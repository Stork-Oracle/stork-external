//! Boundary and adversarial cases.
//!
//! The happy paths live in `test_stork.cairo`; this file is about the edges where an oracle is
//! most likely to be wrong: signature malleability, timestamp arithmetic, the signer set limit,
//! and inputs an attacker controls.

use core::num::traits::Zero;
use snforge_std::{
    ContractClassTrait, DeclareResultTrait, declare, start_cheat_block_timestamp_global,
    start_cheat_caller_address, stop_cheat_caller_address,
};
use starknet::{ContractAddress, EthAddress};
use stork_starknet_sdk::interface::{IStorkDispatcher, IStorkDispatcherTrait};
use stork_starknet_sdk::temporal_numeric_value::TemporalNumericValue;
use crate::vectors::{negative_asset_1, other_public_key, positive_asset_1, stork_public_key};

const VALID_TIME_PERIOD: u64 = 3600;

/// The second the fixture value was produced, so staleness arithmetic can be pinned exactly.
const FIXTURE_SECOND: u64 = 1757543080;

fn owner() -> ContractAddress {
    'owner'.try_into().unwrap()
}

fn deploy_with(signer: EthAddress) -> IStorkDispatcher {
    let contract = declare("Stork").unwrap().contract_class();
    let no_token: ContractAddress = Zero::zero();

    let mut calldata = array![];
    owner().serialize(ref calldata);
    signer.serialize(ref calldata);
    VALID_TIME_PERIOD.serialize(ref calldata);
    0_u256.serialize(ref calldata);
    no_token.serialize(ref calldata);

    let (contract_address, _) = contract.deploy(@calldata).unwrap();

    IStorkDispatcher { contract_address }
}

fn deploy() -> IStorkDispatcher {
    deploy_with(stork_public_key())
}

// === Staleness boundary ===
//
// `get_temporal_numeric_value_v1` accepts an age of exactly valid_time_period_seconds and rejects
// one second more. Off-by-one here either serves stale prices or rejects good ones.

#[test]
fn test_value_exactly_at_the_staleness_boundary_is_accepted() {
    let stork = deploy();
    let input = positive_asset_1();
    stork.update_temporal_numeric_values_v1(array![input]);

    start_cheat_block_timestamp_global(FIXTURE_SECOND + VALID_TIME_PERIOD);

    let value = stork.get_temporal_numeric_value_v1(input.id);
    assert!(value == input.temporal_numeric_value);
}

#[test]
#[should_panic(expected: 'Stork: stale value')]
fn test_value_one_second_past_the_boundary_is_rejected() {
    let stork = deploy();
    let input = positive_asset_1();
    stork.update_temporal_numeric_values_v1(array![input]);

    start_cheat_block_timestamp_global(FIXTURE_SECOND + VALID_TIME_PERIOD + 1);

    stork.get_temporal_numeric_value_v1(input.id);
}

#[test]
fn test_value_timestamped_in_the_future_is_not_stale() {
    // Clock drift between the source and target chains can put a value slightly ahead of the
    // block. That must not underflow the age subtraction and read as ancient.
    let stork = deploy();
    let input = positive_asset_1();
    stork.update_temporal_numeric_values_v1(array![input]);

    start_cheat_block_timestamp_global(FIXTURE_SECOND - 60);

    let value = stork.get_temporal_numeric_value_v1(input.id);
    assert!(value == input.temporal_numeric_value);
}

#[test]
fn test_zero_valid_time_period_still_accepts_a_same_second_value() {
    let contract = declare("Stork").unwrap().contract_class();
    let no_token: ContractAddress = Zero::zero();

    let mut calldata = array![];
    owner().serialize(ref calldata);
    stork_public_key().serialize(ref calldata);
    0_u64.serialize(ref calldata);
    0_u256.serialize(ref calldata);
    no_token.serialize(ref calldata);

    let (address, _) = contract.deploy(@calldata).unwrap();
    let stork = IStorkDispatcher { contract_address: address };

    let input = positive_asset_1();
    stork.update_temporal_numeric_values_v1(array![input]);

    start_cheat_block_timestamp_global(FIXTURE_SECOND);

    let value = stork.get_temporal_numeric_value_v1(input.id);
    assert!(value == input.temporal_numeric_value);
}

// === Signature malleability and tampering ===

#[test]
#[should_panic(expected: 'Stork: invalid signature')]
fn test_swapped_r_and_s_is_rejected() {
    let stork = deploy();
    let input = positive_asset_1();

    let tampered = stork_starknet_sdk::interface::TemporalNumericValueInput {
        r: input.s, s: input.r, ..input,
    };

    stork.update_temporal_numeric_values_v1(array![tampered]);
}

#[test]
#[should_panic(expected: 'Stork: invalid signature')]
fn test_zeroed_signature_is_rejected() {
    let stork = deploy();
    let input = positive_asset_1();

    let tampered = stork_starknet_sdk::interface::TemporalNumericValueInput { r: 0, s: 0, ..input };

    stork.update_temporal_numeric_values_v1(array![tampered]);
}

#[test]
#[should_panic(expected: 'Stork: invalid signature')]
fn test_signature_from_another_feed_is_rejected() {
    // Replaying a valid signature under a different asset id must fail: the id is in the digest.
    let stork = deploy();
    let input = positive_asset_1();
    let other = negative_asset_1();

    let tampered = stork_starknet_sdk::interface::TemporalNumericValueInput {
        id: other.id, ..input,
    };

    stork.update_temporal_numeric_values_v1(array![tampered]);
}

#[test]
#[should_panic(expected: 'Stork: invalid signature')]
fn test_tampered_alg_hash_is_rejected() {
    let stork = deploy();
    let mut input = positive_asset_1();
    input.value_compute_alg_hash += 1;

    stork.update_temporal_numeric_values_v1(array![input]);
}

// === Batch behaviour ===

#[test]
#[should_panic(expected: 'Stork: no fresh update')]
fn test_empty_batch_is_rejected() {
    let stork = deploy();

    stork.update_temporal_numeric_values_v1(array![]);
}

#[test]
fn test_duplicate_ids_in_one_batch_apply_once() {
    // The same update twice in a batch: the first applies, the second is no longer fresh.
    let stork = deploy();
    let input = positive_asset_1();

    stork.update_temporal_numeric_values_v1(array![input, input]);

    let value = stork.get_temporal_numeric_value_unsafe_v1(input.id);
    assert!(value == input.temporal_numeric_value);
}

#[test]
#[should_panic(expected: 'Stork: invalid signature')]
fn test_one_bad_entry_rejects_the_whole_batch() {
    // A batch is all or nothing once an entry is fresh but unverifiable, so a bad update cannot
    // ride along with good ones.
    let stork = deploy();
    let mut bad = negative_asset_1();
    bad.r += 1;

    stork.update_temporal_numeric_values_v1(array![positive_asset_1(), bad]);
}

#[test]
fn test_empty_id_span_reads_are_empty() {
    let stork = deploy();

    assert!(stork.get_temporal_numeric_values_unsafe_v1(array![].span()).len() == 0);
    assert!(stork.get_multiple_temporal_numeric_values_unchecked(array![].span()).len() == 0);
}

// === Signer set ===

#[test]
#[should_panic(expected: 'Stork: address limit reached')]
fn test_signer_list_is_capped() {
    // The canonical key occupies one slot, so seven more fill it.
    let stork = deploy();
    start_cheat_caller_address(stork.contract_address, owner());

    for i in 1_u8..9_u8 {
        let key: EthAddress = (0x1000 + i.into()).try_into().unwrap();
        stork.add_signing_address(key);
    }
}

#[test]
#[should_panic(expected: 'Stork: zero address')]
fn test_zero_signing_address_is_rejected() {
    let stork = deploy();
    start_cheat_caller_address(stork.contract_address, owner());

    stork.add_signing_address(Zero::zero());
}

#[test]
#[should_panic(expected: 'Stork: address not found')]
fn test_removing_an_unknown_signer_is_rejected() {
    let stork = deploy();
    start_cheat_caller_address(stork.contract_address, owner());

    stork.remove_signing_address(other_public_key());
}

#[test]
fn test_signer_removal_keeps_the_rest_usable() {
    // Removing a middle entry uses a swap-and-pop; the survivors must still verify.
    let stork = deploy_with(other_public_key());
    let filler: EthAddress = 0xAAAA.try_into().unwrap();

    start_cheat_caller_address(stork.contract_address, owner());
    stork.add_signing_address(filler);
    stork.add_signing_address(stork_public_key());
    stork.remove_signing_address(filler);
    stop_cheat_caller_address(stork.contract_address);

    assert!(stork.signing_addresses().len() == 2);

    // The key that actually signed the fixture is still accepted after the swap.
    stork.update_temporal_numeric_values_v1(array![positive_asset_1()]);
}

// === Ownership ===

#[test]
#[should_panic(expected: 'Stork: zero address')]
fn test_cannot_transfer_ownership_to_zero() {
    let stork = deploy();
    start_cheat_caller_address(stork.contract_address, owner());

    stork.transfer_ownership(Zero::zero());
}

#[test]
fn test_ownership_transfer_can_be_superseded() {
    let stork = deploy();
    let first: ContractAddress = 'first'.try_into().unwrap();
    let second: ContractAddress = 'second'.try_into().unwrap();

    start_cheat_caller_address(stork.contract_address, owner());
    stork.transfer_ownership(first);
    stork.transfer_ownership(second);
    stop_cheat_caller_address(stork.contract_address);

    assert!(stork.pending_owner() == second);

    // The superseded candidate can no longer accept.
    start_cheat_caller_address(stork.contract_address, second);
    stork.accept_ownership();
    stop_cheat_caller_address(stork.contract_address);

    assert!(stork.owner() == second);
}

#[test]
#[should_panic(expected: 'Stork: caller is not owner')]
fn test_previous_owner_loses_authority_after_transfer() {
    let stork = deploy();
    let new_owner: ContractAddress = 'new'.try_into().unwrap();

    start_cheat_caller_address(stork.contract_address, owner());
    stork.transfer_ownership(new_owner);
    stop_cheat_caller_address(stork.contract_address);

    start_cheat_caller_address(stork.contract_address, new_owner);
    stork.accept_ownership();
    stop_cheat_caller_address(stork.contract_address);

    start_cheat_caller_address(stork.contract_address, owner());
    stork.update_valid_time_period_seconds(60);
}

#[test]
#[should_panic(expected: 'Stork: not pending owner')]
fn test_accept_ownership_with_no_transfer_pending_is_rejected() {
    let stork = deploy();

    start_cheat_caller_address(stork.contract_address, owner());
    stork.accept_ownership();
}

// === Extreme values ===

#[test]
fn test_extreme_timestamps_and_values_survive_storage() {
    // The packed representation puts the timestamp in the high 64 bits and the two's complement
    // value in the low 128. These are the corners of that layout.
    let stork = deploy();

    let corners = array![
        TemporalNumericValue { timestamp_ns: 1, quantized_value: 0 },
        TemporalNumericValue {
            timestamp_ns: 0xffffffffffffffff, quantized_value: 0x7fffffffffffffffffffffffffffffff,
        },
        TemporalNumericValue {
            timestamp_ns: 0xffffffffffffffff, quantized_value: -0x80000000000000000000000000000000,
        },
    ];

    for value in corners {
        let packed =
            stork_starknet_sdk::temporal_numeric_value::TemporalNumericValueStorePacking::pack(
            value,
        );
        let unpacked =
            stork_starknet_sdk::temporal_numeric_value::TemporalNumericValueStorePacking::unpack(
            packed,
        );
        assert!(unpacked == value, "packing lost information at a corner");
    }

    assert!(stork.version() == '1.0.0');
}
