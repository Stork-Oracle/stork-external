package initia_minimove

import (
	"math/big"
	"testing"

	"github.com/Stork-Oracle/stork-external/apps/chain_pusher/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAggregatedSignedPriceToUpdateData(t *testing.T) {
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
