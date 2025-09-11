import { buildModule } from "@nomicfoundation/hardhat-ignition/modules";

const ExampleStorkChainlinkAdapterModule = buildModule("ExampleStorkChainlinkAdapterModule", (m) => {
  // Parameters for the contract deployment
  const storkContractAddress = m.getParameter("storkContractAddress");
  const priceId = m.getParameter("priceId");
  
  const exampleStorkChainlinkAdapter = m.contract(
    "ExampleStorkChainlinkAdapter",
    [storkContractAddress, priceId]
  );

  return { exampleStorkChainlinkAdapter };
});

export default ExampleStorkChainlinkAdapterModule;
