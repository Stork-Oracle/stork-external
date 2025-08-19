# Rust workspace build configuration
WORKSPACE_ROOT := $(CURDIR)
RUST_TARGET_DIR := $(WORKSPACE_ROOT)/target
RUST_INCLUDE_DIR := $(RUST_TARGET_DIR)/include
RUST_LIB_DIR := $(CURDIR)/.lib

# Detect platform and set library extensions
UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Darwin)
    LIB_EXT := dylib
else ifeq ($(UNAME_S),Linux)
    LIB_EXT := so
else
    $(error Unsupported operating system: $(UNAME_S))
endif

# Define library names and paths
SIGNER_LIB_NAME := libsigner_ffi.$(LIB_EXT)
FUEL_LIB_NAME := libfuel_ffi.$(LIB_EXT)

SIGNER_LIB_SRC := $(RUST_TARGET_DIR)/release/$(SIGNER_LIB_NAME)
FUEL_LIB_SRC := $(RUST_TARGET_DIR)/release/$(FUEL_LIB_NAME)

SIGNER_LIB_DEST := $(RUST_LIB_DIR)/$(SIGNER_LIB_NAME)
FUEL_LIB_DEST := $(RUST_LIB_DIR)/$(FUEL_LIB_NAME)

# Header files
SIGNER_HEADER_SRC := $(RUST_INCLUDE_DIR)/signer_ffi.h
FUEL_HEADER_SRC := $(RUST_INCLUDE_DIR)/fuel_ffi.h
SIGNER_HEADER_DEST := $(RUST_LIB_DIR)/signer_ffi.h
FUEL_HEADER_DEST := $(RUST_LIB_DIR)/fuel_ffi.h

# Find workspace members for dependency tracking
WORKSPACE_SOURCES := $(shell find shared/signer/signer_ffi apps/chain_pusher/lib/contract_bindings/fuel/fuel_ffi -name "*.rs" -o -name "Cargo.toml" -o -name "*.json" 2>/dev/null)

# Build all Rust libraries using workspace
$(SIGNER_LIB_SRC) $(FUEL_LIB_SRC): $(WORKSPACE_SOURCES) Cargo.toml
	@echo "Building Rust workspace..."
	@cargo build --release

$(SIGNER_LIB_DEST): $(SIGNER_LIB_SRC)
	@echo "Copying $(SIGNER_LIB_NAME)..."
	@mkdir -p $(RUST_LIB_DIR)
	@cp $(SIGNER_LIB_SRC) $(SIGNER_LIB_DEST)

$(FUEL_LIB_DEST): $(FUEL_LIB_SRC)
	@echo "Copying $(FUEL_LIB_NAME)..."
	@mkdir -p $(RUST_LIB_DIR)
	@cp $(FUEL_LIB_SRC) $(FUEL_LIB_DEST)

$(SIGNER_HEADER_DEST): $(SIGNER_HEADER_SRC)
	@echo "Copying signer_ffi.h..."
	@mkdir -p $(RUST_LIB_DIR)
	@cp $(SIGNER_HEADER_SRC) $(SIGNER_HEADER_DEST)

$(FUEL_HEADER_DEST): $(FUEL_HEADER_SRC)
	@echo "Copying fuel_ffi.h..."
	@mkdir -p $(RUST_LIB_DIR)
	@cp $(FUEL_HEADER_SRC) $(FUEL_HEADER_DEST)

# Main target
.PHONY: rust
rust: $(SIGNER_LIB_DEST) $(FUEL_LIB_DEST) $(SIGNER_HEADER_DEST) $(FUEL_HEADER_DEST)

# Clean targets
.PHONY: clean-rust
clean-rust:
	@echo "Cleaning Rust workspace..."
	@rm -rf $(RUST_LIB_DIR)
	@cargo clean

.PHONY: clean-rust-libs
clean-rust-libs:
	@echo "Cleaning copied libraries..."
	@rm -rf $(RUST_LIB_DIR)
