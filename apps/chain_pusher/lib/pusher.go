package chain_pusher

import (
	"context"
	"math/big"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Pusher struct {
	storkWsEndpoint  string
	storkAuth        string
	chainRpcUrl      string
	contractAddress  string
	assetConfigFile  string
	verifyPublishers bool
	batchingWindow   int
	pollingPeriod    int
	interactor       ContractInteractor
	logger           *zerolog.Logger
}

func NewPusher(storkWsEndpoint, storkAuth, chainRpcUrl, contractAddress, assetConfigFile string, batchingWindow, pollingPeriod int, interactor ContractInteractor, logger *zerolog.Logger) *Pusher {
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
	priceConfig, err := LoadConfig(p.assetConfigFile)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load price config")
	}

	assetIds := make([]AssetId, len(priceConfig.Assets))
	encodedAssetIds := make([]InternalEncodedAssetId, len(priceConfig.Assets))
	i := 0
	for _, entry := range priceConfig.Assets {
		assetIds[i] = entry.AssetId
		encoded, err := stringToByte32(string(entry.EncodedAssetId))
		if err != nil {
			p.logger.Fatal().Err(err).Msg("Failed to convert asset ID")
		}
		encodedAssetIds[i] = encoded
		i++
	}

	storkWsCh := make(chan AggregatedSignedPrice)
	contractCh := make(chan map[InternalEncodedAssetId]InternalTemporalNumericValue)

	storkWs := NewStorkAggregatorWebsocketClient(p.storkWsEndpoint, p.storkAuth, assetIds, p.logger)
	go storkWs.Run(storkWsCh)

	latestContractValueMap := make(map[InternalEncodedAssetId]InternalTemporalNumericValue)
	latestStorkValueMap := make(map[InternalEncodedAssetId]AggregatedSignedPrice)

	initialValues, err := p.interactor.PullValues(encodedAssetIds)
	if err != nil {
		p.logger.Fatal().Err(err).Msg("Failed to pull initial values from contract")
	}
	for encodedAssetId, value := range initialValues {
		latestContractValueMap[encodedAssetId] = value
	}
	p.logger.Info().Msgf("Pulled initial values for %d assets", len(initialValues))

	go p.interactor.ListenContractEvents(ctx, contractCh)
	go p.poll(encodedAssetIds, contractCh)

	ticker := time.NewTicker(time.Duration(p.batchingWindow) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			p.logger.Info().Msg("Pusher stopping due to context cancellation")
			return
		case <-ticker.C:
			updates := make(map[InternalEncodedAssetId]AggregatedSignedPrice)
			for encodedAssetId, latestStorkPrice := range latestStorkValueMap {
				latestValue, ok := latestContractValueMap[encodedAssetId]
				if !ok {
					p.logger.Debug().Msgf("No current value for asset %s", latestStorkPrice.StorkSignedPrice.EncodedAssetId)
					updates[encodedAssetId] = latestStorkPrice
				} else if shouldUpdateAsset(
					latestValue,
					latestStorkPrice,
					priceConfig.Assets[latestStorkPrice.AssetId].FallbackPeriodSecs,
					priceConfig.Assets[latestStorkPrice.AssetId].PercentChangeThreshold,
				) {
					updates[encodedAssetId] = latestStorkPrice
				}
			}

			if len(updates) > 0 {
				err := p.interactor.BatchPushToContract(updates)
				if err != nil {
					p.logger.Error().Err(err).Msg("Failed to push batch to contract")
				}
				// include this to prevent race conditions
				for encodedAssetId, update := range updates {
					quantizedValInt := new(big.Int)
					quantizedValInt.SetString(string(update.StorkSignedPrice.QuantizedPrice), 10)

					latestContractValueMap[encodedAssetId] = InternalTemporalNumericValue{
						TimestampNs:    uint64(update.TimestampNano),
						QuantizedValue: quantizedValInt,
					}
				}
			} else {
				p.logger.Debug().Msg("No updates to push")
			}
		// Handle stork updates
		case valueUpdate := <-storkWsCh:
			encoded, err := stringToByte32(string(valueUpdate.StorkSignedPrice.EncodedAssetId))
			if err != nil {
				p.logger.Error().Err(err).Msg("Failed to convert asset ID")
				continue
			}
			latestStorkValueMap[encoded] = valueUpdate
		// Handle contract updates
		case chainUpdate := <-contractCh:
			for encodedAssetId, storkStructsTemporalNumericValue := range chainUpdate {
				latestContractValueMap[encodedAssetId] = storkStructsTemporalNumericValue
			}
		}
	}
}

func shouldUpdateAsset(latestValue InternalTemporalNumericValue, latestStorkPrice AggregatedSignedPrice, fallbackPeriodSecs uint64, changeThreshold float64) bool {
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

	// Multiply by 100 to get the percentage
	percentChange := new(big.Float).Mul(ratio, big.NewFloat(100))

	thresholdBig := big.NewFloat(changeThreshold)
	return percentChange.Cmp(thresholdBig) > 0
}

func (p *Pusher) poll(encodedAssetIds []InternalEncodedAssetId, ch chan map[InternalEncodedAssetId]InternalTemporalNumericValue) {
	p.logger.Info().Msgf("Polling contract for new values for %d assets", len(encodedAssetIds))
	for range time.Tick(time.Duration(p.pollingPeriod) * time.Second) {
		polledVals, err := p.interactor.PullValues(encodedAssetIds)
		if err != nil {
			p.logger.Error().Err(err).Msg("Failed to poll contract")
			continue
		}
		if len(polledVals) > 0 {
			ch <- polledVals
		}
	}
}
