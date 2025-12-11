// SPDX-License-Identifier: Apache 2
pragma solidity ^0.8.28;

import "@storknetwork/stork-evm-sdk/StorkStructs.sol";

/// @title StorkFastStructs
/// @notice Data structures used by the Stork Fast oracle contract
library StorkFastStructs {
    /// @notice Represents a single update to a temporal numeric value
    struct Asset {
        /// @dev The asset ID of the update
        uint16 assetID;
        /// @dev The Temporal Numeric Value of the update
        StorkStructs.TemporalNumericValue temporalNumericValue;
    }
}
