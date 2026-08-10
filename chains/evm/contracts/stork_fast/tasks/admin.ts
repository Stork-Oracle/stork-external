import { task } from "hardhat/config";

export const verificationFeeInWei = task(
  "verificationFeeInWei",
  "Get the verification fee in wei"
)
  .addPositionalArgument({
    name: "contractAddress",
    description: "The UpgradeableStorkFast contract address",
  })
  .setAction(() => import("./verificationFeeInWei.js"))
  .build();

export const signerAddresses = task(
  "signerAddresses",
  "Get the list of valid signer addresses"
)
  .addPositionalArgument({
    name: "contractAddress",
    description: "The UpgradeableStorkFast contract address",
  })
  .setAction(() => import("./signerAddresses.js"))
  .build();

export const updateVerificationFeeInWei = task(
  "updateVerificationFeeInWei",
  "Update the verification fee in wei"
)
  .addPositionalArgument({
    name: "contractAddress",
    description: "The UpgradeableStorkFast contract address",
  })
  .addPositionalArgument({
    name: "fee",
    description: "The new fee in wei",
  })
  .setAction(() => import("./updateVerificationFeeInWei.js"))
  .build();

export const version = task("version", "Get the contract version")
  .addPositionalArgument({
    name: "contractAddress",
    description: "The UpgradeableStorkFast contract address",
  })
  .setAction(() => import("./version.js"))
  .build();

export const verifyPayload = task(
  "verifyPayload",
  "Verify and deserialize a signed Stork Fast payload on-chain (sends a transaction paying the verification fee)"
)
  .addPositionalArgument({
    name: "contractAddress",
    description: "The UpgradeableStorkFast contract address",
  })
  .addPositionalArgument({
    name: "payload",
    description:
      "The signed ECDSA payload as a hex string (the 'p' field from the Fast WS/REST API)",
  })
  .setAction(() => import("./verifyPayload.js"))
  .build();

export const addSignerAddress = task(
  "addSignerAddress",
  "Add a valid signer address"
)
  .addPositionalArgument({
    name: "contractAddress",
    description: "The UpgradeableStorkFast contract address",
  })
  .addPositionalArgument({
    name: "address",
    description: "The signer address to add",
  })
  .setAction(() => import("./addSignerAddress.js"))
  .build();

export const removeSignerAddress = task(
  "removeSignerAddress",
  "Remove a valid signer address"
)
  .addPositionalArgument({
    name: "contractAddress",
    description: "The UpgradeableStorkFast contract address",
  })
  .addPositionalArgument({
    name: "address",
    description: "The signer address to remove",
  })
  .setAction(() => import("./removeSignerAddress.js"))
  .build();
