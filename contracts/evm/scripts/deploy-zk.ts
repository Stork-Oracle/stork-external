import { utils, Wallet } from "zksync-ethers";
import * as ethers from "ethers";
import { HardhatRuntimeEnvironment } from "hardhat/types";
import { Deployer } from "@matterlabs/hardhat-zksync";

const STORK_PUBLIC_KEY = "0x0a803F9b1CCe32e2773e0d2e98b37E0775cA5d44"
const VALID_TIMEOUT_SECONDS = 3600;
const UPDATE_FEE_IN_WEI = 1;

// An example of a deploy script that will deploy and call a simple contract.
export default async function (hre: HardhatRuntimeEnvironment) {
  console.log(`Running deploy script`);

  // Initialize the wallet.
  const wallet = new Wallet("<WALLET-PRIVATE-KEY>");

  // Create deployer object and load the artifact of the contract we want to deploy.
  const deployer = new Deployer(hre, wallet);
  // Load contract
  const artifact = await deployer.loadArtifact("Greeter");

  // Deploy this contract. The returned object will be of a `Contract` type,
  // similar to the ones in `ethers`.
  const greeting = "Hi there!";
  // `greeting` is an argument for contract constructor.
  const greeterContract = await deployer.deploy(artifact, [greeting]);

  // Show the contract info.
  console.log(`${artifact.contractName} was deployed to ${await greeterContract.getAddress()}`);
}
