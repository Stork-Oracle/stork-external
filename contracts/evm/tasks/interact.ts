import { task } from "hardhat/config";
import { loadContractDeploymentAddress } from "./utils/helpers";

const allowedCommands = [
  "version",
  "updateTemporalNumericValuesV1",
  "getTemporalNumericValueV1",
  "getTemporalNumericValueUnsafeV1",
  "updateValidTimePeriodSeconds",
  "validTimePeriodSeconds",
  "updateSingleUpdateFeeInWei",
  "singleUpdateFeeInWei",
  ,
  "updateStorkPublicKey",
  "storkPublicKey",
];

type AllowedCommands = (typeof allowedCommands)[number];

// An example of a script to interact with the contract
async function main(command: AllowedCommands, args: any) {
  if (!allowedCommands.includes(command)) {
    throw new Error(`Invalid command: ${command}`);
  }

  const contractAddress = await loadContractDeploymentAddress();
  if (!contractAddress) {
    throw new Error(
      "Contract address not found. Please deploy the contract first."
    );
  }
  console.log(`Contract: ${contractAddress}`);

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

  let encoded;

  let returnVal;
  switch (command) {
    case "version":
      const version = await contract.version();
      console.log(`Contract version: ${version}`);
      break;
    case "updateTemporalNumericValuesV1":
      const split = args.split(" ");

      const assetIds = split[0];
      const endpoint = split[1];
      const authKey = split[2];

      const response = await fetch(
        `${endpoint}/v1/prices/latest?assets=${assetIds}`,
        {
          headers: {
            Authorization: `Basic ${authKey}`,
          },
        }
      );

      const rawJson = await response.text();
      const safeJsonText = rawJson.replace(
        /(?<!["\d])\b\d{16,}\b(?!["])/g, // Regex to find large integers not already in quotes
        (match: any) => `"${match}"` // Convert large numbers to strings
      );

      const responseData = JSON.parse(safeJsonText);

      const updates = Object.keys(responseData.data).map((key: any) => {
        const data = responseData.data[key];

        return {
          temporalNumericValue: {
            timestampNs: data.stork_signed_price.timestamped_signature.timestamp,
            quantizedValue: data.stork_signed_price.price,
          },
          id: data.stork_signed_price.encoded_asset_id,
          publisherMerkleRoot: data.stork_signed_price.publisher_merkle_root,
          valueComputeAlgHash: "0x"+ data.stork_signed_price.calculation_alg.checksum,
          r: data.stork_signed_price.timestamped_signature.signature.r,
          s: data.stork_signed_price.timestamped_signature.signature.s,
          v: data.stork_signed_price.timestamped_signature.signature.v,
        };
      });

      const updateResult = await contract.updateTemporalNumericValuesV1(updates, {
        value: updates.length,
      });

      break;
    case "getTemporalNumericValueV1":
      // @ts-expect-error
      encoded = ethers.keccak256(ethers.toUtf8Bytes(args));
      returnVal = await contract.getTemporalNumericValueV1(encoded);
      console.log(returnVal);
      break;
    case "getTemporalNumericValueUnsafeV1":
      // @ts-expect-error
      encoded = ethers.keccak256(ethers.toUtf8Bytes(args));
      returnVal = await contract.getTemporalNumericValueUnsafeV1(encoded);
      console.log(returnVal);
      break;
    case "updateValidTimePeriodSeconds":
      returnVal = await contract.updateValidTimePeriodSeconds(parseInt(args));
      break;
    case "updateSingleUpdateFeeInWei":
      returnVal = await contract.updateSingleUpdateFeeInWei(parseInt(args));
      break;
    case "updateStorkPublicKey":
      returnVal = await contract.updateStorkPublicKey(args);
      break;
    case "validTimePeriodSeconds":
    case "singleUpdateFeeInWei":
    case "storkPublicKey":
      returnVal = await contract[command]();
      console.log(returnVal);
      break;
    default:
      throw new Error(`Invalid command: ${command}`);
  }
}

task("interact", "A task to interact with the proxy contract")
  .addPositionalParam<AllowedCommands>("command", "The command to run")
  .addPositionalParam<string>(
    "args",
    "The arguments for the command",
    undefined,
    undefined,
    true
  )
  .setAction(async (taskArgs) => await main(taskArgs.command, taskArgs.args));
