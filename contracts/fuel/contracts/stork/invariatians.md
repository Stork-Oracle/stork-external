# Stork Invariants

## 1. Initialization Invariants

Single Initialization: The contract can only be initialized once. After storage.initialized is set to true in the initialize function, subsequent calls to initialize should fail due to the require(!storage.initialized.read(), "Already initialized") check.
Atomic Initialization: During the initialize function, storage.initializing is set to true at the start and storage.initialized is set to true at the end. This ensures that initialization is an atomic process, and no reentrancy or partial initialization can occur.

## 2. State Management Invariants

Restricted State Modifications: The state struct in storage contains stork_public_key, single_update_fee_in_wei, and valid_time_period_seconds. These fields can only be modified by the owner through the functions:
update_stork_public_key
update_single_update_fee_in_wei
update_valid_time_period_seconds These functions internally call only_owner, ensuring that only the owner can make changes.

## 3. Temporal Numeric Value Mapping Invariants

Incremental Mapping Count: Each call to create_temporal_numeric_value_mapping increments temporal_numeric_value_mapping_instance_count by 1 and returns a new StorageKey<StorageMap<b256, TemporalNumericValue>> corresponding to the new index.
Immutable Canonical Mapping: The latest_canonical_temporal_numeric_values field in the state struct is set during initialization via _initialize and should not be modified afterward. It remains a reference to the first created mapping.

## 4. Update Logic Invariants

Timestamp Monotonicity: In update_latest_value_if_necessary, a value is updated only if the new timestamp_ns is greater than the existing timestamp_ns for the same id. This ensures that only fresher values are accepted, maintaining a monotonic increase in timestamps.
Signature and Fee Requirements: The update_temporal_numeric_values_v1 function:
Requires a valid Stork signature for each update, verified by _verify_stork_signature_v1.
Checks if the update is fresh (newer timestamp) by calling update_latest_value_if_necessary.
Ensures the total fee paid (msg_amount()) is at least single_update_fee_in_wei * number_of_updates. If not, it reverts with InsufficientFee.

## 5. Signature Verification Invariants

Stork Signature Verification: The _verify_stork_signature_v1 function:
Constructs a message hash using get_stork_message_hash_v1 with fields like id, recvTime, and quantized_value.
Verifies that the signature (r, s) matches the expected stork_public_key.
Publisher Signature Verification: The _verify_publisher_signature_v1 function:
Constructs a message hash using get_publisher_message_hash with fields like oraclePubKey, asset_pair_id, and timestamp.
Verifies that the signature (r, s) matches the expected oraclePubKey.
Merkle Root Verification: The verify_merkle_root function:
Computes the Merkle root from a vector of leaves using compute_merkle_root.
Checks if the computed root matches the provided root.

## 6. Access Control Invariants

Owner-Only Functions: The following functions can only be called by the owner, enforced by the only_owner check:
_update_valid_time_period_seconds
_update_single_update_fee_in_wei
_update_stork_public_key
Owner Initialization: The owner is set during initialize and stored in storage.owner as Initialized(initialOwner). Subsequent calls to owner-restricted functions verify the caller against this value.

## 7. Fee Calculation Invariants

Fee Proportionality: The total fee for updates in update_temporal_numeric_values_v1 is calculated as single_update_fee_in_wei * number_of_updates, where number_of_updates is the count of successfully updated values (i.e., those with newer timestamps).
Fee Enforcement: The contract reverts with InsufficientFee if the amount sent (msg_amount()) is less than the required fee.

## 8. Temporal Value Retrieval Invariants

Safe Retrieval (get_temporal_numeric_value_v1):
Reverts with NotFound if no value exists for the given id (i.e., timestamp_ns == 0).
Reverts with StaleValue if the value’s timestamp is older than valid_time_period_seconds relative to the current block timestamp.
Returns the TemporalNumericValue only if it exists and is not stale.
Unsafe Retrieval (get_temporal_numeric_value_unsafe_v1):
Reverts with NotFound if no value exists for the given id (i.e., timestamp_ns == 0).
Returns the TemporalNumericValue without checking for staleness.

## 9. Version Invariant

Fixed Version: The version function always returns the string "1.0.2", indicating the contract’s version. This should remain constant unless the contract is updated.