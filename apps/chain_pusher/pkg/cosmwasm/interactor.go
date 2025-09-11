package cosmwasm

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/Stork-Oracle/stork-external/apps/chain_pusher/pkg/cosmwasm/bindings"
	"github.com/Stork-Oracle/stork-external/apps/chain_pusher/pkg/pusher"
	"github.com/Stork-Oracle/stork-external/apps/chain_pusher/pkg/types"
	"github.com/rs/zerolog"
)

var ErrFailedToConvertBigInt = errors.New("failed to convert Uint128 string to big.Int")

type ContractInteractor struct {
	logger   zerolog.Logger
	contract *bindings.StorkContract

	pollingPeriodSec int
}

func NewContractInteractor(
	chainGrpcUrl string,
	contractAddress string,
	mnemonic []byte,
	batchingWindow int,
	pollingPeriod int,
	logger zerolog.Logger,
	gasPrice float64,
	gasAdjustment float64,
	denom string,
	chainID string,
	chainPrefix string,
) (*ContractInteractor, error) {
	logger = logger.With().Str("component", "cosmwasm-contract-interactor").Logger()

	mnemonicString := strings.TrimSpace(string(mnemonic))

	contract, err := bindings.NewStorkContract(
		chainGrpcUrl,
		contractAddress,
		mnemonicString,
		gasPrice,
		gasAdjustment,
		denom,
		chainID,
		chainPrefix,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create stork contract: %w", err)
	}

	return &ContractInteractor{
		logger:           logger,
		contract:         contract,
		pollingPeriodSec: pollingPeriod,
	}, nil
}

// ListenContractEvents is a placeholder function for the contract interactor.
// unfortunately, Cosmwasm doesn't currently support websocket RPCs, so we can't listen to events from the contract
// this contract does emit events, so this can be implemented in the future if Cosmwasm re-adds websocket support.
func (sci *ContractInteractor) ListenContractEvents(
	ctx context.Context, ch chan map[types.InternalEncodedAssetID]types.InternalTemporalNumericValue,
) {
	sci.logger.Warn().
		Msg("Cosmwasm pusher does not currently support listening to events via websocket, falling back to polling")
}

func (sci *ContractInteractor) PullValues(
	encodedAssetIDs []types.InternalEncodedAssetID,
) (map[types.InternalEncodedAssetID]types.InternalTemporalNumericValue, error) {
	polledVals := make(map[types.InternalEncodedAssetID]types.InternalTemporalNumericValue)

	for _, encodedAssetID := range encodedAssetIDs {
		var encodeAssetIDInt [32]int
		for i, b := range encodedAssetID {
			encodeAssetIDInt[i] = int(b)
		}

		response, err := sci.contract.GetLatestCanonicalTemporalNumericValueUnchecked(encodeAssetIDInt)
		if err != nil {
			continue
		}

		quantizedValueBigInt := new(big.Int)

		//nolint:mnd // base number.
		quantizedValueBigInt, ok := quantizedValueBigInt.SetString(
			string(response.TemporalNumericValue.QuantizedValue),
			10,
		)
		if !ok {
			return nil, ErrFailedToConvertBigInt
		}

		timestampNs, err := strconv.ParseUint(string(response.TemporalNumericValue.TimestampNs), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse timestamp ns: %w", err)
		}

		polledVals[encodedAssetID] = types.InternalTemporalNumericValue{
			TimestampNs:    timestampNs,
			QuantizedValue: quantizedValueBigInt,
		}
	}

	sci.logger.Debug().Msgf("Pulled %d values from contract", len(polledVals))

	return polledVals, nil
}

func (sci *ContractInteractor) BatchPushToContract(
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

	txHash, err := sci.contract.UpdateTemporalNumericValuesEvm(updateData)
	if err != nil {
		return fmt.Errorf("failed to update temporal numeric values: %w", err)
	}

	sci.logger.Info().
		Int("numUpdates", len(priceUpdates)).
		Str("txHash", txHash).
		Msg("Successfully pushed batch update to contract")

	return nil
}

// GetWalletBalance is a placeholder function to get the balance of the wallet being used to push to the contract.
// todo: implement
//
//nolint:godox // This function has unmet criteria to be implemented.
func (sci *ContractInteractor) GetWalletBalance() (float64, error) {
	return -1, nil
}

func aggregatedSignedPriceToUpdateData(
	price types.AggregatedSignedPrice,
) (bindings.UpdateData, error) {
	signedPrice := price.StorkSignedPrice

	assetID, err := pusher.HexStringToInt32(string(signedPrice.EncodedAssetID))
	if err != nil {
		return bindings.UpdateData{}, fmt.Errorf("failed to convert encoded asset id to byte array: %w", err)
	}

	timestampNs := strconv.FormatUint(signedPrice.TimestampedSignature.TimestampNano, 10)
	quantizedValue := string(signedPrice.QuantizedPrice)
	temporalNumericValue := bindings.TemporalNumericValue{
		QuantizedValue: bindings.Int128(quantizedValue),
		TimestampNs:    bindings.Uint64(timestampNs),
	}

	valueComputeAlgHash, err := pusher.HexStringToInt32(signedPrice.StorkCalculationAlg.Checksum)
	if err != nil {
		return bindings.UpdateData{}, fmt.Errorf("failed to convert value compute alg hash to byte array: %w", err)
	}

	publisherMerkleRoot, err := pusher.HexStringToInt32(signedPrice.PublisherMerkleRoot)
	if err != nil {
		return bindings.UpdateData{}, fmt.Errorf("failed to convert publisher merkle root to byte array: %w", err)
	}

	r, err := pusher.HexStringToInt32(signedPrice.TimestampedSignature.Signature.R)
	if err != nil {
		return bindings.UpdateData{}, fmt.Errorf("failed to convert R to byte array: %w", err)
	}

	s, err := pusher.HexStringToInt32(signedPrice.TimestampedSignature.Signature.S)
	if err != nil {
		return bindings.UpdateData{}, fmt.Errorf("failed to convert S to byte array: %w", err)
	}

	vInt, err := strconv.ParseInt(signedPrice.TimestampedSignature.Signature.V[2:], 16, 8)
	if err != nil {
		return bindings.UpdateData{}, fmt.Errorf("failed to convert V to byte array: %w", err)
	}

	v := int(vInt)

	return bindings.UpdateData{
		ID:                   assetID,
		TemporalNumericValue: temporalNumericValue,
		ValueComputeAlgHash:  valueComputeAlgHash,
		PublisherMerkleRoot:  publisherMerkleRoot,
		R:                    r,
		S:                    s,
		V:                    v,
	}, nil
}
