import { buildModule } from "@nomicfoundation/hardhat-ignition/modules";

export default buildModule("StorkPythAdapter", (m) => {
    const storkPythAdapter = m.contract(
        "StorkPythAdapter",
        [m.getParameter("storkContract")]
    );

    return { storkPythAdapter };
});
