# Kadena Stork Admin CLI

A command-line interface for managing the Stork oracle contract on Kadena blockchain.

## Setup

1. Install dependencies:
```bash
npm install
```

2. Set environment variables:
```bash
export NETWORK_ID="development"  # or "testnet01", "mainnet01"
export CHAIN_ID="1"
export API_HOST="http://localhost:8080"  # or "https://api.testnet.chainweb.com" for testnet
export STORK_CONTRACT_ADDRESS="stork"
export ADMIN_ACCOUNT="k:your-public-key"
export ADMIN_SECRET_KEY="your-secret-key"
```

## Commands

### Deploy Contract
```bash
npm run dev deploy
```

### Initialize Contract
```bash
npm run dev initialize --stork-key 0x0a803F9b1CCe32e2773e0d2e98b37E0775cA5d44 --fee 1
```

### Get Contract State
```bash
npm run dev get-state-info
```

### Update EVM Public Key
```bash
npm run dev update-stork-public-key 0x0a803F9b1CCe32e2773e0d2e98b37E0775cA5d44
```

### Update Fee
```bash
npm run dev update-single-update-fee-in-stu 1000
```

### Write Price Data to Feeds
```bash
npm run dev write-to-feeds "BTCUSD,ETHUSD" "https://rest.jp.stork-oracle.io" "your-auth-key"
```

### Read from Feed
```bash
npm run dev read-from-feed "BTCUSD"
```

## Configuration

The CLI follows the same configuration structure as other Stork chain integrations:

- **Environment Variables**: Configure network, accounts, and contract addresses
- **Key Management**: Uses admin account and secret key for transaction signing
- **Network Support**: Supports development, testnet, and mainnet networks

## Key Features

- **Contract Deployment**: Deploy the Pact contract to Kadena
- **Initialization**: Set up the contract with initial parameters
- **State Management**: Update EVM public keys and fees
- **Feed Operations**: Write price data and read current values
- **Error Handling**: Comprehensive error handling and logging

## Integration with Stork REST API

The CLI integrates with the Stork REST API to fetch the latest price data and write it to the Kadena blockchain. The price data includes:

- Encoded asset IDs
- Temporal numeric values with timestamps
- Publisher merkle roots
- Value computation algorithm hashes
- Cryptographic signatures (r, s, v components)

## Security Notes

- Always verify transaction details before signing
- Keep private keys secure and never commit them to version control
- Use appropriate gas limits and prices for your network
- Test thoroughly on development/testnet before mainnet deployment