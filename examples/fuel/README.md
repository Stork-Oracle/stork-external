# Stork Fuel Example

This is a very simple Fuel project to show how you would use the Stork Sway SDK to consume Stork price updates in your Fuel program.

## Deploy locally
1. Deploy a local version of the [Stork contract](../../contracts/fuel) 
2. Initialize the contract and write some data to it for your desired asset id using the cli in [admin.ts](../../contracts/fuel/cli/admin.ts)
3. Compile and deploy this example contract.
(from `./app`)
```bash
npx fuels deploy
```
4. Read the price from Stork using the cli in [example.ts](./app/example.ts)
```bash
npx ts-node ./app/example.ts read-price <asset_id> <stork_contract_address>
```
