//! A stand-in class used to prove `upgrade` actually replaces the running code.

#[starknet::interface]
pub trait IUpgraded<TContractState> {
    fn version(self: @TContractState) -> felt252;
}

#[starknet::contract]
pub mod UpgradedStork {
    #[storage]
    struct Storage {}

    #[abi(embed_v0)]
    impl UpgradedImpl of super::IUpgraded<ContractState> {
        fn version(self: @ContractState) -> felt252 {
            '2.0.0'
        }
    }
}
