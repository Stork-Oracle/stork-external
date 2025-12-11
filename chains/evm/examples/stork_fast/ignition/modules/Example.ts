import { buildModule } from "@nomicfoundation/hardhat-ignition/modules";

const ExampleModule = buildModule("ExampleModule", (m) => {
  // You'll need to replace this with your actual Stork contract address
  const storkFastContractAddress = m.getParameter("storkFastContractAddress");

  const example = m.contract("Example", [storkFastContractAddress]);

  return { example };
});

export default ExampleModule;
