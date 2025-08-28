package sui

import (
	"math/big"
	"testing"

	"github.com/Stork-Oracle/stork-external/apps/chain_pusher/pkg/sui/bindings"
	"github.com/Stork-Oracle/stork-external/apps/chain_pusher/pkg/types"
	"github.com/rs/zerolog"
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
			name:            "keypair prefix format",
			keyFileContent:  []byte("address: 0x5de3f40b2403956cbd852628dbe1dbe78cefce89ef6db8b38855203bcafe13e6\nkeypair: AJ2dgFpuOlnVZgraFYZAX18Detm77CMKYGdw3o5QdPDy\nflag: 0"),
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

	logger := zerolog.New(nil)
	sci := &ContractInteractor{
		logger: logger,
	}

	tests := []struct {
		name      string
		price     types.AggregatedSignedPrice
		expected  bindings.UpdateData
		wantError bool
	}{
		{
			name: "valid positive price",
			price: types.AggregatedSignedPrice{
				StorkSignedPrice: &types.StorkSignedPrice{
					EncodedAssetID:      "0x1234567890123456789012345678901234567890123456789012345678901234",
					QuantizedPrice:      "1000000000000000000",
					PublisherMerkleRoot: "0xe5ff773b0316059c04aa157898766731017610dcbeede7d7f169bfeaab7cc318",
					StorkCalculationAlg: types.StorkCalculationAlg{
						Checksum: "0x9be7e9f9ed459417d96112a7467bd0b27575a2c7847195c68f805b70ce1795ba",
					},
					TimestampedSignature: types.TimestampedSignature{
						TimestampNano: 1722632569208762117,
						Signature: types.EvmSignature{
							R: "0xb9b3c9f80a355bd0cd6f609fff4a4b15fa4e3b4632adabb74c020f5bcd240741",
							S: "0x16fab526529ac795108d201832cff8c2d2b1c710da6711fe9f7ab288a7149758",
							V: "0x1c",
						},
					},
				},
			},
			expected: bindings.UpdateData{
				ID:                              []byte{0x12, 0x34, 0x56, 0x78, 0x90, 0x12, 0x34, 0x56, 0x78, 0x90, 0x12, 0x34, 0x56, 0x78, 0x90, 0x12, 0x34, 0x56, 0x78, 0x90, 0x12, 0x34, 0x56, 0x78, 0x90, 0x12, 0x34, 0x56, 0x78, 0x90, 0x12, 0x34},
				TemporalNumericValueTimestampNs: 1722632569208762117,
				TemporalNumericValueMagnitude:   big.NewInt(1000000000000000000),
				TemporalNumericValueNegative:    false,
				PublisherMerkleRoot:             []byte{0xe5, 0xff, 0x77, 0x3b, 0x03, 0x16, 0x05, 0x9c, 0x04, 0xaa, 0x15, 0x78, 0x98, 0x76, 0x67, 0x31, 0x01, 0x76, 0x10, 0xdc, 0xbe, 0xed, 0xe7, 0xd7, 0xf1, 0x69, 0xbf, 0xea, 0xab, 0x7c, 0xc3, 0x18},
				ValueComputeAlgHash:             []byte{0x9b, 0xe7, 0xe9, 0xf9, 0xed, 0x45, 0x94, 0x17, 0xd9, 0x61, 0x12, 0xa7, 0x46, 0x7b, 0xd0, 0xb2, 0x75, 0x75, 0xa2, 0xc7, 0x84, 0x71, 0x95, 0xc6, 0x8f, 0x80, 0x5b, 0x70, 0xce, 0x17, 0x95, 0xba},
				R:                               []byte{0xb9, 0xb3, 0xc9, 0xf8, 0x0a, 0x35, 0x5b, 0xd0, 0xcd, 0x6f, 0x60, 0x9f, 0xff, 0x4a, 0x4b, 0x15, 0xfa, 0x4e, 0x3b, 0x46, 0x32, 0xad, 0xab, 0xb7, 0x4c, 0x02, 0x0f, 0x5b, 0xcd, 0x24, 0x07, 0x41},
				S:                               []byte{0x16, 0xfa, 0xb5, 0x26, 0x52, 0x9a, 0xc7, 0x95, 0x10, 0x8d, 0x20, 0x18, 0x32, 0xcf, 0xf8, 0xc2, 0xd2, 0xb1, 0xc7, 0x10, 0xda, 0x67, 0x11, 0xfe, 0x9f, 0x7a, 0xb2, 0x88, 0xa7, 0x14, 0x97, 0x58},
				V:                               byte(0x1c),
			},
			wantError: false,
		},
		{
			name: "valid negative price",
			price: types.AggregatedSignedPrice{
				StorkSignedPrice: &types.StorkSignedPrice{
					EncodedAssetID:      "0x1234567890123456789012345678901234567890123456789012345678901234",
					QuantizedPrice:      "-1000000000000000000",
					PublisherMerkleRoot: "0xe5ff773b0316059c04aa157898766731017610dcbeede7d7f169bfeaab7cc318",
					StorkCalculationAlg: types.StorkCalculationAlg{
						Checksum: "0x9be7e9f9ed459417d96112a7467bd0b27575a2c7847195c68f805b70ce1795ba",
					},
					TimestampedSignature: types.TimestampedSignature{
						TimestampNano: 1722632569208762117,
						Signature: types.EvmSignature{
							R: "0xb9b3c9f80a355bd0cd6f609fff4a4b15fa4e3b4632adabb74c020f5bcd240741",
							S: "0x16fab526529ac795108d201832cff8c2d2b1c710da6711fe9f7ab288a7149758",
							V: "0x1c",
						},
					},
				},
			},
			expected: bindings.UpdateData{
				ID:                              []byte{0x12, 0x34, 0x56, 0x78, 0x90, 0x12, 0x34, 0x56, 0x78, 0x90, 0x12, 0x34, 0x56, 0x78, 0x90, 0x12, 0x34, 0x56, 0x78, 0x90, 0x12, 0x34, 0x56, 0x78, 0x90, 0x12, 0x34, 0x56, 0x78, 0x90, 0x12, 0x34},
				TemporalNumericValueTimestampNs: 1722632569208762117,
				TemporalNumericValueMagnitude:   big.NewInt(1000000000000000000),
				TemporalNumericValueNegative:    true,
				PublisherMerkleRoot:             []byte{0xe5, 0xff, 0x77, 0x3b, 0x03, 0x16, 0x05, 0x9c, 0x04, 0xaa, 0x15, 0x78, 0x98, 0x76, 0x67, 0x31, 0x01, 0x76, 0x10, 0xdc, 0xbe, 0xed, 0xe7, 0xd7, 0xf1, 0x69, 0xbf, 0xea, 0xab, 0x7c, 0xc3, 0x18},
				ValueComputeAlgHash:             []byte{0x9b, 0xe7, 0xe9, 0xf9, 0xed, 0x45, 0x94, 0x17, 0xd9, 0x61, 0x12, 0xa7, 0x46, 0x7b, 0xd0, 0xb2, 0x75, 0x75, 0xa2, 0xc7, 0x84, 0x71, 0x95, 0xc6, 0x8f, 0x80, 0x5b, 0x70, 0xce, 0x17, 0x95, 0xba},
				R:                               []byte{0xb9, 0xb3, 0xc9, 0xf8, 0x0a, 0x35, 0x5b, 0xd0, 0xcd, 0x6f, 0x60, 0x9f, 0xff, 0x4a, 0x4b, 0x15, 0xfa, 0x4e, 0x3b, 0x46, 0x32, 0xad, 0xab, 0xb7, 0x4c, 0x02, 0x0f, 0x5b, 0xcd, 0x24, 0x07, 0x41},
				S:                               []byte{0x16, 0xfa, 0xb5, 0x26, 0x52, 0x9a, 0xc7, 0x95, 0x10, 0x8d, 0x20, 0x18, 0x32, 0xcf, 0xf8, 0xc2, 0xd2, 0xb1, 0xc7, 0x10, 0xda, 0x67, 0x11, 0xfe, 0x9f, 0x7a, 0xb2, 0x88, 0xa7, 0x14, 0x97, 0x58},
				V:                               byte(0x1c),
			},
			wantError: false,
		},
		{
			name: "zero price",
			price: types.AggregatedSignedPrice{
				StorkSignedPrice: &types.StorkSignedPrice{
					EncodedAssetID:      "0x1234567890123456789012345678901234567890123456789012345678901234",
					QuantizedPrice:      "0",
					PublisherMerkleRoot: "0xe5ff773b0316059c04aa157898766731017610dcbeede7d7f169bfeaab7cc318",
					StorkCalculationAlg: types.StorkCalculationAlg{
						Checksum: "0x9be7e9f9ed459417d96112a7467bd0b27575a2c7847195c68f805b70ce1795ba",
					},
					TimestampedSignature: types.TimestampedSignature{
						TimestampNano: 1722632569208762117,
						Signature: types.EvmSignature{
							R: "0xb9b3c9f80a355bd0cd6f609fff4a4b15fa4e3b4632adabb74c020f5bcd240741",
							S: "0x16fab526529ac795108d201832cff8c2d2b1c710da6711fe9f7ab288a7149758",
							V: "0x1c",
						},
					},
				},
			},
			expected: bindings.UpdateData{
				ID:                              []byte{0x12, 0x34, 0x56, 0x78, 0x90, 0x12, 0x34, 0x56, 0x78, 0x90, 0x12, 0x34, 0x56, 0x78, 0x90, 0x12, 0x34, 0x56, 0x78, 0x90, 0x12, 0x34, 0x56, 0x78, 0x90, 0x12, 0x34, 0x56, 0x78, 0x90, 0x12, 0x34},
				TemporalNumericValueTimestampNs: 1722632569208762117,
				TemporalNumericValueMagnitude:   big.NewInt(0),
				TemporalNumericValueNegative:    false,
				PublisherMerkleRoot:             []byte{0xe5, 0xff, 0x77, 0x3b, 0x03, 0x16, 0x05, 0x9c, 0x04, 0xaa, 0x15, 0x78, 0x98, 0x76, 0x67, 0x31, 0x01, 0x76, 0x10, 0xdc, 0xbe, 0xed, 0xe7, 0xd7, 0xf1, 0x69, 0xbf, 0xea, 0xab, 0x7c, 0xc3, 0x18},
				ValueComputeAlgHash:             []byte{0x9b, 0xe7, 0xe9, 0xf9, 0xed, 0x45, 0x94, 0x17, 0xd9, 0x61, 0x12, 0xa7, 0x46, 0x7b, 0xd0, 0xb2, 0x75, 0x75, 0xa2, 0xc7, 0x84, 0x71, 0x95, 0xc6, 0x8f, 0x80, 0x5b, 0x70, 0xce, 0x17, 0x95, 0xba},
				R:                               []byte{0xb9, 0xb3, 0xc9, 0xf8, 0x0a, 0x35, 0x5b, 0xd0, 0xcd, 0x6f, 0x60, 0x9f, 0xff, 0x4a, 0x4b, 0x15, 0xfa, 0x4e, 0x3b, 0x46, 0x32, 0xad, 0xab, 0xb7, 0x4c, 0x02, 0x0f, 0x5b, 0xcd, 0x24, 0x07, 0x41},
				S:                               []byte{0x16, 0xfa, 0xb5, 0x26, 0x52, 0x9a, 0xc7, 0x95, 0x10, 0x8d, 0x20, 0x18, 0x32, 0xcf, 0xf8, 0xc2, 0xd2, 0xb1, 0xc7, 0x10, 0xda, 0x67, 0x11, 0xfe, 0x9f, 0x7a, 0xb2, 0x88, 0xa7, 0x14, 0x97, 0x58},
				V:                               byte(0x1c),
			},
			wantError: false,
		},
		{
			name: "valid price with V=0x1b",
			price: types.AggregatedSignedPrice{
				StorkSignedPrice: &types.StorkSignedPrice{
					EncodedAssetID:      "0x1234567890123456789012345678901234567890123456789012345678901234",
					QuantizedPrice:      "1000000000000000000",
					PublisherMerkleRoot: "0xe5ff773b0316059c04aa157898766731017610dcbeede7d7f169bfeaab7cc318",
					StorkCalculationAlg: types.StorkCalculationAlg{
						Checksum: "0x9be7e9f9ed459417d96112a7467bd0b27575a2c7847195c68f805b70ce1795ba",
					},
					TimestampedSignature: types.TimestampedSignature{
						TimestampNano: 1722632569208762117,
						Signature: types.EvmSignature{
							R: "0xb9b3c9f80a355bd0cd6f609fff4a4b15fa4e3b4632adabb74c020f5bcd240741",
							S: "0x16fab526529ac795108d201832cff8c2d2b1c710da6711fe9f7ab288a7149758",
							V: "0x1b",
						},
					},
				},
			},
			expected: bindings.UpdateData{
				ID:                              []byte{0x12, 0x34, 0x56, 0x78, 0x90, 0x12, 0x34, 0x56, 0x78, 0x90, 0x12, 0x34, 0x56, 0x78, 0x90, 0x12, 0x34, 0x56, 0x78, 0x90, 0x12, 0x34, 0x56, 0x78, 0x90, 0x12, 0x34, 0x56, 0x78, 0x90, 0x12, 0x34},
				TemporalNumericValueTimestampNs: 1722632569208762117,
				TemporalNumericValueMagnitude:   big.NewInt(1000000000000000000),
				TemporalNumericValueNegative:    false,
				PublisherMerkleRoot:             []byte{0xe5, 0xff, 0x77, 0x3b, 0x03, 0x16, 0x05, 0x9c, 0x04, 0xaa, 0x15, 0x78, 0x98, 0x76, 0x67, 0x31, 0x01, 0x76, 0x10, 0xdc, 0xbe, 0xed, 0xe7, 0xd7, 0xf1, 0x69, 0xbf, 0xea, 0xab, 0x7c, 0xc3, 0x18},
				ValueComputeAlgHash:             []byte{0x9b, 0xe7, 0xe9, 0xf9, 0xed, 0x45, 0x94, 0x17, 0xd9, 0x61, 0x12, 0xa7, 0x46, 0x7b, 0xd0, 0xb2, 0x75, 0x75, 0xa2, 0xc7, 0x84, 0x71, 0x95, 0xc6, 0x8f, 0x80, 0x5b, 0x70, 0xce, 0x17, 0x95, 0xba},
				R:                               []byte{0xb9, 0xb3, 0xc9, 0xf8, 0x0a, 0x35, 0x5b, 0xd0, 0xcd, 0x6f, 0x60, 0x9f, 0xff, 0x4a, 0x4b, 0x15, 0xfa, 0x4e, 0x3b, 0x46, 0x32, 0xad, 0xab, 0xb7, 0x4c, 0x02, 0x0f, 0x5b, 0xcd, 0x24, 0x07, 0x41},
				S:                               []byte{0x16, 0xfa, 0xb5, 0x26, 0x52, 0x9a, 0xc7, 0x95, 0x10, 0x8d, 0x20, 0x18, 0x32, 0xcf, 0xf8, 0xc2, 0xd2, 0xb1, 0xc7, 0x10, 0xda, 0x67, 0x11, 0xfe, 0x9f, 0x7a, 0xb2, 0x88, 0xa7, 0x14, 0x97, 0x58},
				V:                               byte(0x1b),
			},
			wantError: false,
		},
		{
			name: "invalid V signature",
			price: types.AggregatedSignedPrice{
				StorkSignedPrice: &types.StorkSignedPrice{
					EncodedAssetID:      "0x1234567890123456789012345678901234567890123456789012345678901234",
					QuantizedPrice:      "1000000000000000000",
					PublisherMerkleRoot: "0xe5ff773b0316059c04aa157898766731017610dcbeede7d7f169bfeaab7cc318",
					StorkCalculationAlg: types.StorkCalculationAlg{
						Checksum: "0x9be7e9f9ed459417d96112a7467bd0b27575a2c7847195c68f805b70ce1795ba",
					},
					TimestampedSignature: types.TimestampedSignature{
						TimestampNano: 1722632569208762117,
						Signature: types.EvmSignature{
							R: "0xb9b3c9f80a355bd0cd6f609fff4a4b15fa4e3b4632adabb74c020f5bcd240741",
							S: "0x16fab526529ac795108d201832cff8c2d2b1c710da6711fe9f7ab288a7149758",
							V: "invalid",
						},
					},
				},
			},
			expected:  bindings.UpdateData{},
			wantError: true,
		},
		{
			name: "invalid encoded asset id",
			price: types.AggregatedSignedPrice{
				StorkSignedPrice: &types.StorkSignedPrice{
					EncodedAssetID:      "invalid hex",
					QuantizedPrice:      "1000000000000000000",
					PublisherMerkleRoot: "0xe5ff773b0316059c04aa157898766731017610dcbeede7d7f169bfeaab7cc318",
					StorkCalculationAlg: types.StorkCalculationAlg{
						Checksum: "0x9be7e9f9ed459417d96112a7467bd0b27575a2c7847195c68f805b70ce1795ba",
					},
					TimestampedSignature: types.TimestampedSignature{
						TimestampNano: 1722632569208762117,
						Signature: types.EvmSignature{
							R: "0xb9b3c9f80a355bd0cd6f609fff4a4b15fa4e3b4632adabb74c020f5bcd240741",
							S: "0x16fab526529ac795108d201832cff8c2d2b1c710da6711fe9f7ab288a7149758",
							V: "0x1c",
						},
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

			result, err := sci.aggregatedSignedPriceToUpdateData(tt.price)

			if tt.wantError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}
