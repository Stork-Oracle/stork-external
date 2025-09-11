library;

use ::temporal_numeric_value::TemporalNumericValue;
use std::string::String;
use std::vm::evm::evm_address::EvmAddress;
use signed_int::i128::I128;

pub struct TemporalNumericValueInput {
    pub temporal_numeric_value: TemporalNumericValue,
    pub id: b256,
    pub publisher_merkle_root: b256,
    pub value_compute_alg_hash: b256,
    pub r: b256,
    pub s: b256,
    pub v: u8,
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
