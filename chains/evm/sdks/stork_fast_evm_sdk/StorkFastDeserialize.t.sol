// SPDX-License-Identifier: Apache 2
pragma solidity ^0.8.28;

import "forge-std/src/Test.sol";
import "./StorkFastDeserializeTestHarness.sol";

contract StorkFastDeserializeTest is Test {
    StorkFastDeserializeTestHarness public testHarness;

    bytes public validPayloadHex_1 =
        hex"690a1cce3cf72e889b699ef800fb80ba47c123ce7422517bfcd2ce0c701423a12938fb6fcf4424c21e39688cd41dd8919b149642b590328ada7532c88b5f0d6b010001187f4dcc627041f8000100000000000000056bc75e2d63100000";

    bytes public validSignatureHex_1 =
        hex"690a1cce3cf72e889b699ef800fb80ba47c123ce7422517bfcd2ce0c701423a12938fb6fcf4424c21e39688cd41dd8919b149642b590328ada7532c88b5f0d6b01";
    bytes public validVerifiablePayloadHex_1 =
        hex"0001187f4dcc627041f8000100000000000000056bc75e2d63100000";
    uint16 public validTaxonomyID_1 = 1;
    uint64 public validTimestampNs_1 = 1765215119172715000;

    bytes public validPayloadHex_1_2_3_4_5_6 =
        hex"5985a517300f6b80a8b6fa7cb2477ef9b57f9cf27706f42e38e096e5350b034730c755a25a416ecd0021ef32a1b36de88f051a5125fb6ae1306dce7c4e9ffa03010001187f4dda764f46c0000100000000000000056bc75e2d631000000002000000000000000ad78ebc5ac62000000003000000000000001043561a882930000000040000000000000015af1d78b58c4000000005000000000000001b1ae4d6e2ef5000000006000000000000002086ac351052600000";

    bytes public validSignatureHex_1_2_3_4_5_6 =
        hex"5985a517300f6b80a8b6fa7cb2477ef9b57f9cf27706f42e38e096e5350b034730c755a25a416ecd0021ef32a1b36de88f051a5125fb6ae1306dce7c4e9ffa0301";
    bytes public validVerifiablePayloadHex_1_2_3_4_5_6 =
        hex"0001187f4dda764f46c0000100000000000000056bc75e2d631000000002000000000000000ad78ebc5ac62000000003000000000000001043561a882930000000040000000000000015af1d78b58c4000000005000000000000001b1ae4d6e2ef5000000006000000000000002086ac351052600000";
    uint16 public validTaxonomyID_1_2_3_4_5_6 = 1;
    uint64 public validTimestampNs_1_2_3_4_5_6 = 1765215179635640000;

    bytes public validNegativePayloadHex_7 =
        hex"e2c0c8ed493d2a7ced7092010bf670c59c33a54d374416a612ebc4064faac878430e96f68a1d2a7f530b51770340e56edca3d587646ee23aef36a3c09321da0e000001187f506ddec854e00007ffffffffffffffda0d8c6cc24a900000";

    bytes public validSignatureHex_7 =
        hex"e2c0c8ed493d2a7ced7092010bf670c59c33a54d374416a612ebc4064faac878430e96f68a1d2a7f530b51770340e56edca3d587646ee23aef36a3c09321da0e00";
    bytes public validVerifiablePayloadHex_7 =
        hex"0001187f506ddec854e00007ffffffffffffffda0d8c6cc24a900000";
    uint16 public validTaxonomyID_7 = 1;
    uint64 public validTimestampNs_7 = 1765218011771852000;

    bytes public invalidSignaturePayloadHex =
        hex"1985a517300f6b80a8b6fa7cb2477ef9b57f9cf27706f42e38e096e5350b034730c755a25a416ecd0021ef32a1b36de88f051a5125fb6ae1306dce7c4e9ffa03010001187f4dda764f46c0000100000000000000056bc75e2d631000000002000000000000000ad78ebc5ac62000000003000000000000001043561a882930000000040000000000000015af1d78b58c4000000005000000000000001b1ae4d6e2ef5000000006000000000000002086ac351052600001";

    bytes public malformedPayloadTooShortHex = hex"0000";

    bytes public malformedPayloadBadLengthHex =
        hex"00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000";

    function setUp() public {
        testHarness = new StorkFastDeserializeTestHarness();
    }

    // ===== SPLIT SIGNED ECDSA PAYLOAD TESTS =====

    function test_SplitSignedECDSA_RevertsWithInvalidPayload_TooShort() public {
        vm.expectRevert(
            abi.encodeWithSelector(StorkFastErrors.InvalidPayload.selector)
        );
        testHarness.splitSignedECDSAPayload(malformedPayloadTooShortHex);
    }

    function test_SplitSignedECDSA_RevertsWithInvalidPayload_BadLength()
        public
    {
        vm.expectRevert(
            abi.encodeWithSelector(StorkFastErrors.InvalidPayload.selector)
        );
        testHarness.splitSignedECDSAPayload(malformedPayloadBadLengthHex);
    }

    function test_SplitSignedECDSA_SuccessfulWithValidPayload_SingleAsset()
        public
        view
    {
        (bytes memory signature, bytes memory verifiablePayload) = testHarness
            .splitSignedECDSAPayload(validPayloadHex_1);
        assertEq(signature.length, 65);
        assertEq(verifiablePayload.length, validPayloadHex_1.length - 65);
        assertEq(signature, validSignatureHex_1);
        assertEq(verifiablePayload, validVerifiablePayloadHex_1);
    }

    function test_SplitSignedECDSA_SuccessfulWithValidPayload_MultipleAssets()
        public
        view
    {
        (bytes memory signature, bytes memory verifiablePayload) = testHarness
            .splitSignedECDSAPayload(validPayloadHex_1_2_3_4_5_6);
        assertEq(signature.length, 65);
        assertEq(
            verifiablePayload.length,
            validPayloadHex_1_2_3_4_5_6.length - 65
        );
        assertEq(signature, validSignatureHex_1_2_3_4_5_6);
        assertEq(verifiablePayload, validVerifiablePayloadHex_1_2_3_4_5_6);
    }

    function test_SplitSignedECDSA_SuccessfulWithValidPayload_NegativeAsset()
        public
        view
    {
        (bytes memory signature, bytes memory verifiablePayload) = testHarness
            .splitSignedECDSAPayload(validNegativePayloadHex_7);
        assertEq(signature.length, 65);
        assertEq(
            verifiablePayload.length,
            validNegativePayloadHex_7.length - 65
        );
        assertEq(signature, validSignatureHex_7);
        assertEq(verifiablePayload, validVerifiablePayloadHex_7);
    }

    // ===== DESERIALIZE SIGNED ECDSA PAYLOAD HEADER TESTS =====

    function test_DeserializeSignedECDSAPayloadHeader_RevertsWithInvalidPayload_TooShort()
        public
    {
        vm.expectRevert(
            abi.encodeWithSelector(StorkFastErrors.InvalidPayload.selector)
        );
        testHarness.deserializeSignedECDSAPayloadHeader(
            malformedPayloadTooShortHex
        );
    }

    function test_DeserializeSignedECDSAPayloadHeader_RevertsWithInvalidPayload_BadLength()
        public
    {
        vm.expectRevert(
            abi.encodeWithSelector(StorkFastErrors.InvalidPayload.selector)
        );
        testHarness.deserializeSignedECDSAPayloadHeader(
            malformedPayloadBadLengthHex
        );
    }

    function test_DeserializeSignedECDSAPayloadHeader_SuccessfulWithValidPayload_SingleAsset()
        public
        view
    {
        (
            bytes memory signature,
            uint16 taxonomyID,
            uint64 timestampNs
        ) = testHarness.deserializeSignedECDSAPayloadHeader(validPayloadHex_1);
        assertEq(signature, validSignatureHex_1);
        assertEq(taxonomyID, validTaxonomyID_1);
        assertEq(timestampNs, validTimestampNs_1);
    }

    function test_DeserializeSignedECDSAPayloadHeader_SuccessfulWithValidPayload_MultipleAssets()
        public
        view
    {
        (
            bytes memory signature,
            uint16 taxonomyID,
            uint64 timestampNs
        ) = testHarness.deserializeSignedECDSAPayloadHeader(
                validPayloadHex_1_2_3_4_5_6
            );
        assertEq(signature, validSignatureHex_1_2_3_4_5_6);
        assertEq(taxonomyID, validTaxonomyID_1_2_3_4_5_6);
        assertEq(timestampNs, validTimestampNs_1_2_3_4_5_6);
    }

    function test_DeserializeSignedECDSAPayloadHeader_SuccessfulWithValidPayload_NegativeAsset()
        public
        view
    {
        (
            bytes memory signature,
            uint16 taxonomyID,
            uint64 timestampNs
        ) = testHarness.deserializeSignedECDSAPayloadHeader(
                validNegativePayloadHex_7
            );
        assertEq(signature, validSignatureHex_7);
        assertEq(taxonomyID, validTaxonomyID_7);
        assertEq(timestampNs, validTimestampNs_7);
    }

    // ===== DESERIALIZE ASSETS FROM SIGNED ECDSA PAYLOAD TESTS =====

    function test_DeserializeAssetsFromSignedECDSAPayload_RevertsWithInvalidPayload_TooShort()
        public
    {
        vm.expectRevert(
            abi.encodeWithSelector(StorkFastErrors.InvalidPayload.selector)
        );
        testHarness.deserializeAssetsFromSignedECDSAPayload(
            malformedPayloadTooShortHex
        );
    }

    function test_DeserializeAssetsFromSignedECDSAPayload_RevertsWithInvalidPayload_BadLength()
        public
    {
        vm.expectRevert(
            abi.encodeWithSelector(StorkFastErrors.InvalidPayload.selector)
        );
        testHarness.deserializeAssetsFromSignedECDSAPayload(
            malformedPayloadBadLengthHex
        );
    }

    function test_DeserializeAssetsFromSignedECDSAPayload_SuccessfulWithValidPayload_SingleAsset()
        public
        view
    {
        StorkFastStructs.Asset[] memory assets = testHarness
            .deserializeAssetsFromSignedECDSAPayload(validPayloadHex_1);
        assertEq(assets.length, 1);
        assertEq(assets[0].assetID, 1);
        assertEq(
            assets[0].temporalNumericValue.quantizedValue,
            100000000000000000000
        );
    }

    function test_DeserializeAssetsFromSignedECDSAPayload_SuccessfulWithValidPayload_MultipleAssets()
        public
        view
    {
        StorkFastStructs.Asset[] memory assets = testHarness
            .deserializeAssetsFromSignedECDSAPayload(
                validPayloadHex_1_2_3_4_5_6
            );
        assertEq(assets.length, 6);
        for (uint256 i = 0; i < assets.length; i++) {
            assertEq(assets[i].assetID, i + 1);
            assertEq(
                int256(assets[i].temporalNumericValue.quantizedValue),
                int256((uint256(i + 1) * 100) * 10 ** 18)
            );
        }
    }

    function test_DeserializeAssetsFromSignedECDSAPayload_SuccessfulWithValidPayload_NegativeAsset()
        public
        view
    {
        StorkFastStructs.Asset[] memory assets = testHarness
            .deserializeAssetsFromSignedECDSAPayload(validNegativePayloadHex_7);
        assertEq(assets.length, 1);
        assertEq(assets[0].assetID, 7);
        assertEq(
            int256(assets[0].temporalNumericValue.quantizedValue),
            -700000000000000000000
        );
    }
}
