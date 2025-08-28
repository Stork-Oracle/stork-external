package sui

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/Stork-Oracle/stork-external/apps/chain_pusher/pkg/pusher"
	"github.com/Stork-Oracle/stork-external/apps/chain_pusher/pkg/sui/bindings"
	"github.com/Stork-Oracle/stork-external/apps/chain_pusher/pkg/types"
	"github.com/coming-chat/go-sui/v2/account"
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
	assetConfigFile string,
	pollingPeriodSec int,
	logger zerolog.Logger,
) (*ContractInteractor, error) {
	logger = logger.With().Str("component", "sui-contract-interactor").Logger()

	account, err := loadPrivateKey(keyFileContent)
	if err != nil {
		return nil, err
	}
	contract, err := bindings.NewStorkContract(rpcUrl, contractAddr, account)
	if err != nil {
		return nil, err
	}
	return &ContractInteractor{
		logger:           logger,
		contract:         contract,
		pollingPeriodSec: pollingPeriodSec,
	}, nil
}

// unfortunately, Sui doesn't currently support websocket RPCs, so we can't listen to events from the contract
// the contract does emit events, so this can be implemented in the future if Sui re-adds websocket support
func (sci *ContractInteractor) ListenContractEvents(ctx context.Context, ch chan map[types.InternalEncodedAssetID]types.InternalTemporalNumericValue) {
	sci.logger.Warn().Msg("Sui does not currently support listening to events via websocket, falling back to polling")
}

func (sci *ContractInteractor) PullValues(encodedAssetIDs []types.InternalEncodedAssetID) (map[types.InternalEncodedAssetID]types.InternalTemporalNumericValue, error) {
	// convert to bindings EncodedAssetID
	bindingsEncodedAssetIDs := []bindings.EncodedAssetID{}
	for _, encodedAssetID := range encodedAssetIDs {
		bindingsEncodedAssetIDs = append(bindingsEncodedAssetIDs, bindings.EncodedAssetID(encodedAssetID))
	}
	values, err := sci.contract.GetMultipleTemporalNumericValuesUnchecked(bindingsEncodedAssetIDs)
	if err != nil {
		return nil, err
	}
	sci.logger.Debug().Msgf("successfully pulled %d values from contract", len(values))

	// convert to map[InternalEncodedAssetID]InternalStorkStructsTemporalNumericValue
	result := make(map[types.InternalEncodedAssetID]types.InternalTemporalNumericValue)
	for _, encodedAssetID := range encodedAssetIDs {
		if value, ok := values[bindings.EncodedAssetID(encodedAssetID)]; ok {
			result[encodedAssetID] = temporalNumericValueToInternal(value)
		}
	}
	return result, nil
}

func (sci *ContractInteractor) BatchPushToContract(priceUpdates map[types.InternalEncodedAssetID]types.AggregatedSignedPrice) error {
	var updateData []bindings.UpdateData
	for _, price := range priceUpdates {
		update, err := sci.aggregatedSignedPriceToUpdateData(price)
		if err != nil {
			return err
		}
		updateData = append(updateData, update)
	}
	digest, err := sci.contract.UpdateMultipleTemporalNumericValuesEvm(updateData)
	if err != nil {
		sci.logger.Error().Err(err).Msg("failed to update multiple temporal numeric values")
		return err
	}
	sci.logger.Info().
		Int("numUpdates", len(priceUpdates)).
		Str("txnDigest", digest).
		Msg("Successfully pushed batch update to contract")
	return nil
}

// todo: implement
func (sci *ContractInteractor) GetWalletBalance() (float64, error) {
	return -1, nil
}

func temporalNumericValueToInternal(value bindings.TemporalNumericValue) types.InternalTemporalNumericValue {
	magnitude := value.QuantizedValue.Magnitude
	negative := value.QuantizedValue.Negative
	signMultiplier := 1
	if negative {
		signMultiplier = -1
	}
	quantizedValue := new(big.Int).Mul(magnitude, big.NewInt(int64(signMultiplier)))

	return types.InternalTemporalNumericValue{
		TimestampNs:    value.TimestampNs,
		QuantizedValue: quantizedValue,
	}
}

func (sci *ContractInteractor) aggregatedSignedPriceToUpdateData(price types.AggregatedSignedPrice) (bindings.UpdateData, error) {
	signedPrice := price.StorkSignedPrice
	assetID, err := pusher.HexStringToByteArray(string(signedPrice.EncodedAssetID))
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

func loadPrivateKey(keyFileContent []byte) (*account.Account, error) {
	lines := strings.Split(string(keyFileContent), "\n")
	if len(lines) == 0 {
		return nil, fmt.Errorf("private key is empty")
	}
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
	privateKey = strings.TrimSpace(privateKey)

	if len(privateKey) == 0 {
		return nil, fmt.Errorf("private key is empty")
	}
	account, err := account.NewAccountWithKeystore(privateKey)
	if err != nil {
		return nil, err
	}
	return account, nil
}
