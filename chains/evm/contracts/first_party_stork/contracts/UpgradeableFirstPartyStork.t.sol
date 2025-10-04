// SPDX-License-Identifier: MIT
pragma solidity ^0.8.28;

import "./UpgradeableFirstPartyStork.sol";
import "@storknetwork/first-party-stork-evm-sdk/FirstPartyStorkStructs.sol";
import "@openzeppelin/contracts/proxy/transparent/TransparentUpgradeableProxy.sol";
import "forge-std/Test.sol";

contract UpgradeableFirstPartyStorkTest is Test {
    UpgradeableFirstPartyStork public implementation;
    UpgradeableFirstPartyStork public stork;
    TransparentUpgradeableProxy public proxy;

    // ==== PUBLISHER ADDRESSES ====
    
    address public owner = address(0x1);
    address public publisher1 = address(0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266);
    address public publisher2 = address(0x16eB47a6BBdF1E1D1E9AC23E6f473f1bCAe519C0);
    address public otherAccount = address(0x2);

    // ==== SINGLE UPDATE FEE ====

    uint256 public singleUpdateFee = 100;

    // ==== ASSET PAIRS ====

    string public ethUsd = "ETHUSD";
    string public btcUsd = "BTCUSD";

    // ==== TEST DATA ====

    FirstPartyStorkStructs.PublisherTemporalNumericValueInput public pub1OldEth = FirstPartyStorkStructs.PublisherTemporalNumericValueInput({
        temporalNumericValue: FirstPartyStorkStructs.TemporalNumericValue({
            timestampNs: 1680210933000000000,
            quantizedValue: 1100000000000000000
        }),
        pubKey: publisher1,
        assetPairId: ethUsd,
        r: 0xff867c1b1658e0b4868dbac1ab0961aebd4ea8308939497c73e76f6c393b158e,
        s: 0x74b19977de6e0a56b7a627e16db9e996548ae12ab10eb209355d13794e28f692,
        v: 0x1c
    });

    FirstPartyStorkStructs.PublisherTemporalNumericValueInput public pub1Eth = FirstPartyStorkStructs.PublisherTemporalNumericValueInput({
        temporalNumericValue: FirstPartyStorkStructs.TemporalNumericValue({
            timestampNs: 1680210934000000000,
            quantizedValue: 1000000000000000000
        }),
        pubKey: publisher1,
        assetPairId: ethUsd,
        r: 0xeabf8494b0b64aab3033dac3c821464324ed861ae68ea6e18fecea05f6675f61,
        s: 0x2fdb84b8b4710b1e1926c8cc4073a3bd7fbfcdd03601303462e097a9f3ae667a,
        v: 0x1b
    });

    FirstPartyStorkStructs.PublisherTemporalNumericValueInput public pub1Btc = FirstPartyStorkStructs.PublisherTemporalNumericValueInput({
        temporalNumericValue: FirstPartyStorkStructs.TemporalNumericValue({
            timestampNs: 1680210934000000000,
            quantizedValue: 2000000000000000000
        }),
        pubKey: publisher1,
        assetPairId: btcUsd,
        r: 0x2b9b43a6d18a4768693b679c7b407a71d41ee682bddc544ecbe8c436110950c3,
        s: 0x08febe3ce25e61e1a8004f7608ff5d4ed56249f5462cabe36b5806bb4e12a44c,
        v: 0x1b
    });

    FirstPartyStorkStructs.PublisherTemporalNumericValueInput public pub1Neg = FirstPartyStorkStructs.PublisherTemporalNumericValueInput({
        temporalNumericValue: FirstPartyStorkStructs.TemporalNumericValue({
            timestampNs: 1680210934000000000,
            quantizedValue: -1000000000000000000
        }),
        pubKey: publisher1,
        assetPairId: btcUsd,
        r: 0xcf78de6532aacc02cfd6fff1148d7f2afcd4892a35aa53a5cc2cc92c95829277,
        s: 0x55e9a40115dc85f0f4ad336e9c3c23a36d057272dbbb9eede3337ebb0ae81eb0,
        v: 0x1b
    });

    FirstPartyStorkStructs.PublisherTemporalNumericValueInput public pub2Btc = FirstPartyStorkStructs.PublisherTemporalNumericValueInput({
        temporalNumericValue: FirstPartyStorkStructs.TemporalNumericValue({
            timestampNs: 1721755261000000000,
            quantizedValue: 66078270000000000000000
        }),
        pubKey: publisher2,
        assetPairId: btcUsd,
        r: 0xbb1e6f87445556233f98c085e2e25e5938bb0fa4eee42b7df06f50836ae4e42e,
        s: 0x79f16b0f10c35db8a08088c53dcaa40355dca57db840eb850b9328fc064c6002,
        v: 0x1c
    });

    // ===== DEPLOY TESTS =====

    function setUp() public {
        implementation = new UpgradeableFirstPartyStork();
        
        bytes memory initializeData = abi.encodeWithSelector(
            UpgradeableFirstPartyStork.initialize.selector,
            owner
        );
        
        proxy = new TransparentUpgradeableProxy(
            address(implementation),
            owner, // admin
            initializeData
        );
        
        stork = UpgradeableFirstPartyStork(address(proxy));
    }

    function test_ShouldReturnOwner() public view {
        assertEq(stork.owner(), owner);
    }

    function test_ShouldNotAllowRenouncingOwnership() public {
        vm.prank(owner);
        vm.expectRevert("Ownable: renouncing ownership is disabled");
        stork.renounceOwnership();
    }

    // ===== CREATE PUBLISHER USER TESTS =====
    
    function test_CreatePublisherUser_Successful() public {
        vm.prank(owner);
        stork.createPublisherUser(publisher1, singleUpdateFee);
    }

    function test_CreatePublisherUser_UpdatesIfExists() public {
        uint256 newSingleUpdateFee = 500;
        
        vm.prank(owner);
        stork.createPublisherUser(publisher1, singleUpdateFee);

        FirstPartyStorkStructs.PublisherUser memory publisherUser = stork.getPublisherUser(publisher1);
        assertEq(publisherUser.pubKey, publisher1);
        assertEq(publisherUser.singleUpdateFee, singleUpdateFee);

        vm.prank(owner);
        stork.createPublisherUser(publisher1, newSingleUpdateFee);

        publisherUser = stork.getPublisherUser(publisher1);
        assertEq(publisherUser.pubKey, publisher1);
        assertEq(publisherUser.singleUpdateFee, newSingleUpdateFee);
    }

    function test_CreatePublisherUser_MultiplePublishers() public {
        vm.prank(owner);
        stork.createPublisherUser(publisher1, singleUpdateFee);
        vm.prank(owner);
        stork.createPublisherUser(publisher2, singleUpdateFee*2);

        FirstPartyStorkStructs.PublisherUser memory publisherUser1 = stork.getPublisherUser(publisher1);
        assertEq(publisherUser1.pubKey, publisher1);
        assertEq(publisherUser1.singleUpdateFee, singleUpdateFee);

        FirstPartyStorkStructs.PublisherUser memory publisherUser2 = stork.getPublisherUser(publisher2);
        assertEq(publisherUser2.pubKey, publisher2);
        assertEq(publisherUser2.singleUpdateFee, singleUpdateFee*2);
    }

    function test_CreatePublisherUser_RevertsIfNotOwner() public {
        vm.prank(otherAccount);
        vm.expectRevert(); // OwnableUnauthorizedAccount
        stork.createPublisherUser(publisher1, singleUpdateFee);
    }

    // ===== DELETE PUBLISHER USER TESTS =====
    
    function test_DeletePublisherUser_Successful() public {
        vm.prank(owner);
        stork.createPublisherUser(publisher1, singleUpdateFee);
        
        FirstPartyStorkStructs.PublisherUser memory publisherUser = stork.getPublisherUser(publisher1);
        assertEq(publisherUser.pubKey, publisher1);
        
        vm.prank(owner);
        stork.deletePublisherUser(publisher1);
        
        // Verify it's deleted
        vm.expectRevert(); // NotFound
        stork.getPublisherUser(publisher1);
    }

    function test_DeletePublisherUser_RevertsIfNotFound() public {
        vm.expectRevert(); // NotFound
        stork.deletePublisherUser(publisher1);
    }

    function test_DeletePublisherUser_RevertsIfNotOwner() public {
        vm.prank(otherAccount);
        vm.expectRevert(); // OwnableUnauthorizedAccount
        stork.deletePublisherUser(publisher1);
    }

    // ===== VERIFY PUBLISHER SIGNATURE TESTS =====
    
    function test_VerifyPublisherSignatureV1_ValidSignature() public view {
        bool result = stork.verifyPublisherSignatureV1(
            pub1Eth.pubKey,
            pub1Eth.assetPairId,
            pub1Eth.temporalNumericValue.timestampNs,
            pub1Eth.temporalNumericValue.quantizedValue,
            pub1Eth.r,
            pub1Eth.s,
            pub1Eth.v
        );
        assertTrue(result);
    }

    function test_VerifyPublisherSignatureV1_InvalidSignature() public view {
        bool result = stork.verifyPublisherSignatureV1(
            pub1Eth.pubKey,
            pub1Eth.assetPairId,
            pub1Eth.temporalNumericValue.timestampNs,
            pub1Eth.temporalNumericValue.quantizedValue,
            pub1Eth.r,
            pub1Eth.s,
            pub1Eth.v + 1
        );
        assertFalse(result);
    }

    // ===== GET PUBLISHER USER TESTS =====
    
    function test_GetPublisherUser_RevertsIfNotFound() public {
        vm.expectRevert(); // NotFound
        stork.getPublisherUser(publisher1);
    }

    function test_GetPublisherUser_ReturnsAfterCreation() public {
        vm.prank(owner);
        stork.createPublisherUser(publisher1, singleUpdateFee);
        
        FirstPartyStorkStructs.PublisherUser memory publisherUser = stork.getPublisherUser(publisher1);
        assertEq(publisherUser.pubKey, publisher1);
        assertEq(publisherUser.singleUpdateFee, singleUpdateFee);
    }

    // ===== UPDATE TEMPORAL NUMERIC VALUES TESTS =====

    function test_GetLatestTemporalNumericValue_RevertsIfNeverUpdated() public {
        vm.expectRevert(); // NotFound
        stork.getLatestTemporalNumericValue(publisher1, ethUsd);
    }

    function test_GetHistoricalTemporalNumericValue_RevertsIfNeverUpdated() public {
        vm.expectRevert(); // NotFound
        stork.getHistoricalTemporalNumericValue(publisher1, ethUsd, 0);
    }

    function test_GetHistoricalRecordsCount_ReturnsZeroForNonExistent() public view {
        uint256 count = stork.getHistoricalRecordsCount(publisher1, ethUsd);
        assertEq(count, 0);
    }

    function test_GetCurrentRoundId_ReturnsZeroForNonExistent() public view {
        uint256 roundId = stork.getCurrentRoundId(publisher1, ethUsd);
        assertEq(roundId, 0);
    }
    
    function test_UpdateTemporalNumericValues_RevertsWithInsufficientFee() public {
        vm.prank(owner);
        stork.createPublisherUser(publisher1, singleUpdateFee);
        vm.deal(publisher1, 1 ether);
        
        FirstPartyStorkStructs.PublisherTemporalNumericValueInput[] memory updateData = 
            new FirstPartyStorkStructs.PublisherTemporalNumericValueInput[](1);
        
        updateData[0] = pub1Eth;

        bool[] memory storeHistoric = new bool[](1);
        storeHistoric[0] = false;
        
        // Should revert with insufficient fee
        vm.expectRevert(); // InsufficientFee
        vm.prank(publisher1);
        stork.updateTemporalNumericValues{value: singleUpdateFee - 1}(updateData, storeHistoric);
    }

    function test_UpdateTemporalNumericValues_SuccessfulUpdate() public {
        vm.prank(owner);
        stork.createPublisherUser(publisher1, singleUpdateFee);
        vm.deal(publisher1, 1 ether);
        
        FirstPartyStorkStructs.PublisherTemporalNumericValueInput[] memory updateData = 
            new FirstPartyStorkStructs.PublisherTemporalNumericValueInput[](1);
        
        updateData[0] = pub1Eth;

        bool[] memory storeHistoric = new bool[](1);
        storeHistoric[0] = false;
        
        vm.prank(publisher1);
        stork.updateTemporalNumericValues{value: singleUpdateFee}(updateData, storeHistoric);
        
        FirstPartyStorkStructs.TemporalNumericValue memory latestValue = 
            stork.getLatestTemporalNumericValue(publisher1, ethUsd);
        
        assertEq(latestValue.timestampNs, pub1Eth.temporalNumericValue.timestampNs);
        assertEq(latestValue.quantizedValue, pub1Eth.temporalNumericValue.quantizedValue);
        
        uint256 roundId = stork.getCurrentRoundId(publisher1, ethUsd);
        assertEq(roundId, 0); // Round ID stays 0 because storeHistoric=false (no historical record created)
    }

    function test_UpdateTemporalNumericValues_SuccessfulUpdateWithHistoric() public {
        vm.prank(owner);
        stork.createPublisherUser(publisher1, singleUpdateFee);
        vm.deal(publisher1, 1 ether);
        
        // Create update data with known good signature
        FirstPartyStorkStructs.PublisherTemporalNumericValueInput[] memory updateData = 
            new FirstPartyStorkStructs.PublisherTemporalNumericValueInput[](1);
        
        updateData[0] = pub1Eth;

        bool[] memory storeHistoric = new bool[](1);
        storeHistoric[0] = true;

        vm.prank(publisher1);
        stork.updateTemporalNumericValues{value: singleUpdateFee}(updateData, storeHistoric);
        
        FirstPartyStorkStructs.TemporalNumericValue memory latestValue = 
            stork.getLatestTemporalNumericValue(publisher1, ethUsd);
        
        assertEq(latestValue.timestampNs, pub1Eth.temporalNumericValue.timestampNs);
        assertEq(latestValue.quantizedValue, pub1Eth.temporalNumericValue.quantizedValue);
        
        uint256 historicalCount = stork.getHistoricalRecordsCount(publisher1, ethUsd);
        assertEq(historicalCount, 1);

        uint256 roundId = stork.getCurrentRoundId(publisher1, ethUsd);
        assertEq(roundId, 1);
        
        FirstPartyStorkStructs.TemporalNumericValue memory historicalValue = 
            stork.getHistoricalTemporalNumericValue(publisher1, ethUsd, 0);
        
        assertEq(historicalValue.timestampNs, pub1Eth.temporalNumericValue.timestampNs);
        assertEq(historicalValue.quantizedValue, pub1Eth.temporalNumericValue.quantizedValue);
    }

    function test_MultipleUpdatesWithHistoric() public {
        vm.prank(owner);
        stork.createPublisherUser(publisher1, singleUpdateFee);
        vm.prank(owner);
        stork.createPublisherUser(publisher2, singleUpdateFee);
        
        vm.deal(publisher1, 1 ether);
        vm.deal(publisher2, 1 ether);
        
        FirstPartyStorkStructs.PublisherTemporalNumericValueInput[] memory updateData1 = 
            new FirstPartyStorkStructs.PublisherTemporalNumericValueInput[](1);
        
        updateData1[0] = pub1Eth;

        bool[] memory storeHistoric1 = new bool[](1);
        storeHistoric1[0] = true;
        
        vm.prank(publisher1);
        stork.updateTemporalNumericValues{value: singleUpdateFee}(updateData1, storeHistoric1);
        
        FirstPartyStorkStructs.PublisherTemporalNumericValueInput[] memory updateData2 = 
            new FirstPartyStorkStructs.PublisherTemporalNumericValueInput[](1);
        
        updateData2[0] = pub2Btc;
        
        bool[] memory storeHistoric2 = new bool[](1);
        storeHistoric2[0] = true;
        
        vm.prank(publisher2);
        stork.updateTemporalNumericValues{value: singleUpdateFee}(updateData2, storeHistoric2);
        
        uint256 historicalCount = stork.getHistoricalRecordsCount(publisher1, ethUsd);
        assertEq(historicalCount, 1);
        
        uint256 roundId = stork.getCurrentRoundId(publisher1, ethUsd);
        assertEq(roundId, 1);

        uint256 historicalCount2 = stork.getHistoricalRecordsCount(publisher2, btcUsd);
        assertEq(historicalCount2, 1);

        uint256 roundId2 = stork.getCurrentRoundId(publisher2, btcUsd);
        assertEq(roundId2, 1);
    }

    function test_UpdateTemporalNumericValues_MultipleUpdatesAtOnce() public {
        vm.prank(owner);
        stork.createPublisherUser(publisher1, singleUpdateFee);
        vm.deal(publisher1, 1 ether);
        
        FirstPartyStorkStructs.PublisherTemporalNumericValueInput[] memory updateData = 
            new FirstPartyStorkStructs.PublisherTemporalNumericValueInput[](2);
        
        updateData[0] = pub1Eth;
        updateData[1] = pub1Btc;
        
        bool[] memory storeHistoric = new bool[](2);
        storeHistoric[0] = true;
        storeHistoric[1] = true;
        
        vm.prank(publisher1);
        stork.updateTemporalNumericValues{value: singleUpdateFee*2}(updateData, storeHistoric);

        uint256 historicalCount = stork.getHistoricalRecordsCount(publisher1, ethUsd);
        assertEq(historicalCount, 1);

        uint256 roundId = stork.getCurrentRoundId(publisher1, ethUsd);
        assertEq(roundId, 1);

        uint256 historicalCount2 = stork.getHistoricalRecordsCount(publisher1, btcUsd);
        assertEq(historicalCount2, 1);

        uint256 roundId2 = stork.getCurrentRoundId(publisher1, btcUsd);
        assertEq(roundId2, 1);
    }

    function test_UpdateTemporalNumericValues_SuccessfulUpdateWithNegativeValue() public {
        vm.prank(owner);
        stork.createPublisherUser(publisher1, singleUpdateFee);
        vm.deal(publisher1, 1 ether);

        FirstPartyStorkStructs.PublisherTemporalNumericValueInput[] memory updateData = 
            new FirstPartyStorkStructs.PublisherTemporalNumericValueInput[](1);
        
        updateData[0] = pub1Neg;

        bool[] memory storeHistoric = new bool[](1);
        storeHistoric[0] = true;

        vm.prank(publisher1);
        stork.updateTemporalNumericValues{value: singleUpdateFee}(updateData, storeHistoric);
        
        FirstPartyStorkStructs.TemporalNumericValue memory latestValue = 
            stork.getLatestTemporalNumericValue(publisher1, btcUsd);
        assertEq(latestValue.quantizedValue, -1000000000000000000);
    }

    function test_UpdateTemporalNumericValues_OlderUpdateIsNotStored() public {
        vm.prank(owner);
        stork.createPublisherUser(publisher1, singleUpdateFee);
        vm.deal(publisher1, 1 ether);
        
        FirstPartyStorkStructs.PublisherTemporalNumericValueInput[] memory updateData = 
            new FirstPartyStorkStructs.PublisherTemporalNumericValueInput[](1);
        
        updateData[0] = pub1Eth;

        bool[] memory storeHistoric = new bool[](1);
        storeHistoric[0] = true;

        vm.prank(publisher1);
        stork.updateTemporalNumericValues{value: singleUpdateFee}(updateData, storeHistoric);

        uint256 historicalCount = stork.getHistoricalRecordsCount(publisher1, ethUsd);
        assertEq(historicalCount, 1);

        FirstPartyStorkStructs.TemporalNumericValue memory historicalValue = 
            stork.getLatestTemporalNumericValue(publisher1, ethUsd);
        assertEq(historicalValue.timestampNs, pub1Eth.temporalNumericValue.timestampNs);
        assertEq(historicalValue.quantizedValue, pub1Eth.temporalNumericValue.quantizedValue);

        updateData[0] = pub1OldEth;

        vm.prank(publisher1);
        vm.expectRevert(); // NoFreshUpdate
        stork.updateTemporalNumericValues{value: singleUpdateFee}(updateData, storeHistoric);

        uint256 historicalCount2 = stork.getHistoricalRecordsCount(publisher1, ethUsd);
        assertEq(historicalCount2, 1);

        FirstPartyStorkStructs.TemporalNumericValue memory historicalValue2 = 
            stork.getLatestTemporalNumericValue(publisher1, ethUsd);
        assertEq(historicalValue2.timestampNs, pub1Eth.temporalNumericValue.timestampNs);
        assertEq(historicalValue2.quantizedValue, pub1Eth.temporalNumericValue.quantizedValue);
    }

    function test_UpdateTemporalNumericValues_SameUpdateTwice() public {
        vm.prank(owner);
        stork.createPublisherUser(publisher1, singleUpdateFee);
        vm.deal(publisher1, 1 ether);
        
        FirstPartyStorkStructs.PublisherTemporalNumericValueInput[] memory updateData = 
            new FirstPartyStorkStructs.PublisherTemporalNumericValueInput[](1);
        
        updateData[0] = pub1Eth;
        
        bool[] memory storeHistoric = new bool[](1);
        storeHistoric[0] = true;
        
        vm.prank(publisher1);
        stork.updateTemporalNumericValues{value: singleUpdateFee}(updateData, storeHistoric);

        vm.prank(publisher1);
        stork.updateTemporalNumericValues{value: singleUpdateFee}(updateData, storeHistoric);

        // Timestamp is same, so both updates are stored
        uint256 historicalCount = stork.getHistoricalRecordsCount(publisher1, ethUsd);
        assertEq(historicalCount, 2);
    }

    function test_UpdateTemporalNumericValues_EmptyUpdate() public {
        vm.prank(owner);
        stork.createPublisherUser(publisher1, singleUpdateFee);
        vm.deal(publisher1, 1 ether);
        
        FirstPartyStorkStructs.PublisherTemporalNumericValueInput[] memory updateData = 
            new FirstPartyStorkStructs.PublisherTemporalNumericValueInput[](0);

        bool[] memory storeHistoric = new bool[](0);

        vm.prank(publisher1);
        vm.expectRevert(); // NoFreshUpdate
        stork.updateTemporalNumericValues{value: singleUpdateFee}(updateData, storeHistoric);

    }

    function test_MultiplePublishersForSameAssetPair() public {
        vm.prank(owner);
        stork.createPublisherUser(publisher1, singleUpdateFee);
        vm.prank(owner);
        stork.createPublisherUser(publisher2, singleUpdateFee);
        
        vm.deal(publisher1, 1 ether);
        vm.deal(publisher2, 1 ether);

        FirstPartyStorkStructs.PublisherTemporalNumericValueInput[] memory updateData1 = 
            new FirstPartyStorkStructs.PublisherTemporalNumericValueInput[](1);
        
        updateData1[0] = pub1Btc;

        bool[] memory storeHistoric1 = new bool[](1);
        storeHistoric1[0] = true;

        vm.prank(publisher1);
        stork.updateTemporalNumericValues{value: singleUpdateFee}(updateData1, storeHistoric1);

        FirstPartyStorkStructs.PublisherTemporalNumericValueInput[] memory updateData2 = 
            new FirstPartyStorkStructs.PublisherTemporalNumericValueInput[](1);
        
        updateData2[0] = pub2Btc;
        
        bool[] memory storeHistoric2 = new bool[](1);
        storeHistoric2[0] = true;
        
        vm.prank(publisher2);
        stork.updateTemporalNumericValues{value: singleUpdateFee}(updateData2, storeHistoric2);

        uint256 historicalCount = stork.getHistoricalRecordsCount(publisher1, btcUsd);
        assertEq(historicalCount, 1);

        uint256 roundId = stork.getCurrentRoundId(publisher1, btcUsd);
        assertEq(roundId, 1);

        uint256 historicalCount2 = stork.getHistoricalRecordsCount(publisher2, btcUsd);
        assertEq(historicalCount2, 1);

        uint256 roundId2 = stork.getCurrentRoundId(publisher2, btcUsd);
        assertEq(roundId2, 1);
    }

    function test_DeletePublisherUser_AfterSuccessfulUpdates() public {
        vm.prank(owner);
        stork.createPublisherUser(publisher1, singleUpdateFee);
        vm.deal(publisher1, 1 ether);
        
        FirstPartyStorkStructs.PublisherTemporalNumericValueInput[] memory updateData = 
            new FirstPartyStorkStructs.PublisherTemporalNumericValueInput[](1);
        
        updateData[0] = pub1Eth;
        
        bool[] memory storeHistoric = new bool[](1);
        storeHistoric[0] = true;
        
        vm.prank(publisher1);
        stork.updateTemporalNumericValues{value: singleUpdateFee}(updateData, storeHistoric);
        
        // Verify update was successful
        FirstPartyStorkStructs.TemporalNumericValue memory latestValue = 
            stork.getLatestTemporalNumericValue(publisher1, ethUsd);
        assertEq(latestValue.timestampNs, pub1Eth.temporalNumericValue.timestampNs);

        uint256 historicalCount = stork.getHistoricalRecordsCount(publisher1, ethUsd);
        assertEq(historicalCount, 1);
        
        vm.prank(owner);
        stork.deletePublisherUser(publisher1);
        
        // Verify publisher user is deleted
        vm.expectRevert(); // NotFound
        stork.getPublisherUser(publisher1);
        
        // Historical data should still be accessible
        FirstPartyStorkStructs.TemporalNumericValue memory stillLatestValue = 
            stork.getLatestTemporalNumericValue(publisher1, ethUsd);
        assertEq(stillLatestValue.timestampNs, pub1Eth.temporalNumericValue.timestampNs);
        
        uint256 stillHistoricalCount = stork.getHistoricalRecordsCount(publisher1, ethUsd);
        assertEq(stillHistoricalCount, 1);
    }
}
