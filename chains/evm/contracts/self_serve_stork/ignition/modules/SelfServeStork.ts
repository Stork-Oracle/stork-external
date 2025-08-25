import { buildModule } from "@nomicfoundation/hardhat-ignition/modules";

export default buildModule("SelfServeStorkModule", (m) => {
  const selfServeStork = m.contract("SelfServeStork", [m.getAccount(0)]);

  return { selfServeStork };
});
