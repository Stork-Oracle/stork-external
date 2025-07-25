#!/bin/bash

# Build the Fuel FFI library for Go integration
set -e

echo "Building Fuel FFI library..."

# Build the Rust library
cargo build --release

echo "Fuel FFI library built successfully!"
echo "Library location: target/release/libfuel_ffi.a (or .so on Linux)"
echo ""
echo "To use this library with Go:"
echo "1. Run: go build in the chain_pusher directory"
echo "2. The CGO flags in fuel.go will automatically link to this library"