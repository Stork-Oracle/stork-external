// contracts/stork/StorkGetters.sol
// SPDX-License-Identifier: Apache 2

pragma solidity >=0.8.24 <0.9.0;

import "@storknetwork/first-party-stork-evm-sdk/FirstPartyStorkStructs.sol";
import "@storknetwork/first-party-stork-evm-sdk/IFirstPartyStorkGetters.sol";
import "@storknetwork/first-party-stork-evm-sdk/FirstPartyStorkErrors.sol";
import "./FirstPartyStorkState.sol";

contract FirstPartyStorkGetters is FirstPartyStorkState, IFirstPartyStorkGetters {
    function getLatestTemporalNumericValue(
        address pubKey,
        string memory assetPairId
    ) public view returns (FirstPartyStorkStructs.TemporalNumericValue memory value) {
        if (_state.latestValues[pubKey][assetPairId].timestampNs == 0) {
            revert FirstPartyStorkErrors.NotFound();
        }

        return _state.latestValues[pubKey][assetPairId];
    }

    function getHistoricalTemporalNumericValue(
        address pubKey,
        string memory assetPairId,
        uint256 roundId
    ) public view returns (FirstPartyStorkStructs.TemporalNumericValue memory) {
        if (roundId >= _state.historicalValues[pubKey][assetPairId].length) {
            revert FirstPartyStorkErrors.NotFound();
        }

        return _state.historicalValues[pubKey][assetPairId][roundId];
    }

    function getHistoricalRecordsCount(
        address pubKey,
        string memory assetPairId
    ) public view returns (uint256) {
        return _state.historicalValues[pubKey][assetPairId].length;
    }

    function getCurrentRoundId(
        address pubKey,
        string memory assetPairId
    ) public view returns (uint256) {
        return _state.currentRoundId[pubKey][assetPairId];
    }

    function getPublisherUser(
        address pubKey
    ) public view returns (FirstPartyStorkStructs.PublisherUser memory) {
        if (_state.publisherUsers[pubKey].pubKey == address(0)) {
            revert FirstPartyStorkErrors.NotFound();
        }

        return _state.publisherUsers[pubKey];
    }
}
