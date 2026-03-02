// SPDX-License-Identifier: Apache-2.0
pragma solidity >=0.8.24 <0.9.0;

import "./StorkStructs.sol";
import "./IStorkEvents.sol";

/// @title IStork
/// @notice Interface for the Stork oracle contract
/// @dev This interface provides access to Stork's temporal numeric values and verification functions
interface IStork is IStorkEvents {
    /// @notice Updates multiple temporal numeric values by verifying signatures and ensuring freshness
    /// @param updateData Array of TemporalNumericValueInput structs containing feed updates
    /// @dev Requires sufficient fee based on the number of updates
    /// @dev Reverts with InvalidSignature if any feed update fails signature verification
    /// @dev Reverts with NoFreshUpdate if none of the provided updates are fresher than current values
    /// @dev Reverts with InsufficientFee if the provided fee is less than the required amount
    function updateTemporalNumericValuesV1(
        StorkStructs.TemporalNumericValueInput[] calldata updateData
    ) external payable;

    /// @notice Retrieves the latest temporal numeric value for the specified feed ID
    /// @param id The identifier of the feed
    /// @return value The latest TemporalNumericValue struct for the feed
    /// @dev Checks for staleness threshold (typically 3600 seconds)
    /// @dev Reverts with NotFound if no value exists for the given feed ID
    /// @dev Reverts with StaleValue if the value is older than the valid time period
    function getTemporalNumericValueV1(
        bytes32 id
    ) external view returns (StorkStructs.TemporalNumericValue memory value);

    /// @notice Retrieves the latest temporal numeric value for the specified feed ID without checking freshness
    /// @param id The identifier of the feed
    /// @return value The latest TemporalNumericValue struct for the feed
    /// @dev Does not check for staleness - use with caution
    /// @dev Reverts with NotFound if no value exists for the given feed ID
    function getTemporalNumericValueUnsafeV1(
        bytes32 id
    ) external view returns (StorkStructs.TemporalNumericValue memory value);

    /// @notice Retrieves the latest temporal numeric values for the specified feed IDs without checking freshness
    /// @param ids The identifiers of the feeds
    /// @return values The latest TemporalNumericValue structs for the feeds
    /// @dev Does not check for staleness - use with caution
    /// @dev Reverts with NotFound if no value exists for the given feed ID
    function getTemporalNumericValuesUnsafeV1(
        bytes32[] calldata ids
    ) external view returns (StorkStructs.TemporalNumericValue[] memory values);

    /// @notice Calculates the total fee required for the given updates
    /// @param updateData Array of TemporalNumericValueInput structs representing updates
    /// @return feeAmount The total fee required for the updates
    function getUpdateFeeV1(
        StorkStructs.TemporalNumericValueInput[] calldata updateData
    ) external view returns (uint feeAmount);

    /// @notice Verifies multiple publisher signatures against the provided Merkle root
    /// @param signatures Array of PublisherSignature structs
    /// @param merkleRoot The Merkle root to validate against
    /// @return bool True if all signatures are valid and match the Merkle root
    function verifyPublisherSignaturesV1(
        StorkStructs.PublisherSignature[] calldata signatures,
        bytes32 merkleRoot
    ) external pure returns (bool);

    /// @notice Retrieves the current version of the contract
    /// @return string The version string (e.g., "1.0.3")
    function version() external pure returns (string memory);
}
