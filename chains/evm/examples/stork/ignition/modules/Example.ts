import { buildModule } from "@nomicfoundation/hardhat-ignition/modules";

const ExampleModule = buildModule("ExampleModule", (m) => {
  // You'll need to replace this with your actual Stork contract address
  const storkContractAddress = m.getParameter("storkContractAddress");
  
  const example = m.contract("Example", [storkContractAddress]);

  return { example };
});

export default ExampleModule;
