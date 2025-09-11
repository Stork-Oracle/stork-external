module example::example{
    use stork::state::StorkState;
    use stork::stork::get_temporal_numeric_value_unchecked;
    use sui::event;

    public struct ExampleStorkPriceEvent has copy, drop {
        timestamp: u64,
        magnitude: u128,
        negative: bool,
    }

    public fun use_stork_price(feed_id: vector<u8>, stork_state: &StorkState) {
        // Get the price
        let price = get_temporal_numeric_value_unchecked(stork_state, feed_id);
        
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