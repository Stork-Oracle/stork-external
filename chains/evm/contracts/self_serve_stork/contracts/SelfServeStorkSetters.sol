// contracts/stork/StorkSetters.sol
// SPDX-License-Identifier: Apache 2

pragma solidity >=0.8.24 <0.9.0;

import "./SelfServeStorkEvents.sol";
import "./SelfServeStorkStructs.sol";
import "./SelfServeStorkState.sol";

contract SelfServeStorkSetters is SelfServeStorkState, ISelfServeStorkEvents {
    function updateLatestValueIfNecessary(
        address pubKey,
        SelfServeStorkStructs.PublisherTemporalNumericValueInput memory input
    ) internal returns (bool) {
        bytes32 assetId = getAssetId(input.assetPairId);
        uint64 latestReceiveTime = _state
        .latestValues[pubKey][assetId].timestampNs;
        if (input.temporalNumericValue.timestampNs < latestReceiveTime) {
            return false;
        }

        _state.latestValues[pubKey][assetId] = input.temporalNumericValue;
        emit ValueUpdate(
            pubKey,
            assetId,
            input.temporalNumericValue.timestampNs,
            input.temporalNumericValue.quantizedValue
        );
        return true;
    }

    function storeHistoricalValue(
        address pubKey,
        SelfServeStorkStructs.PublisherTemporalNumericValueInput memory input
    ) internal returns (bool) {
        bytes32 assetId = getAssetId(input.assetPairId);
        uint64 latestReceiveTime = _state
        .latestValues[pubKey][assetId].timestampNs;
        if (input.temporalNumericValue.timestampNs < latestReceiveTime) {
            return false;
        }

        _state.historicalValues[pubKey][assetId].push(
            SelfServeStorkStructs.TemporalNumericValue(
                input.temporalNumericValue.timestampNs,
                input.temporalNumericValue.quantizedValue
            )
        );
        _state.currentRoundId[pubKey][assetId]++;
        emit HistoricalValueStored(
            pubKey,
            assetId,
            input.temporalNumericValue.timestampNs,
            input.temporalNumericValue.quantizedValue,
            _state.currentRoundId[pubKey][assetId]
        );
        return true;
    }

    function addPublisherUser(
        address pubKey,
        uint256 singleUpdateFee
    ) internal {
        _state.publisherUsers[pubKey] = SelfServeStorkStructs.PublisherUser(
            pubKey,
            singleUpdateFee
        );
        emit PublisherUserAdded(pubKey, singleUpdateFee);
    }

    function removePublisherUser(address pubKey) internal {
        delete _state.publisherUsers[pubKey];
        emit PublisherUserRemoved(pubKey);
    }

    function getAssetId(string memory assetPairId) internal pure returns (bytes32) {
        return keccak256(abi.encodePacked(assetPairId));
    }
}
