import type { HardhatRuntimeEnvironment } from "hardhat/types/hre";
import { Address, hexToBytes, Hex, toHex } from "viem";

interface UseStorkFastTaskArguments {
    exampleContractAddress: string;
    payload: string;
    fee?: string;
}

export default async function (
    taskArguments: UseStorkFastTaskArguments,
    hre: HardhatRuntimeEnvironment,
) {
    const { viem } = await hre.network.connect();
    const publicClient = await viem.getPublicClient();
    const [walletClient] = await viem.getWalletClients();

    const exampleContractArtifact = await hre.artifacts.readArtifact("Example");

    const payload = taskArguments.payload as Hex;

    console.log(`Payload: ${payload}`);
    const hash = await walletClient.writeContract({
        address: taskArguments.exampleContractAddress as Address,
        abi: exampleContractArtifact.abi,
        functionName: "useStorkFast",
        args: [payload],
        value: taskArguments.fee ? BigInt(taskArguments.fee) : BigInt(2),
    });

    console.log(`Transaction hash: ${hash}`);
    console.log("Waiting for confirmation...");

    const receipt = await publicClient.waitForTransactionReceipt({ hash });

    if (receipt.status === "success") {
        console.log("✓ Transaction successful");
    } else {
        console.error("✗ Transaction failed");
        throw new Error("Transaction failed");
    }
}
