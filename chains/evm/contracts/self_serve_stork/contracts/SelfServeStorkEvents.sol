// SPDX-License-Identifier: Apache-2.0
pragma solidity >=0.8.24 <0.9.0;

/// @title ISelfServeStorkEvents
/// @notice Events emitted by the SelfServeStork contract
/// @dev This interface can be used for listening to updates for off-chain and testing purposes
interface ISelfServeStorkEvents {
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

    event HistoricalValueStored(
        address indexed pubKey,
        string indexed assetId,
        uint64 timestampNs,
        int192 quantizedValue,
        uint256 roundId
    );

    event PublisherUserAdded(address indexed pubKey, uint256 singleUpdateFee);

    event PublisherUserRemoved(address indexed pubKey);
}
