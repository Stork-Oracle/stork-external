import { task } from "hardhat/config";
import { CONTRACT_DEPLOYMENT, createFileIfNotExists, getDeployedAddressesPath } from "./utils/helpers";

const STORK_PUBLIC_KEY = "0x0a803F9b1CCe32e2773e0d2e98b37E0775cA5d44"
const VALID_TIMEOUT_SECONDS = 3600;
const UPDATE_FEE_IN_WEI = 1;

async function main() {
  // @ts-expect-error ethers is loaded in hardhat/config
  const [deployer] = await ethers.getSigners();

  // @ts-expect-error artifacts is loaded in hardhat/config
  const UpgradeableStork = await ethers.getContractFactory(
    "UpgradeableStork"
  );

  // @ts-expect-error upgrades is loaded in hardhat/config
  const upgradeableStork = await upgrades.deployProxy(
    UpgradeableStork,
    [deployer.address, STORK_PUBLIC_KEY, VALID_TIMEOUT_SECONDS, UPDATE_FEE_IN_WEI],
    {
      initializer: "initialize",
      kind: "uups",
    }
  );

  await upgradeableStork.deploymentTransaction().wait();

  // write file to store the address at deployments/chain-<chainid>/deployed_addresses.json
  const deployedAddressPath = await getDeployedAddressesPath();
  createFileIfNotExists(deployedAddressPath, JSON.stringify({ [CONTRACT_DEPLOYMENT]: upgradeableStork.target }, null, 2));

  console.log("UpgradeableStork deployed to:", upgradeableStork.target);
}

task("deploy", "A task to deploy the contract")
  .setAction(main);
