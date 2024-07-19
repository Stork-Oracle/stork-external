// SPDX-License-Identifier: Apache 2

pragma solidity >=0.8.24 <0.9.0;

import "./StorkGetters.sol";
import "./StorkSetters.sol";
import "./StorkStructs.sol";
import "./StorkErrors.sol";
import "./StorkVerify.sol";

abstract contract Stork is StorkGetters, StorkSetters, StorkVerify {
    function _initialize(
        address storkPublicKey,
        uint validTimePeriodSeconds,
        uint singleUpdateFeeInWei
    ) internal {
        StorkSetters.setValidTimePeriodSeconds(validTimePeriodSeconds);
        StorkSetters.setSingleUpdateFeeInWei(singleUpdateFeeInWei);
        StorkSetters.setStorkPublicKey(storkPublicKey);
    }

    function updateTemporalNumericValuesV1(
        StorkStructs.TemporalNumericValueInput[] calldata updateData
    ) public payable {
        uint16 numUpdates = 0;
        for (uint i = 0; i < updateData.length; i++) {
            bool verified = verifyStorkSignatureV1(
                storkPublicKey(),
                updateData[i].id,
                updateData[i].temporalNumericValue.timestampNs,
                updateData[i].temporalNumericValue.quantizedValue,
                updateData[i].publisherMerkleRoot,
                updateData[i].valueComputeAlgHash,
                updateData[i].r,
                updateData[i].s,
                updateData[i].v
            );
            if (!verified) revert StorkErrors.InvalidSignature();
            bool updated = updateLatestValueIfNecessary(updateData[i]);
            if (updated) {
                numUpdates++;
            }
        }
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

        if (block.timestamp - (numericValue.timestampNs / 1000000000) > validTimePeriodSeconds()) {
            revert StorkErrors.StaleValue();
        }
        return numericValue;
    }

    function version() public pure returns (string memory) {
        return "1.0.0";
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
}
