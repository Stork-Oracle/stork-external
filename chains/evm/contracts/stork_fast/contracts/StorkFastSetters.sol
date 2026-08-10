// SPDX-License-Identifier: Apache 2
pragma solidity ^0.8.28;

import "./StorkFastState.sol";
import "@storknetwork/stork-fast-evm-sdk/IStorkFastEvents.sol";

contract StorkFastSetters is StorkFastState, IStorkFastEvents {
    uint256 private constant MAX_SIGNER_ADDRESSES = 8;

    function storeAddSignerAddress(address signerAddress) internal {
        require(
            signerAddress != address(0),
            "Signer address cannot be 0 address"
        );
        require(
            !_state.signerAddresses[signerAddress],
            "Signer address already exists"
        );
        require(
            _state.signerAddressList.length < MAX_SIGNER_ADDRESSES,
            "Signer address limit reached"
        );
        _state.signerAddresses[signerAddress] = true;
        _state.signerAddressList.push(signerAddress);
        emit SignerAddressAdded(signerAddress);
    }

    function storeRemoveSignerAddress(address signerAddress) internal {
        require(
            _state.signerAddresses[signerAddress],
            "Signer address does not exist"
        );
        require(
            _state.signerAddressList.length > 1,
            "Cannot remove last signer address"
        );
        _state.signerAddresses[signerAddress] = false;
        for (uint i = 0; i < _state.signerAddressList.length; i++) {
            if (_state.signerAddressList[i] == signerAddress) {
                _state.signerAddressList[i] = _state.signerAddressList[
                    _state.signerAddressList.length - 1
                ];
                _state.signerAddressList.pop();
                break;
            }
        }
        emit SignerAddressRemoved(signerAddress);
    }

    function setVerificationFeeInWei(uint verificationFeeInWei) internal {
        _state.verificationFeeInWei = verificationFeeInWei;
        emit VerificationFeeUpdate(verificationFeeInWei);
    }
}
