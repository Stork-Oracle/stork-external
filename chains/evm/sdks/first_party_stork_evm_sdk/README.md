# FirstParty Stork EVM SDK

This is a Solidity SDK to build EVM contracts that consume FirstParty Stork price feeds. This package is maintained by [Stork Labs](https://stork.network).

It is available on [npm](https://www.npmjs.com/package/@storknetwork/first-party-stork-evm-sdk).

## Overview

The FirstParty Stork EVM SDK allows developers to interact with FirstParty Stork oracle contracts, which enable publishers to directly submit price updates on-chain. Unlike the standard Stork oracle that uses centralized aggregation, FirstParty Stork allows individual publishers to submit their own price data with configurable fees and historical data storage.

## Key Features

- **Publisher-specific data**: Each publisher maintains their own price feeds
- **Historical data storage**: Optional storage of historical price updates with round IDs
- **Configurable fees**: Each publisher can have different update fees
- **Signature verification**: Built-in verification of publisher signatures
- **Event emissions**: Comprehensive event system for tracking updates

## Details

The FirstParty Stork EVM SDK provides a set of useful interfaces and structures for building EVM contracts that consume FirstParty Stork price feeds. Primarily, a consuming contract will be using:

- The `IFirstPartyStork` interface for interacting with the FirstParty Stork oracle contract
- The `getLatestTemporalNumericValue` function to get the latest price update from a specific publisher
- The `getHistoricalTemporalNumericValue` function to get historical price data by round ID
- The `TemporalNumericValue` struct which represents a price update
- The `PublisherUser` struct which represents publisher configuration

## Installation

Install the package in your Solidity project:

```bash
npm install @storknetwork/first-party-stork-evm-sdk
```

## Example Usage

### Basic Price Reading

The following snippet shows how to read the latest price from a specific publisher:

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
    
    // This function reads the latest price from a specific publisher
    function getPrice(address publisher, string memory assetPairId) 
        public view returns (int192 price, uint64 timestamp) {
        FirstPartyStorkStructs.TemporalNumericValue memory value = 
            firstPartyStork.getLatestTemporalNumericValue(publisher, assetPairId);
        return (value.quantizedValue, value.timestampNs);
    }
    
    // This function gets historical price data
    function getHistoricalPrice(address publisher, string memory assetPairId, uint256 roundId)
        public view returns (int192 price, uint64 timestamp) {
        FirstPartyStorkStructs.TemporalNumericValue memory value = 
            firstPartyStork.getHistoricalTemporalNumericValue(publisher, assetPairId, roundId);
        return (value.quantizedValue, value.timestampNs);
    }
}
```

### Publisher Management (Owner Functions)

```solidity
// SPDX-License-Identifier: Apache-2.0
pragma solidity >=0.8.24 <0.9.0;

import "@storknetwork/first-party-stork-evm-sdk/IFirstPartyStork.sol";
import "@storknetwork/first-party-stork-evm-sdk/FirstPartyStorkStructs.sol";

contract PublisherManager {
    IFirstPartyStork public firstPartyStork;
    
    constructor(address _firstPartyStork) {
        firstPartyStork = IFirstPartyStork(_firstPartyStork);
    }
    
    // Add a new publisher (requires owner permissions on the FirstParty Stork contract)
    function addPublisher(address publisherKey, uint256 updateFee) external {
        firstPartyStork.createPublisherUser(publisherKey, updateFee);
    }
    
    // Remove a publisher (requires owner permissions on the FirstParty Stork contract)
    function removePublisher(address publisherKey) external {
        firstPartyStork.deletePublisherUser(publisherKey);
    }
    
    // Get publisher configuration
    function getPublisherInfo(address publisherKey) 
        external view returns (address pubKey, uint256 singleUpdateFee) {
        FirstPartyStorkStructs.PublisherUser memory user = 
            firstPartyStork.getPublisherUser(publisherKey);
        return (user.pubKey, user.singleUpdateFee);
    }
}
```

### Price Update Submission

```solidity
// SPDX-License-Identifier: Apache-2.0
pragma solidity >=0.8.24 <0.9.0;

import "@storknetwork/first-party-stork-evm-sdk/IFirstPartyStork.sol";
import "@storknetwork/first-party-stork-evm-sdk/FirstPartyStorkStructs.sol";

contract PriceSubmitter {
    IFirstPartyStork public firstPartyStork;
    
    constructor(address _firstPartyStork) {
        firstPartyStork = IFirstPartyStork(_firstPartyStork);
    }
    
    // Submit a price update (requires proper signature and fee)
    function submitPriceUpdate(
        FirstPartyStorkStructs.PublisherTemporalNumericValueInput memory updateData,
        bool storeHistoric
    ) external payable {
        FirstPartyStorkStructs.PublisherTemporalNumericValueInput[] memory updates = 
            new FirstPartyStorkStructs.PublisherTemporalNumericValueInput[](1);
        updates[0] = updateData;
        
        bool[] memory storeFlags = new bool[](1);
        storeFlags[0] = storeHistoric;
        
        firstPartyStork.updateTemporalNumericValues{value: msg.value}(updates, storeFlags);
    }
}
```

## Key Differences from Standard Stork SDK

1. **Publisher-specific data**: Each publisher maintains separate price feeds
2. **Historical storage**: Optional round-based historical data storage
3. **Individual fees**: Each publisher can have different update fees
4. **Direct submission**: Publishers submit data directly without central aggregation

## Data Structures

### TemporalNumericValue

- `timestampNs`: Nanosecond precision timestamp
- `quantizedValue`: 192-bit signed integer for the price value

### PublisherTemporalNumericValueInput

- `temporalNumericValue`: The price data
- `pubKey`: Publisher's public key
- `assetPairId`: Asset pair identifier (e.g., "ETHUSD")
- `r`, `s`, `v`: Signature components

### PublisherUser

- `pubKey`: Publisher's public key
- `singleUpdateFee`: Fee required for each update

## Events

The SDK includes comprehensive event interfaces for monitoring:

- `ValueUpdate`: Latest price updates
- `HistoricalValueStored`: Historical data storage
- `PublisherUserAdded`: New publisher registration
- `PublisherUserRemoved`: Publisher removal

## Error Handling

The SDK includes error definitions for common failure cases:

- `InsufficientFee`: Provided fee is too low
- `NoFreshUpdate`: No updates are newer than existing data
- `NotFound`: Requested data doesn't exist
- `InvalidSignature`: Signature verification failed

## Complete Example

A complete working example can be found in the [stork-external repository](https://github.com/stork-oracle/stork-external/tree/main/chains/evm/examples/first_party_stork).
