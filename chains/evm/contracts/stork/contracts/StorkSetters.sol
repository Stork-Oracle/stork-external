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

    function storeAddSigningAddress(address signingAddress) internal {
        require(signingAddress != address(0), "Signing address cannot be 0 address");
        require(!_state.signingAddresses[signingAddress], "Signing address already exists");
        _state.signingAddresses[signingAddress] = true;
        _state.signingAddressList.push(signingAddress);
        emit SigningAddressAdded(signingAddress);
    }

    function storeRemoveSigningAddress(address signingAddress) internal {
        require(signingAddress != address(0), "Signing address cannot be 0 address");
        require(_state.signingAddresses[signingAddress], "Signing address does not exist");
        _state.signingAddresses[signingAddress] = false;
        for (uint i = 0; i < _state.signingAddressList.length; i++) {
            if (_state.signingAddressList[i] == signingAddress) {
                _state.signingAddressList[i] = _state.signingAddressList[_state.signingAddressList.length - 1];
                _state.signingAddressList.pop();
                break;
            }
        }
        emit SigningAddressRemoved(signingAddress);
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
