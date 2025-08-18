import { task } from "hardhat/config";
import { keccak256, toUtf8Bytes } from "ethers";

// An example of a script to interact with the contract
async function readPrice(taskArgs: { 
  exampleContractAddress: string, 
  asset: string 
}, hre: any) {
  // @ts-expect-error artifacts is loaded in hardhat/config
  const contractArtifact = await artifacts.readArtifact("Example");

  // @ts-expect-error ethers is loaded in hardhat/config
  const [signer] = await ethers.getSigners();
  
  // Initialize contract instance for interaction
  const contract = new ethers.Contract(
    taskArgs.exampleContractAddress,
    contractArtifact.abi,
    signer
  );

  // Hash the asset string to get feed_id (same as other chains)
  const feedId = keccak256(toUtf8Bytes(taskArgs.asset));
  
  try {
    console.log(`Reading price for asset: ${taskArgs.asset}`);
    console.log(`Feed ID: ${feedId}`);
    
    // Call the contract function
    const tx = await contract.useStorkPrice(feedId);
    const receipt = await tx.wait();
    
    console.log(`Transaction hash: ${receipt.hash}`);
    
    // Find the event in the transaction receipt
    const event = receipt.logs.find((log: any) => {
      try {
        const parsed = contract.interface.parseLog(log);
        return parsed?.name === 'StorkPriceUsed';
      } catch {
        return false;
      }
    });
    
    if (event) {
      const parsed = contract.interface.parseLog(event);
      console.log(`Price: ${parsed?.args.value}`);
      console.log(`Timestamp: ${parsed?.args.timestamp}`);
    }
    
  } catch (error) {
    console.error("Error reading price:", error);
  }
}

task("read-price", "Reads price from Stork feed")
  .addParam("exampleContractAddress", "The address of the Example contract")
  .addParam("asset", "Asset identifier (will be hashed)")
  .setAction(readPrice);
