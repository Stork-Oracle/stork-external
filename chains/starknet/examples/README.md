# Starknet Examples

Two ways to consume a Stork feed on Starknet.

## Cairo: reading from a contract

[`src/example.cairo`](src/example.cairo) reads a feed through the
[SDK](../sdks/stork_starknet_sdk), showing both the checked read and the unchecked one.

```bash
scarb build
snforge test
```

The tests deploy a real Stork contract, push a genuine signed update through it, and check that
the example both reads a fresh price and refuses a stale one.

## TypeScript: reading from an app

[`app/example.ts`](app/example.ts) reads the same feed over JSON-RPC.

```bash
cd app && npm install

STORK_CONTRACT_ADDRESS=0x<deployed-contract> \
ENCODED_ASSET_ID=0x<encoded-asset-id> \
    npx tsx example.ts
```

It defaults to mainnet; set `STARKNET_RPC_URL` to point elsewhere. It prints both reads so the
difference is visible: the checked read rejects a stale feed, the
unchecked one returns it anyway. The encoded asset id is the keccak256 of the asset symbol;
subscribe to the asset on the Stork network to get it.

Note `quantized_value` is an `i128` serialised as a field element, so negative prices come back as
`P - |v|` and need decoding. Both examples show how.
