// contracts/stork/StorkGetters.sol
// SPDX-License-Identifier: Apache 2

pragma solidity >=0.8.24 <0.9.0;

import "@storknetwork/first-party-stork-evm-sdk/FirstPartyStorkStructs.sol";
import "@storknetwork/first-party-stork-evm-sdk/IFirstPartyStorkGetters.sol";
import "@storknetwork/first-party-stork-evm-sdk/FirstPartyStorkErrors.sol";
import "./FirstPartyStorkState.sol";
import "./FirstPartyStorkHelpers.sol";

contract FirstPartyStorkGetters is
    FirstPartyStorkState,
    FirstPartyStorkHelpers,
    IFirstPartyStorkGetters
{
    function getLatestTemporalNumericValue(
        address pubKey,
        string memory assetPairId
    ) public view returns (StorkStructs.TemporalNumericValue memory value) {
        bytes32 encodedAssetId = getEncodedAssetId(assetPairId);
        if (_state.latestValues[pubKey][encodedAssetId].timestampNs == 0) {
            revert FirstPartyStorkErrors.NotFound();
        }

        return _state.latestValues[pubKey][encodedAssetId];
    }

    function getHistoricalTemporalNumericValue(
        address pubKey,
        string memory assetPairId,
        uint256 roundId
    ) public view returns (StorkStructs.TemporalNumericValue memory) {
        bytes32 encodedAssetId = getEncodedAssetId(assetPairId);
        if (roundId >= _state.historicalValues[pubKey][encodedAssetId].length) {
            revert FirstPartyStorkErrors.NotFound();
        }

        return _state.historicalValues[pubKey][encodedAssetId][roundId];
    }

    function getHistoricalRecordsCount(
        address pubKey,
        string memory assetPairId
    ) public view returns (uint256) {
        bytes32 encodedAssetId = getEncodedAssetId(assetPairId);
        return _state.historicalValues[pubKey][encodedAssetId].length;
    }

    function getCurrentRoundId(
        address pubKey,
        string memory assetPairId
    ) public view returns (uint256) {
        bytes32 encodedAssetId = getEncodedAssetId(assetPairId);
        return _state.currentRoundId[pubKey][encodedAssetId];
    }

    function getPublisherUser(
        address pubKey
    ) public view returns (FirstPartyStorkStructs.PublisherUser memory) {
        if (_state.publisherUsers[pubKey].pubKey == address(0)) {
            revert FirstPartyStorkErrors.NotFound();
        }

        return _state.publisherUsers[pubKey];
    }

    function getSingleUpdateFee(address pubKey) public view returns (uint) {
        return getPublisherUser(pubKey).singleUpdateFee;
    }
}
