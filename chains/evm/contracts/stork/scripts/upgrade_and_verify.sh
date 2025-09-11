#!/bin/bash

# Enable error logging
exec 2> upgrade_error.log

# Array of networks
networks=(
    inMemoryNode
)

# Contract address to verify
CONTRACT_ADDRESS="0xacC0a0cF13571d30B4b8637996F5D6D774d4fd62"

# Log file for successful operations
SUCCESS_LOG="upgrade_success.log"
touch "$SUCCESS_LOG"

for network in "${networks[@]}"; do
    echo "Processing network: $network"
    
    echo "Upgrading contract on $network..."
    if npx hardhat --network "$network" upgrade; then
        echo "[SUCCESS] Upgrade completed on $network" >> "$SUCCESS_LOG"
    else
        echo "[ERROR] Upgrade failed on $network" >&2
        continue
    fi
    
    echo "Verifying contract on $network..."
    if npx hardhat --network "$network" verify "$CONTRACT_ADDRESS"; then
        echo "[SUCCESS] Verification completed on $network" >> "$SUCCESS_LOG"
    else
        echo "[ERROR] Verification failed on $network" >&2
    fi
    
    echo "Completed processing for $network"
    echo "Version is now $(npx hardhat --network $network interact version)"
    echo "----------------------------------------"
done

echo "All networks processed! Check upgrade_success.log for successful operations and upgrade_error.log for errors." 
