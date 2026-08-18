//! Verifies the example contract against a real Stork deployment, including the staleness check
//! a consumer relies on.

use snforge_std::{
    ContractClassTrait, DeclareResultTrait, declare, start_cheat_block_timestamp_global,
};
use starknet::{ContractAddress, EthAddress};
use stork_example::example::{IExampleDispatcher, IExampleDispatcherTrait};
use stork_starknet_sdk::interface::{
    IStorkDispatcher, IStorkDispatcherTrait, TemporalNumericValueInput,
};
use stork_starknet_sdk::temporal_numeric_value::TemporalNumericValue;

const VALID_TIME_PERIOD: u64 = 3600;

/// A real Stork-signed update, from apps/chain_pusher/internal/testutil/testdata.
fn signed_update() -> TemporalNumericValueInput {
    TemporalNumericValueInput {
        id: 0x4de9a89eed25754cfff794b5d7e8a71234cf930ee6bfb71ea8b8aa0ce313699f,
        temporal_numeric_value: TemporalNumericValue {
            timestamp_ns: 1757543080509034593, quantized_value: 100002000000000000000000,
        },
        publisher_merkle_root: 0x7dbca74780f2c8da0f9b17f3bae1280ab787900930c557b215701de7f83c43c0,
        value_compute_alg_hash: 0x9be7e9f9ed459417d96112a7467bd0b27575a2c7847195c68f805b70ce1795ba,
        r: 0x72604d29eb8cdb9009d51e1c2e482e4aa3929d3ac3bf7e9185ceec165edd4490,
        s: 0x1bcfeefd20bdc814c75c0745e1d546fd996cc59d6b3653c5779724c83dfe7e23,
        v: 0x1c,
    }
}

fn deploy() -> (IStorkDispatcher, IExampleDispatcher) {
    let stork_class = declare("Stork").unwrap().contract_class();
    let signer: EthAddress = 0xC4A02e7D370402F4afC36032076B05e74FF81786.try_into().unwrap();
    let owner: ContractAddress = 'owner'.try_into().unwrap();
    let no_token: ContractAddress = 0.try_into().unwrap();

    let mut calldata = array![];
    owner.serialize(ref calldata);
    signer.serialize(ref calldata);
    VALID_TIME_PERIOD.serialize(ref calldata);
    0_u256.serialize(ref calldata);
    no_token.serialize(ref calldata);

    let (stork_address, _) = stork_class.deploy(@calldata).unwrap();

    let example_class = declare("Example").unwrap().contract_class();
    let (example_address, _) = example_class.deploy(@array![]).unwrap();

    (
        IStorkDispatcher { contract_address: stork_address },
        IExampleDispatcher { contract_address: example_address },
    )
}

#[test]
fn test_example_reads_a_fresh_price() {
    let (stork, example) = deploy();
    let update = signed_update();
    stork.update_temporal_numeric_values_v1(array![update]);

    // One second after the value was signed.
    start_cheat_block_timestamp_global(1757543081);

    let value = example.use_stork_price(stork.contract_address, update.id);
    assert!(value == update.temporal_numeric_value);
}

#[test]
#[should_panic(expected: 'Stork: stale value')]
fn test_example_rejects_a_stale_price() {
    let (stork, example) = deploy();
    let update = signed_update();
    stork.update_temporal_numeric_values_v1(array![update]);

    start_cheat_block_timestamp_global(1757543081 + VALID_TIME_PERIOD + 1);

    example.use_stork_price(stork.contract_address, update.id);
}

#[test]
fn test_example_unsafe_read_ignores_staleness() {
    let (stork, example) = deploy();
    let update = signed_update();
    stork.update_temporal_numeric_values_v1(array![update]);

    start_cheat_block_timestamp_global(1757543081 + VALID_TIME_PERIOD + 1);

    let value = example.use_stork_price_unsafe(stork.contract_address, update.id);
    assert!(value == update.temporal_numeric_value);
}

#[test]
#[should_panic(expected: 'Stork: not found')]
fn test_example_rejects_an_unknown_feed() {
    let (stork, example) = deploy();

    example.use_stork_price_unsafe(stork.contract_address, 0xdeadbeef);
}
