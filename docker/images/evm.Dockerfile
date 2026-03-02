FROM node:20-alpine

# Copy in contract and sdk, preserving relative paths
COPY chains/evm/contracts/stork /usr/src/app
COPY chains/evm/sdks/stork_evm_sdk /usr/sdks/stork_evm_sdk

WORKDIR /usr/src/app

ARG STORK_PUBLIC_KEY
ENV STORK_PUBLIC_KEY=${STORK_PUBLIC_KEY}


RUN npm ci
RUN npm install -g wait-on

# Create the Hardhat config directory
RUN mkdir -p /root/.config/hardhat-nodejs

# Pre-seed Hardhat vars and compile at image build time so runtime doesn't need outbound access
# (Hardhat config uses `vars.get(...)`, which requires these to exist.)
RUN npx hardhat vars set PRIVATE_KEY 94c7f304055f90fb58c51e2bb7faa9c46348bacc6035d20aa10bf98e89a4b2c0 && \
    npx hardhat vars set ARBISCAN_API_KEY fake && \
    npx hardhat vars set POLYGON_API_KEY fake && \
    npx hardhat vars set ETHERSCAN_API_KEY fake && \
    npx hardhat vars set CORE_TESTNET_API_KEY fake && \
    npx hardhat vars set CORE_MAINNET_API_KEY fake && \
    npx hardhat vars set ROOTSTOCK_TESTNET_API_KEY fake && \
    npx hardhat vars set SCROLL_MAINNET_API_KEY fake && \
    npx hardhat vars set SONEIUM_MAINNET_RPC_URL fake && \
    npx hardhat vars set ETHERSCAN_SOPHON_API_KEY fake && \
    npx hardhat vars set SONEIUM_MAINNET_BLOCKSCOUT_URL fake && \
    npx hardhat vars set CRONOS_L2_API_KEY fake && \
    npx hardhat compile

# Copy the entrypoint script
COPY docker/scripts/evm-docker-entrypoint.sh /usr/src/app/docker-entrypoint.sh
RUN chmod +x /usr/src/app/docker-entrypoint.sh

# Just run the entrypoint script directly - it already handles everything
ENTRYPOINT [ "/bin/sh", "-c", "/usr/src/app/docker-entrypoint.sh" ]
