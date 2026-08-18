# Stork Starknet SDK

Types and interface for reading Stork price feeds from a Cairo contract.

## Usage

Add it to your `Scarb.toml`:

```toml
[dependencies]
stork_starknet_sdk = { path = "path/to/chains/starknet/sdks/stork_starknet_sdk" }
```

Then call the Stork contract through the dispatcher:

```cairo
use stork_starknet_sdk::interface::{IStorkDispatcher, IStorkDispatcherTrait};
use stork_starknet_sdk::temporal_numeric_value::TemporalNumericValue;

let value: TemporalNumericValue = IStorkDispatcher { contract_address: stork_address }
    .get_temporal_numeric_value_v1(encoded_asset_id);
```

See [the example contract](../../examples/src/example.cairo) for a working consumer.

## Reading values

| Call | Behaviour |
| --- | --- |
| `get_temporal_numeric_value_v1` | Panics if the feed is missing or older than the contract's staleness window. Prefer this. |
| `get_temporal_numeric_value_unsafe_v1` | Panics only if the feed is missing. Use when you handle staleness yourself. |
| `get_temporal_numeric_values_unsafe_v1` | Batch form of the above; panics if *any* id is missing. |
| `get_multiple_temporal_numeric_values_unchecked` | Batch read that never panics; unknown feeds come back with `timestamp_ns == 0`. |

`TemporalNumericValue.quantized_value` is an `i128` quantised to 18 decimals, and can be negative.
It is stored packed into a single `felt252` alongside the timestamp, so a feed update costs one
storage write.

## Modules

| Module | Contents |
| --- | --- |
| `interface` | `IStork`, `TemporalNumericValueInput`, and a minimal `IERC20` for fees |
| `temporal_numeric_value` | `TemporalNumericValue`, `EncodedAssetId`, and the storage packing |
| `events` | Event structs, including `ValueUpdate` |
| `errors` | The panic messages the contract raises |
