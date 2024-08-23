# Stage 1: Build the Rust library
FROM rust:1.70-buster AS rust-build

# Install cross-compilation tools
RUN apt-get update && apt-get install -y \
    gcc-aarch64-linux-gnu \
    libc6-dev-arm64-cross \
    build-essential \
    curl \
    git \
    bash

# Add the ARM64 target
RUN rustup target add aarch64-unknown-linux-gnu

ENV CARGO_TARGET_AARCH64_UNKNOWN_LINUX_GNU_LINKER=aarch64-linux-gnu-gcc

WORKDIR /app/rust/stork

# Copy the Rust source code into the container
COPY rust/stork .

# Install dependencies and build the Rust library for arm64
RUN cargo build --release --target aarch64-unknown-linux-gnu

# Stage 2: Build the Go Library
FROM golang:1.22-bullseye AS go-build

# Install cross-compilation tools
RUN apt-get update && apt-get install -y \
    gcc-aarch64-linux-gnu \
    libc6-dev-arm64-cross \
    build-essential

WORKDIR /app/cli

COPY cli/go.mod cli/go.sum ./
RUN go mod download

# Copy the source code from the cli directory into the container
COPY cli/ .

# Copy the Rust library from the rust-build stage
COPY rust/ /app/rust/

COPY --from=rust-build /app/rust/stork/target/aarch64-unknown-linux-gnu/release/libstork.so /usr/local/lib/
ENV LD_LIBRARY_PATH=/usr/local/lib

RUN CGO_ENABLED=1 GOOS=linux GOARCH=arm64 go build -o /stork .

# Stage 3: Create the final image
FROM debian:buster-slim
COPY --from=go-build /stork /usr/local/bin/stork
COPY --from=rust-build /app/rust/stork/target/aarch64-unknown-linux-gnu/release/libstork.so /usr/local/lib/
ENV LD_LIBRARY_PATH=/usr/local/lib

RUN apt-get update && apt-get install -y \
    libc6 \
    libpthread-stubs0-dev \
    ca-certificates
ENTRYPOINT ["/usr/local/bin/stork"]