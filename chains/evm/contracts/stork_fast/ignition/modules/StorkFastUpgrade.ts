// Use this ignition module to upgrade an already-deployed Stork Fast proxy to
// a new implementation. Requires the 'proxyAddress' module parameter, and the
// deploying account (account 0) must be the contract owner.
//
// Run each upgrade with a fresh --deployment-id following the
// chain-<chainId>-<n> convention (e.g. chain-31337-1 for the first upgrade):
// Ignition journals executed futures per deployment, so re-running under a
// previous deployment id would skip the deploy and upgrade steps entirely.

import { buildModule } from "@nomicfoundation/hardhat-ignition/modules";

const StorkFastUpgradeModule = buildModule("StorkFastUpgradeModule", (m) => {
  const proxy = m.contractAt(
    "UpgradeableStorkFast",
    m.getParameter("proxyAddress"),
    { id: "Proxy" }
  );

  const implementation = m.contract("UpgradeableStorkFast", [], {
    id: "Implementation",
  });

  // UUPS pattern: the proxy delegates upgradeToAndCall to the current
  // implementation, which authorizes it via _authorizeUpgrade (onlyOwner).
  // No re-initialization call is needed, so calldata is empty.
  m.call(proxy, "upgradeToAndCall", [implementation, "0x"]);

  return { proxy, implementation };
});

export default StorkFastUpgradeModule;
