// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.28;

import "@storknetwork/stork-fast-evm-sdk/IStorkFast.sol";
import "@storknetwork/stork-fast-evm-sdk/IStorkFastGetters.sol";
import "@storknetwork/stork-fast-evm-sdk/StorkFastStructs.sol";

/// @title Example Stork Fast consumer contract
/// @notice This contract is just an example of interacting with the Stork Fast contract to verify and deserialize signed ECDSA payloads
/// @dev This contract is not audited and is for demonstration purposes only
contract Example {
    /// @dev The Stork Fast contract interface
    IStorkFast public storkFast;

    /// @notice Dummy event to demonstrate using a Stork Fast asset after verification and deserialization
    /// @param assetID The ID of the asset
    /// @param quantizedValue The quantized value of the asset
    /// @param timestampNs The timestamp in nanoseconds
    event StorkFastAssetVerified(
        uint16 indexed assetID,
        int192 quantizedValue,
        uint64 timestampNs
    );

    /// @notice Constructor
    /// @param _storkFast The address of the Stork Fast contract
    constructor(address _storkFast) {
        storkFast = IStorkFast(_storkFast);
    }

    /// @notice Use the Stork Fast contract to verify and deserialize a signed ECDSA payload
    /// @param payload The signed ECDSA payload
    /// @dev Requires a sufficient fee to cover the verification fee
    /// @dev Reverts with InsufficientFee if the provided fee is less than verification fee
    function useStorkFast(bytes calldata payload) public payable {
        // Ensure we have a sufficient fee to cover the verification fee
        if (msg.value < storkFast.verificationFeeInWei()) {
            revert("Insufficient fee provided");
        }

        // Verify and deserialize the signed ECDSA payload
        StorkFastStructs.Asset[] memory assets = storkFast
            .verifyAndDeserializeSignedECDSAPayload{value: msg.value}(payload);

        // Use the assets for something...
        for (uint i = 0; i < assets.length; i++) {
            emit StorkFastAssetVerified(
                assets[i].assetID,
                assets[i].temporalNumericValue.quantizedValue,
                assets[i].temporalNumericValue.timestampNs
            );
        }
    }
}
