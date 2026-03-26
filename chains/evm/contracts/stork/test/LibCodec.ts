import { ethers, upgrades } from "hardhat";
import { expect } from "chai";
import { loadFixture } from "@nomicfoundation/hardhat-toolbox/network-helpers";

describe("LibCodec", function () {
  const WORDS_PER_ENTRY = 6;

  // Encode a TemporalNumericValueInput into 6 packed uint256 words (off-chain mirror of LibCodec.encode)
  function packEntry(entry: {
    timestampNs: bigint;
    quantizedValue: bigint;
    id: string;
    publisherMerkleRoot: string;
    valueComputeAlgHash: string;
    r: string;
    s: string;
    v: number;
  }): bigint[] {
    const vFlag = BigInt(entry.v - 27) << 255n;
    const ts = BigInt(entry.timestampNs) << 192n;
    const qv = entry.quantizedValue < 0n
      ? (1n << 192n) + entry.quantizedValue  // two's complement for int192
      : entry.quantizedValue;
    const word0 = vFlag | ts | (qv & ((1n << 192n) - 1n));

    return [
      word0,
      BigInt(entry.id),
      BigInt(entry.publisherMerkleRoot),
      BigInt(entry.valueComputeAlgHash),
      BigInt(entry.r),
      BigInt(entry.s),
    ];
  }

  // Decode word0 back to timestampNs, quantizedValue, v
  function unpackWord0(w0: bigint): { timestampNs: bigint; quantizedValue: bigint; v: number } {
    const vFlag = Number(w0 >> 255n);
    const timestampNs = (w0 >> 192n) & 0x7FFFFFFFFFFFFFFFn;
    let quantizedValue = w0 & ((1n << 192n) - 1n);
    // Sign-extend: if bit 191 is set, it's negative
    if (quantizedValue >= (1n << 191n)) {
      quantizedValue -= (1n << 192n);
    }
    return { timestampNs, quantizedValue, v: vFlag + 27 };
  }

  function makeEntry(i: number) {
    return {
      timestampNs: BigInt("1700000000000000000") + BigInt(i),
      quantizedValue: BigInt("1000000000000000000") * BigInt(i + 1),
      id: ethers.keccak256(ethers.AbiCoder.defaultAbiCoder().encode(["string", "uint256"], ["id", i])),
      publisherMerkleRoot: ethers.keccak256(ethers.AbiCoder.defaultAbiCoder().encode(["string", "uint256"], ["merkle", i])),
      valueComputeAlgHash: ethers.keccak256(ethers.AbiCoder.defaultAbiCoder().encode(["string", "uint256"], ["alg", i])),
      r: ethers.keccak256(ethers.AbiCoder.defaultAbiCoder().encode(["string", "uint256"], ["r", i])),
      s: ethers.keccak256(ethers.AbiCoder.defaultAbiCoder().encode(["string", "uint256"], ["s", i])),
      v: 27,
    };
  }

  // Convert to tuple format for ethers struct encoding: [[timestampNs, quantizedValue], id, merkle, alg, r, s, v]
  function toTuple(entry: ReturnType<typeof makeEntry>) {
    return [
      [entry.timestampNs, entry.quantizedValue],
      entry.id,
      entry.publisherMerkleRoot,
      entry.valueComputeAlgHash,
      entry.r,
      entry.s,
      entry.v,
    ];
  }

  // Convert to named struct for contract calls (ethers resolves field names from ABI)
  function toStruct(entry: ReturnType<typeof makeEntry>) {
    return {
      temporalNumericValue: {
        timestampNs: entry.timestampNs,
        quantizedValue: entry.quantizedValue,
      },
      id: entry.id,
      publisherMerkleRoot: entry.publisherMerkleRoot,
      valueComputeAlgHash: entry.valueComputeAlgHash,
      r: entry.r,
      s: entry.s,
      v: entry.v,
    };
  }

  async function deployFixture() {
    const [owner] = await ethers.getSigners();
    const UpgradeableStork = await ethers.getContractFactory("UpgradeableStork");
    const deployed = await upgrades.deployProxy(UpgradeableStork, [
      owner.address,
      "0xC4A02e7D370402F4afC36032076B05e74FF81786",
      60,
      1,
    ]);
    return { deployed, owner };
  }

  describe("pack/unpack roundtrip", function () {
    it("Should roundtrip a single entry", function () {
      const entry = makeEntry(0);
      const words = packEntry(entry);
      expect(words.length).to.equal(WORDS_PER_ENTRY);

      const unpacked = unpackWord0(words[0]);
      expect(unpacked.timestampNs).to.equal(entry.timestampNs);
      expect(unpacked.quantizedValue).to.equal(entry.quantizedValue);
      expect(unpacked.v).to.equal(entry.v);
    });

    it("Should roundtrip negative quantizedValue", function () {
      const entry = { ...makeEntry(0), quantizedValue: -1n };
      const words = packEntry(entry);
      const unpacked = unpackWord0(words[0]);
      expect(unpacked.quantizedValue).to.equal(-1n);
    });

    it("Should roundtrip int192 extremes", function () {
      const int192Min = -(1n << 191n);
      const int192Max = (1n << 191n) - 1n;

      const entryMin = { ...makeEntry(0), quantizedValue: int192Min };
      const entryMax = { ...makeEntry(0), quantizedValue: int192Max };

      expect(unpackWord0(packEntry(entryMin)[0]).quantizedValue).to.equal(int192Min);
      expect(unpackWord0(packEntry(entryMax)[0]).quantizedValue).to.equal(int192Max);
    });

    it("Should roundtrip v=27 and v=28", function () {
      const entry27 = { ...makeEntry(0), v: 27 };
      const entry28 = { ...makeEntry(0), v: 28 };

      expect(unpackWord0(packEntry(entry27)[0]).v).to.equal(27);
      expect(unpackWord0(packEntry(entry28)[0]).v).to.equal(28);
    });

    it("Should roundtrip 3 entries", function () {
      for (let i = 0; i < 3; i++) {
        const entry = makeEntry(i);
        const words = packEntry(entry);
        const unpacked = unpackWord0(words[0]);
        expect(unpacked.timestampNs).to.equal(entry.timestampNs);
        expect(unpacked.quantizedValue).to.equal(entry.quantizedValue);
        expect(unpacked.v).to.equal(entry.v);
        expect(words[1]).to.equal(BigInt(entry.id));
        expect(words[2]).to.equal(BigInt(entry.publisherMerkleRoot));
        expect(words[3]).to.equal(BigInt(entry.valueComputeAlgHash));
        expect(words[4]).to.equal(BigInt(entry.r));
        expect(words[5]).to.equal(BigInt(entry.s));
      }
    });
  });

  describe("on-chain compress/decode via contract", function () {
    it("Should compress struct array to packed words", async function () {
      const { deployed } = await loadFixture(deployFixture);

      const entries = Array.from({ length: 3 }, (_, i) => toStruct(makeEntry(i)));
      const packed: bigint[] = await deployed.compress(entries);

      expect(packed.length).to.equal(3 * WORDS_PER_ENTRY);

      // Verify each entry matches off-chain packing
      for (let i = 0; i < 3; i++) {
        const offchain = packEntry(makeEntry(i));
        for (let j = 0; j < WORDS_PER_ENTRY; j++) {
          expect(packed[i * WORDS_PER_ENTRY + j]).to.equal(offchain[j]);
        }
      }
    });
  });

  describe("calldata size savings", function () {
    it("Should save exactly 640 bytes for 10 entries (24%)", function () {
      const N = 10;

      // ABI-encoded calldata for updateTemporalNumericValuesV1
      const iface = new ethers.Interface([
        "function updateTemporalNumericValuesV1(tuple(tuple(uint64 timestampNs, int192 quantizedValue) temporalNumericValue, bytes32 id, bytes32 publisherMerkleRoot, bytes32 valueComputeAlgHash, bytes32 r, bytes32 s, uint8 v)[] updateData)",
        "function updateTemporalNumericValuesV1Packed(uint256[] packedData)",
      ]);

      const tupleEntries = Array.from({ length: N }, (_, i) => toTuple(makeEntry(i)));
      const abiCalldata = iface.encodeFunctionData("updateTemporalNumericValuesV1", [tupleEntries]);

      // Packed calldata
      const allWords: bigint[] = [];
      for (let i = 0; i < N; i++) {
        allWords.push(...packEntry(makeEntry(i)));
      }
      const packedCalldata = iface.encodeFunctionData("updateTemporalNumericValuesV1Packed", [allWords]);

      const abiSize = (abiCalldata.length - 2) / 2; // hex string to bytes (minus 0x prefix)
      const packedSize = (packedCalldata.length - 2) / 2;
      const saved = abiSize - packedSize;

      console.log(`    ABI calldata:    ${abiSize} bytes`);
      console.log(`    Packed calldata: ${packedSize} bytes`);
      console.log(`    Saved:           ${saved} bytes (${Math.round(saved * 100 / abiSize)}%)`);

      expect(saved).to.equal(640);
      expect(packedSize).to.be.lessThan(abiSize);
    });
  });

  describe("edge cases", function () {
    it("Should reject packed array with invalid length (not divisible by 6)", async function () {
      const { deployed } = await loadFixture(deployFixture);

      const badWords = Array(5).fill(0n);
      await expect(
        deployed.updateTemporalNumericValuesV1Packed(badWords, { value: 1 })
      ).to.be.revertedWithCustomError(deployed, "InvalidLength");
    });
  });
});
