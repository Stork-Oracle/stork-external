import { buildModule } from "@nomicfoundation/hardhat-ignition/modules";

const StorkFastProxyModule = buildModule("StorkFastProxyModule", (m) => {
  const UpgradeableStorkFast = m.contract("UpgradeableStorkFast");

  const proxyAdminOwner = m.getAccount(0);

  const initializeCalldata = m.encodeFunctionCall(UpgradeableStorkFast, "initialize", [proxyAdminOwner, m.getParameter("storkFastAddress"), m.getParameter("verificationFeeInWei")]);

  const proxy = m.contract("TransparentUpgradeableProxy", [
    UpgradeableStorkFast,
    proxyAdminOwner,
    initializeCalldata,
  ]);

  return { proxy };
});

const StorkFastModule = buildModule("StorkFastModule", (m) => {
  const { proxy } = m.useModule(StorkFastProxyModule);

  const StorkFast = m.contractAt("UpgradeableStorkFast", proxy);

  return { StorkFast, proxy };
});
