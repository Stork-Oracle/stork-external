import * as hre from "hardhat";


const errorSignatures = [
    "InsufficientFee()",
    "NoFreshUpdate()",
    "NotFound()",
    "StaleValue()",
    "InvalidSignature()"
];

errorSignatures.forEach((sig) => {
    const errorHash = hre.ethers.keccak256(hre.ethers.toUtf8Bytes(sig));
    const selector = errorHash.substring(0, 10);  // Take first 4 bytes (8 hex characters)
    console.log(`${sig}: ${selector}`);
});
