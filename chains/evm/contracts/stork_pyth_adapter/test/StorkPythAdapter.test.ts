import { expect } from "chai";
import { ethers } from "hardhat";
import { StorkPythAdapter } from "../typechain-types";

describe("StorkPythAdapter", function () {
  let storkPythAdapter: StorkPythAdapter;
  let mockStorkAddress: string;

  beforeEach(async function () {
    // Deploy a mock Stork contract or use a known address
    mockStorkAddress = "0x1234567890123456789012345678901234567890";
    
    const StorkPythAdapterFactory = await ethers.getContractFactory("StorkPythAdapter");
    storkPythAdapter = await StorkPythAdapterFactory.deploy(mockStorkAddress);
  });

  it("Should deploy successfully", async function () {
    expect(await storkPythAdapter.getAddress()).to.be.properAddress;
  });

  it("Should return the correct stork address", async function () {
    expect(await storkPythAdapter.stork()).to.equal(mockStorkAddress);
  });

  it("Should revert on unsupported EMA methods", async function () {
    const dummyId = "0x1234567890123456789012345678901234567890123456789012345678901234";
    
    await expect(storkPythAdapter.getEmaPrice(dummyId))
      .to.be.revertedWith("Not supported");
    
    await expect(storkPythAdapter.getEmaPriceUnsafe(dummyId))
      .to.be.revertedWith("Not supported");
    
    await expect(storkPythAdapter.getEmaPriceNoOlderThan(dummyId, 100))
      .to.be.revertedWith("Not supported");
  });

  it("Should revert on unsupported update methods", async function () {
    await expect(storkPythAdapter.updatePriceFeeds([]))
      .to.be.revertedWith("Not supported");
    
    await expect(storkPythAdapter.getUpdateFee([]))
      .to.be.revertedWith("Not supported");
  });
});
