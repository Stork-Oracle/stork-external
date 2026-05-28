package evm

import (
	"context"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLocalNonceManagerIncrementAndReset(t *testing.T) {
	t.Parallel()

	manager := NewLocalNonceManager()
	manager.nonce = big.NewInt(7)

	err := manager.IncrementNonce(context.Background(), nil, common.Address{})
	require.NoError(t, err)
	assert.Equal(t, int64(8), manager.nonce.Int64())

	err = manager.ResetNonce(context.Background(), nil, common.Address{})
	require.NoError(t, err)
	assert.Nil(t, manager.nonce)
}
