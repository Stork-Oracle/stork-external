# Stork Aptos Contract

This directory contains the Stork Aptos compatible contract written in [Aptos Move](https://aptos.dev/network/blockchain/move).

This contract is used to read and write the latest values from the Stork network on-chain.

### Getting started

```bash
aptos init
aptos move test --move-2 --dev
```

### Local Development

#### Run local node

```bash
aptos node run-local-testnet --with-indexer-api
```

#### Test

```bash
aptos move test --move-2 --dev
```

#### Deploy

```bash
aptos move deploy-object --address-name stork --profile <profile> --move-2
```

#### Upgrade

```bash
aptos move upgrade-object --address-name stork --profile <profile> --object-address <object-address> --move-2
```
