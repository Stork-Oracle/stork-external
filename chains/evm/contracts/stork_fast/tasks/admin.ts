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

export const signerAddress = task(
  "signerAddress",
  "Get the signer address"
)
  .addPositionalArgument({
    name: "contractAddress",
    description: "The UpgradeableStorkFast contract address",
  })
  .setAction(() => import("./signerAddress.js"))
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

export const updateSignerAddress = task(
  "updateSignerAddress",
  "Update the signer address"
)
  .addPositionalArgument({
    name: "contractAddress",
    description: "The UpgradeableStorkFast contract address",
  })
  .addPositionalArgument({
    name: "address",
    description: "The new signer address",
  })
  .setAction(() => import("./updateSignerAddress.js"))
  .build();
