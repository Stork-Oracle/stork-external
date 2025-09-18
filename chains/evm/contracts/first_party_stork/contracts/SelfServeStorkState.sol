// contracts/stork/StorkState.sol
// SPDX-License-Identifier: Apache 2

pragma solidity >=0.8.24 <0.9.0;

import "./SelfServeStorkStructs.sol";

contract SelfServeStorkStorage {
    struct State {
        mapping(address => mapping(string => SelfServeStorkStructs.TemporalNumericValue)) latestValues;
        mapping(address => mapping(string => SelfServeStorkStructs.TemporalNumericValue[])) historicalValues;
        mapping(address => mapping(string => uint256)) currentRoundId;
        mapping(address => SelfServeStorkStructs.PublisherUser) publisherUsers;
    }
}

contract SelfServeStorkState {
    SelfServeStorkStorage.State _state;
}
