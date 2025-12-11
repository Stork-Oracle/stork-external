// SPDX-License-Identifier: Apache 2
pragma solidity ^0.8.28;

/// @title StorkFastErrors
/// @notice Errors use in the Stork Fast contract
library StorkFastErrors {
    /// @notice Insufficient fee is paid to the method
    /// @dev Error code: 0x025dbdd4
    error InsufficientFee();

    /// @notice Signature is invalid
    /// @dev Error code: 0x8baa579f
    error InvalidSignature();

    /// @notice Payload is invalid
    /// @dev Error code: 0x7c6953f9
    error InvalidPayload();
}
