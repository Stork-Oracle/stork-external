// SPDX-License-Identifier: Apache-2.0
pragma solidity >=0.8.24 <0.9.0;

import "./StorkStructs.sol";


interface IStorkTemporalNumericValueUnsafeGetter {
    function getTemporalNumericValueUnsafeV1(
        bytes32 id
    ) external view returns (StorkStructs.TemporalNumericValue memory value);
}