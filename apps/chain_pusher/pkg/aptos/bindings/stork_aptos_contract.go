// Unlike the EVM and Solana bindings, the Aptos bindings are generated from the Move source code, as a tool for this does currently exist.
// Instead, this file contains utility functions for interacting with the Aptos Stork contract.
// These functions are written using the official aptos go sdk.

package bindings

import (
	"fmt"
	"math/big"
	"strconv"
	"sync"

	"github.com/aptos-labs/aptos-go-sdk"
	"github.com/aptos-labs/aptos-go-sdk/bcs"
	"github.com/aptos-labs/aptos-go-sdk/crypto"
)

type StorkContract struct {
	Client          *aptos.Client
	Account         *aptos.Account
	ContractAddress aptos.AccountAddress
}

type EncodedAssetId [32]byte

type TemporalNumericValue struct {
	TimestampNs    uint64
	QuantizedValue I128
}

type I128 struct {
	Magnitude *big.Int
	Negative  bool
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

func NewStorkContract(rpcUrl string, contractAddress string, key *crypto.Ed25519PrivateKey) (*StorkContract, error) {
	config := aptos.NetworkConfig{
		Name:       "",
		ChainId:    0,
		NodeUrl:    rpcUrl,
		FaucetUrl:  "",
		IndexerUrl: "",
	}
	client, err := aptos.NewClient(config)
	if err != nil {
		return nil, err
	}

	account, err := aptos.NewAccountFromSigner(key)
	if err != nil {
		return nil, err
	}

	address := aptos.AccountAddress{}
	err = address.ParseStringRelaxed(contractAddress)
	if err != nil {
		return nil, err
	}

	return &StorkContract{Client: client, Account: account, ContractAddress: address}, nil
}

func (sc *StorkContract) getTemporalNumericValueUnchecked(id EncodedAssetId) (TemporalNumericValue, error) {
	serializer := bcs.Serializer{}
	serializer.WriteBytes(id[:])
	encodedAssetId := serializer.ToBytes()

	payload := &aptos.ViewPayload{
		Module: aptos.ModuleId{
			Address: sc.ContractAddress,
			Name:    "stork",
		},
		Function: "get_temporal_numeric_value_unchecked",
		ArgTypes: []aptos.TypeTag{},
		Args:     [][]byte{encodedAssetId},
	}

	value, err := sc.Client.View(payload)
	if err != nil {
		return TemporalNumericValue{}, err
	}

	if len(value) == 0 {
		return TemporalNumericValue{}, fmt.Errorf("empty response")
	}

	responseMap := value[0].(map[string]interface{})
	timestamp, err := strconv.ParseUint(responseMap["timestamp_ns"].(string), 10, 64)
	if err != nil {
		return TemporalNumericValue{}, fmt.Errorf("failed to parse timestamp: %w", err)
	}

	quantizedValue := responseMap["quantized_value"].(map[string]interface{})
	magnitude := new(big.Int)
	magnitude.SetString(quantizedValue["magnitude"].(string), 10)
	negative := quantizedValue["negative"].(bool)

	return TemporalNumericValue{
		TimestampNs: timestamp,
		QuantizedValue: I128{
			Magnitude: magnitude,
			Negative:  negative,
		},
	}, nil
}

// GetMultipleTemporalNumericValuesUnchecked returns the temporal numeric values for the given feed IDs.
func (sc *StorkContract) GetMultipleTemporalNumericValuesUnchecked(feedIds []EncodedAssetId) (map[EncodedAssetId]TemporalNumericValue, error) {
	response := map[EncodedAssetId]TemporalNumericValue{}
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, id := range feedIds {
		wg.Add(1)
		go func(id EncodedAssetId) {
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
	idsSerializer.Uleb128(uint32(len(updateData)))
	for _, data := range updateData {
		idsSerializer.WriteBytes(data.Id)
	}

	timestampsSerializer.Uleb128(uint32(len(updateData)))
	for _, data := range updateData {
		timestampsSerializer.U64(data.TemporalNumericValueTimestampNs)
	}

	magnitudesSerializer.Uleb128(uint32(len(updateData)))
	for _, data := range updateData {
		magnitudesSerializer.U128(*data.TemporalNumericValueMagnitude)
	}

	negativesSerializer.Uleb128(uint32(len(updateData)))
	for _, data := range updateData {
		negativesSerializer.Bool(data.TemporalNumericValueNegative)
	}

	merkleRootsSerializer.Uleb128(uint32(len(updateData)))
	for _, data := range updateData {
		merkleRootsSerializer.WriteBytes(data.PublisherMerkleRoot)
	}

	algHashesSerializer.Uleb128(uint32(len(updateData)))
	for _, data := range updateData {
		algHashesSerializer.WriteBytes(data.ValueComputeAlgHash)
	}

	rsSerializer.Uleb128(uint32(len(updateData)))
	for _, data := range updateData {
		rsSerializer.WriteBytes(data.R)
	}

	ssSerializer.Uleb128(uint32(len(updateData)))
	for _, data := range updateData {
		ssSerializer.WriteBytes(data.S)
	}

	vsSerializer.Uleb128(uint32(len(updateData)))
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

	submitResponse, err := sc.Client.BuildSignAndSubmitTransaction(sc.Account, aptos.TransactionPayload{Payload: payload})
	if err != nil {
		return "", err
	}

	hash := submitResponse.Hash
	tx, err := sc.Client.WaitForTransaction(hash)
	if err != nil {
		return "", err
	}

	if !tx.Success {
		return "", fmt.Errorf("transaction failed: %s", tx.VmStatus)
	}

	return tx.Hash, nil
}
