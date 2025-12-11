// SPDX-License-Identifier: Apache 2
pragma solidity ^0.8.28;

contract StorkFastStorage {
    struct State {
        // Address derived from the ECDSA keypair used to sign signed ECDSA payloads
        address signerAddress;
        // Fee in wei charged for a single signature verification
        uint256 verificationFeeInWei;
    }
}

contract StorkFastState {
    StorkFastStorage.State _state;
}
