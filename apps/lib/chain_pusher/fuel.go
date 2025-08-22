package chain_pusher

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/Stork-Oracle/stork-external/apps/lib/chain_pusher/contract_bindings/fuel"
	"github.com/rs/zerolog"
)

type FuelContractInteractor struct {
	logger   zerolog.Logger
	contract *fuel.StorkContract
}

func NewFuelContractInteractor(
	rpcUrl string,
	contractAddress string,
	privateKey string,
	logger zerolog.Logger,
) (*FuelContractInteractor, error) {
	logger = logger.With().Str("component", "fuel-contract-interactor").Logger()

	config := fuel.FuelConfig{
		RpcUrl:          rpcUrl,
		ContractAddress: contractAddress,
		PrivateKey:      privateKey,
	}

	contract, err := fuel.NewStorkContract(config.RpcUrl, config.ContractAddress, config.PrivateKey, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create stork contract client: %w", err)
	}

	return &FuelContractInteractor{
		logger:   logger,
		contract: contract,
	}, nil
}

func (fci *FuelContractInteractor) ListenContractEvents(
	ctx context.Context,
	ch chan map[InternalEncodedAssetId]InternalTemporalNumericValue,
) {
	fci.logger.Warn().Msg("Fuel does not currently support listening to events via websocket, falling back to polling")
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
		valueJson, err := fci.contract.GetTemporalNumericValueUncheckedV1([32]byte(idBytes))
		if err != nil {
			if strings.Contains(err.Error(), "feed not found") {
				fci.logger.Warn().Err(err).Str("asset_id", idHex).Msg("No value found")
			} else {
				fci.logger.Warn().Err(err).Str("asset_id", idHex).Msg("Failed to get temporal numeric value")
			}
			continue
		}

		// Convert to internal format

		internalValue := InternalTemporalNumericValue{
			TimestampNs:    valueJson.TimestampNs,
			QuantizedValue: valueJson.QuantizedValue,
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

	var inputs []fuel.FuelTemporalNumericValueInput

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
		fuelInput := fuel.FuelTemporalNumericValueInput{
			TemporalNumericValue: fuel.FuelTemporalNumericValue{
				TimestampNs:    uint64(update.StorkSignedPrice.TimestampedSignature.TimestampNano),
				QuantizedValue: quantizedPriceBigInt,
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

	// Call FFI function
	txHash, err := fci.contract.UpdateTemporalNumericValuesV1(inputs)
	if err != nil {
		return fmt.Errorf("failed to update values on fuel contract: %w", err)
	}

	fci.logger.Info().
		Str("tx_hash", txHash).
		Int("num_updates", len(priceUpdates)).
		Msg("Successfully pushed updates to Fuel contract")

	return nil
}

func (fci *FuelContractInteractor) GetWalletBalance() (float64, error) {
	balance, err := fci.contract.GetWalletBalance()
	if err != nil {
		return 0, fmt.Errorf("failed to get wallet balance: %w", err)
	}

	return float64(balance), nil
}

func (fci *FuelContractInteractor) Close() {
	if fci.contract != nil {
		fci.contract.Close()
	}
}
