package evm

import (
	"crypto/ecdsa"
	"math/big"
	"testing"

	"github.com/Stork-Oracle/stork-external/apps/chain_pusher/internal/testutil"
	"github.com/Stork-Oracle/stork-external/apps/chain_pusher/pkg/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadPrivateKey(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		keyFileContent []byte
		expectedPubKey string // We'll verify by checking the derived public key address
		wantError      bool
	}{
		{
			name:           "valid private key",
			keyFileContent: []byte("ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"),
			expectedPubKey: "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266", // Known address for this private key
			wantError:      false,
		},
		{
			name:           "valid private key with newline",
			keyFileContent: []byte("ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80\n"),
			expectedPubKey: "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266",
			wantError:      false,
		},
		{
			name:           "valid private key with spaces and newlines",
			keyFileContent: []byte("  ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80  \n"),
			expectedPubKey: "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266",
			wantError:      false,
		},
		{
			name:           "valid private key with 0x prefix",
			keyFileContent: []byte("0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"),
			wantError:      true, // crypto.HexToECDSA doesn't accept 0x prefix
		},
		{
			name:           "invalid hex string",
			keyFileContent: []byte("invalid hex"),
			wantError:      true,
		},
		{
			name:           "too short private key",
			keyFileContent: []byte("1234"),
			wantError:      true,
		},
		{
			name:           "empty input",
			keyFileContent: []byte(""),
			wantError:      true,
		},
		{
			name:           "only whitespace",
			keyFileContent: []byte("   \n  \t  "),
			wantError:      true,
		},
		{
			name: "private key with invalid characters",
			keyFileContent: []byte(
				"ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff8g",
			), // 'g' is invalid hex
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := loadPrivateKey(tt.keyFileContent)

			if tt.wantError {
				require.Error(t, err)
				assert.Nil(t, result)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)
			assert.IsType(t, &ecdsa.PrivateKey{}, result)

			// Verify the private key by checking the derived address
			publicKey := result.Public()
			publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
			require.True(t, ok)

			address := crypto.PubkeyToAddress(*publicKeyECDSA)
			assert.Equal(t, tt.expectedPubKey, address.Hex())
		})
	}
}

func TestGetUpdatePayload(t *testing.T) {
	t.Parallel()

	tests := testutil.StandardPriceCase()

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()
		})

		priceUpdates := []types.AggregatedSignedPrice{
			tt.Price,
		}

		result, err := getUpdatePayload(priceUpdates)
		if tt.WantError {
			assert.Error(t, err)

			return
		}

		require.NoError(t, err)
		assert.Len(t, result, 1)

		updatePayload := result[0]
		expectedID := tt.PriceBytes.StorkSignedPrice.EncodedAssetID
		expectedTemporalNumericValueTimestampNs := tt.PriceBytes.StorkSignedPrice.TimestampedSignature.TimestampNano
		expectedTemporalNumericValueQuantizedValue := tt.PriceBytes.StorkSignedPrice.QuantizedPrice
		expectedPublisherMerkleRoot := tt.PriceBytes.StorkSignedPrice.PublisherMerkleRoot
		expectedValueComputeAlgHash := tt.PriceBytes.StorkSignedPrice.StorkCalculationAlg
		expectedR := tt.PriceBytes.StorkSignedPrice.TimestampedSignature.Signature.R
		expectedS := tt.PriceBytes.StorkSignedPrice.TimestampedSignature.Signature.S
		expectedV := tt.PriceBytes.StorkSignedPrice.TimestampedSignature.Signature.V

		assert.Equal(t, expectedID, updatePayload.Id)
		assert.Equal(t, expectedTemporalNumericValueTimestampNs, updatePayload.TemporalNumericValue.TimestampNs)
		assert.Equal(t, expectedTemporalNumericValueQuantizedValue, updatePayload.TemporalNumericValue.QuantizedValue)
		assert.Equal(t, expectedPublisherMerkleRoot, updatePayload.PublisherMerkleRoot)
		assert.Equal(t, expectedValueComputeAlgHash, updatePayload.ValueComputeAlgHash)
		assert.Equal(t, expectedR, updatePayload.R)
		assert.Equal(t, expectedS, updatePayload.S)
		assert.Equal(t, expectedV, updatePayload.V)
	}
}

func TestGetBumpedGasPrices(t *testing.T) {
	t.Parallel()

	eci := &ContractInteractor{}

	tests := []struct {
		name               string
		gasPrice           int64
		gasTipCap          int64
		retryCount         int64
		expectedGasFeeCap  int64
		expectedGasTipCap  int64
		expectedMultiplier float64 
	}{
		{
			name:               "retry 1: 1.2x multiplier",
			gasPrice:           1000,
			gasTipCap:          100,
			retryCount:         1,
			expectedGasFeeCap:  1200,
			expectedGasTipCap:  120,
			expectedMultiplier: 1.2,
		},
		{
			name:               "retry 2: 1.44x multiplier",
			gasPrice:           1000,
			gasTipCap:          100,
			retryCount:         2,
			expectedGasFeeCap:  1440,
			expectedGasTipCap:  144,
			expectedMultiplier: 1.44,
		},
		{
			name:               "retry 3: 1.728x multiplier",
			gasPrice:           1000,
			gasTipCap:          100,
			retryCount:         3,
			expectedGasFeeCap:  1728,
			expectedGasTipCap:  172,
			expectedMultiplier: 1.728,
		},
		{
			name:               "large values retry 1",
			gasPrice:           1000000000,
			gasTipCap:          100000000,
			retryCount:         1,
			expectedGasFeeCap:  1200000000,
			expectedGasTipCap:  120000000,
			expectedMultiplier: 1.2,
		},
		{
			name:               "large values retry 2",
			gasPrice:           50000000000,
			gasTipCap:          2000000000,
			retryCount:         2,
			expectedGasFeeCap:  72000000000,
			expectedGasTipCap:  2880000000,
			expectedMultiplier: 1.44,
		},
		{
			name:               "small values with rounding",
			gasPrice:           10,
			gasTipCap:          5,
			retryCount:         3,
			expectedGasFeeCap:  17,
			expectedGasTipCap:  8,
			expectedMultiplier: 1.728,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gasPrice := big.NewInt(tt.gasPrice)
			gasTipCap := big.NewInt(tt.gasTipCap)

			resultGasFeeCap, resultGasTipCap, err := eci.getBumpedGasPrices(
				gasPrice,
				gasTipCap,
				tt.retryCount,
			)

			require.NoError(t, err)
			require.NotNil(t, resultGasFeeCap)
			require.NotNil(t, resultGasTipCap)

			assert.Equal(t, tt.expectedGasFeeCap, resultGasFeeCap.Int64(),
				"gasFeeCap: expected %d * %.3f = %d, got %d",
				tt.gasPrice, tt.expectedMultiplier, tt.expectedGasFeeCap, resultGasFeeCap.Int64())

			assert.Equal(t, tt.expectedGasTipCap, resultGasTipCap.Int64(),
				"gasTipCap: expected %d * %.3f = %d, got %d",
				tt.gasTipCap, tt.expectedMultiplier, tt.expectedGasTipCap, resultGasTipCap.Int64())
		})
	}
}
