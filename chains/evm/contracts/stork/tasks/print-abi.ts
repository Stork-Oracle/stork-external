import { task } from "hardhat/config";

// An example of a script to interact with the contract
async function main() {
  // @ts-expect-error artifacts is loaded in hardhat/config
  const contractArtifact = await artifacts.readArtifact("UpgradeableStork");

  console.log(JSON.stringify(contractArtifact.abi, null, 2));
}

task("print-abi", "A task to print the ABI of the contract")
  .setAction(main);
