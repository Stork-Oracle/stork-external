package self_serve_evm

import (
	"context"
	"math/big"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"fmt"

	publisher_agent "github.com/Stork-Oracle/stork-external/apps/publisher_agent/pkg"
	"github.com/Stork-Oracle/stork-external/shared/signer"
)

type EvmSelfServeRunner struct {
	config             *EvmSelfServeConfig
	logger             zerolog.Logger
	websocketServer    *WebsocketServer
	contractInteractor *SelfServeContractInteractor
	signedPriceUpdateCh      chan publisher_agent.SignedPriceUpdate[*signer.EvmSignature]
	assetStates        map[string]*AssetPushState
	assetStatesMutex   sync.RWMutex
	ctx                context.Context
	cancel             context.CancelFunc
}

func NewEvmSelfServeRunner(config *EvmSelfServeConfig) *EvmSelfServeRunner {
	ctx, cancel := context.WithCancel(context.Background())

	return &EvmSelfServeRunner{
		config:           config,
		logger:           log.With().Str("component", "evm_runner").Logger(),
		signedPriceUpdateCh:    make(chan publisher_agent.SignedPriceUpdate[*signer.EvmSignature], 1000),
		assetStates:      make(map[string]*AssetPushState),
		assetStatesMutex: sync.RWMutex{},
		ctx:              ctx,
		cancel:           cancel,
	}
}

func (r *EvmSelfServeRunner) Run() {
	r.logger.Info().Msg("Starting EVM Self-Serve Chain Pusher")

	// Initialize asset states
	r.initializeAssetStates()

	// Initialize contract interactor
	var err error
	r.contractInteractor, err = NewSelfServeContractInteractor(
		r.config.ChainRpcUrl,
		r.config.ChainWsUrl,
		r.config.ContractAddress,
		r.config.PrivateKey,
		r.config.GasLimit,
		r.logger,
	)
	if err != nil {
		r.logger.Fatal().Err(err).Msg("Failed to initialize contract interactor")
		return
	}
	defer r.contractInteractor.Close()

	// Initialize websocket server
	r.websocketServer = NewWebsocketServer(r.config.WebsocketPort, r.signedPriceUpdateCh)

	// Start processing goroutines
	go r.processValueUpdates()
	go r.processPushTriggers()

	// Start websocket server (blocking)
	r.logger.Info().Str("port", r.config.WebsocketPort).Msg("Starting WebSocket server")
	if err := r.websocketServer.Start(); err != nil {
		r.logger.Fatal().Err(err).Msg("WebSocket server failed")
	}
}

func (r *EvmSelfServeRunner) Stop() {
	r.logger.Info().Msg("Stopping EVM Self-Serve Chain Pusher")
	r.cancel()

	if r.websocketServer != nil {
		r.websocketServer.Stop()
	}
}

func (r *EvmSelfServeRunner) initializeAssetStates() {
	r.assetStatesMutex.Lock()
	defer r.assetStatesMutex.Unlock()

	for assetId, assetConfig := range r.config.AssetConfig.Assets {
		r.assetStates[assetId] = &AssetPushState{
			AssetId:      assetId,
			Config:       assetConfig,
			LastPrice:    nil,
			LastPushTime: time.Time{},
			PendingSignedPriceUpdate: nil,
			NextPushTime: time.Now().Add(time.Duration(assetConfig.PushIntervalSec) * time.Second),
		}

		r.logger.Info().
			Str("asset", assetId).
			Int("push_interval_sec", assetConfig.PushIntervalSec).
			Float64("percent_threshold", assetConfig.PercentChangeThreshold).
			Msg("Initialized asset push state")
	}
}

func (r *EvmSelfServeRunner) processValueUpdates() {
	r.logger.Info().Msg("Starting value update processor")

	for {
		select {
		case <-r.ctx.Done():
			r.logger.Info().Msg("Value update processor stopped")
			return

		case signedPriceUpdate := <-r.signedPriceUpdateCh:
			r.handleSignedPriceUpdate(signedPriceUpdate)
		}
	}
}

func (r *EvmSelfServeRunner) handleSignedPriceUpdate(signedPriceUpdate publisher_agent.SignedPriceUpdate[*signer.EvmSignature]) {
	r.assetStatesMutex.Lock()
	defer r.assetStatesMutex.Unlock()

	assetId := string(signedPriceUpdate.AssetId)
	assetState, exists := r.assetStates[assetId]
	if !exists {
		r.logger.Debug().Str("asset", assetId).Msg("Received update for unconfigured asset")
		return
	}

	// Convert quantized price to big.Float for comparison
	priceValue, err := r.convertQuantizedPriceToBigFloat(string(signedPriceUpdate.SignedPrice.QuantizedPrice))
	if err != nil {
		r.logger.Error().Err(err).Str("asset", assetId).Msg("Failed to convert quantized price")
		return
	}

	r.logger.Debug().
		Str("asset", assetId).
		Str("price", string(signedPriceUpdate.SignedPrice.QuantizedPrice)).
		Msg("Processing signed price update")

	// Update pending value
	assetState.PendingSignedPriceUpdate = &signedPriceUpdate

	// Check if we should trigger a push based on price change
	if r.shouldPushBasedOnDelta(assetState, priceValue) {
		r.logger.Info().
			Str("asset", assetId).
			Str("old_price", assetState.LastPrice.Text('f', 6)).
			Str("new_price", priceValue.Text('f', 6)).
			Msg("Triggering push due to price delta threshold")

		r.triggerPush(assetState)
	}
}

func (r *EvmSelfServeRunner) shouldPushBasedOnDelta(state *AssetPushState, newPrice *big.Float) bool {
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

func (r *EvmSelfServeRunner) processPushTriggers() {
	r.logger.Info().Msg("Starting push trigger processor")

	ticker := time.NewTicker(1 * time.Second) // Check every second
	defer ticker.Stop()

	for {
		select {
		case <-r.ctx.Done():
			r.logger.Info().Msg("Push trigger processor stopped")
			return

		case <-ticker.C:
			r.checkTimerTriggers()
		}
	}
}

func (r *EvmSelfServeRunner) checkTimerTriggers() {
	r.assetStatesMutex.Lock()
	defer r.assetStatesMutex.Unlock()

	now := time.Now()
	for assetId, state := range r.assetStates {
		if now.After(state.NextPushTime) && state.PendingSignedPriceUpdate != nil {
			r.logger.Info().
				Str("asset", assetId).
				Time("next_push_time", state.NextPushTime).
				Msg("Triggering push due to time interval")

			r.triggerPush(state)
		}
	}
}

func (r *EvmSelfServeRunner) triggerPush(state *AssetPushState) {
	if state.PendingSignedPriceUpdate == nil {
		return
	}

	// Push to contract
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		err := r.contractInteractor.PushSignedPriceUpdate(ctx, state.Config, *state.PendingSignedPriceUpdate)
		if err != nil {
			r.logger.Error().
				Err(err).
				Str("asset", state.AssetId).
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
			Str("asset", state.AssetId).
			Str("value", state.LastPrice.Text('f', 6)).
			Msg("Successfully pushed value to contract")
	}()
}

func (r *EvmSelfServeRunner) convertQuantizedPriceToBigFloat(quantizedPrice string) (*big.Float, error) {
	bf, success := new(big.Float).SetString(quantizedPrice)
	if !success {
		return nil, fmt.Errorf("failed to convert quantized price to big.Float: %s", quantizedPrice)
	}
	return bf, nil
}
