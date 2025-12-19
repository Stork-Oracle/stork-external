// contracts/stork/StorkSetters.sol
// SPDX-License-Identifier: Apache 2

pragma solidity >=0.8.24 <0.9.0;

import "@storknetwork/first-party-stork-evm-sdk/FirstPartyStorkStructs.sol";
import "@storknetwork/first-party-stork-evm-sdk/IFirstPartyStorkEvents.sol";
import "./FirstPartyStorkState.sol";
import "./FirstPartyStorkHelpers.sol";

contract FirstPartyStorkSetters is
    FirstPartyStorkState,
    FirstPartyStorkHelpers,
    IFirstPartyStorkEvents
{
    function updateLatestValueIfNecessary(
        address pubKey,
        FirstPartyStorkStructs.PublisherTemporalNumericValueInput memory input
    ) internal returns (bool) {
        bytes32 encodedAssetId = getEncodedAssetId(input.assetPairId);
        if (
            !isUpdateNecessary(
                pubKey,
                encodedAssetId,
                input.temporalNumericValue.timestampNs
            )
        ) {
            return false;
        }

        _state.latestValues[pubKey][encodedAssetId] = input
            .temporalNumericValue;
        emit ValueUpdate(
            pubKey,
            input.assetPairId,
            input.assetPairId,
            input.temporalNumericValue.timestampNs,
            input.temporalNumericValue.quantizedValue
        );
        return true;
    }

    function storeHistoricalValue(
        address pubKey,
        FirstPartyStorkStructs.PublisherTemporalNumericValueInput memory input
    ) internal returns (bool) {
        bytes32 encodedAssetId = getEncodedAssetId(input.assetPairId);
        if (
            !isUpdateNecessary(
                pubKey,
                encodedAssetId,
                input.temporalNumericValue.timestampNs
            )
        ) {
            return false;
        }

        _state.historicalValues[pubKey][encodedAssetId].push(
            StorkStructs.TemporalNumericValue(
                input.temporalNumericValue.timestampNs,
                input.temporalNumericValue.quantizedValue
            )
        );
        _state.currentRoundId[pubKey][encodedAssetId]++;
        emit HistoricalValueStored(
            pubKey,
            input.assetPairId,
            input.assetPairId,
            input.temporalNumericValue.timestampNs,
            input.temporalNumericValue.quantizedValue,
            _state.currentRoundId[pubKey][encodedAssetId]
        );
        return true;
    }

    function isUpdateNecessary(
        address pubKey,
        bytes32 encodedAssetId,
        uint64 timestampNs
    ) internal view returns (bool) {
        uint64 latestReceiveTime = _state
            .latestValues[pubKey][encodedAssetId]
            .timestampNs;
        return timestampNs > latestReceiveTime;
    }

    function addPublisherUser(
        address pubKey,
        uint256 singleUpdateFee
    ) internal {
        _state.publisherUsers[pubKey] = FirstPartyStorkStructs.PublisherUser(
            pubKey,
            singleUpdateFee
        );
        emit PublisherUserAdded(pubKey, singleUpdateFee);
    }

    function removePublisherUser(address pubKey) internal {
        delete _state.publisherUsers[pubKey];
        emit PublisherUserRemoved(pubKey);
    }
}
