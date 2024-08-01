# Stork Contract

This directory contains a [Hardhat](https://hardhat.org/docs) project used to manage and deploy the Stork EVM compatible contract.

This contract is used to read and write the latest values from the Stork network on-chain.

In order for a new value to be accepted by the contract, the signature associated with the value must be validated against Stork's public key.

This signature is derived from

1. Stork's public key
2. Encoded Asset ID (keccak256 hash of the asset's symbol)
3. Quantized value
4. Timestamp in nanoseconds
5. Merkle root of the signed publisher message hashes
6. Value computation algorithm hash

See `verifyStorkSignatureV1` in `contracts/StorkVerify.sol` for more details.

### Getting started

```
npm i
npx hardhat test
```

### Local Development

#### Run local node

```
npx hardhat node
```

#### Deploy

```
npx hardhat compile
npx hardhat --network inMemoryNode deploy
```

#### Upgrade

```
npx hardhat compile
npx hardhat --network inMemoryNode upgrade
```

#### Verify

```
npx hardhat verify --network <network> <proxy-contract-address>
```

#### Interact

See `tasks/interact.ts` for available methods.

```
npx hardhat --network inMemoryNode inMemoryNode
```

#### Test

```
npx hardhat test
```

#### Generate ABI

```
npx hardhat print-abi > stork.abi
```

### Deploy on-chain

1. Configure your `hardhat.config.ts` with the desired network.
2. Run `npx hardhat --network <network> deploy` to deploy the contract.
3. Deployment will be saved in the `deployments` directory.
