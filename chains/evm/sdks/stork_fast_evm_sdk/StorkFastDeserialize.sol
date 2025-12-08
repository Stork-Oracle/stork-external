// SPDX-License-Identifier: Apache 2
pragma solidity ^0.8.28;

import "@storknetwork/stork-evm-sdk/StorkStructs.sol";
import "./StorkFastStructs.sol";
import "./StorkFastErrors.sol";

/// @title StorkFastDeserialize
/// @notice Library for deserializing Stork Fast signed ECDSA payloads
/// @dev This library provides functions for deserializing signed ECDSA payloads
library StorkFastDeserialize {
    /// @dev The number of bytes representing the taxonomy ID in the signed ECDSA payload
    uint public constant TAXONOMY_ID_BYTES = 2;
    /// @dev The number of bytes representing the timestamp in nanoseconds
    uint public constant TIMESTAMP_NS_BYTES = 8;
    /// @dev The number of bytes representing an asset ID
    uint public constant ASSET_ID_BYTES = 2;
    /// @dev The number of bytes representing a quantized value
    uint public constant QUANTIZED_VALUE_BYTES = 16;

    /// @dev The number of bytes representing the signature in the signed ECDSA payload
    uint public constant SIGNED_ECDSA_SIGNATURE_BYTES = 65;
    /// @dev The number of bytes representing an asset in the signed ECDSA payload
    uint public constant SIGNED_ECDSA_ASSET_BYTES =
        ASSET_ID_BYTES + QUANTIZED_VALUE_BYTES;

    uint public constant SIGNED_ECDSA_SIGNATURE_OFFSET = 0;
    uint public constant SIGNED_ECDSA_TAXONOMY_ID_OFFSET =
        SIGNED_ECDSA_SIGNATURE_OFFSET + SIGNED_ECDSA_SIGNATURE_BYTES;
    uint public constant SIGNED_ECDSA_TIMESTAMP_NS_OFFSET =
        SIGNED_ECDSA_TAXONOMY_ID_OFFSET + TAXONOMY_ID_BYTES;
    uint public constant SIGNED_ECDSA_ASSETS_OFFSET =
        SIGNED_ECDSA_TIMESTAMP_NS_OFFSET + TIMESTAMP_NS_BYTES;

    /// @notice Splits a signed ECDSA payload into a signature and a verifiable payload
    /// @param payload The signed ECDSA payload
    /// @return signature The signature bytes
    /// @return verifiablePayload The verifiable payload bytes, excluding the signature
    function splitSignedECDSAPayload(
        bytes calldata payload
    )
        internal
        pure
        returns (bytes memory signature, bytes memory verifiablePayload)
    {
        signature = payload[
            SIGNED_ECDSA_SIGNATURE_OFFSET:SIGNED_ECDSA_TAXONOMY_ID_OFFSET
        ];
        verifiablePayload = payload[
            SIGNED_ECDSA_TAXONOMY_ID_OFFSET:payload.length
        ];
    }

    /// @notice Deserializes the header of a signed ECDSA payload
    /// @param payload The signed ECDSA payload
    /// @return signature The signature bytes
    /// @return taxonomyID The taxonomy ID as a 16-bit unsigned integer
    /// @return timestampNs The timestamp in nanoseconds as a 64-bit unsigned integer
    function deserializeSignedECDSAPayloadHeader(
        bytes calldata payload
    )
        internal
        pure
        validPayloadLength(payload)
        returns (bytes memory signature, uint16 taxonomyID, uint64 timestampNs)
    {
        bytes memory _signature = payload[
            SIGNED_ECDSA_SIGNATURE_OFFSET:SIGNED_ECDSA_TAXONOMY_ID_OFFSET
        ];
        uint16 _taxonomyID = uint16(
            bytes2(
                payload[
                    SIGNED_ECDSA_TAXONOMY_ID_OFFSET:SIGNED_ECDSA_TIMESTAMP_NS_OFFSET
                ]
            )
        );
        uint64 _timestampNs = uint64(
            bytes8(
                payload[
                    SIGNED_ECDSA_TIMESTAMP_NS_OFFSET:SIGNED_ECDSA_ASSETS_OFFSET
                ]
            )
        );

        return (_signature, _taxonomyID, _timestampNs);
    }

    /// @notice Deserializes the assets from a signed ECDSA payload
    /// @param payload The signed ECDSA payload
    /// @return updates The updates as an array of StorkFastStructs.Update
    function deserializeValuesFromSignedECDSAPayload(
        bytes calldata payload
    )
        internal
        pure
        validPayloadLength(payload)
        returns (StorkFastStructs.Update[] memory updates)
    {
        uint64 timestampNs = uint64(
            bytes8(
                payload[
                    SIGNED_ECDSA_TIMESTAMP_NS_OFFSET:SIGNED_ECDSA_ASSETS_OFFSET
                ]
            )
        );

        uint256 len = payload.length;
        uint16 numUpdates = uint16(
            (len - SIGNED_ECDSA_ASSETS_OFFSET) / SIGNED_ECDSA_ASSET_BYTES
        );

        updates = new StorkFastStructs.Update[](numUpdates);

        uint16 assetIndex = 0;
        for (
            uint i = SIGNED_ECDSA_ASSETS_OFFSET;
            i < len;
            i += SIGNED_ECDSA_ASSET_BYTES
        ) {
            uint16 assetID = uint16(bytes2(payload[i:i + ASSET_ID_BYTES]));

            int128 quantizedValue = int128(
                uint128(
                    bytes16(
                        payload[
                            i + ASSET_ID_BYTES:i +
                                ASSET_ID_BYTES +
                                QUANTIZED_VALUE_BYTES
                        ]
                    )
                )
            );

            updates[assetIndex] = StorkFastStructs.Update(
                assetID,
                StorkStructs.TemporalNumericValue(
                    timestampNs,
                    int192(quantizedValue)
                )
            );
            assetIndex++;
        }

        return updates;
    }

    modifier validPayloadLength(bytes calldata payload) {
        if (payload.length < SIGNED_ECDSA_ASSETS_OFFSET) {
            revert StorkFastErrors.InvalidPayload();
        }
        if (
            (payload.length - SIGNED_ECDSA_ASSETS_OFFSET) %
                SIGNED_ECDSA_ASSET_BYTES !=
            0
        ) {
            revert StorkFastErrors.InvalidPayload();
        }
        _;
    }
}
