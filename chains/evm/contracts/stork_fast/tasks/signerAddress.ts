import type { HardhatRuntimeEnvironment } from "hardhat/types/hre";
import type { Address } from "viem";

interface SignerAddressArguments {
  contractAddress: string;
}

export default async function (
  { contractAddress }: SignerAddressArguments,
  hre: HardhatRuntimeEnvironment
) {
  const { viem } = await hre.network.connect();
  const publicClient = await viem.getPublicClient();

  const contractArtifact = await hre.artifacts.readArtifact("UpgradeableStorkFast");

  console.log(`Contract: ${contractAddress}`);

  const address = await publicClient.readContract({
    address: contractAddress as Address,
    abi: contractArtifact.abi,
    functionName: "signerAddress",
  });

  console.log(`Signer Address: ${address}`);
}
