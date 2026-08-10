// SPDX-License-Identifier: Apache 2
pragma solidity ^0.8.28;

interface IStorkFastGetters {
    /// @notice Retrieves the verification fee in wei
    /// @return uint The verification fee in wei
    function verificationFeeInWei() external view returns (uint);

    /// @notice Retrieves the list of valid signer addresses used for verification
    /// @return address[] The list of valid signer addresses
    function getSignerAddresses() external view returns (address[] memory);

    /// @notice Checks whether an address is a valid signer address
    /// @param signerAddress The address to check
    /// @return bool True if the address is a valid signer address
    function isValidSignerAddress(
        address signerAddress
    ) external view returns (bool);
}
