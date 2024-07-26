import { task } from "hardhat/config";
import { loadContractDeploymentAddress } from "./utils/helpers";

async function main() {
  // @ts-expect-error ethers is loaded in hardhat/config
  const UpgradeableStork = await ethers.getContractFactory(
    "UpgradeableStork"
  );

  const contractAddress = await loadContractDeploymentAddress();
  if (!contractAddress) {
    throw new Error("Contract address not found. Please deploy the contract first.");
  }

  // @ts-expect-error upgrades is loaded in hardhat/config
  // Upgrade the proxy to the new implementation
  const upgraded = await upgrades.upgradeProxy(
    contractAddress,
    UpgradeableStork
  );

  console.log("UpgradeableStork upgraded to:", await upgraded.getAddress());
}

task("upgrade", "A task to upgrade the proxy contract")
    .setAction(async () => await main());
