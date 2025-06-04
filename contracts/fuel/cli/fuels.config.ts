import { createConfig, FuelsConfig, DeployedData, Provider, Address, Wallet, Src14OwnedProxy } from 'fuels';
import fs from 'fs';

const PRIVATE_KEY: string | undefined = process.env.PRIVATE_KEY;
const PROVIDER_URL: string | undefined = process.env.PROVIDER_URL;

if (!PRIVATE_KEY) {
  throw new Error("PRIVATE_KEY is not set");
}

if (!PROVIDER_URL) {
  throw new Error("PROVIDER_URL is not set");
}

export default createConfig({
  contracts: ['../contracts/stork'],
  output: './types',
  privateKey: PRIVATE_KEY,
  providerUrl: PROVIDER_URL,
  forcBuildFlags: ['--release'],
  onDeploy: onDeploy,
});

/**
 * Check the docs:
 * https://docs.fuel.network/docs/fuels-ts/fuels-cli/config-file/
 */

// runs after a successful deploy to configure the proxy contract
async function onDeploy(config: FuelsConfig, data: DeployedData) {
  // get the contract id (which is the proxy contract) 
  const proxyContractId = data.contracts?.find(c => c.name === 'stork')?.contractId;
  if (!proxyContractId) {
    throw new Error("Proxy contract not found");
  }

  const provider = new Provider(PROVIDER_URL!);
  const wallet = Wallet.fromPrivateKey(PRIVATE_KEY!, provider);

  const proxyContract = new Src14OwnedProxy(proxyContractId, wallet);

  // get the implementation contract address
  let implAddress: string = ""
  try {
    const proxyTarget = await proxyContract.functions.proxy_target().get();
    implAddress = proxyTarget.value?.bits.toString() || "";
  } catch (e) {
    console.log("Proxy target not found");
    throw e;
  }

  if (implAddress === "") {
    throw new Error("Implementation contract address not found");
  }

  console.log("Implementation contract address:", implAddress);
  console.log("Proxy contract address:", proxyContractId);

  console.log("Writing implementation address to to contract-ids.json...")
  const contractIds = JSON.parse(fs.readFileSync("types/contract-ids.json", "utf8"));
  contractIds.impl = implAddress;
  fs.writeFileSync("types/contract-ids.json", JSON.stringify(contractIds, null, 2));
}
