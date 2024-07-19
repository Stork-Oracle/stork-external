import { task } from "hardhat/config";

const STORK_PUBLIC_KEY = "0xC4A02e7D370402F4afC36032076B05e74FF81786"
const VALID_TIMEOUT_SECONDS = 6000000000000;
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

  console.log("UpgradeableStork deployed to:", upgradeableStork.target);
}

task("deploy", "A task to deploy the contract")
  .setAction(main);
