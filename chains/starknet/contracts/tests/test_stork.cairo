//! End to end tests against a deployed Stork contract.

use core::num::traits::Zero;
use snforge_std::{
    ContractClassTrait, DeclareResultTrait, EventSpyAssertionsTrait, declare, spy_events,
    start_cheat_block_timestamp_global, start_cheat_caller_address, stop_cheat_caller_address,
};
use starknet::{ContractAddress, EthAddress};
use stork::stork::Stork;
use stork_starknet_sdk::events::ValueUpdate;
use stork_starknet_sdk::interface::{IStorkDispatcher, IStorkDispatcherTrait};
use crate::mock_erc20::{IMockERC20Dispatcher, IMockERC20DispatcherTrait};
use crate::mock_upgrade::{IUpgradedDispatcher, IUpgradedDispatcherTrait};
use crate::vectors::{
    all_vectors, negative_asset_1, other_public_key, positive_asset_1, positive_asset_2,
    stork_public_key,
};

const VALID_TIME_PERIOD: u64 = 3600;

/// A second past the timestamp of the fixture vectors, so freshly pushed values are not stale.
const FIXTURE_TIME: u64 = 1757543205;

fn owner() -> ContractAddress {
    'owner'.try_into().unwrap()
}

fn pusher() -> ContractAddress {
    'pusher'.try_into().unwrap()
}

fn deploy_with(
    signer: EthAddress, single_update_fee: u256, fee_token: ContractAddress,
) -> IStorkDispatcher {
    let contract = declare("Stork").unwrap().contract_class();

    let mut calldata = array![];
    owner().serialize(ref calldata);
    signer.serialize(ref calldata);
    VALID_TIME_PERIOD.serialize(ref calldata);
    single_update_fee.serialize(ref calldata);
    fee_token.serialize(ref calldata);

    let (contract_address, _) = contract.deploy(@calldata).unwrap();

    IStorkDispatcher { contract_address }
}

fn deploy() -> IStorkDispatcher {
    deploy_with(stork_public_key(), 0, Zero::zero())
}

fn deploy_erc20() -> IMockERC20Dispatcher {
    let contract = declare("MockERC20").unwrap().contract_class();
    let (contract_address, _) = contract.deploy(@array![]).unwrap();

    IMockERC20Dispatcher { contract_address }
}

// === Deployment ===

#[test]
fn test_initial_state() {
    let stork = deploy();

    assert!(stork.owner() == owner());
    assert!(stork.pending_owner() == Zero::zero());
    assert!(stork.stork_public_key() == stork_public_key());
    assert!(stork.valid_time_period_seconds() == VALID_TIME_PERIOD);
    assert!(stork.single_update_fee() == 0);
    assert!(stork.version() == '1.0.0');

    // The canonical key is registered as an accepted signer at construction.
    let signers = stork.signing_addresses();
    assert!(signers.len() == 1);
    assert!(*signers.at(0) == stork_public_key());
}

// === Updates ===

#[test]
fn test_update_and_get_value() {
    let stork = deploy();
    let input = positive_asset_1();

    stork.update_temporal_numeric_values_v1(array![input]);

    let value = stork.get_temporal_numeric_value_unsafe_v1(input.id);
    assert!(value == input.temporal_numeric_value);
}

#[test]
fn test_update_negative_value() {
    let stork = deploy();
    let input = negative_asset_1();

    stork.update_temporal_numeric_values_v1(array![input]);

    let value = stork.get_temporal_numeric_value_unsafe_v1(input.id);
    assert!(value.quantized_value == -100000000000000000000);
    assert!(value.timestamp_ns == input.temporal_numeric_value.timestamp_ns);
}

#[test]
fn test_batch_update() {
    let stork = deploy();

    stork.update_temporal_numeric_values_v1(all_vectors());

    let ids = array![positive_asset_1().id, positive_asset_2().id, negative_asset_1().id];
    let values = stork.get_temporal_numeric_values_unsafe_v1(ids.span());

    assert!(values.len() == 3);
    assert!(*values.at(0) == positive_asset_1().temporal_numeric_value);
    assert!(*values.at(1) == positive_asset_2().temporal_numeric_value);
    assert!(*values.at(2) == negative_asset_1().temporal_numeric_value);
}

#[test]
fn test_update_emits_value_update_event() {
    let stork = deploy();
    let input = positive_asset_1();
    let mut spy = spy_events();

    stork.update_temporal_numeric_values_v1(array![input]);

    spy
        .assert_emitted(
            @array![
                (
                    stork.contract_address,
                    Stork::Event::ValueUpdate(
                        ValueUpdate {
                            id: input.id,
                            timestamp_ns: input.temporal_numeric_value.timestamp_ns,
                            quantized_value: input.temporal_numeric_value.quantized_value,
                        },
                    ),
                ),
            ],
        );
}

#[test]
#[should_panic(expected: 'Stork: no fresh update')]
fn test_replaying_the_same_update_reverts() {
    let stork = deploy();

    stork.update_temporal_numeric_values_v1(array![positive_asset_1()]);
    stork.update_temporal_numeric_values_v1(array![positive_asset_1()]);
}

#[test]
fn test_stale_entries_in_a_batch_are_skipped() {
    let stork = deploy();
    stork.update_temporal_numeric_values_v1(array![positive_asset_1()]);

    // The first entry is now stale, but the batch still applies the second.
    stork.update_temporal_numeric_values_v1(array![positive_asset_1(), positive_asset_2()]);

    let value = stork.get_temporal_numeric_value_unsafe_v1(positive_asset_2().id);
    assert!(value == positive_asset_2().temporal_numeric_value);
}

#[test]
#[should_panic(expected: 'Stork: invalid signature')]
fn test_tampered_value_reverts() {
    let stork = deploy();
    let mut input = positive_asset_1();
    input.temporal_numeric_value.quantized_value += 1;

    stork.update_temporal_numeric_values_v1(array![input]);
}

#[test]
#[should_panic(expected: 'Stork: invalid signature')]
fn test_update_signed_by_unknown_key_reverts() {
    let stork = deploy_with(other_public_key(), 0, Zero::zero());

    stork.update_temporal_numeric_values_v1(array![positive_asset_1()]);
}

#[test]
#[should_panic(expected: 'Stork: invalid signature')]
fn test_non_canonical_v_reverts() {
    let stork = deploy();
    let mut input = positive_asset_1();
    input.v = 1;

    stork.update_temporal_numeric_values_v1(array![input]);
}

// === Reads ===

#[test]
#[should_panic(expected: 'Stork: not found')]
fn test_get_unknown_feed_reverts() {
    let stork = deploy();

    stork.get_temporal_numeric_value_unsafe_v1(0xdeadbeef);
}

#[test]
fn test_get_checked_value_within_valid_period() {
    let stork = deploy();
    let input = positive_asset_1();
    stork.update_temporal_numeric_values_v1(array![input]);

    start_cheat_block_timestamp_global(FIXTURE_TIME);

    let value = stork.get_temporal_numeric_value_v1(input.id);
    assert!(value == input.temporal_numeric_value);
}

#[test]
#[should_panic(expected: 'Stork: stale value')]
fn test_get_checked_value_reverts_when_stale() {
    let stork = deploy();
    let input = positive_asset_1();
    stork.update_temporal_numeric_values_v1(array![input]);

    start_cheat_block_timestamp_global(FIXTURE_TIME + VALID_TIME_PERIOD + 1);

    stork.get_temporal_numeric_value_v1(input.id);
}

#[test]
fn test_unchecked_batch_read_tolerates_missing_feeds() {
    let stork = deploy();
    stork.update_temporal_numeric_values_v1(array![positive_asset_1()]);

    let ids = array![positive_asset_1().id, 0xdeadbeef, positive_asset_2().id];
    let values = stork.get_multiple_temporal_numeric_values_unchecked(ids.span());

    assert!(values.len() == 3);
    assert!(*values.at(0) == positive_asset_1().temporal_numeric_value);
    // Unknown feeds read back zeroed rather than reverting the whole batch.
    assert!(*values.at(1).timestamp_ns == 0);
    assert!(*values.at(2).timestamp_ns == 0);
}

#[test]
#[should_panic(expected: 'Stork: not found')]
fn test_unsafe_batch_read_reverts_on_missing_feed() {
    let stork = deploy();
    stork.update_temporal_numeric_values_v1(array![positive_asset_1()]);

    stork.get_temporal_numeric_values_unsafe_v1(array![positive_asset_1().id, 0xdeadbeef].span());
}

#[test]
fn test_get_update_fee() {
    let token = deploy_erc20();
    let stork = deploy_with(stork_public_key(), 42, token.contract_address);

    assert!(stork.get_update_fee_v1(all_vectors().span()) == 168);
    assert!(stork.get_update_fee_v1(array![].span()) == 0);
}

// === Fees ===

#[test]
fn test_fee_is_collected_per_applied_update() {
    let token = deploy_erc20();
    let stork = deploy_with(stork_public_key(), 10, token.contract_address);

    token.mint(pusher(), 1000);
    start_cheat_caller_address(token.contract_address, pusher());
    token.approve(stork.contract_address, 1000);
    stop_cheat_caller_address(token.contract_address);

    start_cheat_caller_address(stork.contract_address, pusher());
    stork.update_temporal_numeric_values_v1(all_vectors());
    stop_cheat_caller_address(stork.contract_address);

    assert!(token.balance_of(pusher()) == 960, "expected 4 updates to cost 40");
    assert!(token.balance_of(stork.contract_address) == 40);
}

#[test]
#[should_panic(expected: 'ERC20: insufficient allowance')]
fn test_update_reverts_without_allowance() {
    let token = deploy_erc20();
    let stork = deploy_with(stork_public_key(), 10, token.contract_address);
    token.mint(pusher(), 1000);

    start_cheat_caller_address(stork.contract_address, pusher());
    stork.update_temporal_numeric_values_v1(array![positive_asset_1()]);
}

#[test]
fn test_no_allowance_needed_when_fee_is_zero() {
    let token = deploy_erc20();
    let stork = deploy_with(stork_public_key(), 0, token.contract_address);

    start_cheat_caller_address(stork.contract_address, pusher());
    stork.update_temporal_numeric_values_v1(array![positive_asset_1()]);
    stop_cheat_caller_address(stork.contract_address);

    assert!(token.balance_of(stork.contract_address) == 0);
}

#[test]
fn test_owner_can_withdraw_fees() {
    let token = deploy_erc20();
    let stork = deploy_with(stork_public_key(), 0, token.contract_address);
    token.mint(stork.contract_address, 500);

    start_cheat_caller_address(stork.contract_address, owner());
    stork.withdraw_fees(owner(), 500);
    stop_cheat_caller_address(stork.contract_address);

    assert!(token.balance_of(owner()) == 500);
}

// === Key rotation ===

#[test]
fn test_added_signing_address_is_accepted() {
    // Deploy against a key that did not sign the fixtures, then add the one that did.
    let stork = deploy_with(other_public_key(), 0, Zero::zero());

    start_cheat_caller_address(stork.contract_address, owner());
    stork.add_signing_address(stork_public_key());
    stop_cheat_caller_address(stork.contract_address);

    stork.update_temporal_numeric_values_v1(array![positive_asset_1()]);

    let signers = stork.signing_addresses();
    assert!(signers.len() == 2);
}

#[test]
fn test_rotating_the_canonical_key_keeps_the_old_one_valid() {
    let stork = deploy();

    start_cheat_caller_address(stork.contract_address, owner());
    stork.update_stork_public_key(other_public_key());
    stop_cheat_caller_address(stork.contract_address);

    assert!(stork.stork_public_key() == other_public_key());

    // Updates signed by the previous key still verify, because it stays in the signer list.
    stork.update_temporal_numeric_values_v1(array![positive_asset_1()]);
}

#[test]
fn test_removed_signing_address_is_rejected() {
    let stork = deploy_with(other_public_key(), 0, Zero::zero());

    start_cheat_caller_address(stork.contract_address, owner());
    stork.add_signing_address(stork_public_key());
    stork.remove_signing_address(stork_public_key());
    stop_cheat_caller_address(stork.contract_address);

    assert!(stork.signing_addresses().len() == 1);
}

#[test]
#[should_panic(expected: 'Stork: invalid signature')]
fn test_update_after_signer_removal_reverts() {
    let stork = deploy_with(other_public_key(), 0, Zero::zero());

    start_cheat_caller_address(stork.contract_address, owner());
    stork.add_signing_address(stork_public_key());
    stork.remove_signing_address(stork_public_key());
    stop_cheat_caller_address(stork.contract_address);

    stork.update_temporal_numeric_values_v1(array![positive_asset_1()]);
}

#[test]
#[should_panic(expected: 'Stork: rotate key first')]
fn test_cannot_remove_canonical_key() {
    let stork = deploy();

    start_cheat_caller_address(stork.contract_address, owner());
    stork.remove_signing_address(stork_public_key());
}

#[test]
#[should_panic(expected: 'Stork: address exists')]
fn test_cannot_add_duplicate_signing_address() {
    let stork = deploy();

    start_cheat_caller_address(stork.contract_address, owner());
    stork.add_signing_address(stork_public_key());
}

// === Admin ===

#[test]
fn test_owner_can_update_config() {
    let stork = deploy();
    start_cheat_caller_address(stork.contract_address, owner());

    stork.update_valid_time_period_seconds(60);
    stork.update_fee_token('token'.try_into().unwrap());
    stork.update_single_update_fee(7);

    stop_cheat_caller_address(stork.contract_address);

    assert!(stork.valid_time_period_seconds() == 60);
    assert!(stork.single_update_fee() == 7);
    assert!(stork.fee_token() == 'token'.try_into().unwrap());
}

#[test]
#[should_panic(expected: 'Stork: caller is not owner')]
fn test_non_owner_cannot_update_fee() {
    let stork = deploy();

    start_cheat_caller_address(stork.contract_address, pusher());
    stork.update_single_update_fee(7);
}

#[test]
#[should_panic(expected: 'Stork: caller is not owner')]
fn test_non_owner_cannot_add_signing_address() {
    let stork = deploy();

    start_cheat_caller_address(stork.contract_address, pusher());
    stork.add_signing_address(other_public_key());
}

#[test]
#[should_panic(expected: 'Stork: caller is not owner')]
fn test_non_owner_cannot_upgrade() {
    let stork = deploy();

    start_cheat_caller_address(stork.contract_address, pusher());
    stork.upgrade('class'.try_into().unwrap());
}

#[test]
fn test_upgrade_replaces_the_running_class() {
    let stork = deploy();
    assert!(stork.version() == '1.0.0');

    let new_class = declare("UpgradedStork").unwrap().contract_class();

    start_cheat_caller_address(stork.contract_address, owner());
    stork.upgrade(*new_class.class_hash);
    stop_cheat_caller_address(stork.contract_address);

    let upgraded = IUpgradedDispatcher { contract_address: stork.contract_address };
    assert!(upgraded.version() == '2.0.0', "contract still runs the old class");
}

#[test]
#[should_panic(expected: 'Stork: zero class hash')]
fn test_upgrade_rejects_zero_class_hash() {
    let stork = deploy();

    start_cheat_caller_address(stork.contract_address, owner());
    stork.upgrade(0.try_into().unwrap());
}

// === Fee configuration ===

// The constructor enforces the same fee/fee-token invariant as the setters below. A reverting
// constructor is not interceptable by the test harness, so it is covered by those instead.

#[test]
#[should_panic(expected: 'Stork: fee token unset')]
fn test_cannot_set_a_fee_before_a_fee_token() {
    let stork = deploy();

    start_cheat_caller_address(stork.contract_address, owner());
    stork.update_single_update_fee(10);
}

#[test]
#[should_panic(expected: 'Stork: fee token unset')]
fn test_cannot_clear_the_fee_token_while_charging() {
    let token = deploy_erc20();
    let stork = deploy_with(stork_public_key(), 10, token.contract_address);

    start_cheat_caller_address(stork.contract_address, owner());
    stork.update_fee_token(Zero::zero());
}

#[test]
fn test_fee_can_be_enabled_then_disabled_in_a_safe_order() {
    let token = deploy_erc20();
    let stork = deploy();

    start_cheat_caller_address(stork.contract_address, owner());
    // Token first, then the fee.
    stork.update_fee_token(token.contract_address);
    stork.update_single_update_fee(10);
    // Unwinding requires the reverse order.
    stork.update_single_update_fee(0);
    stork.update_fee_token(Zero::zero());
    stop_cheat_caller_address(stork.contract_address);

    assert!(stork.single_update_fee() == 0);
    assert!(stork.fee_token() == Zero::zero());
}

// === Ownership ===

#[test]
fn test_two_step_ownership_transfer() {
    let stork = deploy();
    let new_owner = pusher();

    start_cheat_caller_address(stork.contract_address, owner());
    stork.transfer_ownership(new_owner);
    stop_cheat_caller_address(stork.contract_address);

    // Ownership does not move until it is accepted.
    assert!(stork.owner() == owner());
    assert!(stork.pending_owner() == new_owner);

    start_cheat_caller_address(stork.contract_address, new_owner);
    stork.accept_ownership();
    stop_cheat_caller_address(stork.contract_address);

    assert!(stork.owner() == new_owner);
    assert!(stork.pending_owner() == Zero::zero());
}

#[test]
#[should_panic(expected: 'Stork: not pending owner')]
fn test_only_pending_owner_can_accept() {
    let stork = deploy();

    start_cheat_caller_address(stork.contract_address, owner());
    stork.transfer_ownership(pusher());
    stop_cheat_caller_address(stork.contract_address);

    start_cheat_caller_address(stork.contract_address, 'someone'.try_into().unwrap());
    stork.accept_ownership();
}
