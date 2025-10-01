# Stork Intia Minimove Contract

This directory contains the Stork Initia Minimove compatible contract written in [Initia's flavor of Aptos Move](https://docs.initia.xyz/home/core-concepts/initia-and-rollups/rollups/vms/minimove/introduction).

This contract is used to read and write the latest values from the Stork network on-chain.

### Getting started

```bash
minitiad init
minitiad move test --language-version 2.1 --dev
```

### Local Development

#### Run local node

```bash
aptos node run-local-testnet --with-indexer-api
```

#### Test

```bash
minitiad move test --language-version 2.1 --dev
```

#### Deploy

```bash
minitiad move deploy-object --from <key_name> --gas auto --gas-adjustment 1.5 --gas-prices <gas_price> --node <rpc_url> --chain-id <chain_id> --language-version 2.1
```

#### Upgrade

```bash
aptos move upgrade-object --address-name stork --profile <profile> --object-address <object-address> --move-2
```
