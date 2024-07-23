// SPDX-License-Identifier: Apache 2

pragma solidity >=0.8.24 <0.9.0;

library StorkErrors {
    // Insufficient fee is paid to the method.
    error InsufficientFee();
    // There is no fresh update, whereas expected fresh updates.
    error NoFreshUpdate();
    // Not found.
    error NotFound();
    // Requested value is stale.
    error StaleValue();
    // Signature is invalid.
    error InvalidSignature();
}
