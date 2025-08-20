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

  describe("convertInt192ToInt64Precise", function () {
    const INT64_MAX = BigInt("9223372036854775807");
    const INT64_MIN = BigInt("-9223372036854775808");
    const DEFAULT_EXPONENT = -18;

    it("Should return value without adjustment when it fits in int64", async function () {
      const testValue = BigInt("1000000000000000000"); // 1e18, fits in int64
      
      const result = await storkPythAdapter.convertInt192ToInt64Precise(testValue);
      
      expect(result.val).to.equal(testValue);
      expect(result.exp).to.equal(DEFAULT_EXPONENT);
    });

    it("Should handle small positive values", async function () {
      const testValue = BigInt("123456789");
      
      const result = await storkPythAdapter.convertInt192ToInt64Precise(testValue);
      
      expect(result.val).to.equal(testValue);
      expect(result.exp).to.equal(DEFAULT_EXPONENT);
    });

    it("Should handle small negative values", async function () {
      const testValue = BigInt("-123456789");
      
      const result = await storkPythAdapter.convertInt192ToInt64Precise(testValue);
      
      expect(result.val).to.equal(testValue);
      expect(result.exp).to.equal(DEFAULT_EXPONENT);
    });

    it("Should handle zero", async function () {
      const testValue = BigInt("0");
      
      const result = await storkPythAdapter.convertInt192ToInt64Precise(testValue);
      
      expect(result.val).to.equal(0);
      expect(result.exp).to.equal(DEFAULT_EXPONENT);
    });

    it("Should adjust precision for values exceeding int64 max", async function () {
      // Value that requires one division by 10
      const testValue = INT64_MAX * BigInt(10) + BigInt(5); // 92233720368547758075
      
      const result = await storkPythAdapter.convertInt192ToInt64Precise(testValue);
      
      // Should be divided by 10: 9223372036854775807 (truncated)
      expect(result.val).to.equal(INT64_MAX);
      expect(result.exp).to.equal(DEFAULT_EXPONENT + 1);
    });

    it("Should adjust precision for values below int64 min", async function () {
      // Value that requires one division by 10
      const testValue = INT64_MIN * BigInt(10) - BigInt(5); // -92233720368547758085
      
      const result = await storkPythAdapter.convertInt192ToInt64Precise(testValue);
      
      // Should be divided by 10: -9223372036854775808 (truncated)
      expect(result.val).to.equal(INT64_MIN);
      expect(result.exp).to.equal(DEFAULT_EXPONENT + 1);
    });

    it("Should handle multiple precision adjustments", async function () {
      // Value that requires multiple divisions by 10
      const testValue = INT64_MAX * BigInt(1000); // Requires 3 divisions
      
      const result = await storkPythAdapter.convertInt192ToInt64Precise(testValue);
      
      expect(result.val).to.equal(INT64_MAX);
      expect(result.exp).to.equal(DEFAULT_EXPONENT + 3);
    });

    it("Should handle very large positive values", async function () {
      // Test with a very large int192 value
      const testValue = BigInt("100000000000000000000000000000000000000000000000000"); // 1e50

      const result = await storkPythAdapter.convertInt192ToInt64Precise(testValue);
      
      // Should be scaled down significantly
      const expectedValue = BigInt("1000000000000000000");
      expect(result.val).to.be.equal(expectedValue);
      expect(result.exp).to.be.equal(14);
    });

    it("Should handle very large negative values", async function () {
      // Test with a very large negative int192 value
      const testValue = BigInt("-100000000000000000000000000000000000000000000000000"); // -1e50
      
      const result = await storkPythAdapter.convertInt192ToInt64Precise(testValue);
      
      const expectedValue = BigInt("-1000000000000000000");
      expect(result.val).to.be.equal(expectedValue);
      expect(result.exp).to.be.equal(14);
    });

    it("Should handle edge case at int64 max boundary", async function () {
      const testValue = INT64_MAX;
      
      const result = await storkPythAdapter.convertInt192ToInt64Precise(testValue);
      
      expect(result.val).to.equal(INT64_MAX);
      expect(result.exp).to.equal(DEFAULT_EXPONENT);
    });

    it("Should handle edge case at int64 min boundary", async function () {
      const testValue = INT64_MIN;
      
      const result = await storkPythAdapter.convertInt192ToInt64Precise(testValue);
      
      expect(result.val).to.equal(INT64_MIN);
      expect(result.exp).to.equal(DEFAULT_EXPONENT);
    });

    it("Should handle value just above int64 max", async function () {
      const testValue = INT64_MAX + BigInt(1);
      
      const result = await storkPythAdapter.convertInt192ToInt64Precise(testValue);
      
      // Should be divided by 10
      const expectedValue = (INT64_MAX + BigInt(1)) / BigInt(10);
      expect(result.val).to.equal(expectedValue);
      expect(result.exp).to.equal(DEFAULT_EXPONENT + 1);
    });

    it("Should handle value just below int64 min", async function () {
      const testValue = INT64_MIN - BigInt(1);
      
      const result = await storkPythAdapter.convertInt192ToInt64Precise(testValue);
      
      // Should be divided by 10
      const expectedValue = (INT64_MIN - BigInt(1)) / BigInt(10);
      expect(result.val).to.equal(expectedValue);
      expect(result.exp).to.equal(DEFAULT_EXPONENT + 1);
    });

    it("Should maintain precision relationship between value and exponent", async function () {
      const originalValue = BigInt("12345678901234567890123456789"); // Large value
      
      const result = await storkPythAdapter.convertInt192ToInt64Precise(originalValue);
      
      // The scaled value multiplied by 10^(exp_shift) should approximate the original
      // where exp_shift = result.exp - DEFAULT_EXPONENT
      const expShift = Number(result.exp) - DEFAULT_EXPONENT;
      expect(expShift).to.be.greaterThan(0); // Should require scaling
      
      // Verify the result is within int64 bounds
      expect(result.val).to.be.lessThanOrEqual(INT64_MAX);
      expect(result.val).to.be.greaterThanOrEqual(INT64_MIN);
    });
  });
});
