// SPDX-License-Identifier: Apache 2
pragma solidity ^0.8.28;

/// @title StorkFastErrors
/// @notice Errors use in the Stork Fast contract
library StorkFastErrors {
    /// @notice Insufficient fee is paid to the method
    error InsufficientFee();

    /// @notice Signature is invalid
    error InvalidSignature();

    // @notice Payload is invalid
    error InvalidPayload();
}
