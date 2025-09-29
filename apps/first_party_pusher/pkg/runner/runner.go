package runner

import (
	"context"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/rs/zerolog"

	"fmt"

	chain_pusher_types "github.com/Stork-Oracle/stork-external/apps/chain_pusher/pkg/types"
	"github.com/Stork-Oracle/stork-external/apps/first_party_pusher/pkg/types"
	publisher_agent "github.com/Stork-Oracle/stork-external/apps/publisher_agent/pkg"
	"github.com/Stork-Oracle/stork-external/shared"
)

type FirstPartyRunner[T shared.Signature] struct {
	config             *types.FirstPartyConfig
	contractInteractor types.ContractInteractor[T]
	websocketServer    *WebsocketServer[T]

	batchingWindowSecs int
	pollingPeriodSecs  int

	cancel context.CancelFunc
	logger zerolog.Logger
}

func NewFirstPartyRunner[T shared.Signature](
	config *types.FirstPartyConfig,
	contractInteractor types.ContractInteractor[T],
	batchingWindowSecs int,
	pollingPeriodSecs int,
	cancel context.CancelFunc,
	logger zerolog.Logger,
) *FirstPartyRunner[T] {
	return &FirstPartyRunner[T]{
		config:             config,
		contractInteractor: contractInteractor,
		websocketServer:    nil,
		batchingWindowSecs: batchingWindowSecs,
		pollingPeriodSecs:  pollingPeriodSecs,
		cancel:             cancel,
		logger:             logger.With().Str("component", "first_party_runner").Logger(),
	}
}

func (r *FirstPartyRunner[T]) Run(ctx context.Context) {
	r.logger.Info().Msg("Starting EVM First Party Chain Pusher")

	signedPriceUpdateCh := make(chan publisher_agent.SignedPriceUpdate[T], 1000)
	contractUpdateCh := make(chan types.ContractUpdate)

	r.websocketServer = NewWebsocketServer(r.config.WebsocketPort, signedPriceUpdateCh)

	go func() {
		err := r.websocketServer.Start()
		if err != nil {
			r.logger.Fatal().Err(err).Msg("WebSocket server failed")
		}
	}()

	latestPublisherValueMap, latestContractValueMap, pubKeyAssetIDPairs, assetIDtoEncodedAssetID := r.initialize(ctx)

	go r.poll(ctx, contractUpdateCh, pubKeyAssetIDPairs)
	go r.contractInteractor.ListenContractEvents(ctx, contractUpdateCh, pubKeyAssetIDPairs) // todo: doesn't handle errors

	batchingTicker := time.NewTicker(time.Duration(r.batchingWindowSecs) * time.Second)
	defer batchingTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			r.logger.Info().Msg("First Party Runner stopped")

			return

		case signedPriceUpdate := <-signedPriceUpdateCh:
			pubKey := common.HexToAddress(string(signedPriceUpdate.SignedPrice.PublisherKey))
			if _, exists := latestPublisherValueMap[pubKey]; !exists {
				r.logger.Error().Str("pubkey", pubKey.Hex()).
					Msg("Pubkey not found in latest publisher value map")

				continue
			}

			encodedAssetID, exists := assetIDtoEncodedAssetID[signedPriceUpdate.AssetID]
			if !exists {
				r.logger.Error().Str("asset", string(signedPriceUpdate.AssetID)).
					Msg("Asset not found in assetIDtoEncodedAssetID map")

				continue
			}

			latestPublisherValueMap[pubKey][encodedAssetID] = signedPriceUpdate

		case contractUpdate := <-contractUpdateCh:
			for assetID, value := range contractUpdate.LatestContractValueMap {
				if _, exists := latestContractValueMap[contractUpdate.Pubkey]; !exists {
					r.logger.Error().Str("pubkey", contractUpdate.Pubkey.Hex()).
						Msg("Pubkey not found in latest contract value map")

					continue
				}

				encodedAssetID, exists := assetIDtoEncodedAssetID[shared.AssetID(assetID)]
				if !exists {
					r.logger.Error().Str("asset", assetID).
						Msg("Asset not found in assetIDtoEncodedAssetID map")

					continue
				}

				latestContractValueMap[contractUpdate.Pubkey][encodedAssetID] = value
			}

		case <-batchingTicker.C:
			r.handleBatch(ctx, latestPublisherValueMap, latestContractValueMap)
		}
	}
}

func (r *FirstPartyRunner[T]) Stop() {
	r.logger.Info().Msg("Stopping EVM First Party Chain Pusher")
	r.cancel()
	r.contractInteractor.Close()

	if r.websocketServer != nil {
		_ = r.websocketServer.Stop()
	}
}

func (r *FirstPartyRunner[T]) initialize(ctx context.Context) (
	map[common.Address]map[shared.EncodedAssetID]publisher_agent.SignedPriceUpdate[T],
	map[common.Address]map[shared.EncodedAssetID]chain_pusher_types.InternalTemporalNumericValue,
	map[common.Address][]string,
	map[shared.AssetID]shared.EncodedAssetID,
) {
	latestContractValueMap := make(map[common.Address]map[shared.EncodedAssetID]chain_pusher_types.InternalTemporalNumericValue)
	latestPublisherValueMap := make(map[common.Address]map[shared.EncodedAssetID]publisher_agent.SignedPriceUpdate[T])
	pubKeyAssetIDPairs := make(map[common.Address][]string, len(r.config.AssetConfig.Assets))
	assetIDtoEncodedAssetID := make(map[shared.AssetID]shared.EncodedAssetID, len(r.config.AssetConfig.Assets))

	for assetID, assetEntry := range r.config.AssetConfig.Assets {
		if assetEntry.PublicKey == "" {
			r.logger.Error().Str("asset", string(assetID)).Msg("Asset has no specific pub key configured")

			continue
		}

		pubKey := common.HexToAddress(string(assetEntry.PublicKey))
		if _, exists := pubKeyAssetIDPairs[pubKey]; !exists {
			pubKeyAssetIDPairs[pubKey] = make([]string, 0)
		}

		latestPublisherValueMap[pubKey] = make(map[shared.EncodedAssetID]publisher_agent.SignedPriceUpdate[T])
		latestContractValueMap[pubKey] = make(map[shared.EncodedAssetID]chain_pusher_types.InternalTemporalNumericValue)
		pubKeyAssetIDPairs[pubKey] = append(pubKeyAssetIDPairs[pubKey], string(assetID))
		hash := crypto.Keccak256Hash([]byte(assetID))
		assetIDtoEncodedAssetID[assetID] = shared.EncodedAssetID(hash.Hex())
	}

	contractUpdates, err := r.contractInteractor.PullValues(ctx, pubKeyAssetIDPairs) // todo: what about cold start?
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to pull values from contract")
	}

	for _, update := range contractUpdates {
		for assetID, value := range update.LatestContractValueMap {
			if _, exists := assetIDtoEncodedAssetID[shared.AssetID(assetID)]; !exists {
				r.logger.Error().Str("asset", assetID).
					Msg("Asset not found in assetIDtoEncodedAssetID map")

				continue
			}

			encodedAssetID, exists := assetIDtoEncodedAssetID[shared.AssetID(assetID)]
			if !exists {
				r.logger.Error().Str("asset", assetID).
					Msg("Asset not found in assetIDtoEncodedAssetID map")

				continue
			}

			latestContractValueMap[update.Pubkey][encodedAssetID] = value
		}
	}

	return latestPublisherValueMap, latestContractValueMap, pubKeyAssetIDPairs, assetIDtoEncodedAssetID
}

func (r *FirstPartyRunner[T]) handleBatch(
	ctx context.Context,
	latestPublisherValueMap map[common.Address]map[shared.EncodedAssetID]publisher_agent.SignedPriceUpdate[T],
	latestContractValueMap map[common.Address]map[shared.EncodedAssetID]chain_pusher_types.InternalTemporalNumericValue,
) {
	r.logger.Debug().
		Int("num_publisher_updates", len(latestPublisherValueMap)).
		Int("num_contract_updates", len(latestContractValueMap)).
		Msg("Handling batch")

	updates := make(map[chain_pusher_types.AssetEntry]publisher_agent.SignedPriceUpdate[T])

	for pubKey, signedPriceUpdateMap := range latestPublisherValueMap {
		for encodedAssetID, signedPriceUpdate := range signedPriceUpdateMap {
			assetEntry, exists := r.config.AssetConfig.Assets[signedPriceUpdate.AssetID]
			if !exists {
				r.logger.Error().Str("asset", string(signedPriceUpdate.AssetID)).
					Msg("Asset not found in asset config")

				continue
			}

			tnvMap, exists := latestContractValueMap[pubKey]
			if !exists {
				r.logger.Error().Str("pubkey", pubKey.Hex()).
					Msg("Pubkey not found in latest contract value map")

				continue
			}

			latestContractValue, exists := tnvMap[encodedAssetID]
			if !exists {
				r.logger.Info().
					Str("asset", string(signedPriceUpdate.AssetID)).
					Msg("Triggering push due to first price update")

				updates[assetEntry] = signedPriceUpdate

				continue
			}

			if r.shouldPushBasedOnFallback(assetEntry, signedPriceUpdate, latestContractValue) {
				r.logger.Info().
					Str("asset", string(signedPriceUpdate.AssetID)).
					Msg("Triggering push due to fallback period")

				updates[assetEntry] = signedPriceUpdate

				continue
			}

			if r.shouldPushBasedOnDelta(assetEntry, signedPriceUpdate, latestContractValue) {
				r.logger.Info().
					Str("asset", string(signedPriceUpdate.AssetID)).
					Msg("Triggering push due to price delta threshold")

				updates[assetEntry] = signedPriceUpdate
			}
		}
	}
	r.logger.Debug().
		Int("num_updates", len(updates)).
		Msg("Updates to push")

	if len(updates) > 0 {
		go r.pushBatch(ctx, updates, latestContractValueMap)
	}
}

func (r *FirstPartyRunner[T]) shouldPushBasedOnFallback(
	assetEntry chain_pusher_types.AssetEntry,
	signedPriceUpdate publisher_agent.SignedPriceUpdate[T],
	latestContractValue chain_pusher_types.InternalTemporalNumericValue,
) bool {
	// todo: this won't push if data stops flowing, is that what we want?
	lastTime := time.Unix(0, int64(signedPriceUpdate.SignedPrice.TimestampedSignature.TimestampNano))
	if lastTime.After(time.Unix(0, int64(latestContractValue.TimestampNs)).Add(time.Duration(assetEntry.FallbackPeriodSecs) * time.Second)) {
		return true
	}

	return false
}

func (r *FirstPartyRunner[T]) shouldPushBasedOnDelta(
	assetEntry chain_pusher_types.AssetEntry,
	signedPriceUpdate publisher_agent.SignedPriceUpdate[T],
	latestContractValue chain_pusher_types.InternalTemporalNumericValue,
) bool {
	newPrice, err := r.convertQuantizedPriceToBigFloat(string(signedPriceUpdate.SignedPrice.QuantizedPrice))
	if err != nil {
		r.logger.Error().
			Err(err).
			Msg("Failed to convert quantized price to big.Float")

		return false
	}

	contractPrice := big.NewFloat(0).SetInt(latestContractValue.QuantizedValue)

	// Calculate percentage change
	diff := new(big.Float).Sub(newPrice, contractPrice)
	percentChange := new(big.Float).Quo(diff, contractPrice)
	percentChange.Mul(percentChange, big.NewFloat(100))

	absPercentChange := new(big.Float).Abs(percentChange)
	threshold := big.NewFloat(assetEntry.PercentChangeThreshold)

	return absPercentChange.Cmp(threshold) >= 0
}

func (r *FirstPartyRunner[T]) pushBatch(
	ctx context.Context,
	updates map[chain_pusher_types.AssetEntry]publisher_agent.SignedPriceUpdate[T],
	latestContractValueMap map[common.Address]map[shared.EncodedAssetID]chain_pusher_types.InternalTemporalNumericValue,
) {
	r.logger.Debug().
		Int("num_updates", len(updates)).
		Msg("Pushing batch to contract")

	pushCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	err := r.contractInteractor.BatchPushToContract(pushCtx, updates)
	if err != nil {
		r.logger.Error().
			Err(err).
			Msg("Failed to push batch to contract")
	}

	r.logger.Debug().
		Int("num_updates", len(updates)).
		Msg("Updated contract values")

	for entry, update := range updates {
		quantizedValInt := new(big.Int)
		//nolint:mnd // Base number
		quantizedValInt.SetString(string(update.SignedPrice.QuantizedPrice), 10)
		pubKey := common.HexToAddress(string(entry.PublicKey))

		if _, exists := latestContractValueMap[pubKey]; !exists {
			latestContractValueMap[pubKey] = make(map[shared.EncodedAssetID]chain_pusher_types.InternalTemporalNumericValue)
		}

		latestContractValueMap[pubKey][entry.EncodedAssetID] = chain_pusher_types.InternalTemporalNumericValue{
			TimestampNs:    update.SignedPrice.TimestampedSignature.TimestampNano,
			QuantizedValue: quantizedValInt,
		}
	}
}

func (r *FirstPartyRunner[T]) poll(
	ctx context.Context,
	ch chan types.ContractUpdate,
	pubKeyAssetIDPairs map[common.Address][]string,
) {
	r.logger.Debug().Msg("Polling contract for new values")

	pollingTicker := time.NewTicker(time.Duration(r.pollingPeriodSecs) * time.Second)
	defer pollingTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-pollingTicker.C:
			latestContractUpdates, err := r.contractInteractor.PullValues(ctx, pubKeyAssetIDPairs)
			if err != nil {
				r.logger.Error().Err(err).Msg("Failed to pull values from contract")
			}

			for _, update := range latestContractUpdates {
				ch <- update
			}
		}
	}
}

func (r *FirstPartyRunner[T]) convertQuantizedPriceToBigFloat(quantizedPrice string) (*big.Float, error) {
	bf, success := new(big.Float).SetString(quantizedPrice)
	if !success {
		return nil, fmt.Errorf("failed to convert quantized price to big.Float: %s", quantizedPrice)
	}

	return bf, nil
}
