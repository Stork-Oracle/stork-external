import { task } from "hardhat/config";

// An example of a script to interact with the contract
async function getLatestPrice(taskArgs: { exampleContractAddress: string, priceId: string }, hre: any) {
    // @ts-expect-error artifacts is loaded in hardhat/config
    const contractArtifact = await artifacts.readArtifact("ExampleStorkPythAdapter");

    // @ts-expect-error ethers is loaded in hardhat/config
    // Initialize contract instance for interaction
    const contract = new ethers.Contract(
        taskArgs.exampleContractAddress,
        contractArtifact.abi,
        hre.ethers.provider
    );

    let latestPrice = await contract.latestPrice(taskArgs.priceId);

    console.log(`Latest round data: ${latestPrice}`)
}


task("get_latest_price", "Gets latest price from the ExampleStorkPythAdapter contract")
    .addParam("exampleContractAddress", "The address of the ExampleStorkPythAdapter contract")
    .addParam("priceId", "The encoded asset id to pull price for")
    .setAction(getLatestPrice);