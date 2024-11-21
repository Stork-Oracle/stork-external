import { buildModule } from "@nomicfoundation/hardhat-ignition/modules";

export default buildModule("ExampleStorkPythAdapter", (m) => {
    const exampleStorkPythAdapter = m.contract(
        "ExampleStorkPythAdapter",
        ["0x2279B7A0a67DB372996a5FaB50D91eAA73d2eBe6"]
    );

    return { exampleStorkPythAdapter: exampleStorkPythAdapter };
});