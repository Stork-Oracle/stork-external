// SPDX-License-Identifier: Apache-2.0

pragma solidity >=0.8.24 <0.9.0;

interface IStorkGetters {
    /// @notice Retrieves the period (in seconds) that a price feed is considered valid since its publish time
    /// @return uint The number of seconds that data is considered valid
    function validTimePeriodSeconds() external view returns (uint);

    /// @notice Retrieves the single update fee in wei
    /// @return uint The single update fee in wei
    function singleUpdateFeeInWei() external view returns (uint);

    /// @notice Retrieves the Stork public key
    /// @return address The Stork public key
    function storkPublicKey() external view returns (address);
}
