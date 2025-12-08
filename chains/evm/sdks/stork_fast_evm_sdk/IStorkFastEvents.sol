// SPDX-License-Identifier: Apache 2
pragma solidity ^0.8.28;

import "./StorkFastStructs.sol";

/// @title IStorkFastEvents
/// @notice Events emitted by the Stork Fast contract
interface IStorkFastEvents {
    /// @notice Emitted when the Stork Fast Address key is updated
    /// @param newStorkFastAddress The new Stork public key
    event StorkFastAddressUpdate(address indexed newStorkFastAddress);

    /// @notice Emitted when the verification fee in wei is updated
    /// @param newVerificationFeeInWei The new verification fee in wei
    event VerificationFeeUpdate(uint256 newVerificationFeeInWei);
}
