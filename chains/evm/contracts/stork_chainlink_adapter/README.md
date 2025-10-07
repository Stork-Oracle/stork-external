# Stork Chainlink Adapter

This is the Stork Chainlink Adapter for EVM-compatible chains. This package is maintained by [Stork Labs](https://stork.network).

It is available on [npm](https://www.npmjs.com/package/@storknetwork/stork_chainlink_adapter).

This package can be used as an SDK to build contracts that interact with Stork price feeds using Chainlink's familiar AggregatorV3Interface, or deployed as a standalone contract.

## Chainlink Compatibility

The adapter implements Chainlink's [AggregatorV3Interface](https://github.com/smartcontractkit/chainlink-evm/blob/develop/contracts/src/v0.8/shared/interfaces/AggregatorV3Interface.sol), allowing existing Chainlink-integrated applications to seamlessly integrate withStork price feeds with minimal code changes.

**Note: All timestamps are returned in nanoseconds (not seconds) to maintain Stork's high precision.**

## Usage as SDK

Install the package in your Solidity project:

```bash
npm install @storknetwork/stork_chainlink_adapter
```

Import and use in your contract:

```solidity
// SPDX-License-Identifier: Apache-2.0
pragma solidity >=0.8.24 <0.9.0;

import "@storknetwork/stork_chainlink_adapter/contracts/StorkChainlinkAdapter.sol";

contract YourContract {
    StorkChainlinkAdapter public priceAdapter;
    
    constructor(address storkContract, bytes32 priceId) {
        priceAdapter = new StorkChainlinkAdapter(storkContract, priceId);
    }
    
    function getLatestPrice() external view returns (int256) {
        (, int256 price, , , ) = priceAdapter.latestRoundData();
        return price;
    }
}
```

A complete working example can be found in the [stork-external repository](https://github.com/stork-oracle/stork-external/tree/main/chains/evm/examples/stork_chainlink_adapter).


## Deploying

To deploy the contract, clone down the repo and run the following commands from this contract's directory:

1. `npm install`
2. `npx hardhat compile`
3. Set the `storkContract` and `encodedAssetId` in the `ignition/parameters.json` file. Stork contract addresses and encoded asset ids can be found in the [Stork Documentation](https://docs.stork.network/). 
4. Deploy the contract with `npx hardhat --network <network> ignition deploy ignition/modules/StorkChainlinkAdapter.ts --deployment-id chain-<chainId>-<assetId> --parameters ignition/parameters.json`
5. Verify the contract on etherscan with `npx hardhat ignition verify chain-<chainId>-<assetId>`
