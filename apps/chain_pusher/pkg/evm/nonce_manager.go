package evm

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

type NonceManagerI interface {
	GetLatestNonce(ctx context.Context, ethClient *ethclient.Client, address common.Address) (*big.Int, error)
	IncrementNonce(ctx context.Context, ethClient *ethclient.Client, address common.Address) error
	ResetNonce(ctx context.Context, ethClient *ethclient.Client, address common.Address) error
}

type NoopNonceManager struct{}

func NewNoopNonceManager() *NoopNonceManager {
	return &NoopNonceManager{}
}

func (n *NoopNonceManager) GetLatestNonce(ctx context.Context, ethClient *ethclient.Client, address common.Address) (*big.Int, error) {
	return nil, nil
}

func (n *NoopNonceManager) IncrementNonce(ctx context.Context, ethClient *ethclient.Client, address common.Address) error {
	return nil
}

func (n *NoopNonceManager) ResetNonce(ctx context.Context, ethClient *ethclient.Client, address common.Address) error {
	return nil
}

type ServerNonceManager struct {
	usePendingNonce bool
}

func NewServerNonceManager(usePendingNonce bool) *ServerNonceManager {
	return &ServerNonceManager{
		usePendingNonce: usePendingNonce,
	}
}

func (n *ServerNonceManager) GetLatestNonce(ctx context.Context, ethClient *ethclient.Client, address common.Address) (*big.Int, error) {
	if n.usePendingNonce {
		nonce, err := ethClient.PendingNonceAt(ctx, address)
		if err != nil {
			return nil, fmt.Errorf("failed to get latest nonce: %w", err)
		}
		return new(big.Int).SetUint64(nonce), nil
	}
	nonce, err := ethClient.NonceAt(ctx, address, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest nonce: %w", err)
	}
	return new(big.Int).SetUint64(nonce), nil
}

// noop since the nonce is managed by the server
func (n *ServerNonceManager) IncrementNonce(ctx context.Context, ethClient *ethclient.Client, address common.Address) error {
	return nil
}

// noop since the nonce is managed by the server
func (n *ServerNonceManager) ResetNonce(ctx context.Context, ethClient *ethclient.Client, address common.Address) error {
	return nil
}

type LocalNonceManager struct {
	nonce *big.Int
}

func NewLocalNonceManager() *LocalNonceManager {
	return &LocalNonceManager{
		nonce: nil,
	}
}

func (n *LocalNonceManager) GetLatestNonce(ctx context.Context, ethClient *ethclient.Client, address common.Address) (*big.Int, error) {
	if n.nonce == nil {
		nonce, err := ethClient.NonceAt(ctx, address, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to get latest nonce: %w", err)
		}
		n.nonce = new(big.Int).SetUint64(nonce)
	}
	return n.nonce, nil
}

func (n *LocalNonceManager) IncrementNonce(ctx context.Context, ethClient *ethclient.Client, address common.Address) error {
	n.nonce = new(big.Int).Add(n.nonce, big.NewInt(1))
	return nil
}

func (n *LocalNonceManager) ResetNonce(ctx context.Context, ethClient *ethclient.Client, address common.Address) error {
	n.nonce = nil
	return nil
}
