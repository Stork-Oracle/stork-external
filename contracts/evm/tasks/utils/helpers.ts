import fs from "fs";
import path from "path";

export const CONTRACT_DEPLOYMENT = "Stork#UpgradeableStork";

// Function to ensure directory exists
export function ensureDirectoryExistence(filePath: string) {
  const dirname = path.dirname(filePath);
  if (fs.existsSync(dirname)) {
    return true;
  }
  ensureDirectoryExistence(dirname);
  fs.mkdirSync(dirname);
}

// Function to create file if it doesn't exist
export function createFileIfNotExists(filePath: string, content = "") {
  ensureDirectoryExistence(filePath);

  if (!fs.existsSync(filePath)) {
    fs.writeFileSync(filePath, content);
    console.log(`File created: ${filePath}`);
  } else {
    console.log(`File already exists: ${filePath}`);
  }
}

export async function getDeployedAddressesPath() {
  // @ts-expect-error ethers is loaded in hardhat/config
  const chainId = await ethers.provider.getNetwork().then(({ chainId }) => chainId);
  return path.join(__dirname, "..", "..", "deployments", `chain-${chainId}`, "deployed_addresses.json");
}

export async function loadContractDeploymentAddress() {
  const deployedAddressPath = await getDeployedAddressesPath();
  const deployedAddresses = require(deployedAddressPath);
  return deployedAddresses[CONTRACT_DEPLOYMENT];
}
