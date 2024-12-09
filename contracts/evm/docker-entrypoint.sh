#!/bin/sh

# Change to the correct directory
cd /usr/src/app;

# this is a fake private key
npx hardhat vars set PRIVATE_KEY 94c7f304055f90fb58c51e2bb7faa9c46348bacc6035d20aa10bf98e89a4b2c0
npx hardhat vars set ARBISCAN_API_KEY fake
npx hardhat vars set POLYGON_API_KEY fake
npx hardhat vars set ETHERSCAN_API_KEY fake
npx hardhat vars set CORE_TESTNET_API_KEY fake
npx hardhat vars set CORE_MAINNET_API_KEY fake
npx hardhat vars set ROOTSTOCK_TESTNET_API_KEY fake
npx hardhat vars set SONEIUM_MAINNET_RPC_URL fake
npx hardhat vars set SONEIUM_MAINNET_BLOCKSCOUT_URL fake

# Start hardhat node as a background process
npx hardhat node &

npx hardhat compile

# Wait for hardhat node to initialize and then deploy contracts
npx wait-on http://127.0.0.1:8545 && npx hardhat --network inMemoryNode deploy $STORK_PUBLIC_KEY

# The hardhat node process never completes
# Waiting prevents the container from pausing
wait $!
