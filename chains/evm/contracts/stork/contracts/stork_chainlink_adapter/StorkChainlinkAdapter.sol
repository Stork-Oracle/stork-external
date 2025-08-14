// SPDX-License-Identifier: Apache-2.0
pragma solidity >=0.8.24 <0.9.0;

import "@storknetwork/stork-evm-sdk/StorkChainlinkAdapter.sol";

/**
 * @title Deployable Stork Chainlink Adapter
 * @notice Simple deployment wrapper for the Stork Chainlink Adapter from the SDK
 */
contract StorkChainlinkAdapter is StorkChainlinkAdapter {
    constructor(address _stork, bytes32 _priceId) StorkChainlinkAdapter(_stork, _priceId) {}
}
