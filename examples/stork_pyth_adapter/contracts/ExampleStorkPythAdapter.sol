// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import "@pythnetwork/pyth-sdk-solidity/PythStructs.sol";
import "@storknetwork/stork_pyth_adapter/contracts/StorkPythAdapter.sol";

contract ExampleStorkPythAdapter {
    StorkPythAdapter private storkPythAdapter;

    constructor(address storkContract) {
        storkPythAdapter = new StorkPythAdapter(storkContract);
    }

    function latestPrice(bytes32 priceId) public view virtual returns (int64 price, int32 exponent, uint publishTime) {
        PythStructs.Price memory priceStruct = storkPythAdapter.getPriceUnsafe(priceId);
        return (priceStruct.price, priceStruct.expo, priceStruct.publishTime);
    }
}