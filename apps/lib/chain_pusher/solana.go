package chain_pusher

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"

	contract "github.com/Stork-Oracle/stork-external/apps/lib/chain_pusher/contract_bindings/solana"
	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	confirm "github.com/gagliardetto/solana-go/rpc/sendAndConfirmTransaction"
	"github.com/gagliardetto/solana-go/rpc/ws"
	"github.com/rs/zerolog"
	"golang.org/x/time/rate"
)

type SolanaContractInteracter struct {
	logger              zerolog.Logger
	client              *rpc.Client
	wsClient            *ws.Client
	contractAddr        solana.PublicKey
	feedAccounts        []solana.PublicKey
	payer               solana.PrivateKey
	pollingFrequencySec int
}

func NewSolanaContractInteracter(rpcUrl, wsUrl, contractAddr string, privateKeyFile string, assetConfigFile string, pollingFreqSec int, logger zerolog.Logger) (*SolanaContractInteracter, error) {
	logger = logger.With().Str("component", "solana-contract-interactor").Logger()

	client := rpc.New(rpcUrl)
	wsClient, err := ws.Connect(context.Background(), wsUrl)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to connect to Solana WebSocket client")
		return nil, err
	}

	contractPubKey := solana.MustPublicKeyFromBase58(contractAddr)
	if err != nil {
		logger.Fatal().Err(err).Msg("Invalid contract address")
		return nil, err
	}

	payer, err := solana.PrivateKeyFromSolanaKeygenFile(privateKeyFile)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to parse private key")
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
		encodedAssetIdBytes, err := hexStringToByteArray(string(asset.EncodedAssetId))
		if err != nil {
			logger.Fatal().Err(err).Str("assetId", fmt.Sprintf("%v", asset.AssetId)).Msg("Failed to convert encoded asset ID to bytes")
			return nil, err
		}
		//derive pda
		feedAccount, _, err := solana.FindProgramAddress(
			[][]byte{
				[]byte("stork_feed"),
				encodedAssetIdBytes,
			},
			contractPubKey,
		)
		if err != nil {
			logger.Fatal().Err(err).Str("assetId", fmt.Sprintf("%v", asset.AssetId)).Msg("Failed to derive PDA for feed account")
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
		payer:               payer,
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
	polledVals := make(map[InternalEncodedAssetId]InternalStorkStructsTemporalNumericValue)

	for _, encodedAssetId := range encodedAssetIds {
		// Derive the PDA for this asset
		feedAccount, _, err := solana.FindProgramAddress(
			[][]byte{
				[]byte("stork_feed"),
				encodedAssetId[:],
			},
			sci.contractAddr,
		)
		if err != nil {
			sci.logger.Error().Err(err).Str("assetId", hex.EncodeToString(encodedAssetId[:])).Msg("Failed to derive PDA for feed account in PullValues")
			continue
		}

		// Fetch the account data
		accountInfo, err := sci.client.GetAccountInfo(context.Background(), feedAccount)
		if err != nil {
			sci.logger.Error().Err(err).Str("account", feedAccount.String()).Msg("Failed to get account info")
			continue
		}

		if accountInfo == nil || len(accountInfo.Value.Data.GetBinary()) == 0 {
			sci.logger.Debug().Str("assetId", hex.EncodeToString(encodedAssetId[:])).Msg("No value found")
			continue
		}

		// Decode the account data
		decoder := bin.NewBorshDecoder(accountInfo.Value.Data.GetBinary())
		account := &contract.TemporalNumericValueFeedAccount{}
		err = account.UnmarshalWithDecoder(decoder)
		if err != nil {
			sci.logger.Error().Err(err).Str("account", feedAccount.String()).Msg("Failed to decode account data")
			continue
		}

		// Convert to internal format
		polledVals[encodedAssetId] = InternalStorkStructsTemporalNumericValue{
			QuantizedValue: account.LatestValue.QuantizedValue.BigInt(),
			TimestampNs:    account.LatestValue.TimestampNs,
		}
	}

	return polledVals, nil
}

func (sci *SolanaContractInteracter) BatchPushToContract(priceUpdates map[InternalEncodedAssetId]AggregatedSignedPrice) error {
	// Create a rate limiter: 40 requests per 10 seconds (4 per second)
	// with a burst of 10 to stay well within the limits
	limiter := rate.NewLimiter(rate.Every(250*time.Millisecond), 10)

	var wg sync.WaitGroup
	errChan := make(chan error, len(priceUpdates))

	for encodedAssetId, priceUpdate := range priceUpdates {
		wg.Add(1)
		go func(encodedAssetId InternalEncodedAssetId, priceUpdate AggregatedSignedPrice) {
			defer wg.Done()

			// Wait for rate limiter
			err := limiter.Wait(context.Background())
			if err != nil {
				errChan <- fmt.Errorf("rate limiter error: %w", err)
				return
			}

			err = sci.pushSingleUpdateToContract(encodedAssetId, priceUpdate)
			if err != nil {
				errChan <- fmt.Errorf("failed to push update for asset %x: %w", encodedAssetId, err)
			}
		}(encodedAssetId, priceUpdate)
	}

	// Wait for all goroutines to finish
	wg.Wait()
	close(errChan)

	// Collect any errors
	var errs []error
	for err := range errChan {
		errs = append(errs, err)
	}

	// If there were any errors, return them combined
	if len(errs) > 0 {
		return fmt.Errorf("batch push encountered %d errors: %v", len(errs), errs)
	}

	sci.logger.Info().
		Int("numUpdates", len(priceUpdates)).
		Msg("Successfully pushed batch updates to contract")

	return nil
}

func (sci *SolanaContractInteracter) pushSingleUpdateToContract(encodedAssetId InternalEncodedAssetId, priceUpdate AggregatedSignedPrice) error {
	var assetId [32]uint8
	copy(assetId[:], encodedAssetId[:])

	// Derive the PDA for the feed account
	feedAccount, _, err := solana.FindProgramAddress(
		[][]byte{
			[]byte("stork_feed"),
			encodedAssetId[:],
		},
		sci.contractAddr,
	)
	if err != nil {
		return fmt.Errorf("failed to derive PDA for feed account: %w", err)
	}

	// Randomly select a treasury ID (0-255)
	randomId, err := rand.Int(rand.Reader, big.NewInt(256))
	if err != nil {
		return fmt.Errorf("failed to generate random treasury ID: %w", err)
	}
	treasuryId := uint8(randomId.Uint64())

	// Derive the PDA for the selected treasury account
	treasuryAccount, _, err := solana.FindProgramAddress(
		[][]byte{
			[]byte("stork_treasury"),
			[]byte{treasuryId},
		},
		sci.contractAddr,
	)
	if err != nil {
		return fmt.Errorf("failed to derive PDA for treasury account: %w", err)
	}

	// Convert the quantized price to bin.Int128
	quantizedPriceBigInt := new(big.Int)
	quantizedPriceBigInt.SetString(string(priceUpdate.StorkSignedPrice.QuantizedPrice), 10)

	quantizedPrice := bin.Int128{
		Lo: quantizedPriceBigInt.Uint64(),
		Hi: quantizedPriceBigInt.Rsh(quantizedPriceBigInt, 64).Uint64(),
	}

	// Convert strings to [32]byte using stringToByte32
	publisherMerkleRoot, err := stringToByte32(priceUpdate.StorkSignedPrice.PublisherMerkleRoot)
	if err != nil {
		return fmt.Errorf("failed to convert PublisherMerkleRoot: %w", err)
	}

	valueComputeAlgHash, err := stringToByte32(priceUpdate.StorkSignedPrice.StorkCalculationAlg.Checksum)
	if err != nil {
		return fmt.Errorf("failed to convert ValueComputeAlgHash: %w", err)
	}

	r, err := stringToByte32(priceUpdate.StorkSignedPrice.TimestampedSignature.Signature.R)
	if err != nil {
		return fmt.Errorf("failed to convert R: %w", err)
	}

	s, err := stringToByte32(priceUpdate.StorkSignedPrice.TimestampedSignature.Signature.S)
	if err != nil {
		return fmt.Errorf("failed to convert S: %w", err)
	}

	// Create the update instruction
	updateData := contract.TemporalNumericValueEvmInput{
		TemporalNumericValue: contract.TemporalNumericValue{
			TimestampNs:    uint64(priceUpdate.StorkSignedPrice.TimestampedSignature.Timestamp),
			QuantizedValue: quantizedPrice,
		},
		Id:                  assetId,
		PublisherMerkleRoot: publisherMerkleRoot,
		ValueComputeAlgHash: valueComputeAlgHash,
		R:                   r,
		S:                   s,
		V:                   uint8(priceUpdate.StorkSignedPrice.TimestampedSignature.Signature.V[0]),
		TreasuryId:          treasuryId,
	}

	configAccount, _, err := solana.FindProgramAddress(
		[][]byte{
			[]byte("stork_config"),
		},
		sci.contractAddr,
	)
	if err != nil {
		return fmt.Errorf("failed to derive PDA for config account: %w", err)
	}

	instruction, err := contract.NewUpdateTemporalNumericValueEvmInstruction(
		updateData,
		configAccount,
		treasuryAccount,
		feedAccount,
		sci.payer.PublicKey(),
		solana.SystemProgramID,
	).ValidateAndBuild()
	if err != nil {
		return fmt.Errorf("failed to build instruction: %w", err)
	}

	recentBlockHash, err := sci.client.GetRecentBlockhash(context.Background(), rpc.CommitmentFinalized)
	if err != nil {
		return fmt.Errorf("failed to get recent blockhash: %w", err)
	}

	// Create and send the transaction
	tx, err := solana.NewTransaction(
		[]solana.Instruction{instruction},
		recentBlockHash.Value.Blockhash,
		solana.TransactionPayer(sci.payer.PublicKey()),
	)
	if err != nil {
		return fmt.Errorf("failed to create transaction: %w", err)
	}

	_, err = tx.Sign(
		func(key solana.PublicKey) *solana.PrivateKey {
			if key == sci.payer.PublicKey() {
				return &sci.payer
			}
			return nil
		})
	if err != nil {
		return fmt.Errorf("failed to sign transaction: %w", err)
	}

	sig, err := confirm.SendAndConfirmTransaction(
		context.Background(),
		sci.client,
		sci.wsClient,
		tx,
	)

	if err != nil {
		return fmt.Errorf("failed to send and confirm transaction: %w", err)
	}

	sci.logger.Info().
		Str("signature", sig.String()).
		Str("assetId", hex.EncodeToString(encodedAssetId[:])).
		Uint8("treasuryId", treasuryId).
		Msg("Pushed new value to contract")

	return nil
}

func hexStringToByteArray(hexString string) ([]byte, error) {
	// Remove "0x" prefix if present
	hexString = strings.TrimPrefix(hexString, "0x")
	return hex.DecodeString(hexString)
}
