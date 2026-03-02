FROM node:22-alpine

# Copy in first party stork contract
COPY chains/evm/contracts/first_party_stork /usr/src/app
COPY chains/evm/sdks/first_party_stork_evm_sdk /usr/sdks/first_party_stork_evm_sdk
COPY chains/evm/sdks/stork_evm_sdk /usr/sdks/stork_evm_sdk

WORKDIR /usr/sdks/first_party_stork_evm_sdk
RUN npm install

WORKDIR /usr/src/app

ARG PRIVATE_KEY
ARG CONTRACT_ADDRESS
ARG RPC_URL
ARG PUBLISHER_EVM_PUBLIC_KEY

ENV PRIVATE_KEY=${PRIVATE_KEY}
ENV CONTRACT_ADDRESS=${CONTRACT_ADDRESS}
ENV RPC_URL=${RPC_URL}
ENV PUBLISHER_EVM_PUBLIC_KEY=${PUBLISHER_EVM_PUBLIC_KEY}

# Install system dependencies for healthcheck
RUN apk add --no-cache wget

RUN npm install
RUN npm install -g wait-on

# Create the Hardhat config directory
RUN mkdir -p /root/.config/hardhat-nodejs

# Copy the entrypoint script for first party stork
COPY docker/scripts/first-party-evm-docker-entrypoint.sh /usr/src/app/docker-entrypoint.sh
RUN chmod +x /usr/src/app/docker-entrypoint.sh

# Just run the entrypoint script directly - it already handles everything
ENTRYPOINT [ "/bin/sh", "-c", "/usr/src/app/docker-entrypoint.sh" ]
