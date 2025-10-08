# First Party Stork EVM SDK

This is a Solidity SDK to build EVM contracts that consume First Party Stork data. This package is maintained by [Stork Labs](https://stork.network).

It is available on [npm](https://www.npmjs.com/package/@storknetwork/first-party-stork-evm-sdk).

## Pull Model

The First Party Stork EVM SDK allows users to interact with First Party Stork oracle contracts. These contracts enable individual publishers to directly push their own numeric updates on-chain with optional historical data storage. Any user can consume these updates on a pull basis.

## Details

The First Party Stork EVM SDK provides a set of useful interfaces and structures for building EVM contracts that consume First Party Stork data. Primarily, a consuming contract will be using:

- The `IFirstPartyStork` interface for interacting with the First Party Stork oracle contract
- The `getLatestTemporalNumericValue` function to get the latest value update from a specific publisher
- The `getHistoricalTemporalNumericValue` function to get historical value data by round ID
- The `TemporalNumericValue` struct which represents a value update
- The `PublisherUser` struct which represents publisher configuration

## Example

### Reading Data

The following snippet shows how to read the latest value from a specific publisher on chain.

```solidity
// SPDX-License-Identifier: Apache-2.0
pragma solidity >=0.8.24 <0.9.0;

import "@storknetwork/first-party-stork-evm-sdk/IFirstPartyStork.sol";
import "@storknetwork/first-party-stork-evm-sdk/FirstPartyStorkStructs.sol";

contract YourContract {
    IFirstPartyStork public firstPartyStork;
    
    constructor(address _firstPartyStork) {
        firstPartyStork = IFirstPartyStork(_firstPartyStork);
    }
    
    // Read the latest value from a specific publisher
    function getPrice(address publisher, string memory assetPairId) 
        public view returns (int192 value, uint64 timestamp) {
        FirstPartyStorkStructs.TemporalNumericValue memory tnv = 
            firstPartyStork.getLatestTemporalNumericValue(publisher, assetPairId);
        return (tnv.quantizedValue, tnv.timestampNs);
    }
    
    // Read historical value data by round ID
    function getHistoricalPrice(address publisher, string memory assetPairId, uint256 roundId)
        public view returns (int192 value, uint64 timestamp) {
        FirstPartyStorkStructs.TemporalNumericValue memory tnv = 
            firstPartyStork.getHistoricalTemporalNumericValue(publisher, assetPairId, roundId);
        return (tnv.quantizedValue, tnv.timestampNs);
    }
}
```
