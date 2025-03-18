import { task } from "hardhat/config";

task("get-address", "Gets the address for the current signer")
  .setAction(async (args, hre) => {
    const [signer] = await hre.ethers.getSigners();
    const address = await signer.getAddress();
    const balance = await hre.ethers.provider.getBalance(address);
    
    console.log("Address:", address);
    console.log("Balance:", hre.ethers.formatEther(balance), "QUAI");
}); 