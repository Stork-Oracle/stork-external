package self_serve_evm

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/Stork-Oracle/stork-external/apps/self_serve_chain_pusher/pkg/evm/bindings"
	"github.com/Stork-Oracle/stork-external/shared"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rs/zerolog"

	publisher_agent "github.com/Stork-Oracle/stork-external/apps/publisher_agent/pkg"
)

const (
	maxRetryAttempts         = 5
	initialBackoff           = 1 * time.Second
	exponentialBackoffFactor = 1.5
)

type SelfServeContractInteractor struct {
	logger          zerolog.Logger
	client          *ethclient.Client
	wsClient        *ethclient.Client
	contract        *bindings.SelfServeStorkContract
	wsContract      *bindings.SelfServeStorkContract
	privateKey      *ecdsa.PrivateKey
	chainID         *big.Int
	gasLimit        uint64
	contractAddress common.Address
}

func NewSelfServeContractInteractor(
	rpcUrl string,
	wsUrl string,
	contractAddr string,
	privateKey *ecdsa.PrivateKey,
	gasLimit uint64,
	logger zerolog.Logger,
) (*SelfServeContractInteractor, error) {
	client, err := ethclient.Dial(rpcUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RPC: %w", err)
	}

	var wsClient *ethclient.Client
	if wsUrl != "" {
		wsClient, err = ethclient.Dial(wsUrl)
		if err != nil {
			logger.Warn().Err(err).Msg("Failed to connect to WebSocket, using HTTP only")
		}
	}

	chainID, err := client.ChainID(context.Background())
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

	return &SelfServeContractInteractor{
		logger:          logger.With().Str("component", "contract_interactor").Logger(),
		client:          client,
		wsClient:        wsClient,
		contract:        contract,
		wsContract:      wsContract,
		privateKey:      privateKey,
		chainID:         chainID,
		gasLimit:        gasLimit,
		contractAddress: contractAddress,
	}, nil
}

func (ci *SelfServeContractInteractor) PushValue(ctx context.Context, asset AssetPushConfig, value *big.Float, nonce *big.Int) error {
	ci.logger.Info().
		Str("asset", string(asset.AssetID)).
		Str("value", value.Text('f', 6)).
		Str("encoded_asset_id", string(asset.EncodedAssetID)).
		Msg("Pushing value to self-serve contract")

	// Quantize the value (assuming 18 decimal places for int192)
	quantizedValue := new(big.Int)
	scaledValue := new(big.Float).Mul(value, new(big.Float).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)))
	scaledValue.Int(quantizedValue)

	// Create the temporal numeric value
	temporalValue := bindings.SelfServeStorkStructsTemporalNumericValue{
		TimestampNs:    uint64(time.Now().UnixNano()),
		QuantizedValue: quantizedValue,
	}

	// Create the update input with signature data
	// For self-serve, we sign the data ourselves
	updateInput, err := ci.createUpdateInput(temporalValue, asset, nonce)
	if err != nil {
		return fmt.Errorf("failed to create update input: %w", err)
	}

	// Retry logic for transaction submission
	var lastErr error
	backoff := initialBackoff

	for attempt := 0; attempt < maxRetryAttempts; attempt++ {
		if attempt > 0 {
			ci.logger.Warn().
				Int("attempt", attempt+1).
				Dur("backoff", backoff).
				Err(lastErr).
				Msg("Retrying push value transaction")
			time.Sleep(backoff)
			backoff = time.Duration(float64(backoff) * exponentialBackoffFactor)
		}

		txHash, err := ci.submitPushValueTransaction(ctx, []bindings.SelfServeStorkStructsPublisherTemporalNumericValueInput{updateInput})
		if err != nil {
			lastErr = err
			continue
		}

		ci.logger.Info().
			Str("asset", string(asset.AssetID)).
			Str("tx_hash", txHash.Hex()).
			Msg("Successfully submitted push value transaction")
		return nil
	}

	return fmt.Errorf("failed to push value after %d attempts: %w", maxRetryAttempts, lastErr)
}

func (ci *SelfServeContractInteractor) PushSignedPriceUpdate(ctx context.Context, asset AssetPushConfig, signedPriceUpdate publisher_agent.SignedPriceUpdate[*shared.EvmSignature]) error {
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

	for attempt := 0; attempt < maxRetryAttempts; attempt++ {
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

func (ci *SelfServeContractInteractor) convertSignedPriceUpdateToInput(
	signedPriceUpdate publisher_agent.SignedPriceUpdate[*shared.EvmSignature],
	asset AssetPushConfig,
) (bindings.SelfServeStorkStructsPublisherTemporalNumericValueInput, error) {
	// Convert quantized price to big.Int
	quantizedValue, success := new(big.Int).SetString(string(signedPriceUpdate.SignedPrice.QuantizedPrice), 10)
	if !success {
		return bindings.SelfServeStorkStructsPublisherTemporalNumericValueInput{},
			fmt.Errorf("failed to convert quantized price to big.Int: %s", signedPriceUpdate.SignedPrice.QuantizedPrice)
	}

	// Create the temporal numeric value using the signed data timestamp
	temporalValue := bindings.SelfServeStorkStructsTemporalNumericValue{
		TimestampNs:    uint64(signedPriceUpdate.SignedPrice.TimestampedSignature.TimestampNano),
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
	var r [32]byte
	var s [32]byte
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

func (ci *SelfServeContractInteractor) createUpdateInput(
	temporalValue bindings.SelfServeStorkStructsTemporalNumericValue,
	asset AssetPushConfig,
	nonce *big.Int,
) (bindings.SelfServeStorkStructsPublisherTemporalNumericValueInput, error) {
	// Get the public key address from the private key
	pubKeyAddress := crypto.PubkeyToAddress(ci.privateKey.PublicKey)

	// Create the message hash for signing
	// This should match the contract's expected signing format
	messageHash, err := ci.createMessageHash(temporalValue, asset.AssetID, nonce)
	if err != nil {
		return bindings.SelfServeStorkStructsPublisherTemporalNumericValueInput{}, fmt.Errorf("failed to create message hash: %w", err)
	}

	// Sign the message hash
	signature, err := crypto.Sign(messageHash, ci.privateKey)
	if err != nil {
		return bindings.SelfServeStorkStructsPublisherTemporalNumericValueInput{}, fmt.Errorf("failed to sign message: %w", err)
	}

	// Extract r, s, v from signature
	r := [32]byte{}
	s := [32]byte{}
	copy(r[:], signature[0:32])
	copy(s[:], signature[32:64])
	v := signature[64] + 27 // Ethereum v adjustment

	return bindings.SelfServeStorkStructsPublisherTemporalNumericValueInput{
		TemporalNumericValue: temporalValue,
		PubKey:               pubKeyAddress,
		AssetPairId:          string(asset.AssetID),
		R:                    r,
		S:                    s,
		V:                    v,
	}, nil
}

func (ci *SelfServeContractInteractor) createMessageHash(
	temporalValue bindings.SelfServeStorkStructsTemporalNumericValue,
	assetID shared.AssetID,
	nonce *big.Int,
) ([]byte, error) {
	// Create a message hash that matches what the contract expects
	// This is a simplified version - in production, you'd want to match
	// the exact EIP-712 structure that the contract validates
	message := fmt.Sprintf("%s:%d:%s:%s",
		assetID,
		temporalValue.TimestampNs,
		temporalValue.QuantizedValue.String(),
		nonce.String(),
	)

	hash := crypto.Keccak256Hash([]byte(message))
	return hash.Bytes(), nil
}

func (ci *SelfServeContractInteractor) submitPushValueTransaction(
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

func (ci *SelfServeContractInteractor) getTransactionOptions(ctx context.Context) (*bind.TransactOpts, error) {
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

	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)
	auth.GasLimit = ci.gasLimit
	auth.GasPrice = gasPrice
	auth.Context = ctx

	return auth, nil
}

func (ci *SelfServeContractInteractor) Close() {
	if ci.client != nil {
		ci.client.Close()
	}
	if ci.wsClient != nil {
		ci.wsClient.Close()
	}
}
