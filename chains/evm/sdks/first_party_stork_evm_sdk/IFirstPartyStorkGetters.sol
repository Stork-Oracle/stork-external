// SPDX-License-Identifier: Apache-2.0
pragma solidity >=0.8.24 <0.9.0;

import "./FirstPartyStorkStructs.sol";

/// @title IFirstPartyStorkGetters
/// @notice Getter functions for the First Party Stork contract
interface IFirstPartyStorkGetters {
    /// @notice Retrieves the latest temporal numeric value for the specified publisher and asset pair
    /// @param pubKey The publisher's public key
    /// @param assetPairId The asset pair identifier
    /// @return value The latest TemporalNumericValue struct for the publisher and asset pair
    /// @dev Reverts with NotFound if no value exists for the given publisher and asset pair
    function getLatestTemporalNumericValue(
        address pubKey,
        string memory assetPairId
    ) external view returns (FirstPartyStorkStructs.TemporalNumericValue memory);

    /// @notice Retrieves a historical temporal numeric value for the specified publisher, asset pair, and round ID
    /// @param pubKey The publisher's public key
    /// @param assetPairId The asset pair identifier
    /// @param roundId The round ID of the historical record
    /// @return value The TemporalNumericValue struct for the specified round
    /// @dev Reverts with NotFound if the round ID doesn't exist
    function getHistoricalTemporalNumericValue(
        address pubKey,
        string memory assetPairId,
        uint256 roundId
    ) external view returns (FirstPartyStorkStructs.TemporalNumericValue memory);

    /// @notice Retrieves the count of historical records for the specified publisher and asset pair
    /// @param pubKey The publisher's public key
    /// @param assetPairId The asset pair identifier
    /// @return count The number of historical records
    function getHistoricalRecordsCount(
        address pubKey,
        string memory assetPairId
    ) external view returns (uint256);

    /// @notice Retrieves the current round ID for the specified publisher and asset pair
    /// @param pubKey The publisher's public key
    /// @param assetPairId The asset pair identifier
    /// @return roundId The current round ID
    function getCurrentRoundId(
        address pubKey,
        string memory assetPairId
    ) external view returns (uint256);

    /// @notice Retrieves the publisher user configuration
    /// @param pubKey The publisher's public key
    /// @return user The PublisherUser struct containing configuration
    /// @dev Reverts with NotFound if the publisher user doesn't exist
    function getPublisherUser(
        address pubKey
    ) external view returns (FirstPartyStorkStructs.PublisherUser memory);
    
    /// @notice Retrieves the single update fee for a publisher
    /// @param pubKey The publisher's public key
    /// @return fee The single update fee
    function getSingleUpdateFee(
        address pubKey
    ) external view returns (uint);

    /// @notice Calculates the encoded asset ID for the specified asset pair
    /// @param assetPairId The asset pair identifier
    /// @return encodedAssetId The keccak256 hash of the asset pair identifier
    function getEncodedAssetId(
        string memory assetPairId
    ) external pure returns (bytes32);
}
