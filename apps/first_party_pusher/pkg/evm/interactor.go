package first_party_evm

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rs/zerolog"

	"github.com/Stork-Oracle/stork-external/apps/chain_pusher/pkg/pusher"
	chain_pusher_types "github.com/Stork-Oracle/stork-external/apps/chain_pusher/pkg/types"
	"github.com/Stork-Oracle/stork-external/apps/first_party_pusher/pkg/evm/bindings"
	"github.com/Stork-Oracle/stork-external/apps/first_party_pusher/pkg/types"
	publisher_agent "github.com/Stork-Oracle/stork-external/apps/publisher_agent/pkg"
	"github.com/Stork-Oracle/stork-external/shared"
)

const (
	maxRetryAttempts         = 5
	initialBackoff           = 1 * time.Second
	exponentialBackoffFactor = 1.5
)

var (
	ErrEventChannelClosed      = errors.New("event channel is closed")
	ErrMaxRetryAttemptsReached = errors.New("max retry attempts reached")
)

type ContractInteractor struct {
	logger zerolog.Logger

	contract   *bindings.FirstPartyStorkContract
	wsContract *bindings.FirstPartyStorkContract
	client     *ethclient.Client
	wsClient   *ethclient.Client

	privateKey *ecdsa.PrivateKey
	chainID    *big.Int
	gasLimit   uint64
}

func NewContractInteractor(
	rpcUrl string,
	wsUrl string,
	contractAddr string,
	privateKey *ecdsa.PrivateKey,
	gasLimit uint64,
	logger zerolog.Logger,
) (*ContractInteractor, error) {
	client, err := ethclient.Dial(rpcUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RPC: %w", err)
	}

	var wsClient *ethclient.Client
	if wsUrl != "" {
		wsClient, err = ethclient.Dial(wsUrl)
		if err != nil {
			logger.Warn().Err(err).Msg("Failed to connect to WebSocket, using HTTP only")
		} else {
			logger.Info().Msg("Connected to WebSocket endpoint")
		}
	}

	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get chain ID: %w", err)
	}

	contractAddress := common.HexToAddress(contractAddr)

	contract, err := bindings.NewFirstPartyStorkContract(contractAddress, client)
	if err != nil {
		return nil, fmt.Errorf("failed to create contract instance: %w", err)
	}

	var wsContract *bindings.FirstPartyStorkContract
	if wsClient != nil {
		wsContract, err = bindings.NewFirstPartyStorkContract(contractAddress, wsClient)
		if err != nil {
			logger.Warn().Err(err).Msg("Failed to create WebSocket contract instance")
		}
	}

	return &ContractInteractor{
		logger: logger.With().Str("component", "contract_interactor").Logger(),

		contract:   contract,
		wsContract: wsContract,
		client:     client,
		wsClient:   wsClient,
		privateKey: privateKey,
		chainID:    chainID,
		gasLimit:   gasLimit,
	}, nil
}

func (ci *ContractInteractor) CheckPublisherUser(
	pubKey common.Address,
) (bool, error) {
	publisherUser, err := ci.contract.GetPublisherUser(nil, pubKey)
	if err != nil {
		return false, err
	}

	return publisherUser.PubKey == pubKey, nil
}

func (ci *ContractInteractor) PullValues(
	pubKeyAssetIDPairs map[common.Address][]shared.AssetID,
	assetIDtoEncodedAssetID map[shared.AssetID]shared.EncodedAssetID,
) ([]types.ContractUpdate, error) {
	polledVals := make([]types.ContractUpdate, 0)

	for pubKey, assetIDs := range pubKeyAssetIDPairs {
		contractUpdate := types.ContractUpdate{
			Pubkey:           pubKey,
			ContractValueMap: make(map[shared.EncodedAssetID]chain_pusher_types.InternalTemporalNumericValue),
		}
		for _, assetID := range assetIDs {
			storkStructsTemporalNumericValue, err := ci.contract.GetLatestTemporalNumericValue(
				nil,
				pubKey,
				string(assetID),
			)
			if err != nil {
				if strings.Contains(err.Error(), "NotFound()") {
					ci.logger.Warn().Err(err).Str("asset_id", string(assetID)).Msg("No value found")
				} else {
					ci.logger.Warn().Err(err).Str("asset_id", string(assetID)).Msg("Failed to get temporal numeric value")
				}

				continue
			}

			encodedAssetID, exists := assetIDtoEncodedAssetID[shared.AssetID(assetID)]
			if !exists {
				ci.logger.Error().Str("asset_id", string(assetID)).Msg("Asset not found in assetIDtoEncodedAssetID map")

				continue
			}

			contractUpdate.ContractValueMap[encodedAssetID] = chain_pusher_types.InternalTemporalNumericValue(
				storkStructsTemporalNumericValue,
			)
		}

		polledVals = append(polledVals, contractUpdate)
	}

	return polledVals, nil
}

func (ci *ContractInteractor) ListenContractEvents(
	ctx context.Context,
	ch chan types.ContractUpdate,
	pubKeyAssetIDPairs map[common.Address][]shared.AssetID,
) {
	if ci.wsContract == nil {
		ci.logger.Warn().Msg("WebSocket contract not available, cannot listen for events")

		return
	}

	watchOpts := &bind.WatchOpts{Context: context.Background()}

	sub, eventCh, err := ci.setupSubscription(watchOpts, pubKeyAssetIDPairs)
	if err != nil {
		ci.logger.Error().Err(err).Msg("Failed to establish initial subscription")

		return
	}

	defer func() {
		ci.logger.Debug().Msg("Exiting ListenContractEvents")

		if sub != nil {
			sub.Unsubscribe()
			close(eventCh)
		}
	}()

	ci.logger.Info().Msg("Listening for contract events via WebSocket")

	for {
		err = ci.listenLoop(ctx, sub, eventCh, ch)
		if ctx.Err() != nil {
			return
		}

		ci.logger.Warn().Err(err).Msg("Error while watching contract events")

		if sub != nil {
			sub.Unsubscribe()
			sub = nil
		}

		sub, eventCh, err = ci.reconnect(ctx, watchOpts, pubKeyAssetIDPairs)
		if err != nil {
			return
		}
	}
}

func (ci *ContractInteractor) BatchPushToContract(
	updatesByEntry map[chain_pusher_types.AssetEntry]publisher_agent.SignedPriceUpdate[*shared.EvmSignature],
) error {
	updates, historic, err := ci.getUpdatePayload(updatesByEntry)
	if err != nil {
		return fmt.Errorf("failed to convert signed price update: %w", err)
	}

	auth, err := bind.NewKeyedTransactorWithChainID(ci.privateKey, ci.chainID)
	if err != nil {
		return fmt.Errorf("failed to create transactor: %w", err)
	}

	// let the library auto-estimate the gas price
	auth.GasLimit = ci.gasLimit
	auth.Value = big.NewInt(0)

	tx, err := ci.contract.UpdateTemporalNumericValues(auth, updates, historic)
	if err != nil {
		return fmt.Errorf("failed to call UpdateTemporalNumericValues: %w", err)
	}

	ci.logger.Info().
		Int("num_updates", len(updates)).
		Str("tx_hash", tx.Hash().Hex()).
		Msg("Successfully submitted batch signed price update transaction")

	return nil
}

func (ci *ContractInteractor) Close() {
	if ci.client != nil {
		ci.client.Close()
	}

	if ci.wsClient != nil {
		ci.wsClient.Close()
	}
}

func (ci *ContractInteractor) getUpdatePayload(
	updatesByEntry map[chain_pusher_types.AssetEntry]publisher_agent.SignedPriceUpdate[*shared.EvmSignature],
) ([]bindings.FirstPartyStorkStructsPublisherTemporalNumericValueInput, []bool, error) {
	updates := make([]bindings.FirstPartyStorkStructsPublisherTemporalNumericValueInput, 0, len(updatesByEntry))
	historic := make([]bool, 0, len(updatesByEntry))

	for entry, signedPriceUpdate := range updatesByEntry {
		ci.logger.Info().
			Str("asset", string(signedPriceUpdate.AssetID)).
			Str("price", string(signedPriceUpdate.SignedPrice.QuantizedPrice)).
			Str("encoded_asset_id", string(entry.EncodedAssetID)).
			Msg("Pushing signed price update to first party contract")

		// Convert quantized price to big.Int
		quantizedValue, success := new(big.Int).SetString(string(signedPriceUpdate.SignedPrice.QuantizedPrice), 10)
		if !success {
			return nil, nil,
				fmt.Errorf(
					"%w: %s",
					shared.ErrFailedToConvertQuantizedPriceToBigInt,
					signedPriceUpdate.SignedPrice.QuantizedPrice,
				)
		}

		// Create the temporal numeric value using the signed data timestamp
		temporalValue := bindings.FirstPartyStorkStructsTemporalNumericValue{
			TimestampNs:    signedPriceUpdate.SignedPrice.TimestampedSignature.TimestampNano,
			QuantizedValue: quantizedValue,
		}

		// Parse the publisher key
		pubKeyBytes, err := pusher.HexStringToByte20(string(signedPriceUpdate.SignedPrice.PublisherKey))
		if err != nil {
			return nil, nil, fmt.Errorf("failed to decode publisher key: %w", err)
		}

		// Parse the signature components
		rBytes, err := pusher.HexStringToByte32(signedPriceUpdate.SignedPrice.TimestampedSignature.Signature.R)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to decode signature R: %w", err)
		}

		sBytes, err := pusher.HexStringToByte32(signedPriceUpdate.SignedPrice.TimestampedSignature.Signature.S)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to decode signature S: %w", err)
		}

		vInt, err := strconv.ParseInt(signedPriceUpdate.SignedPrice.TimestampedSignature.Signature.V[2:], 16, 8)
		if err != nil || vInt < 0 || vInt > 255 {
			return nil, nil, fmt.Errorf("failed to parse signature V: %w", err)
		}

		updates = append(updates, bindings.FirstPartyStorkStructsPublisherTemporalNumericValueInput{
			TemporalNumericValue: temporalValue,
			PubKey:               pubKeyBytes,
			AssetPairId:          string(entry.AssetID),
			R:                    rBytes,
			S:                    sBytes,
			V:                    uint8(vInt),
		})
		historic = append(historic, entry.Historic)
	}

	return updates, historic, nil
}

func (ci *ContractInteractor) setupSubscription(
	watchOpts *bind.WatchOpts,
	pubKeyAssetIDPairs map[common.Address][]shared.AssetID,
) (ethereum.Subscription, chan *bindings.FirstPartyStorkContractValueUpdate, error) {
	eventCh := make(chan *bindings.FirstPartyStorkContractValueUpdate)
	pubKeys := make([]common.Address, 0, len(pubKeyAssetIDPairs))
	assetIDs := make([]string, 0, len(pubKeyAssetIDPairs))

	for pubKey, assets := range pubKeyAssetIDPairs {
		pubKeys = append(pubKeys, pubKey)
		for _, asset := range assets {
			assetIDs = append(assetIDs, string(asset))
		}
	}

	sub, err := ci.wsContract.WatchValueUpdate(watchOpts, eventCh, pubKeys, assetIDs)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to watch contract events: %w", err)
	}

	return sub, eventCh, nil
}

func (ci *ContractInteractor) listenLoop(
	ctx context.Context,
	sub ethereum.Subscription,
	eventCh chan *bindings.FirstPartyStorkContractValueUpdate,
	outCh chan types.ContractUpdate,
) error {
	for {
		select {
		case <-ctx.Done():
			return nil

		case err := <-sub.Err():
			return fmt.Errorf("error from subscription: %w", err)

		case vLog, ok := <-eventCh:
			if !ok {
				ci.logger.Warn().Msg("Event channel closed, exiting event listener")

				return ErrEventChannelClosed
			}

			tv := chain_pusher_types.InternalTemporalNumericValue{
				QuantizedValue: vLog.QuantizedValue,
				TimestampNs:    vLog.TimestampNs,
			}

			update := types.ContractUpdate{
				Pubkey: vLog.PubKey,
				ContractValueMap: map[shared.EncodedAssetID]chain_pusher_types.InternalTemporalNumericValue{
					shared.EncodedAssetID(vLog.AssetId.Hex()): tv,
				},
			}
			select {
			case outCh <- update:
			case <-ctx.Done():
				return fmt.Errorf("context done: %w", ctx.Err())
			}
		}
	}
}

func (ci *ContractInteractor) reconnect(
	ctx context.Context,
	watchOpts *bind.WatchOpts,
	pubKeyAssetIDPairs map[common.Address][]shared.AssetID,
) (ethereum.Subscription, chan *bindings.FirstPartyStorkContractValueUpdate, error) {
	backoff := initialBackoff
	for retryCount := range maxRetryAttempts {
		backoff = time.Duration(float64(backoff) * exponentialBackoffFactor)
		ci.logger.Info().Dur("backoff", backoff).
			Int("attempt", retryCount+1).
			Int("maxAttempts", maxRetryAttempts).
			Msg("Attempting to reconnect to contract events")

		select {
		case <-ctx.Done():
			return nil, nil, fmt.Errorf("context done: %w", ctx.Err())
		case <-time.After(backoff):
			newSub, newEventCh, err := ci.setupSubscription(watchOpts, pubKeyAssetIDPairs)
			if err != nil {
				ci.logger.Warn().Err(err).Msg("Failed to reconnect to contract events")

				continue
			}

			ci.logger.Info().Msg("Successfully reconnected to contract events")

			return newSub, newEventCh, nil
		}
	}

	ci.logger.Error().Int("maxRetryAttempts", maxRetryAttempts).
		Msg("Max retry attempts reached, giving up on reconnection")

	return nil, nil, ErrMaxRetryAttemptsReached
}
