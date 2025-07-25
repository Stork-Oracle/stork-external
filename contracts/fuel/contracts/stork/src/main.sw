contract;

mod verify;

use std::bytes::*;
use std::bytes_conversions::u64::*;
use std::string::String;
use std::logging::log;
use std::storage::storage_bytes::*;
use std::vm::evm::evm_address::EvmAddress;

use standards::src5::*;
use sway_libs::signed_integers::i128::I128;

use verify::verify_stork_signature;

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
    v: u8,
}

struct State {
    // For verifying the authenticity of the passed data
    stork_public_key: EvmAddress,
    single_update_fee_in_wei: u64,
    // Mapping of cached numeric temporal data
    latest_canonical_temporal_numeric_values: StorageMap<b256, TemporalNumericValue>,
}

storage {
    /// The owner in storage.
    owner: standards::src5::State = standards::src5::State::Uninitialized,
    proposed_owner: Identity = Identity::Address(Address::zero()),
    initialized: bool = false,
    initializing: bool = false,
    state: State = State {
        stork_public_key: EvmAddress::zero(),
        single_update_fee_in_wei: 0,
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
    if (input.temporal_numeric_value.timestamp_ns > latestReceiveTime)
    {
        storage
            .state
            .latest_canonical_temporal_numeric_values
            .insert(input.id, input.temporal_numeric_value);
        let event = StorkEvent::ValueUpdate((input.id, input.temporal_numeric_value));
        log(event);
        return true;
    }
    false
}

#[storage(read, write)]
fn set_stork_public_key(stork_public_key: EvmAddress) {
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

#[storage(read)]
fn _stork_public_key() -> EvmAddress {
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
fn _update_stork_public_key(stork_public_key: EvmAddress) {
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
        initial_owner: Identity,
        stork_public_key: EvmAddress,
        single_update_fee_in_wei: u64,
    );

    #[storage(read)]
    fn single_update_fee_in_wei() -> u64;

    #[storage(read)]
    fn stork_public_key() -> EvmAddress;

    fn verify_stork_signature_v1(
        stork_pubkey: EvmAddress,
        id: b256,
        recv_time: u64,
        quantized_value: I128,
        publisher_merkle_root: b256,
        value_compute_alg_hash: b256,
        r: b256,
        s: b256,
        v: u8,
    ) -> bool;

    #[storage(read, write), payable]
    fn update_temporal_numeric_values_v1(update_data: Vec<TemporalNumericValueInput>);

    #[storage(read)]
    fn get_update_fee_v1(update_data: Vec<TemporalNumericValueInput>) -> u64;

    #[storage(read)]
    fn get_temporal_numeric_value_unchecked_v1(id: b256) -> TemporalNumericValue;

    fn version() -> String;

    #[storage(read, write)]
    fn update_single_update_fee_in_wei(single_update_fee_in_wei: u64);

    #[storage(read, write)]
    fn update_stork_public_key(stork_public_key: EvmAddress);

    #[storage(read, write)]
    fn propose_owner(new_owner: Address);

    #[storage(read, write)]
    fn accept_ownership();
}

impl Stork for Contract {
    #[storage(read, write)]
    fn initialize(
        initial_owner: Identity,
        stork_public_key: EvmAddress,
        single_update_fee_in_wei: u64,
    ) {
        require(!storage.initialized.read(), "Already initialized");
        require(!storage.initializing.read(), "Already initializing");
        storage.initializing.write(true);

        storage
            .owner
            .write(standards::src5::State::Initialized(initial_owner));

        set_single_update_fee_in_wei(single_update_fee_in_wei);
        set_stork_public_key(stork_public_key);
        storage.initialized.write(true);
    }

    #[storage(read)]
    fn single_update_fee_in_wei() -> u64 {
        _single_update_fee_in_wei()
    }

    #[storage(read)]
    fn stork_public_key() -> EvmAddress {
        _stork_public_key()
    }

    fn verify_stork_signature_v1(
        stork_pubkey: EvmAddress,
        id: b256,
        recv_time: u64,
        quantized_value: I128,
        publisher_merkle_root: b256,
        value_compute_alg_hash: b256,
        r: b256,
        s: b256,
        v: u8,
    ) -> bool {
        verify_stork_signature(
            stork_pubkey,
            id,
            recv_time,
            quantized_value,
            publisher_merkle_root,
            value_compute_alg_hash,
            r,
            s,
            v,
        )
    }

    #[storage(read, write), payable]
    fn update_temporal_numeric_values_v1(update_data: Vec<TemporalNumericValueInput>) {
        let mut num_updates = 0;
        let mut i = 0;
        while i < update_data.len() {
            let x = update_data.get(i).unwrap();
            let verified = verify_stork_signature(
                _stork_public_key(),
                x.id,
                x.temporal_numeric_value
                    .timestamp_ns,
                x.temporal_numeric_value
                    .quantized_value,
                x.publisher_merkle_root,
                x.value_compute_alg_hash,
                x.r,
                x.s,
                x.v,
            );

            if (!verified) {
                log(StorkError::InvalidSignature);
                revert(0);
            }
            let updated = update_latest_value_if_necessary(update_data.get(i).unwrap());
            if (updated) {
                num_updates += 1;
            }

            i += 1;
        }
        if (num_updates == 0) {
            log(StorkError::NoFreshUpdate);
            revert(0);
        }

        let required_fee = get_total_fee(num_updates);
        if (std::call_frames::msg_asset_id() != AssetId::base()) {
            log(StorkError::InsufficientFee);
            revert(0)
        }
        if (std::context::msg_amount() < required_fee) {
            log(StorkError::InsufficientFee);
            revert(0);
        }
    }

    #[storage(read)]
    fn get_update_fee_v1(update_data: Vec<TemporalNumericValueInput>) -> u64 {
        get_total_fee(update_data.len())
    }

    #[storage(read)]
    fn get_temporal_numeric_value_unchecked_v1(id: b256) -> TemporalNumericValue {
        let latest_value = match latest_canonical_temporal_numeric_value(id) {
            Ok(value) => value,
            Err(error) => {
                log(error);
                revert(0);
            }
        };
        latest_value
    }

    fn version() -> String {
        return String::from_ascii_str("1.0.0");
    }

    #[storage(read, write)]
    fn update_single_update_fee_in_wei(single_update_fee_in_wei: u64) {
        _update_single_update_fee_in_wei(single_update_fee_in_wei)
    }

    #[storage(read, write)]
    fn update_stork_public_key(stork_public_key: EvmAddress) {
        _update_stork_public_key(stork_public_key)
    }

    #[storage(read, write)]
    fn propose_owner(new_owner: Address) {
        only_owner();
        storage.proposed_owner.write(Identity::Address(new_owner));
    }

    #[storage(read, write)]
    fn accept_ownership() {
        require(storage.proposed_owner.read() == msg_sender().unwrap(), "Only proposed owner can accept ownership");
        storage.owner.write(standards::src5::State::Initialized(storage.proposed_owner.read()));
        storage.proposed_owner.write(Identity::Address(Address::zero()));
    }
}

impl standards::src5::SRC5 for Contract {
    #[storage(read)]
    fn owner() -> standards::src5::State {
        storage.owner.read()
    }
}
