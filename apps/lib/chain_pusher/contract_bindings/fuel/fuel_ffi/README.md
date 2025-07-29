# Fuel FFI Library

This library provides Foreign Function Interface (FFI) bindings for interacting with Fuel blockchain contracts from Go. It wraps the official Fuel Rust SDK (fuels v0.74.0) to provide a C-compatible interface that can be called from Go via CGO.

## Architecture

The library consists of:

- **Rust Library** (`src/lib.rs`): Core Fuel SDK interactions with C-compatible exports
- **C Header** (`target/include/fuel_ffi.h`): C function declarations for FFI interface that are auto generated when building
- **Go Bindings** (`../stork_fuel_contract.go`): Go wrapper that calls FFI functions via CGO

## Building

To build the library:

```bash
# Build release version
cargo build --release
```

This generates the dynamic library at `target/release/libfuel_ffi.<so|dylib|dll>`

