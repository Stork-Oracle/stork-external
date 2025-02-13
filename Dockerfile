# Builder stage
FROM golang:1.22-bookworm AS builder
ARG TARGETPLATFORM
WORKDIR /app

# System dependencies layer
RUN apt-get update && apt-get install -y \
    gcc-aarch64-linux-gnu \
    libc6-dev-arm64-cross \
    g++-x86-64-linux-gnu \
    libc6-dev-amd64-cross \
    build-essential

# Rust installation layer
ENV RUST_VERSION=1.82.0
RUN curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | \
    sh -s -- -y --default-toolchain ${RUST_VERSION} --profile minimal
ENV PATH="/root/.cargo/bin:${PATH}"

# Pre-compile Rust shared library
COPY Makefile ./
COPY mk ./mk/
COPY . .
RUN make rust

# Go dependencies layer
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Source code layer
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    make install

# Release stage
FROM debian:bookworm-slim AS release
WORKDIR /app

RUN apt-get update && \
    apt-get install -y --no-install-recommends \
        ca-certificates \
        curl \
        jq \
        wget \
    && rm -rf /var/lib/apt/lists/* \
    && apt-get clean

# Install CosmWasm libraries
RUN wget https://github.com/CosmWasm/wasmvm/releases/download/v2.2.1/libwasmvm.aarch64.so -O /usr/lib/libwasmvm.aarch64.so && \
    wget https://github.com/CosmWasm/wasmvm/releases/download/v2.2.1/libwasmvm.x86_64.so -O /usr/lib/libwasmvm.x86_64.so && \
    cp "/usr/lib/libwasmvm.$(uname -m | sed 's/x86_64/x86_64/;s/aarch64/aarch64/').so" /usr/lib/libwasmvm.so

COPY --from=builder /app/.lib/libstork.so /usr/local/lib/
ENV LD_LIBRARY_PATH=/usr/local/lib:/usr/lib

# Ensure SERVICE is defined
ARG SERVICE
ENV SERVICE=${SERVICE}
RUN if [ -z "$SERVICE" ]; then echo "SERVICE argument is not defined"; exit 1; fi

COPY --from=builder /app/.bin/${SERVICE} /app/

COPY docker-entrypoint.sh /app/
COPY version.txt /app/
RUN chmod +x /app/docker-entrypoint.sh
ENTRYPOINT ["/app/docker-entrypoint.sh"]
CMD []
