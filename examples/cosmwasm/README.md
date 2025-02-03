# Stork Cosmwasm Example

This directory contains two example contracts that consume Stork price updates. 

1. [Sylvia Example](sylvia) - A contract built with the Sylvia framework.
2. [Cosmwasm Core Example](cosmwasm_core) - A contract built with vanilla cosmwasm.

Both examples have the same interface, so you can use the same [CLI](../app) to interact with both contracts.

### Deploy for testing

For both examples:

1. Deploy a testing version of the [Stork Contract](../../contracts/cosmwasm) to your environment, or [check for an official deployment](https://docs.stork.network/resources/contract-addresses/cosmwasm).
2. Instantiate the Stork contract and write some data to it for your desired asset id using the cli in [admin.ts](../../contracts/cosmwasm/cli/admin.ts)
3. Compile and deploy this example contract. We recommend using a similar method as found [here](../../contracts/cosmwasm/README.md).
4. Instantiate the example contract with the address of the Stork Contract, then read the price from the Stork feed using the cli in [example.ts](../app/example.ts)

