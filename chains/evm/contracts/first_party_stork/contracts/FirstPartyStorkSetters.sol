// contracts/stork/StorkSetters.sol
// SPDX-License-Identifier: Apache 2

pragma solidity >=0.8.24 <0.9.0;

import "./FirstPartyStorkEvents.sol";
import "./FirstPartyStorkStructs.sol";
import "./FirstPartyStorkState.sol";

contract FirstPartyStorkSetters is FirstPartyStorkState, IFirstPartyStorkEvents {
    function updateLatestValueIfNecessary(
        address pubKey,
        FirstPartyStorkStructs.PublisherTemporalNumericValueInput memory input
    ) internal returns (bool) {
        uint64 latestReceiveTime = _state
        .latestValues[pubKey][input.assetPairId].timestampNs;
        if (input.temporalNumericValue.timestampNs < latestReceiveTime) {
            return false;
        }

        _state.latestValues[pubKey][input.assetPairId] = input.temporalNumericValue;
        emit ValueUpdate(
            pubKey,
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
        uint64 latestReceiveTime = _state
        .latestValues[pubKey][input.assetPairId].timestampNs;
        if (input.temporalNumericValue.timestampNs < latestReceiveTime) {
            return false;
        }

        _state.historicalValues[pubKey][input.assetPairId].push(
            FirstPartyStorkStructs.TemporalNumericValue(
                input.temporalNumericValue.timestampNs,
                input.temporalNumericValue.quantizedValue
            )
        );
        _state.currentRoundId[pubKey][input.assetPairId]++;
        emit HistoricalValueStored(
            pubKey,
            input.assetPairId,
            input.temporalNumericValue.timestampNs,
            input.temporalNumericValue.quantizedValue,
            _state.currentRoundId[pubKey][input.assetPairId]
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
