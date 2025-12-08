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
npx hardhat ignition deploy ignition/modules/StorkFast.ts --network inMemoryNode
```

#### Upgrade
