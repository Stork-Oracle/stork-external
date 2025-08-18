// contracts/stork/StorkSetters.sol
// SPDX-License-Identifier: Apache 2

pragma solidity >=0.8.24 <0.9.0;

import "@storknetwork/stork-evm-sdk/IStorkEvents.sol";
import "@storknetwork/stork-evm-sdk/StorkStructs.sol";
import "./StorkState.sol";

contract StorkSetters is StorkState, IStorkEvents {
    function updateLatestValueIfNecessary(
        StorkStructs.TemporalNumericValueInput memory input
    ) internal returns (bool) {
        uint64 latestReceiveTime = _state.latestCanonicalTemporalNumericValues[input.id].timestampNs;
        if (input.temporalNumericValue.timestampNs > latestReceiveTime) {
            _state.latestCanonicalTemporalNumericValues[input.id] = input.temporalNumericValue;
            emit ValueUpdate(
                input.id,
                input.temporalNumericValue.timestampNs,
                input.temporalNumericValue.quantizedValue
            );
            return true;
        }
        return false;
    }

    function setStorkPublicKey(address storkPublicKey) internal {
        require(storkPublicKey != address(0), "Stork public key cannot be 0 address");
        _state.storkPublicKey = storkPublicKey;
        emit StorkPublicKeyUpdate(storkPublicKey);
    }

    function setSingleUpdateFeeInWei(uint fee) internal {
        _state.singleUpdateFeeInWei = fee;
        emit SingleUpdateFeeUpdate(fee);
    }

    function setValidTimePeriodSeconds(uint validTimePeriodSeconds) internal {
        _state.validTimePeriodSeconds = validTimePeriodSeconds;
        emit ValidTimePeriodUpdate(validTimePeriodSeconds);
    }
}
