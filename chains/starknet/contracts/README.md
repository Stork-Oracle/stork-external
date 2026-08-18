# Stork Starknet Contract

This directory contains the Stork compatible contract for Starknet, written in Cairo.

It is the Cairo counterpart of [the EVM contract](../../evm/contracts/stork) and keeps the same
semantics: it stores the latest Stork-signed value per asset feed, and each update is verified
against a secp256k1 signature produced by Stork's EVM signing key. Consumers integrate against
[`IStork`](../sdks/stork_starknet_sdk/src/interface.cairo), which ships in the
[SDK](../sdks/stork_starknet_sdk).

## Verification

Stork signs `keccak256(abi.encodePacked(storkPubKey, id, recvTime, quantizedValue,
publisherMerkleRoot, valueComputeAlgHash))` under the EIP-191 personal-sign prefix. The contract
reproduces that encoding byte for byte in [`src/verify.cairo`](src/verify.cairo) and recovers the
signer with the corelib's secp256k1 syscalls, so the same signed payload the Stork aggregator
produces for Ethereum is accepted here unmodified.

Two details are easy to get wrong and are pinned by tests:

* Ethereum's keccak256 is the byte-reverse of what `core::keccak` returns.
* A negative quantized value is sign-extended to 32 bytes in the *signed message* (two's
  complement), but is serialized as the field element `P - |v|` in *calldata* (Cairo's
  `Serde<i128>`). These are different encodings of the same number.

## Differences from the EVM contract

Both are forced by the platform rather than chosen:

* **Fees** are collected with an ERC20 `transfer_from` instead of `msg.value`, since Starknet has
  no native value transfer. A deployment with `single_update_fee` of 0 never touches the token, so
  the pusher needs no allowance in that case; a non-zero fee requires the pusher to approve the
  contract on the configured `fee_token`.
* **Upgrades** use `replace_class_syscall` rather than a UUPS proxy.

One deliberate behavioural difference: recency is checked *before* signature verification. Verifying
a signature costs roughly 21M L2 gas, and a stale update is discarded either way, so checking the
stored timestamp first avoids paying for updates that will be dropped. This matches the CosmWasm and
Move contracts.

`get_multiple_temporal_numeric_values_unchecked` is an addition. `get_temporal_numeric_values_unsafe_v1`
matches EVM and reverts if any requested feed is missing, which is the wrong shape for a poller: one
not-yet-populated feed would hide every other feed in the batch. The unchecked variant returns a
zeroed value for unknown feeds instead, and is what the chain pusher polls with.

## Development

Requires [Scarb](https://docs.swmansion.com/scarb/) and
[Starknet Foundry](https://foundry-rs.github.io/starknet-foundry/).

### Build

```bash
scarb build
```

### Test

```bash
snforge test
```

The suite runs the contract against real Stork-signed updates taken from
`apps/chain_pusher/internal/testutil/testdata`, plus the hash vectors asserted by the CosmWasm
contract's `verify.rs` tests.

### Deploy

Deployments go to mainnet by default; rehearse on Sepolia first by pointing at a Sepolia RPC.
The `sncast` flow below targets a local devnet:

```bash
starknet-devnet --seed 42

sncast account import --name devnet \
    --address <predeployed-account-address> \
    --private-key <predeployed-private-key> \
    --type oz --url http://127.0.0.1:5050

sncast --account devnet declare --contract-name Stork --url http://127.0.0.1:5050

sncast --account devnet deploy \
    --class-hash <class-hash> \
    --arguments '<initial_owner>, <stork_public_key>, <valid_time_period_seconds>, <single_update_fee>_u256, <fee_token>' \
    --url http://127.0.0.1:5050
```

The constructor takes the owner, the Stork EVM signing key as an `EthAddress`, the staleness
window in seconds, the per-update fee, and the ERC20 the fee is denominated in. Pass `0x0` for
`fee_token` when the fee is zero.

## Operational limits

Signature verification dominates the cost of an update: roughly **21.8M L2 gas per update**,
measured as marginal cost across batch sizes (1 update ~22.6M, 4 updates ~87.9M). Storage and event
costs are negligible next to it.

This sets a ceiling on batch size — divide your target network's per-transaction L2 gas limit by
~21.8M. Confirm the practical maximum on Sepolia before choosing the pusher's batching window,
since an oversized batch fails as a whole.

## Administration

All admin entry points are owner-only. Ownership transfer is two-step, mirroring
`Ownable2StepUpgradeable`.

| Function | Purpose |
| --- | --- |
| `update_stork_public_key` | Rotate the canonical signing key. The previous key stays accepted. |
| `add_signing_address` / `remove_signing_address` | Manage the accepted signer set (max 8). |
| `update_single_update_fee` / `update_fee_token` | Change fee configuration. Set the token before a non-zero fee, and clear the fee before the token: a non-zero fee with no fee token would revert every update. |
| `update_valid_time_period_seconds` | Change the staleness window. |
| `withdraw_fees` | Withdraw collected fees. |
| `upgrade` | Replace the contract class. |
| `transfer_ownership` / `accept_ownership` | Two-step ownership handover. |
