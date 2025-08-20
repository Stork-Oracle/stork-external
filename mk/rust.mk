# Rust workspace build configuration
WORKSPACE_ROOT := $(CURDIR)
RUST_TARGET_DIR := $(WORKSPACE_ROOT)/target
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
SIGNER_HEADER_SRC := $(RUST_TARGET_DIR)/include/signer_ffi.h
FUEL_HEADER_SRC := $(RUST_TARGET_DIR)/include/fuel_ffi.h
SIGNER_HEADER_DEST := $(RUST_LIB_DIR)/signer_ffi.h
FUEL_HEADER_DEST := $(RUST_LIB_DIR)/fuel_ffi.h

.PHONY: build-signer-ffi
build-signer-ffi:
	@echo "Building signer_ffi..."
	@cargo build --release -p signer_ffi

.PHONY: build-fuel-ffi
build-fuel-ffi:
	@echo "Building fuel_ffi..."
	@cargo build --release -p fuel_ffi

# Copy artifacts to lib directory
$(SIGNER_LIB_DEST): build-signer-ffi
	@echo "Copying signer_ffi to $(RUST_LIB_DIR)..."
	@mkdir -p $(RUST_LIB_DIR)
	@cp $(SIGNER_LIB_SRC) $(SIGNER_LIB_DEST)

$(FUEL_LIB_DEST): build-fuel-ffi
	@echo "Copying fuel_ffi to $(RUST_LIB_DIR)..."
	@mkdir -p $(RUST_LIB_DIR)
	@cp $(FUEL_LIB_SRC) $(FUEL_LIB_DEST)

$(SIGNER_HEADER_DEST): build-signer-ffi
	@mkdir -p $(RUST_LIB_DIR)
	@cp $(SIGNER_HEADER_SRC) $(SIGNER_HEADER_DEST)

$(FUEL_HEADER_DEST): build-fuel-ffi
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
	@echo "Cleaning copied libraries..."
	@rm -rf $(RUST_LIB_DIR)


# Lint rust code
.PHONY: lint-rust
lint-rust:
	@echo "Linting Rust workspace..."
	@cargo clippy --all-targets --all-features -- -D warnings

# Format rust code
.PHONY: format-rust
format-rust:
	@echo "Formatting Rust workspace..."
	@cargo fmt --all
