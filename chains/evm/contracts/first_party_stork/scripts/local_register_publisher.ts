import { ethers } from "ethers";

async function main(): Promise<void> {
  const rpcUrl = process.env.RPC_URL;
  const privateKey = process.env.PRIVATE_KEY;
  const contractAddress = process.env.CONTRACT_ADDRESS;
  const publisherEvmPublicKey = process.env.PUBLISHER_EVM_PUBLIC_KEY;

  if (!rpcUrl) {
    throw new Error("RPC_URL is not set");
  }
  if (!privateKey) {
    throw new Error("PRIVATE_KEY is not set");
  }
  if (!contractAddress) {
    throw new Error("CONTRACT_ADDRESS is not set");
  }
  if (!publisherEvmPublicKey) {
    throw new Error("PUBLISHER_EVM_PUBLIC_KEY is not set");
  }

  const provider = new ethers.JsonRpcProvider(rpcUrl, 31337);
  const wallet = new ethers.Wallet(privateKey, provider);

  const abi = [
    "function createPublisherUser(address publisher, uint256 publisherId) returns (uint256)"
  ];

  const contract = new ethers.Contract(contractAddress, abi, wallet);

  const tx = await contract.createPublisherUser(publisherEvmPublicKey, 0n);
  console.log("Sent transaction:", tx.hash);
  const receipt = await tx.wait();
  console.log(
    `Publisher registered successfully! Transaction mined in block: ${receipt.blockNumber}`
  );
}

main().catch((err) => {
  console.error("Failed to register publisher:", err);
  process.exit(1);
});


