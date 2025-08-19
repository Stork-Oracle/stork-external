// SPDX-License-Identifier: Apache-2.0

pragma solidity >=0.8.24 <0.9.0;

import "@pythnetwork/pyth-sdk-solidity/IPyth.sol";
import "@pythnetwork/pyth-sdk-solidity/PythStructs.sol";
import "@storknetwork/stork-evm-sdk/IStork.sol";
import "@storknetwork/stork-evm-sdk/StorkStructs.sol";
import "@storknetwork/stork-evm-sdk/IStorkGetters.sol";

/**
 * @title A port of the IPyth interface that supports Stork price feeds
 */
contract StorkPythAdapter is IPyth {
    IStork public stork;
    IStorkGetters public storkGetters;
    int32 private exponent = -18;
    uint64 private confidenceInterval = 0;

    constructor(address _stork) {
        stork = IStork(_stork);
        storkGetters = IStorkGetters(_stork);  // Same address, different interface
    }

    function convertInt192ToInt64Prescise(int192 value) public view returns (int64 val  , int32 exp) {
        // Use maximum precision
        int32 exp_shift = 0;
        while (value > type(int64).max || value < type(int64).min) {
            value = value / 10;
            exp_shift++;
        }
        require(value >= type(int64).min && value <= type(int64).max, "Overflow");
        return (int64(value), exponent + exp_shift);
    }

    /// @notice Returns the period (in seconds) that a price feed is considered valid since its publish time
    function getValidTimePeriod() external view returns (uint validTimePeriod) {
        return storkGetters.validTimePeriodSeconds();
    }

    /// @notice Returns the price.
    /// @dev Note that confidence intervals are not supported
    /// @return price - please read the documentation of PythStructs.Price to understand how to use this safely.
    function getPrice(
        bytes32 id
    ) external view returns (PythStructs.Price memory price) {
        StorkStructs.TemporalNumericValue memory temporalNumericValue = stork.getTemporalNumericValueV1(id);
        (int64 val, int32 exp) = convertInt192ToInt64Prescise(temporalNumericValue.quantizedValue);
        return PythStructs.Price(
            val,
            confidenceInterval,
            exp,
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
        (int64 val, int32 exp) = convertInt192ToInt64Prescise(temporalNumericValue.quantizedValue);
        return PythStructs.Price(
            val,
            confidenceInterval,
            exp,
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

        (int64 val, int32 exp) = convertInt192ToInt64Prescise(temporalNumericValue.quantizedValue);
        return PythStructs.Price(
            val,
            confidenceInterval,
            exp,
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
    function getTwapUpdateFee(
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

    /// @dev Updates not supported - this contract is read-only
    function parsePriceFeedUpdatesWithConfig(
        bytes[] calldata updateData,
        bytes32[] calldata priceIds,
        uint64 minAllowedPublishTime,
        uint64 maxAllowedPublishTime,
        bool checkUniqueness,
        bool checkUpdateDataIsMinimal,
        bool storeUpdatesIfFresh
    ) external payable returns ( PythStructs.PriceFeed[] memory priceFeeds, uint64[] memory slots) {
        revert("Not supported");
    }

    /// @dev Updates not supported - this contract is read-only
    function parseTwapPriceFeedUpdates(
        bytes[] calldata updateData,
        bytes32[] calldata priceIds
    ) external payable returns (PythStructs.TwapPriceFeed[] memory twapPriceFeeds) {
        revert("Not supported");
    }
    
}
