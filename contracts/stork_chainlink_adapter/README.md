# Stork Chainlink Adapter
This contract is a light wrapper around the [Stork EVM contract](../evm) which conforms to Chainlink's [AggregatorV3Interface](https://github.com/smartcontractkit/chainlink/blob/develop/contracts/src/v0.8/shared/interfaces/AggregatorV3Interface.sol).


## Integrate with Your Solidity Contracts
1. Install the Stork Chainlink Adapter npm package in your project
```
npm i @storknetwork/stork_chainlink_adapter
```

2. Import the Stork Chainlink Adapter contract into your solidity contract using:
```
import "@storknetwork/stork_chainlink_adapter/contracts/StorkChainlinkAdapter.sol";
```

3. Create one StorkChainlinkAdapter for each asset whose price you want to track. This object takes in the contract address of Stork's contract on this chain, and the bytes32-formatted price id for this asset:
```
storkChainlinkAdapter = new StorkChainlinkAdapter(storkContract, priceId);
```

You can see a simple working example of a Solidity contract using this [here](../../examples/stork_chainlink_adapter).


## Publish to npm
1. update the package.json version
2. `npm login`
3. `npm publish --access public`
