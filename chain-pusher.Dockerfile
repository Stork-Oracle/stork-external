# Stage 1: Build the Go Library
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
COPY apps/cmd/chain_pusher ./cmd/chain_pusher/

ARG TARGETPLATFORM
RUN case "$TARGETPLATFORM" in \
        "linux/amd64")  export GOARCH="amd64" && export CC="x86_64-linux-gnu-gcc";; \
        "linux/arm64")  export GOARCH="arm64" && export CC="aarch64-linux-gnu-gcc";; \
        *) echo "Unsupported platform: $TARGETPLATFORM" ; exit 1 ;; \
    esac && \
    CGO_ENABLED=1 GOOS=linux GOARCH=$GOARCH CC=$CC go build -o /stork cmd/chain_pusher/main.go

# Stage 3: Create the final image
FROM debian:bookworm-slim

COPY --from=go-build /stork /usr/local/bin/stork-chain-pusher
ENV LD_LIBRARY_PATH=/usr/local/lib

RUN apt-get update && apt-get install -y \
    libc6 \
    libpthread-stubs0-dev \
    ca-certificates

ENTRYPOINT ["/usr/local/bin/stork-chain-pusher"]
