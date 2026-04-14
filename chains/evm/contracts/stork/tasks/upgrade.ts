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
    UpgradeableStork,
    {
      pollingInterval: 1000,
      timeout: 60000,
      unsafeSkipStorageCheck: true,
    }
  );

  console.log("UpgradeableStork upgraded to:", await upgraded.getAddress());
}

task("upgrade", "A task to upgrade the proxy contract")
    .setAction(async () => await main());


task("prepare-upgrade", "Deploy a new implementation for use with either direct or safe upgrade")
  .setAction(async () => {
    const contractAddress = await loadContractDeploymentAddress();
    if (!contractAddress) {
      throw new Error("Contract address not found. Please deploy the contract first.");
    }

    // @ts-expect-error ethers is loaded in hardhat/config
    const factory = await ethers.getContractFactory("UpgradeableStork");

    // @ts-expect-error upgrades is loaded in hardhat/config
    const newImplAddress = await upgrades.prepareUpgrade(contractAddress, factory, {
      kind: "uups",
      unsafeSkipStorageCheck: true,
    });

    console.log(`New implementation: ${newImplAddress}`);
  });

task("apply-upgrade", "Upgrade the proxy to a prepared implementation (owner must be the deployer key)")
  .addPositionalParam("implAddress", "The new implementation address")
  .setAction(async ({ implAddress }: { implAddress: string }) => {
    const contractAddress = await loadContractDeploymentAddress();
    if (!contractAddress) {
      throw new Error("Contract address not found. Please deploy the contract first.");
    }

    // @ts-expect-error ethers is loaded in hardhat/config
    if (!ethers.isAddress(implAddress)) {
      throw new Error(`Invalid implementation address: ${implAddress}`);
    }

    // @ts-expect-error ethers is loaded in hardhat/config
    const bytecode = await ethers.provider.getCode(implAddress);
    if (bytecode === "0x") {
      throw new Error(`No contract found at implementation address ${implAddress}. Did you run prepare-upgrade first?`);
    }

    // @ts-expect-error artifacts is loaded in hardhat/config
    const contractArtifact = await artifacts.readArtifact("UpgradeableStork");
    // @ts-expect-error ethers is loaded in hardhat/config
    const iface = new ethers.Interface(contractArtifact.abi);
    const missingSelectors: string[] = [];
    for (const fragment of iface.fragments) {
      if (fragment.type === "function") {
        const selector = iface.getFunction(fragment.name)!.selector.slice(2);
        if (!bytecode.includes(selector)) {
          missingSelectors.push(`${fragment.name} (0x${selector})`);
        }
      }
    }
    if (missingSelectors.length > 0) {
      throw new Error(
        `Contract at ${implAddress} is missing expected UpgradeableStork selectors:\n` +
        missingSelectors.map(s => `  - ${s}`).join("\n") +
        "\nAborting upgrade."
      );
    }

    // @ts-expect-error ethers is loaded in hardhat/config
    const [deployer] = await ethers.getSigners();
    // @ts-expect-error ethers is loaded in hardhat/config
    const contract = new ethers.Contract(contractAddress, contractArtifact.abi, deployer);

    console.log(`Upgrading proxy ${contractAddress} to implementation ${implAddress}`);
    const tx = await contract.upgradeToAndCall(implAddress, "0x");
    const receipt = await tx.wait();
    console.log(`Upgraded to: ${implAddress}`);
    console.log(`Transaction: ${receipt.hash}`);
  });
