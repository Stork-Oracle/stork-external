// contracts/stork/StorkState.sol
// SPDX-License-Identifier: Apache 2

pragma solidity >=0.8.24 <0.9.0;

import "@storknetwork/first-party-stork-evm-sdk/FirstPartyStorkStructs.sol";

contract FirstPartyStorkStorage {
    struct State {
        // Mapping of publisher to encodedAssetId to TemporalNumericValue
        mapping(address => mapping(bytes32 => FirstPartyStorkStructs.TemporalNumericValue)) latestValues;
        // Mapping of publisher to encodedAssetId to array of TemporalNumericValue
        mapping(address => mapping(bytes32 => FirstPartyStorkStructs.TemporalNumericValue[])) historicalValues;
        // Mapping of publisher to encodedAssetId to current roundId corresponding to the historical values
        mapping(address => mapping(bytes32 => uint256)) currentRoundId;
        // Mapping of publisher to PublisherUser
        mapping(address => FirstPartyStorkStructs.PublisherUser) publisherUsers;
    }
}

contract FirstPartyStorkState {
    FirstPartyStorkStorage.State _state;
}
