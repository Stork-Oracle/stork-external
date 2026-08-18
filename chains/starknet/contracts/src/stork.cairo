//! The Stork oracle contract for Starknet.
//!
//! This is the Cairo counterpart of `chains/evm/contracts/stork` and behaves the same way: it
//! stores the latest Stork-signed value per asset feed, verifying each update against a secp256k1
//! signature produced by Stork's EVM signing key.
//!
//! Two things differ from the EVM contract, both forced by the platform:
//!
//! * Fees are collected as an ERC20 `transfer_from` rather than `msg.value`, since Starknet has no
//!   native value transfer. A deployment with `single_update_fee` of 0 (the default) never touches
//!   the token, so the pusher needs no allowance in that case.
//! * Upgrades use `replace_class_syscall` rather than a UUPS proxy.

#[starknet::contract]
pub mod Stork {
    use core::num::traits::Zero;
    use starknet::storage::{
        Map, MutableVecTrait, StoragePathEntry, StoragePointerReadAccess, StoragePointerWriteAccess,
        Vec, VecTrait,
    };
    use starknet::syscalls::replace_class_syscall;
    use starknet::{
        ClassHash, ContractAddress, EthAddress, get_block_timestamp, get_caller_address,
        get_contract_address,
    };
    use stork_starknet_sdk::errors::StorkErrors;
    use stork_starknet_sdk::events::{
        FeeTokenUpdate, OwnershipTransferStarted, OwnershipTransferred, SigningAddressAdded,
        SigningAddressRemoved, SingleUpdateFeeUpdate, StorkPublicKeyUpdate, Upgraded,
        ValidTimePeriodUpdate, ValueUpdate,
    };
    use stork_starknet_sdk::interface::{
        IERC20Dispatcher, IERC20DispatcherTrait, IStork, TemporalNumericValueInput,
    };
    use stork_starknet_sdk::temporal_numeric_value::{EncodedAssetId, TemporalNumericValue};
    use crate::verify::verify_stork_evm_signature;

    /// Matches `MAX_SIGNING_ADDRESSES` in `StorkSetters.sol`.
    const MAX_SIGNING_ADDRESSES: u64 = 8;

    const NANOS_PER_SECOND: u64 = 1_000_000_000;

    const VERSION: felt252 = '1.0.0';

    #[storage]
    struct Storage {
        /// The canonical key updates are expected to be signed with.
        stork_public_key: EthAddress,
        /// Every accepted signing key, for rotation with backwards compatibility. The signer is
        /// part of the signed message, so verification has to try each one in turn.
        signing_address_list: Vec<EthAddress>,
        /// O(1) membership check over `signing_address_list`.
        signing_addresses: Map<felt252, bool>,
        single_update_fee: u256,
        fee_token: ContractAddress,
        /// Maximum acceptable age before a value is considered stale. This covers attestation
        /// delay, block time, and potential clock drift between the source and target chains.
        valid_time_period_seconds: u64,
        latest_canonical_temporal_numeric_values: Map<EncodedAssetId, TemporalNumericValue>,
        owner: ContractAddress,
        pending_owner: ContractAddress,
    }

    #[event]
    #[derive(Drop, starknet::Event)]
    pub enum Event {
        ValueUpdate: ValueUpdate,
        StorkPublicKeyUpdate: StorkPublicKeyUpdate,
        SigningAddressAdded: SigningAddressAdded,
        SigningAddressRemoved: SigningAddressRemoved,
        SingleUpdateFeeUpdate: SingleUpdateFeeUpdate,
        FeeTokenUpdate: FeeTokenUpdate,
        ValidTimePeriodUpdate: ValidTimePeriodUpdate,
        OwnershipTransferStarted: OwnershipTransferStarted,
        OwnershipTransferred: OwnershipTransferred,
        Upgraded: Upgraded,
    }

    #[constructor]
    fn constructor(
        ref self: ContractState,
        initial_owner: ContractAddress,
        stork_public_key: EthAddress,
        valid_time_period_seconds: u64,
        single_update_fee: u256,
        fee_token: ContractAddress,
    ) {
        assert(initial_owner.is_non_zero(), StorkErrors::ZERO_ADDRESS);
        // A non-zero fee with no fee token would make every update revert on the token call.
        assert(
            single_update_fee.is_zero() || fee_token.is_non_zero(), StorkErrors::FEE_TOKEN_UNSET,
        );

        self.owner.write(initial_owner);
        self.valid_time_period_seconds.write(valid_time_period_seconds);
        self.single_update_fee.write(single_update_fee);
        self.fee_token.write(fee_token);
        self.set_stork_public_key(stork_public_key);
        self.store_add_signing_address(stork_public_key);

        self.emit(ValidTimePeriodUpdate { valid_time_period_seconds });
        self.emit(SingleUpdateFeeUpdate { single_update_fee });
        self.emit(FeeTokenUpdate { fee_token });
        self.emit(OwnershipTransferred { previous_owner: Zero::zero(), new_owner: initial_owner });
    }

    #[abi(embed_v0)]
    impl StorkImpl of IStork<ContractState> {
        fn update_temporal_numeric_values_v1(
            ref self: ContractState, update_data: Array<TemporalNumericValueInput>,
        ) {
            let mut num_updates: u256 = 0;

            for update in update_data {
                // Check recency before verifying. Signature verification dominates the cost of
                // this call on Starknet, and a stale update is discarded either way.
                if !self.is_fresh(update.id, update.temporal_numeric_value.timestamp_ns) {
                    continue;
                }

                assert(self.is_valid_stork_signer(update), StorkErrors::INVALID_SIGNATURE);

                self
                    .latest_canonical_temporal_numeric_values
                    .entry(update.id)
                    .write(update.temporal_numeric_value);
                self
                    .emit(
                        ValueUpdate {
                            id: update.id,
                            timestamp_ns: update.temporal_numeric_value.timestamp_ns,
                            quantized_value: update.temporal_numeric_value.quantized_value,
                        },
                    );

                num_updates += 1;
            }

            assert(num_updates > 0, StorkErrors::NO_FRESH_UPDATE);

            self.collect_fee(num_updates);
        }

        fn get_temporal_numeric_value_v1(
            self: @ContractState, id: EncodedAssetId,
        ) -> TemporalNumericValue {
            let value = self.get_temporal_numeric_value_unsafe_v1(id);

            let value_seconds = value.timestamp_ns / NANOS_PER_SECOND;
            let now = get_block_timestamp();
            if now >= value_seconds {
                assert(
                    now - value_seconds <= self.valid_time_period_seconds.read(),
                    StorkErrors::STALE_VALUE,
                );
            }

            value
        }

        fn get_temporal_numeric_value_unsafe_v1(
            self: @ContractState, id: EncodedAssetId,
        ) -> TemporalNumericValue {
            let value = self.latest_canonical_temporal_numeric_values.entry(id).read();
            assert(value.timestamp_ns != 0, StorkErrors::NOT_FOUND);

            value
        }

        fn get_temporal_numeric_values_unsafe_v1(
            self: @ContractState, ids: Span<EncodedAssetId>,
        ) -> Array<TemporalNumericValue> {
            let mut values = array![];
            for id in ids {
                values.append(self.get_temporal_numeric_value_unsafe_v1(*id));
            }

            values
        }

        fn get_multiple_temporal_numeric_values_unchecked(
            self: @ContractState, ids: Span<EncodedAssetId>,
        ) -> Array<TemporalNumericValue> {
            let mut values = array![];
            for id in ids {
                values.append(self.latest_canonical_temporal_numeric_values.entry(*id).read());
            }

            values
        }

        fn get_update_fee_v1(
            self: @ContractState, update_data: Span<TemporalNumericValueInput>,
        ) -> u256 {
            self.single_update_fee.read() * update_data.len().into()
        }

        fn version(self: @ContractState) -> felt252 {
            VERSION
        }

        fn stork_public_key(self: @ContractState) -> EthAddress {
            self.stork_public_key.read()
        }

        fn signing_addresses(self: @ContractState) -> Array<EthAddress> {
            let mut addresses = array![];
            for i in 0..self.signing_address_list.len() {
                addresses.append(self.signing_address_list.at(i).read());
            }

            addresses
        }

        fn single_update_fee(self: @ContractState) -> u256 {
            self.single_update_fee.read()
        }

        fn fee_token(self: @ContractState) -> ContractAddress {
            self.fee_token.read()
        }

        fn valid_time_period_seconds(self: @ContractState) -> u64 {
            self.valid_time_period_seconds.read()
        }

        fn update_valid_time_period_seconds(
            ref self: ContractState, valid_time_period_seconds: u64,
        ) {
            self.assert_only_owner();
            self.valid_time_period_seconds.write(valid_time_period_seconds);
            self.emit(ValidTimePeriodUpdate { valid_time_period_seconds });
        }

        fn update_single_update_fee(ref self: ContractState, single_update_fee: u256) {
            self.assert_only_owner();
            // Setting a fee before a fee token would brick updates, so require the token first.
            assert(
                single_update_fee.is_zero() || self.fee_token.read().is_non_zero(),
                StorkErrors::FEE_TOKEN_UNSET,
            );

            self.single_update_fee.write(single_update_fee);
            self.emit(SingleUpdateFeeUpdate { single_update_fee });
        }

        fn update_fee_token(ref self: ContractState, fee_token: ContractAddress) {
            self.assert_only_owner();
            // Clearing the token is only safe once nothing is being charged.
            assert(
                fee_token.is_non_zero() || self.single_update_fee.read().is_zero(),
                StorkErrors::FEE_TOKEN_UNSET,
            );

            self.fee_token.write(fee_token);
            self.emit(FeeTokenUpdate { fee_token });
        }

        fn update_stork_public_key(ref self: ContractState, stork_public_key: EthAddress) {
            self.assert_only_owner();
            self.set_stork_public_key(stork_public_key);

            // Keep the rotated-in key accepted for verification, matching the EVM contract's
            // fallback to `storkPublicKey()`.
            if !self.signing_addresses.entry(stork_public_key.into()).read() {
                self.store_add_signing_address(stork_public_key);
            }
        }

        fn add_signing_address(ref self: ContractState, signing_address: EthAddress) {
            self.assert_only_owner();
            self.store_add_signing_address(signing_address);
        }

        fn remove_signing_address(ref self: ContractState, signing_address: EthAddress) {
            self.assert_only_owner();
            assert(signing_address.is_non_zero(), StorkErrors::ZERO_ADDRESS);
            assert(
                signing_address != self.stork_public_key.read(),
                StorkErrors::CANNOT_REMOVE_CANONICAL,
            );
            assert(
                self.signing_addresses.entry(signing_address.into()).read(),
                StorkErrors::ADDRESS_NOT_FOUND,
            );

            self.signing_addresses.entry(signing_address.into()).write(false);

            // Swap the removed entry with the last one, then drop the tail.
            let len = self.signing_address_list.len();
            for i in 0..len {
                if self.signing_address_list.at(i).read() == signing_address {
                    let last = self.signing_address_list.at(len - 1).read();
                    self.signing_address_list.at(i).write(last);
                    break;
                }
            }
            let _ = self.signing_address_list.pop();

            self.emit(SigningAddressRemoved { signing_address });
        }

        fn withdraw_fees(ref self: ContractState, recipient: ContractAddress, amount: u256) {
            self.assert_only_owner();
            assert(recipient.is_non_zero(), StorkErrors::ZERO_ADDRESS);

            IERC20Dispatcher { contract_address: self.fee_token.read() }
                .transfer(recipient, amount);
        }

        fn upgrade(ref self: ContractState, new_class_hash: ClassHash) {
            self.assert_only_owner();
            assert(new_class_hash.is_non_zero(), StorkErrors::ZERO_CLASS_HASH);

            replace_class_syscall(new_class_hash).unwrap();
            self.emit(Upgraded { class_hash: new_class_hash });
        }

        fn owner(self: @ContractState) -> ContractAddress {
            self.owner.read()
        }

        fn pending_owner(self: @ContractState) -> ContractAddress {
            self.pending_owner.read()
        }

        fn transfer_ownership(ref self: ContractState, new_owner: ContractAddress) {
            self.assert_only_owner();
            assert(new_owner.is_non_zero(), StorkErrors::ZERO_ADDRESS);

            self.pending_owner.write(new_owner);
            self.emit(OwnershipTransferStarted { previous_owner: self.owner.read(), new_owner });
        }

        fn accept_ownership(ref self: ContractState) {
            let caller = get_caller_address();
            assert(caller == self.pending_owner.read(), StorkErrors::NOT_PENDING_OWNER);

            let previous_owner = self.owner.read();
            self.owner.write(caller);
            self.pending_owner.write(Zero::zero());

            self.emit(OwnershipTransferred { previous_owner, new_owner: caller });
        }
    }

    #[generate_trait]
    impl InternalImpl of InternalTrait {
        fn assert_only_owner(self: @ContractState) {
            assert(get_caller_address() == self.owner.read(), StorkErrors::NOT_OWNER);
        }

        /// True if `timestamp_ns` is newer than what is already stored for `id`.
        fn is_fresh(self: @ContractState, id: EncodedAssetId, timestamp_ns: u64) -> bool {
            timestamp_ns > self
                .latest_canonical_temporal_numeric_values
                .entry(id)
                .read()
                .timestamp_ns
        }

        /// True if `update` carries a valid signature from any accepted signing key.
        ///
        /// The canonical key is tried first because it signs essentially every update; the rest of
        /// the list only matters during a key rotation.
        fn is_valid_stork_signer(self: @ContractState, update: TemporalNumericValueInput) -> bool {
            let canonical = self.stork_public_key.read();
            if verify_update(update, canonical) {
                return true;
            }

            let mut valid = false;
            for i in 0..self.signing_address_list.len() {
                let signer = self.signing_address_list.at(i).read();
                if signer != canonical && verify_update(update, signer) {
                    valid = true;
                    break;
                }
            }

            valid
        }

        fn set_stork_public_key(ref self: ContractState, stork_public_key: EthAddress) {
            assert(stork_public_key.is_non_zero(), StorkErrors::ZERO_ADDRESS);

            self.stork_public_key.write(stork_public_key);
            self.emit(StorkPublicKeyUpdate { stork_public_key });
        }

        fn store_add_signing_address(ref self: ContractState, signing_address: EthAddress) {
            assert(signing_address.is_non_zero(), StorkErrors::ZERO_ADDRESS);
            assert(
                !self.signing_addresses.entry(signing_address.into()).read(),
                StorkErrors::ADDRESS_ALREADY_EXISTS,
            );
            assert(
                self.signing_address_list.len() < MAX_SIGNING_ADDRESSES,
                StorkErrors::ADDRESS_LIMIT_REACHED,
            );

            self.signing_addresses.entry(signing_address.into()).write(true);
            self.signing_address_list.push(signing_address);

            self.emit(SigningAddressAdded { signing_address });
        }

        /// Pulls `num_updates * single_update_fee` from the caller. No-op when the fee is 0, which
        /// keeps the common fee-less deployment from requiring an ERC20 allowance.
        fn collect_fee(ref self: ContractState, num_updates: u256) {
            let total_fee = self.single_update_fee.read() * num_updates;
            if total_fee.is_zero() {
                return;
            }

            let paid = IERC20Dispatcher { contract_address: self.fee_token.read() }
                .transfer_from(get_caller_address(), get_contract_address(), total_fee);
            assert(paid, StorkErrors::INSUFFICIENT_FEE);
        }
    }

    /// Free function so the borrow checker does not keep `self` alive across verification.
    fn verify_update(update: TemporalNumericValueInput, signer: EthAddress) -> bool {
        verify_stork_evm_signature(
            signer,
            update.id,
            update.temporal_numeric_value.timestamp_ns,
            update.temporal_numeric_value.quantized_value,
            update.publisher_merkle_root,
            update.value_compute_alg_hash,
            update.r,
            update.s,
            update.v,
        )
    }
}
