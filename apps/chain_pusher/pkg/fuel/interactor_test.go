package fuel

import (
	"math/big"
	"strconv"
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
		expected       string
		wantError      bool
	}{
		{
			name:           "simple private key",
			keyFileContent: []byte("0x1234567890abcdef"),
			expected:       "0x1234567890abcdef",
			wantError:      false,
		},
		{
			name:           "private key with trailing newline",
			keyFileContent: []byte("0x1234567890abcdef\n"),
			expected:       "0x1234567890abcdef",
			wantError:      false,
		},
		{
			name:           "private key without 0x prefix",
			keyFileContent: []byte("1234567890abcdef"),
			expected:       "1234567890abcdef",
			wantError:      false,
		},
		{
			name:           "empty content",
			keyFileContent: []byte(""),
			expected:       "",
			wantError:      true,
		},
		{
			name:           "only newline",
			keyFileContent: []byte("\n"),
			expected:       "",
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

			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAggregatedSignedPriceToTemporalNumericValueInput(t *testing.T) {
	t.Parallel()

	tests := testutil.StandardPriceCase()

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			result, err := aggregatedSignedPriceToTemporalNumericValueInput(tt.Price)
			if tt.WantError {
				assert.Error(t, err)

				return
			}

			expectedID := string(tt.Price.StorkSignedPrice.EncodedAssetID[2:])
			expectedTemporalNumericValueTimestampNs := tt.Price.StorkSignedPrice.TimestampedSignature.TimestampNano
			expectedTemporalNumericValueQuantizedValue, ok := new(
				big.Int,
			).SetString(string(tt.Price.StorkSignedPrice.QuantizedPrice), 10)
			require.True(t, ok)

			expectedPublisherMerkleRoot := tt.Price.StorkSignedPrice.PublisherMerkleRoot[2:]
			expectedValueComputeAlgHash := tt.Price.StorkSignedPrice.StorkCalculationAlg.Checksum[2:]
			expectedR := tt.Price.StorkSignedPrice.TimestampedSignature.Signature.R[2:]
			expectedS := tt.Price.StorkSignedPrice.TimestampedSignature.Signature.S[2:]
			expectedVint64, err := strconv.ParseInt(
				tt.Price.StorkSignedPrice.TimestampedSignature.Signature.V[2:],
				16,
				8,
			)
			require.NoError(t, err)
			assert.Positive(t, expectedVint64)
			assert.Less(t, expectedVint64, int64(256))
			//nolint:gosec // we check the bounds above.
			expectedV := uint8(expectedVint64)

			assert.Equal(t, expectedID, result.ID)
			assert.Equal(t, expectedTemporalNumericValueTimestampNs, result.TemporalNumericValue.TimestampNs)
			assert.Equal(t, expectedTemporalNumericValueQuantizedValue, result.TemporalNumericValue.QuantizedValue)
			assert.Equal(t, expectedPublisherMerkleRoot, result.PublisherMerkleRoot)
			assert.Equal(t, expectedValueComputeAlgHash, result.ValueComputeAlgHash)
			assert.Equal(t, expectedR, result.R)
			assert.Equal(t, expectedS, result.S)
			assert.Equal(t, expectedV, result.V)
		})
	}
}
