import { task } from "hardhat/config";

async function deploy(taskArgs: { storkAddress: string, priceId: string }, hre: any) {
    const storkAddress = taskArgs.storkAddress;

    // @ts-expect-error ethers is loaded in hardhat/config
    const priceId = ethers.encodeBytes32String(taskArgs.priceId);
    console.log(`encoded price id: ${priceId}`)

    // Get the contract factory for ExampleStorkChainlinkAdapter
    const StorkChainlinkAdapter = await hre.ethers.getContractFactory("ExampleStorkChainlinkAdapter");

    // Deploy the contract with constructor arguments
    const storkChainlinkAdapter = await StorkChainlinkAdapter.deploy(storkAddress, priceId);
    const address = await storkChainlinkAdapter.getAddress()

    console.log("Contract deployed to address:", address);
}

task("deploy", "Deploys the ExampleStorkChainlinkAdapter contract")
    .addParam("storkAddress", "The address of the stork contract")
    .addParam("priceId", "The price ID for the contract")
    .setAction(deploy);
