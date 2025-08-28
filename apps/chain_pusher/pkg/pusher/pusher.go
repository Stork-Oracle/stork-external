package pusher

import (
	"context"
	"math/big"
	"time"

	"github.com/Stork-Oracle/stork-external/apps/chain_pusher/pkg/types"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Pusher struct {
	storkWsEndpoint string
	storkAuth       string
	chainRpcUrl     string
	contractAddress string
	assetConfigFile string
	batchingWindow  int
	pollingPeriod   int
	interactor      types.ContractInteractor
	logger          *zerolog.Logger
}

func NewPusher(storkWsEndpoint, storkAuth, chainRpcUrl, contractAddress, assetConfigFile string, batchingWindow, pollingPeriod int, interactor types.ContractInteractor, logger *zerolog.Logger) *Pusher {
	return &Pusher{
		storkWsEndpoint: storkWsEndpoint,
		storkAuth:       storkAuth,
		chainRpcUrl:     chainRpcUrl,
		contractAddress: contractAddress,
		assetConfigFile: assetConfigFile,
		batchingWindow:  batchingWindow,
		pollingPeriod:   pollingPeriod,
		interactor:      interactor,
		logger:          logger,
	}
}

func (p *Pusher) Run(ctx context.Context) {
	priceConfig, err := types.LoadConfig(p.assetConfigFile)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load price config")
	}

	assetIDs := make([]types.AssetID, len(priceConfig.Assets))
	encodedAssetIDs := make([]types.InternalEncodedAssetID, len(priceConfig.Assets))

	i := 0
	for _, entry := range priceConfig.Assets {
		assetIDs[i] = entry.AssetID

		encoded, err := HexStringToByte32(string(entry.EncodedAssetID))
		if err != nil {
			p.logger.Fatal().Err(err).Msg("Failed to convert asset ID")
		}

		encodedAssetIDs[i] = encoded
		i++
	}

	storkWsCh := make(chan types.AggregatedSignedPrice)
	contractCh := make(chan map[types.InternalEncodedAssetID]types.InternalTemporalNumericValue)

	storkWs := NewStorkAggregatorWebsocketClient(p.storkWsEndpoint, p.storkAuth, assetIDs, p.logger)
	go storkWs.Run(storkWsCh)

	latestContractValueMap := make(map[types.InternalEncodedAssetID]types.InternalTemporalNumericValue)
	latestStorkValueMap := make(map[types.InternalEncodedAssetID]types.AggregatedSignedPrice)

	initialValues, err := p.interactor.PullValues(encodedAssetIDs)
	if err != nil {
		p.logger.Fatal().Err(err).Msg("Failed to pull initial values from contract")
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
			updates := make(map[types.InternalEncodedAssetID]types.AggregatedSignedPrice)

			for encodedAssetID, latestStorkPrice := range latestStorkValueMap {
				latestValue, ok := latestContractValueMap[encodedAssetID]
				if !ok {
					p.logger.Debug().Msgf("No current value for asset %s", latestStorkPrice.StorkSignedPrice.EncodedAssetID)
					updates[encodedAssetID] = latestStorkPrice
				} else if shouldUpdateAsset(
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
				}
				// include this to prevent race conditions
				for encodedAssetID, update := range updates {
					quantizedValInt := new(big.Int)
					quantizedValInt.SetString(string(update.StorkSignedPrice.QuantizedPrice), 10)

					latestContractValueMap[encodedAssetID] = types.InternalTemporalNumericValue{
						TimestampNs:    uint64(update.TimestampNano),
						QuantizedValue: quantizedValInt,
					}
				}
			} else {
				p.logger.Debug().Msg("No updates to push")
			}
		// Handle stork updates
		case valueUpdate := <-storkWsCh:
			encoded, err := HexStringToByte32(string(valueUpdate.StorkSignedPrice.EncodedAssetID))
			if err != nil {
				p.logger.Error().Err(err).Msg("Failed to convert asset ID")

				continue
			}

			latestStorkValueMap[encoded] = valueUpdate
		// Handle contract updates
		case chainUpdate := <-contractCh:
			for encodedAssetID, storkStructsTemporalNumericValue := range chainUpdate {
				latestContractValueMap[encodedAssetID] = storkStructsTemporalNumericValue
			}
		}
	}
}

func shouldUpdateAsset(latestValue types.InternalTemporalNumericValue, latestStorkPrice types.AggregatedSignedPrice, fallbackPeriodSecs uint64, changeThreshold float64) bool {
	if uint64(latestStorkPrice.TimestampNano)-latestValue.TimestampNs > fallbackPeriodSecs*1000000000 {
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

	// Multiply by 100 to get the percentage
	percentChange := new(big.Float).Mul(absRatio, big.NewFloat(100))

	thresholdBig := big.NewFloat(changeThreshold)

	return percentChange.Cmp(thresholdBig) > 0
}

func (p *Pusher) poll(encodedAssetIDs []types.InternalEncodedAssetID, ch chan map[types.InternalEncodedAssetID]types.InternalTemporalNumericValue) {
	p.logger.Info().Msgf("Polling contract for new values for %d assets", len(encodedAssetIDs))

	for range time.Tick(time.Duration(p.pollingPeriod) * time.Second) {
		polledVals, err := p.interactor.PullValues(encodedAssetIDs)
		if err != nil {
			p.logger.Error().Err(err).Msg("Failed to poll contract")

			continue
		}

		if len(polledVals) > 0 {
			ch <- polledVals
		}
	}
}
