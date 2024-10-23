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

```
anchor localnet
```

#### Deploy

```
anchor deploy --provider.cluster localnet
```

#### Upgrade

```
anchor upgrade --provider.cluster localnet
```

#### Verify

```
anchor verify --provider.cluster localnet 9yjwoWUgyKeH2cEC4S5G9uudobYAcmDH9zU1mq1hKWyb
```

#### Test

```
anchor test
```

#### Generate IDL

```
anchor build
```
