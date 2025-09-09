FROM node:20-alpine

# Copy in contract and sdk, preserving relative paths
COPY chains/evm/contracts/stork /usr/src/app
COPY chains/evm/sdks/stork_evm_sdk /usr/sdks/stork_evm_sdk

WORKDIR /usr/src/app

ARG STORK_PUBLIC_KEY
ENV STORK_PUBLIC_KEY=${STORK_PUBLIC_KEY}


RUN npm install
RUN npm install -g wait-on

# Create the Hardhat config directory
RUN mkdir -p /root/.config/hardhat-nodejs

# Copy the entrypoint script
COPY docker/scripts/evm-docker-entrypoint.sh /usr/src/app/docker-entrypoint.sh
RUN chmod +x /usr/src/app/docker-entrypoint.sh

# Just run the entrypoint script directly - it already handles everything
ENTRYPOINT [ "/bin/sh", "-c", "/usr/src/app/docker-entrypoint.sh" ]
