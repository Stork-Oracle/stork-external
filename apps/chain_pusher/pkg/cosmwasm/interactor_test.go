package cosmwasm

import (
	"math"
	"strconv"
	"testing"

	"github.com/Stork-Oracle/stork-external/apps/chain_pusher/internal/testutil"
	"github.com/Stork-Oracle/stork-external/apps/chain_pusher/pkg/cosmwasm/bindings"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAggregatedSignedPriceToUpdateData(t *testing.T) {
	t.Parallel()

	testcases := testutil.StandardPriceCase()

	for _, tt := range testcases {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			result, err := aggregatedSignedPriceToUpdateData(tt.Price)

			if tt.WantError {
				assert.Error(t, err)

				return
			}

			require.NoError(t, err)

			var expectedID [32]int
			for i, b := range tt.PriceBytes.StorkSignedPrice.EncodedAssetID {
				expectedID[i] = int(b)
			}

			if tt.PriceBytes.StorkSignedPrice.TimestampedSignature.TimestampNano > math.MaxInt {
				t.Fatalf(
					"timestamp nanoseconds is too large: %d",
					tt.PriceBytes.StorkSignedPrice.TimestampedSignature.TimestampNano,
				)
			}

			//nolint:all // this is safe to convert to an int.
			expectedTemporalNumericValueTimestampNs := bindings.Uint64(strconv.Itoa(
				int(tt.PriceBytes.StorkSignedPrice.TimestampedSignature.TimestampNano),
			))

			expectedTemporalNumericValueQuantizedValue := bindings.Int128(
				tt.PriceBytes.StorkSignedPrice.QuantizedPrice.String(),
			)

			var expectedPublisherMerkleRoot [32]int
			for i, b := range tt.PriceBytes.StorkSignedPrice.PublisherMerkleRoot {
				expectedPublisherMerkleRoot[i] = int(b)
			}

			var expectedValueComputeAlgHash [32]int
			for i, b := range tt.PriceBytes.StorkSignedPrice.StorkCalculationAlg {
				expectedValueComputeAlgHash[i] = int(b)
			}

			var expectedR [32]int
			for i, b := range tt.PriceBytes.StorkSignedPrice.TimestampedSignature.Signature.R {
				expectedR[i] = int(b)
			}

			var expectedS [32]int
			for i, b := range tt.PriceBytes.StorkSignedPrice.TimestampedSignature.Signature.S {
				expectedS[i] = int(b)
			}

			expectedV := int(tt.PriceBytes.StorkSignedPrice.TimestampedSignature.Signature.V)

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
