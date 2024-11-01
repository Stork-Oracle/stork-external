// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import "@storknetwork/stork_chainlink_adapter/contracts/StorkChainlinkAdapter.sol";

contract ExampleStorkChainlinkAdapter {
    StorkChainlinkAdapter storkChainlinkAdapter;

    constructor(address storkContract, bytes32 priceId) {
        storkChainlinkAdapter = new StorkChainlinkAdapter(storkContract, priceId);
    }

    function latestRoundData() public view virtual returns (uint80 roundId, int256 answer, uint256 startedAt, uint256 updatedAt, uint80 answeredInRound) {
        return storkChainlinkAdapter.latestRoundData();
    }
}