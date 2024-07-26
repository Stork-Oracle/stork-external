import { task } from "hardhat/config";
import { loadContractDeploymentAddress } from "./utils/helpers";

const allowedCommands = ["version", "updateTemporalNumericValuesV1", "getTemporalNumericValueV1"] as const;

type AllowedCommands = (typeof allowedCommands)[number];

// An example of a script to interact with the contract
async function main(command: AllowedCommands, args: any) {
  if (!allowedCommands.includes(command)) {
    throw new Error(`Invalid command: ${command}`);
  }

  const contractAddress = await loadContractDeploymentAddress();
  if (!contractAddress) {
    throw new Error("Contract address not found. Please deploy the contract first.");
  }
  console.log(`Running script to interact with contract ${contractAddress}`);

  // @ts-expect-error ethers is loaded in hardhat/config
  const [deployer] = await ethers.getSigners();

  // @ts-expect-error artifacts is loaded in hardhat/config
  const contractArtifact = await artifacts.readArtifact("UpgradeableStork");

  // @ts-expect-error ethers is loaded in hardhat/config
  // Initialize contract instance for interaction
  const contract = new ethers.Contract(
    contractAddress,
    contractArtifact.abi,
    deployer // Interact with the contract on behalf of this wallet
  );

  if (command === "version") {
    const version = await contract.version();
    console.log(`Contract version: ${version}`);
    return;
  } else if (command === "updateTemporalNumericValuesV1") {
    const payload = JSON.parse(args);
    const result = await contract.updateTemporalNumericValuesV1([payload], { value: 1 });
    return;
  } else if (command === "getTemporalNumericValueV1") {
    // @ts-expect-error
    const encoded = ethers.keccak256(ethers.toUtf8Bytes(args));
    const value = await contract.getTemporalNumericValueV1(encoded)
    console.log(`Value for BTCUSD: ${value}`);
    return;
  }
}

task("interact", "A task to interact with the proxy contract")
  .addPositionalParam<AllowedCommands>("command", "The command to run")
  .addPositionalParam<string>("args", "The arguments for the command", undefined, undefined, true)
  .setAction(async (taskArgs) => await main(taskArgs.command, taskArgs.args));
