// SPDX-License-Identifier: Apache 2
pragma solidity ^0.8.28;

import "./StorkFastState.sol";
import "@storknetwork/stork-fast-evm-sdk/IStorkFastEvents.sol";

contract StorkFastSetters is StorkFastState, IStorkFastEvents {
    function setSignerAddress(address signerAddress) internal {
        require(
            signerAddress != address(0),
            "Signer address cannot be 0 address"
        );
        _state.signerAddress = signerAddress;
        emit SignerAddressUpdate(signerAddress);
    }

    function setVerificationFeeInWei(uint verificationFeeInWei) internal {
        _state.verificationFeeInWei = verificationFeeInWei;
        emit VerificationFeeUpdate(verificationFeeInWei);
    }
}
