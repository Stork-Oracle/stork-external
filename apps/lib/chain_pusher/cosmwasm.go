package chain_pusher

import (
	"errors"
	"fmt"
	"math/big"
	"os"
	"strconv"
	"strings"

	contract "github.com/Stork-Oracle/stork-external/apps/lib/chain_pusher/contract_bindings/cosmwasm"
	"github.com/rs/zerolog"
)

type CosmwasmContractInteracter struct {
	logger   zerolog.Logger
	contract *contract.StorkContract

	pollingFrequencySec int
}

func NewCosmwasmContractInteracter(chainGrpcUrl, contractAddress, mnemonicFile string, batchingWindow, pollingFrequency int, logger zerolog.Logger, gasPrice float64, gasAdjustment float64, denom string, chainID string, chainPrefix string) (*CosmwasmContractInteracter, error) {
	logger = logger.With().Str("component", "cosmwasm-contract-interacter").Logger()

	mnemonic, err := os.ReadFile(mnemonicFile)
	if err != nil {
		return nil, err
	}
	mnemonicString := strings.TrimSpace(string(mnemonic))
	logger.Info().Msg("Creating cosmwasm contract interacter")
	contract, err := contract.NewStorkContract(chainGrpcUrl, contractAddress, mnemonicString, gasPrice, gasAdjustment, denom, chainID, chainPrefix)
	if err != nil {
		return nil, err
	}
	return &CosmwasmContractInteracter{
		logger:              logger,
		contract:            contract,
		pollingFrequencySec: pollingFrequency,
	}, nil
}

func (sci *CosmwasmContractInteracter) ListenContractEvents(ch chan map[InternalEncodedAssetId]InternalStorkStructsTemporalNumericValue) {
	sci.logger.Warn().Msg("Cosmwasm pusher does not currently support listening to events via websocket, falling back to polling")
}

func (sci *CosmwasmContractInteracter) PullValues(encodedAssetIds []InternalEncodedAssetId) (map[InternalEncodedAssetId]InternalStorkStructsTemporalNumericValue, error) {
	polledVals := make(map[InternalEncodedAssetId]InternalStorkStructsTemporalNumericValue)
	for _, encodedAssetId := range encodedAssetIds {
		//convert [32]byte to [32]int
		// Convert [32]byte to [32]int using copy and type conversion
		var encodeAssetIdInt [32]int
		for i, b := range encodedAssetId {
			encodeAssetIdInt[i] = int(b)
		}
		response, err := sci.contract.GetLatestCanonicalTemporalNumericValueUnchecked(encodeAssetIdInt)
		if err != nil {
			return nil, err
		}
		quantizedValueBigInt := new(big.Int)
		quantizedValueBigInt, ok := quantizedValueBigInt.SetString(string(response.TemporalNumericValue.QuantizedValue), 10)
		if !ok {
			return nil, errors.New("failed to convert Uint128 string to big.Int")
		}
		timestampNs, err := strconv.ParseUint(string(response.TemporalNumericValue.TimestampNs), 10, 64)
		if err != nil {
			return nil, err
		}
		polledVals[encodedAssetId] = InternalStorkStructsTemporalNumericValue{
			TimestampNs:    timestampNs,
			QuantizedValue: quantizedValueBigInt,
		}
	}
	return polledVals, nil
}

func (sci *CosmwasmContractInteracter) BatchPushToContract(priceUpdates map[InternalEncodedAssetId]AggregatedSignedPrice) error {

	var updateData []contract.UpdateData
	for _, price := range priceUpdates {
		update, err := sci.aggregatedSignedPriceToUpdateData(price)
		if err != nil {
			return err
		}
		updateData = append(updateData, update)
	}
	txHash, err := sci.contract.UpdateTemporalNumericValuesEvm(updateData)
	if err != nil {
		return err
	}
	sci.logger.Info().
		Int("numUpdates", len(priceUpdates)).
		Str("txHash", txHash).
		Msg("Successfully pushed batch update to contract")
	return nil
}

func (sci *CosmwasmContractInteracter) aggregatedSignedPriceToUpdateData(price AggregatedSignedPrice) (contract.UpdateData, error) {
	signedPrice := price.StorkSignedPrice
	assetId, err := hexStringToIntArray(string(signedPrice.EncodedAssetId))
	if err != nil {
		return contract.UpdateData{}, fmt.Errorf("failed to convert encoded asset id to byte array: %w", err)
	}
	timestampNs := uint64(signedPrice.TimestampedSignature.Timestamp)
	quantizedValue := string(signedPrice.QuantizedPrice)
	temporalNumericValue := contract.TemporalNumericValue{
		QuantizedValue: contract.Int128(quantizedValue),
		TimestampNs:    contract.Uint64(timestampNs),
	}
	valueComputeAlgHash, err := hexStringToIntArray(signedPrice.StorkCalculationAlg.Checksum)
	if err != nil {
		return contract.UpdateData{}, fmt.Errorf("failed to convert value compute alg hash to byte array: %w", err)
	}
	publisherMerkleRoot, err := hexStringToIntArray(signedPrice.PublisherMerkleRoot)
	if err != nil {
		return contract.UpdateData{}, fmt.Errorf("failed to convert publisher merkle root to byte array: %w", err)
	}
	r, err := hexStringToIntArray(signedPrice.TimestampedSignature.Signature.R)
	if err != nil {
		return contract.UpdateData{}, fmt.Errorf("failed to convert R to byte array: %w", err)
	}
	s, err := hexStringToIntArray(signedPrice.TimestampedSignature.Signature.S)
	if err != nil {
		return contract.UpdateData{}, fmt.Errorf("failed to convert S to byte array: %w", err)
	}
	vInts, err := hexStringToIntArray(signedPrice.TimestampedSignature.Signature.V)
	if err != nil {
		return contract.UpdateData{}, fmt.Errorf("failed to convert V to byte array: %w", err)
	}
	v := vInts[0]
	return contract.UpdateData{
		Id:                   assetId,
		TemporalNumericValue: temporalNumericValue,
		ValueComputeAlgHash:  valueComputeAlgHash,
		PublisherMerkleRoot:  publisherMerkleRoot,
		R:                    r,
		S:                    s,
		V:                    v,
	}, nil
}

func hexStringToIntArray(hexString string) ([32]int, error) {
	bytes, err := hexStringToByteArray(hexString)
	if err != nil {
		return [32]int{}, fmt.Errorf("failed to convert hex string to byte array: %w", err)
	}
	var result [32]int
	for i, b := range bytes {
		result[i] = int(b)
	}
	return result, nil
}
