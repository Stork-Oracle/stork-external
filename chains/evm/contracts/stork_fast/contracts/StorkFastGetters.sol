// SPDX-License-Identifier: Apache 2
pragma solidity ^0.8.28;

import "./StorkFastState.sol";
import "@storknetwork/stork-fast-evm-sdk/IStorkFastGetters.sol";

contract StorkFastGetters is StorkFastState, IStorkFastGetters {
    function verificationFeeInWei() public view returns (uint) {
        return _state.verificationFeeInWei;
    }

    function storkFastAddress() public view returns (address) {
        return _state.storkFastAddress;
    }
}
