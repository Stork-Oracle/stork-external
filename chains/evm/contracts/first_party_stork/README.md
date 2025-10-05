# First Party Stork Contract

This directory contains a [Hardhat](https://hardhat.org/docs) project used to manage and deploy the First Party Stork EVM compatible contract.

This contract is used to read and write per publisher latest value price updates on-chain. Updates have optional historical data storage.

In order for a new value to be accepted by the contract, the signature associated with the value must be validated against the relevant registered publisher's public key.

This signature is derived from

1. A registered publisher's public key
2. Asset ID
3. Timestamp in seconds
4. Quantized value

See `verifyPublisherSignatureV1` in `contracts/first_party_stork/FirstPartyStorkVerify.sol` for more details.

## Getting Started

### Installation

```bash
npm i
```

### Running Tests

```bash
npx hardhat test
```

## Deployment

### Local Development

#### Run local node

```bash
npx hardhat node
```

#### Deploy to local network

```bash
npx hardhat ignition deploy ignition/modules/FirstPartyStork.ts --network localhost
```

#### Register Publisher on local contract

```bash
RPC_URL=<rpc-url> \
PRIVATE_KEY=<owner-private-key> \
CONTRACT_ADDRESS=<deployed-contract-address> \
PUBLISHER_EVM_PUBLIC_KEY=<publisher-address> \
npx ts-node scripts/local_register_publisher.ts
```

### Running with Docker

To run a local node, deploy the contract to the local network, and register a publisher on the local contract, use the command below from [first_party_pusher](../../../../apps/first_party_pusher/). This is the recommended method.

```bash
docker compose up evm-contract
```
