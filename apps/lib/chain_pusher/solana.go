package chain_pusher

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"
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

type SolanaContractInteractor struct {
	logger              zerolog.Logger
	client              *rpc.Client
	wsClient            *ws.Client
	contractAddr        solana.PublicKey
	feedAccounts        map[InternalEncodedAssetId]solana.PublicKey
	treasuryAccounts    map[uint8]solana.PublicKey
	configAccount       solana.PublicKey
	payer               solana.PrivateKey
	limiter             *rate.Limiter
	pollingFrequencySec int
	batchSize           int
	confirmationInChan  chan solana.Signature
}

// this is a limit imposed by the Solana blockchain and the size of the instruction
const MAX_BATCH_SIZE = 4

const NUM_CONFIRMATION_WORKERS = 10

func NewSolanaContractInteractor(
	rpcUrl string,
	wsUrl string,
	contractAddr string,
	payer []byte,
	assetConfigFile string,
	pollingFreqSec int, logger zerolog.Logger, limitPerSecond int, burstLimit int, batchSize int,
) (*SolanaContractInteractor, error) {
	logger = logger.With().Str("component", "solana-contract-interactor").Logger()

	if 0 < batchSize && batchSize < MAX_BATCH_SIZE {
		logger.Fatal().Msgf("Batch size must be between 1 and %d", MAX_BATCH_SIZE)
	}
	// calculate the time between requests bases on limitPerSecond
	timeBetweenRequestsMs := 1000 / limitPerSecond
	limiter := rate.NewLimiter(rate.Every(time.Duration(timeBetweenRequestsMs)*time.Millisecond), burstLimit)
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

	feedAccounts := make(map[InternalEncodedAssetId]solana.PublicKey)
	for _, asset := range assetConfig.Assets {
		encodedAssetIdBytes, err := hexStringToByteArray(string(asset.EncodedAssetId))
		if err != nil {
			logger.Fatal().Err(err).Str("assetId", fmt.Sprintf("%v", asset.AssetId)).Msg("Failed to convert encoded asset ID to bytes")
			return nil, err
		}
		// derive pda
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

		encodedAssetId := InternalEncodedAssetId(encodedAssetIdBytes)
		feedAccounts[encodedAssetId] = feedAccount
	}

	treasuryAccounts := make(map[uint8]solana.PublicKey)
	for i := 0; i < 256; i++ {
		treasuryAccount, _, err := solana.FindProgramAddress(
			[][]byte{[]byte("stork_treasury"), {uint8(i)}}, contractPubKey)
		if err != nil {
			logger.Fatal().Err(err).Uint8("treasuryId", uint8(i)).Msg("Failed to derive PDA for treasury account")
			return nil, err
		}
		treasuryAccounts[uint8(i)] = treasuryAccount
	}

	configAccount, _, err := solana.FindProgramAddress(
		[][]byte{[]byte("stork_config")}, contractPubKey)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to derive PDA for config account")
		return nil, err
	}

	confirmationInChan := make(chan solana.Signature)
	confirmationOutChan := make(chan solana.Signature)

	contract.SetProgramID(contractPubKey)
	sci := &SolanaContractInteractor{
		logger:              logger,
		client:              client,
		wsClient:            wsClient,
		contractAddr:        contractPubKey,
		feedAccounts:        feedAccounts,
		treasuryAccounts:    treasuryAccounts,
		configAccount:       configAccount,
		payer:               payer,
		limiter:             limiter,
		pollingFrequencySec: pollingFreqSec,
		batchSize:           batchSize,
		confirmationInChan:  confirmationInChan,
	}

	go sci.runUnboundedConfirmationBuffer(confirmationOutChan)
	sci.startConfirmationWorkers(confirmationOutChan, NUM_CONFIRMATION_WORKERS)

	return sci, nil
}

func (sci *SolanaContractInteractor) runUnboundedConfirmationBuffer(confirmationOutChan chan solana.Signature) {
	var queue []solana.Signature
	for {
		if len(queue) == 0 {
			sig := <-sci.confirmationInChan
			queue = append(queue, sig)
		}
		select {
		case sig := <-sci.confirmationInChan:
			queue = append(queue, sig)
		case confirmationOutChan <- queue[0]:
			queue = queue[1:]
		}
	}
}

func (sci *SolanaContractInteractor) startConfirmationWorkers(ch chan solana.Signature, numWorkers int) {
	for i := 0; i < numWorkers; i++ {
		go sci.confirmationWorker(ch)
	}
	sci.logger.Info().Int("numWorkers", numWorkers).Msg("Started confirmation workers")
}

func (sci *SolanaContractInteractor) confirmationWorker(ch chan solana.Signature) {
	for sig := range ch {
		_, err := confirm.WaitForConfirmation(context.Background(), sci.wsClient, sig, nil)
		if err != nil {
			sci.logger.Error().Str("signature", sig.String()).Err(err).Msg("failed to confirm transaction")
		} else {
			sci.logger.Debug().Str("signature", sig.String()).Msg("confirmed transaction")
		}
	}
}

func (sci *SolanaContractInteractor) ListenContractEvents(ctx context.Context, ch chan map[InternalEncodedAssetId]InternalTemporalNumericValue) {
	wg := sync.WaitGroup{}
	for _, feedAccount := range sci.feedAccounts {
		select {
		case <-ctx.Done():
			return
		default:
			wg.Add(1)
			go sci.listenSingleContractEvent(ctx, ch, feedAccount, &wg)
		}
	}
	// Wait indefinitely
	wg.Wait()
}

func (sci *SolanaContractInteractor) listenSingleContractEvent(ctx context.Context, ch chan map[InternalEncodedAssetId]InternalTemporalNumericValue, feedAccount solana.PublicKey, wg *sync.WaitGroup) {
	defer wg.Done()
	sub, err := sci.wsClient.AccountSubscribe(feedAccount, rpc.CommitmentFinalized)
	if err != nil {
		sci.logger.Error().Err(err).Str("account", feedAccount.String()).Msg("Failed to subscribe to feed account")
		return
	}
	defer sub.Unsubscribe()

	for {
		if ctx.Err() != nil {
			return
		}

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
		tv := InternalTemporalNumericValue{
			QuantizedValue: latestValue.QuantizedValue.BigInt(),
			TimestampNs:    latestValue.TimestampNs,
		}

		ch <- map[InternalEncodedAssetId]InternalTemporalNumericValue{account.Id: tv}
	}
}

func (sci *SolanaContractInteractor) PullValues(encodedAssetIds []InternalEncodedAssetId) (map[InternalEncodedAssetId]InternalTemporalNumericValue, error) {
	polledVals := make(map[InternalEncodedAssetId]InternalTemporalNumericValue)

	for _, encodedAssetId := range encodedAssetIds {

		feedAccount := sci.feedAccounts[encodedAssetId]
		accountInfo, err := sci.client.GetAccountInfo(context.Background(), feedAccount)
		if err != nil {
			sci.logger.Error().Err(err).Str("account", feedAccount.String()).Msg("Failed to get account info")
			continue
		}

		if accountInfo == nil || len(accountInfo.Value.Data.GetBinary()) == 0 {
			sci.logger.Debug().Str("assetId", hex.EncodeToString(encodedAssetId[:])).Msg("No value found")
			continue
		}

		decoder := bin.NewBorshDecoder(accountInfo.Value.Data.GetBinary())
		account := &contract.TemporalNumericValueFeedAccount{}
		err = account.UnmarshalWithDecoder(decoder)
		if err != nil {
			sci.logger.Error().Err(err).Str("account", feedAccount.String()).Msg("Failed to decode account data")
			continue
		}

		polledVals[encodedAssetId] = InternalTemporalNumericValue{
			QuantizedValue: account.LatestValue.QuantizedValue.BigInt(),
			TimestampNs:    account.LatestValue.TimestampNs,
		}
	}

	return polledVals, nil
}

func (sci *SolanaContractInteractor) BatchPushToContract(priceUpdates map[InternalEncodedAssetId]AggregatedSignedPrice) error {
	var wg sync.WaitGroup
	errChan := make(chan error, len(priceUpdates))
	sigChan := make(chan solana.Signature, len(priceUpdates))
	priceUpdatesBatches := sci.batchPriceUpdates(priceUpdates)
	for _, priceUpdateBatch := range priceUpdatesBatches {
		wg.Add(1)
		go func(priceUpdateBatch map[InternalEncodedAssetId]AggregatedSignedPrice) {
			defer wg.Done()
			err := sci.limiter.Wait(context.Background())
			if err != nil {
				errChan <- fmt.Errorf("rate limiter error: %w", err)
				return
			}
			sig, err := sci.pushLimitedBatchUpdateToContract(priceUpdateBatch)
			if err != nil {
				errChan <- fmt.Errorf("failed to push batch: %w", err)
			} else {
				sigChan <- sig
			}
		}(priceUpdateBatch)
	}

	wg.Wait()
	close(errChan)
	close(sigChan)
	// Collect any errors
	var errs []error
	for err := range errChan {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return fmt.Errorf("batch push encountered %d errors: %v", len(errs), errs)
	}

	sigs := []string{}
	for sig := range sigChan {
		sigs = append(sigs, sig.String())
	}

	sci.logger.Info().
		Int("numUpdates", len(priceUpdates)).
		Strs("batchTransactionSigs", sigs).
		Msg("Successfully pushed batch updates to contract")

	return nil
}

func (sci *SolanaContractInteractor) batchPriceUpdates(priceUpdates map[InternalEncodedAssetId]AggregatedSignedPrice) []map[InternalEncodedAssetId]AggregatedSignedPrice {
	priceUpdatesBatches := []map[InternalEncodedAssetId]AggregatedSignedPrice{}

	priceUpdatesBatch := make(map[InternalEncodedAssetId]AggregatedSignedPrice)
	i := 0
	for encodedAssetId, priceUpdate := range priceUpdates {
		priceUpdatesBatch[encodedAssetId] = priceUpdate
		i++
		if len(priceUpdatesBatch) == sci.batchSize || i == len(priceUpdates) {
			batchCopy := make(map[InternalEncodedAssetId]AggregatedSignedPrice, len(priceUpdatesBatch))
			for k, v := range priceUpdatesBatch {
				batchCopy[k] = v
			}
			priceUpdatesBatches = append(priceUpdatesBatches, batchCopy)
			priceUpdatesBatch = make(map[InternalEncodedAssetId]AggregatedSignedPrice)
		}
	}

	return priceUpdatesBatches
}

func (sci *SolanaContractInteractor) pushLimitedBatchUpdateToContract(priceUpdates map[InternalEncodedAssetId]AggregatedSignedPrice) (solana.Signature, error) {
	if len(priceUpdates) > MAX_BATCH_SIZE {
		return solana.Signature{}, fmt.Errorf("batch size exceeds limit, skipping update")
	}

	randomId, err := rand.Int(rand.Reader, big.NewInt(256))
	if err != nil {
		return solana.Signature{}, fmt.Errorf("failed to generate random ID: %w", err)
	}
	treasuryId := uint8(randomId.Uint64())

	treasuryAccount := sci.treasuryAccounts[treasuryId]
	instructions := []solana.Instruction{}
	assetIds := []string{}

	for encodedAssetId, priceUpdate := range priceUpdates {
		updateData, err := sci.priceUpdateToTemporalNumericValueEvmInput(priceUpdate, encodedAssetId, treasuryId)
		if err != nil {
			return solana.Signature{}, fmt.Errorf("failed to convert price update to TemporalNumericValueEvmInput: %w", err)
		}

		feedAccount := sci.feedAccounts[encodedAssetId]

		instruction, err := contract.NewUpdateTemporalNumericValueEvmInstruction(
			updateData,
			sci.configAccount,
			treasuryAccount,
			feedAccount,
			sci.payer.PublicKey(),
			solana.SystemProgramID,
		).ValidateAndBuild()
		if err != nil {
			return solana.Signature{}, fmt.Errorf("failed to build instruction: %w", err)
		}
		instructions = append(instructions, instruction)

	}

	recentBlockHash, err := sci.client.GetLatestBlockhash(context.Background(), rpc.CommitmentFinalized)
	if err != nil {
		return solana.Signature{}, fmt.Errorf("failed to get recent blockhash: %w", err)
	}

	tx, err := solana.NewTransaction(
		instructions,
		recentBlockHash.Value.Blockhash,
		solana.TransactionPayer(sci.payer.PublicKey()),
	)
	if err != nil {
		return solana.Signature{}, fmt.Errorf("failed to create transaction: %w", err)
	}

	_, err = tx.Sign(
		func(key solana.PublicKey) *solana.PrivateKey {
			if key == sci.payer.PublicKey() {
				return &sci.payer
			}
			return nil
		})
	if err != nil {
		return solana.Signature{}, fmt.Errorf("failed to sign transaction: %w", err)
	}

	sig, err := sci.client.SendTransaction(context.Background(), tx)
	if err != nil {
		return solana.Signature{}, fmt.Errorf("failed to send transaction: %w", err)
	}
	// check for confirmation without blocking
	sci.confirmationInChan <- sig

	sci.logger.Debug().
		Str("signature", sig.String()).
		Strs("assetIds", assetIds).
		Uint8("treasuryId", treasuryId).
		Msg("Pushed batch update to contract")

	return sig, nil
}

func (sci *SolanaContractInteractor) priceUpdateToTemporalNumericValueEvmInput(priceUpdate AggregatedSignedPrice, encodedAssetId InternalEncodedAssetId, treasuryId uint8) (contract.TemporalNumericValueEvmInput, error) {
	var assetId [32]uint8
	copy(assetId[:], encodedAssetId[:])

	quantizedPrice := sci.quantizedPriceToInt128(priceUpdate.StorkSignedPrice.QuantizedPrice)

	publisherMerkleRootBytes, err := hexStringToByteArray(priceUpdate.StorkSignedPrice.PublisherMerkleRoot)
	if err != nil {
		return contract.TemporalNumericValueEvmInput{}, fmt.Errorf("failed to convert PublisherMerkleRoot: %w", err)
	}
	var publisherMerkleRoot [32]uint8
	copy(publisherMerkleRoot[:], publisherMerkleRootBytes)

	valueComputeAlgHashBytes, err := hexStringToByteArray(priceUpdate.StorkSignedPrice.StorkCalculationAlg.Checksum)
	if err != nil {
		return contract.TemporalNumericValueEvmInput{}, fmt.Errorf("failed to convert ValueComputeAlgHash: %w", err)
	}
	var valueComputeAlgHash [32]uint8
	copy(valueComputeAlgHash[:], valueComputeAlgHashBytes)

	rBytes, err := hexStringToByteArray(priceUpdate.StorkSignedPrice.TimestampedSignature.Signature.R)
	if err != nil {
		return contract.TemporalNumericValueEvmInput{}, fmt.Errorf("failed to convert R: %w", err)
	}
	var r [32]uint8
	copy(r[:], rBytes)

	sBytes, err := hexStringToByteArray(priceUpdate.StorkSignedPrice.TimestampedSignature.Signature.S)
	if err != nil {
		return contract.TemporalNumericValueEvmInput{}, fmt.Errorf("failed to convert S: %w", err)
	}
	var s [32]uint8
	copy(s[:], sBytes)

	vBytes, err := hexStringToByteArray(priceUpdate.StorkSignedPrice.TimestampedSignature.Signature.V)
	if err != nil {
		return contract.TemporalNumericValueEvmInput{}, fmt.Errorf("failed to convert V: %w", err)
	}
	v := uint8(vBytes[0])

	return contract.TemporalNumericValueEvmInput{
		TemporalNumericValue: contract.TemporalNumericValue{
			TimestampNs:    uint64(priceUpdate.StorkSignedPrice.TimestampedSignature.Timestamp),
			QuantizedValue: quantizedPrice,
		},
		Id:                  assetId,
		PublisherMerkleRoot: publisherMerkleRoot,
		ValueComputeAlgHash: valueComputeAlgHash,
		R:                   r,
		S:                   s,
		V:                   v,
		TreasuryId:          treasuryId,
	}, nil
}

func (sci *SolanaContractInteractor) quantizedPriceToInt128(quantizedPrice QuantizedPrice) bin.Int128 {
	quantizedPriceBigInt := new(big.Int)
	quantizedPriceBigInt.SetString(string(quantizedPrice), 10)

	quantizedPrice128 := bin.Int128{
		Lo: quantizedPriceBigInt.Uint64(),
		Hi: quantizedPriceBigInt.Rsh(quantizedPriceBigInt, 64).Uint64(),
	}

	return quantizedPrice128
}
