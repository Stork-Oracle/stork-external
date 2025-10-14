# Initia MiniMove Stork SDK Example
This is a very simple Initia MiniMove project to show how you would use the Stork Initia MiniMove contract to consume Stork price updates in your Initia MiniMove program.

## Deploy locally
1. Deploy a local version of the [Stork contract](../contracts)
2. Initialize the contract and write some data to it for your desired asset id using the cli in [admin.ts](../cli/admin.ts)
3. Set the stork address in the example [Move.toml](./Move.toml)
4. Install dependencies in the app directory:
```bash
cd app && npm install
```
5. Compile and deploy this example contract:
```bash
minitiad move deploy-object example --from <key_name> --gas auto --gas-adjustment 1.5 --gas-prices <gas_price> --node <rpc_url> --chain-id <chain_id> --language-version 2.1
```
6. Read the price from the Stork feed using the cli in [example.ts](./app/example.ts):
```bash
EXAMPLE_PACKAGE_ADDRESS=<package_address> MNEMONIC="<mnemonic>" RPC_URL=<rpc_url> CHAIN_ID=<chain_id> npx ts-node ./app/example.ts read-price <asset_id>
```

Example:
```bash
EXAMPLE_PACKAGE_ADDRESS=0x123... MNEMONIC="your mnemonic here" RPC_URL=https://rest.testnet.initia.xyz CHAIN_ID=initiation-2 npx ts-node ./app/example.ts read-price INITUSD
```
