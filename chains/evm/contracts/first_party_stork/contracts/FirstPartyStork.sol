// SPDX-License-Identifier: Apache 2

pragma solidity >=0.8.24 <0.9.0;

import "@storknetwork/first-party-stork-evm-sdk/IFirstPartyStorkEvents.sol";
import "@storknetwork/first-party-stork-evm-sdk/FirstPartyStorkStructs.sol";
import "@storknetwork/first-party-stork-evm-sdk/FirstPartyStorkErrors.sol";

import "./FirstPartyStorkGetters.sol";
import "./FirstPartyStorkSetters.sol";
import "./FirstPartyStorkVerify.sol";


abstract contract FirstPartyStork is
    FirstPartyStorkGetters,
    FirstPartyStorkSetters,
    FirstPartyStorkVerify
{
    function updateTemporalNumericValues(
        FirstPartyStorkStructs.PublisherTemporalNumericValueInput[]
            calldata updateData,
        bool[] calldata storeHistoric
    ) public payable {
        uint16 numUpdates = 0;
        uint256 requiredFee = 0;
        for (uint i = 0; i < updateData.length; i++) {
            FirstPartyStorkStructs.PublisherTemporalNumericValueInput
                memory input = updateData[i];
            FirstPartyStorkStructs.PublisherUser
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

            if (!verified) revert FirstPartyStorkErrors.InvalidSignature();

            bool updated = updateLatestValueIfNecessary(input.pubKey, input);
            if (updated) {
                numUpdates++;
            }

            if (storeHistoric[i]) {
                storeHistoricalValue(input.pubKey, input);
            }
            requiredFee += publisherUser.singleUpdateFee;
        }

        if (numUpdates == 0) {
            revert FirstPartyStorkErrors.NoFreshUpdate();
        }

        if (msg.value < requiredFee)
            revert FirstPartyStorkErrors.InsufficientFee();
    }

    function createPublisherUser(
        address pubKey,
        uint256 singleUpdateFee
    ) public virtual;

    function deletePublisherUser(address pubKey) public virtual;
}
