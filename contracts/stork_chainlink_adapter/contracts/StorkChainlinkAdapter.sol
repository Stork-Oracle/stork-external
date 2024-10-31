// SPDX-License-Identifier: Apache 2

pragma solidity >=0.8.24 <0.9.0;

/**
 * @title A port of the ChainlinkAggregatorV3 interface that supports Stork price feeds
 */
contract StorkChainlinkAdapter {
    bytes32 public priceId;
    IStorkTemporalNumericValueUnsafeGetter public stork;

    constructor(address _stork, bytes32 _priceId) {
        priceId = _priceId;
        stork = IStorkTemporalNumericValueUnsafeGetter(_stork);
    }

    function decimals() external pure returns (uint8) {
        return 18;
    }

    function description() public pure returns (string memory) {
        return "A port of a chainlink aggregator powered by Stork";
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

    /*
    * @notice This is exactly the same as `latestRoundData`, just including for parity with Chainlink
    * Stork doesn't store roundId on chain so there's no way to access old data by round id
    */
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

interface IStorkTemporalNumericValueUnsafeGetter {
    function getTemporalNumericValueUnsafeV1(
        bytes32 id
    ) external view returns (StorkStructs.TemporalNumericValue memory value);
}

contract StorkStructs {
    struct TemporalNumericValue {
        // slot 1
        // nanosecond level precision timestamp of latest publisher update in batch
        uint64 timestampNs; // 8 bytes
        // should be able to hold all necessary numbers (up to 6277101735386680763835789423207666416102355444464034512895)
        int192 quantizedValue; // 8 bytes
    }
}
