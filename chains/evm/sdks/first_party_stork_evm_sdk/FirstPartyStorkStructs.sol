// SPDX-License-Identifier: Apache-2.0
pragma solidity >=0.8.24 <0.9.0;

import "@storknetwork/stork-evm-sdk/StorkStructs.sol";

/// @title FirstPartyStorkStructs
/// @notice Data structures used by the First Party Stork protocol
library FirstPartyStorkStructs {
    /// @notice Input structure for updating temporal numeric values from publishers
    struct PublisherTemporalNumericValueInput {
        StorkStructs.TemporalNumericValue temporalNumericValue;
        address pubKey;
        string assetPairId;
        bool storeHistorical;
        bytes32 r;
        bytes32 s;
        uint8 v;
    }

    /// @notice Publisher user configuration
    struct PublisherUser {
        address pubKey;
        uint256 singleUpdateFee;
    }
}
