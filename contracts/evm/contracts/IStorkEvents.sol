// SPDX-License-Identifier: Apache-2.0
pragma solidity >=0.8.24 <0.9.0;

/// @title IStorkEvents contains the events that Stork contract emits.
/// @dev This interface can be used for listening to the updates for off-chain and testing purposes.
interface IStorkEvents {
    /// @dev Emitted when the latest value with `id` has received a fresh update.
    /// @param id The Stork Feed ID.
    /// @param timestampNs Publish time of the given update.
    /// @param quantizedValue Value of the given update.
    event ValueUpdate(
        bytes32 indexed id,
        uint64 timestampNs,
        int192 quantizedValue
    );
}
