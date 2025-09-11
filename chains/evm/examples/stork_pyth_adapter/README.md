# Stork Pyth Adapter Contract Example

This is a simple Hardhat project demonstrating how to use the Stork Pyth Adapter in your Solidity contract.

## Deploy locally

1. Deploy a local version of the [Stork contract](../../contracts/stork)
2. Write some data to it for your desired asset id  
3. Update the parameters in [ignition/parameters.json](ignition/parameters.json):
   - Set `storkContractAddress` to your deployed Stork contract address
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
npx hardhat --network inMemoryNode ignition deploy ignition/modules/ExampleStorkPythAdapter.ts --parameters ignition/parameters.json
```
7. Get the latest price from your contract:
```bash
npx hardhat --network inMemoryNode get_latest_price --example-contract-address YOUR_DEPLOYED_CONTRACT_ADDRESS --price-id YOUR_ENCODED_ASSET_ID
```

