// SPDX-License-Identifier: BSD-3-Clause
pragma solidity ^0.8.20;

import "./PriceOracle.sol";
import "./PErc20.sol";
import {AggregatorV3Interface} from "@chainlink/contracts/src/v0.8/shared/interfaces/AggregatorV3Interface.sol";
import "@storknetwork/stork-evm-sdk/IStork.sol";
import "@storknetwork/stork-evm-sdk/StorkStructs.sol";

/**
 * @title HybridStorkPriceOracle
 * @notice Price oracle that can read from both Chainlink feeds and Stork oracle adapters.
 * @dev Stork exposes Chainlink-compatible adapter contracts, so we interact with them using the
 *      same AggregatorV3Interface interface. The oracle tries the preferred feed first (Stork or
 *      Chainlink) and falls back to the other feed before reverting to a manually set price.
 */
contract HybridStorkPriceOracle is PriceOracle {
    struct FeedConfig {
        bool preferStork;
    }

    mapping(address => uint256) private manualPrices;
    mapping(address => bool) public admin;

    mapping(address => AggregatorV3Interface) public chainlinkFeeds;
    mapping(address => bytes32) public storkFeeds;

    mapping(address => uint256) public lastValidChainlinkPrice;
    mapping(address => uint256) public lastValidStorkPrice;

    mapping(address => FeedConfig) public feedConfig;

    uint256 public chainlinkStaleThreshold;
    uint256 public storkStaleThreshold;
    address public owner;
    IStork public storkContract;

    event ManualPricePosted(address indexed asset, uint256 previousPriceMantissa, uint256 newPriceMantissa);
    event ChainlinkFeedRegistered(address indexed asset, address indexed aggregator);
    event StorkFeedRegistered(address indexed asset, bytes32 indexed storkEncodedAssetID);
    event LastChainlinkPriceUpdated(address indexed asset, uint256 priceMantissa);
    event LastStorkPriceUpdated(address indexed asset, uint256 priceMantissa);
    event FeedPreferenceUpdated(address indexed asset, bool preferStork);

    modifier onlyAdmin() {
        require(admin[msg.sender], "HybridOracle: only admin");
        _;
    }

    modifier onlyOwner() {
        require(msg.sender == owner, "HybridOracle: only owner");
        _;
    }

    constructor(uint256 _chainlinkStaleThreshold, address _storkContractAddress, uint256 _storkStaleThresholdSeconds) {
        owner = msg.sender;
        admin[msg.sender] = true;
        chainlinkStaleThreshold = _chainlinkStaleThreshold;
        storkContract = IStork(_storkContractAddress);
        storkStaleThreshold = _storkStaleThresholdSeconds;
    }

    /*//////////////////////////////////////////////////////////////
                        CORE PRICE RETRIEVAL
    //////////////////////////////////////////////////////////////*/

    function getUnderlyingPrice(PToken pToken) public view override returns (uint256) {
        address asset = _getUnderlyingAddress(pToken);
        FeedConfig memory config = feedConfig[asset];

        uint256 price = config.preferStork ? _getStorkPrice(asset) : _getChainlinkPrice(asset);
        if (price == 0) {
            price = config.preferStork ? _getChainlinkPrice(asset) : _getStorkPrice(asset);
        }

        if (price == 0) {
            return manualPrices[asset];
        }

        return price;
    }

    function _getUnderlyingAddress(PToken pToken) internal view returns (address) {
        if (_compareStrings(pToken.symbol(), "pETH")) {
            return 0xEeeeeEeeeEeEeeEeEeEeeEEEeeeeEeeeeeeeEEeE;
        }
        return address(PErc20(address(pToken)).underlying());
    }

    function _getStorkPrice(address asset) internal view returns (uint256) {
        bytes32 storkEncodedAssetID = storkFeeds[asset];
        if (storkEncodedAssetID == bytes32(0)) {
            return 0;
        }
        try storkContract.getTemporalNumericValueUnsafeV1(storkEncodedAssetID) returns (StorkStructs.TemporalNumericValue memory value) {
            if (value.timestampNs / 1000000000 < block.timestamp - storkStaleThreshold) { // check if the price is stale by converting nanoseconds to seconds
                return 0;
            }
            if (value.quantizedValue < 0) { // negative values are not allowed
                return 0;
            }
            return value.quantizedValue; // quantized value is already multiplied by 10^18 to get the correct 18 decimals
        } catch {
        }
        
        return lastValidStorkPrice[storkEncodedAssetID];
    }

    function _getChainlinkPrice(address asset) internal view returns (uint256) {
        return _readFeed(asset, chainlinkFeeds[asset], chainlinkStaleThreshold, lastValidChainlinkPrice[asset]);
    }

    function _readFeed(
        address asset,
        AggregatorV3Interface aggregator,
        uint256 staleThreshold,
        uint256 cachedPrice
    ) internal view returns (uint256) {
        if (address(aggregator) == address(0)) {
            return 0;
        }

        try aggregator.latestRoundData() returns (
            uint80 /* roundId */,
            int256 price,
            uint256 /* startedAt */,
            uint256 updatedAt,
            uint80 /* answeredInRound */
        ) {
            if (price > 0 && block.timestamp - updatedAt <= staleThreshold) {
                uint256 priceMantissa = _scalePrice(uint256(price), aggregator.decimals());
                return priceMantissa;
            }
        } catch {
            // fall through
        }

        return cachedPrice;
    }

    function _scalePrice(uint256 price, uint8 decimals) internal pure returns (uint256) {
        if (decimals < 18) {
            return price * (10 ** (18 - decimals));
        } else if (decimals > 18) {
            return price / (10 ** (decimals - 18));
        }
        return price;
    }

    /*//////////////////////////////////////////////////////////////
                        FEED MANAGEMENT
    //////////////////////////////////////////////////////////////*/

    function setManualPrice(address asset, uint256 priceMantissa) external onlyAdmin {
        emit ManualPricePosted(asset, manualPrices[asset], priceMantissa);
        manualPrices[asset] = priceMantissa;
    }

    function registerChainlinkFeed(address asset, address aggregator) external onlyAdmin {
        require(aggregator != address(0), "HybridOracle: zero aggregator");
        chainlinkFeeds[asset] = AggregatorV3Interface(aggregator);
        emit ChainlinkFeedRegistered(asset, aggregator);
    }

    function removeChainlinkFeed(address asset) external onlyAdmin {
        delete chainlinkFeeds[asset];
        delete lastValidChainlinkPrice[asset];
        emit ChainlinkFeedRegistered(asset, address(0));
    }

    function registerStorkFeed(address asset, bytes32 storkEncodedAssetID) external onlyAdmin {
        require(storkEncodedAssetID != bytes32(0), "HybridOracle: zero stork encoded asset ID");
        storkFeeds[asset] = storkEncodedAssetID;
        emit StorkFeedRegistered(asset, storkEncodedAssetID);
    }

    function removeStorkFeed(address asset) external onlyAdmin {
        delete storkFeeds[asset];
        delete lastValidStorkPrice[asset];
        emit StorkFeedRegistered(asset, bytes32(0));
    }

    // would recommend removing as I dont think this added complexity adds much value
    function updateStorkCache(address[] calldata assets) external {
        for (uint256 i = 0; i < assets.length; i++) {
            address asset = assets[i];
            bytes32 storkEncodedAssetID = storkFeeds[asset];
            if (storkEncodedAssetID == bytes32(0)) continue;
            try storkContract.getTemporalNumericValueUnsafeV1(storkEncodedAssetID) returns (StorkStructs.TemporalNumericValue memory value) {
                if (value.timestampNs / 1000000000 < block.timestamp - storkStaleThreshold) { // check if the price is stale by converting nanoseconds to seconds
                    continue;
                }
                if (value.quantizedValue < 0) { // negative values are not allowed
                    continue;
                }
                lastValidStorkPrice[asset] = value.quantizedValue; // quantized value is already multiplied by 10^18 to get the correct 18 decimals
                emit LastStorkPriceUpdated(asset, value.quantizedValue);
            } catch {
                continue;
            }
        }
    }

    function updateChainlinkCache(address[] calldata assets) external {
        _updateFeedCache(assets, chainlinkFeeds, chainlinkStaleThreshold, lastValidChainlinkPrice, true);
    }

    function _updateFeedCache(
        address[] calldata assets,
        mapping(address => AggregatorV3Interface) storage feeds,
        uint256 staleThreshold,
        mapping(address => uint256) storage cache,
        bool isChainlink
    ) internal {
        for (uint256 i = 0; i < assets.length; i++) {
            address asset = assets[i];
            AggregatorV3Interface aggregator = feeds[asset];
            if (address(aggregator) == address(0)) continue;

            try aggregator.latestRoundData() returns (
                uint80 /* roundId */,
                int256 price,
                uint256 /* startedAt */,
                uint256 updatedAt,
                uint80 /* answeredInRound */
            ) {
                if (price > 0 && block.timestamp - updatedAt <= staleThreshold) {
                    uint256 priceMantissa = _scalePrice(uint256(price), aggregator.decimals());
                    cache[asset] = priceMantissa;
                    if (isChainlink) {
                        emit LastChainlinkPriceUpdated(asset, priceMantissa);
                    } 
                }
            } catch {
                continue;
            }
        }
    }

    function setFeedPreference(address asset, bool preferStork) external onlyAdmin {
        feedConfig[asset].preferStork = preferStork;
        emit FeedPreferenceUpdated(asset, preferStork);
    }

    function setChainlinkStaleThreshold(uint256 newThreshold) external onlyOwner {
        chainlinkStaleThreshold = newThreshold;
    }

    // this should never need to be called, but is here for completeness
    function setStorkContract(address newStorkContract) external onlyOwner {
        storkContract = IStork(newStorkContract);
    }

    function setStorkStaleThreshold(uint256 newThresholdSeconds) external onlyOwner {
        storkStaleThreshold = newThresholdSeconds;
    }

    /*//////////////////////////////////////////////////////////////
                             ACCESS CONTROL
    //////////////////////////////////////////////////////////////*/

    function setOwner(address newOwner) external onlyOwner {
        owner = newOwner;
    }

    function setAdmin(address newAdmin, bool enabled) external onlyOwner {
        admin[newAdmin] = enabled;
    }

    /*//////////////////////////////////////////////////////////////
                             UTILITIES
    //////////////////////////////////////////////////////////////*/

    function _compareStrings(string memory a, string memory b) internal pure returns (bool) {
        return (keccak256(bytes(a)) == keccak256(bytes(b)));
    }
}
