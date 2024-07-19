package main

import (
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var evmpushCmd = &cobra.Command{
	Use:   "evm-push",
	Short: "Push WebSocket prices to EVM contract",
	Run:   runEvmPush,
}

func init() {
	evmpushCmd.Flags().StringP("stork-ws-endpoint", "w", "", "Stork WebSocket endpoint")
	evmpushCmd.Flags().StringP("stork-auth-credentials", "a", "", "Stork auth credentials")
	evmpushCmd.Flags().StringP("chain-rpc-url", "c", "", "Chain RPC URL")
	evmpushCmd.Flags().StringP("contract-address", "x", "", "Contract address")
	evmpushCmd.Flags().StringP("asset-config-file", "f", "", "Asset config file")
	evmpushCmd.Flags().StringP("mnemonic-file", "m", "", "Mnemonic file")
	evmpushCmd.Flags().IntP("batching-window", "b", 10, "Pushing frequency (seconds)")
	evmpushCmd.Flags().IntP("polling-frequency", "p", 5, "Asset Polling frequency (seconds)")

	evmpushCmd.MarkFlagRequired("stork-ws-endpoint")
	evmpushCmd.MarkFlagRequired("stork-auth-credentials")
	evmpushCmd.MarkFlagRequired("chain-rpc-url")
	evmpushCmd.MarkFlagRequired("contract-address")
	evmpushCmd.MarkFlagRequired("asset-config-file")
	evmpushCmd.MarkFlagRequired("mnemonic-file")
}

func runEvmPush(cmd *cobra.Command, args []string) {
	logger := log.With().Str("component", "evm-push").Logger()

	storkAuth, _ := cmd.Flags().GetString("stork-auth-credentials")
	storkWsEndpoint, _ := cmd.Flags().GetString("stork-ws-endpoint")
	chainRpcUrl, _ := cmd.Flags().GetString("chain-rpc-url")
	contractAddress, _ := cmd.Flags().GetString("contract-address")
	assetConfigFile, _ := cmd.Flags().GetString("asset-config-file")
	mnemonicFile, _ := cmd.Flags().GetString("mnemonic-file")
	batchingWindow, _ := cmd.Flags().GetInt("pushing-frequency")
	pollingFrequency, _ := cmd.Flags().GetInt("polling-frequency")

	priceConfig, err := LoadConfig(assetConfigFile)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load price config")
	}

	assetIds := make([]AssetId, len(priceConfig.Assets))
	encodedAssetIds := make([]InternalEncodedAssetId, len(priceConfig.Assets))
	i := 0
	for _, entry := range priceConfig.Assets {
		assetIds[i] = entry.AssetId
		encodedAssetIds[i] = InternalEncodedAssetId(crypto.Keccak256([]byte(entry.AssetId)))
		i++
	}

	storkWsCh := make(chan AggregatedSignedPrice)
	contractCh := make(chan map[InternalEncodedAssetId]StorkStructsTemporalNumericValue)

	storkWs := NewStorkAggregatorWebsocketClient(storkWsEndpoint, storkAuth, assetIds, log.With().Str("component", "stork-ws").Logger())
	go storkWs.Run(storkWsCh)

	storkContractInterfacer := NewStorkContractInterfacer(chainRpcUrl, contractAddress, mnemonicFile, pollingFrequency)
	go storkContractInterfacer.ListenContractEvents(contractCh)
	go storkContractInterfacer.Poll(encodedAssetIds, contractCh)

	updates := make(map[InternalEncodedAssetId]AggregatedSignedPrice)
	latestValueMap := make(map[InternalEncodedAssetId]StorkStructsTemporalNumericValue)

	ticker := time.NewTicker(time.Duration(batchingWindow) * time.Second)
	defer ticker.Stop()

	for {
		select {
		// Push the batch to the contract after waiting the pushing frequency
		case <-ticker.C:
			if len(updates) > 0 {
				err := storkContractInterfacer.BatchPushToContract(updates)
				if err != nil {
					logger.Error().Err(err).Msg("Failed to push batch to contract")
				}
				for encodedAssetId, update := range updates {
					quantizedValInt := new(big.Int)
					quantizedValInt.SetString(string(update.StorkSignedPrice.QuantizedPrice), 10)

					latestValueMap[encodedAssetId] = StorkStructsTemporalNumericValue{
						TimestampNs:    uint64(update.Timestamp),
						QuantizedValue: quantizedValInt,
					}
				}
				updates = make(map[InternalEncodedAssetId]AggregatedSignedPrice)
			} else {
				logger.Debug().Msg("No updates to push")
			}
		// Handle the price updates from the stork websocket server
		case valueUpdate := <-storkWsCh:
			logger.Debug().Msgf("Received price update: %+v", valueUpdate)
			encoded, err := stringToByte32(string(valueUpdate.StorkSignedPrice.EncodedAssetId))
			if err != nil {
				logger.Error().Err(err).Msg("Failed to convert asset ID")
				continue
			}
			currentValue, ok := latestValueMap[encoded]

			if !ok {
				logger.Debug().Msgf("No current value for asset %s", valueUpdate.StorkSignedPrice.EncodedAssetId)
				updates[encoded] = valueUpdate
			} else {
				quantizedVal := new(big.Float)
				quantizedVal.SetString(string(valueUpdate.StorkSignedPrice.QuantizedPrice))

				quantizedCurrVal := new(big.Float)
				quantizedCurrVal.SetInt(currentValue.QuantizedValue)

				// Calculate the absolute difference
				difference := new(big.Float).Sub(quantizedVal, quantizedCurrVal)
				absDifference := new(big.Float).Abs(difference)

				// Calculate the ratio
				ratio := new(big.Float).Quo(absDifference, quantizedCurrVal)

				threshold := big.NewFloat(priceConfig.Assets[valueUpdate.AssetId].Threshold)
				if ratio.Cmp(threshold) > 0 {
					logger.Debug().Msgf("Percentage difference for asset %s is greater than %f", valueUpdate.StorkSignedPrice.EncodedAssetId, priceConfig.Assets[valueUpdate.AssetId].Threshold)
					updates[encoded] = valueUpdate
				} else if uint64(valueUpdate.Timestamp)-currentValue.TimestampNs > uint64(priceConfig.Assets[valueUpdate.AssetId].FallbackPeriodSecs)*1000000000 {
					logger.Debug().Msgf("Fallback period for asset %s has passed", valueUpdate.StorkSignedPrice.EncodedAssetId)
					updates[encoded] = valueUpdate
				}
			}
		// // Handle the contract events
		case chainUpdate := <-contractCh:
			for encodedAssetId, storkStructsTemporalNumericValue := range chainUpdate {
				latestValueMap[encodedAssetId] = storkStructsTemporalNumericValue
			}
		}
	}
}
