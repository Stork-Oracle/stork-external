#!/bin/sh

# Change to the correct directory
cd /usr/src/app

# Set environment variables (example, could be in a .env file)
export PRIVATE_KEY="94c7f304055f90fb58c51e2bb7faa9c46348bacc6035d20aa10bf98e89a4b2c0"
export ARBISCAN_API_KEY="fake"
export POLYGON_API_KEY="fake"
export ETHERSCAN_API_KEY="fake"
export CORE_TESTNET_API_KEY="fake"
export CORE_MAINNET_API_KEY="fake"
export ROOTSTOCK_TESTNET_API_KEY="fake"
export SCROLL_MAINNET_API_KEY="fake"
export SONEIUM_MAINNET_RPC_URL="fake"
export SONEIUM_MAINNET_BLOCKSCOUT_URL="fake"
export CRONOS_L2_API_KEY="fake"

# Start hardhat node as a background process
npx hardhat node &

# Compile contracts
npx hardhat compile

# Wait for hardhat node to initialize and then deploy contracts
npx hardhat --network inMemoryNode deploy $STORK_PUBLIC_KEY

# The hardhat node process never completes
# Waiting prevents the container from pausing
wait $!
