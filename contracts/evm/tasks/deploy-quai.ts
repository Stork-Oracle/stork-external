import { task } from "hardhat/config";
import { quais, Shard } from "quais";
import { createFileIfNotExists } from "./utils/helpers";
import { CONTRACT_DEPLOYMENT } from "./utils/helpers";
import { getDeployedAddressesPath } from "./utils/helpers";
const UpgradeableStorkJson = require("../artifacts/contracts/UpgradeableStork.sol/UpgradeableStork.json");

const STORK_PUBLIC_KEY = "0x0a803F9b1CCe32e2773e0d2e98b37E0775cA5d44"
const VALID_TIMEOUT_SECONDS = 3600;
const UPDATE_FEE_IN_WEI = 1;

async function deployProxy() {

    const provider = new quais.JsonRpcProvider(
        "https://rpc.quai.network", 
        undefined, 
        { usePathing: true }
    );
    const wallet = new quais.Wallet(hre.network.config.accounts[0], provider);
    console.log("Wallet:", wallet.address);
    const ipfsHash = await hre.deployMetadata.pushMetadataToIPFS("UpgradeableStork");

    console.log("Testing RPC connection...");
    const blockNumber = await provider.getBlockNumber(Shard.Cyprus1);
    console.log("Current block number:", blockNumber);
    const UpgradeableStork = new quais.ContractFactory(UpgradeableStorkJson.abi, UpgradeableStorkJson.bytecode, wallet, ipfsHash);
    const upgradeableStork = await quaiUpgrades.deployProxy(
        UpgradeableStork,
        [wallet.address, STORK_PUBLIC_KEY, VALID_TIMEOUT_SECONDS, UPDATE_FEE_IN_WEI],
        {
            initializer: "initialize",
            kind: "uups",
        }
    );

    console.log("Deployment transaction sent, waiting for confirmation...");
    console.log("Transaction hash:", upgradeableStork.deploymentTransaction()?.hash);

    await upgradeableStork.waitForDeployment();

    const deployedAddressPath = await getDeployedAddressesPath();
    createFileIfNotExists(deployedAddressPath, JSON.stringify({ [CONTRACT_DEPLOYMENT]: upgradeableStork.target }, null, 2));

    console.log("UpgradeableStork deployed to:", upgradeableStork.target);
}
task("deploy-quai", "A task to deploy the contract to Quai Network")
  .addOptionalPositionalParam("storkPublicKey", "The public key of the Stork contract")
  .setAction(deployProxy); 
