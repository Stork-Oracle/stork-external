// contracts/stork/StorkGetters.sol
// SPDX-License-Identifier: Apache 2

pragma solidity >=0.8.24 <0.9.0;

import "@storknetwork/stork-evm-sdk/StorkStructs.sol";
import "@storknetwork/stork-evm-sdk/IStorkGetters.sol";

import "./StorkState.sol";

contract StorkGetters is StorkState, IStorkGetters {
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

    function isSigningAddress(address addr) public view returns (bool) {
        if (addr == address(0)) return false;
        return _state.signingAddresses[addr] || addr == storkPublicKey();
    }
}
