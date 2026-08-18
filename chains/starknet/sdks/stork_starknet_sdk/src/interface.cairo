//! The public interface of the Stork Starknet contract.

use starknet::{ClassHash, ContractAddress, EthAddress};
use crate::temporal_numeric_value::{EncodedAssetId, TemporalNumericValue};

/// A single signed update, mirroring `StorkStructs.TemporalNumericValueInput` from the EVM SDK.
#[derive(Copy, Drop, Serde, PartialEq, Debug)]
pub struct TemporalNumericValueInput {
    /// The encoded asset id the update is for.
    pub id: EncodedAssetId,
    /// The signed value.
    pub temporal_numeric_value: TemporalNumericValue,
    /// The merkle root of the publisher prices this value was aggregated from.
    pub publisher_merkle_root: u256,
    /// The checksum of the algorithm used to compute the value.
    pub value_compute_alg_hash: u256,
    /// The `r` component of the Stork signature.
    pub r: u256,
    /// The `s` component of the Stork signature.
    pub s: u256,
    /// The `v` component of the Stork signature. Must be 27 or 28.
    pub v: u8,
}

/// The consumer-facing surface of the contract. Read methods here are what a dapp integrating
/// Stork on Starknet calls; `update_temporal_numeric_values_v1` is what the chain pusher calls.
#[starknet::interface]
pub trait IStork<TContractState> {
    /// Verifies and applies a batch of signed updates.
    ///
    /// Updates that are not newer than the stored value are skipped. Panics with
    /// `Stork: invalid signature` if any non-skipped update fails verification, and with
    /// `Stork: no fresh update` if no update in the batch was fresh.
    ///
    /// The caller is charged `single_update_fee` per applied update, collected in `fee_token`.
    fn update_temporal_numeric_values_v1(
        ref self: TContractState, update_data: Array<TemporalNumericValueInput>,
    );

    /// Returns the latest value for `id`, panicking if it is missing or older than
    /// `valid_time_period_seconds`.
    fn get_temporal_numeric_value_v1(
        self: @TContractState, id: EncodedAssetId,
    ) -> TemporalNumericValue;

    /// Returns the latest value for `id` without a staleness check, panicking if it is missing.
    fn get_temporal_numeric_value_unsafe_v1(
        self: @TContractState, id: EncodedAssetId,
    ) -> TemporalNumericValue;

    /// Batch form of [`get_temporal_numeric_value_unsafe_v1`], which reverts if *any* id is
    /// missing. Consumers reading a fixed set of live feeds want this; pollers do not.
    fn get_temporal_numeric_values_unsafe_v1(
        self: @TContractState, ids: Span<EncodedAssetId>,
    ) -> Array<TemporalNumericValue>;

    /// Batch read that never reverts: unknown feeds come back as a zeroed
    /// [`TemporalNumericValue`], which callers detect by `timestamp_ns == 0`.
    ///
    /// This is what the chain pusher polls with, so that one not-yet-populated feed does not
    /// blind it to every other feed in the batch.
    fn get_multiple_temporal_numeric_values_unchecked(
        self: @TContractState, ids: Span<EncodedAssetId>,
    ) -> Array<TemporalNumericValue>;

    /// Returns the total fee required to submit `update_data`.
    fn get_update_fee_v1(
        self: @TContractState, update_data: Span<TemporalNumericValueInput>,
    ) -> u256;

    /// The contract version, as a short string.
    fn version(self: @TContractState) -> felt252;

    // === Configuration getters ===

    /// The canonical Stork signing key.
    fn stork_public_key(self: @TContractState) -> EthAddress;

    /// All keys currently accepted as Stork signers, including the canonical one.
    fn signing_addresses(self: @TContractState) -> Array<EthAddress>;

    /// The fee charged per applied update, denominated in `fee_token`.
    fn single_update_fee(self: @TContractState) -> u256;

    /// The ERC20 token fees are collected in.
    fn fee_token(self: @TContractState) -> ContractAddress;

    /// How old a value may be before [`get_temporal_numeric_value_v1`] rejects it.
    fn valid_time_period_seconds(self: @TContractState) -> u64;

    // === Admin ===

    fn update_valid_time_period_seconds(ref self: TContractState, valid_time_period_seconds: u64);
    fn update_single_update_fee(ref self: TContractState, single_update_fee: u256);
    fn update_fee_token(ref self: TContractState, fee_token: ContractAddress);
    fn update_stork_public_key(ref self: TContractState, stork_public_key: EthAddress);
    fn add_signing_address(ref self: TContractState, signing_address: EthAddress);
    fn remove_signing_address(ref self: TContractState, signing_address: EthAddress);
    /// Withdraws collected fees to `recipient`.
    fn withdraw_fees(ref self: TContractState, recipient: ContractAddress, amount: u256);
    /// Replaces the contract's class hash. Starknet's native equivalent of a UUPS upgrade.
    fn upgrade(ref self: TContractState, new_class_hash: ClassHash);

    // === Two step ownership, mirroring `Ownable2StepUpgradeable` ===

    fn owner(self: @TContractState) -> ContractAddress;
    fn pending_owner(self: @TContractState) -> ContractAddress;
    fn transfer_ownership(ref self: TContractState, new_owner: ContractAddress);
    fn accept_ownership(ref self: TContractState);
}

/// The subset of ERC20 the contract needs in order to collect and withdraw fees.
#[starknet::interface]
pub trait IERC20<TContractState> {
    fn transfer(ref self: TContractState, recipient: ContractAddress, amount: u256) -> bool;
    fn transfer_from(
        ref self: TContractState, sender: ContractAddress, recipient: ContractAddress, amount: u256,
    ) -> bool;
    fn balance_of(self: @TContractState, account: ContractAddress) -> u256;
}
