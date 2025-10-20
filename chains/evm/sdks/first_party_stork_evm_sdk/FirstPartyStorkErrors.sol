// SPDX-License-Identifier: Apache-2.0
pragma solidity >=0.8.24 <0.9.0;

/// @title FirstPartyStorkErrors
/// @notice Error definitions for the First Party Stork protocol
library FirstPartyStorkErrors {
    /// @notice Insufficient fee is paid to the method
    /// @dev Error code: 0x025dbdd4
    error InsufficientFee();
    
    /// @notice There is no fresh update, whereas expected fresh updates
    /// @dev Error code: 0xde2c57fa
    error NoFreshUpdate();
    
    /// @notice Not found
    /// @dev Error code: 0xc5723b51
    error NotFound();
    
    /// @notice Signature is invalid
    /// @dev Error code: 0x8baa579f
    error InvalidSignature();
}
