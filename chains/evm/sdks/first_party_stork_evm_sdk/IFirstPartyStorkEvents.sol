// SPDX-License-Identifier: Apache-2.0
pragma solidity >=0.8.24 <0.9.0;

/// @title IFirstPartyStorkEvents
/// @notice Events emitted by the First Party Stork contract
/// @dev This interface can be used for listening to updates for off-chain and testing purposes
interface IFirstPartyStorkEvents {
    /// @notice Emitted when the latest value with `assetId` has received a fresh update
    /// @param pubKey The publisher's public key
    /// @param assetId The asset ID
    /// @param timestampNs Publish time of the given update
    /// @param quantizedValue Value of the given update
    event ValueUpdate(
        address indexed pubKey,
        string indexed assetId,
        uint64 timestampNs,
        int192 quantizedValue
    );

    /// @notice Emitted when a historical value is stored
    /// @param pubKey The publisher's public key
    /// @param assetId The asset ID
    /// @param timestampNs Publish time of the given update
    /// @param quantizedValue Value of the given update
    /// @param roundId The round ID of the historical record
    event HistoricalValueStored(
        address indexed pubKey,
        string indexed assetId,
        uint64 timestampNs,
        int192 quantizedValue,
        uint256 roundId
    );

    /// @notice Emitted when a publisher user is added
    /// @param pubKey The publisher's public key
    /// @param singleUpdateFee The fee for a single update
    event PublisherUserAdded(address indexed pubKey, uint256 singleUpdateFee);

    /// @notice Emitted when a publisher user is removed
    /// @param pubKey The publisher's public key
    event PublisherUserRemoved(address indexed pubKey);
}
