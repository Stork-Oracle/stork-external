// SPDX-License-Identifier: Apache-2.0
pragma solidity >=0.8.24 <0.9.0;

import "./FirstPartyStorkStructs.sol";
import "./IFirstPartyStorkEvents.sol";
import "./IFirstPartyStorkGetters.sol";

/// @title IFirstPartyStork
/// @notice Interface for the First Party Stork oracle contract
/// @dev This interface provides access to First Party Stork's temporal numeric values and publisher management functions
interface IFirstPartyStork is IFirstPartyStorkEvents, IFirstPartyStorkGetters {
    /// @notice Updates multiple temporal numeric values by verifying publisher signatures
    /// @param updateData Array of PublisherTemporalNumericValueInput structs containing feed updates
    /// @dev Requires sufficient fee based on the publisher's single update fee
    /// @dev Reverts with InvalidSignature if any feed update fails signature verification
    /// @dev Reverts with NoFreshUpdate if none of the provided updates are fresher than current values
    /// @dev Reverts with InsufficientFee if the provided fee is less than the required amount
    function updateTemporalNumericValues(
        FirstPartyStorkStructs.PublisherTemporalNumericValueInput[] calldata updateData,
    ) external payable;

    /// @notice Verifies a publisher signature for the given parameters
    /// @param publisherPubKey The publisher's public key
    /// @param assetPairId The asset pair identifier
    /// @param timestamp The timestamp of the data (in nanoseconds, but verification uses seconds)
    /// @param value The quantized value
    /// @param r The r component of the signature
    /// @param s The s component of the signature
    /// @param v The v component of the signature
    /// @return bool True if the signature is valid
    function verifyPublisherSignatureV1(
        address publisherPubKey,
        string memory assetPairId,
        uint256 timestamp,
        int256 value,
        bytes32 r,
        bytes32 s,
        uint8 v
    ) external pure returns (bool);

    /// @notice Retrieves the total update fee for a publisher and update
    /// @param updateData Array of PublisherTemporalNumericValueInput structs containing feed updates
    /// @return fee The total update fee
    function getUpdateFeeV1(
        FirstPartyStorkStructs.PublisherTemporalNumericValueInput[]
            calldata updateData
    ) external view returns (uint);
}
