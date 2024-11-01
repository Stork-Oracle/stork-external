# Stork Chainlink Adapter Contract Example
This is a very simple hardhat project to show how you would use the Stork Chainlink Adapter in your Solidity contract.

## Deploy locally
1. Deploy a local version of the [Stork contract](../../contracts/evm) and write some data to it for your desired asset id
2. Update the [ignition module](ignition/modules/stork_chainlink_adapter.ts) to pass your local Stork contract's address as an argument
3. Compile the contract:
```
npx hardhat compile
```
4. Deploy the contract locally and record the address
```
npx hardhat --network localhost ignition deploy ignition/modules/stork_chainlink_adapter.ts
```
5. Get the latest data from your contract:
```
 npx hardhat --network localhost get_latest_round_data --example-contract-address YOUR_LOCAL_EXAMPLE_CONTRACT_ADDRESS   
```