// SPDX-License-Identifier: Apache 2

pragma solidity >=0.8.24 <0.9.0;

import "@pythnetwork/pyth-sdk-solidity/IPyth.sol";
import "@pythnetwork/pyth-sdk-solidity/PythStructs.sol";

/**
 * @title A port of the IPyth interface that supports Stork price feeds
 */
contract StorkPythAdapter is IPyth {
    IStork public stork;
    int32 private exponent = -18;
    uint64 private confidenceInterval = 0;

    constructor(address _stork) {
        stork = IStork(_stork);
    }

    function convertInt192ToInt64(int192 value) public pure returns (int64) {
        require(value >= type(int64).min && value <= type(int64).max, "Overflow");
        return int64(value);
    }

    /// @notice Returns the period (in seconds) that a price feed is considered valid since its publish time
    function getValidTimePeriod() external view returns (uint validTimePeriod) {
        return stork.validTimePeriodSeconds();
    }

    /// @notice Returns the price.
    /// @dev Note that confidence intervals are not supported
    /// @return price - please read the documentation of PythStructs.Price to understand how to use this safely.
    function getPrice(
        bytes32 id
    ) external view returns (PythStructs.Price memory price) {
        StorkStructs.TemporalNumericValue memory temporalNumericValue = stork.getTemporalNumericValueV1(id);
        return PythStructs.Price(
            convertInt192ToInt64(temporalNumericValue.quantizedValue),
            confidenceInterval,
            exponent,
            temporalNumericValue.timestampNs / 1000000000
        );
    }

    /// @notice Returns the price of a price feed without any sanity checks.
    /// @dev Note that confidence intervals are not supported
    /// @return price - please read the documentation of PythStructs.Price to understand how to use this safely.
    function getPriceUnsafe(
        bytes32 id
    ) external view returns (PythStructs.Price memory price) {
        StorkStructs.TemporalNumericValue memory temporalNumericValue = stork.getTemporalNumericValueUnsafeV1(id);
        return PythStructs.Price(
            convertInt192ToInt64(temporalNumericValue.quantizedValue),
            confidenceInterval,
            exponent,
            temporalNumericValue.timestampNs / 1000000000
        );
    }

    /// @notice Returns the price that is no older than `age` seconds of the current time.
    /// @dev Note that confidence intervals are not supported
    /// @return price - please read the documentation of PythStructs.Price to understand how to use this safely.
    function getPriceNoOlderThan(
        bytes32 id,
        uint age
    ) external view returns (PythStructs.Price memory price) {
        StorkStructs.TemporalNumericValue memory temporalNumericValue = stork.getTemporalNumericValueUnsafeV1(id);

        if (block.timestamp - (temporalNumericValue.timestampNs / 1000000000) > age) {
            revert("Value is stale");
        }

        return PythStructs.Price(
            convertInt192ToInt64(temporalNumericValue.quantizedValue),
            confidenceInterval,
            exponent,
            temporalNumericValue.timestampNs / 1000000000
        );
    }

    /// @dev EMA price not supported
    function getEmaPrice(
        bytes32 id
    ) external view returns (PythStructs.Price memory price) {
        revert("Not supported");
    }


    /// @dev EMA price not supported
    function getEmaPriceUnsafe(
        bytes32 id
    ) external view returns (PythStructs.Price memory price) {
        revert("Not supported");
    }

    /// @dev EMA price not supported
    function getEmaPriceNoOlderThan(
        bytes32 id,
        uint age
    ) external view returns (PythStructs.Price memory price) {
        revert("Not supported");
    }

    /// @dev Updates not supported - this contract is read-only
    function updatePriceFeeds(bytes[] calldata updateData) external payable {
        revert("Not supported");
    }

    /// @dev Updates not supported - this contract is read-only
    function updatePriceFeedsIfNecessary(
        bytes[] calldata updateData,
        bytes32[] calldata priceIds,
        uint64[] calldata publishTimes
    ) external payable {
        revert("Not supported");
    }

    /// @dev Updates not supported - this contract is read-only
    function getUpdateFee(
        bytes[] calldata updateData
    ) external view returns (uint feeAmount) {
        revert("Not supported");
    }

    /// @dev Updates not supported - this contract is read-only
    function parsePriceFeedUpdates(
        bytes[] calldata updateData,
        bytes32[] calldata priceIds,
        uint64 minPublishTime,
        uint64 maxPublishTime
    ) external payable returns (PythStructs.PriceFeed[] memory priceFeeds) {
        revert("Not supported");
    }

    /// @dev Updates not supported - this contract is read-only
    function parsePriceFeedUpdatesUnique(
        bytes[] calldata updateData,
        bytes32[] calldata priceIds,
        uint64 minPublishTime,
        uint64 maxPublishTime
    ) external payable returns (PythStructs.PriceFeed[] memory priceFeeds) {
        revert("Not supported");
    }
}


interface IStork {
    function getTemporalNumericValueV1(
        bytes32 id
    ) external view returns (StorkStructs.TemporalNumericValue memory value);

    function getTemporalNumericValueUnsafeV1(
        bytes32 id
    ) external view returns (StorkStructs.TemporalNumericValue memory value);

    function validTimePeriodSeconds() external view returns (uint);
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
