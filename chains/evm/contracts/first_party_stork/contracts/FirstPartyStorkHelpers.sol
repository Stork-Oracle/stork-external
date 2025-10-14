// contracts/stork/StorkHelpers.sol
// SPDX-License-Identifier: Apache 2

pragma solidity >=0.8.24 <0.9.0;

import "@storknetwork/first-party-stork-evm-sdk/IFirstPartyStorkHelpers.sol";

contract FirstPartyStorkHelpers is IFirstPartyStorkHelpers {
    function getEncodedAssetId(
        string memory assetPairId
    ) public pure returns (bytes32) {
        return keccak256(abi.encodePacked(assetPairId));
    }
}

