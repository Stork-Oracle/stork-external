import type { HardhatRuntimeEnvironment } from "hardhat/types/hre";
import type { Address } from "viem";

interface UpgradeArguments {
  proxyAddress: string;
}

export default async function (
  { proxyAddress }: UpgradeArguments,
  hre: HardhatRuntimeEnvironment
) {
  const { viem } = await hre.network.connect();
  const publicClient = await viem.getPublicClient();
  const [walletClient] = await viem.getWalletClients();

  console.log(`Upgrading proxy at: ${proxyAddress}`);
  console.log("Deploying new UpgradeableStorkFast implementation...");

  const newImplementation = await viem.deployContract("UpgradeableStorkFast");
  console.log(`New implementation deployed at: ${newImplementation.address}`);

  console.log("Upgrading proxy to new implementation...");

  const proxyAdminAbi = (await hre.artifacts.readArtifact("ProxyAdmin")).abi;
  const proxyAbi = (await hre.artifacts.readArtifact("TransparentUpgradeableProxy")).abi;

  // Get the ProxyAdmin address from the proxy
  const adminAddress = await publicClient.readContract({
    address: proxyAddress as Address,
    abi: proxyAbi,
    functionName: "admin",
  });

  console.log(`ProxyAdmin address: ${adminAddress}`);

  // Upgrade via ProxyAdmin
  const hash = await walletClient.writeContract({
    address: adminAddress as Address,
    abi: proxyAdminAbi,
    functionName: "upgradeAndCall",
    args: [proxyAddress as Address, newImplementation.address, "0x"],
  });

  console.log(`Upgrade transaction hash: ${hash}`);
  console.log("Waiting for confirmation...");

  const receipt = await publicClient.waitForTransactionReceipt({ hash });

  if (receipt.status === "success") {
    console.log("✓ Proxy upgraded successfully");
  } else {
    console.error("✗ Upgrade failed");
    throw new Error("Upgrade failed");
  }
}
