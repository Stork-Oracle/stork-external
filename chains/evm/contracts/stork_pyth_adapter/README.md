# Stork Pyth Adapter

This is the Stork Pyth Adapter for EVM-compatible chains. This package is maintained by [Stork Labs](https://stork.network).

It is available on [npm](https://www.npmjs.com/package/@storknetwork/stork_pyth_adapter).

This package can be used as an SDK to build contracts that interact with Stork price feeds using Pyth's familiar IPyth interface, or deployed as a standalone contract.

## Pyth Compatibility

The adapter implements Pyth's [IPyth interface](https://github.com/pyth-network/pyth-crosschain/blob/main/target_chains/ethereum/sdk/solidity/IPyth.sol), allowing existing Pyth-integrated applications to seamlessly integrate with Stork price feeds with minimal code changes. Price precision (exponent) is dynamically reduced from Stork's int192 values to fit within the less precise int64 used in the IPyth interface.

**Note that all timestamps are returned in seconds (converted from Stork's nanosecond precision) and confidence intervals are not supported in this adapter.**

## Usage as SDK

Install the package in your Solidity project:

```bash
npm install @storknetwork/stork_pyth_adapter
```

Import and use in your contract:

```solidity
// SPDX-License-Identifier: Apache-2.0
pragma solidity >=0.8.24 <0.9.0;

import "@storknetwork/stork_pyth_adapter/contracts/StorkPythAdapter.sol";

contract YourContract {
    StorkPythAdapter public priceAdapter;
    
    constructor(address storkContract) {
        priceAdapter = new StorkPythAdapter(storkContract);
    }
    
    function getLatestPrice(bytes32 priceId) external view returns (int64) {
        PythStructs.Price memory price = priceAdapter.getPrice(priceId);
        return price.price;
    }
}
```

## Supported Features

This adapter supports the following IPyth interface methods:
- `getValidTimePeriod()` - Returns the validity period for price feeds
- `getPrice(bytes32 id)` - Returns the latest price for a given asset ID
- `getPriceUnsafe(bytes32 id)` - Returns price without validation checks
- `getPriceNoOlderThan(bytes32 id, uint age)` - Returns price with age validation

## Unsupported Features

The following IPyth methods are not supported and will revert:
- EMA price methods (`getEmaPrice`, `getEmaPriceUnsafe`, `getEmaPriceNoOlderThan`)
- Update methods (`updatePriceFeeds`, `updatePriceFeedsIfNecessary`, `parsePriceFeedUpdates`, etc.)
- Fee calculation methods (`getUpdateFee`, `getTwapUpdateFee`)
A complete working example can be found in the [stork-external repository](https://github.com/stork-oracle/stork-external/tree/main/chains/evm/examples/stork_pyth_adapter).

## Deploying

To deploy the contract, clone down the repo and run the following commands from this contract's directory:

1. `npm install`
2. `npx hardhat compile`
3. Set the `storkContract` in the `ignition/parameters.json` file. Stork contract addresses can be found in the [Stork Documentation](https://docs.stork.network/).
4. `npx hardhat ignition deploy ignition/modules/StorkPythAdapter.ts --network <network>`

