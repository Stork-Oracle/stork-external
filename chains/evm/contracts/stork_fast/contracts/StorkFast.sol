// SPDX-License-Identifier: Apache 2
pragma solidity ^0.8.28;

import "@openzeppelin/contracts/utils/cryptography/ECDSA.sol";
import "./StorkFastGetters.sol";
import "./StorkFastSetters.sol";
import "@storknetwork/stork-fast-evm-sdk/StorkFastStructs.sol";
import "@storknetwork/stork-fast-evm-sdk/StorkFastDeserialize.sol";
import "@storknetwork/stork-fast-evm-sdk/StorkFastErrors.sol";
import "@storknetwork/stork-fast-evm-sdk/IStorkFast.sol";

abstract contract StorkFast is StorkFastGetters, StorkFastSetters, IStorkFast {
    function _initialize(
        address[] memory signerAddresses,
        uint verificationFeeInWei
    ) internal {
        require(
            signerAddresses.length > 0,
            "At least one signer address is required"
        );
        for (uint i = 0; i < signerAddresses.length; i++) {
            storeAddSignerAddress(signerAddresses[i]);
        }
        setVerificationFeeInWei(verificationFeeInWei);
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
        signature[64] = bytes1(uint8(signature[64]) + 27);

        (address signer, , ) = ECDSA.tryRecover(messageHash, signature);

        // tryRecover returns address(0) on failure, which is never a valid signer
        return isValidSignerAddress(signer);
    }

    function verifyAndDeserializeSignedECDSAPayload(
        bytes calldata payload
    ) public payable returns (StorkFastStructs.Asset[] memory assets) {
        bool verified = verifySignedECDSAPayload(payload);
        if (!verified) revert StorkFastErrors.InvalidSignature();
        assets = StorkFastDeserialize.deserializeAssetsFromSignedECDSAPayload(
            payload
        );
    }

    function updateVerificationFeeInWei(
        uint verificationFeeInWei
    ) public virtual;

    function addSignerAddress(address signerAddress) public virtual;

    function removeSignerAddress(address signerAddress) public virtual;

    function version() public pure returns (string memory) {
        return "1.0.0";
    }
}
