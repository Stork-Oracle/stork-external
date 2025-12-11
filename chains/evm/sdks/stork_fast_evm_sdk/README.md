# Stork Fast EVM SDK

This is a Solidity SDK for building EVM contracts that consume Stork Fast signed ECDSA update payloads. This package is maintained by [Stork Labs](https://stork.network).

It is available on [npm](https://www.npmjs.com/package/@storknetwork/stork-fast-evm-sdk).

## Verification and Deserialization

The Stork Fast EVM SDK provides a set of useful functions for verifying and deserializing Stork Fast signed ECDSA update payloads. These exist partially as library functions available directly in the `StorkFastDeserialize` library, and partially as contract functions available via the `IStorkFast` interface.

### StorkFastDeserialize Library Functions

The `StorkFastDeserialize` library provides the following functions for deserializing Stork Fast signed ECDSA update payloads:
- `splitSignedECDSAPayload` - Splits a signed ECDSA payload into a signature bytes and a verifiable payload bytes
- `deserializeSignedECDSAPayloadHeader` - Deserializes the header of a signed ECDSA payload
- `deserializeAssetsFromSignedECDSAPayload` - Deserializes the assets from a signed ECDSA payload

### IStorkFast Contract Functions

The `IStorkFast` interface provides a set of useful functions that can be called on an instance of the `StorkFast` contract. These include:
- `verifySignedECDSAPayload` - Verifies a signed ECDSA payload
- `verifyAndDeserializeSignedECDSAPayload` - Verifies and deserializes a signed ECDSA payload

## Example

The following snippet is an example of how to use this SDK to consume Stork Fast signed ECDSA update payloads on-chain. A full example is available [here](https://github.com/stork-oracle/stork-external/tree/main/chains/evm/examples/stork_fast).

```solidity
// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.28;

import "@storknetwork/stork-fast-evm-sdk/IStorkFast.sol";
import "@storknetwork/stork-fast-evm-sdk/StorkFastStructs.sol";

contract YourContract {
    IStorkFast public storkFast;

    constructor(address _storkFast) {
        storkFast = IStorkFast(_storkFast);
    }

    function useStorkFast(bytes calldata payload) public payable returns (StorkFastStructs.Asset[] memory assets) {
        StorkFastStructs.Asset[] memory assets = storkFast.verifyAndDeserializeSignedECDSAPayload{value: msg.value}(payload);

        return assets;
    }
}
```
