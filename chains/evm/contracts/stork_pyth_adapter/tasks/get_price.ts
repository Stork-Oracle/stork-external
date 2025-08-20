import { task } from "hardhat/config";

// An example of a script to interact with the contract
async function getPrice(taskArgs: { contractAddress: string, priceId: string }, hre: any) {
    // @ts-expect-error artifacts is loaded in hardhat/config
    const contractArtifact = await artifacts.readArtifact("StorkPythAdapter");

    // @ts-expect-error ethers is loaded in hardhat/config
    // Initialize contract instance for interaction
    const contract = new ethers.Contract(
        taskArgs.contractAddress,
        contractArtifact.abi,
        hre.ethers.provider
    );

    let price = await contract.getPrice(taskArgs.priceId);

    console.log(`Price data: ${price}`)
}


task("get_price", "Gets latest price from the StorkPythAdapter contract")
    .addParam("contractAddress", "The address of the StorkPythAdapter contract")
    .addParam("priceId", "The price ID to query")
    .setAction(getPrice);
