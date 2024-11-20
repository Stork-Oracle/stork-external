#!/bin/bash
set -eE

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

  # Run the command
  npx hardhat --network sophonTestnet interact updateTemporalNumericValuesV1 $FEEDS https://rest.jp.stork-oracle.network $AUTH_TOKEN --paymaster-address 0x98546B226dbbA8230cf620635a1e4ab01F6A99B2

  # Capture the end time
  end_time=$(date +%s)

  # Calculate the duration
  duration=$((end_time - start_time))

  # Log the duration
  echo "Command execution time: ${duration} seconds"

  # Optional: Add a sleep interval to avoid overwhelming the server
  sleep $SLEEP_INTERVAL
done
