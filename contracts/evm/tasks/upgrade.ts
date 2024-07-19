import { task } from "hardhat/config";

async function main(contractAddress: string) {
  // @ts-expect-error ethers is loaded in hardhat/config
  const UpgradeableStork = await ethers.getContractFactory(
    "UpgradeableStork"
  );

  // @ts-expect-error upgrades is loaded in hardhat/config
  // Upgrade the proxy to the new implementation
  const upgraded = await upgrades.upgradeProxy(
    contractAddress,
    UpgradeableStork
  );

  console.log("UpgradeableStork upgraded to:", await upgraded.getAddress());
}

task("upgrade", "A task to upgrade the proxy contract")
    .addParam("proxyAddress", "The address of the proxy contract")
    .setAction(async (taskArgs) => await main(taskArgs.proxyAddress));
