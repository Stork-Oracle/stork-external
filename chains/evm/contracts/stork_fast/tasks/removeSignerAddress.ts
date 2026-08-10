import type { HardhatRuntimeEnvironment } from "hardhat/types/hre";
import type { Address } from "viem";

interface RemoveSignerAddressArguments {
  contractAddress: string;
  address: string;
}

export default async function (
  { contractAddress, address }: RemoveSignerAddressArguments,
  hre: HardhatRuntimeEnvironment
) {
  const { viem } = await hre.network.connect();
  const publicClient = await viem.getPublicClient();
  const [walletClient] = await viem.getWalletClients();

  const contractArtifact = await hre.artifacts.readArtifact("UpgradeableStorkFast");

  console.log(`Contract: ${contractAddress}`);

  console.log(`Removing signer address ${address}...`);

  const hash = await walletClient.writeContract({
    address: contractAddress as Address,
    abi: contractArtifact.abi,
    functionName: "removeSignerAddress",
    args: [address as Address],
  });

  console.log(`Transaction hash: ${hash}`);
  console.log("Waiting for confirmation...");

  const receipt = await publicClient.waitForTransactionReceipt({ hash });

  if (receipt.status === "success") {
    console.log("✓ Signer address removed successfully");
  } else {
    console.error("✗ Transaction failed");
    throw new Error("Transaction failed");
  }
}
