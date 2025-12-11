import type { HardhatRuntimeEnvironment } from "hardhat/types/hre";
import { Address, hexToBytes, Hex, toHex } from "viem";

interface UseStorkFastTaskArguments {
    exampleContractAddress: string;
    payload: string;
}

export default async function (
    taskArguments: UseStorkFastTaskArguments,
    hre: HardhatRuntimeEnvironment,
) {
    const { viem } = await hre.network.connect();
    const publicClient = await viem.getPublicClient();
    const [walletClient] = await viem.getWalletClients();

    const exampleContractArtifact = await hre.artifacts.readArtifact("Example");

    // Get the Stork Fast contract address from the Example contract
    const storkFastAddress = await publicClient.readContract({
        address: taskArguments.exampleContractAddress as Address,
        abi: exampleContractArtifact.abi,
        functionName: "storkFast",
    }) as Address;

    console.log(`Stork Fast contract: ${storkFastAddress}`);

    // Get the verification fee from the Stork Fast contract
    const verificationFee = await publicClient.readContract({
        address: storkFastAddress,
        abi: [
            {
                name: "verificationFeeInWei",
                type: "function",
                stateMutability: "view",
                inputs: [],
                outputs: [{ type: "uint256" }],
            },
        ],
        functionName: "verificationFeeInWei",
    }) as bigint;

    const feeToUse = taskArguments.fee ? BigInt(taskArguments.fee) : verificationFee;
    console.log(`Verification fee: ${verificationFee} wei`);
    console.log(`Using fee: ${feeToUse} wei`);

    const payload = taskArguments.payload as Hex;

    console.log(`Payload: ${payload}`);
    const hash = await walletClient.writeContract({
        address: taskArguments.exampleContractAddress as Address,
        abi: exampleContractArtifact.abi,
        functionName: "useStorkFast",
        args: [payload],
        value: feeToUse,
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
