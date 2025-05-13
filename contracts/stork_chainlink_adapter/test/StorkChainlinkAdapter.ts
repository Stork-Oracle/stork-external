import { loadFixture } from "@nomicfoundation/hardhat-toolbox/network-helpers";
import { expect } from "chai";
import hre from "hardhat";

describe("StorkChainlinkAdapter", function () {
  // We define a fixture to reuse the same setup in every test.
  // We use loadFixture to run this setup once, snapshot that state,
  // and reset Hardhat Network to that snapshot in every test.
  async function deployStorkChainlinkAdapterFixture() {
    const storkContract = "0x2279B7A0a67DB372996a5FaB50D91eAA73d2eBe6";
    const encodedAssetId =
      "0x4254435553440000000000000000000000000000000000000000000000000000";

    // Contracts are deployed using the first signer/account by default
    const [owner, otherAccount] = await hre.ethers.getSigners();

    const StorkChainlinkAdapter = await hre.ethers.getContractFactory(
      "StorkChainlinkAdapter"
    );
    const storkChainlinkAdapter = await StorkChainlinkAdapter.deploy(
      storkContract,
      encodedAssetId
    );

    return { storkChainlinkAdapter, owner, otherAccount };
  }

  describe("Deployment", function () {
    it("Should set the right unlockTime", async function () {
      const { storkChainlinkAdapter } = await loadFixture(
        deployStorkChainlinkAdapterFixture
      );

      expect(await storkChainlinkAdapter.decimals()).to.equal(18);
      expect(await storkChainlinkAdapter.description()).to.equal(
        "A port of a chainlink aggregator powered by Stork"
      );
      expect(await storkChainlinkAdapter.version()).to.equal(1);
    });
  });
});
