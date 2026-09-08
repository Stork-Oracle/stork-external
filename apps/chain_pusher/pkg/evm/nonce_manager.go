package evm

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"sync"

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

func (n *NoopNonceManager) GetLatestNonce(
	ctx context.Context,
	ethClient *ethclient.Client,
	address common.Address,
) (*big.Int, error) {
	//nolint:nilnil // a noop manager has no nonce to report and no error to raise
	return nil, nil
}

func (n *NoopNonceManager) IncrementNonce(
	ctx context.Context,
	ethClient *ethclient.Client,
	address common.Address,
) error {
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

func (n *ServerNonceManager) GetLatestNonce(
	ctx context.Context,
	ethClient *ethclient.Client,
	address common.Address,
) (*big.Int, error) {
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

// IncrementNonce is a noop since the nonce is managed by the server.
func (n *ServerNonceManager) IncrementNonce(
	ctx context.Context,
	ethClient *ethclient.Client,
	address common.Address,
) error {
	return nil
}

// ResetNonce is a noop since the nonce is managed by the server.
func (n *ServerNonceManager) ResetNonce(
	ctx context.Context,
	ethClient *ethclient.Client,
	address common.Address,
) error {
	return nil
}

type LocalNonceManager struct {
	mu    sync.Mutex
	nonce *big.Int
}

func NewLocalNonceManager() *LocalNonceManager {
	return &LocalNonceManager{
		mu:    sync.Mutex{},
		nonce: nil,
	}
}

func (n *LocalNonceManager) GetLatestNonce(
	ctx context.Context,
	ethClient *ethclient.Client,
	address common.Address,
) (*big.Int, error) {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.nonce == nil {
		nonce, err := ethClient.NonceAt(ctx, address, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to get latest nonce: %w", err)
		}

		n.nonce = new(big.Int).SetUint64(nonce)
	}

	return new(big.Int).Set(n.nonce), nil
}

func (n *LocalNonceManager) IncrementNonce(
	ctx context.Context,
	ethClient *ethclient.Client,
	address common.Address,
) error {
	n.mu.Lock()
	defer n.mu.Unlock()
	// a concurrent ResetNonce may have cleared the nonce; leave it nil so the
	// next GetLatestNonce re-fetches from the chain
	if n.nonce == nil {
		return nil
	}

	n.nonce = new(big.Int).Add(n.nonce, big.NewInt(1))

	return nil
}

func (n *LocalNonceManager) ResetNonce(ctx context.Context, ethClient *ethclient.Client, address common.Address) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.nonce = nil

	return nil
}

type NonceManagerType string

const (
	NonceManagerTypeNoop          NonceManagerType = "noop"
	NonceManagerTypeServer        NonceManagerType = "server"
	NonceManagerTypeServerPending NonceManagerType = "serverPending"
	NonceManagerTypeLocal         NonceManagerType = "local"
)

var ErrUnknownNonceManagerType = errors.New("unknown nonce manager type")

func NewNonceManagerFromType(t NonceManagerType) (NonceManagerI, error) {
	switch t {
	case NonceManagerTypeNoop, "":
		return NewNoopNonceManager(), nil
	case NonceManagerTypeServer:
		return NewServerNonceManager(false), nil
	case NonceManagerTypeServerPending:
		return NewServerNonceManager(true), nil
	case NonceManagerTypeLocal:
		return NewLocalNonceManager(), nil
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnknownNonceManagerType, string(t))
	}
}
