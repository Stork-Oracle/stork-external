package aptos

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/Stork-Oracle/stork-external/apps/chain_pusher/pkg/aptos/bindings"
	"github.com/Stork-Oracle/stork-external/apps/chain_pusher/pkg/pusher"
	"github.com/Stork-Oracle/stork-external/apps/chain_pusher/pkg/types"
	"github.com/aptos-labs/aptos-go-sdk/crypto"
	"github.com/rs/zerolog"
)

var ErrFailedToConvertQuantizedPriceToBigInt = errors.New("failed to convert quantized price to big int")

type ContractInteractor struct {
	logger zerolog.Logger

	pollingPeriodSec int
	privateKey       *crypto.Ed25519PrivateKey
	contractAddress  string

	contract *bindings.StorkContract
}

func NewContractInteractor(
	contractAddr string,
	keyFileContent []byte,
	pollingPeriodSec int,
	logger zerolog.Logger,
) (*ContractInteractor, error) {
	logger = logger.With().Str("component", "aptos-contract-interactor").Logger()

	privateKey, err := loadPrivateKey(keyFileContent)
	if err != nil {
		return nil, err
	}

	return &ContractInteractor{
		logger:           logger,
		pollingPeriodSec: pollingPeriodSec,
		privateKey:       privateKey,
		contractAddress:  contractAddr,
	}, nil
}

func (aci *ContractInteractor) ConnectRest(url string) error {
	contract, err := bindings.NewStorkContract(url, aci.contractAddress, aci.privateKey)
	if err != nil {
		return fmt.Errorf("failed to create stork contract: %w", err)
	}
	aci.contract = contract
	return nil
}

func (aci *ContractInteractor) ConnectWs(url string) error {
	// not implemented
	return nil
}

// ListenContractEvents is a placeholder function for the contract interactor.
// unfortunately, Aptos doesn't currently support websocket RPCs, so we can't listen to events from the contract
// this contract does emit events, so this can be implemented in the future if Aptos re-adds websocket support.
func (aci *ContractInteractor) ListenContractEvents(
	ctx context.Context, ch chan map[types.InternalEncodedAssetID]types.InternalTemporalNumericValue,
) {
	aci.logger.Warn().Msg("Aptos does not currently support listening to events via websocket, falling back to polling")
}

func (aci *ContractInteractor) PullValues(
	encodedAssetIDs []types.InternalEncodedAssetID,
) (map[types.InternalEncodedAssetID]types.InternalTemporalNumericValue, error) {
	// convert to bindings EncodedAssetID
	bindingsEncodedAssetIDs := []bindings.EncodedAssetID{}
	for _, encodedAssetID := range encodedAssetIDs {
		bindingsEncodedAssetIDs = append(bindingsEncodedAssetIDs, bindings.EncodedAssetID(encodedAssetID))
	}

	values, err := aci.contract.GetMultipleTemporalNumericValuesUnchecked(bindingsEncodedAssetIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get multiple temporal numeric values: %w", err)
	}

	aci.logger.Debug().Msgf("successfully pulled %d values from contract", len(values))
	// convert to map[InternalEncodedAssetID]InternalStorkStructsTemporalNumericValue
	result := make(map[types.InternalEncodedAssetID]types.InternalTemporalNumericValue)

	for _, encodedAssetID := range encodedAssetIDs {
		if value, ok := values[bindings.EncodedAssetID(encodedAssetID)]; ok {
			magnitude := value.QuantizedValue.Magnitude
			negative := value.QuantizedValue.Negative

			signMultiplier := 1
			if negative {
				signMultiplier = -1
			}

			quantizedValue := new(big.Int).Mul(magnitude, big.NewInt(int64(signMultiplier)))

			result[encodedAssetID] = types.InternalTemporalNumericValue{
				TimestampNs:    value.TimestampNs,
				QuantizedValue: quantizedValue,
			}
		}
	}

	return result, nil
}

func (aci *ContractInteractor) BatchPushToContract(
	priceUpdates map[types.InternalEncodedAssetID]types.AggregatedSignedPrice,
) error {
	updateData := make([]bindings.UpdateData, 0, len(priceUpdates))

	for _, price := range priceUpdates {
		update, err := aggregatedSignedPriceToUpdateData(price)
		if err != nil {
			return err
		}

		updateData = append(updateData, update)
	}

	hash, err := aci.contract.UpdateMultipleTemporalNumericValuesEvm(updateData)
	if err != nil {
		aci.logger.Error().Err(err).Msg("failed to update multiple temporal numeric values")

		return fmt.Errorf("failed to update multiple temporal numeric values: %w", err)
	}

	aci.logger.Info().
		Int("numUpdates", len(priceUpdates)).
		Str("txnHash", hash).
		Msg("Successfully pushed batch update to contract")

	return nil
}

// GetWalletBalance is a placeholder function to get the balance of the wallet being used to push to the contract.
// todo: implement
//
//nolint:godox // This function has unmet criteria to be implemented.
func (aci *ContractInteractor) GetWalletBalance() (float64, error) {
	return -1, nil
}

func aggregatedSignedPriceToUpdateData(
	price types.AggregatedSignedPrice,
) (bindings.UpdateData, error) {
	signedPrice := price.StorkSignedPrice

	assetID, err := pusher.HexStringToByteArray(string(signedPrice.EncodedAssetID))
	if err != nil {
		return bindings.UpdateData{}, fmt.Errorf("failed to convert encoded asset id to byte array: %w", err)
	}

	timestampNs := signedPrice.TimestampedSignature.TimestampNano
	magnitudeString := string(signedPrice.QuantizedPrice)

	//nolint:mnd // base number.
	magnitude, ok := new(big.Int).SetString(magnitudeString, 10)
	if !ok {
		return bindings.UpdateData{}, ErrFailedToConvertQuantizedPriceToBigInt
	}

	negative := magnitude.Sign() == -1
	magnitude.Abs(magnitude)

	publisherMerkleRoot, err := pusher.HexStringToByteArray(signedPrice.PublisherMerkleRoot)
	if err != nil {
		return bindings.UpdateData{}, fmt.Errorf("failed to convert publisher merkle root to byte array: %w", err)
	}

	valueComputeAlgHash, err := pusher.HexStringToByteArray(signedPrice.StorkCalculationAlg.Checksum)
	if err != nil {
		return bindings.UpdateData{}, fmt.Errorf("failed to convert value compute alg hash to byte array: %w", err)
	}

	r, err := pusher.HexStringToByteArray(signedPrice.TimestampedSignature.Signature.R)
	if err != nil {
		return bindings.UpdateData{}, fmt.Errorf("failed to convert R to byte array: %w", err)
	}

	s, err := pusher.HexStringToByteArray(signedPrice.TimestampedSignature.Signature.S)
	if err != nil {
		return bindings.UpdateData{}, fmt.Errorf("failed to convert S to byte array: %w", err)
	}

	vBytes, err := pusher.HexStringToByteArray(signedPrice.TimestampedSignature.Signature.V)
	if err != nil {
		return bindings.UpdateData{}, fmt.Errorf("failed to convert V to byte array: %w", err)
	}

	v := vBytes[0]

	return bindings.UpdateData{
		ID:                              assetID,
		TemporalNumericValueTimestampNs: timestampNs,
		TemporalNumericValueMagnitude:   magnitude,
		TemporalNumericValueNegative:    negative,
		PublisherMerkleRoot:             publisherMerkleRoot,
		ValueComputeAlgHash:             valueComputeAlgHash,
		R:                               r,
		S:                               s,
		V:                               v,
	}, nil
}

func loadPrivateKey(keyFileContent []byte) (*crypto.Ed25519PrivateKey, error) {
	trimmedKey := strings.TrimSpace(strings.Split(string(keyFileContent), "\n")[0])

	formattedPrivateKey, err := crypto.FormatPrivateKey(trimmedKey, crypto.PrivateKeyVariantEd25519)
	if err != nil {
		return nil, fmt.Errorf("failed to format private key: %w", err)
	}

	privateKey := &crypto.Ed25519PrivateKey{}

	err = privateKey.FromHex(formattedPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to set private key from hex: %w", err)
	}

	return privateKey, nil
}
