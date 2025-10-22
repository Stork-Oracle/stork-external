use starknet::ContractAddress;

#[starknet::interface]
pub trait IStork<TContractState> {
    fn update_temporal_numeric_values(ref self: TContractState, updates: Span<TemporalNumericValueInput>);
    fn update_single_update_fee_in_wei(ref self: TContractState, fee: felt252);
    fn update_stork_public_key(ref self: TContractState, storkPublicKey: StorkPublicKey);

    fn verify_stork_signature(self: @TContractState, update: TemporalNumericValueInput) -> bool;
    fn version(self: @TContractState) -> felt252;

    fn get_temporal_numeric_value(self: @TContractState, encoded: EncodedAssetId) -> TemporalNumericValue;
    fn get_total_fee(self: @TContractState, numUpdates: felt252) -> felt252;
}

#[starknet::contract]
mod Stork {
    use crate::{StorkPublicKey, EncodedAssetId, TemporalNumericValue, TemporalNumericValueInput};
    use starknet::{ContractAddress, get_caller_address};
    use starknet::storage::{Map, StoragePathEntry, StoragePointerReadAccess, StoragePointerWriteAccess};

    #[storage]
    struct Storage {
        // todo: look into bit packing?
        ownerAddress: ContractAddress,
        storkPublicKey: StorkPublicKey,
        singleUpdateFeeInWei: felt252,
        latestValues: Map<EncodedAssetId, TemporalNumericValue>,
    }

    #[constructor]
    fn constructor(
        ref self: ContractState, 
        ownerAddress: ContractAddress, 
        storkPublicKey: StorkPublicKey, 
        singleUpdateFee: felt252,
    ) {
        self.ownerAddress.write(ownerAddress);
        self.storkPublicKey.write(storkPublicKey);
        self.singleUpdateFeeInWei.write(singleUpdateFee);
    }

    #[abi(embed_v0)]
    impl StorkImpl of super::IStork<ContractState> {
        fn update_temporal_numeric_values(ref self: ContractState, updates: Span<TemporalNumericValueInput>) {
            for update in updates {
                assert(self.verify_stork_signature(*update), 'Invalid Stork signature');
                let latestValue = self.latestValues.entry(*update.id).read();

                if latestValue.timestampNs >= *update.temporalNumericValue.timestampNs {
                    continue;
                }

                self.latestValues.entry(*update.id).write(*update.temporalNumericValue);
            }
            // todo: do you enforce paying gas here?
        }

        fn update_single_update_fee_in_wei(ref self: ContractState, fee: felt252) {
            let caller = get_caller_address();
            assert(caller == self.ownerAddress.read(), 'Unauthorized');
            // felt252 is always non-negative
            self.singleUpdateFeeInWei.write(fee);
        }

        fn update_stork_public_key(ref self: ContractState, storkPublicKey: StorkPublicKey) {
            let caller = get_caller_address();
            assert(caller == self.ownerAddress.read(), 'Unauthorized');
            self.storkPublicKey.write(storkPublicKey);
        }

        fn verify_stork_signature(self: @ContractState, update: TemporalNumericValueInput) -> bool {
            // todo: implement
            true // todo: placeholder
        }

        fn version(self: @ContractState) -> felt252 {
            '1.0.0'
        }

        fn get_temporal_numeric_value(self: @ContractState, encoded: EncodedAssetId) -> TemporalNumericValue {
            let tnv = self.latestValues.entry(encoded).read();
            assert(tnv.timestampNs != 0 && tnv.quantizedValue != 0, 'Value not in latestValues map');
            tnv
        }

        fn get_total_fee(self: @ContractState, numUpdates: felt252) -> felt252 {
            numUpdates * self.singleUpdateFeeInWei.read()
        }
    }
}

type StorkPublicKey = ContractAddress;
type EncodedAssetId = felt252;

#[derive(Drop, Copy, Serde, starknet::Store)]
struct TemporalNumericValue {
    timestampNs: u64,
    quantizedValue: felt252,
}

#[derive(Drop, Copy, Serde, starknet::Store)]
struct TemporalNumericValueInput {
    temporalNumericValue: TemporalNumericValue,
    id: EncodedAssetId,
    publisherMerkleRoot: felt252,
    valueComputeAlgHash: felt252,
    r: felt252,
    s: felt252,
}
