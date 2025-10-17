// SPDX-License-Identifier: Apache-2.0
pragma solidity >=0.8.24 <0.9.0;

/// @title IFirstPartyStorkHelpers
/// @notice Helper functions for the First Party Stork contract
interface IFirstPartyStorkHelpers {
    /// @notice Calculates the encoded asset ID for the specified asset pair
    /// @param assetPairId The asset pair identifier
    /// @return encodedAssetId The keccak256 hash of the asset pair identifier
    function getEncodedAssetId(
        string memory assetPairId
    ) external pure returns (bytes32);
}
