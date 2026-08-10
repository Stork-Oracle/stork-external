import type { HardhatRuntimeEnvironment } from "hardhat/types/hre";
import type { Address } from "viem";

interface SignerAddressesArguments {
  contractAddress: string;
}

export default async function (
  { contractAddress }: SignerAddressesArguments,
  hre: HardhatRuntimeEnvironment
) {
  const { viem } = await hre.network.connect();
  const publicClient = await viem.getPublicClient();

  const contractArtifact = await hre.artifacts.readArtifact("UpgradeableStorkFast");

  console.log(`Contract: ${contractAddress}`);

  const addresses = await publicClient.readContract({
    address: contractAddress as Address,
    abi: contractArtifact.abi,
    functionName: "getSignerAddresses",
    args: [],
  });

  console.log(`Signer Addresses: ${(addresses as Address[]).join(", ")}`);
}
