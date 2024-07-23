import { loadFixture } from "@nomicfoundation/hardhat-toolbox/network-helpers";

// @ts-expect-error upgrades is loaded in hardhat/config
import { ethers, upgrades } from "hardhat";

import { expect } from "chai";

describe("UpgradeableStork", function() {
  async function deployUpgradeableStork() {
    // Contracts are deployed using the first signer/account by default
    const [owner, otherAccount] = await ethers.getSigners();
    const STORK_PUBLIC_KEY = "0xC4A02e7D370402F4afC36032076B05e74FF81786";
    const VALID_TIMEOUT_SECONDS = 60;
    const UPDATE_FEE_IN_WEI = 1;

    const UpgradeableStork = await ethers.getContractFactory("UpgradeableStork");

    const deployed = await upgrades.deployProxy(UpgradeableStork, [owner.address, STORK_PUBLIC_KEY, VALID_TIMEOUT_SECONDS, UPDATE_FEE_IN_WEI]);

    return { deployed, owner, otherAccount };
  }

  describe("Deploy", function () {
    it("Should return expected version", async function () {
      const { deployed } = await loadFixture(deployUpgradeableStork);

      expect(await deployed.version()).to.equal("1.0.0");
    });

    it("Should return owner", async function () {
      const { deployed, owner } = await loadFixture(deployUpgradeableStork);

      expect(await deployed.owner()).to.equal(owner.address);
    });

    it("Should return stork public key", async function () {
      const { deployed } = await loadFixture(deployUpgradeableStork);

      expect(await deployed.storkPublicKey()).to.equal("0xC4A02e7D370402F4afC36032076B05e74FF81786");
    });

    it("Should return update fee", async function () {
      const { deployed } = await loadFixture(
        deployUpgradeableStork
      );

      expect(await deployed.singleUpdateFeeInWei()).to.equal(1);
    });

    it("Should return valid timeout", async function () {
      const { deployed } = await loadFixture(
        deployUpgradeableStork
      );

      expect(await deployed.validTimePeriodSeconds()).to.equal(60);
    });
  });

  describe("Upgrade", function () {
    it("Should upgrade successfully", async function () {
      const { deployed } = await loadFixture(deployUpgradeableStork);

      const UpgradeableStorkV2 = await ethers.getContractFactory("UpgradeableStork");

      const upgraded = await upgrades.upgradeProxy(deployed, UpgradeableStorkV2);

      expect(await upgraded.version()).to.equal("1.0.0");
    });

    it("Should revert if not owner", async function () {
      const { deployed, otherAccount } = await loadFixture(deployUpgradeableStork);

      const UpgradeableStorkV2 = await ethers.getContractFactory("UpgradeableStork");

      await expect(upgrades.upgradeProxy(deployed, UpgradeableStorkV2.connect(otherAccount))).to.be.revertedWithCustomError(deployed, "OwnableUnauthorizedAccount");
    });
  });

  describe("updateStorkPublicKey", function () {
    it("Should update successfully", async function () {
      const { deployed } = await loadFixture(deployUpgradeableStork);

      const newPublicKey = "0x1234567890123456789012345678901234567890";

      await deployed.updateStorkPublicKey(newPublicKey);

      expect(await deployed.storkPublicKey()).to.equal(newPublicKey);
    });

    it("Should revert if not owner", async function () {
      const { deployed, otherAccount } = await loadFixture(deployUpgradeableStork);

      const newPublicKey = "0x1234567890123456789012345678901234567890";

      await expect(deployed.connect(otherAccount).updateStorkPublicKey(newPublicKey)).to.be.revertedWithCustomError(deployed, "OwnableUnauthorizedAccount");

      expect(await deployed.storkPublicKey()).to.equal("0xC4A02e7D370402F4afC36032076B05e74FF81786");
    });
  });

  describe("updateSingleUpdateFeeInWei", function () {
    it("Should update successfully", async function () {
      const { deployed } = await loadFixture(deployUpgradeableStork);

      await deployed.updateSingleUpdateFeeInWei(2);

      expect(await deployed.singleUpdateFeeInWei()).to.equal(2);
    });

    it("Should revert if not owner", async function () {
      const { deployed, otherAccount } = await loadFixture(deployUpgradeableStork);

      await expect(deployed.connect(otherAccount).updateSingleUpdateFeeInWei(2)).to.be.revertedWithCustomError(deployed, "OwnableUnauthorizedAccount");

      expect(await deployed.singleUpdateFeeInWei()).to.equal(1);
    });
  });

  describe("updateValidTimePeriodSeconds", function () {
    it("Should update successfully", async function () {
      const { deployed } = await loadFixture(deployUpgradeableStork);

      await deployed.updateValidTimePeriodSeconds(120);

      expect(await deployed.validTimePeriodSeconds()).to.equal(120);
    });

    it("Should revert if not owner", async function () {
      const { deployed, otherAccount } = await loadFixture(deployUpgradeableStork);

      await expect(deployed.connect(otherAccount).updateValidTimePeriodSeconds(120)).to.be.revertedWithCustomError(deployed, "OwnableUnauthorizedAccount");

      expect(await deployed.validTimePeriodSeconds()).to.equal(60);
    });
  });

  describe("updateTemporalNumericValuesV1", function () {
    it("Should update successfully", async function () {
      const { deployed } = await loadFixture(deployUpgradeableStork);

      // values pulled from sample publisher call
      await deployed.updateTemporalNumericValuesV1([
        {
          temporalNumericValue: {
            timestampNs: "1720722087644999936",
            quantizedValue: "60000000000000000000000",
          },
          id: ethers.keccak256(ethers.toUtf8Bytes("BTCUSD")),
          publisherMerkleRoot: ethers.encodeBytes32String("example data"),
          valueComputeAlgHash: ethers.encodeBytes32String("example data"),
          r: "0x3e42e45aadf7da98780de810944ac90424493395c90bf0c21ede86b0d3c2cd7b",
          s: "0x1d853d65ae5be6046dc4199de2a0ee2b7288f51fc4af6946746c425cb8649879",
          v: "0x1c"
        }
      ], { value: 1 });
    });

    it("Should update multiple successfully", async function () {
      const { deployed } = await loadFixture(deployUpgradeableStork);

      // values pulled from sample publisher call
      await deployed.updateTemporalNumericValuesV1([
        {
          temporalNumericValue: {
            timestampNs: "1720722087644999936",
            quantizedValue: "60000000000000000000000",
          },
          id: ethers.keccak256(ethers.toUtf8Bytes("BTCUSD")),
          publisherMerkleRoot: ethers.encodeBytes32String("example data"),
          valueComputeAlgHash: ethers.encodeBytes32String("example data"),
          r: "0x3e42e45aadf7da98780de810944ac90424493395c90bf0c21ede86b0d3c2cd7b",
          s: "0x1d853d65ae5be6046dc4199de2a0ee2b7288f51fc4af6946746c425cb8649879",
          v: "0x1c"
        },
        {
          temporalNumericValue: {
            timestampNs: "1720722554872999936",
            quantizedValue: "3000000000000000000000",
          },
          id: ethers.keccak256(ethers.toUtf8Bytes("ETHUSD")),
          publisherMerkleRoot: ethers.encodeBytes32String("example data"),
          valueComputeAlgHash: ethers.encodeBytes32String("example data"),
          r: "0x67018d101bb11542b3b43048a4d171122e7eb25b8cebd2fe6cb7412cf3438620",
          s: "0x788ca33b146165b588d5e704faf9e4a9fd036d7ae2d88b48e75ee71628ddd657",
          v: "0x1b"
        },
      ], { value: 2 });
    });

    it("Should revert if insufficient fee", async function () {
      const { deployed } = await loadFixture(deployUpgradeableStork);

      await expect(deployed.updateTemporalNumericValuesV1([
        {
          temporalNumericValue: {
            timestampNs: "1720722087644999936",
            quantizedValue: "60000000000000000000000",
          },
          id: ethers.keccak256(ethers.toUtf8Bytes("BTCUSD")),
          publisherMerkleRoot: ethers.encodeBytes32String("example data"),
          valueComputeAlgHash: ethers.encodeBytes32String("example data"),
          r: "0x3e42e45aadf7da98780de810944ac90424493395c90bf0c21ede86b0d3c2cd7b",
          s: "0x1d853d65ae5be6046dc4199de2a0ee2b7288f51fc4af6946746c425cb8649879",
          v: "0x1c"
        }
      ], { value: 0 })).to.be.revertedWithCustomError(deployed, "InsufficientFee");
    });

    it("Should revert if invalid signature", async function () {
      const { deployed } = await loadFixture(deployUpgradeableStork);

      await expect(deployed.updateTemporalNumericValuesV1([
        {
          temporalNumericValue: {
            timestampNs: "1720722087644999936",
            quantizedValue: "70000000000000000000000", // changed value
          },
          id: ethers.keccak256(ethers.toUtf8Bytes("BTCUSD")),
          publisherMerkleRoot: ethers.encodeBytes32String("example data"),
          valueComputeAlgHash: ethers.encodeBytes32String("example data"),
          r: "0x3e42e45aadf7da98780de810944ac90424493395c90bf0c21ede86b0d3c2cd7b",
          s: "0x1d853d65ae5be6046dc4199de2a0ee2b7288f51fc4af6946746c425cb8649879",
          v: "0x1c"
        }
      ], { value: 1 })).to.be.revertedWithCustomError(deployed, "InvalidSignature");
    });
  });

  describe("getUpdateFeeV1", function () {
    it("Should return expected fee", async function () {
      const { deployed } = await loadFixture(deployUpgradeableStork);

      expect(await deployed.getUpdateFeeV1([{
        temporalNumericValue: {
          timestampNs: "1720722087644999936",
          quantizedValue: "60000000000000000000000",
        },
        id: ethers.keccak256(ethers.toUtf8Bytes("BTCUSD")),
        publisherMerkleRoot: ethers.encodeBytes32String("example data"),
        valueComputeAlgHash: ethers.encodeBytes32String("example data"),
        r: "0x3e42e45aadf7da98780de810944ac90424493395c90bf0c21ede86b0d3c2cd7b",
        s: "0x1d853d65ae5be6046dc4199de2a0ee2b7288f51fc4af6946746c425cb8649879",
        v: "0x1c"
      }])).to.equal(1);
    });
  });

  describe("getTemporalNumericValueV1", function () {
    it("Should revert if never updated value", async function () {
      const { deployed } = await loadFixture(deployUpgradeableStork);

      await expect(deployed.getTemporalNumericValueV1(ethers.encodeBytes32String("BTCUSD"))).to.be.revertedWithCustomError(deployed, "NotFound");
    });

    it("Should revert if value is stale", async function () {
      const { deployed } = await loadFixture(deployUpgradeableStork);

      // values pulled from sample publisher call
      await deployed.updateTemporalNumericValuesV1([
        {
          temporalNumericValue: {
            timestampNs: "1720722087644999936",
            quantizedValue: "60000000000000000000000",
          },
          id: ethers.keccak256(ethers.toUtf8Bytes("BTCUSD")),
          publisherMerkleRoot: ethers.encodeBytes32String("example data"),
          valueComputeAlgHash: ethers.encodeBytes32String("example data"),
          r: "0x3e42e45aadf7da98780de810944ac90424493395c90bf0c21ede86b0d3c2cd7b",
          s: "0x1d853d65ae5be6046dc4199de2a0ee2b7288f51fc4af6946746c425cb8649879",
          v: "0x1c"
        }
      ], { value: 1 });

      await expect(deployed.getTemporalNumericValueV1(ethers.keccak256(ethers.toUtf8Bytes("BTCUSD")))).to.be.revertedWithCustomError(deployed, "StaleValue");
    });

    it("Should return expected value", async function () {
      const { deployed } = await loadFixture(deployUpgradeableStork);
      
      // to avoid time period check
      await deployed.updateValidTimePeriodSeconds(100000000000);

      // values pulled from sample publisher call
      await deployed.updateTemporalNumericValuesV1([
        {
          temporalNumericValue: {
            timestampNs: "1720722087644999936",
            quantizedValue: "60000000000000000000000",
          },
          id: ethers.keccak256(ethers.toUtf8Bytes("BTCUSD")),
          publisherMerkleRoot: ethers.encodeBytes32String("example data"),
          valueComputeAlgHash: ethers.encodeBytes32String("example data"),
          r: "0x3e42e45aadf7da98780de810944ac90424493395c90bf0c21ede86b0d3c2cd7b",
          s: "0x1d853d65ae5be6046dc4199de2a0ee2b7288f51fc4af6946746c425cb8649879",
          v: "0x1c"
        }
      ], { value: 1 });

      expect(await deployed.getTemporalNumericValueV1(ethers.keccak256(ethers.toUtf8Bytes("BTCUSD")))).to.deep.equal([
        1720722087644999936n,
        60000000000000000000000n
      ]);
    });

    it("Should return expected value if second value reverts", async function () {
      const { deployed } = await loadFixture(deployUpgradeableStork);
      
      // to avoid time period check
      await deployed.updateValidTimePeriodSeconds(100000000000);

      // values pulled from sample publisher call
      await deployed.updateTemporalNumericValuesV1([
        {
          temporalNumericValue: {
            timestampNs: "1720722087644999936",
            quantizedValue: "60000000000000000000000",
          },
          id: ethers.keccak256(ethers.toUtf8Bytes("BTCUSD")),
          publisherMerkleRoot: ethers.encodeBytes32String("example data"),
          valueComputeAlgHash: ethers.encodeBytes32String("example data"),
          r: "0x3e42e45aadf7da98780de810944ac90424493395c90bf0c21ede86b0d3c2cd7b",
          s: "0x1d853d65ae5be6046dc4199de2a0ee2b7288f51fc4af6946746c425cb8649879",
          v: "0x1c"
        }
      ], { value: 1 });

      await deployed.updateTemporalNumericValuesV1([
        {
          temporalNumericValue: {
            timestampNs: "1720722087644999970", // changed value
            quantizedValue: "60000000000000000000000",
          },
          id: ethers.keccak256(ethers.toUtf8Bytes("BTCUSD")),
          publisherMerkleRoot: ethers.encodeBytes32String("example data"),
          valueComputeAlgHash: ethers.encodeBytes32String("example data"),
          r: "0x3e42e45aadf7da98780de810944ac90424493395c90bf0c21ede86b0d3c2cd7b",
          s: "0x1d853d65ae5be6046dc4199de2a0ee2b7288f51fc4af6946746c425cb8649879",
          v: "0x1c"
        }
      ], { value: 1 }).catch(() => {
        // ignore revert
      });

      expect(await deployed.getTemporalNumericValueV1(ethers.keccak256(ethers.toUtf8Bytes("BTCUSD")))).to.deep.equal([
        1720722087644999936n,
        60000000000000000000000n
      ]);
    });
  });
});
