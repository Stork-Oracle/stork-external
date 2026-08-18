//! Events emitted by the Stork contract.
//!
//! Consumers subscribe to [`ValueUpdate`] to react to feed changes without polling; the rest
//! record configuration changes.

use starknet::{ClassHash, ContractAddress, EthAddress};
use crate::temporal_numeric_value::EncodedAssetId;

/// Emitted whenever a feed advances. The chain pusher subscribes to this to track on-chain state
/// without polling.
#[derive(Drop, starknet::Event)]
pub struct ValueUpdate {
    #[key]
    pub id: EncodedAssetId,
    pub timestamp_ns: u64,
    pub quantized_value: i128,
}

#[derive(Drop, starknet::Event)]
pub struct StorkPublicKeyUpdate {
    pub stork_public_key: EthAddress,
}

#[derive(Drop, starknet::Event)]
pub struct SigningAddressAdded {
    pub signing_address: EthAddress,
}

#[derive(Drop, starknet::Event)]
pub struct SigningAddressRemoved {
    pub signing_address: EthAddress,
}

#[derive(Drop, starknet::Event)]
pub struct SingleUpdateFeeUpdate {
    pub single_update_fee: u256,
}

#[derive(Drop, starknet::Event)]
pub struct FeeTokenUpdate {
    pub fee_token: ContractAddress,
}

#[derive(Drop, starknet::Event)]
pub struct ValidTimePeriodUpdate {
    pub valid_time_period_seconds: u64,
}

#[derive(Drop, starknet::Event)]
pub struct OwnershipTransferStarted {
    #[key]
    pub previous_owner: ContractAddress,
    #[key]
    pub new_owner: ContractAddress,
}

#[derive(Drop, starknet::Event)]
pub struct OwnershipTransferred {
    #[key]
    pub previous_owner: ContractAddress,
    #[key]
    pub new_owner: ContractAddress,
}

#[derive(Drop, starknet::Event)]
pub struct Upgraded {
    pub class_hash: ClassHash,
}
