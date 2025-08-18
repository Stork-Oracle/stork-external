# Stork Chainlink Adapter Contract Example

This is a simple Hardhat project demonstrating how to use the Stork Chainlink Adapter in your Solidity contract.

## Deploy locally

1. Deploy a local version of the [Stork contract](../../contracts/stork) 
2. Write some data to it for your desired asset id
3. Update the parameters in [ignition/parameters.json](ignition/parameters.json):
   - Set `storkContractAddress` to your deployed Stork contract address
   - Set `priceId` to your desired encoded asset id (defaults to BTCUSD)
4. Install dependencies:
```bash
npm install
```
5. Compile the contract:
```bash
npx hardhat compile
```
6. Deploy the contract locally:
```bash
npm run deploy
```
7. Get the latest round data from your contract:
```bash
npx hardhat --network localhost get_latest_round_data --example-contract-address YOUR_DEPLOYED_CONTRACT_ADDRESS
```
