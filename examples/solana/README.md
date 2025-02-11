# Stork Solana SDK Example
This is a very simple Anchor project to show how you would use the Stork Solana SDK to consume Stork price updates in your Anchor program.

## Deploy locally
1. Deploy a local version of the [Stork contract](../../contracts/solana) 
2. Initialize the contract and write some data to it for your desired asset id using the cli in [admin.ts](../../contracts/solana/app/admin.ts)
3. Compile and deploy this example contract.
```
anchor build
```
4. Deploy the contract locally
```
anchor deploy
```
5. Read the price from the Stork feed using the cli in [example.ts](app/example.ts)
```
ANCHOR_PROVIDER_URL=http://localhost:8899 ANCHOR_WALLET=~/.config/solana/id.json npx ts-node ./app/example.ts read-price <encoded_asset_id>
```