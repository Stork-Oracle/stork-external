# Starknet Chain Pusher

The Starknet pusher, structured like every other chain in `apps/chain_pusher/pkg`: `push.go`
defines the cobra command, `interactor.go` implements `types.ContractInteractor`, and `bindings/`
talks to the contract.

The contract it pushes to lives in [`chains/starknet/contracts`](../../../../chains/starknet/contracts).

## Wallet setup

Starknet accounts are contracts, so the pusher needs two things rather than one:

* `--account-address` — the address of the deployed account contract.
* `--private-key-file` — a file containing the hex encoded Stark private key that controls it.

The account must already be deployed and funded.

## Running

```bash
go run ./main.go starknet --help
```

Basic usage:

```bash
go run ./main.go starknet \
    -w wss://api.jp.stork-oracle.network \
    -a <stork-api-key> \
    -r <chain-rpc-url> \
    -c <chain-ws-url> \
    -x <contract-address> \
    -n <account-address> \
    -f <asset-config-file> \
    -k <private-key-file>
```

`--chain-ws-url` is optional. When given, the pusher subscribes to the contract's `ValueUpdate`
events over `starknet_subscribeEvents` and tracks on-chain state without waiting for the poll
interval; without it, it falls back to polling only.

`--fee-token-address` is optional and only affects balance reporting.

## Fees

The contract charges `single_update_fee` per applied update, collected as an ERC20 `transfer_from`.
Deployments with a fee of 0 — the common case — need nothing extra. With a non-zero fee, the
pushing account must approve the Stork contract as a spender on the fee token, or updates will
revert.

## Bindings

There is no `abigen` equivalent for Starknet, so `bindings/` is written by hand and
[`bindings/serde.go`](bindings/serde.go) implements the Cairo calldata layout directly. The two
encodings that matter:

* `u256` is two field elements, low limb first.
* `i128` is a single field element, negative values embedded as `P - |v|`. Note that
  `utils.BigIntToFelt` builds a felt from `big.Int.Bytes()` and therefore drops the sign, so
  negative values must be reduced before conversion.

`serde_test.go` pins both, along with the entry point and event selectors.

## RPC spec versions

Starknet is mid-migration between JSON-RPC 0.9 and 0.10, and providers differ even on the same
network: at the time of writing one Sepolia endpoint serves 0.9.0, another 0.10.2, and a third
0.10.3-rc.0. The push path is verified against all three, and against mainnet at 0.10.2.

That works because `bindings/invoke.go` builds and submits the invoke transaction itself rather
than using `account.BuildAndSendInvokeTxn`. starknet.go's `BroadcastInvokeTxnV3` serializes two
optional fields without `omitempty`, and types one of them as an array where the spec calls for a
base64 string, so nodes reject every transaction it builds:

```
json: cannot unmarshal array into Go struct field BroadcastedTransaction.proof of type core.Base64
```

Transaction hashing and signing still come from starknet.go; only the JSON payloads are written
here, carrying just the fields a Stork update needs.

`--tip` sets an explicit transaction tip in FRI. It is optional: the tip defaults to zero and the
fee is estimated per transaction. Set it to bid for faster inclusion on a busy network.

## Testing

Unit tests:

```bash
go test ./apps/chain_pusher/pkg/starknet/...
```

Integration tests push the signed fixtures from `internal/testutil/testdata` through a real
deployed contract and read them back. They skip unless a node is reachable, so `make
integration-test` stays green without one. Point them at a contract with:

```bash
STARKNET_RPC_URL=<rpc-url> \
STARKNET_ACCOUNT_ADDRESS=0x... \
STARKNET_PRIVATE_KEY=0x... \
STORK_CONTRACT_ADDRESS=0x... \
  go test -tags integration ./apps/chain_pusher/pkg/starknet/...
```

The tests walk the fixtures forward until a batch is fresh, so they can be run repeatedly against
the same contract.
