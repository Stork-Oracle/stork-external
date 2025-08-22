// contracts/stork/StorkGetters.sol
// SPDX-License-Identifier: Apache 2

pragma solidity >=0.8.24 <0.9.0;

import "./SelfServeStorkStructs.sol";
import "./SelfServeStorkState.sol";
import "./SelfServeStorkErrors.sol";

contract SelfServeStorkGetters is SelfServeStorkState {
    function getLatestTemporalNumericValue(
        address pubKey,
        bytes32 assetId
    ) public view returns (SelfServeStorkStructs.TemporalNumericValue memory value) {
        if (_state.latestValues[pubKey][assetId].timestampNs == 0) {
            revert SelfServeStorkErrors.NotFound();
        }

        return _state.latestValues[pubKey][assetId];
    }

    function getHistoricalTemporalNumericValue(
        address pubKey,
        bytes32 assetId,
        uint256 roundId
    ) public view returns (SelfServeStorkStructs.TemporalNumericValue memory) {
        if (roundId >= _state.historicalValues[pubKey][assetId].length) {
            revert SelfServeStorkErrors.NotFound();
        }

        return _state.historicalValues[pubKey][assetId][roundId];
    }

    function getHistoricalRecordsCount(
        address pubKey,
        bytes32 assetId
    ) public view returns (uint256) {
        return _state.historicalValues[pubKey][assetId].length;
    }

    function getCurrentRoundId(
        address pubKey,
        bytes32 assetId
    ) public view returns (uint256) {
        return _state.currentRoundId[pubKey][assetId];
    }

    function getPublisherUser(
        address pubKey
    ) public view returns (SelfServeStorkStructs.PublisherUser memory) {
        if (_state.publisherUsers[pubKey].pubKey == address(0)) {
            revert SelfServeStorkErrors.NotFound();
        }

        return _state.publisherUsers[pubKey];
    }
}
