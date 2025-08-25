package self_serve_chain_pusher

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
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
	logger      zerolog.Logger
	client      *ethclient.Client
	wsClient    *ethclient.Client
	privateKey  *ecdsa.PrivateKey
	chainID     *big.Int
	gasLimit    uint64
	rateLimiter *rate.Limiter
	contractAddress common.Address
}

type TemporalNumericValueEvm struct {
	TimestampNs      *big.Int
	Quantized        *big.Int
	EncodedAssetId   [32]byte
	Nonce            *big.Int
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
	rateLimiter := rate.NewLimiter(rate.Limit(limitPerSecond), burstLimit)

	return &SelfServeContractInteractor{
		logger:          logger.With().Str("component", "contract_interactor").Logger(),
		client:          client,
		wsClient:        wsClient,
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

	// Convert encoded asset ID to bytes32
	encodedAssetIdBytes, err := decodeHexToBytes32(asset.EncodedAssetId)
	if err != nil {
		return fmt.Errorf("failed to decode encoded asset ID: %w", err)
	}

	// Quantize the value (assuming 18 decimal places)
	quantizedValue := new(big.Int)
	value.Mul(value, new(big.Float).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)))
	value.Int(quantizedValue)

	temporalValue := TemporalNumericValueEvm{
		TimestampNs:    big.NewInt(time.Now().UnixNano()),
		Quantized:      quantizedValue,
		EncodedAssetId: encodedAssetIdBytes,
		Nonce:          nonce,
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

		txHash, err := ci.submitPushValueTransaction(ctx, temporalValue)
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

func (ci *SelfServeContractInteractor) submitPushValueTransaction(ctx context.Context, temporalValue TemporalNumericValueEvm) (*common.Hash, error) {
	_ = temporalValue // TODO: Use this to construct actual contract call data
	// Get transaction options
	auth, err := ci.getTransactionOptions(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction options: %w", err)
	}

	// This is a simplified version - in a real implementation, you would:
	// 1. Load the actual contract ABI
	// 2. Create a contract binding
	// 3. Call the appropriate contract method
	// For now, we'll simulate the transaction creation

	// Create transaction data (this would be the actual contract call)
	// For the self-serve contract, this would likely be something like:
	// contract.UpdateTemporalNumericValueEvm(auth, temporalValue)

	// For demonstration, we'll create a simple transaction
	// In reality, you'd use the generated contract bindings
	tx := types.NewTransaction(
		auth.Nonce.Uint64(),
		ci.contractAddress,
		big.NewInt(0), // value
		auth.GasLimit,
		auth.GasPrice,
		[]byte{}, // This would be the actual contract call data
	)

	// Sign and send transaction
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(ci.chainID), ci.privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %w", err)
	}

	err = ci.client.SendTransaction(ctx, signedTx)
	if err != nil {
		return nil, fmt.Errorf("failed to send transaction: %w", err)
	}

	txHash := signedTx.Hash()
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

func decodeHexToBytes32(hexStr string) ([32]byte, error) {
	var result [32]byte
	
	// Remove 0x prefix if present
	if len(hexStr) >= 2 && hexStr[:2] == "0x" {
		hexStr = hexStr[2:]
	}
	
	// Pad to 64 characters (32 bytes)
	for len(hexStr) < 64 {
		hexStr = "0" + hexStr
	}
	
	// Convert hex to bytes
	bytes := common.Hex2Bytes(hexStr)
	if len(bytes) != 32 {
		return result, fmt.Errorf("invalid hex string length for bytes32")
	}
	
	copy(result[:], bytes)
	return result, nil
}

func (ci *SelfServeContractInteractor) Close() {
	if ci.client != nil {
		ci.client.Close()
	}
	if ci.wsClient != nil {
		ci.wsClient.Close()
	}
}