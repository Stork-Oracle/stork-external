// SPDX-License-Identifier: Apache-2.0
pragma solidity >=0.8.24 <0.9.0;

import "@storknetwork/stork-evm-sdk/IStork.sol";
import "@storknetwork/stork-evm-sdk/StorkStructs.sol";

contract Example {
    IStork public stork;
    
    event StorkPriceUsed(
        bytes32 indexed feedId, 
        int192 value, 
        uint64 timestamp
    );
    
    constructor(address _stork) {
        stork = IStork(_stork);
    }
    
    // This function reads the latest price from a Stork feed
    function useStorkPrice(bytes32 feedId) public returns (int192 value, uint64 timestamp) {
        StorkStructs.TemporalNumericValue memory temporalValue = stork.getTemporalNumericValueV1(feedId);
        
        value = temporalValue.quantizedValue;
        timestamp = temporalValue.timestampNs;
        
        // Emit an event with the price data
        emit StorkPriceUsed(feedId, value, timestamp);
        
        return (value, timestamp);
    }
    
    // This function reads the latest price without staleness check
    function useStorkPriceUnsafe(bytes32 feedId) public view returns (int192 value, uint64 timestamp) {
        StorkStructs.TemporalNumericValue memory temporalValue = stork.getTemporalNumericValueUnsafeV1(feedId);
        
        value = temporalValue.quantizedValue;
        timestamp = temporalValue.timestampNs;
        
        return (value, timestamp);
    }
}
