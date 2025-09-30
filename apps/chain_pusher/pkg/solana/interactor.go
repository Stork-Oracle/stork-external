package solana

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/Stork-Oracle/stork-external/apps/chain_pusher/pkg/pusher"
	"github.com/Stork-Oracle/stork-external/apps/chain_pusher/pkg/solana/bindings"
	"github.com/Stork-Oracle/stork-external/apps/chain_pusher/pkg/types"
	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	confirm "github.com/gagliardetto/solana-go/rpc/sendAndConfirmTransaction"
	"github.com/gagliardetto/solana-go/rpc/ws"
	"github.com/rs/zerolog"
	"golang.org/x/time/rate"
)

// NumTreasuryAccounts is the number of treasury accounts to use.
const NumTreasuryAccounts = 256

var ErrBatchSizeExceedsLimit = errors.New("batch size exceeds limit, skipping update")

type ContractInteractor struct {
	logger             zerolog.Logger
	client             *rpc.Client
	wsClient           *ws.Client
	contractAddr       solana.PublicKey
	feedAccounts       map[types.InternalEncodedAssetID]solana.PublicKey
	treasuryAccounts   map[uint8]solana.PublicKey
	configAccount      solana.PublicKey
	payer              solana.PrivateKey
	limiter            *rate.Limiter
	pollingPeriodSec   int
	batchSize          int
	confirmationInChan chan solana.Signature
}

// MaxBatchSize is a limit imposed by the Solana blockchain and the size our update instruction.
const MaxBatchSize = 4

// NumConfirmationWorkers is the number of confirmation workers to run.
const NumConfirmationWorkers = 10

func NewContractInteractor(
	contractAddr string,
	payer []byte,
	assetConfigFile string,
	pollingPeriodSec int, logger zerolog.Logger, limitPerSecond int, burstLimit int, batchSize int,
) (*ContractInteractor, error) {
	logger = logger.With().Str("component", "solana-contract-interactor").Logger()

	if 0 < batchSize && batchSize < MaxBatchSize {
		logger.Fatal().Msgf("Batch size must be between 1 and %d", MaxBatchSize)
	}
	// calculate the time between requests bases on limitPerSecond
	timeBetweenRequests := time.Second / time.Duration(limitPerSecond)
	limiter := rate.NewLimiter(rate.Every(timeBetweenRequests), burstLimit)

	contractPubKey, err := solana.PublicKeyFromBase58(contractAddr)
	if err != nil {
		logger.Fatal().Err(err).Msg("Invalid contract address")

		return nil, fmt.Errorf("invalid contract address: %w", err)
	}

	assetConfig, err := types.LoadConfig(assetConfigFile)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to load asset config")

		return nil, fmt.Errorf("failed to load asset config: %w", err)
	}

	feedAccounts, err := getFeedAccountsFromAssets(assetConfig.Assets, contractPubKey, logger)
	if err != nil {
		return nil, err
	}

	treasuryAccounts, err := getTreasuryAccounts(contractPubKey, logger)
	if err != nil {
		return nil, err
	}

	configAccount, _, err := solana.FindProgramAddress(
		[][]byte{[]byte("stork_config")}, contractPubKey)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to derive PDA for config account")

		return nil, fmt.Errorf("failed to derive PDA for config account: %w", err)
	}

	confirmationInChan := make(chan solana.Signature)
	confirmationOutChan := make(chan solana.Signature)

	bindings.SetProgramID(contractPubKey)
	sci := &ContractInteractor{
		logger:             logger,
		client:             nil,
		wsClient:           nil,
		contractAddr:       contractPubKey,
		feedAccounts:       feedAccounts,
		treasuryAccounts:   treasuryAccounts,
		configAccount:      configAccount,
		payer:              payer,
		limiter:            limiter,
		pollingPeriodSec:   pollingPeriodSec,
		batchSize:          batchSize,
		confirmationInChan: confirmationInChan,
	}

	go sci.runUnboundedConfirmationBuffer(confirmationOutChan)

	sci.startConfirmationWorkers(confirmationOutChan, NumConfirmationWorkers)

	return sci, nil
}

func (sci *ContractInteractor) ConnectHTTP(url string) error {
	client := rpc.New(url)
	sci.client = client
	return nil
}

func (sci *ContractInteractor) ConnectWs(url string) error {
	wsClient, err := ws.Connect(context.Background(), url)
	if err != nil {
		return fmt.Errorf("failed to connect to Solana WebSocket client: %w", err)
	}

	sci.wsClient = wsClient

	return nil
}

func (sci *ContractInteractor) ListenContractEvents(
	ctx context.Context,
	ch chan map[types.InternalEncodedAssetID]types.InternalTemporalNumericValue,
) {
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

func (sci *ContractInteractor) PullValues(
	encodedAssetIDs []types.InternalEncodedAssetID,
) (map[types.InternalEncodedAssetID]types.InternalTemporalNumericValue, error) {
	polledVals := make(map[types.InternalEncodedAssetID]types.InternalTemporalNumericValue)

	for _, encodedAssetID := range encodedAssetIDs {
		feedAccount := sci.feedAccounts[encodedAssetID]

		accountInfo, err := sci.client.GetAccountInfo(context.Background(), feedAccount)
		if err != nil {
			sci.logger.Error().Err(err).Str("account", feedAccount.String()).Msg("Failed to get account info")

			continue
		}

		if accountInfo == nil || len(accountInfo.Value.Data.GetBinary()) == 0 {
			sci.logger.Debug().Str("assetID", hex.EncodeToString(encodedAssetID[:])).Msg("No value found")

			continue
		}

		decoder := bin.NewBorshDecoder(accountInfo.Value.Data.GetBinary())
		account := &bindings.TemporalNumericValueFeedAccount{}

		err = account.UnmarshalWithDecoder(decoder)
		if err != nil {
			sci.logger.Error().Err(err).Str("account", feedAccount.String()).Msg("Failed to decode account data")

			continue
		}

		polledVals[encodedAssetID] = types.InternalTemporalNumericValue{
			QuantizedValue: account.LatestValue.QuantizedValue.BigInt(),
			TimestampNs:    account.LatestValue.TimestampNs,
		}
	}

	return polledVals, nil
}

func (sci *ContractInteractor) BatchPushToContract(
	priceUpdates map[types.InternalEncodedAssetID]types.AggregatedSignedPrice,
) error {
	var wg sync.WaitGroup

	errChan := make(chan error, len(priceUpdates))
	sigChan := make(chan solana.Signature, len(priceUpdates))

	priceUpdatesBatches := sci.batchPriceUpdates(priceUpdates)
	for _, priceUpdateBatch := range priceUpdatesBatches {
		wg.Add(1)

		go func(priceUpdateBatch map[types.InternalEncodedAssetID]types.AggregatedSignedPrice) {
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
	errs := make([]error, 0, len(priceUpdates))
	for err := range errChan {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		//nolint:err113 // This is essentially wrapping the errors
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

// GetWalletBalance is a placeholder function to get the balance of the wallet being used to push to the contract.
// todo: implement
//
//nolint:godox // This function has unmet criteria to be implemented.
func (sci *ContractInteractor) GetWalletBalance() (float64, error) {
	return -1, nil
}

func getFeedAccountsFromAssets(
	assets map[types.AssetID]types.AssetEntry,
	contractPubKey solana.PublicKey,
	logger zerolog.Logger,
) (map[types.InternalEncodedAssetID]solana.PublicKey, error) {
	feedAccounts := make(map[types.InternalEncodedAssetID]solana.PublicKey)

	var (
		encodedAssetIDBytes []byte
		feedAccount         solana.PublicKey
		err                 error
	)

	for _, asset := range assets {
		encodedAssetIDBytes, err = pusher.HexStringToByteArray(string(asset.EncodedAssetID))
		if err != nil {
			logger.Fatal().
				Err(err).
				Str("assetID", fmt.Sprintf("%v", asset.AssetID)).
				Msg("Failed to convert encoded asset ID to bytes")

			return nil, fmt.Errorf("failed to convert encoded asset ID to bytes: %w", err)
		}
		// derive pda
		feedAccount, _, err = solana.FindProgramAddress(
			[][]byte{
				[]byte("stork_feed"),
				encodedAssetIDBytes,
			},
			contractPubKey,
		)
		if err != nil {
			logger.Fatal().
				Err(err).
				Str("assetID", fmt.Sprintf("%v", asset.AssetID)).
				Msg("Failed to derive PDA for feed account")

			return nil, fmt.Errorf("failed to derive PDA for feed account: %w", err)
		}

		encodedAssetID := types.InternalEncodedAssetID(encodedAssetIDBytes)
		feedAccounts[encodedAssetID] = feedAccount
	}

	return feedAccounts, nil
}

func getTreasuryAccounts(
	contractPubKey solana.PublicKey,
	logger zerolog.Logger,
) (map[uint8]solana.PublicKey, error) {
	treasuryAccounts := make(map[uint8]solana.PublicKey)

	for i := range NumTreasuryAccounts {
		//nolint:gosec // "i" is clearly constrained to uint8 range
		uint8i := uint8(i)

		var (
			treasuryAccount solana.PublicKey
			err             error
		)

		treasuryAccount, _, err = solana.FindProgramAddress(
			[][]byte{[]byte("stork_treasury"), {uint8i}}, contractPubKey)
		if err != nil {
			logger.Fatal().Err(err).Uint8("treasuryID", uint8i).Msg("Failed to derive PDA for treasury account")

			return nil, fmt.Errorf("failed to derive PDA for treasury account: %w", err)
		}

		treasuryAccounts[uint8i] = treasuryAccount
	}

	return treasuryAccounts, nil
}

func (sci *ContractInteractor) runUnboundedConfirmationBuffer(confirmationOutChan chan solana.Signature) {
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

func (sci *ContractInteractor) startConfirmationWorkers(ch chan solana.Signature, numWorkers int) {
	for range numWorkers {
		go sci.confirmationWorker(ch)
	}

	sci.logger.Info().Int("numWorkers", numWorkers).Msg("Started confirmation workers")
}

func (sci *ContractInteractor) confirmationWorker(ch chan solana.Signature) {
	for sig := range ch {
		_, err := confirm.WaitForConfirmation(context.Background(), sci.wsClient, sig, nil)
		if err != nil {
			sci.logger.Error().Str("signature", sig.String()).Err(err).Msg("failed to confirm transaction")
		} else {
			sci.logger.Debug().Str("signature", sig.String()).Msg("confirmed transaction")
		}
	}
}

func (sci *ContractInteractor) listenSingleContractEvent(
	ctx context.Context,
	ch chan map[types.InternalEncodedAssetID]types.InternalTemporalNumericValue,
	feedAccount solana.PublicKey,
	wg *sync.WaitGroup,
) {
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

		var msg *ws.AccountResult

		msg, err = sub.Recv()
		if err != nil {
			sci.logger.Error().Err(err).Str("account", feedAccount.String()).Msg("Error receiving contract events")

			return
		}

		data := msg.Value.Data.GetBinary()

		decoder := bin.NewBorshDecoder(data)

		account := &bindings.TemporalNumericValueFeedAccount{}

		err = account.UnmarshalWithDecoder(decoder)
		if err != nil {
			sci.logger.Error().Err(err).Msg("Error getting account from message")

			continue
		}

		latestValue := account.LatestValue
		tv := types.InternalTemporalNumericValue{
			QuantizedValue: latestValue.QuantizedValue.BigInt(),
			TimestampNs:    latestValue.TimestampNs,
		}

		ch <- map[types.InternalEncodedAssetID]types.InternalTemporalNumericValue{account.Id: tv}
	}
}

func (sci *ContractInteractor) batchPriceUpdates(
	priceUpdates map[types.InternalEncodedAssetID]types.AggregatedSignedPrice,
) []map[types.InternalEncodedAssetID]types.AggregatedSignedPrice {
	priceUpdatesBatches := []map[types.InternalEncodedAssetID]types.AggregatedSignedPrice{}

	priceUpdatesBatch := make(map[types.InternalEncodedAssetID]types.AggregatedSignedPrice)
	i := 0

	for encodedAssetID, priceUpdate := range priceUpdates {
		priceUpdatesBatch[encodedAssetID] = priceUpdate

		i++
		if len(priceUpdatesBatch) == sci.batchSize || i == len(priceUpdates) {
			batchCopy := make(map[types.InternalEncodedAssetID]types.AggregatedSignedPrice, len(priceUpdatesBatch))
			for k, v := range priceUpdatesBatch {
				batchCopy[k] = v
			}

			priceUpdatesBatches = append(priceUpdatesBatches, batchCopy)
			priceUpdatesBatch = make(map[types.InternalEncodedAssetID]types.AggregatedSignedPrice)
		}
	}

	return priceUpdatesBatches
}

func (sci *ContractInteractor) pushLimitedBatchUpdateToContract(
	priceUpdates map[types.InternalEncodedAssetID]types.AggregatedSignedPrice,
) (solana.Signature, error) {
	if len(priceUpdates) > MaxBatchSize {
		return solana.Signature{}, ErrBatchSizeExceedsLimit
	}

	randomID, err := rand.Int(rand.Reader, big.NewInt(NumTreasuryAccounts))
	if err != nil {
		return solana.Signature{}, fmt.Errorf("failed to generate random ID: %w", err)
	}

	//nolint:gosec // "randomID" is clearly constrained to uint8 range
	treasuryID := uint8(randomID.Uint64())

	treasuryAccount := sci.treasuryAccounts[treasuryID]
	instructions := []solana.Instruction{}
	assetIDs := []string{}

	var updateData bindings.TemporalNumericValueEvmInput
	for encodedAssetID, priceUpdate := range priceUpdates {
		updateData, err = sci.priceUpdateToTemporalNumericValueEvmInput(priceUpdate, treasuryID)
		if err != nil {
			return solana.Signature{}, fmt.Errorf(
				"failed to convert price update to TemporalNumericValueEvmInput: %w",
				err,
			)
		}

		feedAccount := sci.feedAccounts[encodedAssetID]

		var instruction *bindings.Instruction

		instruction, err = bindings.NewUpdateTemporalNumericValueEvmInstruction(
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
		Strs("assetIDs", assetIDs).
		Uint8("treasuryID", treasuryID).
		Msg("Pushed batch update to contract")

	return sig, nil
}

func (sci *ContractInteractor) priceUpdateToTemporalNumericValueEvmInput(
	priceUpdate types.AggregatedSignedPrice,
	treasuryID uint8,
) (bindings.TemporalNumericValueEvmInput, error) {
	var assetID [32]uint8

	encodedAssetIDBytes, err := pusher.HexStringToByte32(string(priceUpdate.StorkSignedPrice.EncodedAssetID))
	if err != nil {
		return bindings.TemporalNumericValueEvmInput{}, fmt.Errorf("failed to convert EncodedAssetID: %w", err)
	}

	copy(assetID[:], encodedAssetIDBytes[:])

	quantizedPrice := quantizedPriceToInt128(priceUpdate.StorkSignedPrice.QuantizedPrice)

	publisherMerkleRootBytes, err := pusher.HexStringToByteArray(priceUpdate.StorkSignedPrice.PublisherMerkleRoot)
	if err != nil {
		return bindings.TemporalNumericValueEvmInput{}, fmt.Errorf("failed to convert PublisherMerkleRoot: %w", err)
	}

	var publisherMerkleRoot [32]uint8
	copy(publisherMerkleRoot[:], publisherMerkleRootBytes)

	valueComputeAlgHashBytes, err := pusher.HexStringToByteArray(
		priceUpdate.StorkSignedPrice.StorkCalculationAlg.Checksum,
	)
	if err != nil {
		return bindings.TemporalNumericValueEvmInput{}, fmt.Errorf("failed to convert ValueComputeAlgHash: %w", err)
	}

	var valueComputeAlgHash [32]uint8
	copy(valueComputeAlgHash[:], valueComputeAlgHashBytes)

	rBytes, err := pusher.HexStringToByteArray(priceUpdate.StorkSignedPrice.TimestampedSignature.Signature.R)
	if err != nil {
		return bindings.TemporalNumericValueEvmInput{}, fmt.Errorf("failed to convert R: %w", err)
	}

	//nolint:varnamelen // "r" is a valid variable name in the context of rsv signature
	var r [32]uint8
	copy(r[:], rBytes)

	sBytes, err := pusher.HexStringToByteArray(priceUpdate.StorkSignedPrice.TimestampedSignature.Signature.S)
	if err != nil {
		return bindings.TemporalNumericValueEvmInput{}, fmt.Errorf("failed to convert S: %w", err)
	}

	var s [32]uint8
	copy(s[:], sBytes)

	vBytes, err := pusher.HexStringToByteArray(priceUpdate.StorkSignedPrice.TimestampedSignature.Signature.V)
	if err != nil {
		return bindings.TemporalNumericValueEvmInput{}, fmt.Errorf("failed to convert V: %w", err)
	}

	v := vBytes[0]

	return bindings.TemporalNumericValueEvmInput{
		TemporalNumericValue: bindings.TemporalNumericValue{
			TimestampNs:    priceUpdate.StorkSignedPrice.TimestampedSignature.TimestampNano,
			QuantizedValue: quantizedPrice,
		},
		Id:                  assetID,
		PublisherMerkleRoot: publisherMerkleRoot,
		ValueComputeAlgHash: valueComputeAlgHash,
		R:                   r,
		S:                   s,
		V:                   v,
		TreasuryId:          treasuryID,
	}, nil
}

//nolint:mnd // twos compliment conversions contain magic numbers
func quantizedPriceToInt128(quantizedPrice types.QuantizedPrice) bin.Int128 {
	quantizedPriceBigInt := new(big.Int)
	quantizedPriceBigInt.SetString(string(quantizedPrice), 10)

	// Handle two's complement for signed 128-bit representation
	if quantizedPriceBigInt.Sign() >= 0 {
		quantizedPrice128 := bin.Int128{
			Lo: quantizedPriceBigInt.Uint64(),
			Hi: new(big.Int).Rsh(quantizedPriceBigInt, 64).Uint64(),
		}

		return quantizedPrice128
	} else {
		maxUint128 := new(big.Int).Lsh(big.NewInt(1), 128)
		maxUint128.Sub(maxUint128, big.NewInt(1))

		absValue := new(big.Int).Abs(quantizedPriceBigInt)

		twosComplement := new(big.Int).Sub(maxUint128, absValue)
		twosComplement.Add(twosComplement, big.NewInt(1))

		quantizedPrice128 := bin.Int128{
			Lo: twosComplement.Uint64(),
			Hi: new(big.Int).Rsh(twosComplement, 64).Uint64(),
		}

		return quantizedPrice128
	}
}
