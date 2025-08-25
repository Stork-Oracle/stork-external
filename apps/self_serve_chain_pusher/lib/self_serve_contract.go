package self_serve_chain_pusher

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"time"

	contract_bindings "github.com/Stork-Oracle/stork-external/apps/self_serve_chain_pusher/lib/contract_bindings/evm"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rs/zerolog"
	"golang.org/x/time/rate"
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
	contract        *contract_bindings.SelfServeStorkContract
	wsContract      *contract_bindings.SelfServeStorkContract
	privateKey      *ecdsa.PrivateKey
	chainID         *big.Int
	gasLimit        uint64
	rateLimiter     *rate.Limiter
	contractAddress common.Address
}


func NewSelfServeContractInteractor(
	rpcUrl string,
	wsUrl string,
	contractAddr string,
	privateKey *ecdsa.PrivateKey,
	gasLimit uint64,
	limitPerSecond float64,
	burstLimit int,
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
	contract, err := contract_bindings.NewSelfServeStorkContract(contractAddress, client)
	if err != nil {
		return nil, fmt.Errorf("failed to create contract instance: %w", err)
	}

	var wsContract *contract_bindings.SelfServeStorkContract
	if wsClient != nil {
		wsContract, err = contract_bindings.NewSelfServeStorkContract(contractAddress, wsClient)
		if err != nil {
			logger.Warn().Err(err).Msg("Failed to create WebSocket contract instance")
		}
	}

	rateLimiter := rate.NewLimiter(rate.Limit(limitPerSecond), burstLimit)

	return &SelfServeContractInteractor{
		logger:          logger.With().Str("component", "contract_interactor").Logger(),
		client:          client,
		wsClient:        wsClient,
		contract:        contract,
		wsContract:      wsContract,
		privateKey:      privateKey,
		chainID:         chainID,
		gasLimit:        gasLimit,
		rateLimiter:     rateLimiter,
		contractAddress: contractAddress,
	}, nil
}

func (ci *SelfServeContractInteractor) PushValue(ctx context.Context, asset AssetPushConfig, value *big.Float, nonce *big.Int) error {
	ci.logger.Info().
		Str("asset", asset.AssetId).
		Str("value", value.Text('f', 6)).
		Str("encoded_asset_id", asset.EncodedAssetId).
		Msg("Pushing value to self-serve contract")

	// Wait for rate limiter
	if err := ci.rateLimiter.Wait(ctx); err != nil {
		return fmt.Errorf("rate limiter error: %w", err)
	}

	// Quantize the value (assuming 18 decimal places for int192)
	quantizedValue := new(big.Int)
	scaledValue := new(big.Float).Mul(value, new(big.Float).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)))
	scaledValue.Int(quantizedValue)

	// Create the temporal numeric value
	temporalValue := contract_bindings.SelfServeStorkStructsTemporalNumericValue{
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

		txHash, err := ci.submitPushValueTransaction(ctx, []contract_bindings.SelfServeStorkStructsPublisherTemporalNumericValueInput{updateInput})
		if err != nil {
			lastErr = err
			continue
		}

		ci.logger.Info().
			Str("asset", asset.AssetId).
			Str("tx_hash", txHash.Hex()).
			Msg("Successfully submitted push value transaction")
		return nil
	}

	return fmt.Errorf("failed to push value after %d attempts: %w", maxRetryAttempts, lastErr)
}

func (ci *SelfServeContractInteractor) createUpdateInput(
	temporalValue contract_bindings.SelfServeStorkStructsTemporalNumericValue,
	asset AssetPushConfig,
	nonce *big.Int,
) (contract_bindings.SelfServeStorkStructsPublisherTemporalNumericValueInput, error) {
	// Get the public key address from the private key
	pubKeyAddress := crypto.PubkeyToAddress(ci.privateKey.PublicKey)

	// Create the message hash for signing
	// This should match the contract's expected signing format
	messageHash, err := ci.createMessageHash(temporalValue, asset.AssetId, nonce)
	if err != nil {
		return contract_bindings.SelfServeStorkStructsPublisherTemporalNumericValueInput{}, fmt.Errorf("failed to create message hash: %w", err)
	}

	// Sign the message hash
	signature, err := crypto.Sign(messageHash, ci.privateKey)
	if err != nil {
		return contract_bindings.SelfServeStorkStructsPublisherTemporalNumericValueInput{}, fmt.Errorf("failed to sign message: %w", err)
	}

	// Extract r, s, v from signature
	r := [32]byte{}
	s := [32]byte{}
	copy(r[:], signature[0:32])
	copy(s[:], signature[32:64])
	v := signature[64] + 27 // Ethereum v adjustment

	return contract_bindings.SelfServeStorkStructsPublisherTemporalNumericValueInput{
		TemporalNumericValue: temporalValue,
		PubKey:               pubKeyAddress,
		AssetPairId:          asset.AssetId,
		R:                    r,
		S:                    s,
		V:                    v,
	}, nil
}

func (ci *SelfServeContractInteractor) createMessageHash(
	temporalValue contract_bindings.SelfServeStorkStructsTemporalNumericValue,
	assetId string,
	nonce *big.Int,
) ([]byte, error) {
	// Create a message hash that matches what the contract expects
	// This is a simplified version - in production, you'd want to match
	// the exact EIP-712 structure that the contract validates
	message := fmt.Sprintf("%s:%d:%s:%s",
		assetId,
		temporalValue.TimestampNs,
		temporalValue.QuantizedValue.String(),
		nonce.String(),
	)

	hash := crypto.Keccak256Hash([]byte(message))
	return hash.Bytes(), nil
}

func (ci *SelfServeContractInteractor) submitPushValueTransaction(
	ctx context.Context,
	updateData []contract_bindings.SelfServeStorkStructsPublisherTemporalNumericValueInput,
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

	// If gasLimit is 0, we would estimate gas here
	if ci.gasLimit == 0 {
		// Gas estimation would be done here for the specific contract call
		auth.GasLimit = 200000 // Default fallback
	}

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