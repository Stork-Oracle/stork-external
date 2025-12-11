# Stork EVM SDK Example

This is a very simple Hardhat project to show how you would use the Stork Fast EVM SDK to consume Stork Fast Signed ECDSA update payloads in your Solidity contract.

## Deploy locally

1. Deploy a local version of the [Stork Fast contract](../../contracts/stork_fast) 
2. Initialize the contract and write some data to it for your desired asset id using the CLI tools
3. Update the [ignition module parameters](ignition/parameters.json) to pass your local Stork Fast contract's address as the `storkFastContractAddress` parameter
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
npx hardhat --network hardhatMainnet ignition deploy ignition/modules/Example.ts --parameters ignition/parameters.json
```
7. Use the Stork Fast contract to verify a signed ECDSA payload:
```bash
npx hardhat --network hardhatMainnet use-stork-fast 0xYOUR_EXAMPLE_CONTRACT_ADDRESS 0xYOUR_SIGNED_PAYLOAD 
```
