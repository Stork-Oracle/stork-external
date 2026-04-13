// SPDX-License-Identifier: Apache 2

pragma solidity 0.8.24;

import "forge-std/src/Test.sol";
import "./LibCodec.sol";
import "./StorkStructs.sol";

contract LibCodecHarness {
    function encode(
        StorkStructs.TemporalNumericValueInput[] memory inputs
    ) external pure returns (uint256[] memory) {
        return LibCodec.encode(inputs);
    }

    function decode(
        uint256[] calldata words
    ) external pure returns (StorkStructs.TemporalNumericValueInput[] memory) {
        return LibCodec.decode(words);
    }

    function roundtrip(
        uint256[] calldata words
    ) external pure returns (uint256[] memory) {
        StorkStructs.TemporalNumericValueInput[] memory decoded = LibCodec.decode(words);
        return LibCodec.encode(decoded);
    }
}

contract LibCodecTest is Test {
    LibCodecHarness harness;

    function setUp() public {
        harness = new LibCodecHarness();
    }

    // ── Fuzz: encode → decode roundtrip ─────────────────────────────────

    function testFuzz_roundtrip_single(
        uint64 timestampNs,
        int192 quantizedValue,
        bytes32 id,
        bytes32 publisherMerkleRoot,
        bytes32 valueComputeAlgHash,
        bytes32 r,
        bytes32 s,
        uint8 v
    ) public view {
        v = uint8(bound(v, 27, 28));
        // timestampNs must fit in 63 bits (bit 63 reserved for v_flag)
        timestampNs = uint64(bound(timestampNs, 0, uint64((1 << 63) - 1)));

        StorkStructs.TemporalNumericValueInput[] memory inputs =
            new StorkStructs.TemporalNumericValueInput[](1);

        inputs[0] = StorkStructs.TemporalNumericValueInput({
            temporalNumericValue: StorkStructs.TemporalNumericValue({
                timestampNs: timestampNs,
                quantizedValue: quantizedValue
            }),
            id: id,
            publisherMerkleRoot: publisherMerkleRoot,
            valueComputeAlgHash: valueComputeAlgHash,
            r: r,
            s: s,
            v: v
        });

        uint256[] memory encoded = harness.encode(inputs);
        assertEq(encoded.length, 6, "encoded length");

        StorkStructs.TemporalNumericValueInput[] memory decoded = harness.decode(encoded);
        assertEq(decoded.length, 1, "decoded length");

        _assertInputEq(inputs[0], decoded[0]);
    }

    function testFuzz_roundtrip_multi(
        uint64[3] memory timestamps,
        int192[3] memory values,
        bytes32[3] memory ids,
        uint8[3] memory vs
    ) public view {
        StorkStructs.TemporalNumericValueInput[] memory inputs =
            new StorkStructs.TemporalNumericValueInput[](3);

        for (uint256 i; i < 3; ++i) {
            vs[i] = uint8(bound(vs[i], 27, 28));
            timestamps[i] = uint64(bound(timestamps[i], 0, uint64((1 << 63) - 1)));
            inputs[i] = StorkStructs.TemporalNumericValueInput({
                temporalNumericValue: StorkStructs.TemporalNumericValue({
                    timestampNs: timestamps[i],
                    quantizedValue: values[i]
                }),
                id: ids[i],
                publisherMerkleRoot: bytes32(uint256(i + 100)),
                valueComputeAlgHash: bytes32(uint256(i + 200)),
                r: bytes32(uint256(i + 300)),
                s: bytes32(uint256(i + 400)),
                v: vs[i]
            });
        }

        uint256[] memory encoded = harness.encode(inputs);
        assertEq(encoded.length, 18, "encoded length for 3 entries");

        StorkStructs.TemporalNumericValueInput[] memory decoded = harness.decode(encoded);
        assertEq(decoded.length, 3, "decoded length");

        for (uint256 i; i < 3; ++i) {
            _assertInputEq(inputs[i], decoded[i]);
        }
    }

    // ── Fuzz: manual word construction → decode → encode roundtrip ──────

    function testFuzz_calldata_roundtrip(
        uint64 timestampNs,
        int192 quantizedValue,
        bytes32 id,
        bytes32 merkle,
        bytes32 algHash,
        bytes32 r,
        bytes32 s,
        uint8 v
    ) public view {
        v = uint8(bound(v, 27, 28));
        timestampNs = uint64(bound(timestampNs, 0, uint64((1 << 63) - 1)));

        uint256[] memory words = new uint256[](6);
        words[0] = (uint256(v - 27) << 255)
            | (uint256(timestampNs) << 192)
            | uint256(uint192(quantizedValue));
        words[1] = uint256(id);
        words[2] = uint256(merkle);
        words[3] = uint256(algHash);
        words[4] = uint256(r);
        words[5] = uint256(s);

        uint256[] memory result = harness.roundtrip(words);

        for (uint256 i; i < 6; ++i) {
            assertEq(result[i], words[i], string.concat("word mismatch at ", vm.toString(i)));
        }
    }

    // ── Edge cases ──────────────────────────────────────────────────────

    function test_empty() public view {
        uint256[] memory empty = new uint256[](0);
        StorkStructs.TemporalNumericValueInput[] memory decoded = harness.decode(empty);
        assertEq(decoded.length, 0);
    }

    function test_invalidLength_reverts() public {
        uint256[] memory bad = new uint256[](5); // not divisible by 6
        vm.expectRevert(LibCodec.InvalidLength.selector);
        harness.decode(bad);
    }

    function test_timestampOverflow_reverts() public {
        // timestampNs == 2^63 should revert
        StorkStructs.TemporalNumericValueInput[] memory inputs =
            new StorkStructs.TemporalNumericValueInput[](1);
        inputs[0] = _makeInput(uint64(1 << 63), int192(0), bytes32(0), 27);
        vm.expectRevert(LibCodec.TimestampOverflow.selector);
        harness.encode(inputs);
    }

    function testFuzz_timestampOverflow_reverts(uint64 timestampNs) public {
        // Any value >= 2^63 must revert
        vm.assume(timestampNs >= uint64(1 << 63));
        StorkStructs.TemporalNumericValueInput[] memory inputs =
            new StorkStructs.TemporalNumericValueInput[](1);
        inputs[0] = _makeInput(timestampNs, int192(0), bytes32(0), 27);
        vm.expectRevert(LibCodec.TimestampOverflow.selector);
        harness.encode(inputs);
    }

    function test_invalidV_reverts() public {
        StorkStructs.TemporalNumericValueInput[] memory inputs =
            new StorkStructs.TemporalNumericValueInput[](1);
        inputs[0] = _makeInput(1000, int192(0), bytes32(0), 26); // v=26 invalid
        vm.expectRevert(LibCodec.InvalidV.selector);
        harness.encode(inputs);
    }

    function testFuzz_negative_quantizedValue(
        uint64 ts,
        bytes32 id
    ) public view {
        ts = uint64(bound(ts, 0, uint64((1 << 63) - 1)));
        int192 negValue = -1;

        StorkStructs.TemporalNumericValueInput[] memory inputs =
            new StorkStructs.TemporalNumericValueInput[](1);

        inputs[0] = StorkStructs.TemporalNumericValueInput({
            temporalNumericValue: StorkStructs.TemporalNumericValue({
                timestampNs: ts,
                quantizedValue: negValue
            }),
            id: id,
            publisherMerkleRoot: bytes32(0),
            valueComputeAlgHash: bytes32(0),
            r: bytes32(0),
            s: bytes32(0),
            v: 27
        });

        uint256[] memory encoded = harness.encode(inputs);
        StorkStructs.TemporalNumericValueInput[] memory decoded = harness.decode(encoded);

        assertEq(decoded[0].temporalNumericValue.quantizedValue, negValue, "negative value roundtrip");
    }

    function testFuzz_int192_extremes(uint64 ts) public view {
        ts = uint64(bound(ts, 0, uint64((1 << 63) - 1)));
        int192 minVal = type(int192).min;
        int192 maxVal = type(int192).max;

        StorkStructs.TemporalNumericValueInput[] memory inputs =
            new StorkStructs.TemporalNumericValueInput[](2);

        inputs[0] = _makeInput(ts, minVal, bytes32(uint256(1)), 27);
        inputs[1] = _makeInput(ts, maxVal, bytes32(uint256(2)), 28);

        uint256[] memory encoded = harness.encode(inputs);
        StorkStructs.TemporalNumericValueInput[] memory decoded = harness.decode(encoded);

        assertEq(decoded[0].temporalNumericValue.quantizedValue, minVal, "int192 min");
        assertEq(decoded[1].temporalNumericValue.quantizedValue, maxVal, "int192 max");
        assertEq(decoded[0].v, 27, "v=27");
        assertEq(decoded[1].v, 28, "v=28");
    }

    // ── Calldata size comparison ───────────────────────────────────────

    function test_calldataSavings_10entries() public {
        uint256 N = 10;

        StorkStructs.TemporalNumericValueInput[] memory inputs =
            new StorkStructs.TemporalNumericValueInput[](N);
        for (uint256 i; i < N; ++i) {
            inputs[i] = StorkStructs.TemporalNumericValueInput({
                temporalNumericValue: StorkStructs.TemporalNumericValue({
                    timestampNs: uint64(1700000000000000000 + i),
                    quantizedValue: int192(int256(1e18 * int256(i + 1)))
                }),
                id: bytes32(uint256(keccak256(abi.encode("id", i)))),
                publisherMerkleRoot: bytes32(uint256(keccak256(abi.encode("merkle", i)))),
                valueComputeAlgHash: bytes32(uint256(keccak256(abi.encode("alg", i)))),
                r: bytes32(uint256(keccak256(abi.encode("r", i)))),
                s: bytes32(uint256(keccak256(abi.encode("s", i)))),
                v: 27
            });
        }

        // ── Original ABI-encoded calldata ──
        bytes memory abiCalldata = abi.encodeWithSignature(
            "updateTemporalNumericValuesV1(((uint64,int192),bytes32,bytes32,bytes32,bytes32,bytes32,uint8)[])",
            inputs
        );

        // ── Packed calldata ──
        uint256[] memory packed = harness.encode(inputs);
        bytes memory packedCalldata = abi.encodeWithSignature(
            "updateTemporalNumericValuesV1Packed(uint256[])",
            packed
        );

        uint256 abiSize = abiCalldata.length;
        uint256 packedSize = packedCalldata.length;
        uint256 saved = abiSize - packedSize;
        uint256 pctSaved = (saved * 100) / abiSize;

        // Count zero vs non-zero bytes for L2 gas estimation
        // EIP-2028: 4 gas per zero byte, 16 gas per non-zero byte
        uint256 abiGas = _calldataGas(abiCalldata);
        uint256 packedGas = _calldataGas(packedCalldata);
        uint256 gasSaved = abiGas - packedGas;
        uint256 gasPctSaved = (gasSaved * 100) / abiGas;

        emit log_named_uint("ABI calldata (bytes)", abiSize);
        emit log_named_uint("Packed calldata (bytes)", packedSize);
        emit log_named_uint("Saved (bytes)", saved);
        emit log_named_uint("Saved bytes (%)", pctSaved);
        emit log("");
        emit log_named_uint("ABI calldata gas (zero=4, nonzero=16)", abiGas);
        emit log_named_uint("Packed calldata gas", packedGas);
        emit log_named_uint("Saved gas", gasSaved);
        emit log_named_uint("Saved gas (%)", gasPctSaved);

        emit log("");
        emit log("--- ABI breakdown ---");
        emit log_named_uint("  ABI zero bytes", _countZeroBytes(abiCalldata));
        emit log_named_uint("  ABI non-zero bytes", abiSize - _countZeroBytes(abiCalldata));
        emit log("--- Packed breakdown ---");
        emit log_named_uint("  Packed zero bytes", _countZeroBytes(packedCalldata));
        emit log_named_uint("  Packed non-zero bytes", packedSize - _countZeroBytes(packedCalldata));

        assertLt(packedSize, abiSize, "packed should be smaller");
    }

    // ── Helpers ──────────────────────────────────────────────────────────

    function _makeInput(
        uint64 ts,
        int192 qv,
        bytes32 id,
        uint8 v
    ) internal pure returns (StorkStructs.TemporalNumericValueInput memory) {
        return StorkStructs.TemporalNumericValueInput({
            temporalNumericValue: StorkStructs.TemporalNumericValue({
                timestampNs: ts,
                quantizedValue: qv
            }),
            id: id,
            publisherMerkleRoot: bytes32(0),
            valueComputeAlgHash: bytes32(0),
            r: bytes32(0),
            s: bytes32(0),
            v: v
        });
    }

    function _calldataGas(bytes memory data) internal pure returns (uint256 gas_) {
        for (uint256 i; i < data.length; ++i) {
            gas_ += data[i] == 0 ? 4 : 16;
        }
    }

    function _countZeroBytes(bytes memory data) internal pure returns (uint256 count) {
        for (uint256 i; i < data.length; ++i) {
            if (data[i] == 0) count++;
        }
    }

    function _assertInputEq(
        StorkStructs.TemporalNumericValueInput memory a,
        StorkStructs.TemporalNumericValueInput memory b
    ) internal pure {
        assert(a.temporalNumericValue.timestampNs == b.temporalNumericValue.timestampNs);
        assert(a.temporalNumericValue.quantizedValue == b.temporalNumericValue.quantizedValue);
        assert(a.id == b.id);
        assert(a.publisherMerkleRoot == b.publisherMerkleRoot);
        assert(a.valueComputeAlgHash == b.valueComputeAlgHash);
        assert(a.r == b.r);
        assert(a.s == b.s);
        assert(a.v == b.v);
    }
}
