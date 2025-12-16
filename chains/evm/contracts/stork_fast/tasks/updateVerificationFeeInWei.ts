import type { HardhatRuntimeEnvironment } from "hardhat/types/hre";
import type { Address } from "viem";
import { formatEther } from "viem";

interface UpdateVerificationFeeInWeiArguments {
  contractAddress: string;
  fee: string;
}

export default async function (
  { contractAddress, fee }: UpdateVerificationFeeInWeiArguments,
  hre: HardhatRuntimeEnvironment
) {
  const { viem } = await hre.network.connect();
  const publicClient = await viem.getPublicClient();
  const [walletClient] = await viem.getWalletClients();

  const contractArtifact = await hre.artifacts.readArtifact("UpgradeableStorkFast");

  console.log(`Contract: ${contractAddress}`);

  const feeInWei = BigInt(fee);
  console.log(
    `Setting verification fee to ${feeInWei} wei (${formatEther(feeInWei)} ETH)...`
  );

  const hash = await walletClient.writeContract({
    address: contractAddress as Address,
    abi: contractArtifact.abi,
    functionName: "updateVerificationFeeInWei",
    args: [feeInWei],
  });

  console.log(`Transaction hash: ${hash}`);
  console.log("Waiting for confirmation...");

  const receipt = await publicClient.waitForTransactionReceipt({ hash });

  if (receipt.status === "success") {
    console.log("✓ Verification fee updated successfully");
  } else {
    console.error("✗ Transaction failed");
    throw new Error("Transaction failed");
  }
}
