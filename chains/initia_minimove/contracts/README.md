# Stork Intia Minimove Contract

This directory contains the Stork Initia Minimove compatible contract written in [Initia's flavor of Aptos Move](https://docs.initia.xyz/home/core-concepts/initia-and-rollups/rollups/vms/minimove/introduction).

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
