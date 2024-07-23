import { task } from "hardhat/config";

// An example of a script to interact with the contract
async function main(contractAddress: string) {
  console.log(`Running script to interact with contract ${contractAddress}`);

  // @ts-expect-error ethers is loaded in hardhat/config
  const [deployer] = await ethers.getSigners();

  // @ts-expect-error artifacts is loaded in hardhat/config
  const contractArtifact = await artifacts.readArtifact("UpgradeableStork");

  // @ts-expect-error ethers is loaded in hardhat/config
  // Initialize contract instance for interaction
  const contract = new ethers.Contract(
    contractAddress,
    contractArtifact.abi,
    deployer // Interact with the contract on behalf of this wallet
  );

  const version = await contract.version();
  console.log(`Contract version: ${version}`);

  const result = await contract.updateTemporalNumericValuesV1([
    {
      temporalNumericValue: {
        timestampNs: "3",
        quantizedValue: "60000000000000000000000",
      },
      // @ts-expect-error
      id: ethers.keccak256(ethers.toUtf8Bytes("BTCUSD")),
      // @ts-expect-error
      publisherMerkleRoot: ethers.encodeBytes32String("example data"),
      // @ts-expect-error
      valueComputeAlgHash: ethers.encodeBytes32String("example data"),
      r: "0x3e42e45aadf7da98780de810944ac90424493395c90bf0c21ede86b0d3c2cd7b",
      s: "0x1d853d65ae5be6046dc4199de2a0ee2b7288f51fc4af6946746c425cb8649879",
      v: "0x1c"
    }
  ], { value: 1 });

  console.log(`Contract version: ${result}`);

  // @ts-expect-error
  const value = await contract.getTemporalNumericValueV1(ethers.keccak256(ethers.toUtf8Bytes("BTCUSD")))

  console.log(`Value for BTCUSD: ${value}`);
}

task("interact", "A task to interact with the proxy contract")
  .addParam("proxyAddress", "The address of the proxy contract")
  .setAction(async (taskArgs) => await main(taskArgs.proxyAddress));
