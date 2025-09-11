# Builder stage
FROM golang:1.24.3-bookworm AS builder
ARG TARGETPLATFORM
WORKDIR /app

# System dependencies layer
RUN apt-get update && apt-get install -y \
    gcc-aarch64-linux-gnu \
    libc6-dev-arm64-cross \
    g++-x86-64-linux-gnu \
    libc6-dev-amd64-cross \
    build-essential


# mount cache for incremental rust builds



# Rust installation layer
ENV RUST_VERSION=1.88.0
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
