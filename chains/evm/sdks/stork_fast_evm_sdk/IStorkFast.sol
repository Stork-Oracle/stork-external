// SPDX-License-Identifier: Apache 2
pragma solidity ^0.8.28;

import "./StorkFastStructs.sol";

/// @title StorkFast
/// @notice Interface for the Stork Fast oracle contract
/// @dev This interface provides access to Stork Fast's verification functions
interface IStorkFast {
    /// @notice Verifies a signed ECDSA update payload
    /// @param payload The signed ECDSA payload
    /// @dev Requires sufficient fee
    /// @dev Reverts with InsufficientFee if the provided fee is less than the required amount
    /// @return bool True if the signature is valid
    function verifySignedECDSAPayload(
        bytes calldata payload
    ) external payable returns (bool);

    /// @notice Verifies and deserializes a signed ECDSA update payload
    /// @param payload The signed ECDSA payload
    /// @dev Requires sufficient fee
    /// @dev Reverts with InsufficientFee if the provided fee is less than the required amount
    /// @dev Reverts with InvalidSignature if signature verification fails
    /// @return updates Array of updates
    function verifyAndDeserializeSignedECDSAPayload(
        bytes calldata payload
    ) external payable returns (StorkFastStructs.Update[] memory updates);
}
