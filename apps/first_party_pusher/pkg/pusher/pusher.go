package pusher

import (
	"context"
	"math/big"
	"time"

	chain_pusher_types "github.com/Stork-Oracle/stork-external/apps/chain_pusher/pkg/types"
	"github.com/Stork-Oracle/stork-external/apps/first_party_pusher/pkg/types"
	publisher_agent "github.com/Stork-Oracle/stork-external/apps/publisher_agent/pkg"
	"github.com/Stork-Oracle/stork-external/shared"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rs/zerolog"
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

	latestPublisherValueMap, latestContractValueMap, pubKeyAssetIDPairs := r.initialize()

	go r.poll(ctx, contractUpdateCh, pubKeyAssetIDPairs)
	go r.contractInteractor.ListenContractEvents(ctx, contractUpdateCh, pubKeyAssetIDPairs)

	batchingTicker := time.NewTicker(time.Duration(r.batchingWindowSecs) * time.Second)
	defer batchingTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			r.logger.Info().Msg("First Party Runner stopped")

			return

		case signedPriceUpdate := <-signedPriceUpdateCh:
			pubKey := common.HexToAddress(string(signedPriceUpdate.SignedPrice.PublisherKey))

			publisherAssetPair := types.PublisherAssetPair{
				Address: pubKey,
				AssetID: signedPriceUpdate.AssetID,
			}

			latestPublisherValueMap[publisherAssetPair] = signedPriceUpdate

		case contractUpdate := <-contractUpdateCh:
			for assetID, value := range contractUpdate.ContractValueMap {
				publisherAssetPair := types.PublisherAssetPair{
					Address: contractUpdate.Pubkey,
					AssetID: assetID,
				}

				latestContractValueMap[publisherAssetPair] = value
			}

		case <-batchingTicker.C:
			r.handleBatch(latestPublisherValueMap, latestContractValueMap)
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

func (r *FirstPartyRunner[T]) initialize() (
	map[types.PublisherAssetPair]publisher_agent.SignedPriceUpdate[T],
	map[types.PublisherAssetPair]chain_pusher_types.InternalTemporalNumericValue,
	map[common.Address][]shared.AssetID,
) {
	latestPublisherValueMap := make(map[types.PublisherAssetPair]publisher_agent.SignedPriceUpdate[T])
	latestContractValueMap := make(map[types.PublisherAssetPair]chain_pusher_types.InternalTemporalNumericValue)
	pubKeyAssetIDPairs := make(map[common.Address][]shared.AssetID, len(r.config.AssetConfig.Assets))

	// populate pubKeyAssetIDPairs and assetIDtoEncodedAssetID
	for assetID, assetEntry := range r.config.AssetConfig.Assets {
		if assetEntry.PublicKey == "" {
			r.logger.Error().Str("asset", string(assetID)).Msg("Asset has no specific pub key configured")

			continue
		}

		pubKey := common.HexToAddress(string(assetEntry.PublicKey))
		if _, exists := pubKeyAssetIDPairs[pubKey]; !exists {
			pubKeyAssetIDPairs[pubKey] = make([]shared.AssetID, 0)
		}

		pubKeyAssetIDPairs[pubKey] = append(pubKeyAssetIDPairs[pubKey], assetID)
	}

	contractUpdates, err := r.contractInteractor.PullValues(pubKeyAssetIDPairs)
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to pull values from contract - expected if cold starting")
	}

	// populate latestContractValueMap
	for _, update := range contractUpdates {
		for assetID, value := range update.ContractValueMap {
			publisherAssetPair := types.PublisherAssetPair{
				Address: update.Pubkey,
				AssetID: assetID,
			}

			latestContractValueMap[publisherAssetPair] = value
		}
	}

	for pubKey := range pubKeyAssetIDPairs {
		isPublisherUser, _ := r.contractInteractor.CheckPublisherUser(pubKey)

		if !isPublisherUser {
			r.logger.Fatal().Str("pubkey", pubKey.Hex()).Msg("Publisher user not found")
		}
	}

	return latestPublisherValueMap, latestContractValueMap, pubKeyAssetIDPairs
}

func (r *FirstPartyRunner[T]) handleBatch(
	latestPublisherValueMap map[types.PublisherAssetPair]publisher_agent.SignedPriceUpdate[T],
	latestContractValueMap map[types.PublisherAssetPair]chain_pusher_types.InternalTemporalNumericValue,
) {
	r.logger.Debug().
		Int("num_publisher_updates", len(latestPublisherValueMap)).
		Int("num_contract_updates", len(latestContractValueMap)).
		Msg("Handling batch")

	updates := make(map[types.AssetEntry]publisher_agent.SignedPriceUpdate[T])

	for publisherAssetPair, signedPriceUpdate := range latestPublisherValueMap {
		assetEntry, exists := r.config.AssetConfig.Assets[signedPriceUpdate.AssetID]
		if !exists {
			r.logger.Error().Str("asset", string(signedPriceUpdate.AssetID)).
				Msg("Asset not found in asset config")

			continue
		}

		latestContractValue, exists := latestContractValueMap[publisherAssetPair]
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

	r.logger.Debug().
		Int("num_updates", len(updates)).
		Msg("Updates to push")

	if len(updates) > 0 {
		r.pushBatch(updates, latestContractValueMap)
	}
}

func (r *FirstPartyRunner[T]) shouldPushBasedOnFallback(
	assetEntry types.AssetEntry,
	signedPriceUpdate publisher_agent.SignedPriceUpdate[T],
	latestContractValue chain_pusher_types.InternalTemporalNumericValue,
) bool {
	publisherTimestampNs := signedPriceUpdate.SignedPrice.TimestampedSignature.TimestampNano

	return publisherTimestampNs-latestContractValue.TimestampNs > assetEntry.FallbackPeriodSecs*uint64(time.Second)
}

func (r *FirstPartyRunner[T]) shouldPushBasedOnDelta(
	assetEntry types.AssetEntry,
	signedPriceUpdate publisher_agent.SignedPriceUpdate[T],
	latestContractValue chain_pusher_types.InternalTemporalNumericValue,
) bool {
	newPrice := new(big.Float)
	newPrice.SetString(string(signedPriceUpdate.SignedPrice.QuantizedPrice))

	contractPrice := new(big.Float)
	contractPrice.SetInt(latestContractValue.QuantizedValue)

	// Calculate absolute difference
	diff := new(big.Float).Sub(newPrice, contractPrice)
	absDiff := new(big.Float).Abs(diff)

	if contractPrice.Sign() == 0 {
		return absDiff.Sign() != 0
	}

	ratio := new(big.Float).Quo(absDiff, contractPrice)
	absRatio := new(big.Float).Abs(ratio)

	percentChange := new(big.Float).Mul(absRatio, big.NewFloat(100))
	threshold := big.NewFloat(assetEntry.PercentChangeThreshold)

	return percentChange.Cmp(threshold) > 0
}

func (r *FirstPartyRunner[T]) pushBatch(
	updates map[types.AssetEntry]publisher_agent.SignedPriceUpdate[T],
	latestContractValueMap map[types.PublisherAssetPair]chain_pusher_types.InternalTemporalNumericValue,
) {
	r.logger.Debug().
		Int("num_updates", len(updates)).
		Msg("Pushing batch to contract")

	err := r.contractInteractor.BatchPushToContract(updates)
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

		publisherAssetPair := types.PublisherAssetPair{
			Address: common.HexToAddress(string(entry.PublicKey)),
			AssetID: entry.AssetID,
		}

		latestContractValueMap[publisherAssetPair] = chain_pusher_types.InternalTemporalNumericValue{
			TimestampNs:    update.SignedPrice.TimestampedSignature.TimestampNano,
			QuantizedValue: quantizedValInt,
		}
	}
}

func (r *FirstPartyRunner[T]) poll(
	ctx context.Context,
	ch chan types.ContractUpdate,
	pubKeyAssetIDPairs map[common.Address][]shared.AssetID,
) {
	r.logger.Debug().Msg("Polling contract for new values")

	pollingTicker := time.NewTicker(time.Duration(r.pollingPeriodSecs) * time.Second)
	defer pollingTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-pollingTicker.C:
			latestContractUpdates, err := r.contractInteractor.PullValues(pubKeyAssetIDPairs)
			if err != nil {
				r.logger.Error().Err(err).Msg("Failed to pull values from contract")
			}

			for _, update := range latestContractUpdates {
				ch <- update
			}
		}
	}
}
