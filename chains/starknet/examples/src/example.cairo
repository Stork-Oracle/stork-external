//! A contract that reads a Stork price feed.
//!
//! The Stork contract address is passed in here for clarity; a real contract would more likely
//! store it, or hard code it as a constant.

use stork_starknet_sdk::temporal_numeric_value::{EncodedAssetId, TemporalNumericValue};

#[starknet::interface]
pub trait IExample<TContractState> {
    /// Reads a feed, rejecting values older than the Stork contract's staleness window.
    fn use_stork_price(
        ref self: TContractState, stork_address: starknet::ContractAddress, id: EncodedAssetId,
    ) -> TemporalNumericValue;

    /// Reads a feed without the staleness check. Only do this if you handle staleness yourself.
    fn use_stork_price_unsafe(
        self: @TContractState, stork_address: starknet::ContractAddress, id: EncodedAssetId,
    ) -> TemporalNumericValue;
}

#[starknet::contract]
pub mod Example {
    use starknet::ContractAddress;
    use stork_starknet_sdk::interface::{IStorkDispatcher, IStorkDispatcherTrait};
    use stork_starknet_sdk::temporal_numeric_value::{EncodedAssetId, TemporalNumericValue};

    #[storage]
    struct Storage {}

    #[event]
    #[derive(Drop, starknet::Event)]
    pub enum Event {
        PriceUsed: PriceUsed,
    }

    #[derive(Drop, starknet::Event)]
    pub struct PriceUsed {
        #[key]
        pub id: EncodedAssetId,
        pub timestamp_ns: u64,
        pub quantized_value: i128,
    }

    #[abi(embed_v0)]
    impl ExampleImpl of super::IExample<ContractState> {
        fn use_stork_price(
            ref self: ContractState, stork_address: ContractAddress, id: EncodedAssetId,
        ) -> TemporalNumericValue {
            // Panics if the feed is unknown or older than the contract's valid time period, so a
            // consumer never acts on a stale price by accident.
            let value = IStorkDispatcher { contract_address: stork_address }
                .get_temporal_numeric_value_v1(id);

            self
                .emit(
                    PriceUsed {
                        id,
                        timestamp_ns: value.timestamp_ns,
                        quantized_value: value.quantized_value,
                    },
                );

            value
        }

        fn use_stork_price_unsafe(
            self: @ContractState, stork_address: ContractAddress, id: EncodedAssetId,
        ) -> TemporalNumericValue {
            IStorkDispatcher { contract_address: stork_address }
                .get_temporal_numeric_value_unsafe_v1(id)
        }
    }
}
