import { utils, Wallet } from "zksync-ethers";
import { HardhatRuntimeEnvironment } from "hardhat/types";
import { Deployer } from '@matterlabs/hardhat-zksync-deploy/dist/deployer';


import { vars } from "hardhat/config";

import * as hre from "hardhat";
import { artifacts } from "hardhat";
import { loadContractDeploymentAddress } from "../tasks/utils/helpers";

const DEPLOYMENT = "Stork#UpgradeableStorkZK";

// An example of a deploy script that will deploy and call a simple contract.
const interact = async function (hre: HardhatRuntimeEnvironment) {
  console.log(`Running interact script`);

  const contractAddress = await loadContractDeploymentAddress(DEPLOYMENT);
  if (!contractAddress) {
    throw new Error(
      "Contract address not found. Please deploy the contract first."
    );
  }
  console.log(`Contract: ${contractAddress}`);

  const [interacter] = await hre.ethers.getSigners();

  const contractArtifact = await artifacts.readArtifact("UpgradeableStorkZK");

  const contract = new hre.ethers.Contract(
    contractAddress,
    contractArtifact.abi,
    interacter // Interact with the contract on behalf of this wallet
  );

  const params = utils.getPaymasterParams(
    "0x950e3Bb8C6bab20b56a70550EC037E22032A413e", // Paymaster address
    {
      type: "General",
      innerInput: new Uint8Array(),
    }
  );

//   console.log(await contract.updateSingleUpdateFeeInWei.send(0, {
//     customData: {
//         gasPerPubdata: utils.DEFAULT_GAS_PER_PUBDATA_LIMIT,
//         paymasterParams: params,
//       }
//   }))

    await contract.updateTemporalNumericValuesV1.send([{
        temporalNumericValue: {
            timestampNs: "1725636529768330580",
            quantizedValue: "54225426073500002000000"
        },
        id: "0x7404e3d104ea7841c3d9e6fd20adfe99b4ad586bc08d8f3bd3afef894cf184de",
        publisherMerkleRoot: "0x2ff65fb2ac4f11fb5e1ee43444edfd1b447ae45ea67d89be495a7cec88bbd243",
        valueComputeAlgHash: "0x9be7e9f9ed459417d96112a7467bd0b27575a2c7847195c68f805b70ce1795ba",
        r: "0xcbd81a4c9d6730d6ff53b1ad816379216e6fcdcb517a64210f38730e7e5bed57",
        s: "0x4ee4f3ef1443f2c729b0adadcb5ede04dcd70899e5a0ea7d56c0c04c33f4c8b1",
        v: "0x1c"
    }], {
        value: 0,
        customData: {
            gasPerPubdata: utils.DEFAULT_GAS_PER_PUBDATA_LIMIT,
            paymasterParams: params,
        }
    }).catch((e) => {
        console.log(e.info)
        });

    const returnVal = await contract.getTemporalNumericValueV1("0x7404e3d104ea7841c3d9e6fd20adfe99b4ad586bc08d8f3bd3afef894cf184de");
    // const returnVal = await contract.storkPublicKey();
    console.log(`Current value: ${returnVal}`);
}

interact(hre)
  .then(() => process.exit(0))
  .catch(error => {
    console.error(error);
    process.exit(1);
  });
