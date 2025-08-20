# Stork EVM SDK Example

This is a very simple Hardhat project to show how you would use the Stork EVM SDK to consume Stork price updates in your Solidity contract.

## Deploy locally

1. Deploy a local version of the [Stork contract](../../contracts/stork) 
2. Initialize the contract and write some data to it for your desired asset id using the CLI tools
3. Update the [ignition module](ignition/modules/Example.ts) to pass your local Stork contract's address as the `storkContractAddress` parameter
4. Install dependencies:
```bash
npm install
```
5. Compile the contract:
```bash
npx hardhat compile
```
6. Deploy the contract locally and record the address:
```bash
npx hardhat --network inMemoryNode ignition deploy ignition/modules/Example.ts --parameters ignition/parameters.json
```
7. Read the price from the Stork feed:
```bash
npx hardhat --network inMemoryNode read-price --example-contract-address YOUR_EXAMPLE_CONTRACT_ADDRESS --asset BTCUSD
```
