#!/usr/bin/env bash
#
# End to end check of the Starknet push path against a local devnet.
#
# Builds the contract, deploys it with the CLI, then pushes real Stork-signed fixtures through it
# with the Go pusher and reads them back. This is the whole loop the pusher performs in production,
# minus the Stork websocket feed.
#
# Usage: chains/starknet/scripts/e2e.sh [--keep]
#   --keep   leave the devnet running after the check

set -euo pipefail

KEEP_DEVNET=false
[[ "${1:-}" == "--keep" ]] && KEEP_DEVNET=true

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
CONTRACTS_DIR="$REPO_ROOT/chains/starknet/contracts"
CLI_DIR="$REPO_ROOT/chains/starknet/cli"

DEVNET_PORT="${DEVNET_PORT:-5050}"
DEVNET_URL="http://127.0.0.1:$DEVNET_PORT"
DEVNET_SEED=42

# Predeployed by `starknet-devnet --seed 42`.
ACCOUNT_ADDRESS=0x34ba56f92265f0868c57d3fe72ecab144fc96f97954bbbc4252cef8e8a979ba
PRIVATE_KEY=0xb137668388dbe9acdfa3bc734cc2c469

# The key that signed the fixtures in apps/chain_pusher/internal/testutil/testdata.
STORK_KEY=0xC4A02e7D370402F4afC36032076B05e74FF81786

# The pinned Go client emits the SNIP-36 `proof_facts` field, which nodes below spec 0.10.2 reject,
# so 0.10.0 is not good enough here.
MIN_SPEC="0.10.2"

DEVNET_PID=""
DEVNET_LOG="$(mktemp)"

step() { printf '\n\033[1m==> %s\033[0m\n' "$1"; }
fail() { printf '\033[31merror: %s\033[0m\n' "$1" >&2; exit 1; }

cleanup() {
  if [[ -n "$DEVNET_PID" && "$KEEP_DEVNET" == false ]]; then
    kill "$DEVNET_PID" 2>/dev/null || true
    wait "$DEVNET_PID" 2>/dev/null || true
    echo "devnet stopped"
  elif [[ -n "$DEVNET_PID" ]]; then
    echo "devnet left running on $DEVNET_URL (pid $DEVNET_PID)"
  fi
}
trap cleanup EXIT

step "Checking prerequisites"
for tool in scarb snforge starknet-devnet node go; do
  command -v "$tool" >/dev/null || fail "$tool is not installed"
done
echo "all present"

step "Building the contract"
(cd "$CONTRACTS_DIR" && scarb build)

step "Running contract tests"
(cd "$CONTRACTS_DIR" && snforge test)

step "Running example tests"
(cd "$REPO_ROOT/chains/starknet/examples" && scarb build && snforge test)

step "Starting devnet on port $DEVNET_PORT"
if curl -s -m 2 -X POST "$DEVNET_URL" \
     -H 'Content-Type: application/json' \
     -d '{"jsonrpc":"2.0","method":"starknet_specVersion","params":[],"id":1}' >/dev/null 2>&1; then
  echo "reusing the devnet already listening on $DEVNET_URL"
else
  starknet-devnet --seed "$DEVNET_SEED" --host 127.0.0.1 --port "$DEVNET_PORT" >"$DEVNET_LOG" 2>&1 &
  DEVNET_PID=$!

  for _ in $(seq 1 30); do
    sleep 1
    curl -s -m 2 -X POST "$DEVNET_URL" \
      -H 'Content-Type: application/json' \
      -d '{"jsonrpc":"2.0","method":"starknet_specVersion","params":[],"id":1}' >/dev/null 2>&1 && break
  done
fi

SPEC=$(curl -s -m 5 -X POST "$DEVNET_URL" \
  -H 'Content-Type: application/json' \
  -d '{"jsonrpc":"2.0","method":"starknet_specVersion","params":[],"id":1}' |
  sed -n 's/.*"result":"\([^"]*\)".*/\1/p')

[[ -n "$SPEC" ]] || { cat "$DEVNET_LOG" >&2; fail "devnet did not come up"; }
echo "devnet is serving JSON-RPC spec $SPEC"

if [[ "$(printf '%s\n%s\n' "$MIN_SPEC" "$SPEC" | sort -V | head -1)" != "$MIN_SPEC" ]]; then
  fail "devnet serves JSON-RPC spec $SPEC, but the pinned Go client needs $MIN_SPEC or newer.
       Transactions will be rejected with \`unknown field 'proof_facts'\`.
       Install a newer starknet-devnet:  asdf install starknet-devnet 0.9.1
       then put it first on PATH, or set DEVNET_PORT to one already serving $MIN_SPEC+."
fi

step "Installing CLI dependencies"
# The deploy CLI is local tooling and is not part of the repo, so say so plainly rather than
# failing inside npm.
[[ -d "$CLI_DIR" ]] || fail "no CLI at $CLI_DIR.
       Deployment tooling is kept out of this repo; this script needs a local copy."
(cd "$CLI_DIR" && [[ -d node_modules ]] || npm install --silent)
echo "ready"

step "Deploying the contract"
DEPLOY_OUTPUT=$(cd "$CLI_DIR" && \
  STARKNET_RPC_URL="$DEVNET_URL" \
  STARKNET_ACCOUNT_ADDRESS="$ACCOUNT_ADDRESS" \
  STARKNET_PRIVATE_KEY="$PRIVATE_KEY" \
  npx tsx admin.ts deploy --stork-public-key "$STORK_KEY" 2>&1)
echo "$DEPLOY_OUTPUT"

CONTRACT_ADDRESS=$(echo "$DEPLOY_OUTPUT" | sed -n 's/.*export STORK_CONTRACT_ADDRESS=\(0x[0-9a-f]*\).*/\1/p')
[[ -n "$CONTRACT_ADDRESS" ]] || fail "could not parse the deployed contract address"

step "Reading the deployed configuration back"
(cd "$CLI_DIR" && \
  STARKNET_RPC_URL="$DEVNET_URL" \
  STARKNET_ACCOUNT_ADDRESS="$ACCOUNT_ADDRESS" \
  STARKNET_PRIVATE_KEY="$PRIVATE_KEY" \
  STORK_CONTRACT_ADDRESS="$CONTRACT_ADDRESS" \
  npx tsx admin.ts info)

step "Pushing signed fixtures with the Go pusher"
(cd "$REPO_ROOT" && \
  STORK_CONTRACT_ADDRESS="$CONTRACT_ADDRESS" \
  STARKNET_DEVNET_URL="$DEVNET_URL" \
  go test -count=1 -tags integration -v -run TestIntegration ./apps/chain_pusher/pkg/starknet/)

printf '\n\033[32m==> push path verified end to end\033[0m\n'
echo "    contract: $CONTRACT_ADDRESS"
echo "    devnet:   $DEVNET_URL (spec $SPEC)"
