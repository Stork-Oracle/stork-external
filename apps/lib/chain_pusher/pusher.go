package chain_pusher

import (
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
	mnemonicFile     string
	verifyPublishers bool
	batchingWindow   int
	pollingFrequency int
	interacter       ContractInteracter
	logger           *zerolog.Logger
}

func NewPusher(storkWsEndpoint, storkAuth, chainRpcUrl, contractAddress, assetConfigFile, mnemonicFile string, verifyPublishers bool, batchingWindow, pollingFrequency int, interacter ContractInteracter, logger *zerolog.Logger) *Pusher {
	return &Pusher{
		storkWsEndpoint:  storkWsEndpoint,
		storkAuth:        storkAuth,
		chainRpcUrl:      chainRpcUrl,
		contractAddress:  contractAddress,
		assetConfigFile:  assetConfigFile,
		mnemonicFile:     mnemonicFile,
		verifyPublishers: verifyPublishers,
		batchingWindow:   batchingWindow,
		pollingFrequency: pollingFrequency,
		interacter:       interacter,
		logger:           logger,
	}
}

func (p *Pusher) Run() {
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
	contractCh := make(chan map[InternalEncodedAssetId]InternalStorkStructsTemporalNumericValue)

	storkWs := NewStorkAggregatorWebsocketClient(p.storkWsEndpoint, p.storkAuth, assetIds, p.logger)
	go storkWs.Run(storkWsCh)

	latestContractValueMap := make(map[InternalEncodedAssetId]InternalStorkStructsTemporalNumericValue)
	latestStorkValueMap := make(map[InternalEncodedAssetId]AggregatedSignedPrice)

	initialValues, err := p.interacter.PullValues(encodedAssetIds)
	if err != nil {
		p.logger.Fatal().Err(err).Msg("Failed to pull initial values from contract")
	}
	for encodedAssetId, value := range initialValues {
		latestContractValueMap[encodedAssetId] = value
	}
	p.logger.Info().Msgf("Pulled initial values for %d assets", len(initialValues))

	go p.interacter.ListenContractEvents(contractCh)
	go p.poll(encodedAssetIds, contractCh)

	ticker := time.NewTicker(time.Duration(p.batchingWindow) * time.Second)
	defer ticker.Stop()

	for {
		select {
		// Determine updates after the batching window has passed and push them to the contract
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
				err := p.interacter.BatchPushToContract(updates)
				if err != nil {
					p.logger.Error().Err(err).Msg("Failed to push batch to contract")
				}
				// include this to prevent race conditions
				for encodedAssetId, update := range updates {
					quantizedValInt := new(big.Int)
					quantizedValInt.SetString(string(update.StorkSignedPrice.QuantizedPrice), 10)

					latestContractValueMap[encodedAssetId] = InternalStorkStructsTemporalNumericValue{
						TimestampNs:    uint64(update.Timestamp),
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

func shouldUpdateAsset(latestValue InternalStorkStructsTemporalNumericValue, latestStorkPrice AggregatedSignedPrice, fallbackPeriodSecs uint64, changeThreshold float64) bool {
	if uint64(latestStorkPrice.Timestamp)-latestValue.TimestampNs > fallbackPeriodSecs*1000000000 {
		return true
	}

	quantizedVal := new(big.Float)
	quantizedVal.SetString(string(latestStorkPrice.StorkSignedPrice.QuantizedPrice))

	quantizedCurrVal := new(big.Float)
	quantizedCurrVal.SetInt(latestValue.QuantizedValue)

	// Calculate the absolute difference
	difference := new(big.Float).Sub(quantizedVal, quantizedCurrVal)
	absDifference := new(big.Float).Abs(difference)

	// Calculate the ratio
	ratio := new(big.Float).Quo(absDifference, quantizedCurrVal)

	// Multiply by 100 to get the percentage
	percentChange := new(big.Float).Mul(ratio, big.NewFloat(100))

	thresholdBig := big.NewFloat(changeThreshold)
	return percentChange.Cmp(thresholdBig) > 0
}

func (p *Pusher) poll(encodedAssetIds []InternalEncodedAssetId, ch chan map[InternalEncodedAssetId]InternalStorkStructsTemporalNumericValue) {
	p.logger.Info().Msgf("Polling contract for new values for %d assets", len(encodedAssetIds))
	for _ = range time.Tick(time.Duration(p.pollingFrequency) * time.Second) {
		polledVals, err := p.interacter.PullValues(encodedAssetIds)
		if err != nil {
			p.logger.Error().Err(err).Msg("Failed to poll contract")
			continue
		}
		if len(polledVals) > 0 {
			ch <- polledVals
		}
	}
}
