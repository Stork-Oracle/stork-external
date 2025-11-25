#!/bin/bash

# Get the auth token from an arg
FEEDS=$1
if [ -z "$FEEDS" ]; then
  echo "Feeds are required"
  exit 1
fi

AUTH_TOKEN=$2
if [ -z "$AUTH_TOKEN" ]; then
  echo "Auth token is required"
  exit 1
fi

SLEEP_INTERVAL=${3:-60}

# Loop indefinitely
while true; do
  # Capture the start time
  start_time=$(date +%s)

  # Run the command and capture output
  output=$(npx hardhat --network sophonMainnet interact updateTemporalNumericValuesV1 $FEEDS https://rest.jp.stork-oracle.network $AUTH_TOKEN --paymaster-address 0x98546B226dbbA8230cf620635a1e4ab01F6A99B2 2>&1)
  echo "$output"

  # Check if output contains error messages (hardhat may exit with 0 even when errors occur)
  if echo "$output" | grep -qiE "error|unexpected error|nonce too high"; then
    echo "Command failed, creating a new private key"
    npx hardhat vars set PRIVATE_KEY $(openssl rand -hex 32)
    continue
  fi

  # Capture the end time
  end_time=$(date +%s)

  # Calculate the duration
  duration=$((end_time - start_time))

  # Log the duration
  echo "$(date '+%Y-%m-%d %H:%M:%S') - Command execution time: ${duration} seconds"

  # Optional: Add a sleep interval to avoid overwhelming the server
  sleep $SLEEP_INTERVAL
done
