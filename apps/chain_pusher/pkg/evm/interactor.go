package evm

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/Masterminds/semver"
	"github.com/Stork-Oracle/stork-external/apps/chain_pusher/pkg/evm/bindings"
	"github.com/Stork-Oracle/stork-external/apps/chain_pusher/pkg/pusher"
	"github.com/Stork-Oracle/stork-external/apps/chain_pusher/pkg/types"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	etherrors "github.com/ethereum/go-ethereum/core/txpool"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rs/zerolog"
)

var (
	ErrCastingPublicKeyToECDSA = errors.New("error casting public key to ECDSA")
	ErrMaxRetryAttemptsReached = errors.New("max retry attempts reached")
	ErrEventChannelClosed      = errors.New("event channel is closed")
)

type ContractInteractor struct {
	logger zerolog.Logger

	contractAddress common.Address
	privateKey      *ecdsa.PrivateKey
	gasLimit        uint64
	nonceManager    NonceManagerI

	contract        *bindings.StorkContract
	wsContract      *bindings.StorkContract
	client          *ethclient.Client
	useSyncSend     bool
	version         *semver.Version
	gasFeeCap       *big.Int
	gasTipCap       *big.Int
	singleUpdateFee *big.Int
	lastSetGasCaps  time.Time

	chainID *big.Int

	verifyPublishers bool
}

const (
	// 1 * (1.5 ^ 10) = 57.66 seconds (last attempt delay).
	maxRetryAttempts         = 10
	initialBackoff           = 1 * time.Second
	exponentialBackoffFactor = 1.5
	gasCalcResetInterval     = 5 * time.Minute
	maxTransactionAttempts   = 3
	gasBumpNumerator         = 120
	gasBumpDenominator       = 100
	gasLimitMultiplier       = 1.1
)

func NewContractInteractor(
	contractAddr string,
	keyFileContent []byte,
	nonceManager NonceManagerI,
	verifyPublishers bool,
	logger zerolog.Logger,
	gasLimit uint64,
	useSyncSend bool,
) (*ContractInteractor, error) {
	privateKey, err := loadPrivateKey(keyFileContent)
	if err != nil {
		return nil, err
	}

	return &ContractInteractor{
		logger: logger,

		contractAddress: common.HexToAddress(contractAddr),
		privateKey:      privateKey,
		nonceManager:    nonceManager,
		gasLimit:        gasLimit,

		verifyPublishers: verifyPublishers,

		contract:        nil,
		wsContract:      nil,
		client:          nil,
		chainID:         nil,
		version:         nil,
		gasFeeCap:       nil,
		gasTipCap:       nil,
		useSyncSend:     useSyncSend,
		singleUpdateFee: nil,
		lastSetGasCaps:  time.Time{},
	}, nil
}

func (eci *ContractInteractor) ConnectHTTP(ctx context.Context, url string) error {
	client, err := ethclient.Dial(url)
	if err != nil {
		return fmt.Errorf("failed to connect to RPC: %w", err)
	}

	chainID, err := client.ChainID(ctx)
	if err != nil {
		return fmt.Errorf("failed to get network ID: %w", err)
	}

	contract, err := bindings.NewStorkContract(eci.contractAddress, client)
	if err != nil {
		return fmt.Errorf("failed to create contract instance: %w", err)
	}

	eci.contract = contract
	eci.client = client
	eci.chainID = chainID

	// load version
	versionStr, err := eci.contract.Version(makeCallOpts(ctx))
	if err != nil {
		eci.logger.Error().Err(err).Msg("Failed to get contract version")
	}
	version, err := semver.NewVersion(versionStr)
	if err != nil {
		eci.logger.Error().Err(err).Msg("Failed to parse contract version")
	}
	eci.version = version
	eci.logger.Info().Interface("version", eci.version).Msg("contract version")

	// set single update fee
	singleUpdateFee, err := eci.getSingleUpdateFee(ctx)
	if err != nil {
		return fmt.Errorf("failed to get single update fee: %w", err)
	}
	eci.singleUpdateFee = singleUpdateFee

	return nil
}

func (eci *ContractInteractor) ConnectWs(ctx context.Context, url string) error {
	var wsClient *ethclient.Client

	var err error

	if url != "" {
		wsClient, err = ethclient.DialContext(ctx, url)
		if err != nil {
			return fmt.Errorf("failed to connect to WS: %w", err)
		} else {
			eci.logger.Info().Msg("Connected to WebSocket endpoint")
		}
	}

	var wsContract *bindings.StorkContract
	if wsClient != nil {
		wsContract, err = bindings.NewStorkContract(eci.contractAddress, wsClient)
		if err != nil {
			return fmt.Errorf("failed to create WebSocket contract instance: %w", err)
		}
	}

	eci.wsContract = wsContract

	return nil
}

func (eci *ContractInteractor) ListenContractEvents(
	ctx context.Context, ch chan map[types.InternalEncodedAssetID]types.InternalTemporalNumericValue,
) {
	if eci.wsContract == nil {
		eci.logger.Warn().Msg("WebSocket contract not available, cannot listen for events")

		return
	}

	sub, eventCh, err := setupSubscription(eci, makeWatchOpts(ctx))
	if err != nil {
		eci.logger.Warn().Err(err).Msg("Failed to establish initial subscription")

		return
	}

	defer func() {
		eci.logger.Debug().Msg("Exiting ListenContractEvents")

		if sub != nil {
			sub.Unsubscribe()
			close(eventCh)
		}
	}()

	eci.logger.Info().Msg("Listening for contract events via WebSocket")

	for {
		err = eci.listenLoop(ctx, sub, eventCh, ch)
		if ctx.Err() != nil {
			return
		}

		eci.logger.Warn().Err(err).Msg("Error while watching contract events")

		if sub != nil {
			sub.Unsubscribe()
			sub = nil
		}

		sub, eventCh, err = eci.reconnect(ctx, makeWatchOpts(ctx))
		if err != nil {
			return
		}
	}
}

func (eci *ContractInteractor) PullValues(
	ctx context.Context,
	encodedAssetIDs []types.InternalEncodedAssetID,
) (map[types.InternalEncodedAssetID]types.InternalTemporalNumericValue, error) {
	if eci.version != nil && eci.version.Compare(semver.MustParse("1.0.5")) >= 0 {
		return eci.batchPullValues(ctx, encodedAssetIDs)
	} else {
		return eci.individuallyPullValues(ctx, encodedAssetIDs)
	}
}

func makeCallOpts(ctx context.Context) *bind.CallOpts {
	return &bind.CallOpts{Context: ctx} //nolint:exhaustruct
}

func makeWatchOpts(ctx context.Context) *bind.WatchOpts {
	return &bind.WatchOpts{Context: ctx}
}

func getUpdatePayload(
	priceUpdates []types.AggregatedSignedPrice,
) ([]bindings.StorkStructsTemporalNumericValueInput, error) {
	updates := make([]bindings.StorkStructsTemporalNumericValueInput, len(priceUpdates))
	i := 0

	for _, priceUpdate := range priceUpdates {
		quantizedPriceBigInt := new(big.Int)
		//nolint:mnd // base number.
		quantizedPriceBigInt.SetString(string(priceUpdate.StorkSignedPrice.QuantizedPrice), 10)

		encodedAssetID, err := pusher.HexStringToByte32(string(priceUpdate.StorkSignedPrice.EncodedAssetID))
		if err != nil {
			return nil, fmt.Errorf("failed to parse signature R: %w", err)
		}

		rBytes, err := pusher.HexStringToByte32(priceUpdate.StorkSignedPrice.TimestampedSignature.Signature.R)
		if err != nil {
			return nil, fmt.Errorf("failed to parse signature S: %w", err)
		}

		sBytes, err := pusher.HexStringToByte32(priceUpdate.StorkSignedPrice.TimestampedSignature.Signature.S)
		if err != nil {
			return nil, fmt.Errorf("failed to parse signature S: %w", err)
		}

		publisherMerkleRoot, err := pusher.HexStringToByte32(priceUpdate.StorkSignedPrice.PublisherMerkleRoot)
		if err != nil {
			return nil, fmt.Errorf("failed to parse PublisherMerkleRoot: %w", err)
		}

		checksum, err := pusher.HexStringToByte32(priceUpdate.StorkSignedPrice.StorkCalculationAlg.Checksum)
		if err != nil {
			return nil, fmt.Errorf("failed to parse signature V: %w", err)
		}

		vInt, err := strconv.ParseInt(priceUpdate.StorkSignedPrice.TimestampedSignature.Signature.V[2:], 16, 8)
		if err != nil || vInt < 0 || vInt > 255 {
			return nil, fmt.Errorf("failed to parse signature V: %w", err)
		}

		updates[i] = bindings.StorkStructsTemporalNumericValueInput{
			TemporalNumericValue: bindings.StorkStructsTemporalNumericValue{
				TimestampNs:    priceUpdate.StorkSignedPrice.TimestampedSignature.TimestampNano,
				QuantizedValue: quantizedPriceBigInt,
			},
			Id:                  encodedAssetID,
			PublisherMerkleRoot: publisherMerkleRoot,
			ValueComputeAlgHash: checksum,
			R:                   rBytes,
			S:                   sBytes,
			V:                   uint8(vInt),
		}
		i++
	}

	return updates, nil
}

type verifyPayload struct {
	pubSigs    []bindings.StorkStructsPublisherSignature
	merkleRoot [32]byte
}

func getVerifyPublishersPayloads(
	priceUpdates map[types.InternalEncodedAssetID]types.AggregatedSignedPrice,
) ([]verifyPayload, error) {
	payloads := make([]verifyPayload, len(priceUpdates))
	i := 0

	for _, priceUpdate := range priceUpdates {
		merkleRootBytes, err := pusher.HexStringToByte32(priceUpdate.StorkSignedPrice.PublisherMerkleRoot)
		if err != nil {
			return nil, fmt.Errorf("failed to parse PublisherMerkleRoot: %w", err)
		}

		payloads[i] = verifyPayload{
			pubSigs:    make([]bindings.StorkStructsPublisherSignature, len(priceUpdate.SignedPrices)),
			merkleRoot: merkleRootBytes,
		}
		j := 0

		var (
			pubKeyBytes [20]byte
			rBytes      [32]byte
			sBytes      [32]byte
			vInt        int64
		)

		for _, signedPrice := range priceUpdate.SignedPrices {
			pubKeyBytes, err = pusher.HexStringToByte20(string(signedPrice.PublisherKey))
			if err != nil {
				return nil, fmt.Errorf("failed to parse PublisherMerkleRoot: %w", err)
			}

			quantizedPriceBigInt := new(big.Int)
			//nolint:mnd // base number.
			quantizedPriceBigInt.SetString(string(signedPrice.QuantizedPrice), 10)

			rBytes, err = pusher.HexStringToByte32(signedPrice.TimestampedSignature.Signature.R)
			if err != nil {
				return nil, fmt.Errorf("failed to parse signature R: %w", err)
			}

			sBytes, err = pusher.HexStringToByte32(signedPrice.TimestampedSignature.Signature.S)
			if err != nil {
				return nil, fmt.Errorf("failed to parse signature S: %w", err)
			}

			vInt, err = strconv.ParseInt(signedPrice.TimestampedSignature.Signature.V[2:], 16, 8)
			if err != nil || vInt < 0 || vInt > 255 {
				return nil, fmt.Errorf("failed to parse signature V: %w", err)
			}

			payloads[i].pubSigs[j] = bindings.StorkStructsPublisherSignature{
				PubKey:         pubKeyBytes,
				AssetPairId:    signedPrice.ExternalAssetID,
				Timestamp:      signedPrice.TimestampedSignature.TimestampNano / uint64(time.Second),
				QuantizedValue: quantizedPriceBigInt,
				R:              rBytes,
				S:              sBytes,
				V:              uint8(vInt),
			}
			j++
		}

		i++
	}

	return payloads, nil
}

func (eci *ContractInteractor) BatchPushToContract(
	ctx context.Context,
	priceUpdates map[types.InternalEncodedAssetID]types.AggregatedSignedPrice,
) error {
	if eci.verifyPublishers {
		publisherVerifyPayloads, err := getVerifyPublishersPayloads(priceUpdates)
		if err != nil {
			return err
		}

		var verified bool
		for i := range publisherVerifyPayloads {
			verified, err = eci.contract.VerifyPublisherSignaturesV1(
				makeCallOpts(ctx),
				publisherVerifyPayloads[i].pubSigs,
				publisherVerifyPayloads[i].merkleRoot,
			)
			if err != nil {
				eci.logger.Error().Err(err).Msg("Failed to verify publisher signatures")

				return fmt.Errorf("failed to verify publisher signatures: %w", err)
			}

			if !verified {
				eci.logger.Error().Msg("Publisher signatures not verified, skipping update")

				return nil
			}
		}
	}
	// convert to []types.AggregatedSignedPrice
	priceUpdatesSlice := make([]types.AggregatedSignedPrice, 0, len(priceUpdates))
	for _, priceUpdate := range priceUpdates {
		priceUpdatesSlice = append(priceUpdatesSlice, priceUpdate)
	}

	updatePayload, err := getUpdatePayload(priceUpdatesSlice)
	if err != nil {
		return err
	}

	// this is the same logic as whats on the contract, but do it locally to avoid an rpc call
	fee := eci.getUpdateFee(updatePayload)

	tx, err := eci.submitTransaction(ctx, updatePayload, fee)
	if err != nil {
		if errors.Is(err, etherrors.ErrReplaceUnderpriced) {
			eci.logger.Warn().Err(err).Msg("Transaction underpriced, retrying with bumped gas prices")

			tx, err = eci.retryTransaction(ctx, updatePayload, fee)
			if err != nil {
				return fmt.Errorf("failed to retry transaction submission: %w", err)
			}
		} else {
			return fmt.Errorf("failed to submit transaction: %w", err)
		}
	}

	eci.logger.Debug().
		Str("txHash", tx.Hash().Hex()).
		Int("numUpdates", len(updatePayload)).
		Uint64("gasPrice", tx.GasPrice().Uint64()).
		Msg("Pushed new values to contract")

	return nil
}

func (eci *ContractInteractor) getUpdateFee(updatePayload []bindings.StorkStructsTemporalNumericValueInput) *big.Int {
	fee := new(big.Int).Mul(eci.singleUpdateFee, big.NewInt(int64(len(updatePayload))))
	return fee
}

func (eci *ContractInteractor) GetWalletBalance(ctx context.Context) (float64, error) {
	publicKey := eci.privateKey.Public()

	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return -1, ErrCastingPublicKeyToECDSA
	}

	address := crypto.PubkeyToAddress(*publicKeyECDSA)

	balance, err := eci.client.BalanceAt(ctx, address, nil)
	if err != nil {
		return -1, fmt.Errorf("failed to get wallet balance: %w", err)
	}

	balanceFloat, _ := balance.Float64()

	return balanceFloat, nil
}

func (eci *ContractInteractor) batchPullValues(
	ctx context.Context,
	encodedAssetIDs []types.InternalEncodedAssetID,
) (map[types.InternalEncodedAssetID]types.InternalTemporalNumericValue, error) {
	polledVals := make(map[types.InternalEncodedAssetID]types.InternalTemporalNumericValue)

	compatibleEncodedAssetIDs := make([][32]byte, 0, len(encodedAssetIDs))
	for _, encodedAssetID := range encodedAssetIDs {
		compatibleEncodedAssetIDs = append(compatibleEncodedAssetIDs, encodedAssetID)
	}

	storkStructsTemporalNumericValues, err := eci.contract.GetTemporalNumericValuesUnsafeV1(
		makeCallOpts(ctx), compatibleEncodedAssetIDs,
	)
	if err != nil {
		if strings.Contains(err.Error(), "NotFound()") || strings.Contains(err.Error(), "0xc5723b51") {
			eci.logger.Warn().Err(err).Msg("No value found")

			return polledVals, nil
		}

		return nil, fmt.Errorf("failed to get temporal numeric values: %w", err)
	}

	for i, storkStructsTemporalNumericValue := range storkStructsTemporalNumericValues {
		polledVals[encodedAssetIDs[i]] = types.InternalTemporalNumericValue(storkStructsTemporalNumericValue)
	}

	return polledVals, nil
}

func (eci *ContractInteractor) individuallyPullValues(
	ctx context.Context,
	encodedAssetIDs []types.InternalEncodedAssetID,
) (map[types.InternalEncodedAssetID]types.InternalTemporalNumericValue, error) {
	polledVals := make(map[types.InternalEncodedAssetID]types.InternalTemporalNumericValue)

	var failedToGetLatestValueErr error

	for _, encodedAssetID := range encodedAssetIDs {
		storkStructsTemporalNumericValue, err := eci.contract.GetTemporalNumericValueUnsafeV1(
			makeCallOpts(ctx), encodedAssetID,
		)
		if err != nil {
			if strings.Contains(err.Error(), "NotFound()") || strings.Contains(err.Error(), "0xc5723b51") {
				eci.logger.Warn().Err(err).Str("assetID", hex.EncodeToString(encodedAssetID[:])).Msg("No value found")
			} else {
				eci.logger.Warn().Err(err).Str("assetID", hex.EncodeToString(encodedAssetID[:])).Msg("Failed to get latest value")
				failedToGetLatestValueErr = err
			}

			continue
		}

		polledVals[encodedAssetID] = types.InternalTemporalNumericValue(storkStructsTemporalNumericValue)
	}

	if failedToGetLatestValueErr != nil {
		err := fmt.Errorf(
			"failed to pull at least one value from the contract. Last error: %w",
			failedToGetLatestValueErr,
		)

		return polledVals, err
	}

	return polledVals, nil
}

//nolint:ireturn // interface return acceptable here.
func setupSubscription(
	eci *ContractInteractor,
	watchOpts *bind.WatchOpts,
) (ethereum.Subscription, chan *bindings.StorkContractValueUpdate, error) {
	eventCh := make(chan *bindings.StorkContractValueUpdate)

	sub, err := eci.wsContract.WatchValueUpdate(watchOpts, eventCh, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to watch contract events: %w", err)
	}

	return sub, eventCh, nil
}

func (eci *ContractInteractor) listenLoop(
	ctx context.Context,
	sub ethereum.Subscription,
	eventCh chan *bindings.StorkContractValueUpdate,
	outCh chan map[types.InternalEncodedAssetID]types.InternalTemporalNumericValue,
) error {
	for {
		select {
		case <-ctx.Done():
			return nil

		case err := <-sub.Err():
			return fmt.Errorf("error from subscription: %w", err)

		case vLog, ok := <-eventCh:
			if !ok {
				eci.logger.Warn().Msg("Event channel closed, exiting event listener")

				return ErrEventChannelClosed
			}

			tv := types.InternalTemporalNumericValue{
				QuantizedValue: vLog.QuantizedValue,
				TimestampNs:    vLog.TimestampNs,
			}
			select {
			case outCh <- map[types.InternalEncodedAssetID]types.InternalTemporalNumericValue{vLog.Id: tv}:
			case <-ctx.Done():
				return fmt.Errorf("context done: %w", ctx.Err())
			}
		}
	}
}

//nolint:ireturn // interface return acceptable here.
func (eci *ContractInteractor) reconnect(
	ctx context.Context,
	watchOpts *bind.WatchOpts,
) (ethereum.Subscription, chan *bindings.StorkContractValueUpdate, error) {
	backoff := initialBackoff
	for retryCount := range maxRetryAttempts {
		backoff = time.Duration(float64(backoff) * exponentialBackoffFactor)
		eci.logger.Info().Dur("backoff", backoff).
			Int("attempt", retryCount+1).
			Int("maxAttempts", maxRetryAttempts).
			Msg("Attempting to reconnect to contract events")

		select {
		case <-ctx.Done():
			return nil, nil, fmt.Errorf("context done: %w", ctx.Err())
		case <-time.After(backoff):
			newSub, newEventCh, err := setupSubscription(eci, watchOpts)
			if err != nil {
				eci.logger.Warn().Err(err).Msg("Failed to reconnect to contract events")

				continue
			}

			eci.logger.Info().Msg("Successfully reconnected to contract events")

			return newSub, newEventCh, nil
		}
	}

	eci.logger.Error().Int("maxRetryAttempts", maxRetryAttempts).
		Msg("Max retry attempts reached, giving up on reconnection")

	return nil, nil, ErrMaxRetryAttemptsReached
}

func (eci *ContractInteractor) submitTransaction(
	ctx context.Context,
	updatePayload []bindings.StorkStructsTemporalNumericValueInput,
	fee *big.Int,
) (*ethtypes.Transaction, error) {
	auth, err := bind.NewKeyedTransactorWithChainID(eci.privateKey, eci.chainID)
	if err != nil {
		return nil, fmt.Errorf("failed to get auth data: %w", err)
	}

	if time.Since(eci.lastSetGasCaps) > gasCalcResetInterval {
		eci.gasFeeCap = nil
		eci.gasTipCap = nil

		singleUpdateFee, err := eci.getSingleUpdateFee(ctx)
		if err != nil {
			eci.logger.Error().Err(err).Msg("failed to get single update fee")
		} else {
			eci.singleUpdateFee = singleUpdateFee
		}
		eci.lastSetGasCaps = time.Now()
	}
	nonce, err := eci.nonceManager.GetLatestNonce(ctx, eci.client, crypto.PubkeyToAddress(eci.privateKey.PublicKey))
	if err != nil {
		return nil, fmt.Errorf("failed to get latest nonce: %w", err)
	}

	auth.Context = ctx
	auth.Value = fee
	auth.NoSend = true // always send the transaction manually
	auth.Nonce = nonce
	auth.GasLimit = eci.gasLimit // default 0

	if eci.gasFeeCap != nil {
		auth.GasFeeCap = eci.gasFeeCap
	}

	if eci.gasTipCap != nil {
		auth.GasTipCap = eci.gasTipCap
	}

	tx, err := eci.contract.UpdateTemporalNumericValuesV1(auth, updatePayload)
	if err != nil {
		if revertData, ok := ethclient.RevertErrorData(err); ok {
			eci.logger.Error().Str("revertData", hex.EncodeToString(revertData)).Msg("transaction reverted with data")
		}
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	if eci.useSyncSend {
		receipt, txErr := eci.client.SendTransactionSync(ctx, tx, nil)
		err := eci.nonceManager.IncrementNonce(ctx, eci.client, crypto.PubkeyToAddress(eci.privateKey.PublicKey))
		if err != nil {
			return nil, fmt.Errorf("failed to increment nonce: %w", err)
		}

		if txErr != nil {
			if strings.Contains(txErr.Error(), "nonce") {
				eci.logger.Warn().Msg("Nonce mismatch, resetting nonce")
				err := eci.nonceManager.ResetNonce(ctx, eci.client, crypto.PubkeyToAddress(eci.privateKey.PublicKey))
				if err != nil {
					return nil, fmt.Errorf("failed to reset nonce: %w", err)
				}
			}
			return nil, fmt.Errorf("failed to send transaction: %w", txErr)
		}

		if receipt.Status != 1 {
			return nil, fmt.Errorf("eth_sendRawTransactionSync transaction failed")
		}

		return tx, nil
	} else {
		txErr := eci.client.SendTransaction(ctx, tx)
		err := eci.nonceManager.IncrementNonce(ctx, eci.client, crypto.PubkeyToAddress(eci.privateKey.PublicKey))
		if err != nil {
			return nil, fmt.Errorf("failed to increment nonce: %w", err)
		}

		if txErr != nil {
			if revertData, ok := ethclient.RevertErrorData(txErr); ok {
				eci.logger.Error().Str("revertData", hex.EncodeToString(revertData)).Msg("transaction reverted with data")
			} else if strings.Contains(txErr.Error(), "nonce") {
				eci.logger.Warn().Msg("Nonce mismatch, resetting nonce")
				err := eci.nonceManager.ResetNonce(ctx, eci.client, crypto.PubkeyToAddress(eci.privateKey.PublicKey))
				if err != nil {
					return nil, fmt.Errorf("failed to reset nonce: %w", err)
				}
			}
			return nil, fmt.Errorf("failed to send transaction: %w", txErr)
		}
	}

	eci.gasFeeCap = tx.GasFeeCap()
	eci.gasTipCap = tx.GasTipCap()

	return tx, nil
}

func (eci *ContractInteractor) getSingleUpdateFee(ctx context.Context) (*big.Int, error) {
	singleUpdateFee, err := eci.contract.SingleUpdateFeeInWei(makeCallOpts(ctx))
	if err != nil {
		return nil, fmt.Errorf("failed to get single update fee: %w", err)
	}
	return singleUpdateFee, nil
}

func (eci *ContractInteractor) retryTransaction(
	ctx context.Context,
	updatePayload []bindings.StorkStructsTemporalNumericValueInput,
	fee *big.Int,
) (*ethtypes.Transaction, error) {
	var lastErr error

	for retryCount := range maxTransactionAttempts {
		gasTipCap, err := eci.client.SuggestGasTipCap(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get gas tip cap: %w", err)
		}

		// gas price is used to estimate gas fee
		gasPrice, err := eci.client.SuggestGasPrice(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get gas price: %w", err)
		}

		newGasFeeCap, newGasTipCap := getBumpedGasPrices(gasPrice, gasTipCap, int64(retryCount+1))
		eci.logger.Debug().
			Str("gasFeeCap", newGasFeeCap.String()).
			Str("gasTipCap", newGasTipCap.String()).
			Msg("Retrying with bumped gas prices")

		eci.gasFeeCap = newGasFeeCap
		eci.gasTipCap = newGasTipCap

		tx, err := eci.submitTransaction(ctx, updatePayload, fee)

		lastErr = err
		if err == nil {
			return tx, nil
		}
	}

	return nil, lastErr
}

// For simplicity, this function assumes the mnemonic file contains the private key directly.
func loadPrivateKey(mnemonicFile []byte) (*ecdsa.PrivateKey, error) {
	// remove any trailing newline characters
	dataString := strings.TrimSpace(string(mnemonicFile))

	privateKey, err := crypto.HexToECDSA(dataString)
	if err != nil {
		return nil, fmt.Errorf("failed to load private key: %w", err)
	}

	return privateKey, nil
}

// To replace a transaction, the gas price must be bumped by at least 10%.
func getBumpedGasPrices(
	gasPrice *big.Int,
	gasTipCap *big.Int,
	retryCount int64,
) (*big.Int, *big.Int) {
	retryBig := big.NewInt(retryCount)

	bumpNumeratorPower := new(big.Int).Exp(big.NewInt(gasBumpNumerator), retryBig, nil)
	bumpDenominatorPower := new(big.Int).Exp(big.NewInt(gasBumpDenominator), retryBig, nil)

	retryGasFeeCap := new(big.Int).Mul(gasPrice, bumpNumeratorPower)
	retryGasFeeCap.Div(retryGasFeeCap, bumpDenominatorPower)

	retryGasTipCap := new(big.Int).Mul(gasTipCap, bumpNumeratorPower)
	retryGasTipCap.Div(retryGasTipCap, bumpDenominatorPower)

	return retryGasFeeCap, retryGasTipCap
}
