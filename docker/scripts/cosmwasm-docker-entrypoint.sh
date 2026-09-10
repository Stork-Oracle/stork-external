#!/bin/sh

# Starts a single-node wasmd chain, deploys the Stork CosmWasm contract to it, and keeps the node running.
# The chain ID, denom, and pusher mnemonic must match the defaults in
# apps/chain_pusher/pkg/cosmwasm/interactor_integration_test.go.

set -eu

CHAIN_ID=stork-local
DENOM=stake
# Well-known BIP-39 test mnemonic. It is only ever funded on this throwaway local chain.
PUSHER_MNEMONIC="abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
WASM_FILE=/opt/stork_cw.wasm
# The docker-compose healthcheck waits for this file.
READY_FILE=/tmp/stork-contract-address
KEYRING="--keyring-backend test"

# Converts a 0x-prefixed EVM address into the JSON byte array the contract expects.
evm_address_to_json_bytes() {
	hex=${1#0x}
	if [ ${#hex} -ne 40 ]; then
		echo "expected a 20-byte hex address, got: $1" >&2
		return 1
	fi

	bytes=""
	while [ -n "$hex" ]; do
		rest=${hex#??}
		bytes="${bytes:+$bytes,}$(printf '%d' "0x${hex%"$rest"}")"
		hex=$rest
	done

	echo "[$bytes]"
}

# Waits for a transaction to be included in a block and fails if it was rejected.
wait_for_tx() {
	for _ in $(seq 60); do
		if result=$(wasmd query tx "$1" --output json 2>/dev/null); then
			if [ "$(echo "$result" | jq -r '.code // 0')" != "0" ]; then
				echo "transaction $1 failed: $(echo "$result" | jq -r '.raw_log')" >&2
				return 1
			fi

			return 0
		fi

		sleep 1
	done

	echo "timed out waiting for transaction $1" >&2
	return 1
}

# Broadcasts a wasm transaction from the deployer account and waits for it to be included.
broadcast_wasm_tx() {
	result=$(wasmd tx wasm "$@" --from deployer --chain-id "$CHAIN_ID" --gas auto --gas-adjustment 1.5 \
		--gas-prices "0$DENOM" --yes --output json $KEYRING)
	if [ "$(echo "$result" | jq -r '.code // 0')" != "0" ]; then
		echo "broadcast failed: $(echo "$result" | jq -r '.raw_log')" >&2
		return 1
	fi

	wait_for_tx "$(echo "$result" | jq -r '.txhash')"
}

STORK_EVM_PUBLIC_KEY_JSON=$(evm_address_to_json_bytes "${STORK_PUBLIC_KEY:?STORK_PUBLIC_KEY must be set}")

wasmd init local --chain-id "$CHAIN_ID" > /dev/null 2>&1
# Produce blocks every 500ms instead of every 1s so tests spend less time waiting for inclusion.
sed -i 's/^timeout_commit = .*/timeout_commit = "500ms"/' "$HOME/.wasmd/config/config.toml"

wasmd keys add deployer $KEYRING > /dev/null 2>&1
echo "$PUSHER_MNEMONIC" | wasmd keys add pusher --recover $KEYRING > /dev/null 2>&1
wasmd genesis add-genesis-account deployer "1000000000000$DENOM" $KEYRING
wasmd genesis add-genesis-account pusher "1000000000000$DENOM" $KEYRING
wasmd genesis gentx deployer "250000000$DENOM" --chain-id "$CHAIN_ID" $KEYRING > /dev/null 2>&1
wasmd genesis collect-gentxs > /dev/null 2>&1

wasmd start --rpc.laddr tcp://0.0.0.0:26657 --log_level warn &
NODE_PID=$!
trap 'kill "$NODE_PID"' INT TERM

until wasmd status 2>&1 | jq -e '.sync_info.latest_block_height | tonumber > 0' > /dev/null 2>&1; do
	kill -0 "$NODE_PID"
	sleep 1
done

broadcast_wasm_tx store "$WASM_FILE"
broadcast_wasm_tx instantiate 1 \
	"{\"stork_evm_public_key\":$STORK_EVM_PUBLIC_KEY_JSON,\"single_update_fee\":{\"amount\":\"1\",\"denom\":\"$DENOM\"}}" \
	--label stork --no-admin

CONTRACT_ADDRESS=$(wasmd query wasm list-contract-by-code 1 --output json | jq -r '.contracts[0]')

# Make sure the contract verifies signatures against the key that signed the test data.
DEPLOYED_KEY_JSON=$(wasmd query wasm contract-state smart "$CONTRACT_ADDRESS" '{"get_stork_evm_public_key":{}}' \
	--output json | jq -c '.data.stork_evm_public_key')
if [ "$DEPLOYED_KEY_JSON" != "$STORK_EVM_PUBLIC_KEY_JSON" ]; then
	echo "contract has stork_evm_public_key $DEPLOYED_KEY_JSON, expected $STORK_EVM_PUBLIC_KEY_JSON" >&2
	exit 1
fi

echo "$CONTRACT_ADDRESS" > "$READY_FILE"
echo "Stork contract deployed at $CONTRACT_ADDRESS"

# The node never exits on its own; waiting keeps the container running.
wait "$NODE_PID"
