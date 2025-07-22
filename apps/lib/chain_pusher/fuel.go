package chain_pusher

/*
#cgo LDFLAGS: -L./fuel_ffi/target/release -lfuel_ffi
#cgo CFLAGS: -I./fuel_ffi/src
#include "fuel.h"
#include <stdlib.h>
*/
import "C"

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"
	"unsafe"

	"github.com/rs/zerolog"
)

type FuelContractInteractor struct {
	logger      zerolog.Logger
	client      *C.FuelClient
	rpcUrl      string
	contractId  string
	pollingPeriod int
}

type FuelConfig struct {
	RpcUrl         string `json:"rpc_url"`
	ContractAddress string `json:"contract_address"`
	PrivateKey     string `json:"private_key"`
}

type FuelTemporalNumericValue struct {
	TimestampNs     uint64 `json:"timestamp_ns"`
	QuantizedValue  string `json:"quantized_value"` // Using string for i128
}

type FuelTemporalNumericValueInput struct {
	TemporalNumericValue    FuelTemporalNumericValue `json:"temporal_numeric_value"`
	Id                      string                   `json:"id"`
	PublisherMerkleRoot     string                   `json:"publisher_merkle_root"`
	ValueComputeAlgHash     string                   `json:"value_compute_alg_hash"`
	R                       string                   `json:"r"`
	S                       string                   `json:"s"`
	V                       uint8                    `json:"v"`
}

func NewFuelContractInteractor(
	rpcUrl string,
	contractAddress string,
	privateKey string,
	pollingPeriod int,
	logger zerolog.Logger,
) (*FuelContractInteractor, error) {
	logger = logger.With().Str("component", "fuel-contract-interactor").Logger()

	config := FuelConfig{
		RpcUrl:         rpcUrl,
		ContractAddress: contractAddress,
		PrivateKey:     privateKey,
	}

	configJson, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal fuel config: %w", err)
	}

	configCStr := C.CString(string(configJson))
	defer C.free(unsafe.Pointer(configCStr))

	client := C.fuel_client_new(configCStr)
	if client == nil {
		return nil, fmt.Errorf("failed to create fuel client")
	}

	return &FuelContractInteractor{
		logger:        logger,
		client:        client,
		rpcUrl:        rpcUrl,
		contractId:    contractAddress,
		pollingPeriod: pollingPeriod,
	}, nil
}

func (fci *FuelContractInteractor) ListenContractEvents(
	ctx context.Context, 
	ch chan map[InternalEncodedAssetId]InternalTemporalNumericValue,
) {
	fci.logger.Warn().Msg("Fuel does not currently support listening to events via websocket, falling back to polling")
	// TODO: Implement event listening when Fuel supports it
}

func (fci *FuelContractInteractor) PullValues(
	encodedAssetIds []InternalEncodedAssetId,
) (map[InternalEncodedAssetId]InternalTemporalNumericValue, error) {
	result := make(map[InternalEncodedAssetId]InternalTemporalNumericValue)

	for _, assetId := range encodedAssetIds {
		// Convert asset ID to hex string
		idHex := hex.EncodeToString(assetId[:])

		// Convert hex string to bytes for FFI call
		idBytes, err := hex.DecodeString(idHex)
		if err != nil {
			fci.logger.Error().Err(err).Str("asset_id", idHex).Msg("Failed to decode asset ID")
			continue
		}

		// Ensure we have exactly 32 bytes
		if len(idBytes) != 32 {
			fci.logger.Error().Str("asset_id", idHex).Msg("Asset ID must be 32 bytes")
			continue
		}

		// Call FFI function
		valueJson := C.fuel_get_latest_value(fci.client, (*C.uchar)(unsafe.Pointer(&idBytes[0])))
		if valueJson == nil {
			// No value found for this asset
			continue
		}

		// Convert C string to Go string and free it
		valueStr := C.GoString(valueJson)
		C.fuel_free_string(valueJson)

		// Parse the JSON response
		var fuelValue FuelTemporalNumericValue
		if err := json.Unmarshal([]byte(valueStr), &fuelValue); err != nil {
			fci.logger.Error().Err(err).Str("asset_id", idHex).Msg("Failed to parse temporal numeric value")
			continue
		}

		// Convert to internal format
		quantizedValueBig := new(big.Int)
		quantizedValueBig.SetString(fuelValue.QuantizedValue, 10)

		internalValue := InternalTemporalNumericValue{
			TimestampNs:     fuelValue.TimestampNs,
			QuantizedValue:  quantizedValueBig,
		}

		result[assetId] = internalValue
	}

	return result, nil
}

func (fci *FuelContractInteractor) BatchPushToContract(
	priceUpdates map[InternalEncodedAssetId]AggregatedSignedPrice,
) error {
	if len(priceUpdates) == 0 {
		return nil
	}

	var inputs []FuelTemporalNumericValueInput
	
	for _, update := range priceUpdates {
		if update.StorkSignedPrice == nil {
			fci.logger.Error().Str("asset_id", string(update.AssetId)).Msg("StorkSignedPrice is nil")
			continue
		}

		// Parse quantized price
		quantizedPriceBigInt := new(big.Int)
		quantizedPriceBigInt.SetString(string(update.StorkSignedPrice.QuantizedPrice), 10)

		// Parse signature components
		rBytes, err := stringToByte32(update.StorkSignedPrice.TimestampedSignature.Signature.R)
		if err != nil {
			fci.logger.Error().Err(err).Msg("Failed to parse signature R")
			continue
		}

		sBytes, err := stringToByte32(update.StorkSignedPrice.TimestampedSignature.Signature.S)
		if err != nil {
			fci.logger.Error().Err(err).Msg("Failed to parse signature S")
			continue
		}

		// Parse encoded asset ID
		encodedAssetIdBytes, err := stringToByte32(string(update.StorkSignedPrice.EncodedAssetId))
		if err != nil {
			fci.logger.Error().Err(err).Msg("Failed to parse encoded asset ID")
			continue
		}

		// Parse publisher merkle root
		publisherMerkleRootBytes, err := stringToByte32(update.StorkSignedPrice.PublisherMerkleRoot)
		if err != nil {
			fci.logger.Error().Err(err).Msg("Failed to parse publisher merkle root")
			continue
		}

		// Parse value compute algorithm hash
		valueComputeAlgHashBytes, err := stringToByte32(update.StorkSignedPrice.StorkCalculationAlg.Checksum)
		if err != nil {
			fci.logger.Error().Err(err).Msg("Failed to parse value compute alg hash")
			continue
		}

		// Convert V from string to uint8 (remove "0x" prefix and parse as hex)
		vInt, err := strconv.ParseInt(update.StorkSignedPrice.TimestampedSignature.Signature.V[2:], 16, 8)
		if err != nil {
			fci.logger.Error().Err(err).Msg("Failed to parse signature V")
			continue
		}

		// Convert internal format to Fuel format
		fuelInput := FuelTemporalNumericValueInput{
			TemporalNumericValue: FuelTemporalNumericValue{
				TimestampNs:    uint64(update.StorkSignedPrice.TimestampedSignature.TimestampNano),
				QuantizedValue: quantizedPriceBigInt.String(),
			},
			Id:                  hex.EncodeToString(encodedAssetIdBytes[:]),
			PublisherMerkleRoot: hex.EncodeToString(publisherMerkleRootBytes[:]),
			ValueComputeAlgHash: hex.EncodeToString(valueComputeAlgHashBytes[:]),
			R:                   hex.EncodeToString(rBytes[:]),
			S:                   hex.EncodeToString(sBytes[:]),
			V:                   uint8(vInt),
		}

		inputs = append(inputs, fuelInput)
	}

	// Serialize inputs to JSON
	inputsJson, err := json.Marshal(inputs)
	if err != nil {
		return fmt.Errorf("failed to marshal fuel inputs: %w", err)
	}

	// Call FFI function
	inputsCStr := C.CString(string(inputsJson))
	defer C.free(unsafe.Pointer(inputsCStr))

	txHashPtr := C.fuel_update_values(fci.client, inputsCStr)
	if txHashPtr == nil {
		return fmt.Errorf("failed to update values on fuel contract")
	}

	// Get transaction hash and free it
	txHash := C.GoString(txHashPtr)
	C.fuel_free_string(txHashPtr)

	fci.logger.Info().
		Str("tx_hash", txHash).
		Int("num_updates", len(priceUpdates)).
		Msg("Successfully pushed updates to Fuel contract")

	return nil
}

func (fci *FuelContractInteractor) GetWalletBalance() (float64, error) {
	balance := C.fuel_get_wallet_balance(fci.client)
	
	// Convert from wei to ETH (assuming Fuel uses similar units)
	balanceFloat := float64(balance) / 1e9 // Fuel uses 9 decimals
	
	return balanceFloat, nil
}

func (fci *FuelContractInteractor) Close() {
	if fci.client != nil {
		C.fuel_client_free(fci.client)
		fci.client = nil
	}
}

// Helper function to create a logger for Fuel pusher
func FuelPusherLogger(rpcUrl, contractAddress string) zerolog.Logger {
	return AppLogger("fuel").With().
		Str("chainRpcUrl", rpcUrl).
		Str("contractAddress", contractAddress).
		Logger()
}