// SPDX-License-Identifier: Apache 2

pragma solidity >=0.8.24 <0.9.0;

import "@storknetwork/stork-evm-sdk/StorkStructs.sol";

/// @title LibCodec
/// @notice Flat uint256[] codec for TemporalNumericValueInput to minimize calldata cost.
/// @dev Each entry is 6 consecutive uint256 words:
///   word[0]: v_flag (1 bit) | timestampNs (63 bits) | quantizedValue (192 bits)
///            v_flag = v - 27 (bit 255), timestampNs in bits [254:192], quantizedValue in bits [191:0]
///   word[1]: id                  (bytes32)
///   word[2]: publisherMerkleRoot (bytes32)
///   word[3]: valueComputeAlgHash (bytes32)
///   word[4]: r                   (bytes32)
///   word[5]: s                   (bytes32)
///
///   Calldata per entry: 6 × 32 = 192 bytes.
library LibCodec {
    uint256 internal constant WORDS_PER_ENTRY = 6;
    uint256 internal constant V_FLAG_BIT = 1 << 255;

    error InvalidLength();
    error InvalidV();
    error TimestampOverflow();

    /// @notice Encode an array of TemporalNumericValueInput into a flat uint256 array.
    /// @dev Intended for off-chain use to produce compact calldata.
    function encode(
        StorkStructs.TemporalNumericValueInput[] memory inputs
    ) internal pure returns (uint256[] memory words) {
        uint256 len = inputs.length;
        words = new uint256[](len * WORDS_PER_ENTRY);

        for (uint256 i; i < len; ++i) {
            uint256 base = i * WORDS_PER_ENTRY;
            StorkStructs.TemporalNumericValueInput memory inp = inputs[i];

            uint8 v = inp.v;
            if (v != 27 && v != 28) revert InvalidV();
            if (inp.temporalNumericValue.timestampNs >= 1 << 63) revert TimestampOverflow();

            // bit 255: v_flag (v - 27), bits [254:192]: timestampNs, bits [191:0]: quantizedValue
            words[base] = (uint256(v - 27) << 255)
                | (uint256(inp.temporalNumericValue.timestampNs) << 192)
                | uint256(uint192(inp.temporalNumericValue.quantizedValue));

            words[base + 1] = uint256(inp.id);
            words[base + 2] = uint256(inp.publisherMerkleRoot);
            words[base + 3] = uint256(inp.valueComputeAlgHash);
            words[base + 4] = uint256(inp.r);
            words[base + 5] = uint256(inp.s);
        }
    }

    /// @notice Decode a flat uint256 calldata array into TemporalNumericValueInput structs.
    /// @dev Used on-chain to unpack compact calldata.
    function decode(
        uint256[] calldata words
    ) internal pure returns (StorkStructs.TemporalNumericValueInput[] memory inputs) {
        if (words.length % WORDS_PER_ENTRY != 0) revert InvalidLength();
        uint256 len = words.length / WORDS_PER_ENTRY;
        inputs = new StorkStructs.TemporalNumericValueInput[](len);

        for (uint256 i; i < len; ++i) {
            uint256 base = i * WORDS_PER_ENTRY;
            uint256 w0 = words[base];

            inputs[i] = StorkStructs.TemporalNumericValueInput({
                temporalNumericValue: StorkStructs.TemporalNumericValue({
                    // bits [254:192], mask out bit 255
                    timestampNs: uint64((w0 >> 192) & 0x7FFFFFFFFFFFFFFF),
                    quantizedValue: int192(uint192(w0))
                }),
                id: bytes32(words[base + 1]),
                publisherMerkleRoot: bytes32(words[base + 2]),
                valueComputeAlgHash: bytes32(words[base + 3]),
                r: bytes32(words[base + 4]),
                s: bytes32(words[base + 5]),
                // bit 255 → 0 or 1, add 27
                v: uint8((w0 >> 255) + 27)
            });
        }
    }
}
