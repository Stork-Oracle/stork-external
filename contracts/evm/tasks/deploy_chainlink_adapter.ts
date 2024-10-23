import { task } from "hardhat/config";
import { createFileIfNotExists, getDeployedAddressesPath } from "./utils/helpers";


const STORK_CHAINLINK_ADAPTER_CONTRACT_DEPLOYMENT = "Stork#StorkChainlinkAdapter";

async function main(priceId: string, storkContractAddress: string) {
    // @ts-expect-error ethers is loaded in hardhat/config
    const priceIdBytes = ethers.encodeBytes32String(priceId);

    // @ts-expect-error ethers is loaded in hardhat/config
    const StorkChainlinkAdapterFactory = await ethers.getContractFactory(
        "StorkChainlinkAdapter"
    );

    const storkChainlinkAdapterContract = await StorkChainlinkAdapterFactory.deploy(storkContractAddress, priceIdBytes);

    // @ts-ignore
    await storkChainlinkAdapterContract.deploymentTransaction().wait();

    // todo: do we need to track this? or can we skip since we're not upgrading these
    // // write file to store the address at deployments/chain-<chainid>/deployed_addresses.json
    // const deployedAddressPath = await getDeployedAddressesPath();
    // createFileIfNotExists(deployedAddressPath, JSON.stringify({ [STORK_CHAINLINK_ADAPTER_CONTRACT_DEPLOYMENT]: storkChainlinkAdapterContract.target }, null, 2));

    console.log("StorkChainlinkAdapter deployed to:", storkChainlinkAdapterContract.target);
}

task("deploy-chainlink-adapter", "A task to deploy the contract")
    .addPositionalParam<string>("priceId", "The Price ID to pull prices for (e.g. BTCUSD)")
    .addPositionalParam<string>("storkContractAddress", "The stork contract address to pull prices from")
    .setAction(async (taskArgs) => await main(taskArgs.priceId, taskArgs.storkContractAddress));
