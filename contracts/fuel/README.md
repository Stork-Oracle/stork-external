# Stork Fuel Contract

This directory contains the Stork compatible Fuel contract written in [Sway](https://docs.fuel.network/docs/sway/), as well as a CLI tool used to manage the Stork Fuel compatible contract.

This contract is used to read and write the latest values from the Stork network on-chain. For reading values on chain, see the [stork-fuel-sdk](../../sdks/fuel/stork-fuel-sdk).

### Getting started

From `contracts/stork/`

```bash
forc test
cargo test
```

### Local Development

For any command that deploys, including:

- `npx fuels deploy`
- `npx fuels test`

please ensure you have removed the `address` under the `[proxy]` section from [Forc.toml](contracts/stork/Forc.toml) if you intend for a new proxy to be deployed, as this is set by those commands and assume a proxy already exists at the address if set. If you do not see the `address` field, you are good to go.

#### Run local node

This command also builds, generates types, and deploys contract + proxy.

From `cli/`

```bash
npx fuels dev
```

#### Build

This command also generates types for the contract in `cli/types/`.

From `cli/`

```bash
npx fuels build
```

#### Deploy

This command writes the proxy address to [Forc.toml](contracts/stork/Forc.toml), as well as the proxy and impl addresses to [contract-ids.json](cli/contract-ids.json).

From `cli/`

```bash
npx fuels deploy
```

#### Upgrade

Upgrading the contract is the same as deploying, but uses the `[proxy]` `address` field in [Forc.toml](contracts/stork/Forc.toml) as the proxy to upgrade. If you intend to upgrade a contract, ensure this is set to the proxy address you intend to upgrade.

```bash
npx fuels deploy
```
