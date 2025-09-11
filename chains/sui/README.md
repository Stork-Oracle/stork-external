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
#### Test

```bash
sui move test
```

#### Deploy

```bash
sui move build
sui client publish
```

#### Upgrade

```bash
sui move build
sui client upgrade --upgrade-capability
```

Upon upgrading, you must call migrate on the new contract via the admin cli to enable the new contract and disable the old.

#### Note

Sui packages are capable of automatically handling addresses via the Move.lock file. This allows for easy inclusion as a dependency via a github url without having to manually specify the deployed address. Because of this, the Move.lock file must be checked into version control, and must be pushed to the remote repository whenever the package address changes. The Move.lock file should not be updated manually.

