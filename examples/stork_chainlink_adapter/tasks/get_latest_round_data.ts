import { task } from "hardhat/config";

// An example of a script to interact with the contract
async function getLatestRoundData(taskArgs: { exampleContractAddress: string, priceId: string }, hre: any) {
    // @ts-expect-error artifacts is loaded in hardhat/config
    const contractArtifact = await artifacts.readArtifact("ExampleStorkChainlinkAdapter");

    // @ts-expect-error ethers is loaded in hardhat/config
    // Initialize contract instance for interaction
    const contract = new ethers.Contract(
        taskArgs.exampleContractAddress,
        contractArtifact.abi,
        hre.ethers.provider
    );

    let latestRoundData = await contract.latestRoundData();

    console.log(`Latest round data: ${latestRoundData}`)
}


task("get_latest_round_data", "Gets latest price from the ExampleStorkChainlinkAdapter contract")
    .addParam("exampleContractAddress", "The address of the ExampleStorkChainlinkAdapter contract")
    .setAction(getLatestRoundData);