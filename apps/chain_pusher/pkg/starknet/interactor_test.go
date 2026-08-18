package starknet

import (
	"math/big"
	"testing"

	"github.com/Stork-Oracle/stork-external/apps/chain_pusher/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAggregatedSignedPriceToUpdateData(t *testing.T) {
	t.Parallel()

	for _, tt := range testutil.StandardPriceCase() {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			result, err := aggregatedSignedPriceToUpdateData(tt.Price)

			if tt.WantError {
				assert.Error(t, err)

				return
			}

			require.NoError(t, err)

			expected := tt.PriceBytes.StorkSignedPrice

			assert.Equal(t, expected.EncodedAssetID, [32]byte(result.ID))
			assert.Equal(t, expected.TimestampedSignature.TimestampNano, result.TemporalNumericValue.TimestampNs)
			assert.Equal(t, expected.QuantizedPrice, result.TemporalNumericValue.QuantizedValue)
			assert.Equal(t, expected.PublisherMerkleRoot, result.PublisherMerkleRoot)
			assert.Equal(t, expected.StorkCalculationAlg, result.ValueComputeAlgHash)
			assert.Equal(t, expected.TimestampedSignature.Signature.R, result.R)
			assert.Equal(t, expected.TimestampedSignature.Signature.S, result.S)
			assert.Equal(t, expected.TimestampedSignature.Signature.V, result.V)
		})
	}
}

func TestAggregatedSignedPriceToUpdateDataPreservesNegativePrices(t *testing.T) {
	t.Parallel()

	prices, err := testutil.LoadAggregatedSignedPrices()
	require.NoError(t, err)

	price, err := prices.NextNegativeAsset1()
	require.NoError(t, err)

	update, err := aggregatedSignedPriceToUpdateData(*price)
	require.NoError(t, err)

	assert.Negative(t, update.TemporalNumericValue.QuantizedValue.Sign())
	assert.Equal(
		t,
		string(price.StorkSignedPrice.QuantizedPrice),
		update.TemporalNumericValue.QuantizedValue.String(),
	)
}

func TestLoadPrivateKey(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		content   string
		want      *big.Int
		wantError bool
	}{
		{
			name:    "hex with 0x prefix",
			content: "0x0000000000000000000000000000000000000000000000000000000000000001",
			want:    big.NewInt(1),
		},
		{
			name:    "hex without prefix",
			content: "ff",
			want:    big.NewInt(255),
		},
		{
			name:    "trailing newline and whitespace are ignored",
			content: "  0x2a  \n",
			want:    big.NewInt(42),
		},
		{
			name:    "only the first line is read",
			content: "0x2a\nignored\n",
			want:    big.NewInt(42),
		},
		{
			name:      "empty file",
			content:   "\n",
			wantError: true,
		},
		{
			name:      "not hex",
			content:   "not-a-key",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := loadPrivateKey([]byte(tt.content))

			if tt.wantError {
				assert.Error(t, err)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, 0, tt.want.Cmp(got), "want %s, got %s", tt.want, got)
		})
	}
}
