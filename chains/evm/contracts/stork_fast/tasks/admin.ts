import { scope } from "hardhat/config";
import type { HardhatRuntimeEnvironment } from "hardhat/types";
import type { Address } from "viem";
import { formatEther } from "viem";

const initializeContract = async (
  hre: HardhatRuntimeEnvironment,
  contractAddress: string
) => {
  const { viem, artifacts } = hre;
  const publicClient = await viem.getPublicClient();
  const [walletClient] = await viem.getWalletClients();

  const contractArtifact = await artifacts.readArtifact("UpgradeableStorkFast");

  console.log(`Network: ${hre.network.name}`);
  console.log(`Contract: ${contractAddress}`);

  return { publicClient, walletClient, abi: contractArtifact.abi };
};

const adminScope = scope("admin", "Admin tasks for StorkFast contract");

adminScope
  .task("verificationFeeInWei", "Get the verification fee in wei")
  .addPositionalParam<string>(
    "contractAddress",
    "The UpgradeableStorkFast contract address"
  )
  .setAction(
    async (
      { contractAddress }: { contractAddress: string },
      hre: HardhatRuntimeEnvironment
    ) => {
      const { publicClient, abi } = await initializeContract(
        hre,
        contractAddress
      );

      const fee = await publicClient.readContract({
        address: contractAddress as Address,
        abi,
        functionName: "verificationFeeInWei",
      });

      console.log(`Verification Fee: ${fee} wei (${formatEther(fee)} ETH)`);
    }
  );

adminScope
  .task("storkFastAddress", "Get the Stork Fast address")
  .addPositionalParam<string>(
    "contractAddress",
    "The UpgradeableStorkFast contract address"
  )
  .setAction(
    async (
      { contractAddress }: { contractAddress: string },
      hre: HardhatRuntimeEnvironment
    ) => {
      const { publicClient, abi } = await initializeContract(
        hre,
        contractAddress
      );

      const address = await publicClient.readContract({
        address: contractAddress as Address,
        abi,
        functionName: "storkFastAddress",
      });

      console.log(`Stork Fast Address: ${address}`);
    }
  );

adminScope
  .task("updateVerificationFeeInWei", "Update the verification fee in wei")
  .addPositionalParam<string>(
    "contractAddress",
    "The UpgradeableStorkFast contract address"
  )
  .addPositionalParam<string>("fee", "The new fee in wei")
  .setAction(
    async (
      { contractAddress, fee }: { contractAddress: string; fee: string },
      hre: HardhatRuntimeEnvironment
    ) => {
      const { publicClient, walletClient, abi } = await initializeContract(
        hre,
        contractAddress
      );

      const feeInWei = BigInt(fee);
      console.log(
        `Setting verification fee to ${feeInWei} wei (${formatEther(feeInWei)} ETH)...`
      );

      const hash = await walletClient.writeContract({
        address: contractAddress as Address,
        abi,
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
  );

adminScope
  .task("updateStorkFastAddress", "Update the Stork Fast address")
  .addPositionalParam<string>(
    "contractAddress",
    "The UpgradeableStorkFast contract address"
  )
  .addPositionalParam<string>("address", "The new Stork Fast address")
  .setAction(
    async (
      {
        contractAddress,
        address,
      }: { contractAddress: string; address: string },
      hre: HardhatRuntimeEnvironment
    ) => {
      const { publicClient, walletClient, abi } = await initializeContract(
        hre,
        contractAddress
      );

      console.log(`Setting Stork Fast address to ${address}...`);

      const hash = await walletClient.writeContract({
        address: contractAddress as Address,
        abi,
        functionName: "updateStorkFastAddress",
        args: [address as Address],
      });

      console.log(`Transaction hash: ${hash}`);
      console.log("Waiting for confirmation...");

      const receipt = await publicClient.waitForTransactionReceipt({ hash });

      if (receipt.status === "success") {
        console.log("✓ Stork Fast address updated successfully");
      } else {
        console.error("✗ Transaction failed");
        throw new Error("Transaction failed");
      }
    }
  );
