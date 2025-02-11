package chain_pusher

import (
	"fmt"
	"math/big"
	"os"
	"strings"

	contract "github.com/Stork-Oracle/stork-external/apps/lib/chain_pusher/contract_bindings/aptos"
	"github.com/rs/zerolog"
)

type AptosContractInteracter struct {
	logger   zerolog.Logger
	contract *contract.StorkContract

	pollingFrequencySec int
}

func NewAptosContractInteracter(rpcUrl, contractAddr, privateKeyFile string, assetConfigFile string, pollingFreqSec int, logger zerolog.Logger) (*AptosContractInteracter, error) {
	logger = logger.With().Str("component", "aptos-contract-interactor").Logger()

	keyFileContent, err := os.ReadFile(privateKeyFile)
	if err != nil {
		return nil, err
	}

	privateKey := strings.TrimSpace(strings.Split(string(keyFileContent), "\n")[0])

	contract, err := contract.NewStorkContract(rpcUrl, contractAddr, privateKey)
	if err != nil {
		return nil, err
	}

	return &AptosContractInteracter{
		logger:              logger,
		contract:            contract,
		pollingFrequencySec: pollingFreqSec,
	}, nil
}

// unfortunately, Aptos doesn't currently support websocket RPCs, so we can't listen to events from the contract
// the contract does emit events, so this can be implemented in the future if Aptos re-adds websocket support
func (aci *AptosContractInteracter) ListenContractEvents(ch chan map[InternalEncodedAssetId]InternalStorkStructsTemporalNumericValue) {
	aci.logger.Warn().Msg("Aptos does not currently support listening to events via websocket, falling back to polling")
}

func (aci *AptosContractInteracter) PullValues(encodedAssetIds []InternalEncodedAssetId) (map[InternalEncodedAssetId]InternalStorkStructsTemporalNumericValue, error) {
	// convert to bindings EncodedAssetId
	bindingsEncodedAssetIds := []contract.EncodedAssetId{}
	for _, encodedAssetId := range encodedAssetIds {
		bindingsEncodedAssetIds = append(bindingsEncodedAssetIds, contract.EncodedAssetId(encodedAssetId))
	}
	values, err := aci.contract.GetMultipleTemporalNumericValuesUnchecked(bindingsEncodedAssetIds)
	aci.logger.Debug().Msgf("successfully pulled %d values from contract", len(values))
	if err != nil {
		return nil, err
	}

	// convert to map[InternalEncodedAssetId]InternalStorkStructsTemporalNumericValue
	result := make(map[InternalEncodedAssetId]InternalStorkStructsTemporalNumericValue)
	for _, encodedAssetId := range encodedAssetIds {
		if value, ok := values[contract.EncodedAssetId(encodedAssetId)]; ok {

			magnitude := value.QuantizedValue.Magnitude
			negative := value.QuantizedValue.Negative
			signMultiplier := 1
			if negative {
				signMultiplier = -1
			}
			quantizedValue := new(big.Int).Mul(magnitude, big.NewInt(int64(signMultiplier)))

			result[encodedAssetId] = InternalStorkStructsTemporalNumericValue{
				TimestampNs:    value.TimestampNs,
				QuantizedValue: quantizedValue,
			}
		}
	}
	return result, nil
}

func (aci *AptosContractInteracter) BatchPushToContract(priceUpdates map[InternalEncodedAssetId]AggregatedSignedPrice) error {

	var updateData []contract.UpdateData
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

func (aci *AptosContractInteracter) aggregatedSignedPriceToAptosUpdateData(price AggregatedSignedPrice) (contract.UpdateData, error) {
	signedPrice := price.StorkSignedPrice
	assetId, err := hexStringToByteArray(string(signedPrice.EncodedAssetId))
	if err != nil {
		return contract.UpdateData{}, fmt.Errorf("failed to convert encoded asset id to byte array: %w", err)
	}
	timestampNs := uint64(signedPrice.TimestampedSignature.Timestamp)
	magnitude_string := string(signedPrice.QuantizedPrice)
	magnitude, ok := new(big.Int).SetString(magnitude_string, 10)
	if !ok {
		return contract.UpdateData{}, fmt.Errorf("failed to convert quantized price to big int")
	}
	negative := magnitude.Sign() == -1
	magnitude.Abs(magnitude)

	publisherMerkleRoot, err := hexStringToByteArray(signedPrice.PublisherMerkleRoot)
	if err != nil {
		return contract.UpdateData{}, fmt.Errorf("failed to convert publisher merkle root to byte array: %w", err)
	}

	valueComputeAlgHash, err := hexStringToByteArray(signedPrice.StorkCalculationAlg.Checksum)
	if err != nil {
		return contract.UpdateData{}, fmt.Errorf("failed to convert value compute alg hash to byte array: %w", err)
	}

	r, err := hexStringToByteArray(signedPrice.TimestampedSignature.Signature.R)
	if err != nil {
		return contract.UpdateData{}, fmt.Errorf("failed to convert R to byte array: %w", err)
	}

	s, err := hexStringToByteArray(signedPrice.TimestampedSignature.Signature.S)
	if err != nil {
		return contract.UpdateData{}, fmt.Errorf("failed to convert S to byte array: %w", err)
	}

	vBytes, err := hexStringToByteArray(signedPrice.TimestampedSignature.Signature.V)
	if err != nil {
		return contract.UpdateData{}, fmt.Errorf("failed to convert V to byte array: %w", err)
	}
	v := byte(vBytes[0])

	return contract.UpdateData{
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
