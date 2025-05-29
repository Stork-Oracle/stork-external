contract;

use std::bytes::*;
use std::bytes_conversions::u64::*;
use std::crypto::secp256k1::Secp256k1;
use std::crypto::message::Message;
use std::string::String;
use standards::src5::*;

enum StorkEvent {
    ValueUpdate: (b256, u64, [u8; 24])
}

enum StorkError {
    InsufficientFee: (),
    NoFreshUpdate: (),
    NotFound: (),
    StaleValue: (),
    InvalidSignature: (),
}

struct TemporalNumericValue {
    // slot 1
    // nanosecond level precision timestamp of latest publisher update in batch
    timestamp_ns: u64, // 8 bytes
    // should be able to hold all necessary numbers (up to 6277101735386680763835789423207666416102355444464034512895)
    quantized_value: [u8; 24], // 8 bytes
}

struct TemporalNumericValueInput {
    temporal_numeric_value: TemporalNumericValue,
    id: b256,
    publisher_merkle_root: b256,
    value_compute_alg_hash: b256,
    r: b256, 
    s: b256, 
}

struct PublisherSignature {
    pub_key: Identity,
    asset_pair_id: String,
    timestamp: u64, // 8 bytes
    quantized_value: b256, // 8 bytes
    r: b256,
    s: b256,
}

struct State {
    // For verifying the authenticity of the passed data
    stork_public_key: Identity,
    single_update_fee_in_wei: u64,
    /// Maximum acceptable time period before value is considered to be stale.
    /// This includes attestation delay, block time, and potential clock drift
    /// between the source/target chains.
    valid_time_period_seconds: u64,
    // Mapping of cached numeric temporal data
    latest_canonical_temporal_numeric_values: Option<StorageKey<StorageMap<b256, TemporalNumericValue>>>,
}

storage {
    /// The owner in storage.
    owner: standards::src5::State = standards::src5::State::Uninitialized,
    initialized: bool = false,
    initializing: bool = false,
    temporal_numeric_value_mapping_instance_count: u64 = 0,
    temporal_numeric_value_mapping_instances: StorageMap<u64, StorageMap<b256, TemporalNumericValue>> = StorageMap {},
    state: State = State {
        stork_public_key: Identity::Address(Address::zero()),
        single_update_fee_in_wei: 0,
        valid_time_period_seconds: 0,
        latest_canonical_temporal_numeric_values: None,
    },
}

#[storage(read, write)]
fn create_temporal_numeric_value_mapping() -> StorageKey<StorageMap<b256, TemporalNumericValue>> {
    let i = storage.temporal_numeric_value_mapping_instance_count.read();
    let result = storage.temporal_numeric_value_mapping_instances.get(i);
    storage.temporal_numeric_value_mapping_instance_count.write(i + 1);
    result
}

#[storage(read)]
fn latest_canonical_temporal_numeric_value(id: b256) -> TemporalNumericValue {
    storage.state.read().latest_canonical_temporal_numeric_values.unwrap().get(id).read()
}

#[storage(read, write)]
fn update_latest_value_if_necessary(input: TemporalNumericValueInput) -> bool {
    let mut _state = storage.state.read();
    let latestReceiveTime = _state.latest_canonical_temporal_numeric_values.unwrap().get(input.id).read().timestamp_ns;
    if (input.temporal_numeric_value.timestamp_ns > latestReceiveTime) {
        _state.latest_canonical_temporal_numeric_values.unwrap().insert(input.id, input.temporal_numeric_value);
        log(StorkEvent::ValueUpdate((
            input.id,
            input.temporal_numeric_value.timestamp_ns,
            input.temporal_numeric_value.quantized_value
        )));
        return true;
    }
    false
}

#[storage(read, write)]
fn set_stork_public_key(stork_public_key: Identity) {
    let mut state = storage.state.read();
    state.stork_public_key = stork_public_key;
    storage.state.write(state);
}

#[storage(read, write)]
fn set_single_update_fee_in_wei(fee: u64) {
    let mut state = storage.state.read();
    state.single_update_fee_in_wei = fee;
    storage.state.write(state);
}

#[storage(read, write)]
fn set_valid_time_period_seconds(valid_time_period_seconds: u64) {
    let mut state = storage.state.read();
    state.valid_time_period_seconds = valid_time_period_seconds;
    storage.state.write(state);
}

fn get_eth_signed_message_hash32(message: b256) -> b256 {
    let mut bytes = Bytes::new();
    
    let s = "\x19Ethereum Signed Message:\n32";
    let a = Bytes::from(raw_slice::from_parts::<u8>(s.as_ptr(), s.len()));
    for x in a.iter() {
        bytes.push(x);
    }

    let b = Bytes::from(message);
    for x in b.iter() {
        bytes.push(x);
    }

    std::hash::keccak256(bytes)
}

fn get_stork_message_hash_v1(
    storkPubKey: Identity,
    id: b256,
    recvTime: u64,
    quantized_value: [u8; 24],
    publisher_merkle_root: b256,
    value_compute_alg_hash: b256
) -> b256 {
    let mut bytes = Bytes::new();
    
    let storkPubKey = Bytes::from(match storkPubKey {
        Identity::Address(x) => b256::from(x),
        Identity::ContractId(x) => b256::from(x),
    });
    for x in storkPubKey.iter() {
        bytes.push(x);
    }
    
    let id = Bytes::from(id);
    for x in id.iter() {
        bytes.push(x);
    }
    
    let recvTime = recvTime.to_be_bytes();
    for x in recvTime.iter() {
        bytes.push(x);
    }

    let mut i = 0;
    while i < 24 {
        bytes.push(quantized_value[i]);
        i += 1;
    }
    
    let publisher_merkle_root = Bytes::from(publisher_merkle_root);
    for x in publisher_merkle_root.iter() {
        bytes.push(x);
    }
    
    let value_compute_alg_hash = Bytes::from(value_compute_alg_hash);
    for x in value_compute_alg_hash.iter() {
        bytes.push(x);
    }
    std::hash::keccak256(bytes)
}

fn get_publisher_message_hash(
    oracleName: Identity,
    asset_pair_id: String,
    timestamp: u64,
    value: b256
) -> b256 {
    let mut bytes = Bytes::new();

    let oracleName = Bytes::from(match oracleName {
        Identity::Address(x) => b256::from(x),
        Identity::ContractId(x) => b256::from(x),
    });
    for x in oracleName.iter() {
        bytes.push(x);
    }

    
    let asset_pair_id = asset_pair_id.as_bytes();
    for x in asset_pair_id.iter() {
        bytes.push(x);
    }

    let timestamp = timestamp.to_be_bytes();
    for x in timestamp.iter() {
        bytes.push(x);
    }

    let value = Bytes::from(value);
    for x in value.iter() {
        bytes.push(x);
    }

    std::hash::keccak256(bytes)
}

fn get_signer(
    signedMessageHash: b256,
    r: b256,
    s: b256,
) -> Identity {
    Identity::Address(Address::from(Secp256k1::from((r, s)).address(Message::from(signedMessageHash)).unwrap()))
}

fn compute_merkle_root(leaves: Vec<b256>) -> b256 {
    require(leaves.len() > 0, "No leaves provided");

    let mut leaves = leaves;

    while (leaves.len() > 1) {
        if (leaves.len() % 2 != 0) {
            // If odd number of leaves, duplicate the last one
            let mut extendedLeaves = Vec::with_capacity(leaves.len() + 1);
            let mut i = 0;
            while i < leaves.len() {
                extendedLeaves.set(i, leaves.get(i).unwrap());
                i += 1;
            }
            extendedLeaves.set(leaves.len(), leaves.get(leaves.len() - 1).unwrap());
            leaves = extendedLeaves;
        }

        let mut nextLevel = Vec::with_capacity(leaves.len() / 2);
        let mut i = 0;
        while i < leaves.len() {
            let mut bytes = Bytes::new();
            for x in Bytes::from(leaves.get(i).unwrap()).iter() {
                bytes.push(x);
            }

            for x in Bytes::from(leaves.get(i + 1).unwrap()).iter() {
                bytes.push(x);
            }

            nextLevel.set(i / 2, std::hash::keccak256(bytes));
            i += 1;
        }
        leaves = nextLevel;
    }

    leaves.get(0).unwrap()
}

#[storage(read, write)]
fn _initialize(stork_public_key: Identity, valid_time_period_seconds: u64, single_update_fee_in_wei: u64) {
    let mut state = storage.state.read();
    state.latest_canonical_temporal_numeric_values = Some(create_temporal_numeric_value_mapping());
    storage.state.write(state);

    set_valid_time_period_seconds(valid_time_period_seconds);
    set_single_update_fee_in_wei(single_update_fee_in_wei);
    set_stork_public_key(stork_public_key);
}

fn _verify_stork_signature_v1(
    storkPubKey: Identity,
    id: b256,
    recvTime: u64,
    quantized_value: [u8; 24],
    publisher_merkle_root: b256,
    value_compute_alg_hash: b256,
    r: b256,
    s: b256,
) -> bool {
    let msgHash = get_stork_message_hash_v1(
        storkPubKey,
        id,
        recvTime,
        quantized_value,
        publisher_merkle_root,
        value_compute_alg_hash
    );

    let signedMessageHash = get_eth_signed_message_hash32(msgHash);

    // Verify hash was generated by the actual user
    let signer = get_signer(signedMessageHash, r, s);
    signer == storkPubKey
}

#[storage(read)]
fn _stork_public_key() -> Identity {
    storage.state.read().stork_public_key
}

#[storage(read)]
fn _single_update_fee_in_wei() -> u64 {
    storage.state.read().single_update_fee_in_wei
}

#[storage(read)]
fn get_total_fee(totalNumUpdates: u64) -> u64 {
    totalNumUpdates * _single_update_fee_in_wei()
}

#[storage(read)]
fn _valid_time_period_seconds() -> u64 {
    storage.state.read().valid_time_period_seconds
}

fn _verify_merkle_root(leaves: Vec<b256>, root: b256) -> bool {
    compute_merkle_root(leaves) == root
}

fn _verify_publisher_signature_v1(
    oraclePubKey: Identity,
    asset_pair_id: String,
    timestamp: u64,
    value: b256,
    r: b256,
    s: b256,
) -> bool {
    let msgHash = get_publisher_message_hash(
        oraclePubKey,
        asset_pair_id,
        timestamp,
        value
    );
    let signedMessageHash = get_eth_signed_message_hash32(msgHash);

    // Verify hash was generated by the actual user
    let signer = get_signer(signedMessageHash, r, s);
    signer == oraclePubKey
}

#[storage(read, write)]
fn _update_valid_time_period_seconds(valid_time_period_seconds: u64) {
    only_owner();
    set_valid_time_period_seconds(valid_time_period_seconds);
}

#[storage(read, write)]
fn _update_single_update_fee_in_wei(maxStorkPerBlock: u64) {
    only_owner();
    set_single_update_fee_in_wei(maxStorkPerBlock);
}

#[storage(read, write)]
fn _update_stork_public_key(stork_public_key: Identity) {
    only_owner();
    set_stork_public_key(stork_public_key);
}

#[storage(read)]
fn only_owner() {
    match storage.owner.read() {
        standards::src5::State::Uninitialized => {},
        standards::src5::State::Initialized(owner) => {
            require(msg_sender().unwrap() == owner, "Only Owner");
        },
        standards::src5::State::Revoked => {}
    }
}

abi Stork {    
    #[storage(read, write)]
    fn initialize(
        initialOwner: Identity,
        stork_public_key: Identity,
        valid_time_period_seconds: u64,
        single_update_fee_in_wei: u64
    );

    #[storage(read)]
    fn single_update_fee_in_wei() -> u64;
    
    #[storage(read)]
    fn valid_time_period_seconds() -> u64;
    
    #[storage(read)]
    fn stork_public_key() -> Identity;

    fn verify_merkle_root(leaves: Vec<b256>, root: b256) -> bool;

    fn verify_publisher_signature_v1(
        oraclePubKey: Identity,
        asset_pair_id: String,
        timestamp: u64,
        value: b256,
        r: b256,
        s: b256,
    ) -> bool;

    fn verify_stork_signature_v1(
        storkPubKey: Identity,
        id: b256,
        recvTime: u64,
        quantized_value: [u8; 24],
        publisher_merkle_root: b256,
        value_compute_alg_hash: b256,
        r: b256,
        s: b256,
    ) -> bool;

    #[storage(read, write), payable]
    fn update_temporal_numeric_values_v1(updateData: Vec<TemporalNumericValueInput>);

    #[storage(read)]
    fn get_update_fee_v1(updateData: Vec<TemporalNumericValueInput>) -> u64;

    #[storage(read)]
    fn get_temporal_numeric_value_v1(id: b256) -> TemporalNumericValue;

    #[storage(read)]
    fn get_temporal_numeric_value_unsafe_v1(id: b256) -> TemporalNumericValue;

    fn verify_publisher_signatures_v1(signatures: Vec<PublisherSignature>, merkleRoot: b256) -> bool;

    fn version() -> String;

    #[storage(read, write)]
    fn update_valid_time_period_seconds(valid_time_period_seconds: u64);

    #[storage(read, write)]
    fn update_single_update_fee_in_wei(single_update_fee_in_wei: u64);

    #[storage(read, write)]
    fn update_stork_public_key(stork_public_key: Identity);
}

impl Stork for Contract {
    #[storage(read, write)]
    fn initialize(
        initialOwner: Identity,
        stork_public_key: Identity,
        valid_time_period_seconds: u64,
        single_update_fee_in_wei: u64
    ) {
        require(!storage.initialized.read(), "Already initialized");
        require(!storage.initializing.read(), "Already initializing");
        storage.initializing.write(true);

        storage.owner.write(standards::src5::State::Initialized(initialOwner));

        _initialize(stork_public_key, valid_time_period_seconds, single_update_fee_in_wei);

        storage.initialized.write(true);
    }

    #[storage(read)]
    fn single_update_fee_in_wei() -> u64 {
        _single_update_fee_in_wei()
    }
    
    #[storage(read)]
    fn valid_time_period_seconds() -> u64 {
        _valid_time_period_seconds()
    }
    
    #[storage(read)]
    fn stork_public_key() -> Identity {
        _stork_public_key()
    }

    fn verify_merkle_root(leaves: Vec<b256>, root: b256) -> bool {
        _verify_merkle_root(leaves, root)
    }

    fn verify_publisher_signature_v1(
        oraclePubKey: Identity,
        asset_pair_id: String,
        timestamp: u64,
        value: b256,
        r: b256,
        s: b256,
    ) -> bool {
        _verify_publisher_signature_v1(oraclePubKey, asset_pair_id, timestamp, value, r, s)
    }

    fn verify_stork_signature_v1(
        storkPubKey: Identity,
        id: b256,
        recvTime: u64,
        quantized_value: [u8; 24],
        publisher_merkle_root: b256,
        value_compute_alg_hash: b256,
        r: b256,
        s: b256,
    ) -> bool {
        _verify_stork_signature_v1(
            storkPubKey,
            id,
            recvTime,
            quantized_value,
            publisher_merkle_root,
            value_compute_alg_hash,
            r,
            s,
        )
    }

    #[storage(read, write), payable]
    fn update_temporal_numeric_values_v1(updateData: Vec<TemporalNumericValueInput>) {
        let mut numUpdates = 0;
        let mut i = 0;
        while i < updateData.len() {
            let x = updateData.get(i).unwrap();
            let verified = _verify_stork_signature_v1(
                _stork_public_key(),
                x.id,
                x.temporal_numeric_value.timestamp_ns,
                x.temporal_numeric_value.quantized_value,
                x.publisher_merkle_root,
                x.value_compute_alg_hash,
                x.r,
                x.s,
            );

            if (!verified) {
                log(StorkError::InvalidSignature);
                revert(0);
            }
            let updated = update_latest_value_if_necessary(updateData.get(i).unwrap());
            if (updated) {
                numUpdates += 1;
            }

            i += 1;
        }
        if (numUpdates == 0) {
            log(StorkError::NoFreshUpdate);
            revert(0);
        }

        let requiredFee = get_total_fee(numUpdates);
        if (std::context::msg_amount() < requiredFee) {
            log(StorkError::InsufficientFee);
            revert(0);
        }
    
    }

    #[storage(read)]
    fn get_update_fee_v1(updateData: Vec<TemporalNumericValueInput>) -> u64 {
        get_total_fee(updateData.len())
    }

    #[storage(read)]
    fn get_temporal_numeric_value_v1(id: b256) -> TemporalNumericValue {
        let numericValue = latest_canonical_temporal_numeric_value(id);
        if (numericValue.timestamp_ns == 0) {
            log(StorkError::NotFound);
            revert(0);
        }

        if (std::block::timestamp() - (numericValue.timestamp_ns / 1000000000) > _valid_time_period_seconds()) {
            log(StorkError::StaleValue);
            revert(0);
        }
        numericValue
    }

    #[storage(read)]
    fn get_temporal_numeric_value_unsafe_v1(id: b256) -> TemporalNumericValue {
        let numericValue = latest_canonical_temporal_numeric_value(id);
        if (numericValue.timestamp_ns == 0) {
            log(StorkError::NotFound);
            revert(0);
        }

        numericValue
    }

    fn verify_publisher_signatures_v1(signatures: Vec<PublisherSignature>, merkleRoot: b256) -> bool {
        let mut hashes = Vec::with_capacity(signatures.len());

        let mut i = 0;
        while i < signatures.len() {
            let s = signatures.get(i).unwrap();
            if(!_verify_publisher_signature_v1(
                s.pub_key,
                s.asset_pair_id,
                s.timestamp,
                s.quantized_value,
                s.r,
                s.s,
            )) {
                return false;
            }

            let computed = get_publisher_message_hash(
                s.pub_key,
                s.asset_pair_id,
                s.timestamp,
                s.quantized_value
            );
            hashes.set(i, computed);
        }
        
        _verify_merkle_root(hashes, merkleRoot)
    }

    fn version() -> String {
        return String::from_ascii_str("1.0.2");
    }

    #[storage(read, write)]
    fn update_valid_time_period_seconds(valid_time_period_seconds: u64) {
        _update_valid_time_period_seconds(valid_time_period_seconds)
    }

    #[storage(read, write)]
    fn update_single_update_fee_in_wei(single_update_fee_in_wei: u64) {
        _update_single_update_fee_in_wei(single_update_fee_in_wei)
    }

    #[storage(read, write)]
    fn update_stork_public_key(stork_public_key: Identity) {
        _update_stork_public_key(stork_public_key)
    }
}

impl standards::src5::SRC5 for Contract {
    #[storage(read)]
    fn owner() -> standards::src5::State {
        storage.owner.read()
    }
}

