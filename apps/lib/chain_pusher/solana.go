package chain_pusher

import (
	"context"
	"fmt"

	contract "github.com/Stork-Oracle/stork-external/apps/lib/chain_pusher/contract_bindings/solana"
	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gagliardetto/solana-go/rpc/ws"
	"github.com/rs/zerolog"
)

type SolanaContractInteracter struct {
	logger              zerolog.Logger
	client              *rpc.Client
	wsClient            *ws.Client
	contractAddr        solana.PublicKey
	feedAccounts        []solana.PublicKey
	pollingFrequencySec int
}

func NewSolanaContractInteracter(rpcUrl, wsUrl, contractAddr string, assetConfigFile string, pollingFreqSec int, logger zerolog.Logger) (*SolanaContractInteracter, error) {
	logger = logger.With().Str("component", "solana-contract-interactor").Logger()

	client := rpc.New(rpcUrl)
	wsClient, err := ws.Connect(context.Background(), wsUrl)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to connect to Solana WebSocket client")
		return nil, err
	}

	contractPubKey, err := solana.PublicKeyFromBase58(contractAddr)
	if err != nil {
		logger.Fatal().Err(err).Msg("Invalid contract address")
		return nil, err
	}

	assetConfig, err := LoadConfig(assetConfigFile)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to load asset config")
		return nil, err
	}

	feedAccounts := make([]solana.PublicKey, len(assetConfig.Assets))
	i := 0
	for _, asset := range assetConfig.Assets {
		encoded := []byte(asset.EncodedAssetId)

		//derive pda
		feedAccount, _, err := solana.FindProgramAddress(
			[][]byte{
				[]byte("stork-feed"),
				encoded,
			},
			contractPubKey,
		)
		if err != nil {
			logger.Fatal().Err(err).Msg("Failed to derive PDA for feed account")
			return nil, err
		}
		feedAccounts[i] = feedAccount
		i++
	}

	return &SolanaContractInteracter{
		logger:              logger,
		client:              client,
		wsClient:            wsClient,
		contractAddr:        contractPubKey,
		feedAccounts:        feedAccounts,
		pollingFrequencySec: pollingFreqSec,
	}, nil
}

func (sci *SolanaContractInteracter) ListenContractEvents(ch chan map[InternalEncodedAssetId]InternalStorkStructsTemporalNumericValue) {
	for _, feedAccount := range sci.feedAccounts {
		sub, err := sci.wsClient.AccountSubscribe(feedAccount, rpc.CommitmentFinalized)
		if err != nil {
			sci.logger.Error().Err(err).Str("account", feedAccount.String()).Msg("Failed to subscribe to feed account")
			continue
		}
		defer sub.Unsubscribe()

		go func(sub *ws.AccountSubscription, feedAccount solana.PublicKey) {
			for {
				msg, err := sub.Recv()
				if err != nil {
					sci.logger.Error().Err(err).Str("account", feedAccount.String()).Msg("Error receiving contract events")
					return
				}

				data := msg.Value.Data.GetBinary()

				decoder := bin.NewBorshDecoder(data)

				account := &contract.TemporalNumericValueFeedAccount{}

				err = account.UnmarshalWithDecoder(decoder)

				if err != nil {
					sci.logger.Error().Err(err).Msg("Error getting account from message")
					continue
				}

				latestValue := account.LatestValue
				tv := InternalStorkStructsTemporalNumericValue{
					QuantizedValue: latestValue.QuantizedValue.BigInt(),
					TimestampNs:    latestValue.TimestampNs,
				}

				ch <- map[InternalEncodedAssetId]InternalStorkStructsTemporalNumericValue{account.Id: tv}
			}
		}(sub, feedAccount)
	}

	// Wait indefinitely
	select {}
}

func (sci *SolanaContractInteracter) PullValues(encodedAssetIds []InternalEncodedAssetId) (map[InternalEncodedAssetId]InternalStorkStructsTemporalNumericValue, error) {
	// Implement logic to pull values from the Solana contract
	return nil, fmt.Errorf("PullValues not implemented")
}

func (sci *SolanaContractInteracter) BatchPushToContract(priceUpdates map[InternalEncodedAssetId]AggregatedSignedPrice) error {
	// Implement logic to push updates to the Solana contract
	return fmt.Errorf("BatchPushToContract not implemented")
}
