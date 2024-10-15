# Stage 1: Build the Rust binary
FROM rust:1.80-bookworm
# Install cross-compilation tools
RUN apt-get update &&\
    apt-get install -y \
    gcc-aarch64-linux-gnu \
    libc6-dev-arm64-cross \
    g++-x86-64-linux-gnu libc6-dev-amd64-cross \
    build-essential \
    curl \
    git \
    bash


ENV CARGO_TARGET_AARCH64_UNKNOWN_LINUX_GNU_LINKER=aarch64-linux-gnu-gcc
ENV CARGO_TARGET_X86_64_UNKNOWN_LINUX_GNU_LINKER=x86_64-linux-gnu-gcc

WORKDIR /app/rust/stork

# Copy the Rust source code into the container
COPY apps/lib/signer/rust/stork .

# Install dependencies and build the Rust library depending on the target platform
ARG TARGETPLATFORM
RUN case "$TARGETPLATFORM" in \
        "linux/amd64")  export TARGET="x86_64-unknown-linux-gnu" ;; \
        "linux/arm64")  export TARGET="aarch64-unknown-linux-gnu" ;; \
        *) echo "Unsupported platform: $TARGETPLATFORM" ; exit 1 ;; \
    esac && \
    rustup target add $TARGET && \
    cargo build --release --target $TARGET && \
    cp /app/rust/stork/target/$TARGET/release/libstork.so /usr/local/lib/