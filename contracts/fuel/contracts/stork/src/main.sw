contract;

use std::bytes::*;
use std::bytes_conversions::u64::*;
use std::crypto::secp256k1::Secp256k1;
use std::crypto::message::Message;
use std::string::String;
use std::logging::log;
use std::u128::U128;

use standards::src5::*;
use sway_libs::signed_integers::i128::I128;

use stork_sway_sdk::errors::StorkError;
use stork_sway_sdk::temporal_numeric_value::TemporalNumericValue;
use stork_sway_sdk::events::StorkEvent;

struct TemporalNumericValueInput {
    temporal_numeric_value: TemporalNumericValue,
    id: b256,
    publisher_merkle_root: b256,
    value_compute_alg_hash: b256,
    r: b256, 
    s: b256, 
}


struct State {
    // For verifying the authenticity of the passed data
    stork_public_key: Identity,
    single_update_fee_in_wei: u64,
    // Mapping of cached numeric temporal data
    latest_canonical_temporal_numeric_values: StorageMap<b256, TemporalNumericValue>,
}

storage {
    /// The owner in storage.
    owner: standards::src5::State = standards::src5::State::Uninitialized,
    initialized: bool = false,
    initializing: bool = false,
    state: State = State {
        stork_public_key: Identity::Address(Address::zero()),
        single_update_fee_in_wei: 0,
        // valid_time_period_seconds: 0,
        latest_canonical_temporal_numeric_values: StorageMap::<b256, TemporalNumericValue> {},
    },
}


#[storage(read)]
fn latest_canonical_temporal_numeric_value(id: b256) -> Result<TemporalNumericValue, StorkError> {
    let map: StorageKey<StorageMap<b256, TemporalNumericValue>> = storage.state.latest_canonical_temporal_numeric_values;
    match map.get(id).try_read() {
        Some(tnv) => Ok(tnv),
        None => Err(StorkError::FeedNotFound),
    }
}

#[storage(read, write)]
fn update_latest_value_if_necessary(input: TemporalNumericValueInput) -> bool {
    let mut latestReceiveTime = 0;
    match latest_canonical_temporal_numeric_value(input.id) {
        Ok(tnv) => {
            latestReceiveTime = tnv.get_timestamp_ns();
        },
        _ => {},
    }
    if (input.temporal_numeric_value.timestamp_ns > latestReceiveTime) {
        storage.state.latest_canonical_temporal_numeric_values.insert(input.id, input.temporal_numeric_value);
        let event = StorkEvent::ValueUpdate((input.id, input.temporal_numeric_value));
        log(event);
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


fn get_signer(
    signedMessageHash: b256,
    r: b256,
    s: b256,
) -> Identity {
    Identity::Address(Address::from(Secp256k1::from((r, s)).address(Message::from(signedMessageHash)).unwrap()))
}

// helper function to convert I128 to [u8; 24]
fn i128_to_be_bytes(value: I128) -> [u8; 24] {
    let mut bytes = [0u8; 24];
    
    // Get the underlying U128 value
    let mut u128_value = value.underlying();
    
    // If the value is greater than indent (positive number)
    // subtract indent to get the actual positive value
    if u128_value > I128::indent() {
        u128_value = u128_value - I128::indent();
    } else if u128_value < I128::indent() {
        // For negative numbers, calculate two's complement
        // First get the absolute value (distance from indent)
        u128_value = I128::indent() - u128_value;
        // Then convert to two's complement
        u128_value = !u128_value + U128::from(1u64);
        // Set the sign bit
        bytes[8] = 0x80;
    }
    
    // Convert U128 to bytes, filling the last 16 bytes
    let mut i = 23;
    while i >= 8 {
        let bytes_val = (u128_value.binary_and(U128::from((0, 255)))).as_u64().unwrap().try_as_u8();
        if bytes_val.is_none() {
            log(StorkError::InvalidSignature);
            revert(0);
        }
        // safe unwrap
        bytes[i] = bytes_val.unwrap();
        u128_value = u128_value >> 8;
        i -= 1;
    }
    
    bytes
}

fn _verify_stork_signature_v1(
    storkPubKey: Identity,
    id: b256,
    recvTime: u64,
    quantized_value: I128,
    publisher_merkle_root: b256,
    value_compute_alg_hash: b256,
    r: b256,
    s: b256,
) -> bool {
    // convert quantized_value to [u8; 24]
    let quantized_value = i128_to_be_bytes(quantized_value);

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
        single_update_fee_in_wei: u64
    );

    #[storage(read)]
    fn single_update_fee_in_wei() -> u64;
    
 
    
    #[storage(read)]
    fn stork_public_key() -> Identity;


    fn verify_stork_signature_v1(
        storkPubKey: Identity,
        id: b256,
        recvTime: u64,
        quantized_value: I128,
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
    fn get_temporal_numeric_value_unchecked_v1(id: b256) -> TemporalNumericValue;

    fn version() -> String;

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
        single_update_fee_in_wei: u64
    ) {
        require(!storage.initialized.read(), "Already initialized");
        require(!storage.initializing.read(), "Already initializing");
        storage.initializing.write(true);

        storage.owner.write(standards::src5::State::Initialized(initialOwner));

        set_single_update_fee_in_wei(single_update_fee_in_wei);
        set_stork_public_key(stork_public_key);
        storage.initialized.write(true);
    }

    #[storage(read)]
    fn single_update_fee_in_wei() -> u64 {
        _single_update_fee_in_wei()
    }
    
    
    #[storage(read)]
    fn stork_public_key() -> Identity {
        _stork_public_key()
    }

    fn verify_stork_signature_v1(
        storkPubKey: Identity,
        id: b256,
        recvTime: u64,
        quantized_value: I128,
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
    fn get_temporal_numeric_value_unchecked_v1(id: b256) -> TemporalNumericValue {
        let latestValueResult: Result<TemporalNumericValue, StorkError> = latest_canonical_temporal_numeric_value(id);
        if (latestValueResult.is_err()) {
            log(StorkError::FeedNotFound);
            revert(0);
        }

        // This unwrap is safe as we've checked the error case
        latestValueResult.unwrap()
    }

    fn version() -> String {
        return String::from_ascii_str("1.0.2");
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

