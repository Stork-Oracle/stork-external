// SPDX-License-Identifier: Apache 2

pragma solidity >=0.8.24 <0.9.0;

library StorkErrors {
    // Insufficient fee is paid to the method.
    // 0x025dbdd4
    error InsufficientFee();
    // There is no fresh update, whereas expected fresh updates.
    // 0xde2c57fa
    error NoFreshUpdate();
    // Not found.
    // 0xc5723b51
    error NotFound();
    // Requested value is stale.
    // 0x24c4fe43
    error StaleValue();
    // Signature is invalid.
    // 0x8baa579f
    error InvalidSignature();
}
