// These bindings are not generated.
// Instead, this file contains utility functions for interacting with the Aptos Stork contract.

package bindings

import (
	"errors"
	"fmt"
	"math"
	"math/big"
	"net/http"
	"net/http/cookiejar"
	"strconv"
	"sync"
	"time"

	"github.com/aptos-labs/aptos-go-sdk"
	"github.com/aptos-labs/aptos-go-sdk/bcs"
	"github.com/aptos-labs/aptos-go-sdk/crypto"
)

const DefaultHttpClientTimeout = 5 * time.Second

var (
	ErrInvalidLengths = errors.New("invalid lengths")
	ErrEmptyResponse  = errors.New("empty response")
	ErrWrongType      = errors.New("wrong type")
	ErrTxFailed       = errors.New("transaction failed")
)

type StorkContract struct {
	Client          *aptos.Client
	Account         *aptos.Account
	ContractAddress aptos.AccountAddress
}

type EncodedAssetID [32]byte

type TemporalNumericValue struct {
	TimestampNs    uint64
	QuantizedValue I128
}

type I128 struct {
	Magnitude *big.Int
	Negative  bool
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

func NewStorkContract(rpcUrl string, contractAddress string, key *crypto.Ed25519PrivateKey) (*StorkContract, error) {
	config := aptos.NetworkConfig{
		Name:       "",
		ChainId:    0,
		NodeUrl:    rpcUrl,
		FaucetUrl:  "",
		IndexerUrl: "",
	}

	// pass custom http client as aptos client calls are not context aware, they hardcode a timeout of 60 seconds
	// we hardcode our own shorter timeout - this is ultimately an imperfect workaround.
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	httpClient := &http.Client{
		Jar:     jar,
		Timeout: DefaultHttpClientTimeout,
	}

	client, err := aptos.NewClient(config, httpClient)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	account, err := aptos.NewAccountFromSigner(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create account from signer: %w", err)
	}

	address := aptos.AccountAddress{}

	err = address.ParseStringRelaxed(contractAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to parse contract address: %w", err)
	}

	return &StorkContract{Client: client, Account: account, ContractAddress: address}, nil
}

// GetMultipleTemporalNumericValuesUnchecked returns the temporal numeric values for the given feed IDs.
func (sc *StorkContract) GetMultipleTemporalNumericValuesUnchecked(
	feedIDs []EncodedAssetID,
) (map[EncodedAssetID]TemporalNumericValue, error) {
	response := map[EncodedAssetID]TemporalNumericValue{}

	var (
		mu sync.Mutex
		wg sync.WaitGroup
	)

	for _, id := range feedIDs {
		wg.Add(1)

		go func(id EncodedAssetID) {
			defer wg.Done()

			value, err := sc.getTemporalNumericValueUnchecked(id)
			if err != nil {
				// unfortunately, errors from view functions are pretty bad, so we assume an error means the value is not available
				return
			}

			mu.Lock()

			response[id] = value

			mu.Unlock()
		}(id)
	}

	wg.Wait()

	return response, nil
}

//nolint:funlen // This is a long function but does related work.
func (sc *StorkContract) UpdateMultipleTemporalNumericValuesEvm(updateData []UpdateData) (string, error) {
	// Create separate serializers for each vector
	idsSerializer := bcs.Serializer{}
	timestampsSerializer := bcs.Serializer{}
	magnitudesSerializer := bcs.Serializer{}
	negativesSerializer := bcs.Serializer{}
	merkleRootsSerializer := bcs.Serializer{}
	algHashesSerializer := bcs.Serializer{}
	rsSerializer := bcs.Serializer{}
	ssSerializer := bcs.Serializer{}
	vsSerializer := bcs.Serializer{}

	// Serialize each vector with its own serializer
	if len(updateData) > math.MaxUint32 {
		return "", ErrInvalidLengths
	}

	//nolint:all // safe to cast to uint32.
	lenUpdateData := uint32(len(updateData))

	idsSerializer.Uleb128(lenUpdateData)

	for _, data := range updateData {
		idsSerializer.WriteBytes(data.ID)
	}

	timestampsSerializer.Uleb128(lenUpdateData)

	for _, data := range updateData {
		timestampsSerializer.U64(data.TemporalNumericValueTimestampNs)
	}

	magnitudesSerializer.Uleb128(lenUpdateData)

	for _, data := range updateData {
		magnitudesSerializer.U128(*data.TemporalNumericValueMagnitude)
	}

	negativesSerializer.Uleb128(lenUpdateData)

	for _, data := range updateData {
		negativesSerializer.Bool(data.TemporalNumericValueNegative)
	}

	merkleRootsSerializer.Uleb128(lenUpdateData)

	for _, data := range updateData {
		merkleRootsSerializer.WriteBytes(data.PublisherMerkleRoot)
	}

	algHashesSerializer.Uleb128(lenUpdateData)

	for _, data := range updateData {
		algHashesSerializer.WriteBytes(data.ValueComputeAlgHash)
	}

	rsSerializer.Uleb128(lenUpdateData)

	for _, data := range updateData {
		rsSerializer.WriteBytes(data.R)
	}

	ssSerializer.Uleb128(lenUpdateData)

	for _, data := range updateData {
		ssSerializer.WriteBytes(data.S)
	}

	vsSerializer.Uleb128(lenUpdateData)

	for _, data := range updateData {
		vsSerializer.U8(data.V)
	}

	// Create the transaction payload with all serialized vectors
	payload := &aptos.EntryFunction{
		Module: aptos.ModuleId{
			Address: sc.ContractAddress,
			Name:    "stork",
		},
		Function: "update_multiple_temporal_numeric_values_evm",
		ArgTypes: []aptos.TypeTag{},
		Args: [][]byte{
			idsSerializer.ToBytes(),
			timestampsSerializer.ToBytes(),
			magnitudesSerializer.ToBytes(),
			negativesSerializer.ToBytes(),
			merkleRootsSerializer.ToBytes(),
			algHashesSerializer.ToBytes(),
			rsSerializer.ToBytes(),
			ssSerializer.ToBytes(),
			vsSerializer.ToBytes(),
		},
	}

	submitResponse, err := sc.Client.BuildSignAndSubmitTransaction(
		sc.Account,
		aptos.TransactionPayload{Payload: payload},
	)
	if err != nil {
		return "", fmt.Errorf("failed to build sign and submit transaction: %w", err)
	}

	hash := submitResponse.Hash

	tx, err := sc.Client.WaitForTransaction(hash)
	if err != nil {
		return "", fmt.Errorf("failed to wait for transaction: %w", err)
	}

	if !tx.Success {
		return "", fmt.Errorf("%s: %w", tx.VmStatus, ErrTxFailed)
	}

	return tx.Hash, nil
}

func (sc *StorkContract) getTemporalNumericValueUnchecked(id EncodedAssetID) (TemporalNumericValue, error) {
	serializer := bcs.Serializer{}
	serializer.WriteBytes(id[:])
	encodedAssetID := serializer.ToBytes()

	payload := &aptos.ViewPayload{
		Module: aptos.ModuleId{
			Address: sc.ContractAddress,
			Name:    "stork",
		},
		Function: "get_temporal_numeric_value_unchecked",
		ArgTypes: []aptos.TypeTag{},
		Args:     [][]byte{encodedAssetID},
	}

	value, err := sc.Client.View(payload)
	if err != nil {
		return TemporalNumericValue{}, fmt.Errorf("failed to get temporal numeric value: %w", err)
	}

	if len(value) == 0 {
		return TemporalNumericValue{}, ErrEmptyResponse
	}

	responseMap, ok := value[0].(map[string]interface{})
	if !ok {
		return TemporalNumericValue{}, ErrWrongType
	}

	timestampString, ok := responseMap["timestamp_ns"].(string)
	if !ok {
		return TemporalNumericValue{}, ErrWrongType
	}

	timestamp, err := strconv.ParseUint(timestampString, 10, 64)
	if err != nil {
		return TemporalNumericValue{}, fmt.Errorf("failed to parse timestamp: %w", err)
	}

	quantizedValue, ok := responseMap["quantized_value"].(map[string]interface{})
	if !ok {
		return TemporalNumericValue{}, ErrWrongType
	}

	magnitudeString, ok := quantizedValue["magnitude"].(string)
	if !ok {
		return TemporalNumericValue{}, ErrWrongType
	}

	magnitude := new(big.Int)
	//nolint:mnd // base number.
	magnitude.SetString(magnitudeString, 10)

	negative, ok := quantizedValue["negative"].(bool)
	if !ok {
		return TemporalNumericValue{}, ErrWrongType
	}

	return TemporalNumericValue{
		TimestampNs: timestamp,
		QuantizedValue: I128{
			Magnitude: magnitude,
			Negative:  negative,
		},
	}, nil
}
