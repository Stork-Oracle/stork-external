package solana

import (
	"testing"

	"github.com/Stork-Oracle/stork-external/apps/chain_pusher/internal/testutil"
	"github.com/Stork-Oracle/stork-external/apps/chain_pusher/pkg/types"
	bin "github.com/gagliardetto/binary"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBatchPriceUpdates(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		batchSize       int
		numUpdates      int
		expectedBatches int
	}{
		{"empty updates", 2, 0, 0},
		{"single update", 2, 1, 1},
		{"exact batch size", 2, 2, 1},
		{"multiple batches", 2, 3, 2},
		{"large batch", 4, 10, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create test price updates map
			priceUpdates := make(map[types.InternalEncodedAssetID]types.AggregatedSignedPrice)
			for i := range tt.numUpdates {
				var assetID types.InternalEncodedAssetID

				assetID[0] = byte(i + 1) // Create unique asset IDs
				priceUpdates[assetID] = types.AggregatedSignedPrice{}
			}

			sci := &ContractInteractor{
				batchSize: tt.batchSize,
			}

			batches := sci.batchPriceUpdates(priceUpdates)

			assert.Len(t, batches, tt.expectedBatches)

			// Verify each batch size is correct
			for i, batch := range batches {
				if i == len(batches)-1 {
					// Last batch might be smaller
					assert.LessOrEqual(t, len(batch), tt.batchSize)
				} else {
					assert.Len(t, batch, tt.batchSize)
				}
			}
		})
	}
}

func TestPriceUpdateToTemporalNumericValueEvmInput(t *testing.T) {
	t.Parallel()

	tests := testutil.StandardPriceCase()

	sci := &ContractInteractor{}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			result, err := sci.priceUpdateToTemporalNumericValueEvmInput(tt.Price, 0)

			if tt.WantError {
				assert.Error(t, err)

				return
			}

			require.NoError(t, err)
			// convert bytes representation to expected
			expectedID := tt.PriceBytes.StorkSignedPrice.EncodedAssetID
			expectedTemporalNumericValueTimestampNs := tt.PriceBytes.StorkSignedPrice.TimestampedSignature.TimestampNano
			expectedTemporalNumericValueQuantizedValue := quantizedPriceToInt128(
				types.QuantizedPrice(tt.PriceBytes.StorkSignedPrice.QuantizedPrice.String()),
			)
			expectedPublisherMerkleRoot := tt.PriceBytes.StorkSignedPrice.PublisherMerkleRoot
			expectedValueComputeAlgHash := tt.PriceBytes.StorkSignedPrice.StorkCalculationAlg
			expectedR := tt.PriceBytes.StorkSignedPrice.TimestampedSignature.Signature.R
			expectedS := tt.PriceBytes.StorkSignedPrice.TimestampedSignature.Signature.S
			expectedV := tt.PriceBytes.StorkSignedPrice.TimestampedSignature.Signature.V
			expectedTreasuryID := uint8(0)

			assert.Equal(t, expectedID, result.Id)
			assert.Equal(t, expectedTemporalNumericValueTimestampNs, result.TemporalNumericValue.TimestampNs)
			assert.Equal(t, expectedTemporalNumericValueQuantizedValue, result.TemporalNumericValue.QuantizedValue)
			assert.Equal(t, expectedPublisherMerkleRoot, result.PublisherMerkleRoot)
			assert.Equal(t, expectedValueComputeAlgHash, result.ValueComputeAlgHash)
			assert.Equal(t, expectedR, result.R)
			assert.Equal(t, expectedS, result.S)
			assert.Equal(t, expectedV, result.V)
			assert.Equal(t, expectedTreasuryID, result.TreasuryId)
		})
	}
}

func TestQuantizedPriceToInt128(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		quantizedPrice types.QuantizedPrice
		expected       bin.Int128
	}{
		{
			name:           "zero value",
			quantizedPrice: types.QuantizedPrice("0"),
			expected:       bin.Int128{Lo: 0, Hi: 0},
		},
		{
			name:           "small positive number",
			quantizedPrice: types.QuantizedPrice("1000"),
			expected:       bin.Int128{Lo: 1000, Hi: 0},
		},
		{
			name:           "large positive number",
			quantizedPrice: types.QuantizedPrice("10000000000000000000000000"),
			expected:       bin.Int128{Lo: 1590897978359414784, Hi: 542101},
		},
		{
			name:           "small negative number",
			quantizedPrice: types.QuantizedPrice("-1000"),
			expected:       bin.Int128{Lo: 18446744073709550616, Hi: 18446744073709551615},
		},
		{
			name:           "large negative number",
			quantizedPrice: types.QuantizedPrice("-10000000000000000000000000"),
			expected:       bin.Int128{Lo: 16855846095350136832, Hi: 18446744073709009514},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := quantizedPriceToInt128(tt.quantizedPrice)
			assert.Equal(t, tt.expected, result)
		})
	}
}
