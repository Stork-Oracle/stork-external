package starknet

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/Stork-Oracle/stork-external/apps/chain_pusher/pkg/pusher"
	"github.com/Stork-Oracle/stork-external/apps/chain_pusher/pkg/starknet/bindings"
	"github.com/Stork-Oracle/stork-external/apps/chain_pusher/pkg/types"
	"github.com/Stork-Oracle/stork-external/shared"
	"github.com/rs/zerolog"
)

var (
	ErrInvalidPrivateKey = errors.New("failed to parse private key as a hex encoded field element")
	ErrEmptyPrivateKey   = errors.New("private key file is empty")
)

const (
	// eventChannelBufferSize bounds the buffer between the websocket subscription and the pusher.
	eventChannelBufferSize = 128

	// weiPerEth is used to render the account balance in whole tokens, both STRK and ETH being
	// 18 decimal tokens.
	weiPerEth = 1e18
)

// ContractInteractor implements types.ContractInteractor for the Stork Cairo contract.
type ContractInteractor struct {
	logger zerolog.Logger

	contractAddress string
	accountAddress  string
	feeTokenAddress string
	privateKey      *big.Int
	tip             string

	contract *bindings.StorkContract
}

// NewContractInteractor builds an interactor. Providers are attached later by ConnectHTTP and
// ConnectWs, matching how the pusher drives every other chain.
func NewContractInteractor(
	contractAddress string,
	accountAddress string,
	feeTokenAddress string,
	tip string,
	keyFileContent []byte,
	logger zerolog.Logger,
) (*ContractInteractor, error) {
	logger = logger.With().Str("component", "starknet-contract-interactor").Logger()

	privateKey, err := loadPrivateKey(keyFileContent)
	if err != nil {
		return nil, err
	}

	contract, err := bindings.NewStorkContract(contractAddress, accountAddress, privateKey, tip)
	if err != nil {
		return nil, fmt.Errorf("failed to create stork contract: %w", err)
	}

	return &ContractInteractor{
		logger:          logger,
		contractAddress: contractAddress,
		accountAddress:  accountAddress,
		feeTokenAddress: feeTokenAddress,
		privateKey:      privateKey,
		tip:             tip,
		contract:        contract,
	}, nil
}

func (sci *ContractInteractor) ConnectHTTP(ctx context.Context, url string) error {
	err := sci.contract.ConnectHTTP(ctx, url)
	if err != nil {
		return fmt.Errorf("failed to connect to HTTP RPC: %w", err)
	}

	if sci.contract.SpecVersionWarning != nil {
		sci.logger.Warn().
			Err(sci.contract.SpecVersionWarning).
			Msg("Node JSON-RPC spec version differs from the one this client implements, continuing anyway")
	}

	return nil
}

func (sci *ContractInteractor) ConnectWs(ctx context.Context, url string) error {
	if url == "" {
		sci.logger.Info().Msg("No websocket URL configured, falling back to polling")

		return nil
	}

	err := sci.contract.ConnectWs(ctx, url)
	if err != nil {
		return fmt.Errorf("failed to connect to websocket RPC: %w", err)
	}

	return nil
}

// ListenContractEvents streams ValueUpdate events into ch.
//
// Starknet exposes a native `starknet_subscribeEvents` websocket method, so unlike the Move and
// CosmWasm pushers this does not have to degrade to polling, provided a websocket URL was given.
func (sci *ContractInteractor) ListenContractEvents(
	ctx context.Context,
	ch chan map[types.InternalEncodedAssetID]types.InternalTemporalNumericValue,
) {
	eventCh := make(chan bindings.ValueUpdateEvent, eventChannelBufferSize)

	errCh, err := sci.contract.SubscribeValueUpdates(ctx, eventCh)
	if err != nil {
		sci.logger.Warn().Err(err).Msg("Failed to subscribe to contract events, falling back to polling")

		return
	}

	sci.logger.Info().Msg("Listening for contract events")

	for {
		select {
		case <-ctx.Done():
			return
		case subErr := <-errCh:
			if subErr != nil {
				sci.logger.Error().Err(subErr).Msg("Contract event subscription ended, falling back to polling")
			}

			return
		case event := <-eventCh:
			assetID := types.InternalEncodedAssetID(event.ID)
			ch <- map[types.InternalEncodedAssetID]types.InternalTemporalNumericValue{
				assetID: {
					TimestampNs:    event.Value.TimestampNs,
					QuantizedValue: event.Value.QuantizedValue,
				},
			}
		}
	}
}

func (sci *ContractInteractor) PullValues(
	ctx context.Context,
	encodedAssetIDs []types.InternalEncodedAssetID,
) (map[types.InternalEncodedAssetID]types.InternalTemporalNumericValue, error) {
	ids := make([]bindings.EncodedAssetID, 0, len(encodedAssetIDs))
	for _, encodedAssetID := range encodedAssetIDs {
		ids = append(ids, bindings.EncodedAssetID(encodedAssetID))
	}

	values, err := sci.contract.GetMultipleTemporalNumericValuesUnchecked(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("failed to pull values from contract: %w", err)
	}

	polledVals := make(map[types.InternalEncodedAssetID]types.InternalTemporalNumericValue, len(values))
	for id, value := range values {
		polledVals[types.InternalEncodedAssetID(id)] = types.InternalTemporalNumericValue{
			TimestampNs:    value.TimestampNs,
			QuantizedValue: value.QuantizedValue,
		}
	}

	sci.logger.Debug().Msgf("Pulled %d values from contract", len(polledVals))

	return polledVals, nil
}

func (sci *ContractInteractor) BatchPushToContract(
	ctx context.Context,
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

	txHash, err := sci.contract.UpdateTemporalNumericValuesV1(ctx, updateData)
	if err != nil {
		return fmt.Errorf("failed to update temporal numeric values: %w", err)
	}

	sci.logger.Debug().
		Int("numUpdates", len(priceUpdates)).
		Str("txHash", txHash).
		Msgf(
			"Successfully pushed batch update of %d asset%s to contract",
			len(priceUpdates), pusher.Pluralize(len(priceUpdates)),
		)

	return nil
}

// GetWalletBalance returns the pushing account's balance in whole fee tokens.
func (sci *ContractInteractor) GetWalletBalance(ctx context.Context) (float64, error) {
	if sci.feeTokenAddress == "" {
		return -1, nil
	}

	balance, err := sci.contract.GetBalance(ctx, sci.feeTokenAddress)
	if err != nil {
		return -1, fmt.Errorf("failed to get wallet balance: %w", err)
	}

	asFloat, _ := new(big.Float).Quo(new(big.Float).SetInt(balance), big.NewFloat(weiPerEth)).Float64()

	return asFloat, nil
}

// aggregatedSignedPriceToUpdateData converts a Stork websocket price into contract calldata.
func aggregatedSignedPriceToUpdateData(
	price types.AggregatedSignedPrice,
) (bindings.UpdateData, error) {
	signedPrice := price.StorkSignedPrice

	assetID, err := pusher.HexStringToByte32(string(signedPrice.EncodedAssetID))
	if err != nil {
		return bindings.UpdateData{}, fmt.Errorf("failed to convert encoded asset id to byte array: %w", err)
	}

	//nolint:mnd // base number.
	quantizedValue, ok := new(big.Int).SetString(string(signedPrice.QuantizedPrice), 10)
	if !ok {
		return bindings.UpdateData{}, shared.ErrFailedToConvertQuantizedPriceToBigInt
	}

	publisherMerkleRoot, err := pusher.HexStringToByte32(signedPrice.PublisherMerkleRoot)
	if err != nil {
		return bindings.UpdateData{}, fmt.Errorf("failed to convert publisher merkle root to byte array: %w", err)
	}

	valueComputeAlgHash, err := pusher.HexStringToByte32(signedPrice.StorkCalculationAlg.Checksum)
	if err != nil {
		return bindings.UpdateData{}, fmt.Errorf("failed to convert value compute alg hash to byte array: %w", err)
	}

	sigR, err := pusher.HexStringToByte32(signedPrice.TimestampedSignature.Signature.R)
	if err != nil {
		return bindings.UpdateData{}, fmt.Errorf("failed to convert R to byte array: %w", err)
	}

	sigS, err := pusher.HexStringToByte32(signedPrice.TimestampedSignature.Signature.S)
	if err != nil {
		return bindings.UpdateData{}, fmt.Errorf("failed to convert S to byte array: %w", err)
	}

	vBytes, err := pusher.HexStringToByteArray(signedPrice.TimestampedSignature.Signature.V)
	if err != nil {
		return bindings.UpdateData{}, fmt.Errorf("failed to convert V to byte array: %w", err)
	}

	if len(vBytes) == 0 {
		return bindings.UpdateData{}, bindings.ErrInvalidSignatureV
	}

	return bindings.UpdateData{
		ID: assetID,
		TemporalNumericValue: bindings.TemporalNumericValue{
			TimestampNs:    signedPrice.TimestampedSignature.TimestampNano,
			QuantizedValue: quantizedValue,
		},
		PublisherMerkleRoot: publisherMerkleRoot,
		ValueComputeAlgHash: valueComputeAlgHash,
		R:                   sigR,
		S:                   sigS,
		V:                   vBytes[0],
	}, nil
}

// loadPrivateKey reads a hex encoded Stark private key from the key file.
func loadPrivateKey(keyFileContent []byte) (*big.Int, error) {
	trimmed := strings.TrimSpace(strings.Split(string(keyFileContent), "\n")[0])
	if trimmed == "" {
		return nil, ErrEmptyPrivateKey
	}

	trimmed = strings.TrimPrefix(trimmed, "0x")

	//nolint:mnd // base number.
	privateKey, ok := new(big.Int).SetString(trimmed, 16)
	if !ok {
		return nil, ErrInvalidPrivateKey
	}

	return privateKey, nil
}
