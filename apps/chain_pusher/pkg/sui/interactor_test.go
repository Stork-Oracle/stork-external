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
		{
			name: "large positive value",
			input: bindings.TemporalNumericValue{
				TimestampNs: 1722632569208762117,
				QuantizedValue: bindings.I128{
					Magnitude: big.NewInt(999999999999999999),
					Negative:  false,
				},
			},
			expected: types.InternalTemporalNumericValue{
				TimestampNs:    1722632569208762117,
				QuantizedValue: big.NewInt(999999999999999999),
			},
		},
		{
			name: "large negative value",
			input: bindings.TemporalNumericValue{
				TimestampNs: 1722632569208762117,
				QuantizedValue: bindings.I128{
					Magnitude: big.NewInt(999999999999999999),
					Negative:  true,
				},
			},
			expected: types.InternalTemporalNumericValue{
				TimestampNs:    1722632569208762117,
				QuantizedValue: big.NewInt(-999999999999999999),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := temporalNumericValueToInternal(tt.input)

			assert.Equal(t, tt.expected.TimestampNs, result.TimestampNs)
			assert.Equal(t, tt.expected.QuantizedValue, result.QuantizedValue)
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
			name:            "keypair prefix with extra whitespace",
			keyFileContent:  []byte("address: 0x5de3f40b2403956cbd852628dbe1dbe78cefce89ef6db8b38855203bcafe13e6\nkeypair:      AJ2dgFpuOlnVZgraFYZAX18Detm77CMKYGdw3o5QdPDy   \nflag: 0"),
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
			name:            "single line with whitespace",
			keyFileContent:  []byte("  AJ2dgFpuOlnVZgraFYZAX18Detm77CMKYGdw3o5QdPDy  "),
			expectedAddress: "0x5de3f40b2403956cbd852628dbe1dbe78cefce89ef6db8b38855203bcafe13e6",
			wantError:       false,
		},
		{
			name:            "single line with newline",
			keyFileContent:  []byte("AJ2dgFpuOlnVZgraFYZAX18Detm77CMKYGdw3o5QdPDy\n"),
			expectedAddress: "",
			wantError:       true,
		},
		{
			name:            "Invalid keypair",
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
		{
			name:            "only whitespace",
			keyFileContent:  []byte("   \n  \t  "),
			expectedAddress: "",
			wantError:       true,
		},
		{
			name:            "keypair prefix with empty value",
			keyFileContent:  []byte("keypair: "),
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
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedAddress, result.Address)
			}
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
		wantError bool
	}{
		{
			name: "valid positive price",
			price: types.AggregatedSignedPrice{
				StorkSignedPrice: &types.StorkSignedPrice{
					EncodedAssetId:      "0x1234567890123456789012345678901234567890123456789012345678901234",
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
			wantError: false,
		},
		{
			name: "valid negative price",
			price: types.AggregatedSignedPrice{
				StorkSignedPrice: &types.StorkSignedPrice{
					EncodedAssetId:      "0x1234567890123456789012345678901234567890123456789012345678901234",
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
			wantError: false,
		},
		{
			name: "zero price",
			price: types.AggregatedSignedPrice{
				StorkSignedPrice: &types.StorkSignedPrice{
					EncodedAssetId:      "0x1234567890123456789012345678901234567890123456789012345678901234",
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
			wantError: false,
		},
		{
			name: "invalid hex in encoded asset id",
			price: types.AggregatedSignedPrice{
				StorkSignedPrice: &types.StorkSignedPrice{
					EncodedAssetId:      "invalid hex",
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
			wantError: true,
		},
		{
			name: "invalid quantized price",
			price: types.AggregatedSignedPrice{
				StorkSignedPrice: &types.StorkSignedPrice{
					EncodedAssetId:      "0x1234567890123456789012345678901234567890123456789012345678901234",
					QuantizedPrice:      "not a number",
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
			wantError: true,
		},
		{
			name: "invalid hex in publisher merkle root",
			price: types.AggregatedSignedPrice{
				StorkSignedPrice: &types.StorkSignedPrice{
					EncodedAssetId:      "0x1234567890123456789012345678901234567890123456789012345678901234",
					QuantizedPrice:      "1000000000000000000",
					PublisherMerkleRoot: "invalid hex",
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
			wantError: true,
		},
		{
			name: "invalid hex in signature R",
			price: types.AggregatedSignedPrice{
				StorkSignedPrice: &types.StorkSignedPrice{
					EncodedAssetId:      "0x1234567890123456789012345678901234567890123456789012345678901234",
					QuantizedPrice:      "1000000000000000000",
					PublisherMerkleRoot: "0xe5ff773b0316059c04aa157898766731017610dcbeede7d7f169bfeaab7cc318",
					StorkCalculationAlg: types.StorkCalculationAlg{
						Checksum: "0x9be7e9f9ed459417d96112a7467bd0b27575a2c7847195c68f805b70ce1795ba",
					},
					TimestampedSignature: types.TimestampedSignature{
						TimestampNano: 1722632569208762117,
						Signature: types.EvmSignature{
							R: "invalid hex",
							S: "0x16fab526529ac795108d201832cff8c2d2b1c710da6711fe9f7ab288a7149758",
							V: "0x1c",
						},
					},
				},
			},
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

			// Validate basic fields
			assert.Equal(t, uint64(tt.price.StorkSignedPrice.TimestampedSignature.TimestampNano), result.TemporalNumericValueTimestampNs)

			// Validate magnitude handling
			expectedMagnitude := new(big.Int)
			expectedMagnitude.SetString(string(tt.price.StorkSignedPrice.QuantizedPrice), 10)
			expectedNegative := expectedMagnitude.Sign() == -1
			expectedMagnitude.Abs(expectedMagnitude)

			assert.Equal(t, expectedMagnitude, result.TemporalNumericValueMagnitude)
			assert.Equal(t, expectedNegative, result.TemporalNumericValueNegative)

			// Validate V byte conversion
			assert.Equal(t, byte(0x1c), result.V)
		})
	}
}
