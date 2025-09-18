// SPDX-License-Identifier: Apache-2.0
pragma solidity >=0.8.24 <0.9.0;

/// @title StorkStructs
/// @notice Data structures used by the Stork protocol
library SelfServeStorkStructs {
    /// @notice Represents a temporal numeric value with timestamp and quantized value
    struct TemporalNumericValue {
        /// @dev Nanosecond level precision timestamp of latest publisher update in batch
        uint64 timestampNs; // 8 bytes
        /// @dev Should be able to hold all necessary numbers (up to 6277101735386680763835789423207666416102355444464034512895)
        int192 quantizedValue; // 24 bytes
    }

    /// @notice Input structure for updating temporal numeric values
    struct PublisherTemporalNumericValueInput {
        TemporalNumericValue temporalNumericValue;
        address pubKey;
        string assetPairId;
        bytes32 r;
        bytes32 s;
        uint8 v;
    }

    struct PublisherUser {
        address pubKey;
        uint256 singleUpdateFee;
    }
}
