# Stork Solana Contract

This directory contains an [Anchor](https://www.anchor-lang.com/) project used to manage and deploy the Stork Solana compatible contract.

This contract is used to read and write the latest values from the Stork network on-chain.

### Getting started

```
solana-keygen new --outfile ~/.config/solana/id.json
export COPYFILE_DISABLE=1 # for macos
yarn install
anchor test
```

### Local Development

#### Run local node

It appears that the contract will be deployed to the localnet cluster by default.

```
anchor localnet
```

#### Deploy

```
anchor deploy
```

#### Upgrade

```
anchor upgrade
```

#### Verify

```
anchor verify <program-id>
```

#### Test

```
anchor test
```

#### Generate IDL

```
anchor build
```
