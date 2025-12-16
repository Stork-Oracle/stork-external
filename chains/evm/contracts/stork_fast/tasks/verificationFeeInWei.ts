import type { HardhatRuntimeEnvironment } from "hardhat/types/hre";
import type { Address } from "viem";
import { formatEther } from "viem";

interface VerificationFeeInWeiArguments {
  contractAddress: string;
}

export default async function (
  { contractAddress }: VerificationFeeInWeiArguments,
  hre: HardhatRuntimeEnvironment
) {
  const { viem } = await hre.network.connect();
  const publicClient = await viem.getPublicClient();

  const contractArtifact = await hre.artifacts.readArtifact("UpgradeableStorkFast");

  console.log(`Contract: ${contractAddress}`);

  const fee = await publicClient.readContract({
    address: contractAddress as Address,
    abi: contractArtifact.abi,
    functionName: "verificationFeeInWei",
  });

  console.log(`Verification Fee: ${fee} wei (${formatEther(fee)} ETH)`);
}
