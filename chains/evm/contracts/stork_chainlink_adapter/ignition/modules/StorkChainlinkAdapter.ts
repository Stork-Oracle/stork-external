import { buildModule } from "@nomicfoundation/hardhat-ignition/modules";

export default buildModule("StorkChainlinkAdapter", (m) => {
    const storkChainlinkAdapter = m.contract(
        "StorkChainlinkAdapter",
        [m.getParameter("storkContract"), m.getParameter("encodedAssetId")]
    );

    return { storkChainlinkAdapter };
});
