import { task } from "hardhat/config";
import { loadContractDeploymentAddress } from "./utils/helpers";

// before calling this, make sure you compile the contracts using `npx hardhat --network sophonTestnet compile --force`
// paymaster address is sophon paymaster address - 0x98546B226dbbA8230cf620635a1e4ab01F6A99B2

async function main({ paymasterAddress }: { paymasterAddress: string }) {
  // @ts-expect-error ethers is loaded in hardhat/config
  const UpgradeableStork = await ethers.getContractFactory(
    "UpgradeableStork"
  );

  const contractAddress = await loadContractDeploymentAddress();
  if (!contractAddress) {
    throw new Error("Contract address not found. Please deploy the contract first.");
  }

  let deployParams = {};
  if (paymasterAddress) {
    // @ts-expect-error zksyncEthers is loaded in hardhat/config
    const params = zksyncEthers.utils.getPaymasterParams(
        paymasterAddress,
        {
            type: 'General',
            innerInput: new Uint8Array(),
        },
    );

    deployParams = {
        paymasterParams: params,
    };
  }

//   @ts-expect-error zkUpgrades is loaded in hardhat/config
  const upgraded = await zkUpgrades.upgradeProxy(
    contractAddress,
    UpgradeableStork,
    deployParams,
  );

  console.log("UpgradeableStork upgraded to:", await upgraded.getAddress());
}

task("upgrade-zk", "A task to upgrade the proxy contract on zk")
    .addOptionalParam('paymasterAddress', 'The address of the paymaster')
    .setAction(main);
