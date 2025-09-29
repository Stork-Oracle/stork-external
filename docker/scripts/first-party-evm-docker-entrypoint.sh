#!/bin/sh

# Change to the correct directory
cd /usr/src/app;

# Using a fake private key for local development
npx hardhat vars set PRIVATE_KEY 94c7f304055f90fb58c51e2bb7faa9c46348bacc6035d20aa10bf98e89a4b2c0

# Start hardhat node as a background process
npx hardhat node &

# Compile contracts
npx hardhat compile

# Wait for hardhat node to initialize and then deploy contracts
npx wait-on http://127.0.0.1:8545 && echo "Deploying First Party Stork contract..." && npx hardhat ignition deploy ignition/modules/FirstPartyStork.ts --network hardhatLocal --reset

# Register the publisher after deployment
echo "Registering publisher with deployed contract..."

# Get the deployed contract address
DEPLOYED_ADDRESS=$(cat ignition/deployments/chain-31337/deployed_addresses.json | grep -o '"FirstPartyStorkModule#UpgradeableFirstPartyStork": "[^"]*"' | cut -d'"' -f4)
echo "Contract deployed at: $DEPLOYED_ADDRESS"

# Register publisher using hardhat console
# This calls createPublisherUser(address publisher, uint256 fee) with fee=0
npx hardhat console --network hardhatLocal << EOF
const contract = await ethers.getContractAt("UpgradeableFirstPartyStork", "$DEPLOYED_ADDRESS");
const [owner] = await ethers.getSigners();
const publisherAddress = "0x99e295e85cb07C16B7BB62A44dF532A7F2620237";
const fee = 0;
console.log("Registering publisher:", publisherAddress);
const tx = await contract.createPublisherUser(publisherAddress, fee);
await tx.wait();
console.log("Publisher registered successfully! Tx:", tx.hash);
EOF
  
echo "Publisher registration complete"

# The hardhat node process never completes
# Waiting prevents the container from pausing
wait $!
