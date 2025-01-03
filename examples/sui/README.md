# Stork Sui SDK Example
This is a very simple Sui project to show how you would use the Stork Sui program to consume Stork price updates in your Sui program.

## Deploy locally
1. Deploy a local version of the [Stork contract](../../contracts/sui) 
2. Initialize the contract and write some data to it for your desired asset id using the cli in [admin.ts](../../contracts/sui/cli/admin.ts)
3. Compile and deploy this example contract.
```bash
sui move build
```
4. Deploy the contract locally
```bash
sui client publish
```
5. Read the price from the Stork feed using the cli in [example.ts](example.ts)
```bash
EXAMPLE_PACKAGE_ADDRESS=<package_address> RPC_ALIAS=<rpc_alias> SUI_KEY_ALIAS=<key_alias> npx ts-node ./app/example.ts read-price <asset_id> <stork_contract_address>
```