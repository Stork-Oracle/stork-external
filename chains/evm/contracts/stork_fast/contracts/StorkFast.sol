// SPDX-License-Identifier: Apache 2
pragma solidity ^0.8.28;

import "@openzeppelin/contracts/utils/cryptography/ECDSA.sol";
import "./StorkFastGetters.sol";
import "./StorkFastSetters.sol";
import "@storknetwork/stork-fast-evm-sdk/StorkFastStructs.sol";
import "@storknetwork/stork-fast-evm-sdk/StorkFastDeserialize.sol";
import "@storknetwork/stork-fast-evm-sdk/StorkFastErrors.sol";

abstract contract StorkFast is StorkFastGetters, StorkFastSetters {
    function _initialize(
        address storkFastAddress,
        uint verificationFeeInWei
    ) internal {
        StorkFastSetters.setStorkFastAddress(storkFastAddress);
        StorkFastSetters.setVerificationFeeInWei(verificationFeeInWei);
    }

    function verifySignedECDSAPayload(
        bytes calldata payload
    ) public payable returns (bool) {
        if (msg.value < verificationFeeInWei()) {
            revert StorkFastErrors.InsufficientFee();
        }
        (
            bytes memory signature,
            bytes memory verifiablePayload
        ) = StorkFastDeserialize.splitSignedECDSAPayload(payload);
        bytes32 messageHash = keccak256(verifiablePayload);

        (address signer, , ) = ECDSA.tryRecover(messageHash, signature);

        return signer == storkFastAddress();
    }

    function verifyAndDeserializeSignedECDSAPayload(
        bytes calldata payload
    ) public payable returns (StorkFastStructs.Update[] memory updates) {
        bool verified = verifySignedECDSAPayload{value: msg.value}(payload);
        if (!verified) revert StorkFastErrors.InvalidSignature();
        updates = StorkFastDeserialize.deserializeValuesFromSignedECDSAPayload(
            payload
        );
    }

    function updateVerificationFeeInWei(
        uint verificationFeeInWei
    ) public virtual;

    function updateStorkFastAddress(address storkFastAddress) public virtual;

    function version() public pure returns (string memory) {
        return "1.0.0";
    }
}
