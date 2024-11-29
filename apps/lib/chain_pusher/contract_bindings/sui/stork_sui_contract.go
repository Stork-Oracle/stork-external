// Unlike the EVM and Solana bindings, the Sui bindings are not generated from the Move source code, as a tool for this does not currently exist.
// Instead, this file contains utility functions for interacting with the Sui Stork contract.
// These functions are written using https://github.com/block-vision/sui-go-sdk

package contract_bindings_sui

import (
	_ "github.com/block-vision/sui-go-sdk/constant"
	_ "github.com/block-vision/sui-go-sdk/models"
	_ "github.com/block-vision/sui-go-sdk/signer"
	_ "github.com/block-vision/sui-go-sdk/sui"
	_ "github.com/block-vision/sui-go-sdk/utils"
)

type StorkContract struct {
	client *client.Client
	address string
}

func NewStorkContract(rpcUrl string) (*StorkContract, error) {}

// Listens for any events emitted by the Stork contract
func (sc *StorkContract) ListenContractEvents(ch chan map[InternalEncodedAssetId]InternalStorkStructsTemporalNumericValue) {}

// stork functions
func (sc *StorkContract) InitStork(adminCap AdminCap, storkSuiPublicKey address, storkEvmPublicKey []byte) {}

func (sc *StorkContract) update_single_temporal_numeric_value_evm(updateData UpdateTemporalNumericValueEvmInput) {}

func (sc *StorkContract) update_multiple_temporal_numeric_value_evm(updateData [UpdateTemporalNumericValueEvmInputVec) {}

// State functions
func (sc *StorkContract) get_stork_evm_public_key() []byte {}

func (sc *StorkContract) get_stork_sui_public_key() address {}

func (sc *StorkContract) get_single_update_fee_in_mist() uint64 {}

func (sc *StorkContract) get_version() uint64 {}

// state admin functions
func (sc *StorkContract) update_single_update_fee_in_mist(newSingleUpdateFeeInMist uint64) {}

func (sc *StorkContract) update_stork_sui_public_key(newStorkSuiPublicKey address) {}

func (sc *StorkContract) update_stork_evm_public_key(newStorkEvmPublicKey []byte) {}

func (sc *StorkContract) withdraw_fees(adminCap AdminCap) {}

func (sc *StorkContract) migrate(adminCap AdminCap) {}

// utility functions
func (sc *StorkContract) get_stork_state() StorkState {}

func (sc *StorkContract) get_admin_cap() AdminCap {}
