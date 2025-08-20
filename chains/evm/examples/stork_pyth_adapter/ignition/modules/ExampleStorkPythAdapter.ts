import { buildModule } from "@nomicfoundation/hardhat-ignition/modules";

const ExampleStorkPythAdapterModule = buildModule("ExampleStorkPythAdapterModule", (m) => {
  // Parameter for the contract deployment
  const storkContractAddress = m.getParameter("storkContractAddress");
  
  const exampleStorkPythAdapter = m.contract(
    "ExampleStorkPythAdapter",
    [storkContractAddress]
  );

  return { exampleStorkPythAdapter };
});

export default ExampleStorkPythAdapterModule;
