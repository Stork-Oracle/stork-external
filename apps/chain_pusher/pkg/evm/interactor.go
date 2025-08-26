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

	"github.com/Stork-Oracle/stork-external/apps/chain_pusher/pkg/evm/bindings"
	"github.com/Stork-Oracle/stork-external/apps/chain_pusher/pkg/pusher"
	"github.com/Stork-Oracle/stork-external/apps/chain_pusher/pkg/types"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rs/zerolog"
)

type EvmContractInteractor struct {
	logger zerolog.Logger

	contract   *bindings.StorkContract
	wsContract *bindings.StorkContract
	client     *ethclient.Client

	privateKey *ecdsa.PrivateKey
	chainID    *big.Int
	gasLimit   uint64

	verifyPublishers bool
}

const (
	// 1 * (1.5 ^ 10) = 57.66 seconds (last attempt delay)
	maxRetryAttempts         = 10
	initialBackoff           = 1 * time.Second
	exponentialBackoffFactor = 1.5
)

func NewEvmContractInteractor(
	rpcUrl string,
	wsUrl string,
	contractAddr string,
	mnemonic []byte,
	verifyPublishers bool,
	logger zerolog.Logger,
	gasLimit uint64,
) (*EvmContractInteractor, error) {
	privateKey, err := loadPrivateKey(mnemonic)
	if err != nil {
		return nil, err
	}

	client, err := ethclient.Dial(rpcUrl)
	if err != nil {
		return nil, err
	}

	var wsClient *ethclient.Client
	if wsUrl != "" {
		wsClient, err = ethclient.Dial(wsUrl)
		if err != nil {
			logger.Warn().Err(err).Msg("Failed to connect to WebSocket endpoint")
		} else {
			logger.Info().Msg("Connected to WebSocket endpoint")
		}
	}

	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		return nil, err
	}

	contractAddress := common.HexToAddress(contractAddr)
	contract, err := bindings.NewStorkContract(contractAddress, client)
	if err != nil {
		return nil, err
	}

	var wsContract *bindings.StorkContract
	if wsClient != nil {
		wsContract, err = bindings.NewStorkContract(contractAddress, wsClient)
		if err != nil {
			logger.Warn().Err(err).Msg("Failed to create WebSocket contract instance")
		}
	}

	return &EvmContractInteractor{
		logger: logger,

		contract:   contract,
		wsContract: wsContract,
		client:     client,
		privateKey: privateKey,
		chainID:    chainID,
		gasLimit:   gasLimit,

		verifyPublishers: verifyPublishers,
	}, nil
}

func (eci *EvmContractInteractor) ListenContractEvents(
	ctx context.Context, ch chan map[types.InternalEncodedAssetId]types.InternalTemporalNumericValue,
) {
	if eci.wsContract == nil {
		eci.logger.Warn().Msg("WebSocket contract not available, cannot listen for events")
		return
	}

	watchOpts := &bind.WatchOpts{Context: context.Background()}

	sub, eventCh, err := setupSubscription(eci, watchOpts)
	if err != nil {
		eci.logger.Error().Err(err).Msg("Failed to establish initial subscription")
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
		err := eci.listenLoop(ctx, sub, eventCh, ch)
		if ctx.Err() != nil {
			return
		}

		eci.logger.Warn().Err(err).Msg("Error while watching contract events")
		if sub != nil {
			sub.Unsubscribe()
			sub = nil
		}

		sub, eventCh, err = eci.reconnect(ctx, watchOpts)
		if err != nil {
			return
		}
	}
}

func setupSubscription(
	eci *EvmContractInteractor,
	watchOpts *bind.WatchOpts,
) (ethereum.Subscription, chan *bindings.StorkContractValueUpdate, error) {
	eventCh := make(chan *bindings.StorkContractValueUpdate)
	sub, err := eci.wsContract.WatchValueUpdate(watchOpts, eventCh, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to watch contract events: %w", err)
	}
	return sub, eventCh, nil
}

func (eci *EvmContractInteractor) listenLoop(
	ctx context.Context,
	sub ethereum.Subscription,
	eventCh chan *bindings.StorkContractValueUpdate,
	outCh chan map[types.InternalEncodedAssetId]types.InternalTemporalNumericValue,
) error {
	for {
		select {
		case <-ctx.Done():
			return nil

		case err := <-sub.Err():
			return err

		case vLog, ok := <-eventCh:
			if !ok {
				eci.logger.Warn().Msg("Event channel closed, exiting event listener")
				return errors.New("event channel closed is closed")
			}

			tv := types.InternalTemporalNumericValue{
				QuantizedValue: vLog.QuantizedValue,
				TimestampNs:    vLog.TimestampNs,
			}
			select {
			case outCh <- map[types.InternalEncodedAssetId]types.InternalTemporalNumericValue{vLog.Id: tv}:
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}
}

func (eci *EvmContractInteractor) reconnect(
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
			return nil, nil, ctx.Err()
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
	return nil, nil, errors.New("max retry attempts reached")
}

func (eci *EvmContractInteractor) PullValues(encodedAssetIds []types.InternalEncodedAssetId) (map[types.InternalEncodedAssetId]types.InternalTemporalNumericValue, error) {
	polledVals := make(map[types.InternalEncodedAssetId]types.InternalTemporalNumericValue)
	for _, encodedAssetId := range encodedAssetIds {
		storkStructsTemporalNumericValue, err := eci.contract.GetTemporalNumericValueUnsafeV1(nil, encodedAssetId)
		if err != nil {
			if strings.Contains(err.Error(), "NotFound()") {
				eci.logger.Warn().Err(err).Str("assetId", hex.EncodeToString(encodedAssetId[:])).Msg("No value found")
			} else {
				eci.logger.Warn().Err(err).Str("assetId", hex.EncodeToString(encodedAssetId[:])).Msg("Failed to get latest value")
			}

			continue
		}
		polledVals[encodedAssetId] = types.InternalTemporalNumericValue(storkStructsTemporalNumericValue)
	}
	return polledVals, nil
}

func getUpdatePayload(priceUpdates map[types.InternalEncodedAssetId]types.AggregatedSignedPrice) ([]bindings.StorkStructsTemporalNumericValueInput, error) {
	updates := make([]bindings.StorkStructsTemporalNumericValueInput, len(priceUpdates))
	i := 0
	for _, priceUpdate := range priceUpdates {

		quantizedPriceBigInt := new(big.Int)
		quantizedPriceBigInt.SetString(string(priceUpdate.StorkSignedPrice.QuantizedPrice), 10)

		encodedAssetId, err := pusher.StringToByte32(string(priceUpdate.StorkSignedPrice.EncodedAssetId))
		if err != nil {
			return nil, err
		}

		rBytes, err := pusher.StringToByte32(priceUpdate.StorkSignedPrice.TimestampedSignature.Signature.R)
		if err != nil {
			return nil, err
		}

		sBytes, err := pusher.StringToByte32(priceUpdate.StorkSignedPrice.TimestampedSignature.Signature.S)
		if err != nil {
			return nil, err
		}

		publisherMerkleRoot, err := pusher.StringToByte32(priceUpdate.StorkSignedPrice.PublisherMerkleRoot)
		if err != nil {
			return nil, err
		}

		checksum, err := pusher.StringToByte32(priceUpdate.StorkSignedPrice.StorkCalculationAlg.Checksum)
		if err != nil {
			return nil, err
		}

		vInt, err := strconv.ParseInt(priceUpdate.StorkSignedPrice.TimestampedSignature.Signature.V[2:], 16, 8)
		if err != nil {
			return nil, err
		}

		updates[i] = bindings.StorkStructsTemporalNumericValueInput{
			TemporalNumericValue: bindings.StorkStructsTemporalNumericValue{
				TimestampNs:    uint64(priceUpdate.StorkSignedPrice.TimestampedSignature.TimestampNano),
				QuantizedValue: quantizedPriceBigInt,
			},
			Id:                  encodedAssetId,
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

func getVerifyPublishersPayloads(priceUpdates map[types.InternalEncodedAssetId]types.AggregatedSignedPrice) ([]verifyPayload, error) {
	payloads := make([]verifyPayload, len(priceUpdates))
	i := 0
	for _, priceUpdate := range priceUpdates {
		merkleRootBytes, err := pusher.StringToByte32(priceUpdate.StorkSignedPrice.PublisherMerkleRoot)
		if err != nil {
			return nil, err
		}

		payloads[i] = verifyPayload{
			pubSigs:    make([]bindings.StorkStructsPublisherSignature, len(priceUpdate.SignedPrices)),
			merkleRoot: merkleRootBytes,
		}
		j := 0
		for _, signedPrice := range priceUpdate.SignedPrices {
			pubKeyBytes, err := pusher.StringToByte20(string(signedPrice.PublisherKey))
			if err != nil {
				return nil, err
			}

			quantizedPriceBigInt := new(big.Int)
			quantizedPriceBigInt.SetString(string(signedPrice.QuantizedPrice), 10)

			rBytes, err := pusher.StringToByte32(signedPrice.TimestampedSignature.Signature.R)
			if err != nil {
				return nil, err
			}

			sBytes, err := pusher.StringToByte32(signedPrice.TimestampedSignature.Signature.S)
			if err != nil {
				return nil, err
			}

			vInt, err := strconv.ParseInt(signedPrice.TimestampedSignature.Signature.V[2:], 16, 8)
			if err != nil {
				return nil, err
			}

			payloads[i].pubSigs[j] = bindings.StorkStructsPublisherSignature{
				PubKey:         pubKeyBytes,
				AssetPairId:    signedPrice.ExternalAssetId,
				Timestamp:      uint64(signedPrice.TimestampedSignature.TimestampNano) / 1000000000,
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

func (eci *EvmContractInteractor) BatchPushToContract(priceUpdates map[types.InternalEncodedAssetId]types.AggregatedSignedPrice) error {
	if eci.verifyPublishers {
		publisherVerifyPayloads, err := getVerifyPublishersPayloads(priceUpdates)
		if err != nil {
			return err
		}
		for i := range publisherVerifyPayloads {
			verified, err := eci.contract.VerifyPublisherSignaturesV1(nil, publisherVerifyPayloads[i].pubSigs, publisherVerifyPayloads[i].merkleRoot)
			if err != nil {
				eci.logger.Error().Err(err).Msg("Failed to verify publisher signatures")
				return err
			}
			if !verified {
				eci.logger.Error().Msg("Publisher signatures not verified, skipping update")
				return nil
			}
		}
	}

	updatePayload, err := getUpdatePayload(priceUpdates)
	if err != nil {
		return err
	}

	fee, err := eci.contract.GetUpdateFeeV1(nil, updatePayload)
	if err != nil {
		return err
	}

	auth, err := bind.NewKeyedTransactorWithChainID(eci.privateKey, eci.chainID)
	if err != nil {
		return err
	}

	// let the library auto-estimate the gas price
	auth.GasLimit = eci.gasLimit
	auth.Value = fee

	tx, err := eci.contract.UpdateTemporalNumericValuesV1(auth, updatePayload)
	if err != nil {
		return err
	}

	eci.logger.Info().
		Str("txHash", tx.Hash().Hex()).
		Int("numUpdates", len(updatePayload)).
		Uint64("gasPrice", tx.GasPrice().Uint64()).
		Msg("Pushed new values to contract")
	return nil
}

func (eci *EvmContractInteractor) GetWalletBalance() (float64, error) {
	publicKey := eci.privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return -1, fmt.Errorf("error casting public key to ECDSA")
	}

	address := crypto.PubkeyToAddress(*publicKeyECDSA)

	balance, err := eci.client.BalanceAt(context.Background(), address, nil)
	if err != nil {
		return -1, err
	}
	balanceFloat, _ := balance.Float64()

	return balanceFloat, nil
}

// For simplicity, this function assumes the mnemonic file contains the private key directly
func loadPrivateKey(mnemonicFile []byte) (*ecdsa.PrivateKey, error) {
	// remove any trailing newline characters
	dataString := strings.TrimSpace(string(mnemonicFile))

	privateKey, err := crypto.HexToECDSA(dataString)
	if err != nil {
		return nil, err
	}
	return privateKey, nil
}
