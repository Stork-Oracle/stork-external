// These binding are not generated.
// Instead, this file contains utility functions for interacting with the Sui Stork contract
// over the Sui fullnode gRPC v2 API.

package bindings

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"maps"
	"math/big"
	"sort"
	"strconv"

	"github.com/Stork-Oracle/go-sui-sdk/v2/account"
	"github.com/Stork-Oracle/go-sui-sdk/v2/lib"
	"github.com/Stork-Oracle/go-sui-sdk/v2/sui_types"
	"github.com/Stork-Oracle/stork-external/apps/chain_pusher/pkg/pusher"
	rpcv2 "github.com/Stork-Oracle/stork-external/apps/chain_pusher/pkg/sui/rpc/v2"
	"github.com/fardream/go-bcs/bcs"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

var (
	ErrFeedRegistryNotFound       = errors.New("feed registry not found")
	ErrFieldNotFound              = errors.New("field not found")
	ErrWrongType                  = errors.New("wrong type")
	ErrStorkStateIDRequired       = errors.New("stork state ID is required")
	ErrStateNotShared             = errors.New("stork state object is not a shared object")
	ErrInsufficientGasCoins       = errors.New("insufficient gas coin balance")
	ErrUnsupportedSignatureScheme = errors.New("unsupported signature scheme")
	ErrEmptyResponse              = errors.New("empty response")
)

const (
	// suiCoinObjectType is the on-chain object type of owned SUI coins (coins are
	// Coin<T> wrapper objects); used to filter ListOwnedObjects for gas selection.
	suiCoinObjectType = "0x2::coin::Coin<0x2::sui::SUI>"
	// suiCoinType is the bare currency type parameter T; used by GetBalance.
	suiCoinType         = "0x2::sui::SUI"
	dynamicFieldsPage   = 1000
	maxGasCoinCandidate = 100
	uleb128MaxShiftBits = 28
)

type StorkContract struct {
	Client          *Client
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

// NewStorkContract creates a StorkContract client backed by the Sui fullnode gRPC v2 API.
// rpcUrl is a gRPC endpoint (e.g. "fullnode.mainnet.sui.io:443" or "https://host").
// storkStateID is the object ID of the shared StorkState created when the contract was
// initialized.
func NewStorkContract(
	ctx context.Context,
	rpcUrl string,
	contractAddress string,
	account *account.Account,
	storkStateID string,
) (*StorkContract, error) {
	if storkStateID == "" {
		return nil, ErrStorkStateIDRequired
	}

	client, err := DialGrpc(rpcUrl)
	if err != nil {
		return nil, err
	}

	contractAddr, err := sui_types.NewAddressFromHex(contractAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to convert contract address to Sui address: %w", err)
	}

	state, err := getStorkState(ctx, client, storkStateID)
	if err != nil {
		return nil, err
	}

	return &StorkContract{Client: client, Account: account, ContractAddress: *contractAddr, State: state}, nil
}

// GetMultipleTemporalNumericValuesUnchecked gets multiple temporal numeric values at a time for efficiency.
func (sc *StorkContract) GetMultipleTemporalNumericValuesUnchecked(
	ctx context.Context, feedIDs []EncodedAssetID,
) (map[EncodedAssetID]TemporalNumericValue, error) {
	feedIDsMap := sc.State.FeedRegistry.Entries

	unknownFeedIDs := []EncodedAssetID{}

	for _, feedID := range feedIDs {
		if _, ok := feedIDsMap[feedID]; !ok {
			unknownFeedIDs = append(unknownFeedIDs, feedID)
		}
	}

	resolvedFeedIDs, err := sc.getFeedIDs(ctx, unknownFeedIDs)
	if err != nil {
		return nil, err
	}

	maps.Copy(feedIDsMap, resolvedFeedIDs)

	requests := []*rpcv2.GetObjectRequest{}
	for _, feedObjectID := range feedIDsMap {
		requests = append(requests, &rpcv2.GetObjectRequest{
			ObjectId: proto.String(feedObjectID.String()),
		})
	}

	response, err := sc.Client.Ledger.BatchGetObjects(ctx, &rpcv2.BatchGetObjectsRequest{
		Requests: requests,
		ReadMask: &fieldmaskpb.FieldMask{Paths: []string{"object_id", "json"}},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get feed objects: %w", err)
	}

	result := make(map[EncodedAssetID]TemporalNumericValue)

	var (
		id    EncodedAssetID
		value TemporalNumericValue
	)

	for _, objectResult := range response.GetObjects() {
		object := objectResult.GetObject()
		if object == nil {
			return nil, fmt.Errorf(
				"failed to get feed object: %w: %s", ErrEmptyResponse, objectResult.GetError().GetMessage(),
			)
		}

		id, value, err = parseFeedToTemporalNumericValue(object)
		if err != nil {
			return nil, err
		}

		result[id] = value
	}

	return result, nil
}

func parseFeedToTemporalNumericValue(feed *rpcv2.Object) (EncodedAssetID, TemporalNumericValue, error) {
	var id EncodedAssetID

	fields, err := objectFields(feed)
	if err != nil {
		return id, TemporalNumericValue{}, err
	}

	assetIDBytes, err := nestedBase64Bytes(fields, "asset_id")
	if err != nil {
		return id, TemporalNumericValue{}, err
	}

	copy(id[:], assetIDBytes)

	latestValue, ok := fields["latest_value"].(map[string]any)
	if !ok {
		return id, TemporalNumericValue{}, fmt.Errorf("latest_value is not a map: %w", ErrWrongType)
	}

	timestampNsStr, ok := latestValue["timestamp_ns"].(string)
	if !ok {
		return id, TemporalNumericValue{}, fmt.Errorf("timestamp_ns is not a string: %w", ErrWrongType)
	}

	timestampNs, err := strconv.ParseUint(timestampNsStr, 10, 64)
	if err != nil {
		return id, TemporalNumericValue{}, fmt.Errorf("failed to parse timestamp_ns: %w", err)
	}

	quantizedValue, ok := latestValue["quantized_value"].(map[string]any)
	if !ok {
		return id, TemporalNumericValue{}, fmt.Errorf("quantized_value is not a map: %w", ErrWrongType)
	}

	magnitudeStr, ok := quantizedValue["magnitude"].(string)
	if !ok {
		return id, TemporalNumericValue{}, fmt.Errorf("magnitude is not a string: %w", ErrWrongType)
	}

	//nolint:mnd // Base number
	magnitude, ok := new(big.Int).SetString(magnitudeStr, 10)
	if !ok {
		return id, TemporalNumericValue{}, fmt.Errorf("failed to parse magnitude: %w", ErrWrongType)
	}

	negative, ok := quantizedValue["negative"].(bool)
	if !ok {
		return id, TemporalNumericValue{}, fmt.Errorf("negative is not a bool: %w", ErrWrongType)
	}

	return id, TemporalNumericValue{
		TimestampNs: timestampNs,
		QuantizedValue: I128{
			Magnitude: magnitude,
			Negative:  negative,
		},
	}, nil
}

//nolint:funlen,cyclop,maintidx // This function assembles a multi-command programmable transaction.
func (sc *StorkContract) UpdateMultipleTemporalNumericValuesEvm(
	ctx context.Context,
	updateData []UpdateData,
) (string, error) {
	ptb := sui_types.NewProgrammableTransactionBuilder()

	// get reference gas price
	referenceGasPrice, err := sc.getReferenceGasPrice(ctx)
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

	gasBudget, err := sc.getGasBudgetFromDryRun(ctx, &pt, referenceGasPrice)
	if err != nil {
		return "", err
	}

	gasCoins, err := sc.pickGasCoins(ctx, *address, totalFeeAmount+gasBudget)
	if err != nil {
		return "", err
	}

	tx := sui_types.NewProgrammable(
		*address,
		gasCoins,
		pt,
		gasBudget,
		referenceGasPrice,
	)

	txBytes, err := bcs.Marshal(tx)
	if err != nil {
		return "", fmt.Errorf("failed to marshal transaction: %w", err)
	}

	signature, err := sc.Account.SignSecureWithoutEncode(txBytes, sui_types.DefaultIntent())
	if err != nil {
		return "", fmt.Errorf("failed to sign transaction: %w", err)
	}

	sigBytes, err := signatureBytes(signature)
	if err != nil {
		return "", err
	}

	txResponse, err := sc.Client.Execution.ExecuteTransaction(ctx, &rpcv2.ExecuteTransactionRequest{
		Transaction: &rpcv2.Transaction{Bcs: &rpcv2.Bcs{Value: txBytes}},
		Signatures:  []*rpcv2.UserSignature{{Bcs: &rpcv2.Bcs{Value: sigBytes}}},
		ReadMask:    &fieldmaskpb.FieldMask{Paths: []string{"transaction.digest"}},
	})
	if err != nil {
		return "", fmt.Errorf("failed to execute transaction: %w", err)
	}

	return txResponse.GetTransaction().GetDigest(), nil
}

// GetWalletBalance returns the account's total SUI balance in MIST (the smallest
// denomination), matching the other chain interactors which report raw base units.
func (sc *StorkContract) GetWalletBalance(ctx context.Context) (float64, error) {
	response, err := sc.Client.State.GetBalance(ctx, &rpcv2.GetBalanceRequest{
		Owner:    proto.String(sc.Account.Address),
		CoinType: proto.String(suiCoinType),
	})
	if err != nil {
		return 0, fmt.Errorf("failed to get balance: %w", err)
	}

	return float64(response.GetBalance().GetBalance()), nil
}

//nolint:cyclop,funlen // This is a long and complex function due to interface destructuring
func getStorkState(
	ctx context.Context,
	client *Client,
	configuredStateID string,
) (StorkState, error) {
	storkStateID, err := sui_types.NewAddressFromHex(configuredStateID)
	if err != nil {
		return StorkState{}, fmt.Errorf("failed to parse configured stork state ID: %w", err)
	}

	response, err := client.Ledger.GetObject(ctx, &rpcv2.GetObjectRequest{
		ObjectId: proto.String(storkStateID.String()),
		ReadMask: &fieldmaskpb.FieldMask{Paths: []string{"object_id", "version", "owner", "json"}},
	})
	if err != nil {
		return StorkState{}, fmt.Errorf("failed to get object: %w", err)
	}

	object := response.GetObject()
	if object == nil {
		return StorkState{}, fmt.Errorf("failed to get stork state object: %w", ErrEmptyResponse)
	}

	fields, err := objectFields(object)
	if err != nil {
		return StorkState{}, err
	}

	evmPublicKeyBytes, err := nestedBase64Bytes(fields, "stork_evm_public_key")
	if err != nil {
		return StorkState{}, err
	}

	storkEvmPublicKey := hex.EncodeToString(evmPublicKeyBytes)

	storkSuiPublicKeyString, ok := fields["stork_sui_address"].(string)
	if !ok {
		return StorkState{}, fmt.Errorf("stork sui address is not a string: %w", ErrWrongType)
	}

	storkSuiPublicKey, err := sui_types.NewAddressFromHex(storkSuiPublicKeyString)
	if err != nil {
		return StorkState{}, fmt.Errorf("failed to get address from hex: %w", err)
	}

	singleUpdateFeeInMistString, ok := fields["single_update_fee_in_mist"].(string)
	if !ok {
		return StorkState{}, fmt.Errorf("single update fee in mist is not a string: %w", ErrWrongType)
	}

	singleUpdateFeeInMist, err := strconv.ParseUint(singleUpdateFeeInMistString, 10, 64)
	if err != nil {
		return StorkState{}, fmt.Errorf("failed to parse single update fee in mist: %w", err)
	}

	owner := object.GetOwner()
	if owner.GetKind() != rpcv2.Owner_SHARED {
		return StorkState{}, fmt.Errorf("%w: owner kind is %s", ErrStateNotShared, owner.GetKind())
	}

	version := object.GetVersion()
	initialSharedVersion := owner.GetVersion()

	// registry
	stateDynamicFields, err := listAllDynamicFields(ctx, client, storkStateID.String())
	if err != nil {
		return StorkState{}, err
	}

	registryID := sui_types.SuiAddress{}

	var (
		nameBytes       []byte
		registryAddress *sui_types.SuiAddress
	)

	for _, dynamicField := range stateDynamicFields {
		nameBytes, err = decodeBcsBytes(dynamicField.GetName().GetValue())
		if err != nil {
			return StorkState{}, fmt.Errorf("failed to decode dynamic field name: %w", err)
		}

		if bytes.Equal(nameBytes, []byte("temporal_numeric_value_feed_registry")) {
			registryAddress, err = sui_types.NewAddressFromHex(dynamicFieldObjectID(dynamicField))
			if err != nil {
				return StorkState{}, fmt.Errorf("failed to parse registry object ID: %w", err)
			}

			registryID = *registryAddress

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
	ctx context.Context,
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

	// Checks are disabled so the empty gas payment and placeholder budget are not
	// validated against the sender's balance, matching JSON-RPC dry-run semantics.
	simulateResponse, err := sc.Client.Execution.SimulateTransaction(ctx, &rpcv2.SimulateTransactionRequest{
		Transaction: &rpcv2.Transaction{Bcs: &rpcv2.Bcs{Value: txBytes}},
		ReadMask: &fieldmaskpb.FieldMask{
			Paths: []string{"transaction.effects.status", "transaction.effects.gas_used"},
		},
		Checks: rpcv2.SimulateTransactionRequest_DISABLED.Enum(),
	})
	if err != nil {
		return 0, fmt.Errorf("dry run failed: %w", err)
	}

	effects := simulateResponse.GetTransaction().GetEffects()
	if effects == nil {
		return 0, fmt.Errorf("dry run returned no effects: %w", ErrEmptyResponse)
	}

	if !effects.GetStatus().GetSuccess() {
		//nolint:err113 // This is essentially wrapping an error
		return 0, fmt.Errorf("dry run failed: %s", effects.GetStatus().GetError().GetDescription())
	}

	gasUsed := effects.GetGasUsed()

	computationCost, err := pusher.SafeUint64ToInt64(gasUsed.GetComputationCost())
	if err != nil {
		return 0, fmt.Errorf("failed to convert computation cost to int64: %w", err)
	}

	storageCost, err := pusher.SafeUint64ToInt64(gasUsed.GetStorageCost())
	if err != nil {
		return 0, fmt.Errorf("failed to convert storage cost to int64: %w", err)
	}

	storageRebate, err := pusher.SafeUint64ToInt64(gasUsed.GetStorageRebate())
	if err != nil {
		return 0, fmt.Errorf("failed to convert storage rebate to int64: %w", err)
	}

	gasFeeUint64, err := pusher.SafeInt64ToUint64(computationCost + storageCost - storageRebate)
	if err != nil {
		return 0, fmt.Errorf("failed to convert gas fee to uint64: %w", err)
	}

	return gasFeeUint64, nil
}

func (sc *StorkContract) getReferenceGasPrice(ctx context.Context) (uint64, error) {
	response, err := sc.Client.Ledger.GetEpoch(ctx, &rpcv2.GetEpochRequest{
		ReadMask: &fieldmaskpb.FieldMask{Paths: []string{"epoch", "reference_gas_price"}},
	})
	if err != nil {
		return 0, fmt.Errorf("failed to get reference gas price: %w", err)
	}

	return response.GetEpoch().GetReferenceGasPrice(), nil
}

func (sc *StorkContract) getFeedIDs(
	ctx context.Context,
	feedIDs []EncodedAssetID,
) (map[EncodedAssetID]sui_types.SuiAddress, error) {
	feedIDsMap := make(map[EncodedAssetID]sui_types.SuiAddress)

	if len(feedIDs) == 0 {
		return feedIDsMap, nil
	}

	registryEntries, err := listAllDynamicFields(ctx, sc.Client, sc.State.FeedRegistry.ID.String())
	if err != nil {
		return nil, err
	}

	entriesByAssetID := make(map[EncodedAssetID]sui_types.SuiAddress)

	var (
		nameBytes    []byte
		feedObjectID *sui_types.SuiAddress
	)

	for _, entry := range registryEntries {
		nameBytes, err = decodeBcsBytes(entry.GetName().GetValue())
		if err != nil {
			return nil, fmt.Errorf("failed to decode registry entry name: %w", err)
		}

		if len(nameBytes) != len(EncodedAssetID{}) {
			continue
		}

		feedObjectID, err = sui_types.NewAddressFromHex(dynamicFieldObjectID(entry))
		if err != nil {
			return nil, fmt.Errorf("failed to parse feed object ID: %w", err)
		}

		var assetID EncodedAssetID

		copy(assetID[:], nameBytes)

		entriesByAssetID[assetID] = *feedObjectID
	}

	for _, feedID := range feedIDs {
		if feedObjectID, ok := entriesByAssetID[feedID]; ok {
			feedIDsMap[feedID] = feedObjectID
		}
	}

	return feedIDsMap, nil
}

// pickGasCoins selects SUI gas coins owned by the sender totaling at least requiredAmount.
func (sc *StorkContract) pickGasCoins(
	ctx context.Context,
	owner sui_types.SuiAddress,
	requiredAmount uint64,
) ([]*sui_types.ObjectRef, error) {
	response, err := sc.Client.State.ListOwnedObjects(ctx, &rpcv2.ListOwnedObjectsRequest{
		Owner:      proto.String(owner.String()),
		ObjectType: proto.String(suiCoinObjectType),
		PageSize:   proto.Uint32(maxGasCoinCandidate),
		ReadMask:   &fieldmaskpb.FieldMask{Paths: []string{"object_id", "version", "digest", "balance"}},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list owned coins: %w", err)
	}

	coins := response.GetObjects()

	// largest balances first so the gas payment stays small
	sort.Slice(coins, func(i, j int) bool {
		return coins[i].GetBalance() > coins[j].GetBalance()
	})

	coinRefs := []*sui_types.ObjectRef{}
	total := uint64(0)

	var (
		objectID *sui_types.SuiAddress
		digest   *sui_types.Digest
	)

	for _, coin := range coins {
		objectID, err = sui_types.NewAddressFromHex(coin.GetObjectId())
		if err != nil {
			return nil, fmt.Errorf("failed to parse coin object ID: %w", err)
		}

		digest, err = sui_types.NewDigest(coin.GetDigest())
		if err != nil {
			return nil, fmt.Errorf("failed to parse coin digest: %w", err)
		}

		coinRefs = append(coinRefs, &sui_types.ObjectRef{
			ObjectId: *objectID,
			Version:  coin.GetVersion(),
			Digest:   *digest,
		})

		total += coin.GetBalance()
		if total >= requiredAmount {
			return coinRefs, nil
		}
	}

	return nil, fmt.Errorf("%w: have %d, need %d", ErrInsufficientGasCoins, total, requiredAmount)
}

// listAllDynamicFields pages through every dynamic field of parent.
func listAllDynamicFields(ctx context.Context, client *Client, parent string) ([]*rpcv2.DynamicField, error) {
	dynamicFields := []*rpcv2.DynamicField{}

	var pageToken []byte

	for {
		response, err := client.State.ListDynamicFields(ctx, &rpcv2.ListDynamicFieldsRequest{
			Parent:    proto.String(parent),
			PageSize:  proto.Uint32(dynamicFieldsPage),
			PageToken: pageToken,
			ReadMask:  &fieldmaskpb.FieldMask{Paths: []string{"kind", "field_id", "name", "child_id"}},
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list dynamic fields: %w", err)
		}

		dynamicFields = append(dynamicFields, response.GetDynamicFields()...)

		pageToken = response.GetNextPageToken()
		if len(pageToken) == 0 {
			return dynamicFields, nil
		}
	}
}

// dynamicFieldObjectID returns the object ID a dynamic field points at: the child object
// for dynamic object fields, otherwise the field object itself.
func dynamicFieldObjectID(dynamicField *rpcv2.DynamicField) string {
	if childID := dynamicField.GetChildId(); childID != "" {
		return childID
	}

	return dynamicField.GetFieldId()
}

// objectFields returns the Move struct fields of an object from its JSON rendering.
func objectFields(object *rpcv2.Object) (map[string]any, error) {
	jsonValue := object.GetJson()
	if jsonValue == nil {
		return nil, fmt.Errorf("object has no json content: %w", ErrFieldNotFound)
	}

	fields, ok := jsonValue.AsInterface().(map[string]any)
	if !ok {
		return nil, fmt.Errorf("object json is not a map: %w", ErrWrongType)
	}

	return fields, nil
}

// nestedBase64Bytes extracts fields[key]["bytes"] (a base64 string in the gRPC JSON
// rendering of Move byte vectors) and decodes it.
func nestedBase64Bytes(fields map[string]any, key string) ([]byte, error) {
	inner, ok := fields[key].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("%s is not a map: %w", key, ErrWrongType)
	}

	encoded, ok := inner["bytes"].(string)
	if !ok {
		return nil, fmt.Errorf("%s bytes is not a string: %w", key, ErrWrongType)
	}

	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("failed to decode %s bytes: %w", key, err)
	}

	return decoded, nil
}

// decodeBcsBytes decodes a BCS byte vector: a ULEB128 length prefix followed by that many
// bytes. Structs wrapping a single vector<u8> (e.g. EncodedAssetId) serialize identically.
func decodeBcsBytes(data []byte) ([]byte, error) {
	length := 0
	shift := 0
	offset := 0

	for {
		if offset >= len(data) {
			return nil, fmt.Errorf("truncated ULEB128 length prefix: %w", ErrWrongType)
		}

		b := data[offset]
		offset++

		//nolint:mnd // ULEB128 uses 7 value bits per byte
		length |= int(b&0x7f) << shift

		if b&0x80 == 0 {
			break
		}

		shift += 7
		if shift > uleb128MaxShiftBits {
			return nil, fmt.Errorf("ULEB128 length prefix too large: %w", ErrWrongType)
		}
	}

	if len(data)-offset != length {
		return nil, fmt.Errorf("BCS byte vector length mismatch: %w", ErrWrongType)
	}

	return data[offset:], nil
}

// signatureBytes returns the serialized (flag || signature || public key) bytes of a
// signature, as expected by the gRPC UserSignature bcs field.
func signatureBytes(signature sui_types.Signature) ([]byte, error) {
	switch {
	case signature.Ed25519SuiSignature != nil:
		return signature.Ed25519SuiSignature.Signature[:], nil
	case signature.Secp256k1SuiSignature != nil:
		return signature.Secp256k1SuiSignature.Signature, nil
	case signature.Secp256r1SuiSignature != nil:
		return signature.Secp256r1SuiSignature.Signature, nil
	default:
		return nil, ErrUnsupportedSignatureScheme
	}
}
