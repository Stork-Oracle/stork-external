// SPDX-License-Identifier: Apache 2

pragma solidity >=0.8.24 <0.9.0;

import "./SelfServeStorkEvents.sol";
import "./SelfServeStorkStructs.sol";
import "./SelfServeStorkErrors.sol";

import "./SelfServeStorkGetters.sol";
import "./SelfServeStorkSetters.sol";
import "./SelfServeStorkVerify.sol";


abstract contract SelfServeStork is
    SelfServeStorkGetters,
    SelfServeStorkSetters,
    SelfServeStorkVerify
{
    function updateTemporalNumericValues(
        SelfServeStorkStructs.PublisherTemporalNumericValueInput[]
            calldata updateData,
        bool storeHistoric
    ) public payable {
        uint16 numUpdates = 0;
        uint256 requiredFee = 0;
        for (uint i = 0; i < updateData.length; i++) {
            SelfServeStorkStructs.PublisherTemporalNumericValueInput
                memory input = updateData[i];
            SelfServeStorkStructs.PublisherUser
                memory publisherUser = getPublisherUser(input.pubKey);

            bool verified = verifyPublisherSignatureV1(
                input.pubKey,
                input.assetPairId,
                input.temporalNumericValue.timestampNs,
                input.temporalNumericValue.quantizedValue,
                input.r,
                input.s,
                input.v
            );

            if (!verified) revert SelfServeStorkErrors.InvalidSignature();

            bool updated = updateLatestValueIfNecessary(input.pubKey, input);
            if (updated) {
                numUpdates++;
            }

            if (storeHistoric) {
                storeHistoricalValue(input.pubKey, input);
            }
            requiredFee += publisherUser.singleUpdateFee;
        }

        if (numUpdates == 0) {
            revert SelfServeStorkErrors.NoFreshUpdate();
        }

        if (msg.value < requiredFee)
            revert SelfServeStorkErrors.InsufficientFee();
    }

    function createPublisherUser(
        address pubKey,
        uint256 singleUpdateFee
    ) public virtual;

    function deletePublisherUser(address pubKey) public virtual;
}
