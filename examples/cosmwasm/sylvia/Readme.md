# Stork Sylvia Cosmwasm Example

This is a very simple Sylvia project to show how you would use the Stork Cosmwasm Contract to consume Stork price updates in your Sylvia contract.

### Deploy for testing

1. Deploy a testing version of the [Stork Contract](../../contracts/cosmwasm) to your environment, or [check for an official deployment](https://docs.stork.network/resources/contract-addresses/cosmwasm).
2. Instantiate the Stork contract and write some data to it for your desired asset id using the cli in [admin.ts](../../contracts/cosmwasm/cli/admin.ts)
3. Compile and deploy this example contract.
5. Instantiate the example contract with the address of the Stork Contract, then read the price from the Stork feed using the cli in [example.ts](../app/example.ts)