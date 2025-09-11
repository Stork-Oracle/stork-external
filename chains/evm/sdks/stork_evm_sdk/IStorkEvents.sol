// SPDX-License-Identifier: Apache-2.0
pragma solidity >=0.8.24 <0.9.0;

/// @title IStorkEvents
/// @notice Events emitted by the Stork contract
/// @dev This interface can be used for listening to updates for off-chain and testing purposes
interface IStorkEvents {
    /// @notice Emitted when the latest value with `id` has received a fresh update
    /// @param id The Stork Feed ID
    /// @param timestampNs Publish time of the given update
    /// @param quantizedValue Value of the given update
    event ValueUpdate(
        bytes32 indexed id,
        uint64 timestampNs,
        int192 quantizedValue
    );

    /// @notice Emitted when the Stork public key is updated
    /// @param newStorkPublicKey The new Stork public key
    event StorkPublicKeyUpdate(address indexed newStorkPublicKey);

    /// @notice Emitted when the Stork single update fee is updated
    /// @param newSingleUpdateFee The new Stork single update fee
    event SingleUpdateFeeUpdate(uint256 newSingleUpdateFee);

    /// @notice Emitted when the Stork valid time period is updated
    /// @param newValidTimePeriod The new Stork valid time period
    event ValidTimePeriodUpdate(uint256 newValidTimePeriod);
}
