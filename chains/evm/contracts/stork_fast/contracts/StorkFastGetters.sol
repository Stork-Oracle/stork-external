// SPDX-License-Identifier: Apache 2
pragma solidity ^0.8.28;

import "./StorkFastState.sol";

contract StorkFastGetters is StorkFastState {
    function verificationFeeInWei() public view returns (uint) {
        return _state.verificationFeeInWei;
    }

    function storkFastAddress() public view returns (address) {
        return _state.storkFastAddress;
    }
}
