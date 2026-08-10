// SPDX-License-Identifier: Apache 2
pragma solidity ^0.8.28;

contract StorkFastStorage {
    struct State {
        // Set of valid signer addresses for key rotation with backwards compatibility.
        // Addresses are derived from the ECDSA keypairs used to sign signed ECDSA payloads.
        // signerAddresses provides O(1) membership check during signature verification;
        // signerAddressList enables enumeration via getSignerAddresses()
        mapping(address => bool) signerAddresses;
        address[] signerAddressList;
        // Fee in wei charged for a single signature verification
        uint256 verificationFeeInWei;
    }
}

contract StorkFastState {
    StorkFastStorage.State _state;
}
