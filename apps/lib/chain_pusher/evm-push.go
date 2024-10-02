package chain_pusher

import (
	"math/big"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var EvmpushCmd = &cobra.Command{
	Use:   "evm",
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
	EvmpushCmd.Flags().StringP(StorkWebsocketEndpointFlag, "w", "", "Stork WebSocket endpoint")
	EvmpushCmd.Flags().StringP(StorkAuthCredentialsFlag, "a", "", "Stork auth credentials - base64(username:password)")
	EvmpushCmd.Flags().StringP(ChainRpcUrlFlag, "c", "", "Chain RPC URL")
	EvmpushCmd.Flags().StringP(ContractAddressFlag, "x", "", "Contract address")
	EvmpushCmd.Flags().StringP(AssetConfigFileFlag, "f", "", "Asset config file")
	EvmpushCmd.Flags().StringP(MnemonicFileFlag, "m", "", "Mnemonic file")
	EvmpushCmd.Flags().BoolP(VerifyPublishersFlag, "v", false, "Verify the publisher signed prices before pushing stork signed value to contract")
	EvmpushCmd.Flags().IntP(BatchingWindowFlag, "b", 5, "Batching window (seconds)")
	EvmpushCmd.Flags().IntP(PollingFrequencyFlag, "p", 3, "Asset Polling frequency (seconds)")

	EvmpushCmd.MarkFlagRequired(StorkWebsocketEndpointFlag)
	EvmpushCmd.MarkFlagRequired(StorkAuthCredentialsFlag)
	EvmpushCmd.MarkFlagRequired(ChainRpcUrlFlag)
	EvmpushCmd.MarkFlagRequired(ContractAddressFlag)
	EvmpushCmd.MarkFlagRequired(AssetConfigFileFlag)
	EvmpushCmd.MarkFlagRequired(MnemonicFileFlag)
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

	latestContractValueMap := make(map[InternalEncodedAssetId]StorkStructsTemporalNumericValue)
	latestStorkValueMap := make(map[InternalEncodedAssetId]AggregatedSignedPrice)

	initialValues, err := storkContractInterfacer.PullValues(encodedAssetIds)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to pull initial values from contract")
	}
	for encodedAssetId, value := range initialValues {
		latestContractValueMap[encodedAssetId] = value
	}
	logger.Info().Msgf("Pulled initial values for %d assets", len(initialValues))

	go storkContractInterfacer.ListenContractEvents(contractCh)
	go storkContractInterfacer.Poll(encodedAssetIds, contractCh)

	ticker := time.NewTicker(time.Duration(batchingWindow) * time.Second)
	defer ticker.Stop()

	for {
		select {
		// Determine updates after the batching window has passed and push them to the contract
		case <-ticker.C:
			updates := make(map[InternalEncodedAssetId]AggregatedSignedPrice)
			for encodedAssetId, latestStorkPrice := range latestStorkValueMap {
				latestValue, ok := latestContractValueMap[encodedAssetId]
				if !ok {
					logger.Debug().Msgf("No current value for asset %s", latestStorkPrice.StorkSignedPrice.EncodedAssetId)
					updates[encodedAssetId] = latestStorkPrice
				} else if shouldUpdate(
					latestValue,
					latestStorkPrice,
					priceConfig.Assets[latestStorkPrice.AssetId].FallbackPeriodSecs,
					priceConfig.Assets[latestStorkPrice.AssetId].PercentChangeThreshold,
				) {
					updates[encodedAssetId] = latestStorkPrice
				}
			}

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
			} else {
				logger.Debug().Msg("No updates to push")
			}
		// Handle stork updates
		case valueUpdate := <-storkWsCh:
			encoded, err := stringToByte32(string(valueUpdate.StorkSignedPrice.EncodedAssetId))
			if err != nil {
				logger.Error().Err(err).Msg("Failed to convert asset ID")
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

func shouldUpdate(latestValue StorkStructsTemporalNumericValue, latestStorkPrice AggregatedSignedPrice, fallbackPeriodSecs uint64, changeThreshold float64) bool {
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
