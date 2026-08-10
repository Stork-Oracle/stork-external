// SPDX-License-Identifier: Apache 2
pragma solidity ^0.8.28;

import "./StorkFastStructs.sol";

/// @title IStorkFastEvents
/// @notice Events emitted by the Stork Fast contract
interface IStorkFastEvents {
    /// @notice Emitted when a signer address is added
    /// @param signerAddress The added signer address
    event SignerAddressAdded(address indexed signerAddress);

    /// @notice Emitted when a signer address is removed
    /// @param signerAddress The removed signer address
    event SignerAddressRemoved(address indexed signerAddress);

    /// @notice Emitted when the verification fee in wei is updated
    /// @param newVerificationFeeInWei The new verification fee in wei
    event VerificationFeeUpdate(uint256 newVerificationFeeInWei);
}
