// SPDX-License-Identifier: Apache-2.0
pragma solidity >=0.8.24 <0.9.0;

library SelfServeStorkUtils {
    function getAssetId(string memory assetPairId) public pure returns (bytes32) {
        return keccak256(abi.encodePacked(assetPairId));
    }
}
