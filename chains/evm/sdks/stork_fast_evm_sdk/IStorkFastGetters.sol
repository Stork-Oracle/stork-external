// SPDX-License-Identifier: Apache 2
pragma solidity ^0.8.28;

interface IStorkFastGetters {
    /// @notice Retrieves the verification fee in wei
    /// @return uint The verification fee in wei
    function verificationFeeInWei() external view returns (uint);

    /// @notice Retrieves the Stork Fast Address used for verification
    /// @return address The Stork Fast Address
    function storkFastAddress() external view returns (address);
}
