# Define the Rust source directory
RUST_SRC_DIR ?= apps/lib/signer/rust/stork
FUEL_FFI_SRC_DIR ?= apps/lib/chain_pusher/contract_bindings/fuel/fuel_ffi
RUST_LIB_DIR:= $(CURDIR)/.lib

# Detect the host operating system
UNAME_S := $(shell uname -s)
UNAME_M := $(shell uname -m)

# Define targets based on the host operating system. We compile for Linux within
# docker, where we can respect the TARGETPLATFORM variable (set by docker itself
# during the build). On macOS, we just want to compile something that will run
# on the current system, so we infer the target from the host architecture.
# https://doc.rust-lang.org/rustc/platform-support.html
ifeq ($(UNAME_S),Darwin)
    STORK_LIB_NAME := libstork.dylib
    FUEL_FFI_LIB_NAME := libfuel_ffi.dylib
    ifeq ($(UNAME_M),x86_64)
        RUST_TARGET := x86_64-apple-darwin
    else
        RUST_TARGET := aarch64-apple-darwin
    endif
else ifeq ($(UNAME_S),Linux)
    STORK_LIB_NAME := libstork.so
    FUEL_FFI_LIB_NAME := libfuel_ffi.so
    ifeq ($(TARGETPLATFORM),linux/amd64)
        RUST_TARGET := x86_64-unknown-linux-gnu
    else ifeq ($(TARGETPLATFORM),linux/arm64)
        RUST_TARGET := aarch64-unknown-linux-gnu
    else
        $(error Unsupported TARGETPLATFORM: $(TARGETPLATFORM))
    endif
else
    $(error Unsupported operating system: $(UNAME_S))
endif

LIBSTORK_DIR := $(RUST_SRC_DIR)/target/$(RUST_TARGET)/release
LIBSTORK := $(RUST_LIB_DIR)/$(STORK_LIB_NAME)

FUEL_FFI_LIB_DIR := $(FUEL_FFI_SRC_DIR)/target/$(RUST_TARGET)/release
LIBFUEL_FFI := $(RUST_LIB_DIR)/$(FUEL_FFI_LIB_NAME)

# Add header file definitions
FUEL_FFI_HEADER := $(RUST_LIB_DIR)/fuel_ffi.h

# Find all Rust source files for proper dependencies
RUST_SOURCES := $(shell find $(RUST_SRC_DIR) -name "*.rs" 2>/dev/null || echo "")
FUEL_FFI_SOURCES := $(shell find $(FUEL_FFI_SRC_DIR) -name "*.rs" 2>/dev/null || echo "")

# Debug what find actually returns
$(info DEBUG: FUEL_FFI_SRC_DIR = $(FUEL_FFI_SRC_DIR))
$(info DEBUG: FUEL_FFI_SOURCES = $(FUEL_FFI_SOURCES))

FUEL_FFI_CARGO := $(FUEL_FFI_SRC_DIR)/Cargo.toml
FUEL_FFI_ABI := $(FUEL_FFI_SRC_DIR)/stork-abi.json

$(LIBSTORK): $(RUST_SOURCES)
	@echo "Building $(STORK_LIB_NAME) for $(RUST_TARGET)..."
	@mkdir -p $(RUST_LIB_DIR)
	@cd $(RUST_SRC_DIR) && \
	rustup target add $(RUST_TARGET) && \
	CARGO_NET_GIT_FETCH_WITH_CLI=false cargo build --release --target $(RUST_TARGET) && \
	cp target/$(RUST_TARGET)/release/$(STORK_LIB_NAME) $(LIBSTORK)

$(LIBFUEL_FFI): $(FUEL_FFI_SOURCES) $(FUEL_FFI_CARGO) $(FUEL_FFI_ABI)
	@echo "Building $(FUEL_FFI_LIB_NAME) for $(RUST_TARGET)..."
	@mkdir -p $(RUST_LIB_DIR)
	@cd $(FUEL_FFI_SRC_DIR) && \
	rustup target add $(RUST_TARGET) && \
	CARGO_NET_GIT_FETCH_WITH_CLI=false cargo build --release --target $(RUST_TARGET) && \
	cp target/$(RUST_TARGET)/release/$(FUEL_FFI_LIB_NAME) $(LIBFUEL_FFI) && \
	cp target/include/fuel_ffi.h $(FUEL_FFI_HEADER)

.PHONY: rust
rust: $(LIBSTORK) $(LIBFUEL_FFI)

.PHONY: clean-rust
clean-rust:
	@rm -rf $(RUST_LIB_DIR)
	@cd $(RUST_SRC_DIR) && cargo clean
	@cd $(FUEL_FFI_SRC_DIR) && cargo clean
