// Unlike the EVM and Solana bindings, the Sui bindings are not generated from the Move source code, as a tool for this does not currently exist.
// Instead, this file contains utility functions for interacting with the Sui Stork contract.
// These functions are written using https://github.com/block-vision/sui-go-sdk

package contract_bindings_sui

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strconv"

	"github.com/coming-chat/go-sui/v2/account"
	"github.com/coming-chat/go-sui/v2/client"
	"github.com/coming-chat/go-sui/v2/lib"
	"github.com/coming-chat/go-sui/v2/sui_types"
	"github.com/coming-chat/go-sui/v2/types"
	"github.com/fardream/go-bcs/bcs"
)

type StorkContract struct {
	Client            *client.Client
	Account           *account.Account
	ContractAddress   sui_types.SuiAddress
	State             StorkState
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
	InitialSharedVersion  uint64
}

type FeedRegistry struct {
	Id      sui_types.SuiAddress
	Entries map[EncodedAssetId]sui_types.SuiAddress
}

type TemporalNumericValue struct {
	TimestampNs    uint64
	QuantizedValue I128
}

type EncodedAssetId [32]byte

type I128 struct {
	Magnitude *big.Int
	Negative  bool
}

type U128 struct {
	Value []byte
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

	referenceGasPriceResult, err := client.GetReferenceGasPrice(context.Background())
	if err != nil {
		return nil, err
	}
	referenceGasPrice := referenceGasPriceResult.Uint64()

	return &StorkContract{Client: client, Account: account, ContractAddress: *contractAddr, State: state, ReferenceGasPrice: referenceGasPrice}, nil
}

func (sc *StorkContract) SetReferenceGasPrice(referenceGasPrice uint64) {
	sc.ReferenceGasPrice = referenceGasPrice
}

// gets multiple temporal numeric values at a time to save on RPC calls
func (sc *StorkContract) GetMultipleTemporalNumericValuesUnchecked(feedIds []EncodedAssetId) (map[EncodedAssetId]TemporalNumericValue, error) {
	feedIdsMap := sc.State.FeedRegistry.Entries

	unknownFeedIDs := []EncodedAssetId{}
	for _, feedId := range feedIds {
		if _, ok := feedIdsMap[feedId]; !ok {
			unknownFeedIDs = append(unknownFeedIDs, feedId)
		}
	}

	resolvedFeedIDs, err := sc.getFeedIds(unknownFeedIDs)
	if err != nil {
		return nil, err
	}
	// feedIDsMap = append(feedIdsMap, resolvedFeedIDs...)
	for feedId, feedObjectId := range resolvedFeedIDs {
		feedIdsMap[feedId] = feedObjectId
	}
	feedIDs := []sui_types.SuiAddress{}
	for _, feedID := range feedIdsMap {
		feedIDs = append(feedIDs, feedID)
	}

	options := &types.SuiObjectDataOptions{
		ShowContent: true,
	}

	feeds, err := sc.Client.MultiGetObjects(context.Background(), feedIDs, options)
	if err != nil {
		return nil, err
	}

	result := make(map[EncodedAssetId]TemporalNumericValue)

	for _, feed := range feeds {
		var id EncodedAssetId
		slice := feed.Data.Content.Data.MoveObject.Fields.(map[string]interface{})["asset_id"].(map[string]interface{})["fields"].(map[string]interface{})["bytes"]
		idBytes, err := interfaceSliceToBytes(slice)
		if err != nil {
			return nil, err
		}
		copy(id[:], idBytes)

		latestValueFields := feed.Data.Content.Data.MoveObject.Fields.(map[string]interface{})["latest_value"].(map[string]interface{})["fields"]
		timestampNs, err := strconv.ParseUint(latestValueFields.(map[string]interface{})["timestamp_ns"].(string), 10, 64)
		if err != nil {
			return nil, err
		}
		magnitude := big.Int{}
		temp := latestValueFields.(map[string]interface{})["quantized_value"].(map[string]interface{})["fields"].(map[string]interface{})["magnitude"].(string)
		magnitude.SetString(temp, 10)
		negative := latestValueFields.(map[string]interface{})["quantized_value"].(map[string]interface{})["fields"].(map[string]interface{})["negative"].(bool)

		quantizedValue := I128{
			Magnitude: &magnitude,
			Negative:  negative,
		}
		latestValue := TemporalNumericValue{
			TimestampNs:    timestampNs,
			QuantizedValue: quantizedValue,
		}
		result[id] = latestValue
	}
	return result, nil
}

func (sc *StorkContract) UpdateMultipleTemporalNumericValuesEvm(updateData []UpdateData) error {
	ptb := sui_types.NewProgrammableTransactionBuilder()

	// fee
	totalFeeAmount := sc.State.SingleUpdateFeeInMist * uint64(len(updateData))
	address, err := sui_types.NewAddressFromHex(sc.Account.Address)
	if err != nil {
		return err
	}
	feeArg, err := ptb.Pure(totalFeeAmount)
	if err != nil {
		return err
	}
	splitCoinResult := ptb.Command(
		sui_types.Command{
			SplitCoins: &struct {
				Argument  sui_types.Argument
				Arguments []sui_types.Argument
			}{
				Argument:  sui_types.Argument{GasCoin: &lib.EmptyEnum{}},
				Arguments: []sui_types.Argument{feeArg},
			},
		},
	)

	//deconstruct update data into arrays
	ids := [][]byte{}
	temporalNumericValueTimestampNss := []uint64{}
	temporalNumericValueMagnitudes := []*big.Int{}
	temporalNumericValueNegatives := []bool{}
	publisherMerkleRoots := [][]byte{}
	valueComputeAlgHashes := [][]byte{}
	rs := [][]byte{}
	ss := [][]byte{}
	vs := []byte{}
	for _, update := range updateData {
		ids = append(ids, update.Id)
		temporalNumericValueTimestampNss = append(temporalNumericValueTimestampNss, update.TemporalNumericValueTimestampNs)
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

	magnitudeBytes := make([]bcs.Uint128, len(temporalNumericValueMagnitudes))
	for i, magnitude := range temporalNumericValueMagnitudes {
		u128val, err := bcs.NewUint128FromBigInt(magnitude)
		if err != nil {
			return err
		}
		magnitudeBytes[i] = *u128val
	}

	magnitudesArg, err := ptb.Pure(magnitudeBytes)
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
				Package:  sc.ContractAddress,
				Module:   "update_temporal_numeric_value_evm_input_vec",
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
			Id                   sui_types.SuiAddress
			InitialSharedVersion uint64
			Mutable              bool
		}{
			Id:                   sc.State.Id,
			InitialSharedVersion: sc.State.InitialSharedVersion,
			Mutable:              true,
		},
	})
	if err != nil {
		return err
	}

	// stork::update_multiple_temporal_numeric_values_evm
	ptb.Command(
		sui_types.Command{
			MoveCall: &sui_types.ProgrammableMoveCall{
				Package:  sc.ContractAddress,
				Module:   "stork",
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

	coins, err := sc.Client.GetCoins(context.Background(), *address, nil, nil, 100)
	if err != nil {
		return err
	}

	gasBudget, err := sc.getGasBudgetFromDryRun(&pt)
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
		*address,
		pickedCoins.CoinRefs(),
		pt,
		gasBudget,
		sc.ReferenceGasPrice,
	)
	txBytes, err := bcs.Marshal(tx)
	if err != nil {
		return err
	}

	signatures, err := sc.Account.SignSecureWithoutEncode(txBytes, sui_types.DefaultIntent())
	if err != nil {
		return err
	}
	_, err = sc.Client.ExecuteTransactionBlock(context.Background(), txBytes, []any{signatures}, nil, types.TxnRequestTypeWaitForEffectsCert)
	if err != nil {
		return err
	}
	return nil
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
	limit := uint(1)
	event, err := client.QueryEvents(context.Background(), eventFilter, nil, &limit, false)
	if err != nil {
		return StorkState{}, err
	}

	storkStateIdHex := event.Data[0].ParsedJson.(map[string]interface{})["stork_state_id"].(string)
	storkStateId, err := sui_types.NewAddressFromHex(storkStateIdHex)
	if err != nil {
		return StorkState{}, err
	}

	options := &types.SuiObjectDataOptions{
		ShowContent: true,
		ShowOwner:   true,
	}
	object, err := client.GetObject(context.Background(), *storkStateId, options)
	if err != nil {
		return StorkState{}, err
	}
	fields := object.Data.Content.Data.MoveObject.Fields.(map[string]interface{})
	storkEvmPublicKeyBytes := fields["stork_evm_public_key"].(map[string]interface{})["fields"].(map[string]interface{})["bytes"]
	byteSlice, err := interfaceSliceToBytes(storkEvmPublicKeyBytes)
	if err != nil {
		return StorkState{}, fmt.Errorf("failed to convert public key bytes: %w", err)
	}

	storkEvmPublicKey := hex.EncodeToString(byteSlice)
	storkSuiPublicKeyString := fields["stork_sui_public_key"].(string)
	storkSuiPublicKey, err := sui_types.NewAddressFromHex(storkSuiPublicKeyString)
	if err != nil {
		return StorkState{}, err
	}
	singleUpdateFeeInMist, err := strconv.ParseUint(fields["single_update_fee_in_mist"].(string), 10, 64)
	if err != nil {
		return StorkState{}, err
	}
	version := object.Data.Version.Uint64()
	initialSharedVersion := uint64(*object.Data.Owner.Shared.InitialSharedVersion)
	// registry
	stateDynamicFields, err := client.GetDynamicFields(context.Background(), *storkStateId, nil, nil)
	if err != nil {
		return StorkState{}, err
	}
	registryId := sui_types.SuiAddress{}
	for _, dynamicField := range stateDynamicFields.Data {
		nameBytes, err := interfaceSliceToBytes(dynamicField.Name.Value)
		if err != nil {
			return StorkState{}, fmt.Errorf("failed to convert name bytes: %w", err)
		}
		if bytes.Equal(nameBytes, []byte("temporal_numeric_value_feed_registry")) {
			registryId = dynamicField.ObjectId
			break
		}
	}

	if registryId == (sui_types.SuiAddress{}) {
		return StorkState{}, errors.New("feed registry not found")
	}

	feedIds := make(map[EncodedAssetId]sui_types.SuiAddress)

	registry := FeedRegistry{Id: registryId, Entries: feedIds}

	return StorkState{Id: *storkStateId, StorkEvmPublicKey: storkEvmPublicKey, StorkSuiPublicKey: *storkSuiPublicKey, SingleUpdateFeeInMist: singleUpdateFeeInMist, Version: version, InitialSharedVersion: initialSharedVersion, FeedRegistry: registry}, nil
}

func (sc *StorkContract) getGasBudgetFromDryRun(pt *sui_types.ProgrammableTransaction) (uint64, error) {
	address, err := sui_types.NewAddressFromHex(sc.Account.Address)
	if err != nil {
		return 0, err
	}
	tx := sui_types.NewProgrammable(
		*address,
		nil,
		*pt,
		uint64(10e9),
		sc.ReferenceGasPrice,
	)

	txBytes, err := bcs.Marshal(tx)
	if err != nil {
		return 0, err
	}
	dryRunResult, err := sc.Client.DryRunTransaction(context.Background(), txBytes)
	if err != nil {
		return 0, fmt.Errorf("dry run failed: %w", err)
	}

	if !dryRunResult.Effects.Data.IsSuccess() {
		return 0, fmt.Errorf("dry run failed: %s", dryRunResult.Effects.Data.V1.Status.Error)
	}

	gasUsed := dryRunResult.Effects.Data.GasFee()

	return uint64(gasUsed), nil
}

func (sc *StorkContract) getFeedIds(feedIds []EncodedAssetId) (map[EncodedAssetId]sui_types.SuiAddress, error) {
	feedIdsMap := make(map[EncodedAssetId]sui_types.SuiAddress)

	registryId := sc.State.FeedRegistry.Id

	registryEntries, err := sc.Client.GetDynamicFields(context.Background(), registryId, nil, nil)
	if err != nil {
		return feedIdsMap, err
	}

	for _, feedId := range feedIds {
		for _, entry := range registryEntries.Data {
			nameBytes, err := interfaceSliceToBytes(entry.Name.Value.(map[string]interface{})["bytes"])
			if err != nil {
				return nil, fmt.Errorf("failed to convert name2 bytes: %w", err)
			}
			if bytes.Equal(nameBytes, feedId[:]) {
				feedObjectId := entry.ObjectId
				feedIdsMap[feedId] = feedObjectId
				break
			}
		}
	}

	return feedIdsMap, nil
}

func interfaceSliceToBytes(slice interface{}) ([]byte, error) {
	interfaceSlice, ok := slice.([]interface{})
	if !ok {
		return nil, fmt.Errorf("input is not a slice of interfaces, but a %T", slice)
	}

	byteSlice := make([]byte, len(interfaceSlice))
	for i, v := range interfaceSlice {
		floatVal, ok := v.(float64)
		if !ok {
			return nil, fmt.Errorf("element at index %d is not a float64", i)
		}
		byteSlice[i] = byte(floatVal)
	}
	return byteSlice, nil
}
