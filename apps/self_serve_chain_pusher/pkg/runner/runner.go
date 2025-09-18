package runner

import (
	"context"
	"math/big"
	"sync"
	"time"

	"github.com/rs/zerolog"

	"fmt"

	publisher_agent "github.com/Stork-Oracle/stork-external/apps/publisher_agent/pkg"
	"github.com/Stork-Oracle/stork-external/apps/self_serve_chain_pusher/pkg/types"
	"github.com/Stork-Oracle/stork-external/shared"
)

type SelfServeRunner struct {
	config             *types.SelfServeConfig
	contractInteractor types.ContractInteractor
	websocketServer    *WebsocketServer

	signedPriceUpdateCh chan publisher_agent.SignedPriceUpdate[*shared.EvmSignature]
	assetStates         map[shared.AssetID]*types.AssetPushState
	assetStatesMutex    sync.RWMutex

	cancel context.CancelFunc
	logger zerolog.Logger
}

func NewSelfServeRunner(
	config *types.SelfServeConfig,
	contractInteractor types.ContractInteractor,
	cancel context.CancelFunc,
	logger zerolog.Logger,
) *SelfServeRunner {
	return &SelfServeRunner{
		config:              config,
		contractInteractor:  contractInteractor,
		websocketServer:     nil,
		signedPriceUpdateCh: make(chan publisher_agent.SignedPriceUpdate[*shared.EvmSignature], 1000),
		assetStates:         make(map[shared.AssetID]*types.AssetPushState),
		assetStatesMutex:    sync.RWMutex{},
		cancel:              cancel,
		logger:              logger.With().Str("component", "self_serve_runner").Logger(),
	}
}

func (r *SelfServeRunner) Run(ctx context.Context) {
	r.logger.Info().Msg("Starting EVM Self-Serve Chain Pusher")

	// Initialize asset states
	r.initializeAssetStates()

	// // TODO: move out of runner and into
	// // Initialize contract interactor
	// var err error
	// r.contractInteractor, err = NewContractInteractor(
	// 	r.config.ChainRpcUrl,
	// 	r.config.ChainWsUrl,
	// 	r.config.ContractAddress,
	// 	r.config.PrivateKey,
	// 	r.config.GasLimit,
	// 	r.logger,
	// )
	// if err != nil {
	// 	r.logger.Fatal().Err(err).Msg("Failed to initialize contract interactor")
	// 	return
	// }
	// defer r.contractInteractor.Close()

	// Initialize websocket server
	r.websocketServer = NewWebsocketServer(r.config.WebsocketPort, r.signedPriceUpdateCh)

	// Start processing goroutines
	go r.processValueUpdates(ctx)
	go r.processPushTriggers(ctx)

	// Start websocket server (blocking)
	r.logger.Info().Str("port", r.config.WebsocketPort).Msg("Starting WebSocket server")
	if err := r.websocketServer.Start(); err != nil {
		r.logger.Fatal().Err(err).Msg("WebSocket server failed")
	}
}

func (r *SelfServeRunner) Stop() {
	r.logger.Info().Msg("Stopping EVM Self-Serve Chain Pusher")
	r.cancel()

	if r.websocketServer != nil {
		r.websocketServer.Stop()
	}
}

func (r *SelfServeRunner) initializeAssetStates() {
	r.assetStatesMutex.Lock()
	defer r.assetStatesMutex.Unlock()

	for assetID, assetConfig := range r.config.AssetConfig.Assets {
		r.assetStates[assetID] = &types.AssetPushState{
			AssetID:                  assetID,
			Config:                   assetConfig,
			LastPrice:                nil,
			LastPushTime:             time.Time{},
			PendingSignedPriceUpdate: nil,
			NextPushTime:             time.Now().Add(time.Duration(assetConfig.PushIntervalSec) * time.Second),
		}

		r.logger.Info().
			Str("asset", string(assetID)).
			Int("push_interval_sec", assetConfig.PushIntervalSec).
			Float64("percent_threshold", assetConfig.PercentChangeThreshold).
			Msg("Initialized asset push state")
	}
}

func (r *SelfServeRunner) processValueUpdates(ctx context.Context) {
	r.logger.Info().Msg("Starting value update processor")

	for {
		select {
		case <-ctx.Done():
			r.logger.Info().Msg("Value update processor stopped")
			return

		case signedPriceUpdate := <-r.signedPriceUpdateCh:
			r.handleSignedPriceUpdate(ctx, signedPriceUpdate)
		}
	}
}

func (r *SelfServeRunner) handleSignedPriceUpdate(ctx context.Context, signedPriceUpdate publisher_agent.SignedPriceUpdate[*shared.EvmSignature]) {
	r.assetStatesMutex.Lock()
	defer r.assetStatesMutex.Unlock()

	assetID := signedPriceUpdate.AssetID
	assetState, exists := r.assetStates[assetID]
	if !exists {
		r.logger.Debug().Str("asset", string(assetID)).Msg("Received update for unconfigured asset")
		return
	}

	// Convert quantized price to big.Float for comparison
	priceValue, err := r.convertQuantizedPriceToBigFloat(string(signedPriceUpdate.SignedPrice.QuantizedPrice))
	if err != nil {
		r.logger.Error().Err(err).Str("asset", string(assetID)).Msg("Failed to convert quantized price")
		return
	}

	r.logger.Debug().
		Str("asset", string(assetID)).
		Str("price", string(signedPriceUpdate.SignedPrice.QuantizedPrice)).
		Msg("Processing signed price update")

	// Update pending value
	assetState.PendingSignedPriceUpdate = &signedPriceUpdate

	// Check if we should trigger a push based on price change
	if r.shouldPushBasedOnDelta(assetState, priceValue) {
		r.logger.Info().
			Str("asset", string(assetID)).
			Str("old_price", assetState.LastPrice.Text('f', 6)).
			Str("new_price", priceValue.Text('f', 6)).
			Msg("Triggering push due to price delta threshold")

		r.triggerPush(ctx, assetState)
	}
}

func (r *SelfServeRunner) shouldPushBasedOnDelta(state *types.AssetPushState, newPrice *big.Float) bool {
	if state.LastPrice == nil {
		return true // First price update
	}

	// Calculate percentage change
	diff := new(big.Float).Sub(newPrice, state.LastPrice)
	percentChange := new(big.Float).Quo(diff, state.LastPrice)
	percentChange.Mul(percentChange, big.NewFloat(100))

	absPercentChange := new(big.Float).Abs(percentChange)
	threshold := big.NewFloat(state.Config.PercentChangeThreshold)

	return absPercentChange.Cmp(threshold) >= 0
}

func (r *SelfServeRunner) processPushTriggers(ctx context.Context) {
	r.logger.Info().Msg("Starting push trigger processor")

	ticker := time.NewTicker(1 * time.Second) // Check every second
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			r.logger.Info().Msg("Push trigger processor stopped")
			return

		case <-ticker.C:
			r.checkTimerTriggers(ctx)
		}
	}
}

func (r *SelfServeRunner) checkTimerTriggers(ctx context.Context) {
	r.assetStatesMutex.Lock()
	defer r.assetStatesMutex.Unlock()

	now := time.Now()
	for assetID, state := range r.assetStates {
		if now.After(state.NextPushTime) && state.PendingSignedPriceUpdate != nil {
			r.logger.Info().
				Str("asset", string(assetID)).
				Time("next_push_time", state.NextPushTime).
				Msg("Triggering push due to time interval")

			r.triggerPush(ctx, state)
		}
	}
}

func (r *SelfServeRunner) triggerPush(parentCtx context.Context, state *types.AssetPushState) {
	if state.PendingSignedPriceUpdate == nil {
		return
	}

	// Push to contract
	go func() {
		ctx, cancel := context.WithTimeout(parentCtx, 30*time.Second)
		defer cancel()

		err := r.contractInteractor.PushSignedPriceUpdate(ctx, state.Config, *state.PendingSignedPriceUpdate)
		if err != nil {
			r.logger.Error().
				Err(err).
				Str("asset", string(state.AssetID)).
				Msg("Failed to push value to contract")
			return
		}

		// Update state after successful push
		r.assetStatesMutex.Lock()
		priceValue, _ := r.convertQuantizedPriceToBigFloat(string(state.PendingSignedPriceUpdate.SignedPrice.QuantizedPrice))
		state.LastPrice = priceValue
		state.LastPushTime = time.Now()
		state.NextPushTime = time.Now().Add(time.Duration(state.Config.PushIntervalSec) * time.Second)
		state.PendingSignedPriceUpdate = nil
		r.assetStatesMutex.Unlock()

		r.logger.Info().
			Str("asset", string(state.AssetID)).
			Str("value", state.LastPrice.Text('f', 6)).
			Msg("Successfully pushed value to contract")
	}()
}

func (r *SelfServeRunner) convertQuantizedPriceToBigFloat(quantizedPrice string) (*big.Float, error) {
	bf, success := new(big.Float).SetString(quantizedPrice)
	if !success {
		return nil, fmt.Errorf("failed to convert quantized price to big.Float: %s", quantizedPrice)
	}
	return bf, nil
}
