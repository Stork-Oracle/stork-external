import type { HardhatRuntimeEnvironment } from "hardhat/types/hre";
import type { Address, Hex } from "viem";

interface VerifyPayloadArguments {
  contractAddress: string;
  payload: string;
}

export default async function (
  { contractAddress, payload }: VerifyPayloadArguments,
  hre: HardhatRuntimeEnvironment
) {
  const { viem } = await hre.network.connect();
  const publicClient = await viem.getPublicClient();
  const [walletClient] = await viem.getWalletClients();

  const contractArtifact = await hre.artifacts.readArtifact("UpgradeableStorkFast");

  console.log(`Contract: ${contractAddress}`);

  if (!payload.startsWith("0x")) {
    payload = `0x${payload}`;
  }

  const fee = (await publicClient.readContract({
    address: contractAddress as Address,
    abi: contractArtifact.abi,
    functionName: "verificationFeeInWei",
    args: [],
  })) as bigint;

  console.log(`Verification fee: ${fee} wei`);
  console.log("Calling verifyAndDeserializeSignedECDSAPayload...");

  const hash = await walletClient.writeContract({
    address: contractAddress as Address,
    abi: contractArtifact.abi,
    functionName: "verifyAndDeserializeSignedECDSAPayload",
    args: [payload as Hex],
    value: fee,
  });

  console.log(`Transaction hash: ${hash}`);
  console.log("Waiting for confirmation...");

  const receipt = await publicClient.waitForTransactionReceipt({ hash });

  if (receipt.status === "success") {
    console.log(`✓ Payload verified on-chain (gas used: ${receipt.gasUsed})`);
  } else {
    console.error("✗ Transaction reverted — payload failed verification");
    throw new Error("Transaction reverted");
  }
}
