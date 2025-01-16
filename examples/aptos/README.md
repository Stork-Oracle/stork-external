# Aptos Stork SDK Example
This is a very simple Aptos project to show how you would use the Stork Aptos program to consume Stork price updates in your Aptos program.

## Deploy locally
1. Deploy a local version of the [Stork contract](../../contracts/aptos) 
2. Initialize the contract and write some data to it for your desired asset id using the cli in [admin.ts](../../contracts/aptos/cli/admin.ts)
3. Set the stork address in the example [Move.toml](./Move.toml)
3. Compile and deploy this example contract.
```bash
aptos move deploy-object --address-name example --profile <profile_name> --move-2
```
4. Read the price from the Stork feed using the cli in [example.ts](./app/example.ts)
```bash
EXAMPLE_PACKAGE_ADDRESS=<package_address> RPC_ALIAS=<rpc_alias> PRIVATE_KEY=<private_key> npx ts-node ./app/example.ts read-price <asset_id>
```