// SPDX-License-Identifier: Apache 2

pragma solidity >=0.8.24 <0.9.0;

import {Stork} from "./Stork.sol";
import "./StorkStructs.sol";
import {UpgradeableStork} from "./UpgradeableStork.sol";

// todo harry: should we be explicit about forking here

// This interface is forked from the Zerolend Adapter found here:
// https://github.com/zerolend/pyth-oracles/blob/master/contracts/PythAggregatorV3.sol
// Original license found under licenses/zerolend-pyth-oracles.md

/**
 * @title A port of the ChainlinkAggregatorV3 interface that supports Stork price feeds
 * @notice This does not store any roundId information on-chain. Please review the code before using this implementation.
 * Users should deploy an instance of this contract to wrap every price feed id that they need to use.
 */
contract StorkChainlinkAdapter {
    bytes32 public priceId;
    Stork public stork;

    constructor(address _stork, bytes32 _priceId) {
        priceId = _priceId;
        stork = Stork(_stork);
    }

    function decimals() external pure returns (uint8) {
        return 18;
    }

    function description() public pure returns (string memory) {
        return "A port of a chainlink aggregator powered by Stork network feeds";
    }

    function version() public pure returns (uint256) {
        return 1;
    }

    function latestAnswer() public view virtual returns (int256) {
        return stork.getTemporalNumericValueUnsafeV1(priceId).quantizedValue;
    }

    function latestTimestamp() public view returns (uint256) {
        return stork.getTemporalNumericValueUnsafeV1(priceId).timestampNs;
    }

    function latestRound() public view returns (uint256) {
        // use timestamp as the round id
        return latestTimestamp();
    }

    function getAnswer(uint256) public view returns (int256) {
        return latestAnswer();
    }

    function getTimestamp(uint256) external view returns (uint256) {
        return latestTimestamp();
    }

    function getRoundData(
        uint80 _roundId
    )
    external
    view
    returns (
        uint80 roundId,
        int256 answer,
        uint256 startedAt,
        uint256 updatedAt,
        uint80 answeredInRound
    )
    {
        StorkStructs.TemporalNumericValue memory value = stork.getTemporalNumericValueUnsafeV1(priceId);
        return (
            _roundId,
            value.quantizedValue,
            value.timestampNs,
            value.timestampNs,
            _roundId
        );
    }

    function latestRoundData()
    external
    view
    returns (
        uint80 roundId,
        int256 answer,
        uint256 startedAt,
        uint256 updatedAt,
        uint80 answeredInRound
    )
    {
        StorkStructs.TemporalNumericValue memory value = stork.getTemporalNumericValueUnsafeV1(priceId);
        roundId = uint80(value.timestampNs);
        return (
            roundId,
            value.quantizedValue,
            value.timestampNs,
            value.timestampNs,
            roundId
        );
    }
}