package aptos

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/Stork-Oracle/stork-external/apps/chain_pusher/pkg/aptos/bindings"
	"github.com/Stork-Oracle/stork-external/apps/chain_pusher/pkg/pusher"
	"github.com/Stork-Oracle/stork-external/apps/chain_pusher/pkg/types"
	"github.com/aptos-labs/aptos-go-sdk/crypto"
	"github.com/rs/zerolog"
)

type ContractInteractor struct {
	logger   zerolog.Logger
	contract *bindings.StorkContract

	pollingPeriodSec int
}

func NewContractInteractor(
	rpcUrl string,
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

	contract, err := bindings.NewStorkContract(rpcUrl, contractAddr, privateKey)
	if err != nil {
		return nil, err
	}

	return &ContractInteractor{
		logger:           logger,
		contract:         contract,
		pollingPeriodSec: pollingPeriodSec,
	}, nil
}

// unfortunately, Aptos doesn't currently support websocket RPCs, so we can't listen to events from the contract
// the contract does emit events, so this can be implemented in the future if Aptos re-adds websocket support
func (aci *ContractInteractor) ListenContractEvents(
	ctx context.Context, ch chan map[types.InternalEncodedAssetId]types.InternalTemporalNumericValue,
) {
	aci.logger.Warn().Msg("Aptos does not currently support listening to events via websocket, falling back to polling")
}

func (aci *ContractInteractor) PullValues(encodedAssetIds []types.InternalEncodedAssetId) (map[types.InternalEncodedAssetId]types.InternalTemporalNumericValue, error) {
	// convert to bindings EncodedAssetId
	bindingsEncodedAssetIds := []bindings.EncodedAssetId{}
	for _, encodedAssetId := range encodedAssetIds {
		bindingsEncodedAssetIds = append(bindingsEncodedAssetIds, bindings.EncodedAssetId(encodedAssetId))
	}
	values, err := aci.contract.GetMultipleTemporalNumericValuesUnchecked(bindingsEncodedAssetIds)
	aci.logger.Debug().Msgf("successfully pulled %d values from contract", len(values))
	if err != nil {
		return nil, err
	}

	// convert to map[InternalEncodedAssetId]InternalStorkStructsTemporalNumericValue
	result := make(map[types.InternalEncodedAssetId]types.InternalTemporalNumericValue)
	for _, encodedAssetId := range encodedAssetIds {
		if value, ok := values[bindings.EncodedAssetId(encodedAssetId)]; ok {

			magnitude := value.QuantizedValue.Magnitude
			negative := value.QuantizedValue.Negative
			signMultiplier := 1
			if negative {
				signMultiplier = -1
			}
			quantizedValue := new(big.Int).Mul(magnitude, big.NewInt(int64(signMultiplier)))

			result[encodedAssetId] = types.InternalTemporalNumericValue{
				TimestampNs:    value.TimestampNs,
				QuantizedValue: quantizedValue,
			}
		}
	}
	return result, nil
}

func (aci *ContractInteractor) BatchPushToContract(priceUpdates map[types.InternalEncodedAssetId]types.AggregatedSignedPrice) error {
	var updateData []bindings.UpdateData
	for _, price := range priceUpdates {
		update, err := aci.aggregatedSignedPriceToAptosUpdateData(price)
		if err != nil {
			return err
		}
		updateData = append(updateData, update)
	}
	hash, err := aci.contract.UpdateMultipleTemporalNumericValuesEvm(updateData)
	if err != nil {
		aci.logger.Error().Err(err).Msg("failed to update multiple temporal numeric values")
		return err
	}
	aci.logger.Info().
		Int("numUpdates", len(priceUpdates)).
		Str("txnHash", hash).
		Msg("Successfully pushed batch update to contract")
	return nil
}

// todo: implement
func (aci *ContractInteractor) GetWalletBalance() (float64, error) {
	return -1, nil
}

func (aci *ContractInteractor) aggregatedSignedPriceToAptosUpdateData(price types.AggregatedSignedPrice) (bindings.UpdateData, error) {
	signedPrice := price.StorkSignedPrice
	assetId, err := pusher.HexStringToByteArray(string(signedPrice.EncodedAssetId))
	if err != nil {
		return bindings.UpdateData{}, fmt.Errorf("failed to convert encoded asset id to byte array: %w", err)
	}
	timestampNs := uint64(signedPrice.TimestampedSignature.TimestampNano)
	magnitude_string := string(signedPrice.QuantizedPrice)
	magnitude, ok := new(big.Int).SetString(magnitude_string, 10)
	if !ok {
		return bindings.UpdateData{}, fmt.Errorf("failed to convert quantized price to big int")
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
	v := byte(vBytes[0])

	return bindings.UpdateData{
		Id:                              assetId,
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
		return nil, err
	}
	privateKey := &crypto.Ed25519PrivateKey{}
	err = privateKey.FromHex(formattedPrivateKey)
	if err != nil {
		return nil, err
	}
	return privateKey, nil
}
