// Unlike the EVM and Solana bindings, the Sui bindings are not generated from the Move source code, as a tool for this does not currently exist.
// Instead, this file contains utility functions for interacting with the Sui Stork contract.
// These functions are written using https://github.com/block-vision/sui-go-sdk

package contract_bindings_sui

import (
	"context"
	"math/big"

	"github.com/coming-chat/go-sui/v2/account"
	"github.com/coming-chat/go-sui/v2/client"
	"github.com/coming-chat/go-sui/v2/sui_types"
	"github.com/coming-chat/go-sui/v2/types"
)

type StorkContract struct {
	client           *client.Client
	account          *account.Account
	contract_address sui_types.SuiAddress
	state            StorkState
}

type MultipleUpdateData struct {
	Ids                              [][]byte
	TemporalNumericValueTimestampNss []big.Int
	TemporalNumericValueMagnitudes   []big.Int
	TemporalNumericValueNegatives    []bool
	PublisherMerkleRoots             [][]byte
	ValueComputeAlgHashes            [][]byte
	Rs                               [][]byte
	SS                               [][]byte
	Vs                               []byte
}

type StorkState struct {
	Id                    sui_types.SuiAddress
	StorkSuiPublicKey     sui_types.SuiAddress
	StorkEvmPublicKey     string
	SingleUpdateFeeInMist uint64
	Version               uint64
}

func NewStorkContract(rpcUrl string, contractAddress string, key string) (*StorkContract, error) {
	client, err := client.Dial(rpcUrl)
	if err != nil {
		return nil, err
	}
	account, err := account.NewAccountWithKeystore(key)
	if err != nil {
		return nil, err
	}

	return &StorkContract{client: client, account: account, contract_address: contractAddress, state: getStorkState()}, nil
}

func NewMultipleUpdateData(
	Ids [][]byte,
	TemporalNumericValueTimestampNss []big.Int,
	TemporalNumericValueMagnitudes []big.Int,
	TemporalNumericValueNegatives []bool,
	PublisherMerkleRoots [][]byte,
	ValueComputeAlgHashes [][]byte,
	Rs [][]byte,
	SS [][]byte,
	Vs []byte,
) MultipleUpdateData {
	return MultipleUpdateData{
		Ids:                              Ids,
		TemporalNumericValueTimestampNss: TemporalNumericValueTimestampNss,
		TemporalNumericValueMagnitudes:   TemporalNumericValueMagnitudes,
		TemporalNumericValueNegatives:    TemporalNumericValueNegatives,
		PublisherMerkleRoots:             PublisherMerkleRoots,
		ValueComputeAlgHashes:            ValueComputeAlgHashes,
		Rs:                               Rs,
		SS:                               SS,
		Vs:                               Vs,
	}
}

// Listens for any events emitted by the Stork contract
func (sc *StorkContract) ListenContractEvents(ch chan map[InternalEncodedAssetId]InternalStorkStructsTemporalNumericValue) {

}

func (sc *StorkContract) UpdateMultipleTemporalNumericValueEvm(updateData MultipleUpdateData) {

}

// utility functions
func getStorkState(contractAddress sui_types.SuiAddress, client *client.Client) (StorkState, error) {
	eventFilter := types.EventFilter{
		MoveModule: &struct {
			Package sui_types.ObjectID `json:"package"`
			Module  string             `json:"module"`
		}{
			Package: contractAddress,
			Module:  "stork",
		},
	}
	limit := uint32(1)
	event, err := client.QueryEvents(context.Background(), eventFilter, nil, &limit, false)
	if err != nil {
		return StorkState{}, err
	}

	stork_state_id_hex := event.Data[0].ParsedJson.(map[string]interface{})["stork_state_id"].(string)
	stork_state_id, err := sui_types.NewAddressFromHex(stork_state_id_hex)
	if err != nil {
		return StorkState{}, err
	}

	options := &types.SuiObjectDataOptions{
		ShowContent: true,
	}
	object, err := client.GetObject(context.Background(), *stork_state_id, options)
	if err != nil {
		return StorkState{}, err
	}
	fields := object.Data.Content.Data.MoveObject.Fields.(map[string]interface{})
	stork_evm_public_key := fields["stork_evm_public_key"].(map[string]interface{})["bytes"].(string)
	stork_sui_public_key_string := fields["stork_sui_public_key"].(map[string]interface{})["address"].(string)
	stork_sui_public_key, err := sui_types.NewAddressFromHex(stork_sui_public_key_string)
	if err != nil {
		return StorkState{}, err
	}
	single_update_fee_in_mist := uint64(fields["single_update_fee_in_mist"].(float64))
	version, err := getVersion(*stork_state_id, client)
	if err != nil {
		return StorkState{}, err
	}

	return StorkState{Id: *stork_state_id, StorkEvmPublicKey: stork_evm_public_key, StorkSuiPublicKey: *stork_sui_public_key, SingleUpdateFeeInMist: single_update_fee_in_mist, Version: version}, nil
}
