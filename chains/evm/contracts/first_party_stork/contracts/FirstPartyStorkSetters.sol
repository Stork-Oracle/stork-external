// contracts/stork/StorkSetters.sol
// SPDX-License-Identifier: Apache 2

pragma solidity >=0.8.24 <0.9.0;

import "@storknetwork/first-party-stork-evm-sdk/FirstPartyStorkStructs.sol";
import "@storknetwork/first-party-stork-evm-sdk/IFirstPartyStorkEvents.sol";
import "./FirstPartyStorkState.sol";
import "./FirstPartyStorkGetters.sol";

contract FirstPartyStorkSetters is FirstPartyStorkState, FirstPartyStorkGetters, IFirstPartyStorkEvents {
    function updateLatestValueIfNecessary(
        address pubKey,
        FirstPartyStorkStructs.PublisherTemporalNumericValueInput memory input
    ) internal returns (bool) {
        bytes32 encodedAssetId = getEncodedAssetId(input.assetPairId);
        uint64 latestReceiveTime = _state
        .latestValues[pubKey][encodedAssetId].timestampNs;
        if (input.temporalNumericValue.timestampNs < latestReceiveTime) {
            return false;
        }

        _state.latestValues[pubKey][encodedAssetId] = input.temporalNumericValue;
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
        uint64 latestReceiveTime = _state
        .latestValues[pubKey][encodedAssetId].timestampNs;
        if (input.temporalNumericValue.timestampNs < latestReceiveTime) {
            return false;
        }

        _state.historicalValues[pubKey][encodedAssetId].push(
            FirstPartyStorkStructs.TemporalNumericValue(
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
