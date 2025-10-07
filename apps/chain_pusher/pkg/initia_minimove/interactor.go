package initia_minimove

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/Stork-Oracle/stork-external/apps/chain_pusher/pkg/initia_minimove/bindings"
	"github.com/Stork-Oracle/stork-external/apps/chain_pusher/pkg/pusher"
	"github.com/Stork-Oracle/stork-external/apps/chain_pusher/pkg/types"
	"github.com/rs/zerolog"
)

var ErrFailedToConvertQuantizedPriceToBigInt = errors.New("failed to convert quantized price to big int")

type ContractInteractor struct {
	logger   zerolog.Logger
	contract *bindings.StorkContract

	pollingPeriodSec int
}

func NewContractInteractor(
	rpcUrl string,
	contractAddr string,
	mnemonic []byte,
	pollingPeriodSec int,
	logger zerolog.Logger,
	gasPrice float64,
	gasAdjustment float64,
	denom string,
	chainID string,
) (*ContractInteractor, error) {
	logger = logger.With().Str("component", "initia-minimove-contract-interactor").Logger()

	mnemonicString := strings.TrimSpace(string(mnemonic))

	contract, err := bindings.NewStorkContract(
		rpcUrl,
		contractAddr,
		mnemonicString,
		gasPrice,
		gasAdjustment,
		denom,
		chainID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create stork contract: %w", err)
	}

	return &ContractInteractor{
		logger:           logger,
		contract:         contract,
		pollingPeriodSec: pollingPeriodSec,
	}, nil
}

// ListenContractEvents is a placeholder function for the contract interactor.
// unfortunately, Initia MiniMove doesn't currently support websocket RPCs, so we can't listen to events from the contract
// this contract does emit events, so this can be implemented in the future if Initia re-adds websocket support.
func (ici *ContractInteractor) ListenContractEvents(
	ctx context.Context, ch chan map[types.InternalEncodedAssetID]types.InternalTemporalNumericValue,
) {
	ici.logger.Warn().Msg("Initia MiniMove does not currently support listening to events via websocket, falling back to polling")
}

func (ici *ContractInteractor) PullValues(
	encodedAssetIDs []types.InternalEncodedAssetID,
) (map[types.InternalEncodedAssetID]types.InternalTemporalNumericValue, error) {
	polledVals := make(map[types.InternalEncodedAssetID]types.InternalTemporalNumericValue)

	for _, encodedAssetID := range encodedAssetIDs {
		value, err := ici.contract.GetTemporalNumericValueUnchecked(encodedAssetID[:])
		if err != nil {
			if errors.Is(err, bindings.ErrFeedNotFound) {
				ici.logger.Warn().Err(err).Str("assetID", hex.EncodeToString(encodedAssetID[:])).Msg("No value found")
			} else {
				ici.logger.Warn().Err(err).Str("assetID", hex.EncodeToString(encodedAssetID[:])).Msg("Failed to get latest value")
			}

			continue
		}

		magnitude := value.QuantizedValue.Magnitude
		negative := value.QuantizedValue.Negative

		signMultiplier := 1
		if negative {
			signMultiplier = -1
		}

		quantizedValue := new(big.Int).Mul(magnitude, big.NewInt(int64(signMultiplier)))

		polledVals[encodedAssetID] = types.InternalTemporalNumericValue{
			TimestampNs:    value.TimestampNs,
			QuantizedValue: quantizedValue,
		}
	}

	ici.logger.Debug().Msgf("Pulled %d values from contract", len(polledVals))

	return polledVals, nil
}

func (ici *ContractInteractor) BatchPushToContract(
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

	hash, err := ici.contract.UpdateMultipleTemporalNumericValuesEvm(updateData)
	if err != nil {
		ici.logger.Error().Err(err).Msg("failed to update multiple temporal numeric values")

		return fmt.Errorf("failed to update multiple temporal numeric values: %w", err)
	}

	ici.logger.Info().
		Int("numUpdates", len(priceUpdates)).
		Str("txnHash", hash).
		Msg("Successfully pushed batch update to contract")

	return nil
}

// GetWalletBalance is a placeholder function to get the balance of the wallet being used to push to the contract.
// todo: implement
//
//nolint:godox // This function has unmet criteria to be implemented.
func (ici *ContractInteractor) GetWalletBalance() (float64, error) {
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
