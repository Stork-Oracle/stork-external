package chain_pusher

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

	contract_bindings "github.com/Stork-Oracle/stork-external/apps/lib/chain_pusher/contract_bindings/evm"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rs/zerolog"
)

type EvmContractInteractor struct {
	logger zerolog.Logger

	contract   *contract_bindings.StorkContract
	wsContract *contract_bindings.StorkContract
	client     *ethclient.Client

	privateKey *ecdsa.PrivateKey
	chainID    *big.Int

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
	contract, err := contract_bindings.NewStorkContract(contractAddress, client)
	if err != nil {
		return nil, err
	}

	var wsContract *contract_bindings.StorkContract
	if wsClient != nil {
		wsContract, err = contract_bindings.NewStorkContract(contractAddress, wsClient)
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

		verifyPublishers: verifyPublishers,
	}, nil
}

func (sci *EvmContractInteractor) ListenContractEvents(
	ctx context.Context, ch chan map[InternalEncodedAssetId]InternalTemporalNumericValue,
) {
	if sci.wsContract == nil {
		sci.logger.Warn().Msg("WebSocket contract not available, cannot listen for events")
		return
	}

	watchOpts := &bind.WatchOpts{Context: context.Background()}

	sub, eventCh, err := setupSubscription(sci, watchOpts)
	if err != nil {
		sci.logger.Error().Err(err).Msg("Failed to establish initial subscription")
		return
	}

	defer func() {
		sci.logger.Debug().Msg("Exiting ListenContractEvents")
		if sub != nil {
			sub.Unsubscribe()
			close(eventCh)
		}
	}()

	sci.logger.Info().Msg("Listening for contract events via WebSocket")
	for {
		err := sci.listenLoop(ctx, sub, eventCh, ch)
		if ctx.Err() != nil {
			return
		}

		sci.logger.Warn().Err(err).Msg("Error while watching contract events")
		if sub != nil {
			sub.Unsubscribe()
			close(eventCh)
			sub = nil
		}

		sub, eventCh, err = sci.reconnect(ctx, watchOpts)
		if err != nil {
			return
		}
	}
}

func setupSubscription(
	sci *EvmContractInteractor,
	watchOpts *bind.WatchOpts,
) (ethereum.Subscription, chan *contract_bindings.StorkContractValueUpdate, error) {
	eventCh := make(chan *contract_bindings.StorkContractValueUpdate)
	sub, err := sci.wsContract.WatchValueUpdate(watchOpts, eventCh, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to watch contract events: %w", err)
	}
	return sub, eventCh, nil
}

func (sci *EvmContractInteractor) listenLoop(
	ctx context.Context,
	sub ethereum.Subscription,
	eventCh chan *contract_bindings.StorkContractValueUpdate,
	outCh chan map[InternalEncodedAssetId]InternalTemporalNumericValue,
) error {
	for {
		select {
		case <-ctx.Done():
			return nil

		case err := <-sub.Err():
			return err

		case vLog, ok := <-eventCh:
			if !ok {
				sci.logger.Warn().Msg("Event channel closed, exiting event listener")
				return errors.New("event channel closed is closed")
			}

			tv := InternalTemporalNumericValue{
				QuantizedValue: vLog.QuantizedValue,
				TimestampNs:    vLog.TimestampNs,
			}
			select {
			case outCh <- map[InternalEncodedAssetId]InternalTemporalNumericValue{vLog.Id: tv}:
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}
}

func (sci *EvmContractInteractor) reconnect(
	ctx context.Context,
	watchOpts *bind.WatchOpts,
) (ethereum.Subscription, chan *contract_bindings.StorkContractValueUpdate, error) {
	backoff := initialBackoff
	for retryCount := range maxRetryAttempts {
		backoff = time.Duration(float64(backoff) * exponentialBackoffFactor)
		sci.logger.Info().Dur("backoff", backoff).
			Int("attempt", retryCount+1).
			Int("maxAttempts", maxRetryAttempts).
			Msg("Attempting to reconnect to contract events")

		select {
		case <-ctx.Done():
			return nil, nil, ctx.Err()
		case <-time.After(backoff):
			newSub, newEventCh, err := setupSubscription(sci, watchOpts)
			if err != nil {
				sci.logger.Error().Err(err).Msg("Failed to reconnect to contract events")

				continue
			}

			sci.logger.Info().Msg("Successfully reconnected to contract events")
			return newSub, newEventCh, nil
		}
	}

	sci.logger.Error().Int("maxRetryAttempts", maxRetryAttempts).
		Msg("Max retry attempts reached, giving up on reconnection")
	return nil, nil, errors.New("max retry attempts reached")
}

func (sci *EvmContractInteractor) PullValues(encodedAssetIds []InternalEncodedAssetId) (map[InternalEncodedAssetId]InternalTemporalNumericValue, error) {
	polledVals := make(map[InternalEncodedAssetId]InternalTemporalNumericValue)
	for _, encodedAssetId := range encodedAssetIds {
		storkStructsTemporalNumericValue, err := sci.contract.GetTemporalNumericValueUnsafeV1(nil, encodedAssetId)
		if err != nil {
			if strings.Contains(err.Error(), "NotFound()") {
				sci.logger.Warn().Err(err).Str("assetId", hex.EncodeToString(encodedAssetId[:])).Msg("No value found")
			} else {
				sci.logger.Warn().Err(err).Str("assetId", hex.EncodeToString(encodedAssetId[:])).Msg("Failed to get latest value")
			}

			continue
		}
		polledVals[encodedAssetId] = InternalTemporalNumericValue(storkStructsTemporalNumericValue)
	}
	return polledVals, nil
}

func getUpdatePayload(priceUpdates map[InternalEncodedAssetId]AggregatedSignedPrice) ([]contract_bindings.StorkStructsTemporalNumericValueInput, error) {
	updates := make([]contract_bindings.StorkStructsTemporalNumericValueInput, len(priceUpdates))
	i := 0
	for _, priceUpdate := range priceUpdates {

		quantizedPriceBigInt := new(big.Int)
		quantizedPriceBigInt.SetString(string(priceUpdate.StorkSignedPrice.QuantizedPrice), 10)

		encodedAssetId, err := stringToByte32(string(priceUpdate.StorkSignedPrice.EncodedAssetId))
		if err != nil {
			return nil, err
		}

		rBytes, err := stringToByte32(priceUpdate.StorkSignedPrice.TimestampedSignature.Signature.R)
		if err != nil {
			return nil, err
		}

		sBytes, err := stringToByte32(priceUpdate.StorkSignedPrice.TimestampedSignature.Signature.S)
		if err != nil {
			return nil, err
		}

		publisherMerkleRoot, err := stringToByte32(priceUpdate.StorkSignedPrice.PublisherMerkleRoot)
		if err != nil {
			return nil, err
		}

		checksum, err := stringToByte32(priceUpdate.StorkSignedPrice.StorkCalculationAlg.Checksum)
		if err != nil {
			return nil, err
		}

		vInt, err := strconv.ParseInt(priceUpdate.StorkSignedPrice.TimestampedSignature.Signature.V[2:], 16, 8)
		if err != nil {
			return nil, err
		}

		updates[i] = contract_bindings.StorkStructsTemporalNumericValueInput{
			TemporalNumericValue: contract_bindings.StorkStructsTemporalNumericValue{
				TimestampNs:    uint64(priceUpdate.StorkSignedPrice.TimestampedSignature.Timestamp),
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
	pubSigs    []contract_bindings.StorkStructsPublisherSignature
	merkleRoot [32]byte
}

func getVerifyPublishersPayloads(priceUpdates map[InternalEncodedAssetId]AggregatedSignedPrice) ([]verifyPayload, error) {
	payloads := make([]verifyPayload, len(priceUpdates))
	i := 0
	for _, priceUpdate := range priceUpdates {
		merkleRootBytes, err := stringToByte32(priceUpdate.StorkSignedPrice.PublisherMerkleRoot)
		if err != nil {
			return nil, err
		}

		payloads[i] = verifyPayload{
			pubSigs:    make([]contract_bindings.StorkStructsPublisherSignature, len(priceUpdate.SignedPrices)),
			merkleRoot: merkleRootBytes,
		}
		j := 0
		for _, signedPrice := range priceUpdate.SignedPrices {
			pubKeyBytes, err := stringToByte20(string(signedPrice.PublisherKey))
			if err != nil {
				return nil, err
			}

			quantizedPriceBigInt := new(big.Int)
			quantizedPriceBigInt.SetString(string(signedPrice.QuantizedPrice), 10)

			rBytes, err := stringToByte32(signedPrice.TimestampedSignature.Signature.R)
			if err != nil {
				return nil, err
			}

			sBytes, err := stringToByte32(signedPrice.TimestampedSignature.Signature.S)
			if err != nil {
				return nil, err
			}

			vInt, err := strconv.ParseInt(signedPrice.TimestampedSignature.Signature.V[2:], 16, 8)
			if err != nil {
				return nil, err
			}

			payloads[i].pubSigs[j] = contract_bindings.StorkStructsPublisherSignature{
				PubKey:         pubKeyBytes,
				AssetPairId:    signedPrice.ExternalAssetId,
				Timestamp:      uint64(signedPrice.TimestampedSignature.Timestamp) / 1000000000,
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

func (sci *EvmContractInteractor) BatchPushToContract(priceUpdates map[InternalEncodedAssetId]AggregatedSignedPrice) error {
	if sci.verifyPublishers {
		publisherVerifyPayloads, err := getVerifyPublishersPayloads(priceUpdates)
		if err != nil {
			return err
		}
		for i := range publisherVerifyPayloads {
			verified, err := sci.contract.VerifyPublisherSignaturesV1(nil, publisherVerifyPayloads[i].pubSigs, publisherVerifyPayloads[i].merkleRoot)
			if err != nil {
				sci.logger.Error().Err(err).Msg("Failed to verify publisher signatures")
				return err
			}
			if !verified {
				sci.logger.Error().Msg("Publisher signatures not verified, skipping update")
				return nil
			}
		}
	}

	updatePayload, err := getUpdatePayload(priceUpdates)
	if err != nil {
		return err
	}

	fee, err := sci.contract.GetUpdateFeeV1(nil, updatePayload)
	if err != nil {
		return err
	}

	auth, err := bind.NewKeyedTransactorWithChainID(sci.privateKey, sci.chainID)
	if err != nil {
		return err
	}

	// let the library auto-estimate the gas price
	auth.GasLimit = 0
	auth.Value = fee

	tx, err := sci.contract.UpdateTemporalNumericValuesV1(auth, updatePayload)
	if err != nil {
		return err
	}

	sci.logger.Info().
		Str("txHash", tx.Hash().Hex()).
		Int("numUpdates", len(updatePayload)).
		Uint64("gasPrice", tx.GasPrice().Uint64()).
		Msg("Pushed new values to contract")
	return nil
}

func (sci *EvmContractInteractor) GetWalletBalance() (float64, error) {
	publicKey := sci.privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return -1, fmt.Errorf("error casting public key to ECDSA")
	}

	address := crypto.PubkeyToAddress(*publicKeyECDSA)

	balance, err := sci.client.BalanceAt(context.Background(), address, nil)
	if err != nil {
		return -1, err
	}
	balanceFloat, _ := balance.Float64()

	return balanceFloat, nil
}
