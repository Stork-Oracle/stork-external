// Unlike the EVM and Solana bindings, the Sui bindings are not generated from the Move source code, as a tool for this does not currently exist.
// Instead, this file contains utility functions for interacting with the Sui Stork contract.
// These functions are written using https://github.com/block-vision/sui-go-sdk

package contract_bindings_sui

import (
	"context"
	"math/big"

	"github.com/Stork-Oracle/stork-external/apps/lib/chain_pusher/model"
	"github.com/coming-chat/go-sui/v2/account"
	"github.com/coming-chat/go-sui/v2/client"
	"github.com/coming-chat/go-sui/v2/lib"
	"github.com/coming-chat/go-sui/v2/sui_types"
	"github.com/coming-chat/go-sui/v2/types"
)

type StorkContract struct {
	Client           *client.Client
	Account          *account.Account
	ContractAddress  sui_types.SuiAddress
	State            StorkState
	ReferenceGasPrice uint64
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

type UpdateData struct {
	Id                              []byte
	TemporalNumericValueTimestampNs uint64
	TemporalNumericValueMagnitude   *big.Int
	TemporalNumericValueNegative    bool
	PublisherMerkleRoot             []byte
	ValueComputeAlgHash             []byte
	R                               []byte
	S                               []byte
	V                               byte
}

type StorkState struct {
	Id                    sui_types.SuiAddress
	StorkSuiPublicKey     sui_types.SuiAddress
	StorkEvmPublicKey     string
	SingleUpdateFeeInMist uint64
	Version               uint64
	FeedRegistry          FeedRegistry
}

type FeedRegistry struct {
	Id sui_types.SuiAddress
	Entries map[model.InternalEncodedAssetId]sui_types.SuiAddress
}

type TemporalNumericValue struct {
	TimestampNs uint64
	QuantizedValue *big.Int
}

type I128 struct {
	Magnitude *big.Int
	Negative bool
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
	contractAddr, err := sui_types.NewAddressFromHex(contractAddress)
	if err != nil {
		return nil, err
	}
	state, err := getStorkState(*contractAddr, client)
	if err != nil {
		return nil, err
	}

	reference_gas_price, err := client.GetReferenceGasPrice()	
	
	if err != nil {
		return nil, err
	}

	return &StorkContract{Client: client, Account: account, ContractAddress: *contractAddr, State: state, ReferenceGasPrice: reference_gas_price}, nil
}

func (sc *StorkContract) SetReferenceGasPrice(reference_gas_price uint64) {
	sc.reference_gas_price = reference_gas_price
}

// gets multiple temporal numeric values at a time to save on RPC calls
func (sc *StorkContract) GetMultipleTemporalNumericValuesUnchecked(feedIds []model.InternalEncodedAssetId) (map[model.InternalEncodedAssetId]model.InternalStorkStructsTemporalNumericValue, error) {
	feed_ids_map := sc.State.FeedRegistry.Entries

	unknown_feed_ids := []model.InternalEncodedAssetId{}
	for _, feedId := range feedIds {
		if _, ok := feed_ids_map[feedId]; !ok {
			unknown_feed_ids = append(unknown_feed_ids, feedId)
		}
	}
	
	resolved_feed_ids := sc.getFeedIds(unknown_feed_ids)
	feed_ids_map = append(feed_ids_map, resolved_feed_ids...)
	feed_ids := []sui_types.SuiAddress{}
	for _, feed_id := range feed_ids_map {
		feed_ids = append(feed_ids, feed_id)
	}

	options := &types.SuiObjectDataOptions{
		ShowContent: true,
	}

	feeds, err := sc.Client.MultiGetObjects(context.Background(), feed_ids, options)
	if err != nil {
		return nil, err
	}

	result := make(map[model.InternalEncodedAssetId]model.InternalStorkStructsTemporalNumericValue)
	
	for _, feed := range feeds {
		var id model.InternalEncodedAssetId
		copy(id[:], feed.Data.Content.Data.MoveObject.Fields.(map[string]interface{})["asset_id"].([]byte))
		var latestValue TemporalNumericValue
		latestValueFields := feed.Data.Content.Data.MoveObject.Fields.(map[string]interface{})["latest_value"]
		
		if err := types.JsonUnmarshal(latestValueFields, &latestValue); err != nil {
			return nil, fmt.Errorf("failed to unmarshal latest value: %w", err)
		}
		
		// Convert negative bool to int
		signInt := 1
		if latestValue.QuantizedValue.Negative {
			signInt = -1
		}
		quantizedValue := latestValue.QuantizedValue.Magnitude.Mul(big.NewInt(int64(signInt)))

		// Convert to the model type
		modelValue := model.InternalStorkStructsTemporalNumericValue{
			TimestampNs: latestValue.TimestampNs,
			QuantizedValue: quantizedValue,
		}
		
		result[id] = modelValue
	}
	
	return result, nil
}

func (sc *StorkContract) UpdateMultipleTemporalNumericValueEvm(updateData []UpdateData) error {
	ptb := sui_types.NewProgrammableTransactionBuilder()

	// fee
	feeArg, err := ptb.Pure(sc.State.SingleUpdateFeeInMist)
	if err != nil {
		return err
	}
	splitCoinResult := ptb.Command(
		sui_types.Command{
			SplitCoins: &struct {
				Argument  sui_types.Argument
				Arguments []sui_types.Argument
			}{
				Argument:  sui_types.Argument{GasCoin: &lib.EmptyEnum,
				Arguments: []sui_types.Argument{feeArg},
			},
		},
	)

	//deconstruct update data into arrays
	ids := [][]byte{}
	temporalNumericValueTimestampNss := []big.Int{}
	temporalNumericValueMagnitudes := []*big.Int{}
	temporalNumericValueNegatives := []bool{}
	publisherMerkleRoots := [][]byte{}
	valueComputeAlgHashes := [][]byte{}
	rs := [][]byte{}
	ss := [][]byte{}
	vs := []byte{}
	for _, update := range updateData {
		ids = append(ids, update.Id)
		temporalNumericValueTimestampNss = append(temporalNumericValueTimestampNss, big.Int(update.TemporalNumericValueTimestampNs))
		temporalNumericValueMagnitudes = append(temporalNumericValueMagnitudes, update.TemporalNumericValueMagnitude)
		temporalNumericValueNegatives = append(temporalNumericValueNegatives, update.TemporalNumericValueNegative)
		publisherMerkleRoots = append(publisherMerkleRoots, update.PublisherMerkleRoot)
		valueComputeAlgHashes = append(valueComputeAlgHashes, update.ValueComputeAlgHash)
		rs = append(rs, update.R)
		ss = append(ss, update.S)
		vs = append(vs, update.V)
	}
	idsArg, err := ptb.Pure(ids)
	if err != nil {
		return err
	}
	timestampNssArg, err := ptb.Pure(temporalNumericValueTimestampNss)
	if err != nil {
		return err
	}
	magnitudesArg, err := ptb.Pure(temporalNumericValueMagnitudes)
	if err != nil {
		return err
	}
	negativesArg, err := ptb.Pure(temporalNumericValueNegatives)
	if err != nil {
		return err
	}
	publisherMerkleRootsArg, err := ptb.Pure(publisherMerkleRoots)
	if err != nil {
		return err
	}
	valueComputeAlgHashesArg, err := ptb.Pure(valueComputeAlgHashes)
	if err != nil {
		return err
	}
	rsArg, err := ptb.Pure(rs)
	if err != nil {
		return err
	}
	ssArg, err := ptb.Pure(ss)
	if err != nil {
		return err
	}
	vsArg, err := ptb.Pure(vs)
	if err != nil {
		return err
	}

	//update_temporal_numeric_value_evm_input_vec::new
	updateTemporalNumericValueEvmInputVec := ptb.Command(
		sui_types.Command{
			MoveCall: &sui_types.ProgrammableMoveCall{
				Package: sc.contract_address,
				Module:  "update_temporal_numeric_value_evm_input_vec",
				Function: "new",
				Arguments: []sui_types.Argument{
					idsArg,
					timestampNssArg,
					magnitudesArg,
					negativesArg,
					publisherMerkleRootsArg,
					valueComputeAlgHashesArg,
					rsArg,
					ssArg,
					vsArg,
				},
			},
		},
	)

	stateArg, err := ptb.Obj(sui_types.ObjectArg{
		SharedObject: &struct {
			Id:                   sc.state.Id,
			InitialSharedVersion: sc.state.Version,
			Mutable:              false,
		},
	})
	if err != nil {
		return err
	}

	//stork::update_multiple_temporal_numeric_values_evm
	ptb.Command(
		sui_types.Command{
			MoveCall: &sui_types.ProgrammableMoveCall{
				Package: sc.contract_address,
				Module:  "stork",
				Function: "update_multiple_temporal_numeric_values_evm",
				Arguments: []sui_types.Argument{
					stateArg,
					updateTemporalNumericValueEvmInputVec,
					splitCoinResult,
				},
			},
		},
	)

	pt := ptb.Finish()
	gasBudget, err := sc.getGasBudgetFromDryRun(pt)
	if err != nil {
		return err
	}
	totalFeeAmount := sc.state.SingleUpdateFeeInMist * uint64(len(updateData))

	coins, err := sc.client.GetCoins(context.Background(), sc.account.Address, nil, nil, 100)[0]
	if err != nil {
		return err
	}
	
	pickedCoins, err := types.PickupCoins(
		coins,
		*big.NewInt(int64(totalFeeAmount)),
		gasBudget,
		100,
		10,
	)
	if err != nil {
		return err
	}
	tx := sui_types.NewProgrammable(
		sc.account.Address,
		pickedCoins,
		pt,
		gasBudget,
		sc.reference_gas_price,
	)
	return sc.client.ExecuteTransactionBlock(context.Background(), tx, sc.account)
}

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
	version := object.Data.Version.Uint64()

	// registry
	state_dynamic_fields, err := client.GetDynamicFields(context.Background(), *stork_state_id, nil, nil)
	if err != nil {
		return StorkState{}, err
	}
	registry_id := sui_types.SuiAddress{}
	for _, dynamic_field := range state_dynamic_fields.Data {
		if bytes.Equal(dynamic_field.Name.Value, []byte("temporal_numeric_value_feed_registry")) {
			registry_id = dynamic_field.Value.ObjectID
			break
		}
	asset_}

	if registry_id == sui_types.SuiAddress{} {
		return StorkState{}, errors.New("feed registry not found")
	}

	feed_ids := make(map[model.InternalEncodedAssetId]sui_types.SuiAddress)

	registry := FeedRegistry{Id: registry_id, Entries: feed_ids}

	return StorkState{Id: *stork_state_id, StorkEvmPublicKey: stork_evm_public_key, StorkSuiPublicKey: *stork_sui_public_key, SingleUpdateFeeInMist: single_update_fee_in_mist, Version: version, FeedRegistry: registry}, nil
}

func (sc *StorkContract) GetFeedIds() (map[model.InternalEncodedAssetId]sui_types.SuiAddress, error) {

}

func (sc *StorkContract) getGasBudgetFromDryRun(pt *sui_types.ProgrammableTransaction) (uint64, error) {
	tx := sui_types.NewProgrammable(
		sc.account.Address,
		nil,
		pt,
		sc.reference_gas_price,
	)

	txBytes, err := bcs.Marshal(tx)
	if err != nil {	
		return 0, err
	}
	dryRunResult, err := sc.client.DryRunTransaction(context.Background(), txBytes)
	if err != nil {
        return nil, fmt.Errorf("dry run failed: %w", err)
    }
    
    if !dryRunResult.Effects.Data.IsSuccess() {
        return nil, fmt.Errorf("dry run failed: %s", dryRunResult.Effects.V1.Status.Error)
    }

	gasUsed := dryRunResult.Effects.GasFee()

	return gasUsed, nil
}

func (sc *StorkContract) getFeedIds(feedIds []model.InternalEncodedAssetId) (map[model.InternalEncodedAssetId]sui_types.SuiAddress, error) {
	feed_ids := make(map[model.InternalEncodedAssetId]sui_types.SuiAddress)
	
	registry_id := sc.State.FeedRegistry.Id

	registry_entries, err := sc.Client.GetDynamicFields(context.Background(), registry_id, nil, nil)
	if err != nil {
		return feed_ids, err
	}

	
	for _, feedId := range feedIds {
		for _, entry := range registry_entries.Data {
			if bytes.Equal(entry.Name.Value.Bytes, feedId) {
				feed_object_id := entry.Value.ObjectID
				feed_ids[feedId] = feed_object_id
				break
			}
		}
	}

	return feed_ids, nil
}
