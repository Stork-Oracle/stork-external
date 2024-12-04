package chain_pusher

import (
	"fmt"
	"math/big"
	"os"
	"strings"

	contract "github.com/Stork-Oracle/stork-external/apps/lib/chain_pusher/contract_bindings/sui"
	"github.com/rs/zerolog"
)

type SuiContractInteracter struct {
	logger   zerolog.Logger
	contract *contract.StorkContract

	pollingFrequencySec int
}

func NewSuiContractInteracter(rpcUrl, contractAddr, privateKeyFile string, assetConfigFile string, pollingFreqSec int, logger zerolog.Logger) (*SuiContractInteracter, error) {
	logger = logger.With().Str("component", "sui-contract-interactor").Logger()

	keyFileContent, err := os.ReadFile(privateKeyFile)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(keyFileContent), "\n")
	var privateKey string
	for _, line := range lines {
		if strings.HasPrefix(line, "keypair:") {
			privateKey = strings.TrimSpace(line[len("keypair:"):])
			break
		}
	}
	if privateKey == "" && len(lines) == 1 {
		privateKey = strings.TrimSpace(lines[0])
	}
	contract, err := contract.NewStorkContract(rpcUrl, contractAddr, privateKey)
	if err != nil {
		return nil, err
	}
	return &SuiContractInteracter{
		logger:              logger,
		contract:            contract,
		pollingFrequencySec: pollingFreqSec,
	}, nil
}

// unfortunately, Sui doesn't currently support websocket RPCs, so we can't listen to events from the contract
// the contract does emit events, so this can be implemented in the future if Sui re-adds websocket support
func (sci *SuiContractInteracter) ListenContractEvents(ch chan map[InternalEncodedAssetId]InternalStorkStructsTemporalNumericValue) {
	sci.logger.Warn().Msg("Sui does not currently support listening to events via websocket, falling back to polling")
}

func (sci *SuiContractInteracter) PullValues(encodedAssetIds []InternalEncodedAssetId) (map[InternalEncodedAssetId]InternalStorkStructsTemporalNumericValue, error) {
	// convert to bindings EncodedAssetId
	bindingsEncodedAssetIds := []contract.EncodedAssetId{}
	for _, encodedAssetId := range encodedAssetIds {
		bindingsEncodedAssetIds = append(bindingsEncodedAssetIds, contract.EncodedAssetId(encodedAssetId))
	}
	values, err := sci.contract.GetMultipleTemporalNumericValuesUnchecked(bindingsEncodedAssetIds)
	if err != nil {
		return nil, err
	}
	sci.logger.Debug().Msgf("successfully pulled %d values from contract", len(values))

	// convert to map[InternalEncodedAssetId]InternalStorkStructsTemporalNumericValue
	result := make(map[InternalEncodedAssetId]InternalStorkStructsTemporalNumericValue)
	for _, encodedAssetId := range encodedAssetIds {
		if value, ok := values[contract.EncodedAssetId(encodedAssetId)]; ok {
			result[encodedAssetId] = temporalNumericValueToInternal(value)
		}
	}
	return result, nil
}

func (sci *SuiContractInteracter) BatchPushToContract(priceUpdates map[InternalEncodedAssetId]AggregatedSignedPrice) error {

	var updateData []contract.UpdateData
	for _, price := range priceUpdates {
		update, err := sci.aggregatedSignedPriceToUpdateData(price)
		if err != nil {
			return err
		}
		updateData = append(updateData, update)
	}
	err := sci.contract.UpdateMultipleTemporalNumericValuesEvm(updateData)
	if err != nil {
		sci.logger.Error().Err(err).Msg("failed to update multiple temporal numeric values")
		return err
	}
	sci.logger.Info().Msg("successfully updated multiple temporal numeric values")
	return nil
}

func temporalNumericValueToInternal(value contract.TemporalNumericValue) InternalStorkStructsTemporalNumericValue {
	magnitude := value.QuantizedValue.Magnitude
	negative := value.QuantizedValue.Negative
	signMultiplier := 1
	if negative {
		signMultiplier = -1
	}
	quantizedValue := new(big.Int).Mul(magnitude, big.NewInt(int64(signMultiplier)))

	return InternalStorkStructsTemporalNumericValue{
		TimestampNs:    value.TimestampNs,
		QuantizedValue: quantizedValue,
	}
}

func (sci *SuiContractInteracter) aggregatedSignedPriceToUpdateData(price AggregatedSignedPrice) (contract.UpdateData, error) {
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
