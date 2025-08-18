# Stork EVM SDK Example

This is a very simple Hardhat project to show how you would use the Stork EVM SDK to consume Stork price updates in your Solidity contract.

## Deploy locally

1. Deploy a local version of the [Stork contract](../../contracts/stork) 
2. Initialize the contract and write some data to it for your desired asset id using the CLI tools
3. Update the [ignition module](ignition/modules/ExampleStorkSDK.ts) to pass your local Stork contract's address as the `storkContractAddress` parameter
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
npx hardhat --network localhost ignition deploy ignition/modules/ExampleStorkSDK.ts --parameters '{"ExampleStorkSDKModule":{"storkContractAddress":"YOUR_STORK_CONTRACT_ADDRESS"}}'
```
7. Read the price from the Stork feed:
```bash
npx hardhat --network localhost read-price --example-contract-address YOUR_EXAMPLE_CONTRACT_ADDRESS --asset BTCUSD
```

## Example Usage

The contract provides two main functions:

- `useStorkPrice(bytes32 feedId)` - Reads the latest price with staleness check and emits an event
- `useStorkPriceUnsafe(bytes32 feedId)` - Reads the latest price without staleness check (view function)

The asset identifier (like "BTCUSD") is automatically hashed using keccak256 to create the feedId, matching the pattern used in other Stork chain examples.
