package aptos

import (
	"math/big"
	"testing"

	"github.com/Stork-Oracle/stork-external/apps/chain_pusher/pkg/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadPrivateKey(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		keyFileContent []byte
		expectedPubKey string
		wantError      bool
	}{
		{
			name:           "valid private key",
			keyFileContent: []byte("8a7b8c1d2e3f4a5b6c7d8e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b"),
			expectedPubKey: "0x86f0cdc8814eee42f6ced3d82efd52e1b2a8d57210a21e59524457c85d3d6cb3",
			wantError:      false,
		},
		{
			name:           "valid private key with newline",
			keyFileContent: []byte("8a7b8c1d2e3f4a5b6c7d8e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b\n"),
			expectedPubKey: "0x86f0cdc8814eee42f6ced3d82efd52e1b2a8d57210a21e59524457c85d3d6cb3",
			wantError:      false,
		},
		{
			name:           "invalid private key format",
			keyFileContent: []byte("invalid_key"),
			expectedPubKey: "",
			wantError:      true,
		},
		{
			name:           "empty content",
			keyFileContent: []byte(""),
			expectedPubKey: "",
			wantError:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := loadPrivateKey(tt.keyFileContent)

			if tt.wantError {
				assert.Error(t, err)

				return
			}

			pubKey := result.PubKey()
			assert.Equal(t, tt.expectedPubKey, pubKey.ToHex())
			require.NoError(t, err)
			assert.NotNil(t, result)
		})
	}
}

func TestAggregatedSignedPriceToAptosUpdateData(t *testing.T) {
	t.Parallel()

	tests := testutil.StandardPriceCase()

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			result, err := aggregatedSignedPriceToUpdateData(tt.Price)

			if tt.WantError {
				assert.Error(t, err)

				return
			}

			require.NoError(t, err)

			expectedID := tt.PriceBytes.StorkSignedPrice.EncodedAssetID[:]
			expectedTemporalNumericValueTimestampNs := tt.PriceBytes.StorkSignedPrice.TimestampedSignature.TimestampNano
			expectedTemporalNumericValueQuantizedValueMagnitude := new(
				big.Int,
			).Abs(tt.PriceBytes.StorkSignedPrice.QuantizedPrice)
			expectedTemporalNumericValueQuantizedValueNegative := tt.PriceBytes.StorkSignedPrice.QuantizedPrice.Sign() == -1
			expectedTemporalNumericValuePublisherMerkleRoot := tt.PriceBytes.StorkSignedPrice.PublisherMerkleRoot[:]
			expectedTemporalNumericValueValueComputeAlgHash := tt.PriceBytes.StorkSignedPrice.StorkCalculationAlg[:]
			expectedTemporalNumericValueR := tt.PriceBytes.StorkSignedPrice.TimestampedSignature.Signature.R[:]
			expectedTemporalNumericValueS := tt.PriceBytes.StorkSignedPrice.TimestampedSignature.Signature.S[:]
			expectedTemporalNumericValueV := tt.PriceBytes.StorkSignedPrice.TimestampedSignature.Signature.V

			assert.Equal(t, expectedID, result.ID)
			assert.Equal(t, expectedTemporalNumericValueTimestampNs, result.TemporalNumericValueTimestampNs)
			assert.Equal(t, expectedTemporalNumericValueQuantizedValueMagnitude, result.TemporalNumericValueMagnitude)
			assert.Equal(t, expectedTemporalNumericValueQuantizedValueNegative, result.TemporalNumericValueNegative)
			assert.Equal(t, expectedTemporalNumericValuePublisherMerkleRoot, result.PublisherMerkleRoot)
			assert.Equal(t, expectedTemporalNumericValueValueComputeAlgHash, result.ValueComputeAlgHash)
			assert.Equal(t, expectedTemporalNumericValueR, result.R)
			assert.Equal(t, expectedTemporalNumericValueS, result.S)
			assert.Equal(t, expectedTemporalNumericValueV, result.V)
		})
	}
}
