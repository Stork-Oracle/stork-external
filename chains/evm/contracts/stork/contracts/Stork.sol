// SPDX-License-Identifier: Apache 2

pragma solidity >=0.8.24 <0.9.0;

import "@storknetwork/stork-evm-sdk/IStorkEvents.sol";
import "@storknetwork/stork-evm-sdk/StorkStructs.sol";
import "@storknetwork/stork-evm-sdk/StorkErrors.sol";
import "@storknetwork/stork-evm-sdk/IStork.sol";
import "@storknetwork/stork-evm-sdk/LibCodec.sol";

import "./StorkGetters.sol";
import "./StorkSetters.sol";
import "./StorkVerify.sol";

abstract contract Stork is StorkGetters, StorkSetters, StorkVerify, IStork {
    function _initialize(
        address storkPublicKey,
        uint validTimePeriodSeconds,
        uint singleUpdateFeeInWei
    ) internal {
        StorkSetters.setValidTimePeriodSeconds(validTimePeriodSeconds);
        StorkSetters.setSingleUpdateFeeInWei(singleUpdateFeeInWei);
        StorkSetters.setStorkPublicKey(storkPublicKey);
    }

    function _isValidStorkSigner(
        bytes32 id,
        uint256 recvTime,
        int256 quantizedValue,
        bytes32 publisherMerkleRoot,
        bytes32 valueComputeAlgHash,
        bytes32 r,
        bytes32 s,
        uint8 v
    ) private view returns (bool) {
        address[] memory list = _state.signingAddressList;
        for (uint i = 0; i < list.length; i++) {
            if (verifyStorkSignatureV1(list[i], id, recvTime, quantizedValue, publisherMerkleRoot, valueComputeAlgHash, r, s, v)) {
                return true;
            }
        }
        // Legacy fallback: storkPublicKey always reflects the current canonical signing key
        return verifyStorkSignatureV1(storkPublicKey(), id, recvTime, quantizedValue, publisherMerkleRoot, valueComputeAlgHash, r, s, v);
    }

    function updateTemporalNumericValuesV1(
        StorkStructs.TemporalNumericValueInput[] calldata updateData
    ) public payable {
        uint16 numUpdates = 0;
        for (uint i = 0; i < updateData.length; i++) {
            if (!_isValidStorkSigner(
                updateData[i].id,
                updateData[i].temporalNumericValue.timestampNs,
                updateData[i].temporalNumericValue.quantizedValue,
                updateData[i].publisherMerkleRoot,
                updateData[i].valueComputeAlgHash,
                updateData[i].r,
                updateData[i].s,
                updateData[i].v
            )) revert StorkErrors.InvalidSignature();
            if (updateLatestValueIfNecessary(updateData[i])) numUpdates++;
        }
        _validateUpdatesAndFee(numUpdates);
    }

    /// @notice Same as updateTemporalNumericValuesV1 but accepts flat uint256[] calldata.
    /// @dev Each entry is 6 consecutive uint256 words (see LibCodec for layout).
    ///      Saves ~24% calldata bytes (~8% calldata gas) vs ABI-encoded struct array.
    function updateTemporalNumericValuesV1Packed(
        uint256[] calldata packedData
    ) public payable {
        StorkStructs.TemporalNumericValueInput[] memory updateData = LibCodec.decode(packedData);
        uint16 numUpdates = 0;
        for (uint i = 0; i < updateData.length; i++) {
            if (!_isValidStorkSigner(
                updateData[i].id,
                updateData[i].temporalNumericValue.timestampNs,
                updateData[i].temporalNumericValue.quantizedValue,
                updateData[i].publisherMerkleRoot,
                updateData[i].valueComputeAlgHash,
                updateData[i].r,
                updateData[i].s,
                updateData[i].v
            )) revert StorkErrors.InvalidSignature();
            if (updateLatestValueIfNecessary(updateData[i])) numUpdates++;
        }
        _validateUpdatesAndFee(numUpdates);
    }

    function _validateUpdatesAndFee(uint16 numUpdates) private {
        if (numUpdates == 0) revert StorkErrors.NoFreshUpdate();
        uint requiredFee = getTotalFee(numUpdates);
        if (msg.value < requiredFee) revert StorkErrors.InsufficientFee();
    }

    function getUpdateFeeV1(
        StorkStructs.TemporalNumericValueInput[] calldata updateData
    ) public view returns (uint feeAmount) {
        return getTotalFee(updateData.length);
    }

    function getTemporalNumericValueV1(
        bytes32 id
    ) public view returns (StorkStructs.TemporalNumericValue memory value) {
        StorkStructs.TemporalNumericValue memory numericValue = latestCanonicalTemporalNumericValue(id);
        if (numericValue.timestampNs == 0) {
            revert StorkErrors.NotFound();
        }

        uint64 lastTimestampSeconds = numericValue.timestampNs / 1000000000;
        if (block.timestamp >= lastTimestampSeconds && block.timestamp - lastTimestampSeconds > validTimePeriodSeconds()) {
            revert StorkErrors.StaleValue();
        }
        return numericValue;
    }

    function getTemporalNumericValueUnsafeV1(
        bytes32 id
    ) public view returns (StorkStructs.TemporalNumericValue memory value) {
        StorkStructs.TemporalNumericValue memory numericValue = latestCanonicalTemporalNumericValue(id);
        if (numericValue.timestampNs == 0) {
            revert StorkErrors.NotFound();
        }

        return numericValue;
    }

    function getTemporalNumericValuesUnsafeV1(
        bytes32[] calldata ids
    ) public view returns (StorkStructs.TemporalNumericValue[] memory values) {
        values = new StorkStructs.TemporalNumericValue[](ids.length);
        for (uint i = 0; i < ids.length; i++) {
            values[i] = latestCanonicalTemporalNumericValue(ids[i]);
            if (values[i].timestampNs == 0) {
                revert StorkErrors.NotFound();
            }
        }
        return values;
    }

    function verifyPublisherSignaturesV1(
        StorkStructs.PublisherSignature[] calldata signatures,
        bytes32 merkleRoot
    ) public pure returns (bool) {
        bytes32[] memory hashes = new bytes32[](signatures.length);

        for (uint i = 0; i < signatures.length; i++) {
            if(!verifyPublisherSignatureV1(
                signatures[i].pubKey,
                signatures[i].assetPairId,
                signatures[i].timestamp,
                signatures[i].quantizedValue,
                signatures[i].r,
                signatures[i].s,
                signatures[i].v
            )) return false;
            bytes32 computed = getPublisherMessageHash(
                signatures[i].pubKey,
                signatures[i].assetPairId,
                signatures[i].timestamp,
                signatures[i].quantizedValue
            );
            hashes[i] = computed;
        }
        return verifyMerkleRoot(hashes, merkleRoot);
    }

    function packTemporalNumericValueInputs(
        StorkStructs.TemporalNumericValueInput[] calldata inputs
    ) external pure returns (uint256[] memory) {
        return LibCodec.encode(inputs);
    }

    function version() public pure returns (string memory) {
        return "1.0.6";
    }

    function getTotalFee(
        uint totalNumUpdates
    ) private view returns (uint requiredFee) {
        return totalNumUpdates * singleUpdateFeeInWei();
    }

    function updateValidTimePeriodSeconds(
        uint validTimePeriodSeconds
    ) public virtual;

    function updateSingleUpdateFeeInWei(
        uint singleUpdateFeeInWei
    ) public virtual;

    function updateStorkPublicKey(address storkPublicKey) public virtual;

    function addSigningAddress(address signingAddress) public virtual;

    function removeSigningAddress(address signingAddress) public virtual;
}
