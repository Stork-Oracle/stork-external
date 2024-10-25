# Stork Chainlink Adapter Contract Example
This is a very simple hardhat project to show how you would use the Stork Chainlink Adapter in your Solidity contract.

## Deploy locally
1. Deploy a local version of the [Stork contract](../../contracts/evm) and write some data to it for your desired asset id
2. Compile the contract:
```
npx hardhat compile
```
3. Deploy the contract locally and record the address
```
npx hardhat --network localhost deploy --stork-address YOUR_LOCAL_STORK_CONTRACT_ADDRESS --price-id YOUR_PRICE_ID
```
4. Get the latest data from your contract:
```
 npx hardhat --network localhost get_latest_round_data --example-contract-address YOUR_LOCAL_EXAMPLE_CONTRACT_ADDRESS   
```