package sui

import (
	"math/big"
	"testing"

	"github.com/Stork-Oracle/stork-external/apps/chain_pusher/pkg/internal/testutil"
	"github.com/Stork-Oracle/stork-external/apps/chain_pusher/pkg/sui/bindings"
	"github.com/Stork-Oracle/stork-external/apps/chain_pusher/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTemporalNumericValueToInternal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    bindings.TemporalNumericValue
		expected types.InternalTemporalNumericValue
	}{
		{
			name: "positive value",
			input: bindings.TemporalNumericValue{
				TimestampNs: 1722632569208762117,
				QuantizedValue: bindings.I128{
					Magnitude: big.NewInt(1000000000000000000),
					Negative:  false,
				},
			},
			expected: types.InternalTemporalNumericValue{
				TimestampNs:    1722632569208762117,
				QuantizedValue: big.NewInt(1000000000000000000),
			},
		},
		{
			name: "negative value",
			input: bindings.TemporalNumericValue{
				TimestampNs: 1722632569208762117,
				QuantizedValue: bindings.I128{
					Magnitude: big.NewInt(1000000000000000000),
					Negative:  true,
				},
			},
			expected: types.InternalTemporalNumericValue{
				TimestampNs:    1722632569208762117,
				QuantizedValue: big.NewInt(-1000000000000000000),
			},
		},
		{
			name: "zero value",
			input: bindings.TemporalNumericValue{
				TimestampNs: 1000000000000,
				QuantizedValue: bindings.I128{
					Magnitude: big.NewInt(0),
					Negative:  false,
				},
			},
			expected: types.InternalTemporalNumericValue{
				TimestampNs:    1000000000000,
				QuantizedValue: big.NewInt(0),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := temporalNumericValueToInternal(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestLoadPrivateKey(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		keyFileContent  []byte
		expectedAddress string
		wantError       bool
	}{
		{
			name: "keypair prefix format",
			keyFileContent: []byte(
				"address: 0x5de3f40b2403956cbd852628dbe1dbe78cefce89ef6db8b38855203bcafe13e6\nkeypair: AJ2dgFpuOlnVZgraFYZAX18Detm77CMKYGdw3o5QdPDy\nflag: 0",
			),
			expectedAddress: "0x5de3f40b2403956cbd852628dbe1dbe78cefce89ef6db8b38855203bcafe13e6",
			wantError:       false,
		},
		{
			name:            "single line without prefix",
			keyFileContent:  []byte("AJ2dgFpuOlnVZgraFYZAX18Detm77CMKYGdw3o5QdPDy"),
			expectedAddress: "0x5de3f40b2403956cbd852628dbe1dbe78cefce89ef6db8b38855203bcafe13e6",
			wantError:       false,
		},
		{
			name:            "invalid keypair",
			keyFileContent:  []byte("B2dgFpuOlnVZgraFYZAX18Detm77CMKYGdw3o5QdPDy"),
			expectedAddress: "",
			wantError:       true,
		},
		{
			name:            "empty content",
			keyFileContent:  []byte(""),
			expectedAddress: "",
			wantError:       true,
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
			assert.Equal(t, tt.expectedAddress, result.Address)
		})
	}
}

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

			// convert bytes representation to expected
			expectedID := tt.PriceBytes.StorkSignedPrice.EncodedAssetID[:]
			expectedTemporalNumericValueMagnitude := new(big.Int).Abs(tt.PriceBytes.StorkSignedPrice.QuantizedPrice)
			expectedTemporalNumericValueNegative := tt.PriceBytes.StorkSignedPrice.QuantizedPrice.Sign() == -1
			expectedTemporalNumericValueTimestampNs := tt.PriceBytes.StorkSignedPrice.TimestampedSignature.TimestampNano
			expectedTemporalNumericValuePublisherMerkleRoot := tt.PriceBytes.StorkSignedPrice.PublisherMerkleRoot[:]
			expectedTemporalNumericValueValueComputeAlgHash := tt.PriceBytes.StorkSignedPrice.StorkCalculationAlg[:]
			expectedTemporalNumericValueR := tt.PriceBytes.StorkSignedPrice.TimestampedSignature.Signature.R[:]
			expectedTemporalNumericValueS := tt.PriceBytes.StorkSignedPrice.TimestampedSignature.Signature.S[:]
			expectedTemporalNumericValueV := tt.PriceBytes.StorkSignedPrice.TimestampedSignature.Signature.V

			assert.Equal(t, expectedID, result.ID)
			assert.Equal(t, expectedTemporalNumericValueMagnitude, result.TemporalNumericValueMagnitude)
			assert.Equal(t, expectedTemporalNumericValueNegative, result.TemporalNumericValueNegative)
			assert.Equal(t, expectedTemporalNumericValueTimestampNs, result.TemporalNumericValueTimestampNs)
			assert.Equal(t, expectedTemporalNumericValuePublisherMerkleRoot, result.PublisherMerkleRoot)
			assert.Equal(t, expectedTemporalNumericValueValueComputeAlgHash, result.ValueComputeAlgHash)
			assert.Equal(t, expectedTemporalNumericValueR, result.R)
			assert.Equal(t, expectedTemporalNumericValueS, result.S)
			assert.Equal(t, expectedTemporalNumericValueV, result.V)
		})
	}
}
