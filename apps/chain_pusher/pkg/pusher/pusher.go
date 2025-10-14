package pusher

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/Stork-Oracle/stork-external/apps/chain_pusher/pkg/types"
	"github.com/Stork-Oracle/stork-external/shared"
	"github.com/rs/zerolog"
)

// Pusher is a struct that contains the configuration for the Pusher.
type Pusher struct {
	storkWsEndpoint string
	storkAuth       string
	chainRpcUrl     string
	chainWsRpcUrl   string
	contractAddress string
	assetConfigFile string
	batchingWindow  int
	pollingPeriod   int
	interactor      types.ContractInteractor
	logger          *zerolog.Logger
}

// NewPusher creates a new Pusher with the given parameters.
func NewPusher(
	storkWsEndpoint, storkAuth, chainRpcUrl, chainWsRpcUrl, contractAddress, assetConfigFile string,
	batchingWindow, pollingPeriod int,
	interactor types.ContractInteractor,
	logger *zerolog.Logger,
) *Pusher {
	return &Pusher{
		storkWsEndpoint: storkWsEndpoint,
		storkAuth:       storkAuth,
		chainRpcUrl:     chainRpcUrl,
		chainWsRpcUrl:   chainWsRpcUrl,
		contractAddress: contractAddress,
		assetConfigFile: assetConfigFile,
		batchingWindow:  batchingWindow,
		pollingPeriod:   pollingPeriod,
		interactor:      interactor,
		logger:          logger,
	}
}

// Run starts the Pusher.
func (p *Pusher) Run(ctx context.Context) {
	p.logger.Info().Str("wsRpcUrl", p.chainWsRpcUrl).Msg("Connecting to WS RPC URL")

	err := p.interactor.ConnectWs(p.chainWsRpcUrl)
	if err != nil {
		p.logger.Error().Err(err).
			Str("wsRpcUrl", p.chainWsRpcUrl).
			Msg("failed to connect to ws RPC")
	}

	p.logger.Info().Str("httpRpcUrl", p.chainRpcUrl).Msg("Connecting to HTTP RPC URL")

	err = p.interactor.ConnectHTTP(p.chainRpcUrl)
	if err != nil {
		p.logger.Error().Err(err).
			Str("httpRpcUrl", p.chainRpcUrl).
			Msg("failed to connect to HTTP RPC")
	}

	priceConfig, assetIDs, encodedAssetIDs, err := p.initializeAssets()
	if err != nil {
		p.logger.Fatal().Err(err).Msg("Failed to initialize assets")
	}

	storkWsCh := make(chan types.AggregatedSignedPrice)
	contractCh := make(chan map[types.InternalEncodedAssetID]types.InternalTemporalNumericValue)

	storkWs := NewStorkAggregatorWebsocketClient(p.storkWsEndpoint, p.storkAuth, assetIDs, p.logger)
	go storkWs.Run(storkWsCh)

	latestContractValueMap := make(map[types.InternalEncodedAssetID]types.InternalTemporalNumericValue)
	latestStorkValueMap := make(map[types.InternalEncodedAssetID]types.AggregatedSignedPrice)

	initialValues, err := p.interactor.PullValues(encodedAssetIDs)
	if err != nil {
		p.logger.Error().Err(err).Msg("Failed to pull initial values from contract")

		p.logger.Info().Str("httpRpcUrl", p.chainRpcUrl).Msg("Reconnecting to HTTP RPC URL")

		err = p.interactor.ConnectHTTP(p.chainRpcUrl)
		if err != nil {
			p.logger.Error().Err(err).Msg("failed to reconnect to HTTP RPC")
		}
	}

	for encodedAssetID, value := range initialValues {
		latestContractValueMap[encodedAssetID] = value
	}

	p.logger.Info().Msgf("Pulled initial values for %d assets", len(initialValues))

	go p.interactor.ListenContractEvents(ctx, contractCh)
	go p.poll(encodedAssetIDs, contractCh)

	ticker := time.NewTicker(time.Duration(p.batchingWindow) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			p.logger.Info().Msg("Pusher stopping due to context cancellation")

			return
		case <-ticker.C:
			p.handlePushUpdates(latestContractValueMap, latestStorkValueMap, priceConfig)
		// Handle stork updates
		case valueUpdate := <-storkWsCh:
			p.handleStorkUpdate(valueUpdate, latestStorkValueMap)
		// Handle contract updates
		case chainUpdate := <-contractCh:
			p.handleContractUpdate(chainUpdate, latestContractValueMap)
		}
	}
}

func shouldUpdateAsset(
	latestValue types.InternalTemporalNumericValue,
	latestStorkPrice types.AggregatedSignedPrice,
	fallbackPeriodSecs uint64,
	changeThreshold float64,
) bool {
	if latestStorkPrice.TimestampNano-latestValue.TimestampNs > fallbackPeriodSecs*uint64(time.Second) {
		return true
	}

	quantizedVal := new(big.Float)
	quantizedVal.SetString(string(latestStorkPrice.StorkSignedPrice.QuantizedPrice))

	quantizedCurrVal := new(big.Float)
	quantizedCurrVal.SetInt(latestValue.QuantizedValue)

	// Calculate the absolute difference
	difference := new(big.Float).Sub(quantizedVal, quantizedCurrVal)
	absDifference := new(big.Float).Abs(difference)

	if quantizedCurrVal.Sign() == 0 {
		return quantizedVal.Sign() != 0
	}

	// Calculate the ratio
	ratio := new(big.Float).Quo(absDifference, quantizedCurrVal)
	absRatio := new(big.Float).Abs(ratio)

	//nolint:mnd // The purpose of 100 here is evident in order to get a percentage
	percentChange := new(big.Float).Mul(absRatio, big.NewFloat(100))

	thresholdBig := big.NewFloat(changeThreshold)

	return percentChange.Cmp(thresholdBig) > 0
}

func (p *Pusher) poll(
	encodedAssetIDs []types.InternalEncodedAssetID,
	ch chan map[types.InternalEncodedAssetID]types.InternalTemporalNumericValue,
) {
	p.logger.Info().Msgf("Polling contract for new values for %d assets", len(encodedAssetIDs))

	for range time.Tick(time.Duration(p.pollingPeriod) * time.Second) {
		polledVals, err := p.interactor.PullValues(encodedAssetIDs)
		if err != nil {
			p.logger.Error().Err(err).Msg("Failed to poll contract")

			p.logger.Info().Str("httpRpcUrl", p.chainRpcUrl).Msg("Reconnecting to HTTP RPC URL")

			err = p.interactor.ConnectHTTP(p.chainRpcUrl)
			if err != nil {
				p.logger.Error().Err(err).Msg("failed to reconnect to HTTP RPC")
			}
		}

		if len(polledVals) > 0 {
			ch <- polledVals
		}
	}
}

// initializeAssets loads config and prepares asset IDs.
func (p *Pusher) initializeAssets() (*types.AssetConfig, []shared.AssetID, []types.InternalEncodedAssetID, error) {
	priceConfig, err := types.LoadConfig(p.assetConfigFile)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to load price config: %w", err)
	}

	assetIDs := make([]shared.AssetID, len(priceConfig.Assets))
	encodedAssetIDs := make([]types.InternalEncodedAssetID, len(priceConfig.Assets))

	i := 0
	for _, entry := range priceConfig.Assets {
		assetIDs[i] = entry.AssetID

		var encoded [32]byte

		encoded, err = HexStringToByte32(string(entry.EncodedAssetID))
		if err != nil {
			return nil, nil, nil, err
		}

		encodedAssetIDs[i] = encoded
		i++
	}

	return priceConfig, assetIDs, encodedAssetIDs, nil
}

// handlePushUpdates processes updates and pushes them to the contract.
func (p *Pusher) handlePushUpdates(
	latestContractValueMap map[types.InternalEncodedAssetID]types.InternalTemporalNumericValue,
	latestStorkValueMap map[types.InternalEncodedAssetID]types.AggregatedSignedPrice,
	priceConfig *types.AssetConfig,
) {
	updates := make(map[types.InternalEncodedAssetID]types.AggregatedSignedPrice)

	for encodedAssetID, latestStorkPrice := range latestStorkValueMap {
		latestValue, ok := latestContractValueMap[encodedAssetID]
		if !ok {
			p.logger.Debug().
				Msgf("No current value for asset %s", latestStorkPrice.StorkSignedPrice.EncodedAssetID)
			updates[encodedAssetID] = latestStorkPrice

			continue
		}

		if shouldUpdateAsset(
			latestValue,
			latestStorkPrice,
			priceConfig.Assets[latestStorkPrice.AssetID].FallbackPeriodSecs,
			priceConfig.Assets[latestStorkPrice.AssetID].PercentChangeThreshold,
		) {
			updates[encodedAssetID] = latestStorkPrice
		}
	}

	if len(updates) > 0 {
		err := p.interactor.BatchPushToContract(updates)
		if err != nil {
			p.logger.Error().Err(err).Msg("Failed to push batch to contract")

			p.logger.Info().Str("httpRpcUrl", p.chainRpcUrl).Msg("Reconnecting to HTTP RPC URL")

			err = p.interactor.ConnectHTTP(p.chainRpcUrl)
			if err != nil {
				p.logger.Error().Err(err).Msg("failed to reconnect to HTTP RPC")
			}
		} else {
			for encodedAssetID, update := range updates {
				quantizedValInt := new(big.Int)
				//nolint:mnd // Base number
				quantizedValInt.SetString(string(update.StorkSignedPrice.QuantizedPrice), 10)

				latestContractValueMap[encodedAssetID] = types.InternalTemporalNumericValue{
					TimestampNs:    update.TimestampNano,
					QuantizedValue: quantizedValInt,
				}
			}
		}
	}
}

// handleStorkUpdate processes updates from the Stork websocket.
func (p *Pusher) handleStorkUpdate(
	valueUpdate types.AggregatedSignedPrice,
	latestStorkValueMap map[types.InternalEncodedAssetID]types.AggregatedSignedPrice,
) {
	encoded, err := HexStringToByte32(string(valueUpdate.StorkSignedPrice.EncodedAssetID))
	if err != nil {
		p.logger.Error().Err(err).Msg("Failed to convert asset ID")

		return
	}

	latestStorkValueMap[encoded] = valueUpdate
}

// handleContractUpdate processes updates from contract events.
func (p *Pusher) handleContractUpdate(
	chainUpdate map[types.InternalEncodedAssetID]types.InternalTemporalNumericValue,
	latestContractValueMap map[types.InternalEncodedAssetID]types.InternalTemporalNumericValue,
) {
	for encodedAssetID, storkStructsTemporalNumericValue := range chainUpdate {
		latestContractValueMap[encodedAssetID] = storkStructsTemporalNumericValue
	}
}
