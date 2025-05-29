# Stork Proxy Invariants

## 1. Initialization Invariants

Single Owner Initialization:
The owner can only be set once. The set_owner function checks that storage::SRC14.owner is State::Uninitialized before setting it to State::Initialized(owner). If the owner is already initialized, the function reverts with the error "Owner already initialized".
Once set, subsequent calls to set_owner will fail.
Proxy Target Initialization:
The target field in storage is initialized to ContractId::zero() by default. It can be updated multiple times via the set_proxy_target function, but only by the owner.

## 2. State Management Invariants

Owner State:
The owner field in storage (storage::SRC14.owner) can be in one of three states:
State::Uninitialized (default),
State::Initialized(Identity) (after set_owner is called),
State::Revoked (possible but not explicitly set in the provided code).
Once set to Initialized, the owner cannot be changed due to the check in set_owner.
Target Contract:
The target field (storage::SRC14.target) holds the ContractId of the contract to which method calls are forwarded via the fallback function. It can be updated by the owner using set_proxy_target.

## 3. Access Control Invariants

Owner-Only Functions:
The set_proxy_target function can only be called by the owner. This is enforced by the only_owner function, which:
Allows execution if the owner is Uninitialized or Revoked (no revert in these cases),
Reverts with AccessError::NotOwner if the owner is Initialized and the caller (msg_sender()) does not match the stored owner.
Owner Setting:
The set_owner function can only succeed when the owner is Uninitialized. After initialization, any attempt to call it will revert.

## 4. Fallback Function Invariants

Proxy Behavior:
The fallback function forwards any method call not explicitly defined in the contract to the target contract using run_external.
This means that calls to undefined functions are proxied to the contract specified in storage::SRC14.target.
Target Requirement:
The fallback function assumes target is readable from storage. If target is ContractId::zero() (the default value), calling run_external with this zero address may result in undefined behavior or reversion, depending on the platform's implementation.

## 5. Storage Layout Invariants

Specific Storage Slots:
The target is stored at a specific slot: 0x7bb458adc1d118713319a5baa00a2d049dd64d2916477d2688d76970c898cd55 (computed as sha256("storage_SRC14_0")).
The owner is stored as part of the SRC14 storage structure, following the standard SRC14 layout.
This ensures a predictable storage layout accessible by external contracts or tools.

## 6. Function Behavior Invariants

set_proxy_target(new_target: ContractId):
Can only be called by the owner (enforced by only_owner).
Updates the target field in storage to new_target.
proxy_target() -> Option<ContractId>:
Returns Some(ContractId) if the target is set (readable from storage), otherwise None if reading fails (though this is unlikely given the default ContractId::zero()).
proxy_owner() -> State:
Returns the current state of the owner: Uninitialized, Initialized(Identity), or Revoked.
set_owner(owner: Identity):
Can only be called when the owner is Uninitialized.
Sets the owner to State::Initialized(owner) and prevents further changes.
only_owner():
Does not revert if the owner is Uninitialized or Revoked.
Reverts with AccessError::NotOwner if the owner is Initialized and the caller is not the owner.

## 7. Error Handling Invariants

Access Control Errors:
The only_owner function reverts with AccessError::NotOwner if the owner is Initialized and the caller does not match the stored owner identity.
Initialization Errors:
The set_owner function reverts with "Owner already initialized" if the owner is not Uninitialized when called.

## How to Use These Invariants for Testing

You can design tests to verify each invariant as follows:

### Initialization:

Call set_owner twice and verify that the second call reverts.
Set the owner, then call set_proxy_target to ensure it works only for the owner.
State Management:
Test that proxy_owner reflects the correct state (Uninitialized initially, Initialized after set_owner).
Update target via set_proxy_target and confirm it persists in storage.
Access Control:
Call set_proxy_target with a non-owner identity (after setting the owner) and verify it reverts with AccessError::NotOwner.
Fallback Behavior:
Call an undefined function and confirm it forwards to the target contract. Test with target set to ContractId::zero() to observe platform-specific behavior.
Storage:
Verify that target and owner are stored and readable at their expected slots.
Function Behavior:
Test each function (set_proxy_target, proxy_target, proxy_owner, set_owner) with valid and invalid inputs to ensure correct behavior.
Error Handling:
Trigger error conditions (e.g., non-owner calling set_proxy_target, multiple set_owner calls) and confirm the contract reverts with the expected errors.