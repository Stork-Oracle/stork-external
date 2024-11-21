# Stork Pyth Adapter
This contract is a light wrapper around the [Stork EVM contract](../evm) which conforms to Pyth's [IPyth interface](https://github.com/pyth-network/pyth-sdk-solidity/blob/main/IPyth.sol).


## Integrate with Your Solidity Contracts
1. Install the Stork Pyth Adapter npm package in your project
```
npm i @storknetwork/stork_pyth_adapter
```

2. Import the Stork Pyth Adapter contract into your solidity contract using:
```
import "@storknetwork/stork_pyth_adapter/contracts/StorkPythAdapter.sol";
```

3. Create a StorkPythAdapter contract. This object takes in the contract address of Stork's contract on this chain:
```
storkPythAdapter = new StorkPythAdapter(storkContract);
```

You can see a simple working example of a Solidity contract using this [here](../../examples/stork_pyth_adapter).


## Publish to npm
1. update the package.json version
2. `npm login`
3. `npm publish --access public`
