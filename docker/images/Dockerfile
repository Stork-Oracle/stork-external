# Rust base stage
FROM lukemathwalker/cargo-chef:latest-rust-1.88.0-bookworm AS rust-base

# Install precompiled sccache based on target architecture
ARG TARGETARCH
ARG SCCACHE_VERSION=0.7.7

RUN case "$TARGETARCH" in \
      "amd64") ARCH="x86_64" ;; \
      "arm64") ARCH="aarch64" ;; \
      *) echo "Unsupported architecture: $TARGETARCH" && exit 1 ;; \
    esac && \
    curl -L "https://github.com/mozilla/sccache/releases/download/v${SCCACHE_VERSION}/sccache-v${SCCACHE_VERSION}-${ARCH}-unknown-linux-musl.tar.gz" | \
    tar xz -C /usr/local/bin --strip-components=1 "sccache-v${SCCACHE_VERSION}-${ARCH}-unknown-linux-musl/sccache" && \
    chmod +x /usr/local/bin/sccache

# Set sccache environment variables
ENV RUSTC_WRAPPER=sccache SCCACHE_DIR=/sccache

# Rust cargo-chef planner stage
FROM rust-base AS rust-planner
WORKDIR /app
COPY . .
RUN cargo chef prepare --recipe-path recipe.json

# Rust builder stage
FROM rust-base AS rust-builder
WORKDIR /app
COPY --from=rust-planner /app/recipe.json recipe.json
RUN --mount=type=cache,target=/usr/local/cargo/registry \
    --mount=type=cache,target=/usr/local/cargo/git \
    --mount=type=cache,target=$SCCACHE_DIR,sharing=locked \
    cargo chef cook --release --recipe-path recipe.json
COPY . .
RUN --mount=type=cache,target=/usr/local/cargo/registry \
    --mount=type=cache,target=/usr/local/cargo/git \
    --mount=type=cache,target=$SCCACHE_DIR,sharing=locked \
    make rust

# Go builder stage
FROM golang:1.24.3-bookworm AS builder
ARG TARGETPLATFORM
WORKDIR /app

# System dependencies
RUN apt-get update && apt-get install -y \
    gcc-aarch64-linux-gnu \
    libc6-dev-arm64-cross \
    g++-x86-64-linux-gnu \
    libc6-dev-amd64-cross \
    build-essential

# Copy Rust libraries from rust-builder
COPY --from=rust-builder /app/.lib/ ./.lib/

# Go dependencies
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Source code
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    make install NO_RUST=1

# Release stage
FROM debian:bookworm-slim AS release
WORKDIR /app

RUN apt-get update && \
    apt-get install -y --no-install-recommends \
        ca-certificates \
        curl \
        jq \
    && rm -rf /var/lib/apt/lists/* \
    && apt-get clean

# Copy libraries from builder stage
COPY --from=builder /app/.lib/libwasmvm.*.so /usr/local/lib/
COPY --from=builder /app/.lib/libsigner_ffi.so /usr/local/lib/
COPY --from=builder /app/.lib/libfuel_ffi.so /usr/local/lib/
ENV LD_LIBRARY_PATH=/usr/local/lib

# Ensure SERVICE is defined
ARG SERVICE
ENV SERVICE=${SERVICE}
RUN if [ -z "$SERVICE" ]; then echo "SERVICE argument is not defined"; exit 1; fi

COPY --from=builder /app/.bin/${SERVICE} /app/

COPY docker/scripts/docker-entrypoint.sh /app/
COPY version.txt /app/
RUN chmod +x /app/docker-entrypoint.sh
ENTRYPOINT ["/app/docker-entrypoint.sh"]
CMD []
