module example::example {

    use stork::stork;
    use aptos_framework::event;

    #[event]
    public struct ExampleStorkPriceEvent has copy, drop, store {
        timestamp: u64,
        magnitude: u128,
        negative: bool,
    }

    public entry fun use_stork_price(asset_id: vector<u8>) {
        // Get the price
        let price = stork::get_temporal_numeric_value_unchecked(asset_id);

        let timestamp = price.get_timestamp_ns();
        let i128value = price.get_quantized_value();

        let magnitude = i128value.get_magnitude();
        let negative = i128value.is_negative();

        // Do something with the price
        // example: emit an event
        event::emit(ExampleStorkPriceEvent {
            timestamp,
            magnitude,
            negative,
        });
    }
}