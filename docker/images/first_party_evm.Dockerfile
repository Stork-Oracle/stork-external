FROM node:22-alpine

# Copy in first party stork contract
COPY chains/evm/contracts/first_party_stork /usr/src/app

WORKDIR /usr/src/app

ARG STORK_PUBLIC_KEY
ENV STORK_PUBLIC_KEY=${STORK_PUBLIC_KEY}

# TODO: see if i need this
ARG CONTRACT_OWNER_ADDRESS
ENV CONTRACT_OWNER_ADDRESS=${CONTRACT_OWNER_ADDRESS}

# Install system dependencies for healthcheck
RUN apk add --no-cache wget

RUN npm install
RUN npm install -g wait-on

# Create the Hardhat config directory
RUN mkdir -p /root/.config/hardhat-nodejs

# Copy the entrypoint script for first party stork
COPY docker/scripts/first-party-evm-docker-entrypoint.sh /usr/src/app/docker-entrypoint.sh
RUN chmod +x /usr/src/app/docker-entrypoint.sh

ENTRYPOINT [ "/bin/sh", "-c", "/usr/src/app/docker-entrypoint.sh" ]
