// contracts/State.sol
// SPDX-License-Identifier: Apache 2

pragma solidity >=0.8.24 <0.9.0;

import "./StorkStructs.sol";

contract StorkStorage {
    struct State {
        // For verifying the authenticity of the passed data
        address storkPublicKey;
        uint singleUpdateFeeInWei;
        /// Maximum acceptable time period before value is considered to be stale.
        /// This includes attestation delay, block time, and potential clock drift
        /// between the source/target chains.
        uint validTimePeriodSeconds;
        // Mapping of cached numeric temporal data
        mapping(bytes32 => StorkStructs.TemporalNumericValue) latestCanonicalTemporalNumericValues;
    }
}

contract StorkState {
    StorkStorage.State _state;
}
