// contracts/Setters.sol
// SPDX-License-Identifier: Apache 2

pragma solidity >=0.8.24 <0.9.0;

import "./StorkState.sol";
import "./IStorkEvents.sol";

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
        _state.storkPublicKey = storkPublicKey;
    }

    function setSingleUpdateFeeInWei(uint fee) internal {
        _state.singleUpdateFeeInWei = fee;
    }

    function setValidTimePeriodSeconds(uint validTimePeriodSeconds) internal {
        _state.validTimePeriodSeconds = validTimePeriodSeconds;
    }
}
