# Stork Sui Contract

This directory contains an [Sui Move](https://sui.io/move) project used to manage and deploy the Stork Sui compatible contract.

This contract is used to read and write the latest values from the Stork network on-chain.

### Getting started


```bash
sui client
sui keytool generate ed25519
sui move test
```
### Local Development

#### Run local node

```bash
RUST_LOG="off,sui_node=info" sui start --with-faucet --force-regenesis
```

#### Deploy

```bash
sui move build
sui move publish
```

#### Upgrade

```bash
sui move build
sui move publish
```


#### Test

```bash
sui move test
```

