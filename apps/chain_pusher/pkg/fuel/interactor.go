package fuel

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strconv"

	"github.com/Stork-Oracle/stork-external/apps/chain_pusher/pkg/fuel/bindings"
	"github.com/Stork-Oracle/stork-external/apps/chain_pusher/pkg/pusher"
	"github.com/Stork-Oracle/stork-external/apps/chain_pusher/pkg/types"
	"github.com/rs/zerolog"
)

var ErrPrivateKeyEmpty = errors.New("private key cannot be empty")

type ContractInteractor struct {
	logger          zerolog.Logger
	privateKey      string
	contractAddress string
	contract        *bindings.StorkContract
}

func NewContractInteractor(
	contractAddress string,
	keyFileContent []byte,
	logger zerolog.Logger,
) (*ContractInteractor, error) {
	logger = logger.With().Str("component", "fuel-contract-interactor").Logger()

	privateKey, err := loadPrivateKey(keyFileContent)
	if err != nil {
		return nil, fmt.Errorf("failed to load private key: %w", err)
	}

	return &ContractInteractor{
		contract:        nil,
		logger:          logger,
		privateKey:      privateKey,
		contractAddress: contractAddress,
	}, nil
}

func (fci *ContractInteractor) ConnectHTTP(_ context.Context, url string) error {
	config := bindings.Config{
		RpcUrl:          url,
		ContractAddress: fci.contractAddress,
		PrivateKey:      fci.privateKey,
	}

	contract, err := bindings.NewStorkContract(config.RpcUrl, config.ContractAddress, config.PrivateKey)
	if err != nil {
		return fmt.Errorf("failed to create stork contract client: %w", err)
	}

	fci.contract = contract

	return nil
}

func (fci *ContractInteractor) ConnectWs(ctx context.Context, url string) error {
	// not implemented
	return nil
}

func (fci *ContractInteractor) ListenContractEvents(
	ctx context.Context,
	ch chan map[types.InternalEncodedAssetID]types.InternalTemporalNumericValue,
) {
	fci.logger.Warn().Msg("Fuel does not currently support listening to events via websocket, falling back to polling")
}

func (fci *ContractInteractor) PullValues(
	_ context.Context, // a 5 second timeout is hardcoded in the ffi library
	encodedAssetIDs []types.InternalEncodedAssetID,
) (map[types.InternalEncodedAssetID]types.InternalTemporalNumericValue, error) {
	result := make(map[types.InternalEncodedAssetID]types.InternalTemporalNumericValue)

	if len(encodedAssetIDs) == 0 {
		return result, nil
	}

	// Convert all asset IDs to [32]byte format
	ids := make([][32]byte, 0, len(encodedAssetIDs))
	validAssetIDs := make([]types.InternalEncodedAssetID, 0, len(encodedAssetIDs))

	for _, assetID := range encodedAssetIDs {
		// Convert asset ID to hex string
		idHex := hex.EncodeToString(assetID[:])

		// Convert hex string to bytes for FFI call
		idBytes, err := hex.DecodeString(idHex)
		if err != nil {
			fci.logger.Error().Err(err).Str("asset_id", idHex).Msg("Failed to decode asset ID")
			continue
		}

		// Ensure we have exactly 32 bytes
		//nolint:mnd // 32 bytes is the expected length for a Fuel asset ID.
		if len(idBytes) != 32 {
			fci.logger.Error().Str("asset_id", idHex).Msg("Asset ID must be 32 bytes")
			continue
		}

		ids = append(ids, [32]byte(idBytes))
		validAssetIDs = append(validAssetIDs, assetID)
	}

	if len(ids) == 0 {
		return result, nil
	}

	// Call batch FFI function
	values, err := fci.contract.GetTemporalNumericValuesUncheckedV1(ids)
	if err != nil {
		return result, fmt.Errorf("failed to get temporal numeric values: %w", err)
	}

	// Map results back to asset IDs
	for i, value := range values {
		if i >= len(validAssetIDs) {
			fci.logger.Warn().Int("index", i).Msg("Received more values than requested")
			break
		}

		internalValue := types.InternalTemporalNumericValue{
			TimestampNs:    value.TimestampNs,
			QuantizedValue: value.QuantizedValue,
		}

		result[validAssetIDs[i]] = internalValue
	}

	return result, nil
}

func (fci *ContractInteractor) BatchPushToContract(
	_ context.Context, // a 5 second timeout is hardcoded in the ffi library
	priceUpdates map[types.InternalEncodedAssetID]types.AggregatedSignedPrice,
) error {
	if len(priceUpdates) == 0 {
		return nil
	}

	inputs := make([]bindings.TemporalNumericValueInput, 0, len(priceUpdates))

	for _, update := range priceUpdates {
		if update.StorkSignedPrice == nil {
			fci.logger.Error().Str("asset_id", string(update.AssetID)).Msg("StorkSignedPrice is nil")

			continue
		}

		fuelInput, err := aggregatedSignedPriceToTemporalNumericValueInput(update)
		if err != nil {
			fci.logger.Error().Err(err).Str("asset_id", string(update.AssetID)).Msg("Failed to convert price update")

			continue
		}

		inputs = append(inputs, fuelInput)
	}

	// Call FFI function
	txHash, err := fci.contract.UpdateTemporalNumericValuesV1(inputs)
	if err != nil {
		return fmt.Errorf("failed to update values on fuel contract: %w", err)
	}

	fci.logger.Debug().
		Str("tx_hash", txHash).
		Int("num_updates", len(priceUpdates)).
		Msg("Successfully pushed updates to Fuel contract")

	return nil
}

// GetWalletBalance uses a 5 second timeout is hardcoded in the ffi library.
func (fci *ContractInteractor) GetWalletBalance(_ context.Context) (float64, error) {
	balance, err := fci.contract.GetWalletBalance()
	if err != nil {
		return 0, fmt.Errorf("failed to get wallet balance: %w", err)
	}

	return float64(balance), nil
}

func (fci *ContractInteractor) Close() {
	if fci.contract != nil {
		fci.contract.Close()
	}
}

func aggregatedSignedPriceToTemporalNumericValueInput(
	update types.AggregatedSignedPrice,
) (bindings.TemporalNumericValueInput, error) {
	// Parse quantized price
	quantizedPriceBigInt := new(big.Int)
	//nolint:mnd // base number.
	quantizedPriceBigInt.SetString(string(update.StorkSignedPrice.QuantizedPrice), 10)

	// Parse signature components
	rBytes, err := pusher.HexStringToByte32(update.StorkSignedPrice.TimestampedSignature.Signature.R)
	if err != nil {
		return bindings.TemporalNumericValueInput{}, fmt.Errorf("failed to parse signature R: %w", err)
	}

	sBytes, err := pusher.HexStringToByte32(update.StorkSignedPrice.TimestampedSignature.Signature.S)
	if err != nil {
		return bindings.TemporalNumericValueInput{}, fmt.Errorf("failed to parse signature S: %w", err)
	}

	// Parse encoded asset ID
	encodedAssetIDBytes, err := pusher.HexStringToByte32(string(update.StorkSignedPrice.EncodedAssetID))
	if err != nil {
		return bindings.TemporalNumericValueInput{}, fmt.Errorf("failed to parse encoded asset ID: %w", err)
	}

	// Parse publisher merkle root
	publisherMerkleRootBytes, err := pusher.HexStringToByte32(update.StorkSignedPrice.PublisherMerkleRoot)
	if err != nil {
		return bindings.TemporalNumericValueInput{}, fmt.Errorf("failed to parse publisher merkle root: %w", err)
	}

	// Parse value compute algorithm hash
	valueComputeAlgHashBytes, err := pusher.HexStringToByte32(update.StorkSignedPrice.StorkCalculationAlg.Checksum)
	if err != nil {
		return bindings.TemporalNumericValueInput{}, fmt.Errorf("failed to parse value compute alg hash: %w", err)
	}

	// Convert V from string to uint8 (remove "0x" prefix and parse as hex)
	vInt, err := strconv.ParseInt(update.StorkSignedPrice.TimestampedSignature.Signature.V[2:], 16, 8)
	if err != nil || vInt < 0 || vInt > 255 {
		return bindings.TemporalNumericValueInput{}, fmt.Errorf("failed to parse signature V: %w", err)
	}

	// Convert internal format to Fuel format
	return bindings.TemporalNumericValueInput{
		TemporalNumericValue: bindings.TemporalNumericValue{
			TimestampNs:    update.StorkSignedPrice.TimestampedSignature.TimestampNano,
			QuantizedValue: quantizedPriceBigInt,
		},
		ID:                  hex.EncodeToString(encodedAssetIDBytes[:]),
		PublisherMerkleRoot: hex.EncodeToString(publisherMerkleRootBytes[:]),
		ValueComputeAlgHash: hex.EncodeToString(valueComputeAlgHashBytes[:]),
		R:                   hex.EncodeToString(rBytes[:]),
		S:                   hex.EncodeToString(sBytes[:]),
		V:                   uint8(vInt),
	}, nil
}

func loadPrivateKey(keyFileContent []byte) (string, error) {
	// Extract private key from file content (assuming it's on the first line)
	privateKeyStr := string(keyFileContent)
	if len(privateKeyStr) > 0 && privateKeyStr[len(privateKeyStr)-1] == '\n' {
		privateKeyStr = privateKeyStr[:len(privateKeyStr)-1]
	}

	// Validate that private key is not empty
	if len(privateKeyStr) == 0 {
		return "", ErrPrivateKeyEmpty
	}

	return privateKeyStr, nil
}
