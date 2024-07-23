package main

import (
	"math/big"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var evmpushCmd = &cobra.Command{
	Use:   "evm-push",
	Short: "Push WebSocket prices to EVM contract",
	Run:   runEvmPush,
}

// required
const StorkWebsocketEndpointFlag = "stork-ws-endpoint"
const StorkAuthCredentialsFlag = "stork-auth-credentials"
const ChainRpcUrlFlag = "chain-rpc-url"
const ContractAddressFlag = "contract-address"
const AssetConfigFileFlag = "asset-config-file"
const MnemonicFileFlag = "mnemonic-file"

// optional
const VerifyPublishersFlag = "verify-publishers"
const BatchingWindowFlag = "batching-window"
const PollingFrequencyFlag = "polling-frequency"

func init() {
	evmpushCmd.Flags().StringP(StorkWebsocketEndpointFlag, "w", "", "Stork WebSocket endpoint")
	evmpushCmd.Flags().StringP(StorkAuthCredentialsFlag, "a", "", "Stork auth credentials")
	evmpushCmd.Flags().StringP(ChainRpcUrlFlag, "c", "", "Chain RPC URL")
	evmpushCmd.Flags().StringP(ContractAddressFlag, "x", "", "Contract address")
	evmpushCmd.Flags().StringP(AssetConfigFileFlag, "f", "", "Asset config file")
	evmpushCmd.Flags().StringP(MnemonicFileFlag, "m", "", "Mnemonic file")
	evmpushCmd.Flags().BoolP(VerifyPublishersFlag, "v", false, "Verify the contract")
	evmpushCmd.Flags().IntP(BatchingWindowFlag, "b", 10, "Batching window (seconds)")
	evmpushCmd.Flags().IntP(PollingFrequencyFlag, "p", 5, "Asset Polling frequency (seconds)")

	evmpushCmd.MarkFlagRequired(StorkWebsocketEndpointFlag)
	evmpushCmd.MarkFlagRequired(StorkAuthCredentialsFlag)
	evmpushCmd.MarkFlagRequired(ChainRpcUrlFlag)
	evmpushCmd.MarkFlagRequired(ContractAddressFlag)
	evmpushCmd.MarkFlagRequired(AssetConfigFileFlag)
	evmpushCmd.MarkFlagRequired(MnemonicFileFlag)
}

func runEvmPush(cmd *cobra.Command, args []string) {
	storkWsEndpoint, _ := cmd.Flags().GetString(StorkWebsocketEndpointFlag)
	storkAuth, _ := cmd.Flags().GetString(StorkAuthCredentialsFlag)
	chainRpcUrl, _ := cmd.Flags().GetString(ChainRpcUrlFlag)
	contractAddress, _ := cmd.Flags().GetString(ContractAddressFlag)
	assetConfigFile, _ := cmd.Flags().GetString(AssetConfigFileFlag)
	mnemonicFile, _ := cmd.Flags().GetString(MnemonicFileFlag)
	verifyPublishers, _ := cmd.Flags().GetBool(VerifyPublishersFlag)
	batchingWindow, _ := cmd.Flags().GetInt(BatchingWindowFlag)
	pollingFrequency, _ := cmd.Flags().GetInt(PollingFrequencyFlag)

	logger := EvmPusherLogger(chainRpcUrl, contractAddress)

	priceConfig, err := LoadConfig(assetConfigFile)
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
			logger.Fatal().Err(err).Msg("Failed to convert asset ID")
		}
		encodedAssetIds[i] = encoded
		i++
	}

	storkWsCh := make(chan AggregatedSignedPrice)
	contractCh := make(chan map[InternalEncodedAssetId]StorkStructsTemporalNumericValue)

	storkWs := NewStorkAggregatorWebsocketClient(storkWsEndpoint, storkAuth, assetIds, logger)
	go storkWs.Run(storkWsCh)

	storkContractInterfacer := NewStorkContractInterfacer(chainRpcUrl, contractAddress, mnemonicFile, pollingFrequency, verifyPublishers, logger)

	initialValues, err := storkContractInterfacer.PullValues(encodedAssetIds)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to pull initial values from contract")
	}
	latestContractValueMap := make(map[InternalEncodedAssetId]StorkStructsTemporalNumericValue)
	for encodedAssetId, value := range initialValues {
		latestContractValueMap[encodedAssetId] = value
	}
	logger.Info().Msgf("Pulled initial values for %d assets", len(initialValues))

	go storkContractInterfacer.ListenContractEvents(contractCh)
	go storkContractInterfacer.Poll(encodedAssetIds, contractCh)

	updates := make(map[InternalEncodedAssetId]AggregatedSignedPrice)

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
				// include this to prevent race conditions
				for encodedAssetId, update := range updates {
					quantizedValInt := new(big.Int)
					quantizedValInt.SetString(string(update.StorkSignedPrice.QuantizedPrice), 10)

					latestContractValueMap[encodedAssetId] = StorkStructsTemporalNumericValue{
						TimestampNs:    uint64(update.Timestamp),
						QuantizedValue: quantizedValInt,
					}
				}
				updates = make(map[InternalEncodedAssetId]AggregatedSignedPrice)
			} else {
				logger.Debug().Msg("No updates to push")
			}
		// Handle updates from the stork websocket server
		case valueUpdate := <-storkWsCh:
			logger.Debug().Msgf("Received price update: %+v", valueUpdate)
			encoded, err := stringToByte32(string(valueUpdate.StorkSignedPrice.EncodedAssetId))
			if err != nil {
				logger.Error().Err(err).Msg("Failed to convert asset ID")
				continue
			}
			currentValue, ok := latestContractValueMap[encoded]

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

				// Multiply by 100 to get the percentage
				percentChange := new(big.Float).Mul(ratio, big.NewFloat(100))

				threshold := big.NewFloat(priceConfig.Assets[valueUpdate.AssetId].Threshold)
				if percentChange.Cmp(threshold) > 0 {
					logger.Debug().Msgf("Percentage difference for asset %s is greater than %f", valueUpdate.StorkSignedPrice.EncodedAssetId, priceConfig.Assets[valueUpdate.AssetId].Threshold)
					updates[encoded] = valueUpdate
				} else if uint64(valueUpdate.Timestamp)-currentValue.TimestampNs > uint64(priceConfig.Assets[valueUpdate.AssetId].FallbackPeriodSecs)*1000000000 {
					logger.Debug().Msgf("Fallback period for asset %s has passed", valueUpdate.StorkSignedPrice.EncodedAssetId)
					updates[encoded] = valueUpdate
				}
			}
		// Handle contract updates
		case chainUpdate := <-contractCh:
			for encodedAssetId, storkStructsTemporalNumericValue := range chainUpdate {
				latestContractValueMap[encodedAssetId] = storkStructsTemporalNumericValue
			}
		}
	}
}
