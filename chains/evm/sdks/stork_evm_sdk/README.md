# Stork EVM SDK

This is a Solidity SDK to build EVM contracts that consume Stork price feeds. This package is maintained by [Stork Labs](https://stork.network).

It is available on [npm](https://www.npmjs.com/package/@storknetwork/stork-evm-sdk).

## Pull Model

The Stork EVM SDK allows users to consume Stork price updates on a pull basis. This puts the responsibility of submitting the price updates on-chain to the user whenever they want to interact with an app that consumes Stork price feeds. Stork Labs maintains a [Chain Pusher](https://github.com/stork-oracle/stork-external/tree/main/apps/chain_pusher/README.md) in order to do this.

## Details

The Stork EVM SDK provides a set of useful interfaces and structures for building EVM contracts that consume Stork price feeds. Primarily, a consuming contract will be using:

- The `IStork` interface for interacting with the Stork oracle contract
- The `getTemporalNumericValueV1` function to get the latest price update with staleness check
- The `getTemporalNumericValueUnsafeV1` function to get the latest price update without staleness check
- The `TemporalNumericValue` struct which represents a price update

## Example

The following snippet is an example of how to use this SDK to consume Stork price feeds on chain. A full example is available [here](https://github.com/stork-oracle/stork-external/tree/main/chains/evm/examples/stork).

```solidity
// SPDX-License-Identifier: Apache-2.0
pragma solidity >=0.8.24 <0.9.0;

import "@stork-network/stork-evm-sdk/IStork.sol";
import "@stork-network/stork-evm-sdk/StorkStructs.sol";

contract YourContract {
    IStork public stork;
    
    constructor(address _stork) {
        stork = IStork(_stork);
    }
    
    // This function reads the latest price from a Stork feed
    function getPrice(bytes32 feedId) public view returns (int192 price, uint64 timestamp) {
        StorkStructs.TemporalNumericValue memory value = stork.getTemporalNumericValueV1(feedId);
        return (value.quantizedValue, value.timestampNs);
    }
}
