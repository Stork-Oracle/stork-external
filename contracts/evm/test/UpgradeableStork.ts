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

  describe('verifyPublisherSignatureV1', function () {
    it("Should verify the given ETH signature for arguments", async function () {
      const { deployed } = await loadFixture(deployUpgradeableStork);

      const sig = {
        pubKey: "0x0810E094a8b0e750c7ACB66F469AfBBd595FF69b",
        assetPairId: "ETHUSD",
        timestamp: 1680210934,
        quantizedValue: '1000000000000000000',
        r: "0xd80926f0433827d55e17bc77953b44788fb40057c55b2578da4f59361d758555",
        s: "0x69703bad148facb6ba7e5d61676240d6e50162d97e0e7e31d7c7ccd35db6df5f",
        v: 0x1b
      };

      expect(await deployed.verifyPublisherSignatureV1(
        sig.pubKey,
        sig.assetPairId,
        sig.timestamp,
        sig.quantizedValue,
        sig.r,
        sig.s,
        sig.v
      )).to.eq(true);
    });

    it("Should verify the given BTC signature for arguments", async function () {
      const { deployed } = await loadFixture(deployUpgradeableStork);

      const sig = {
        pubKey: "0x16eb47a6bbdf1e1d1e9ac23e6f473f1bcae519c0",
        assetPairId: "BTCUSD",
        timestamp: 1721755261,
        quantizedValue: '66078270000000000000000',
        r: "0xbb1e6f87445556233f98c085e2e25e5938bb0fa4eee42b7df06f50836ae4e42e",
        s: "0x79f16b0f10c35db8a08088c53dcaa40355dca57db840eb850b9328fc064c6002",
        v: 0x1c
      };

      expect(await deployed.verifyPublisherSignatureV1(
        sig.pubKey,
        sig.assetPairId,
        sig.timestamp,
        sig.quantizedValue,
        sig.r,
        sig.s,
        sig.v
      )).to.eq(true);
    });

    it("Should return false if invalid signature", async function () {
      const { deployed } = await loadFixture(deployUpgradeableStork);

      const sig = {
        pubKey: "0x0810E094a8b0e750c7ACB66F469AfBBd595FF69b",
        assetPairId: "ETHUSD",
        timestamp: 1680210934,
        quantizedValue: '1000000000000000000',
        r: "0xd80926f0433827d55e17bc77953b44788fb40057c55b2578da4f59361d758555",
        s: "0x69703bad148facb6ba7e5d61676240d6e50162d97e0e7e31d7c7ccd35db6df5f",
        v: 0x1c // changed from 0x1b
      };

      expect(await deployed.verifyPublisherSignatureV1(
        sig.pubKey,
        sig.assetPairId,
        sig.timestamp,
        sig.quantizedValue,
        sig.r,
        sig.s,
        sig.v
      )).to.eq(false);
    });
  });

  describe('verifyStorkSignatureV1', function () {
    it("Should verify the given signature for arguments", async function () {
      const { deployed } = await loadFixture(deployUpgradeableStork);

      const STORK_PUBKEY = "0xC4A02e7D370402F4afC36032076B05e74FF81786"
      const ID = ethers.keccak256(ethers.toUtf8Bytes("BTCUSD"))
      const TIMESTAMP = "1720722087644999936"
      const VALUE = "60000000000000000000000"
      const MERKLE_ROOT = ethers.encodeBytes32String("example data")
      const VALUE_COMPUTE_ALG_HASH = ethers.encodeBytes32String("example data")
      const R = "0x3e42e45aadf7da98780de810944ac90424493395c90bf0c21ede86b0d3c2cd7b"
      const S = "0x1d853d65ae5be6046dc4199de2a0ee2b7288f51fc4af6946746c425cb8649879"
      const V = "0x1c"

      expect(await deployed.verifyStorkSignatureV1(
        STORK_PUBKEY,
        ID,
        TIMESTAMP,
        VALUE,
        MERKLE_ROOT,
        VALUE_COMPUTE_ALG_HASH,
        R,
        S,
        V
      )).to.eq(true);
    });

    it("Should return false if invalid signature", async function () {
      const { deployed } = await loadFixture(deployUpgradeableStork);

      const STORK_PUBKEY = "0xC4A02e7D370402F4afC36032076B05e74FF81786" 
      const ID = ethers.keccak256(ethers.toUtf8Bytes("BTCUSD"))
      const TIMESTAMP = "1720722087644999936"
      const VALUE = "60000000000000000000000"
      const MERKLE_ROOT = ethers.encodeBytes32String("example data")
      const VALUE_COMPUTE_ALG_HASH = ethers.encodeBytes32String("example data")
      const R = "0x3e42e45aadf7da98780de810944ac90424493395c90bf0c21ede86b0d3c2cd7b"
      const S = "0x1d853d65ae5be6046dc4199de2a0ee2b7288f51fc4af6946746c425cb8649879"
      const V = "0x1b" // changed from 0x1c

      expect(await deployed.verifyStorkSignatureV1(
        STORK_PUBKEY,
        ID,
        TIMESTAMP,
        VALUE,
        MERKLE_ROOT,
        VALUE_COMPUTE_ALG_HASH,
        R,
        S,
        V
      )).to.eq(false);
    });
  });

  describe('verifyMerkleRoot', function () {
    it("Should work with one value", async function () {
      const { deployed } = await loadFixture(deployUpgradeableStork);

      const hashes = [
        "0xa3330b2fc8da019adc16cfe62cfba5a2494e1e0d82a3410ca8395774565c8f61",
      ];

      expect(await deployed.verifyMerkleRoot(hashes, "0xa3330b2fc8da019adc16cfe62cfba5a2494e1e0d82a3410ca8395774565c8f61")).to.eq(true);
    });

    it("Should work with two values", async function () {
      const { deployed } = await loadFixture(deployUpgradeableStork);

      const hashes = [
        "0xa3330b2fc8da019adc16cfe62cfba5a2494e1e0d82a3410ca8395774565c8f61",
        "0x1fd10d34dcb1ef1ea4965bb60c494bf0782a6f0a7ad3354d1f1f3dc61fb65b21",
      ];

      expect(await deployed.verifyMerkleRoot(hashes, "0xa856b2ea34e72a9024cb0e70df0ff6e642f73c8db6ee506b4770796977e9497f")).to.eq(true);
    });

    it("Should work with two values reversed", async function () {
      const { deployed } = await loadFixture(deployUpgradeableStork);

      const hashes = [
        "0x1fd10d34dcb1ef1ea4965bb60c494bf0782a6f0a7ad3354d1f1f3dc61fb65b21",
        "0xa3330b2fc8da019adc16cfe62cfba5a2494e1e0d82a3410ca8395774565c8f61",
      ];

      expect(await deployed.verifyMerkleRoot(hashes, "0x6c9a6f99ea9fa21d9029392c5fc9674e918f3e20dcf29c5fb04a43560e620d5b")).to.eq(true);
    });

    it("Should compute the merkle root for even number of values", async function () {
      const { deployed } = await loadFixture(deployUpgradeableStork);

      const hashes = [
        "0xca978112ca1bbdcafac231b39a23dc4da786eff8147c4e72b9807785afee48bb",
        "0x3e23e8160039594a33894f6564e1b1348bbd7a0088d42c4acb73eeaed59c009d",
        "0x2e7d2c03a9507ae265ecf5b5356885a53393a2029d241394997265a1a25aefc6",
        "0x18ac3e7343f016890c510e93f935261169d9e3f565436429830faf0934f4f8e4",
      ];

      expect(await deployed.verifyMerkleRoot(hashes, "0x50607deffb9baf552082f3460d5c8f8d521ca672862844064fc7396c8bf2715d")).to.eq(true);
    });

    it("Should compute the merkle root for odd number of values", async function () {
      const { deployed } = await loadFixture(deployUpgradeableStork);

      const hashes = [
        "0xca978112ca1bbdcafac231b39a23dc4da786eff8147c4e72b9807785afee48bb",
        "0x3e23e8160039594a33894f6564e1b1348bbd7a0088d42c4acb73eeaed59c009d",
        "0x2e7d2c03a9507ae265ecf5b5356885a53393a2029d241394997265a1a25aefc6",
      ];

      expect(await deployed.verifyMerkleRoot(hashes, "0x7af913bc71f11ec77c2fee5c62f94368d78bedb9a312b04f793dcb59cd14ec89")).to.eq(true);
    });
  });

  describe("verifyPublisherSignaturesV1", function () {
    it("Should verify the given signatures for arguments", async function () {
      const { deployed } = await loadFixture(deployUpgradeableStork);

      expect(await deployed.verifyPublisherSignaturesV1(
        [
          {
            pubKey: "0x0810E094a8b0e750c7ACB66F469AfBBd595FF69b",
            assetPairId: "ETHUSD",
            timestamp: 1680210934,
            quantizedValue: '1000000000000000000',
            r: "0xd80926f0433827d55e17bc77953b44788fb40057c55b2578da4f59361d758555",
            s: "0x69703bad148facb6ba7e5d61676240d6e50162d97e0e7e31d7c7ccd35db6df5f",
            v: 0x1b
          },
          {
            pubKey: "0x16eb47a6bbdf1e1d1e9ac23e6f473f1bcae519c0",
            assetPairId: "BTCUSD",
            timestamp: 1721755261,
            quantizedValue: '66078270000000000000000',
            r: "0xbb1e6f87445556233f98c085e2e25e5938bb0fa4eee42b7df06f50836ae4e42e",
            s: "0x79f16b0f10c35db8a08088c53dcaa40355dca57db840eb850b9328fc064c6002",
            v: 0x1c
          }
        ],
        "0xa856b2ea34e72a9024cb0e70df0ff6e642f73c8db6ee506b4770796977e9497f", // taken from above tests
      )).to.eq(true);
    })
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
