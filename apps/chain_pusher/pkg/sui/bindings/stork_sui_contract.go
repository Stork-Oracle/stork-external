// These binding are not generated.
// Instead, this file contains utility functions for interacting with the Sui Stork contract.

package bindings

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strconv"

	"github.com/Stork-Oracle/stork-external/apps/chain_pusher/pkg/pusher"
	"github.com/coming-chat/go-sui/v2/account"
	sui_client "github.com/coming-chat/go-sui/v2/client"
	"github.com/coming-chat/go-sui/v2/lib"
	"github.com/coming-chat/go-sui/v2/sui_types"
	"github.com/coming-chat/go-sui/v2/types"
	"github.com/fardream/go-bcs/bcs"
)

var (
	ErrFeedRegistryNotFound = errors.New("feed registry not found")
	ErrFieldNotFound        = errors.New("field not found")
	ErrWrongType            = errors.New("wrong type")
)

type StorkContract struct {
	Client          *sui_client.Client
	Account         *account.Account
	ContractAddress sui_types.SuiAddress
	State           StorkState
}

type MultipleUpdateData struct {
	IDs                              [][]byte
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
	ID                              []byte
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
	ID                    sui_types.SuiAddress
	StorkSuiPublicKey     sui_types.SuiAddress
	StorkEvmPublicKey     string
	SingleUpdateFeeInMist uint64
	Version               uint64
	FeedRegistry          FeedRegistry
	InitialSharedVersion  uint64
}

type FeedRegistry struct {
	ID      sui_types.SuiAddress
	Entries map[EncodedAssetID]sui_types.SuiAddress
}

type TemporalNumericValue struct {
	TimestampNs    uint64
	QuantizedValue I128
}

type EncodedAssetID [32]byte

type I128 struct {
	Magnitude *big.Int
	Negative  bool
}

type U128 struct {
	Value []byte
}

func NewStorkContract(rpcUrl string, contractAddress string, account *account.Account) (*StorkContract, error) {
	client, err := sui_client.Dial(rpcUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to dial Sui client: %w", err)
	}

	contractAddr, err := sui_types.NewAddressFromHex(contractAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to convert contract address to Sui address: %w", err)
	}

	state, err := getStorkState(*contractAddr, client)
	if err != nil {
		return nil, err
	}

	return &StorkContract{Client: client, Account: account, ContractAddress: *contractAddr, State: state}, nil
}

// GetMultipleTemporalNumericValuesUnchecked gets multiple temporal numeric values at a time for efficiency.
//

func (sc *StorkContract) GetMultipleTemporalNumericValuesUnchecked(
	feedIDs []EncodedAssetID,
) (map[EncodedAssetID]TemporalNumericValue, error) {
	feedIDsMap := sc.State.FeedRegistry.Entries

	unknownFeedIDs := []EncodedAssetID{}

	for _, feedID := range feedIDs {
		if _, ok := feedIDsMap[feedID]; !ok {
			unknownFeedIDs = append(unknownFeedIDs, feedID)
		}
	}

	resolvedFeedIDs, err := sc.getFeedIDs(unknownFeedIDs)
	if err != nil {
		return nil, err
	}

	for feedID, feedObjectID := range resolvedFeedIDs {
		feedIDsMap[feedID] = feedObjectID
	}

	feedAddresses := []sui_types.SuiAddress{}
	for _, feedID := range feedIDsMap {
		feedAddresses = append(feedAddresses, feedID)
	}

	options := &types.SuiObjectDataOptions{
		ShowContent: true,
	}

	feeds, err := sc.Client.MultiGetObjects(context.Background(), feedAddresses, options)
	if err != nil {
		return nil, fmt.Errorf("failed to get feed objects: %w", err)
	}

	result := make(map[EncodedAssetID]TemporalNumericValue)

	var (
		id    EncodedAssetID
		value TemporalNumericValue
	)

	for _, feed := range feeds {
		id, value, err = parseFeedToTemporalNumericValue(feed)
		if err != nil {
			return nil, err
		}

		result[id] = value
	}

	return result, nil
}

//nolint:cyclop,funlen // This is a long and complex function but does clearly related work.
func parseFeedToTemporalNumericValue(feed types.SuiObjectResponse) (EncodedAssetID, TemporalNumericValue, error) {
	var id EncodedAssetID

	// Extract asset_id
	fields, ok := feed.Data.Content.Data.MoveObject.Fields.(map[string]interface{})
	if !ok {
		return id, TemporalNumericValue{}, fmt.Errorf("failed to get fields: %w", ErrFieldNotFound)
	}

	assetID, exists := fields["asset_id"]
	if !exists {
		return id, TemporalNumericValue{}, fmt.Errorf("asset_id field not found: %w", ErrFieldNotFound)
	}

	assetIDMap, ok := assetID.(map[string]interface{})
	if !ok {
		return id, TemporalNumericValue{}, fmt.Errorf("asset_id is not a map: %w", ErrFieldNotFound)
	}

	assetIDFields, exists := assetIDMap["fields"]
	if !exists {
		return id, TemporalNumericValue{}, fmt.Errorf("fields not found in asset_id: %w", ErrFieldNotFound)
	}

	assetIDFieldsMap, ok := assetIDFields.(map[string]interface{})
	if !ok {
		return id, TemporalNumericValue{}, fmt.Errorf("asset_id fields is not a map: %w", ErrFieldNotFound)
	}

	bytesField, exists := assetIDFieldsMap["bytes"]
	if !exists {
		return id, TemporalNumericValue{}, fmt.Errorf("bytes not found in asset_id fields: %w", ErrFieldNotFound)
	}

	idBytes, err := interfaceSliceToBytes(bytesField)
	if err != nil {
		return id, TemporalNumericValue{}, err
	}

	copy(id[:], idBytes)

	// Extract latest_value
	latestValue, exists := fields["latest_value"]
	if !exists {
		return id, TemporalNumericValue{}, fmt.Errorf("latest_value field not found: %w", ErrFieldNotFound)
	}

	latestValueMap, ok := latestValue.(map[string]interface{})
	if !ok {
		return id, TemporalNumericValue{}, fmt.Errorf("latest_value is not a map: %w", ErrFieldNotFound)
	}

	latestValueFields, exists := latestValueMap["fields"]
	if !exists {
		return id, TemporalNumericValue{}, fmt.Errorf("fields not found in latest_value: %w", ErrFieldNotFound)
	}

	latestValueFieldsMap, ok := latestValueFields.(map[string]interface{})
	if !ok {
		return id, TemporalNumericValue{}, fmt.Errorf("latest_value fields is not a map: %w", ErrFieldNotFound)
	}

	// Extract timestamp_ns
	timestampNsField, exists := latestValueFieldsMap["timestamp_ns"]
	if !exists {
		return id, TemporalNumericValue{}, fmt.Errorf("timestamp_ns not found: %w", ErrFieldNotFound)
	}

	timestampNsStr, ok := timestampNsField.(string)
	if !ok {
		return id, TemporalNumericValue{}, fmt.Errorf("timestamp_ns is not a string: %w", ErrFieldNotFound)
	}

	timestampNs, err := strconv.ParseUint(timestampNsStr, 10, 64)
	if err != nil {
		return id, TemporalNumericValue{}, fmt.Errorf("failed to parse timestamp ns: %w", err)
	}

	// Extract quantized_value
	quantizedValueField, exists := latestValueFieldsMap["quantized_value"]
	if !exists {
		return id, TemporalNumericValue{}, fmt.Errorf("quantized_value not found: %w", ErrFieldNotFound)
	}

	quantizedValueMap, ok := quantizedValueField.(map[string]interface{})
	if !ok {
		return id, TemporalNumericValue{}, fmt.Errorf("quantized_value is not a map: %w", ErrFieldNotFound)
	}

	quantizedValueFields, exists := quantizedValueMap["fields"]
	if !exists {
		return id, TemporalNumericValue{}, fmt.Errorf("fields not found in quantized_value: %w", ErrFieldNotFound)
	}

	quantizedValueFieldsMap, ok := quantizedValueFields.(map[string]interface{})
	if !ok {
		return id, TemporalNumericValue{}, fmt.Errorf("quantized_value fields is not a map: %w", ErrFieldNotFound)
	}

	// Extract magnitude
	magnitudeField, exists := quantizedValueFieldsMap["magnitude"]
	if !exists {
		return id, TemporalNumericValue{}, fmt.Errorf("magnitude not found: %w", ErrFieldNotFound)
	}

	magnitudeStr, ok := magnitudeField.(string)
	if !ok {
		return id, TemporalNumericValue{}, fmt.Errorf("magnitude is not a string: %w", ErrFieldNotFound)
	}

	magnitude := big.Int{}
	//nolint:mnd // Base number
	magnitude.SetString(magnitudeStr, 10)

	// Extract negative
	negativeField, exists := quantizedValueFieldsMap["negative"]
	if !exists {
		return id, TemporalNumericValue{}, fmt.Errorf("negative not found: %w", ErrFieldNotFound)
	}

	negative, ok := negativeField.(bool)
	if !ok {
		return id, TemporalNumericValue{}, fmt.Errorf("negative is not a bool: %w", ErrFieldNotFound)
	}

	quantizedValue := I128{
		Magnitude: &magnitude,
		Negative:  negative,
	}

	value := TemporalNumericValue{
		TimestampNs:    timestampNs,
		QuantizedValue: quantizedValue,
	}

	return id, value, nil
}

//nolint:cyclop,funlen,maintidx // This is a long and complex function but does related work.
func (sc *StorkContract) UpdateMultipleTemporalNumericValuesEvm(updateData []UpdateData) (string, error) {
	ptb := sui_types.NewProgrammableTransactionBuilder()

	// get reference gas price
	referenceGasPrice, err := sc.getReferenceGasPrice()
	if err != nil {
		return "", err
	}

	// fee
	totalFeeAmount := sc.State.SingleUpdateFeeInMist * uint64(len(updateData))

	address, err := sui_types.NewAddressFromHex(sc.Account.Address)
	if err != nil {
		return "", fmt.Errorf("failed to get address from hex: %w", err)
	}

	feeArg, err := ptb.Pure(totalFeeAmount)
	if err != nil {
		return "", fmt.Errorf("failed to create pure field for total fee amount: %w", err)
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

	// deconstruct update data into arrays
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
		ids = append(ids, update.ID)
		temporalNumericValueTimestampNss = append(
			temporalNumericValueTimestampNss,
			update.TemporalNumericValueTimestampNs,
		)
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
		return "", fmt.Errorf("failed to create pure field for ids: %w", err)
	}

	timestampNssArg, err := ptb.Pure(temporalNumericValueTimestampNss)
	if err != nil {
		return "", fmt.Errorf("failed to create pure field for timestamp nss: %w", err)
	}

	magnitudeBytes := make([]bcs.Uint128, len(temporalNumericValueMagnitudes))

	var u128val *bcs.Uint128

	for i, magnitude := range temporalNumericValueMagnitudes {
		u128val, err = bcs.NewUint128FromBigInt(magnitude)
		if err != nil {
			return "", fmt.Errorf("failed to create uint128 from big int: %w", err)
		}

		magnitudeBytes[i] = *u128val
	}

	magnitudesArg, err := ptb.Pure(magnitudeBytes)
	if err != nil {
		return "", fmt.Errorf("failed to create pure field for magnitudes: %w", err)
	}

	negativesArg, err := ptb.Pure(temporalNumericValueNegatives)
	if err != nil {
		return "", fmt.Errorf("failed to create pure field for negatives: %w", err)
	}

	publisherMerkleRootsArg, err := ptb.Pure(publisherMerkleRoots)
	if err != nil {
		return "", fmt.Errorf("failed to create pure field for publisher merkle roots: %w", err)
	}

	valueComputeAlgHashesArg, err := ptb.Pure(valueComputeAlgHashes)
	if err != nil {
		return "", fmt.Errorf("failed to create pure field for value compute alg hashes: %w", err)
	}

	rsArg, err := ptb.Pure(rs)
	if err != nil {
		return "", fmt.Errorf("failed to create pure field for rs: %w", err)
	}

	ssArg, err := ptb.Pure(ss)
	if err != nil {
		return "", fmt.Errorf("failed to create pure field for ss: %w", err)
	}

	vsArg, err := ptb.Pure(vs)
	if err != nil {
		return "", fmt.Errorf("failed to create pure field for vs: %w", err)
	}

	// update_temporal_numeric_value_evm_input_vec::new
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
			//nolint:revive // Forced by anonymous struct in object arg
			Id                   sui_types.SuiAddress
			InitialSharedVersion uint64
			Mutable              bool
		}{
			Id:                   sc.State.ID,
			InitialSharedVersion: sc.State.InitialSharedVersion,
			Mutable:              true,
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to create object: %w", err)
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

	//nolint:mnd // 100 is being used as a arbitrarily large limit and thus a permissible magic number
	coins, err := sc.Client.GetCoins(context.Background(), *address, nil, nil, 100)
	if err != nil {
		return "", fmt.Errorf("failed to get coins: %w", err)
	}

	gasBudget, err := sc.getGasBudgetFromDryRun(&pt, referenceGasPrice)
	if err != nil {
		return "", err
	}

	totalFeeAmountInt64, err := pusher.SafeUint64ToInt64(totalFeeAmount)
	if err != nil {
		return "", fmt.Errorf("failed to convert total fee amount to int64: %w", err)
	}

	pickedCoins, err := types.PickupCoins(
		coins,
		*big.NewInt(totalFeeAmountInt64),
		gasBudget,
		0,
		0,
	)
	if err != nil {
		return "", fmt.Errorf("failed to pick up coins: %w", err)
	}

	tx := sui_types.NewProgrammable(
		*address,
		pickedCoins.CoinRefs(),
		pt,
		gasBudget,
		referenceGasPrice,
	)

	txBytes, err := bcs.Marshal(tx)
	if err != nil {
		return "", fmt.Errorf("failed to marshal transaction: %w", err)
	}

	signatures, err := sc.Account.SignSecureWithoutEncode(txBytes, sui_types.DefaultIntent())
	if err != nil {
		return "", fmt.Errorf("failed to sign transaction: %w", err)
	}

	txResponse, err := sc.Client.ExecuteTransactionBlock(
		context.Background(),
		txBytes,
		[]any{signatures},
		nil,
		types.TxnRequestTypeWaitForEffectsCert,
	)
	if err != nil {
		return "", fmt.Errorf("failed to execute transaction block: %w", err)
	}

	digest := txResponse.Digest.String()

	return digest, nil
}

func getOriginalContractAddress(
	contractAddress sui_types.SuiAddress,
	client *sui_client.Client,
) (sui_types.SuiAddress, error) {
	method := sui_client.SuiMethod("getNormalizedMoveModulesByPackage")

	var result interface{}

	err := client.CallContext(
		context.Background(),
		&result,
		method, // This is the constant defined in method.go
		contractAddress,
	)
	if err != nil {
		return sui_types.SuiAddress{}, fmt.Errorf("failed to get address from hex: %w", err)
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		return sui_types.SuiAddress{}, fmt.Errorf("result is not a map: %w", ErrFieldNotFound)
	}

	adminField, exists := resultMap["admin"]
	if !exists {
		return sui_types.SuiAddress{}, fmt.Errorf("admin field not found: %w", ErrFieldNotFound)
	}

	adminMap, ok := adminField.(map[string]interface{})
	if !ok {
		return sui_types.SuiAddress{}, fmt.Errorf("admin is not a map: %w", ErrFieldNotFound)
	}

	addressField, exists := adminMap["address"]
	if !exists {
		return sui_types.SuiAddress{}, fmt.Errorf("address field not found: %w", ErrFieldNotFound)
	}

	addressString, ok := addressField.(string)
	if !ok {
		return sui_types.SuiAddress{}, fmt.Errorf("address is not a string: %w", ErrFieldNotFound)
	}

	address, err := sui_types.NewAddressFromHex(addressString)
	if err != nil {
		return sui_types.SuiAddress{}, fmt.Errorf("failed to get address from hex: %w", err)
	}

	return *address, nil
}

//nolint:cyclop,funlen // This is a long and complex function due to interface destructuring
func getStorkState(contractAddress sui_types.SuiAddress, client *sui_client.Client) (StorkState, error) {
	originalContractAddress, err := getOriginalContractAddress(contractAddress, client)
	if err != nil {
		return StorkState{}, err
	}

	eventFilter := types.EventFilter{
		MoveModule: &struct {
			Package sui_types.ObjectID `json:"package"`
			Module  string             `json:"module"`
		}{
			Package: originalContractAddress,
			Module:  "stork",
		},
	}
	limit := uint(1)

	event, err := client.QueryEvents(context.Background(), eventFilter, nil, &limit, false)
	if err != nil {
		return StorkState{}, fmt.Errorf("failed to query events: %w", err)
	}

	parsedJSON, ok := event.Data[0].ParsedJson.(map[string]interface{})
	if !ok {
		return StorkState{}, fmt.Errorf("parsed json is not a map: %w", ErrWrongType)
	}

	storkStateIDField, exists := parsedJSON["stork_state_id"]
	if !exists {
		return StorkState{}, fmt.Errorf("stork_state_id field not found: %w", ErrFieldNotFound)
	}

	storkStateIDHex, ok := storkStateIDField.(string)
	if !ok {
		return StorkState{}, fmt.Errorf("stork_state_id is not a string: %w", ErrWrongType)
	}

	storkStateID, err := sui_types.NewAddressFromHex(storkStateIDHex)
	if err != nil {
		return StorkState{}, fmt.Errorf("failed to get address from hex: %w", err)
	}

	options := &types.SuiObjectDataOptions{
		ShowContent: true,
		ShowOwner:   true,
	}

	object, err := client.GetObject(context.Background(), *storkStateID, options)
	if err != nil {
		return StorkState{}, fmt.Errorf("failed to get object: %w", err)
	}

	fields, ok := object.Data.Content.Data.MoveObject.Fields.(map[string]interface{})
	if !ok {
		return StorkState{}, fmt.Errorf("fields is not a map: %w", ErrWrongType)
	}

	storkEvmPublicKeyField, exists := fields["stork_evm_public_key"]
	if !exists {
		return StorkState{}, fmt.Errorf("stork evm public key field not found: %w", ErrFieldNotFound)
	}

	storkEvmPublicKeyMap, ok := storkEvmPublicKeyField.(map[string]interface{})
	if !ok {
		return StorkState{}, fmt.Errorf("stork evm public key is not a map: %w", ErrWrongType)
	}

	storkEvmPublicKeyFields, exists := storkEvmPublicKeyMap["fields"]
	if !exists {
		return StorkState{}, fmt.Errorf("stork evm public key fields field not found: %w", ErrFieldNotFound)
	}

	fieldsMap, ok := storkEvmPublicKeyFields.(map[string]interface{})
	if !ok {
		return StorkState{}, fmt.Errorf("stork evm public key fields is not a map: %w", ErrWrongType)
	}

	fieldsMapBytesField, exists := fieldsMap["bytes"]
	if !exists {
		return StorkState{}, fmt.Errorf("stork evm public key fields map bytes field not found: %w", ErrFieldNotFound)
	}

	byteSlice, err := interfaceSliceToBytes(fieldsMapBytesField)
	if err != nil {
		return StorkState{}, fmt.Errorf("failed to convert public key bytes: %w", err)
	}

	storkEvmPublicKey := hex.EncodeToString(byteSlice)

	storkSuiPublicKeyString, ok := fields["stork_sui_address"].(string)
	if !ok {
		return StorkState{}, fmt.Errorf("stork sui address is not a string: %w", ErrWrongType)
	}

	storkSuiPublicKey, err := sui_types.NewAddressFromHex(storkSuiPublicKeyString)
	if err != nil {
		return StorkState{}, fmt.Errorf("failed to get address from hex: %w", err)
	}

	singleUpdateFeeInMistField, exists := fields["single_update_fee_in_mist"]
	if !exists {
		return StorkState{}, fmt.Errorf("single update fee in mist field not found: %w", ErrFieldNotFound)
	}

	singleUpdateFeeInMistString, ok := singleUpdateFeeInMistField.(string)
	if !ok {
		return StorkState{}, fmt.Errorf("single update fee in mist is not a string: %w", ErrWrongType)
	}

	singleUpdateFeeInMist, err := strconv.ParseUint(singleUpdateFeeInMistString, 10, 64)
	if err != nil {
		return StorkState{}, fmt.Errorf("failed to parse single update fee in mist: %w", err)
	}

	version := object.Data.Version.Uint64()
	initialSharedVersion := *object.Data.Owner.Shared.InitialSharedVersion
	// registry
	stateDynamicFields, err := client.GetDynamicFields(context.Background(), *storkStateID, nil, nil)
	if err != nil {
		return StorkState{}, fmt.Errorf("failed to get dynamic fields: %w", err)
	}

	registryID := sui_types.SuiAddress{}

	for _, dynamicField := range stateDynamicFields.Data {
		var nameBytes []byte

		nameBytes, err = interfaceSliceToBytes(dynamicField.Name.Value)
		if err != nil {
			return StorkState{}, fmt.Errorf("failed to convert name bytes: %w", err)
		}

		if bytes.Equal(nameBytes, []byte("temporal_numeric_value_feed_registry")) {
			registryID = dynamicField.ObjectId

			break
		}
	}

	if registryID == (sui_types.SuiAddress{}) {
		return StorkState{}, ErrFeedRegistryNotFound
	}

	feedIDs := make(map[EncodedAssetID]sui_types.SuiAddress)

	registry := FeedRegistry{ID: registryID, Entries: feedIDs}

	return StorkState{
		ID:                    *storkStateID,
		StorkEvmPublicKey:     storkEvmPublicKey,
		StorkSuiPublicKey:     *storkSuiPublicKey,
		SingleUpdateFeeInMist: singleUpdateFeeInMist,
		Version:               version,
		InitialSharedVersion:  initialSharedVersion,
		FeedRegistry:          registry,
	}, nil
}

func (sc *StorkContract) getGasBudgetFromDryRun(
	pt *sui_types.ProgrammableTransaction,
	referenceGasPrice uint64,
) (uint64, error) {
	address, err := sui_types.NewAddressFromHex(sc.Account.Address)
	if err != nil {
		return 0, fmt.Errorf("failed to get address from hex: %w", err)
	}

	tx := sui_types.NewProgrammable(
		*address,
		nil,
		*pt,
		//nolint:mnd // 10e9 is an arbitrarily large gas budget and thus a permissible magic number
		uint64(10e9),
		referenceGasPrice,
	)

	txBytes, err := bcs.Marshal(tx)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal transaction: %w", err)
	}

	dryRunResult, err := sc.Client.DryRunTransaction(context.Background(), txBytes)
	if err != nil {
		return 0, fmt.Errorf("dry run failed: %w", err)
	}

	if !dryRunResult.Effects.Data.IsSuccess() {
		//nolint:err113 // This is essentially wrapping an error
		return 0, fmt.Errorf("dry run failed: %s", dryRunResult.Effects.Data.V1.Status.Error)
	}

	gasUsed := dryRunResult.Effects.Data.GasFee()

	gasUsedUint64, err := pusher.SafeInt64ToUint64(gasUsed)
	if err != nil {
		return 0, fmt.Errorf("failed to convert gas used to uint64: %w", err)
	}

	return gasUsedUint64, nil
}

func (sc *StorkContract) getReferenceGasPrice() (uint64, error) {
	referenceGasPriceResult, err := sc.Client.GetReferenceGasPrice(context.Background())
	if err != nil {
		return 0, fmt.Errorf("failed to get reference gas price: %w", err)
	}

	return referenceGasPriceResult.Uint64(), nil
}

func (sc *StorkContract) getFeedIDs(feedIDs []EncodedAssetID) (map[EncodedAssetID]sui_types.SuiAddress, error) {
	feedIDsMap := make(map[EncodedAssetID]sui_types.SuiAddress)

	registryID := sc.State.FeedRegistry.ID

	registryEntries, err := sc.Client.GetDynamicFields(context.Background(), registryID, nil, nil)
	if err != nil {
		return feedIDsMap, fmt.Errorf("failed to get dynamic fields: %w", err)
	}

	registryEntriesData := registryEntries.Data

	var nameBytes []byte

	for _, feedID := range feedIDs {
		for _, entry := range registryEntriesData {
			valueMap, exists := entry.Name.Value.(map[string]interface{})
			if !exists {
				return nil, fmt.Errorf("value is not a map: %w", ErrWrongType)
			}

			bytesField, exists := valueMap["bytes"]
			if !exists {
				return nil, fmt.Errorf("name field bytes not found: %w", ErrFieldNotFound)
			}

			nameBytes, err = interfaceSliceToBytes(bytesField)
			if err != nil {
				return nil, fmt.Errorf("failed to convert name bytes: %w", err)
			}

			if bytes.Equal(nameBytes, feedID[:]) {
				feedObjectID := entry.ObjectId
				feedIDsMap[feedID] = feedObjectID

				break
			}
		}
	}

	return feedIDsMap, nil
}

func interfaceSliceToBytes(slice interface{}) ([]byte, error) {
	interfaceSlice, ok := slice.([]interface{})
	if !ok {
		return nil, fmt.Errorf("input is not a slice of interfaces, but a %T: %w", slice, ErrWrongType)
	}

	byteSlice := make([]byte, len(interfaceSlice))

	var floatVal float64

	for i, v := range interfaceSlice {
		floatVal, ok = v.(float64)
		if !ok {
			return nil, fmt.Errorf("element at index %d is not a float64: %w", i, ErrWrongType)
		}

		byteSlice[i] = byte(floatVal)
	}

	return byteSlice, nil
}
