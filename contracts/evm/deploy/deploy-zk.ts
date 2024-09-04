import { utils, Wallet } from "zksync-ethers";
import * as ethers from "ethers";
import { HardhatRuntimeEnvironment } from "hardhat/types";
import { Deployer } from "@matterlabs/hardhat-zksync";

const STORK_PUBLIC_KEY = "0x0a803F9b1CCe32e2773e0d2e98b37E0775cA5d44";
const VALID_TIMEOUT_SECONDS = 3600;
const UPDATE_FEE_IN_WEI = 1;

import { vars } from "hardhat/config";

// An example of a deploy script that will deploy and call a simple contract.
export default async function (hre: HardhatRuntimeEnvironment) {
  console.log(`Running deploy script`);

  // Initialize the wallet.
  const wallet = new Wallet(vars.get("PRIVATE_KEY"));

  // Create deployer object and load the artifact of the contract we want to deploy.
  const deployer = new Deployer(hre, wallet);
  // Load contract

  const artNames = await hre.artifacts.getAllFullyQualifiedNames();
  const paths = await hre.artifacts.getArtifactPaths();
  console.log(artNames);
  console.log(paths);
  const artifact = await deployer.loadArtifact("UpgradeableStork");

  const params = utils.getPaymasterParams(
    "0x950e3Bb8C6bab20b56a70550EC037E22032A413e", // Paymaster address
    {
      type: "General",
      innerInput: new Uint8Array(),
    }
  );

  const deployed = await hre.zkUpgrades.deployProxy(
    deployer.zkWallet,
    artifact,
    [
      "0x0a803F9b1CCe32e2773e0d2e98b37E0775cA5d44",
      STORK_PUBLIC_KEY,
      VALID_TIMEOUT_SECONDS,
      UPDATE_FEE_IN_WEI,
    ],
    { 
      initializer: "initialize",
      kind: "uups",
    }
  );

  await deployed.waitForDeployment();
  console.log("deployed to:", await deployed.getAddress());
}
