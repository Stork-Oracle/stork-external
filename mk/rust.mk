# Define the Rust source directory
RUST_SRC_DIR ?= apps/lib/signer/rust/stork
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
    LIB_NAME := libstork.dylib
    ifeq ($(UNAME_M),x86_64)
        RUST_TARGET := x86_64-apple-darwin
    else
        RUST_TARGET := aarch64-apple-darwin
    endif
else ifeq ($(UNAME_S),Linux)
    LIB_NAME := libstork.so
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
LIBSTORK := $(RUST_LIB_DIR)/$(LIB_NAME)

$(LIBSTORK): $(RUST_SRC_DIR)
	@echo "Building $(LIB_NAME) for $(RUST_TARGET)..."
	@mkdir -p $(RUST_LIB_DIR)
	@cd $(RUST_SRC_DIR) && \
	rustup target add $(RUST_TARGET) && \
	CARGO_NET_GIT_FETCH_WITH_CLI=false cargo build --release --target $(RUST_TARGET) && \
	cp target/$(RUST_TARGET)/release/$(LIB_NAME) $(LIBSTORK)

.PHONY: rust
rust: $(LIBSTORK)

.PHONY: clean-rust
clean-rust:
	@rm -rf $(RUST_LIB_DIR)
	@cd $(RUST_SRC_DIR) && cargo clean
