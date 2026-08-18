//! Panic messages for the Stork Starknet contract.
//!
//! Starknet has no typed errors, so these short strings play the role of the EVM contract's
//! `StorkErrors` custom errors.

pub mod StorkErrors {
    pub const INVALID_SIGNATURE: felt252 = 'Stork: invalid signature';
    pub const NO_FRESH_UPDATE: felt252 = 'Stork: no fresh update';
    pub const NOT_FOUND: felt252 = 'Stork: not found';
    pub const STALE_VALUE: felt252 = 'Stork: stale value';
    pub const INSUFFICIENT_FEE: felt252 = 'Stork: insufficient fee';
    pub const FEE_TOKEN_UNSET: felt252 = 'Stork: fee token unset';
    pub const NOT_OWNER: felt252 = 'Stork: caller is not owner';
    pub const NOT_PENDING_OWNER: felt252 = 'Stork: not pending owner';
    pub const ZERO_ADDRESS: felt252 = 'Stork: zero address';
    pub const ZERO_CLASS_HASH: felt252 = 'Stork: zero class hash';
    pub const ADDRESS_ALREADY_EXISTS: felt252 = 'Stork: address exists';
    pub const ADDRESS_NOT_FOUND: felt252 = 'Stork: address not found';
    pub const ADDRESS_LIMIT_REACHED: felt252 = 'Stork: address limit reached';
    pub const CANNOT_REMOVE_CANONICAL: felt252 = 'Stork: rotate key first';
}
