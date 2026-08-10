// Use this ignition module to deploy the Stork Fast contract for the first time.

import { buildModule } from "@nomicfoundation/hardhat-ignition/modules";

const StorkFastProxyModule = buildModule("StorkFastProxyModule", (m) => {
  const implementation = m.contract("UpgradeableStorkFast", [], {
    id: "Implementation",
  });

  const initialOwner = m.getAccount(0);

  const initializeCalldata = m.encodeFunctionCall(implementation, "initialize", [initialOwner, m.getParameter("signerAddresses"), m.getParameter("verificationFeeInWei")]);

  // UUPS pattern: upgrades are performed by calling upgradeToAndCall on the
  // proxy itself, authorized by the contract owner via _authorizeUpgrade.
  const proxy = m.contract("ERC1967Proxy", [
    implementation,
    initializeCalldata,
  ]);

  return { proxy };
});

const StorkFastModule = buildModule("StorkFastModule", (m) => {
  const { proxy } = m.useModule(StorkFastProxyModule);

  const StorkFast = m.contractAt("UpgradeableStorkFast", proxy);

  return { StorkFast, proxy };
});

export default StorkFastModule;
export { StorkFastProxyModule };
