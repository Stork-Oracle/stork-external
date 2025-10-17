import { buildModule } from "@nomicfoundation/hardhat-ignition/modules";

const ProxyModule = buildModule("ProxyModule", (m) => {
  const UpgradeableFirstPartyStork = m.contract("UpgradeableFirstPartyStork");

  const proxyAdminOwner = m.getAccount(0);
  
  const initializeCalldata = m.encodeFunctionCall(UpgradeableFirstPartyStork, "initialize", [proxyAdminOwner]);

  const proxy = m.contract("TransparentUpgradeableProxy", [
    UpgradeableFirstPartyStork,
    proxyAdminOwner,
    initializeCalldata,
  ]);

  return { proxy };
});
  
const FirstPartyStorkModule = buildModule("FirstPartyStorkModule", (m) => {
  const { proxy } = m.useModule(ProxyModule);

  const FirstPartyStork = m.contractAt("UpgradeableFirstPartyStork", proxy);

  return { FirstPartyStork, proxy };
});

export default FirstPartyStorkModule;
