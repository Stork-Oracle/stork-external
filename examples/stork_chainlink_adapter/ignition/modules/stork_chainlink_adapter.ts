import { buildModule } from "@nomicfoundation/hardhat-ignition/modules";

export default buildModule("ExampleStorkChainlinkAdapter", (m) => {
    const exampleStorkChainlinkAdapter = m.contract(
        "ExampleStorkChainlinkAdapter",
        ["0x2279B7A0a67DB372996a5FaB50D91eAA73d2eBe6", "0x4254435553440000000000000000000000000000000000000000000000000000"]
    );

    return { exampleStorkChainlinkAdapter: exampleStorkChainlinkAdapter };
});