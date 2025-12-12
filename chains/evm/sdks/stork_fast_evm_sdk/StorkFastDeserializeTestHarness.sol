// SPDX-License-Identifier: Apache 2
pragma solidity ^0.8.28;

import "./StorkFastDeserialize.sol";

/// @title StorkFastDeserializeTestHarness
/// @dev Contract wrapper around StorkFastDeserialize for unit testing
contract StorkFastDeserializeTestHarness {
    /// @dev Wrapper around StorkFastDeserialize.splitSignedECDSAPayload
    function splitSignedECDSAPayload(
        bytes calldata payload
    )
        public
        pure
        returns (bytes memory signature, bytes memory verifiablePayload)
    {
        return StorkFastDeserialize.splitSignedECDSAPayload(payload);
    }

    /// @dev Wrapper around StorkFastDeserialize.deserializeSignedECDSAPayloadHeader
    function deserializeSignedECDSAPayloadHeader(
        bytes calldata payload
    )
        public
        pure
        returns (bytes memory signature, uint16 taxonomyID, uint64 timestampNs)
    {
        return
            StorkFastDeserialize.deserializeSignedECDSAPayloadHeader(payload);
    }

    /// @dev Wrapper around StorkFastDeserialize.deserializeAssetsFromSignedECDSAPayload
    function deserializeAssetsFromSignedECDSAPayload(
        bytes calldata payload
    ) public pure returns (StorkFastStructs.Asset[] memory assets) {
        return
            StorkFastDeserialize.deserializeAssetsFromSignedECDSAPayload(
                payload
            );
    }
}
