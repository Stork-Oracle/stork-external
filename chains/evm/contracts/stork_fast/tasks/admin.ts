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

export const storkFastAddress = task(
  "storkFastAddress",
  "Get the Stork Fast address"
)
  .addPositionalArgument({
    name: "contractAddress",
    description: "The UpgradeableStorkFast contract address",
  })
  .setAction(() => import("./storkFastAddress.js"))
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

export const updateStorkFastAddress = task(
  "updateStorkFastAddress",
  "Update the Stork Fast address"
)
  .addPositionalArgument({
    name: "contractAddress",
    description: "The UpgradeableStorkFast contract address",
  })
  .addPositionalArgument({
    name: "address",
    description: "The new Stork Fast address",
  })
  .setAction(() => import("./updateStorkFastAddress.js"))
  .build();
