// SPDX-License-Identifier: Apache-2.0
pragma solidity >=0.8.24 <0.9.0;

import "@storknetwork/stork-evm-sdk/StorkPythAdapter.sol";

/**
 * @title Deployable Stork Pyth Adapter
 * @notice Simple deployment wrapper for the Stork Pyth Adapter from the SDK
 */
contract StorkPythAdapter is StorkPythAdapter {
    constructor(address _stork) StorkPythAdapter(_stork) {}
}
