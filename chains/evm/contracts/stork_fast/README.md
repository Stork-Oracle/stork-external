# Sample Hardhat 3 Beta Project (`node:test` and `viem`)

This directory contains a [Hardhat](https://hardhat.org/docs) project for the Stork Fast EVM compatible contract.

This contract is used to verify and deserialize Stork Fast signed ECDSA payloads.

In order to verify a signed ECDSA payload, use the `verifySignedECDSAPayload` function.

To also deserialize the update values from the signed ECDSA payload, use the `verifyAndDeserializeSignedECDSAPayload` function.

### Getting started

```bash
npm i
```

### Running Tests

```bash
npx hardhat test
```

### Local Development

#### Run local node

```bash
npx hardhat node
```

#### Deploy

```bash
npx hardhat compile
npx hardhat ignition deploy ignition/modules/StorkFast.ts --network hardhatLocal --parameters ignition/parameters.json
```

Note: the `--parameters` flag is required — the module has no defaults for `signerAddresses` and `verificationFeeInWei`, and Ignition does not auto-load `ignition/parameters.json`. Parameters must be keyed under `StorkFastProxyModule` (the module that reads them), as in the checked-in file. `--verify` is omitted here since Etherscan verification does not work against a local node.

#### Interact

See `tasks/admin.ts` for available methods.

```bash
npx hardhat --network hardhatLocal <method>
```

### Deploy on-chain

1. Configure your `hardhat.config.ts` with the desired network.
2. Update the `ignition/parameters.json` file with the desired signer addresses and verification fee in wei (keyed under `StorkFastProxyModule`).
3. Run `npx hardhat ignition deploy ignition/modules/StorkFast.ts --network <network> --parameters ignition/parameters.json --verify` to deploy the contract.
4. Deployment will be saved in the `ignition/deployments` directory.
