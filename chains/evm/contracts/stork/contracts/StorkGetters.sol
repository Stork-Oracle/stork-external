// contracts/Getters.sol
// SPDX-License-Identifier: Apache 2

pragma solidity >=0.8.24 <0.9.0;

import "./StorkStructs.sol";
import "./StorkState.sol";

contract StorkGetters is StorkState {
    function latestCanonicalTemporalNumericValue(
        bytes32 id
    ) internal view returns (StorkStructs.TemporalNumericValue memory value) {
        return _state.latestCanonicalTemporalNumericValues[id];
    }

    function singleUpdateFeeInWei() public view returns (uint) {
        return _state.singleUpdateFeeInWei;
    }

    function validTimePeriodSeconds() public view returns (uint) {
        return _state.validTimePeriodSeconds;
    }

    function storkPublicKey() public view returns (address) {
        return _state.storkPublicKey;
    }
}
