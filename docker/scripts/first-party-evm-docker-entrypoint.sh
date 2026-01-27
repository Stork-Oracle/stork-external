#!/bin/sh

# Change to the correct directory
cd /usr/src/app;

# Clean any existing log file and start hardhat node as a background process
rm -f /tmp/registered.log
npx hardhat node &

npx hardhat compile

echo "Waiting for hardhat node to initialize..."
npx wait-on $RPC_URL
echo "Hardhat node initialized"
# Wait for hardhat node to initialize and then deploy contracts
echo "Deploying contracts..."
npx hardhat ignition deploy ignition/modules/FirstPartyStork.ts --network hardhatLocal --reset
echo "Contracts deployed"

# Run the TypeScript registration script and log success
npx ts-node scripts/local_register_publisher.ts && echo "REGISTRATION_COMPLETE" >> /tmp/registered.log

# The hardhat node process never completes
# Waiting prevents the container from pausing
wait $!
