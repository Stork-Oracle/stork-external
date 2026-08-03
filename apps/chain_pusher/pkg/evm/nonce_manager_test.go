package evm

import (
	"context"
	"math/big"
	"sync"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLocalNonceManagerIncrementAfterReset(t *testing.T) {
	t.Parallel()

	nm := NewLocalNonceManager()
	nm.nonce = big.NewInt(5)

	require.NoError(t, nm.ResetNonce(context.Background(), nil, common.Address{}))
	// nonce is nil after reset; increment must not panic and must leave it nil
	// so the next GetLatestNonce re-fetches from the chain
	require.NoError(t, nm.IncrementNonce(context.Background(), nil, common.Address{}))
	assert.Nil(t, nm.nonce)
}

func TestLocalNonceManagerIncrement(t *testing.T) {
	t.Parallel()

	nm := NewLocalNonceManager()
	nm.nonce = big.NewInt(5)

	require.NoError(t, nm.IncrementNonce(context.Background(), nil, common.Address{}))
	assert.Equal(t, big.NewInt(6), nm.nonce)
}

func TestLocalNonceManagerConcurrentAccess(t *testing.T) {
	t.Parallel()

	nm := NewLocalNonceManager()
	nm.nonce = big.NewInt(0)

	ctx := context.Background()
	addr := common.Address{}

	var wg sync.WaitGroup
	for range 100 {
		wg.Add(3)

		go func() {
			defer wg.Done()

			assert.NoError(t, nm.IncrementNonce(ctx, nil, addr))
		}()

		go func() {
			defer wg.Done()

			assert.NoError(t, nm.ResetNonce(ctx, nil, addr))
		}()

		go func() {
			defer wg.Done()

			// simulate the read side of GetLatestNonce with a pre-populated nonce
			nm.mu.Lock()
			if nm.nonce == nil {
				nm.nonce = big.NewInt(0)
			}
			_ = new(big.Int).Set(nm.nonce)
			nm.mu.Unlock()
		}()
	}

	wg.Wait()
}
