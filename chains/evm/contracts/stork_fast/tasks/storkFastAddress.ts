import type { HardhatRuntimeEnvironment } from "hardhat/types/hre";
import type { Address } from "viem";

interface StorkFastAddressArguments {
  contractAddress: string;
}

export default async function (
  { contractAddress }: StorkFastAddressArguments,
  hre: HardhatRuntimeEnvironment
) {
  const { viem } = await hre.network.connect();
  const publicClient = await viem.getPublicClient();

  const contractArtifact = await hre.artifacts.readArtifact("UpgradeableStorkFast");

  console.log(`Contract: ${contractAddress}`);

  const address = await publicClient.readContract({
    address: contractAddress as Address,
    abi: contractArtifact.abi,
    functionName: "storkFastAddress",
  });

  console.log(`Stork Fast Address: ${address}`);
}
