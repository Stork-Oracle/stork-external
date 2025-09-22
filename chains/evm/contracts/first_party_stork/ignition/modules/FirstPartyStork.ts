import { buildModule } from "@nomicfoundation/hardhat-ignition/modules";

const ProxyModule = buildModule("ProxyModule", (m) => {
  const UpgradeableSelfServeStork = m.contract("UpgradeableSelfServeStork");

  const proxy = m.contract("ERC1967Proxy", [
    UpgradeableSelfServeStork,
    "0x",
  ]);

  return { proxy };
});
  
const SelfServeStorkModule = buildModule("SelfServeStorkModule", (m) => {
  const { proxy } = m.useModule(ProxyModule);

  const SelfServeStork = m.contractAt("SelfServeStork", proxy);

  return { SelfServeStork, proxy };
});

export default SelfServeStorkModule;
