package cosmwasm

import (
	"testing"

	"github.com/Stork-Oracle/stork-external/apps/chain_pusher/pkg/cosmwasm/bindings"
	"github.com/Stork-Oracle/stork-external/apps/chain_pusher/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAggregatedSignedPriceToUpdateData(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		input     types.AggregatedSignedPrice
		expected  bindings.UpdateData
		wantError bool
	}{
		{
			name: "valid conversion",
			input: types.AggregatedSignedPrice{
				StorkSignedPrice: &types.StorkSignedPrice{
					EncodedAssetId: "0x1234567890123456789012345678901234567890123456789012345678901234",
					QuantizedPrice: "123456789",
					TimestampedSignature: types.TimestampedSignature{
						TimestampNano: 1640995200000000000,
						Signature: types.EvmSignature{
							R: "0xabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcd",
							S: "0x1234567812345678123456781234567812345678123456781234567812345678",
							V: "0x1b",
						},
					},
					PublisherMerkleRoot: "0xfedcba0987654321fedcba0987654321fedcba0987654321fedcba0987654321",
					StorkCalculationAlg: types.StorkCalculationAlg{
						Checksum: "0x9876543210abcdef9876543210abcdef9876543210abcdef9876543210abcdef",
					},
				},
			},
			expected: bindings.UpdateData{
				Id: [32]int{0x12, 0x34, 0x56, 0x78, 0x90, 0x12, 0x34, 0x56, 0x78, 0x90, 0x12, 0x34, 0x56, 0x78, 0x90, 0x12, 0x34, 0x56, 0x78, 0x90, 0x12, 0x34, 0x56, 0x78, 0x90, 0x12, 0x34, 0x56, 0x78, 0x90, 0x12, 0x34},
				TemporalNumericValue: bindings.TemporalNumericValue{
					QuantizedValue: bindings.Int128("123456789"),
					TimestampNs:    bindings.Uint64("1640995200000000000"),
				},
				ValueComputeAlgHash: [32]int{0x98, 0x76, 0x54, 0x32, 0x10, 0xab, 0xcd, 0xef, 0x98, 0x76, 0x54, 0x32, 0x10, 0xab, 0xcd, 0xef, 0x98, 0x76, 0x54, 0x32, 0x10, 0xab, 0xcd, 0xef, 0x98, 0x76, 0x54, 0x32, 0x10, 0xab, 0xcd, 0xef},
				PublisherMerkleRoot: [32]int{0xfe, 0xdc, 0xba, 0x09, 0x87, 0x65, 0x43, 0x21, 0xfe, 0xdc, 0xba, 0x09, 0x87, 0x65, 0x43, 0x21, 0xfe, 0xdc, 0xba, 0x09, 0x87, 0x65, 0x43, 0x21, 0xfe, 0xdc, 0xba, 0x09, 0x87, 0x65, 0x43, 0x21},
				R:                   [32]int{0xab, 0xcd, 0xef, 0xab, 0xcd, 0xef, 0xab, 0xcd, 0xef, 0xab, 0xcd, 0xef, 0xab, 0xcd, 0xef, 0xab, 0xcd, 0xef, 0xab, 0xcd, 0xef, 0xab, 0xcd, 0xef, 0xab, 0xcd, 0xef, 0xab, 0xcd, 0xef, 0xab, 0xcd},
				S:                   [32]int{0x12, 0x34, 0x56, 0x78, 0x12, 0x34, 0x56, 0x78, 0x12, 0x34, 0x56, 0x78, 0x12, 0x34, 0x56, 0x78, 0x12, 0x34, 0x56, 0x78, 0x12, 0x34, 0x56, 0x78, 0x12, 0x34, 0x56, 0x78, 0x12, 0x34, 0x56, 0x78},
				V:                   27,
			},
			wantError: false,
		},
		{
			name: "v parameter 0x1c",
			input: types.AggregatedSignedPrice{
				StorkSignedPrice: &types.StorkSignedPrice{
					EncodedAssetId: "0x1111111111111111111111111111111111111111111111111111111111111111",
					QuantizedPrice: "987654321",
					TimestampedSignature: types.TimestampedSignature{
						TimestampNano: 1641081600000000000,
						Signature: types.EvmSignature{
							R: "0x1111111111111111111111111111111111111111111111111111111111111111",
							S: "0x2222222222222222222222222222222222222222222222222222222222222222",
							V: "0x1c",
						},
					},
					PublisherMerkleRoot: "0x3333333333333333333333333333333333333333333333333333333333333333",
					StorkCalculationAlg: types.StorkCalculationAlg{
						Checksum: "0x4444444444444444444444444444444444444444444444444444444444444444",
					},
				},
			},
			expected: bindings.UpdateData{
				Id: [32]int{17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17},
				TemporalNumericValue: bindings.TemporalNumericValue{
					QuantizedValue: bindings.Int128("987654321"),
					TimestampNs:    bindings.Uint64("1641081600000000000"),
				},
				ValueComputeAlgHash: [32]int{68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68},
				PublisherMerkleRoot: [32]int{51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51, 51},
				R:                   [32]int{17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17},
				S:                   [32]int{34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34},
				V:                   28,
			},
			wantError: false,
		},
		{
			name: "invalid encoded asset id",
			input: types.AggregatedSignedPrice{
				StorkSignedPrice: &types.StorkSignedPrice{
					EncodedAssetId: "invalid hex",
					QuantizedPrice: "123456789",
					TimestampedSignature: types.TimestampedSignature{
						TimestampNano: 1640995200000000000,
						Signature: types.EvmSignature{
							R: "0x1234567890123456789012345678901234567890123456789012345678901234",
							S: "0xabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcd",
							V: "0x1b",
						},
					},
					PublisherMerkleRoot: "0xfedcba0987654321fedcba0987654321fedcba0987654321fedcba0987654321",
					StorkCalculationAlg: types.StorkCalculationAlg{
						Checksum: "0x9876543210abcdef9876543210abcdef9876543210abcdef9876543210abcdef",
					},
				},
			},
			expected:  bindings.UpdateData{},
			wantError: true,
		},
		{
			name: "invalid R value",
			input: types.AggregatedSignedPrice{
				StorkSignedPrice: &types.StorkSignedPrice{
					EncodedAssetId: "0x1234567890123456789012345678901234567890123456789012345678901234",
					QuantizedPrice: "123456789",
					TimestampedSignature: types.TimestampedSignature{
						TimestampNano: 1640995200000000000,
						Signature: types.EvmSignature{
							R: "invalid hex",
							S: "0xabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcd",
							V: "0x1b",
						},
					},
					PublisherMerkleRoot: "0xfedcba0987654321fedcba0987654321fedcba0987654321fedcba0987654321",
					StorkCalculationAlg: types.StorkCalculationAlg{
						Checksum: "0x9876543210abcdef9876543210abcdef9876543210abcdef9876543210abcdef",
					},
				},
			},
			expected:  bindings.UpdateData{},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			sci := &ContractInteractor{}
			result, err := sci.aggregatedSignedPriceToUpdateData(tt.input)

			if tt.wantError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}
