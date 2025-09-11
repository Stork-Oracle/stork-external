contract;

use stork_sway_sdk::interface::Stork;
use std::logging::log;
use signed_int::i128::I128;

struct ExampleStorkPriceEvent {
    timestamp: u64,
    quantized_value: I128,
}

abi Example {
    fn use_stork_price(feed_id: b256, stork_contract_address: b256);
}

impl Example for Contract {
    // Pass in the contract address as a b256, this could also be stored in storage, or written as a constant in the contract
    fn use_stork_price(feed_id: b256, stork_contract_address: b256) {
        // Get the stork contract
        let stork_contract = abi(Stork, stork_contract_address);

        // Get the price
        let price = stork_contract.get_temporal_numeric_value_unchecked_v1(feed_id);

        let timestamp = price.get_timestamp_ns();
        let quantized_value = price.get_quantized_value();

        // Do something with the price
        // example: Log the price

        log(ExampleStorkPriceEvent {
            timestamp,
            quantized_value,
        });
    }
}
