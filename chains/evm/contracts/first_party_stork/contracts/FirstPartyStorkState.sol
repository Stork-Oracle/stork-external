// contracts/stork/StorkState.sol
// SPDX-License-Identifier: Apache 2

pragma solidity >=0.8.24 <0.9.0;

import "./FirstPartyStorkStructs.sol";

contract FirstPartyStorkStorage {
    struct State {
        mapping(address => mapping(string => FirstPartyStorkStructs.TemporalNumericValue)) latestValues;
        mapping(address => mapping(string => FirstPartyStorkStructs.TemporalNumericValue[])) historicalValues;
        mapping(address => mapping(string => uint256)) currentRoundId;
        mapping(address => FirstPartyStorkStructs.PublisherUser) publisherUsers;
    }
}

contract FirstPartyStorkState {
    FirstPartyStorkStorage.State _state;
}
