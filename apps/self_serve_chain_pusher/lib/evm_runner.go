package self_serve_chain_pusher

import (
	"context"
	"math/big"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type EvmSelfServeRunner struct {
	config              *EvmSelfServeConfig
	logger              zerolog.Logger
	websocketServer     *WebsocketServer
	contractInteractor  *SelfServeContractInteractor
	valueUpdateCh       chan ValueUpdate
	assetStates         map[string]*AssetPushState
	assetStatesMutex    sync.RWMutex
	ctx                 context.Context
	cancel              context.CancelFunc
}

func NewEvmSelfServeRunner(config *EvmSelfServeConfig) *EvmSelfServeRunner {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &EvmSelfServeRunner{
		config:           config,
		logger:           log.With().Str("component", "evm_runner").Logger(),
		valueUpdateCh:    make(chan ValueUpdate, 1000),
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
		r.config.LimitPerSecond,
		r.config.BurstLimit,
		r.logger,
	)
	if err != nil {
		r.logger.Fatal().Err(err).Msg("Failed to initialize contract interactor")
		return
	}
	defer r.contractInteractor.Close()

	// Initialize websocket server
	r.websocketServer = NewWebsocketServer(r.config.WebsocketPort, r.valueUpdateCh)

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
			PendingValue: nil,
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
			
		case valueUpdate := <-r.valueUpdateCh:
			r.handleValueUpdate(valueUpdate)
		}
	}
}

func (r *EvmSelfServeRunner) handleValueUpdate(valueUpdate ValueUpdate) {
	r.assetStatesMutex.Lock()
	defer r.assetStatesMutex.Unlock()

	assetState, exists := r.assetStates[valueUpdate.Asset]
	if !exists {
		r.logger.Debug().Str("asset", valueUpdate.Asset).Msg("Received update for unconfigured asset")
		return
	}

	r.logger.Debug().
		Str("asset", valueUpdate.Asset).
		Str("value", valueUpdate.Value.Text('f', 6)).
		Msg("Processing value update")

	// Update pending value
	assetState.PendingValue = &valueUpdate

	// Check if we should trigger a push based on price change
	if r.shouldPushBasedOnDelta(assetState, valueUpdate.Value) {
		r.logger.Info().
			Str("asset", valueUpdate.Asset).
			Str("old_price", assetState.LastPrice.Text('f', 6)).
			Str("new_price", valueUpdate.Value.Text('f', 6)).
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
		if now.After(state.NextPushTime) && state.PendingValue != nil {
			r.logger.Info().
				Str("asset", assetId).
				Time("next_push_time", state.NextPushTime).
				Msg("Triggering push due to time interval")
			
			r.triggerPush(state)
		}
	}
}

func (r *EvmSelfServeRunner) triggerPush(state *AssetPushState) {
	if state.PendingValue == nil {
		return
	}

	// Generate nonce for this push
	nonce := GenerateNonce()

	// Push to contract
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		err := r.contractInteractor.PushValue(ctx, state.Config, state.PendingValue.Value, nonce)
		if err != nil {
			r.logger.Error().
				Err(err).
				Str("asset", state.AssetId).
				Msg("Failed to push value to contract")
			return
		}

		// Update state after successful push
		r.assetStatesMutex.Lock()
		state.LastPrice = new(big.Float).Set(state.PendingValue.Value)
		state.LastPushTime = time.Now()
		state.NextPushTime = time.Now().Add(time.Duration(state.Config.PushIntervalSec) * time.Second)
		state.PendingValue = nil
		r.assetStatesMutex.Unlock()

		r.logger.Info().
			Str("asset", state.AssetId).
			Str("value", state.LastPrice.Text('f', 6)).
			Msg("Successfully pushed value to contract")
	}()
}