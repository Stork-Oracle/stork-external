# Stage 1: Build the Rust binary
FROM rust:1.80-bookworm AS rust-build
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

# Stage 2: Build the Go Library
FROM golang:1.22-bookworm AS go-build

# Install cross-compilation tools
RUN apt-get update && apt-get install -y \
    gcc-aarch64-linux-gnu \
    libc6-dev-arm64-cross \
    g++-x86-64-linux-gnu libc6-dev-amd64-cross \
    build-essential

WORKDIR /app/cli

COPY go.mod go.sum ./
RUN go mod download

# Copy the source code from the lib and cmd directories into the container
COPY apps/lib ./lib/
COPY apps/cmd/publisher_agent ./cmd/publisher_agent/

COPY --from=rust-build /usr/local/lib/libstork.so /usr/local/lib/
ENV LD_LIBRARY_PATH=/usr/local/lib

ARG TARGETPLATFORM
RUN case "$TARGETPLATFORM" in \
        "linux/amd64")  export GOARCH="amd64" && export CC="x86_64-linux-gnu-gcc";; \
        "linux/arm64")  export GOARCH="arm64" && export CC="aarch64-linux-gnu-gcc";; \
        *) echo "Unsupported platform: $TARGETPLATFORM" ; exit 1 ;; \
    esac && \
    CGO_ENABLED=1 GOOS=linux GOARCH=$GOARCH CC=$CC go build -o /stork cmd/publisher_agent/main.go

# Stage 3: Create the final image
FROM debian:bookworm-slim

COPY --from=go-build /stork /usr/local/bin/stork-publisher-agent
COPY --from=rust-build /usr/local/lib/libstork.so /usr/local/lib/
ENV LD_LIBRARY_PATH=/usr/local/lib

RUN apt-get update && apt-get install -y \
    libc6 \
    libpthread-stubs0-dev \
    ca-certificates

ENTRYPOINT ["/usr/local/bin/stork-publisher-agent"]
