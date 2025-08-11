import { task } from "hardhat/config";

// An example of a script to interact with the contract
async function getLatestRoundData(taskArgs: { contractAddress: string }, hre: any) {
    // @ts-expect-error artifacts is loaded in hardhat/config
    const contractArtifact = await artifacts.readArtifact("StorkChainlinkAdapter");

    // @ts-expect-error ethers is loaded in hardhat/config
    // Initialize contract instance for interaction
    const contract = new ethers.Contract(
        taskArgs.contractAddress,
        contractArtifact.abi,
        hre.ethers.provider
    );

    let latestRoundData = await contract.latestRoundData();

    console.log(`Latest round data: ${latestRoundData}`)
}


task("get_latest_round_data", "Gets latest price from the StorkChainlinkAdapter contract")
    .addParam("contractAddress", "The address of the StorkChainlinkAdapter contract")
    .setAction(getLatestRoundData);
