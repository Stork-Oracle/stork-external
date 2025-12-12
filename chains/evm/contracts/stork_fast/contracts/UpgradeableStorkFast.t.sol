// SPDX-License-Identifier: MIT
pragma solidity ^0.8.28;

import "./UpgradeableStorkFast.sol";
import "@storknetwork/stork-fast-evm-sdk/StorkFastStructs.sol";
import "@storknetwork/stork-fast-evm-sdk/StorkFastDeserialize.sol";
import "@storknetwork/stork-fast-evm-sdk/StorkFastErrors.sol";
import "@openzeppelin/contracts/proxy/transparent/TransparentUpgradeableProxy.sol";
import "forge-std/Test.sol";

contract UpgradeableStorkFastTest is Test {
    UpgradeableStorkFast public implementation;
    UpgradeableStorkFast public storkFast;
    TransparentUpgradeableProxy public proxy;

    // ==== TEST ADDRESSES ====

    address public owner = address(0x1);
    address public otherAccount = address(0x2);

    address public signerAddress =
        address(0xC4A02e7D370402F4afC36032076B05e74FF81786);

    // ==== VERIFICATION FEE ====

    uint256 public verificationFee = 1;

    // ==== TEST PAYLOADS ====

    // Test payloads have the asset IDs after _ present
    // Each asset has the unscaled value of asset ID * 100 (negative if specified)

    bytes public validPayloadHex_1 =
        hex"690a1cce3cf72e889b699ef800fb80ba47c123ce7422517bfcd2ce0c701423a12938fb6fcf4424c21e39688cd41dd8919b149642b590328ada7532c88b5f0d6b010001187f4dcc627041f8000100000000000000056bc75e2d63100000";

    bytes public validPayloadHex_1_2_3_4_5_6 =
        hex"5985a517300f6b80a8b6fa7cb2477ef9b57f9cf27706f42e38e096e5350b034730c755a25a416ecd0021ef32a1b36de88f051a5125fb6ae1306dce7c4e9ffa03010001187f4dda764f46c0000100000000000000056bc75e2d631000000002000000000000000ad78ebc5ac62000000003000000000000001043561a882930000000040000000000000015af1d78b58c4000000005000000000000001b1ae4d6e2ef5000000006000000000000002086ac351052600000";

    bytes public validNegativePayloadHex_7 =
        hex"e2c0c8ed493d2a7ced7092010bf670c59c33a54d374416a612ebc4064faac878430e96f68a1d2a7f530b51770340e56edca3d587646ee23aef36a3c09321da0e000001187f506ddec854e00007ffffffffffffffda0d8c6cc24a900000";

    bytes public invalidSignaturePayloadHex =
        hex"1985a517300f6b80a8b6fa7cb2477ef9b57f9cf27706f42e38e096e5350b034730c755a25a416ecd0021ef32a1b36de88f051a5125fb6ae1306dce7c4e9ffa03010001187f4dda764f46c0000100000000000000056bc75e2d631000000002000000000000000ad78ebc5ac62000000003000000000000001043561a882930000000040000000000000015af1d78b58c4000000005000000000000001b1ae4d6e2ef5000000006000000000000002086ac351052600000";

    bytes public malformedPayloadTooShortHex = hex"0000";

    bytes public malformedPayloadBadLengthHex =
        hex"00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000";

    // ===== DEPLOY TESTS =====

    function setUp() public {
        implementation = new UpgradeableStorkFast();

        bytes memory initializeData = abi.encodeWithSelector(
            UpgradeableStorkFast.initialize.selector,
            owner,
            signerAddress,
            verificationFee
        );

        proxy = new TransparentUpgradeableProxy(
            address(implementation),
            owner, // admin
            initializeData
        );

        storkFast = UpgradeableStorkFast(payable(address(proxy)));
    }

    function test_ShouldReturnOwner() public view {
        assertEq(storkFast.owner(), owner);
    }

    function test_ShouldReturnSignerAddress() public view {
        assertEq(storkFast.signerAddress(), signerAddress);
    }

    function test_ShouldReturnVerificationFee() public view {
        assertEq(storkFast.verificationFeeInWei(), verificationFee);
    }

    function test_ShouldNotAllowRenouncingOwnership() public {
        vm.prank(owner);
        vm.expectRevert("Ownable: renouncing ownership is disabled");
        storkFast.renounceOwnership();
    }

    function test_ShouldReturnVersion() public view {
        string memory version = storkFast.version();
        assertEq(version, "1.0.0");
    }

    // ===== UPDATE SIGNER ADDRESS TESTS =====

    function test_UpdateSignerAddress_Successful() public {
        address newAddress = address(0x123);

        vm.prank(owner);
        storkFast.updateSignerAddress(newAddress);

        assertEq(storkFast.signerAddress(), newAddress);
    }

    function test_UpdateSignerAddress_RevertsIfNotOwner() public {
        address newAddress = address(0x123);

        vm.prank(otherAccount);
        vm.expectRevert(); // OwnableUnauthorizedAccount
        storkFast.updateSignerAddress(newAddress);
    }

    function test_UpdateSignerAddress_RevertsIfZeroAddress() public {
        vm.prank(owner);
        vm.expectRevert("Signer address cannot be 0 address");
        storkFast.updateSignerAddress(address(0));
    }

    // ===== UPDATE VERIFICATION FEE TESTS =====

    function test_UpdateVerificationFeeInWei_Successful() public {
        uint256 newFee = 500;

        vm.prank(owner);
        storkFast.updateVerificationFeeInWei(newFee);

        assertEq(storkFast.verificationFeeInWei(), newFee);
    }

    function test_UpdateVerificationFeeInWei_RevertsIfNotOwner() public {
        uint256 newFee = 500;

        vm.prank(otherAccount);
        vm.expectRevert(); // OwnableUnauthorizedAccount
        storkFast.updateVerificationFeeInWei(newFee);
    }

    function test_UpdateVerificationFeeInWei_CanSetToZero() public {
        vm.prank(owner);
        storkFast.updateVerificationFeeInWei(0);

        assertEq(storkFast.verificationFeeInWei(), 0);
    }

    // ===== VERIFY SIGNED ECDSA PAYLOAD TESTS =====

    function test_VerifySignedECDSAPayload_RevertsWithInsufficientFee() public {
        vm.expectRevert(
            abi.encodeWithSelector(StorkFastErrors.InsufficientFee.selector)
        );
        storkFast.verifySignedECDSAPayload{value: verificationFee - 1}(
            validPayloadHex_1
        );
    }

    function test_VerifySignedECDSAPayload_SuccessfulWithValidSignature()
        public
    {
        uint256 initialBalance = address(storkFast).balance;

        bool result = storkFast.verifySignedECDSAPayload{
            value: verificationFee
        }(validPayloadHex_1);
        assertTrue(result);
        assertGt(address(storkFast).balance, initialBalance);
    }

    function test_VerifySignedECDSAPayload_SuccessfulWithValidSignature_MultipleAssets()
        public
    {
        uint256 initialBalance = address(storkFast).balance;
        bool result = storkFast.verifySignedECDSAPayload{
            value: verificationFee
        }(validPayloadHex_1_2_3_4_5_6);
        assertTrue(result);
        assertGt(address(storkFast).balance, initialBalance);
    }

    function test_VerifySignedECDSAPayload_SuccessfulWithValidSignature_NegativeAsset()
        public
    {
        uint256 initialBalance = address(storkFast).balance;
        bool result = storkFast.verifySignedECDSAPayload{
            value: verificationFee
        }(validNegativePayloadHex_7);
        assertTrue(result);
        assertGt(address(storkFast).balance, initialBalance);
    }

    function test_VerifySignedECDSAPayload_FailsWithInvalidSignature() public {
        bool result = storkFast.verifySignedECDSAPayload{
            value: verificationFee
        }(invalidSignaturePayloadHex);
        assertFalse(result);
    }

    function test_VerifySignedECDSAPayload_SuccessfulWithExcessFee() public {
        uint256 initialBalance = address(storkFast).balance;
        bool result = storkFast.verifySignedECDSAPayload{
            value: verificationFee * 2
        }(validPayloadHex_1);
        assertTrue(result);
        assertGt(address(storkFast).balance, initialBalance);
    }

    // ===== VERIFY AND DESERIALIZE SIGNED ECDSA PAYLOAD TESTS =====

    function test_VerifyAndDeserialize_RevertsWithInsufficientFee() public {
        vm.expectRevert(
            abi.encodeWithSelector(StorkFastErrors.InsufficientFee.selector)
        );
        storkFast.verifyAndDeserializeSignedECDSAPayload{
            value: verificationFee - 1
        }(validPayloadHex_1);
    }

    function test_VerifyAndDeserialize_RevertsWithInvalidSignature() public {
        vm.expectRevert(
            abi.encodeWithSelector(StorkFastErrors.InvalidSignature.selector)
        );
        storkFast.verifyAndDeserializeSignedECDSAPayload{
            value: verificationFee
        }(invalidSignaturePayloadHex);
    }

    function test_VerifyAndDeserialize_SuccessfulWithSingleAsset() public {
        uint256 initialBalance = address(storkFast).balance;

        StorkFastStructs.Asset[] memory updates = storkFast
            .verifyAndDeserializeSignedECDSAPayload{value: verificationFee}(
            validPayloadHex_1
        );

        assertGt(address(storkFast).balance, initialBalance);
        assertEq(updates.length, 1);
        assertEq(updates[0].assetID, 1);
        assertEq(
            updates[0].temporalNumericValue.quantizedValue,
            100000000000000000000
        );
    }

    function test_VerifyAndDeserialize_SuccessfulWithSingleAsset_NegativeAsset()
        public
    {
        uint256 initialBalance = address(storkFast).balance;
        StorkFastStructs.Asset[] memory updates = storkFast
            .verifyAndDeserializeSignedECDSAPayload{value: verificationFee}(
            validNegativePayloadHex_7
        );
        assertGt(address(storkFast).balance, initialBalance);
        assertEq(updates.length, 1);
        assertEq(updates[0].assetID, 7);
        assertEq(
            updates[0].temporalNumericValue.quantizedValue,
            -700000000000000000000
        );
    }

    function test_VerifyAndDeserialize_SuccessfulWithMultipleAssets() public {
        uint256 initialBalance = address(storkFast).balance;
        StorkFastStructs.Asset[] memory updates = storkFast
            .verifyAndDeserializeSignedECDSAPayload{value: verificationFee}(
            validPayloadHex_1_2_3_4_5_6
        );

        assertGt(address(storkFast).balance, initialBalance);
        assertEq(updates.length, 6);
        assertEq(updates[0].assetID, 1);
        for (uint256 i = 0; i < updates.length; i++) {
            uint16 assetID = updates[i].assetID;
            assertEq(assetID, i + 1);
            assertEq(
                int256(updates[i].temporalNumericValue.quantizedValue),
                int256((uint256(assetID) * 100) * 10 ** 18)
            );
        }
    }

    function test_VerifyAndDeserialize_RevertsWithMalformedPayloadTooShort()
        public
    {
        vm.expectRevert(
            abi.encodeWithSelector(StorkFastErrors.InvalidPayload.selector)
        );
        storkFast.verifyAndDeserializeSignedECDSAPayload{
            value: verificationFee
        }(malformedPayloadTooShortHex);
    }

    function test_VerifyAndDeserialize_RevertsWithMalformedPayloadBadLength()
        public
    {
        vm.expectRevert(
            abi.encodeWithSelector(StorkFastErrors.InvalidPayload.selector)
        );
        storkFast.verifyAndDeserializeSignedECDSAPayload{
            value: verificationFee
        }(malformedPayloadBadLengthHex);
    }
}
