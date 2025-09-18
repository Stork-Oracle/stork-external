package self_serve_evm

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rs/zerolog"

	"github.com/Stork-Oracle/stork-external/apps/chain_pusher/pkg/pusher"
	publisher_agent "github.com/Stork-Oracle/stork-external/apps/publisher_agent/pkg"
	"github.com/Stork-Oracle/stork-external/apps/self_serve_chain_pusher/pkg/evm/bindings"
	"github.com/Stork-Oracle/stork-external/apps/self_serve_chain_pusher/pkg/types"
	"github.com/Stork-Oracle/stork-external/shared"
)

const (
	maxRetryAttempts         = 5
	initialBackoff           = 1 * time.Second
	exponentialBackoffFactor = 1.5
)

type ContractInteractor struct {
	logger zerolog.Logger

	contract   *bindings.SelfServeStorkContract
	wsContract *bindings.SelfServeStorkContract
	client     *ethclient.Client
	wsClient   *ethclient.Client

	privateKey *ecdsa.PrivateKey
	chainID    *big.Int
	gasLimit   uint64
}

func NewContractInteractor(
	rpcUrl string,
	wsUrl string,
	contractAddr string,
	privateKey *ecdsa.PrivateKey,
	gasLimit uint64,
	logger zerolog.Logger,
) (*ContractInteractor, error) {
	client, err := ethclient.Dial(rpcUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RPC: %w", err)
	}

	var wsClient *ethclient.Client
	if wsUrl != "" {
		wsClient, err = ethclient.Dial(wsUrl)
		if err != nil {
			logger.Warn().Err(err).Msg("Failed to connect to WebSocket, using HTTP only")
		} else {
			logger.Info().Msg("Connected to WebSocket endpoint")
		}
	}

	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get chain ID: %w", err)
	}

	contractAddress := common.HexToAddress(contractAddr)

	contract, err := bindings.NewSelfServeStorkContract(contractAddress, client)
	if err != nil {
		return nil, fmt.Errorf("failed to create contract instance: %w", err)
	}

	var wsContract *bindings.SelfServeStorkContract
	if wsClient != nil {
		wsContract, err = bindings.NewSelfServeStorkContract(contractAddress, wsClient)
		if err != nil {
			logger.Warn().Err(err).Msg("Failed to create WebSocket contract instance")
		}
	}

	return &ContractInteractor{
		logger: logger.With().Str("component", "contract_interactor").Logger(),

		contract:   contract,
		wsContract: wsContract,
		client:     client,
		wsClient:   wsClient,
		privateKey: privateKey,
		chainID:    chainID,
		gasLimit:   gasLimit,
	}, nil
}

func (ci *ContractInteractor) PushSignedPriceUpdate(ctx context.Context, asset types.AssetPushConfig, signedPriceUpdate publisher_agent.SignedPriceUpdate[*shared.EvmSignature]) error {
	ci.logger.Info().
		Str("asset", string(signedPriceUpdate.AssetID)).
		Str("price", string(signedPriceUpdate.SignedPrice.QuantizedPrice)).
		Str("encoded_asset_id", string(asset.EncodedAssetID)).
		Msg("Pushing signed price update to self-serve contract")

	// Convert the signed price update to contract input
	updateInput, err := ci.convertSignedPriceUpdateToInput(signedPriceUpdate, asset)
	if err != nil {
		return fmt.Errorf("failed to convert signed price update: %w", err)
	}

	// Retry logic for transaction submission
	var lastErr error

	backoff := initialBackoff

	for attempt := range maxRetryAttempts {
		if attempt > 0 {
			ci.logger.Warn().
				Int("attempt", attempt+1).
				Dur("backoff", backoff).
				Err(lastErr).
				Msg("Retrying push signed price update transaction")
			time.Sleep(backoff)
			backoff = time.Duration(float64(backoff) * exponentialBackoffFactor)
		}

		txHash, err := ci.submitPushValueTransaction(ctx, []bindings.SelfServeStorkStructsPublisherTemporalNumericValueInput{updateInput})
		if err != nil {
			lastErr = err

			continue
		}

		ci.logger.Info().
			Str("asset", string(signedPriceUpdate.AssetID)).
			Str("tx_hash", txHash.Hex()).
			Msg("Successfully submitted signed price update transaction")
		return nil
	}

	return fmt.Errorf("failed to push signed price update after %d attempts: %w", maxRetryAttempts, lastErr)
}

// TODO: this is not in our existing pushers
func (ci *ContractInteractor) Close() {
	if ci.client != nil {
		ci.client.Close()
	}

	if ci.wsClient != nil {
		ci.wsClient.Close()
	}
}

func (ci *ContractInteractor) convertSignedPriceUpdateToInput(
	signedPriceUpdate publisher_agent.SignedPriceUpdate[*shared.EvmSignature],
	asset types.AssetPushConfig,
) (bindings.SelfServeStorkStructsPublisherTemporalNumericValueInput, error) {
	// Convert quantized price to big.Int
	quantizedValue, success := new(big.Int).SetString(string(signedPriceUpdate.SignedPrice.QuantizedPrice), 10)
	if !success {
		return bindings.SelfServeStorkStructsPublisherTemporalNumericValueInput{},
			fmt.Errorf("%w: %s", shared.ErrFailedToConvertQuantizedPriceToBigInt, signedPriceUpdate.SignedPrice.QuantizedPrice)
	}

	// Create the temporal numeric value using the signed data timestamp
	temporalValue := bindings.SelfServeStorkStructsTemporalNumericValue{
		TimestampNs:    signedPriceUpdate.SignedPrice.TimestampedSignature.TimestampNano,
		QuantizedValue: quantizedValue,
	}

	// Parse the publisher key
	pubKeyBytes, err := hex.DecodeString(strings.TrimPrefix(string(signedPriceUpdate.SignedPrice.PublisherKey), "0x"))
	if err != nil {
		return bindings.SelfServeStorkStructsPublisherTemporalNumericValueInput{},
			fmt.Errorf("failed to decode publisher key: %w", err)
	}

	var pubKeyAddress common.Address
	copy(pubKeyAddress[:], pubKeyBytes)

	// Parse the signature components
	rBytes, err := hex.DecodeString(strings.TrimPrefix(signedPriceUpdate.SignedPrice.TimestampedSignature.Signature.R, "0x"))
	if err != nil {
		return bindings.SelfServeStorkStructsPublisherTemporalNumericValueInput{},
			fmt.Errorf("failed to decode signature R: %w", err)
	}
	sBytes, err := hex.DecodeString(strings.TrimPrefix(signedPriceUpdate.SignedPrice.TimestampedSignature.Signature.S, "0x"))
	if err != nil {
		return bindings.SelfServeStorkStructsPublisherTemporalNumericValueInput{},
			fmt.Errorf("failed to decode signature S: %w", err)
	}
	vBytes, err := hex.DecodeString(strings.TrimPrefix(signedPriceUpdate.SignedPrice.TimestampedSignature.Signature.V, "0x"))
	if err != nil {
		return bindings.SelfServeStorkStructsPublisherTemporalNumericValueInput{},
			fmt.Errorf("failed to decode signature V: %w", err)
	}

	// Convert to fixed-size arrays
	var r, s [32]byte

	copy(r[:], rBytes)
	copy(s[:], sBytes)

	v := vBytes[0] // V is a single byte

	return bindings.SelfServeStorkStructsPublisherTemporalNumericValueInput{
		TemporalNumericValue: temporalValue,
		PubKey:               pubKeyAddress,
		AssetPairId:          string(asset.AssetID),
		R:                    r,
		S:                    s,
		V:                    v,
	}, nil
}

func (ci *ContractInteractor) submitPushValueTransaction(
	ctx context.Context,
	updateData []bindings.SelfServeStorkStructsPublisherTemporalNumericValueInput,
) (*common.Hash, error) {
	// Get transaction options
	auth, err := ci.getTransactionOptions(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction options: %w", err)
	}

	// Call the contract's UpdateTemporalNumericValues method
	// storeHistoric is set to false for basic functionality
	tx, err := ci.contract.UpdateTemporalNumericValues(auth, updateData, false)
	if err != nil {
		return nil, fmt.Errorf("failed to call UpdateTemporalNumericValues: %w", err)
	}

	txHash := tx.Hash()
	return &txHash, nil
}

func (ci *ContractInteractor) getTransactionOptions(ctx context.Context) (*bind.TransactOpts, error) {
	nonce, err := ci.client.PendingNonceAt(ctx, crypto.PubkeyToAddress(ci.privateKey.PublicKey))
	if err != nil {
		return nil, fmt.Errorf("failed to get nonce: %w", err)
	}

	gasPrice, err := ci.client.SuggestGasPrice(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get gas price: %w", err)
	}

	auth, err := bind.NewKeyedTransactorWithChainID(ci.privateKey, ci.chainID)
	if err != nil {
		return nil, fmt.Errorf("failed to create transactor: %w", err)
	}

	nonceInt, err := pusher.SafeUint64ToInt64(nonce)
	if err != nil {
		return nil, fmt.Errorf("failed to convert nonce to uint64: %w", err)
	}

	auth.Nonce = big.NewInt(nonceInt)
	auth.Value = big.NewInt(0)
	auth.GasLimit = ci.gasLimit
	auth.GasPrice = gasPrice
	auth.Context = ctx

	return auth, nil
}
