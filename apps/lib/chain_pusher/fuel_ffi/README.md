# Fuel FFI Library

This library provides Foreign Function Interface (FFI) bindings for interacting with Fuel blockchain contracts from Go. It wraps the official Fuel Rust SDK (fuels v0.70.4) to provide a C-compatible interface that can be called from Go via CGO.

## Architecture

The library consists of:

- **Rust Library** (`src/lib.rs`): Core Fuel SDK interactions with C-compatible exports
- **C Header** (`src/fuel.h`): C function declarations for FFI interface  
- **Go Bindings** (`../fuel.go`): Go wrapper that calls FFI functions via CGO

## Building

To build the library:

```bash
# Build release version
cargo build --release

# Or use the convenience script
./build.sh
```

This generates the static library at `target/release/libfuel_ffi.a` (or `.so` on Linux).

## API

### Client Management
- `fuel_client_new(config_json)` - Create new Fuel client from JSON config
- `fuel_client_free(client)` - Clean up client resources

### Contract Interactions  
- `fuel_get_latest_value(client, id)` - Get latest temporal numeric value for asset ID
- `fuel_update_values(client, inputs_json)` - Batch update multiple values on contract
- `fuel_get_wallet_balance(client)` - Get wallet balance in base units

### Memory Management
- `fuel_free_string(s)` - Free strings returned by FFI functions

## Configuration

Client configuration is passed as JSON:

```json
{
  "rpc_url": "https://testnet.fuel.network/graphql",
  "contract_address": "0x1234567890abcdef...",
  "private_key": "0x1234567890abcdef..."
}
```

## Error Handling

FFI functions return null pointers on error. The Go wrapper handles error checking and provides proper Go error types.

## Development Notes

- Uses tokio runtime for async operations
- Handles JSON serialization/deserialization for complex types
- Manages C string memory allocation/deallocation
- Thread-safe through Arc<Runtime> for async operations

## Dependencies

- `fuels = "0.70.4"` - Official Fuel Rust SDK
- `tokio` - Async runtime
- `serde` + `serde_json` - JSON serialization
- `hex` - Hex string utilities
- `libc` - C FFI types