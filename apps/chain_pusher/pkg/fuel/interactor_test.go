package fuel

import (
	"testing"

	"github.com/Stork-Oracle/stork-external/apps/chain_pusher/pkg/types"
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
			name:           "private key with multiple trailing newlines",
			keyFileContent: []byte("0x1234567890abcdef\n\n"),
			expected:       "0x1234567890abcdef\n",
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
		{
			name:           "private key without 0x prefix",
			keyFileContent: []byte("1234567890abcdef"),
			expected:       "1234567890abcdef",
			wantError:      false,
		},
		{
			name:           "private key with spaces (no trimming of spaces)",
			keyFileContent: []byte("  0x1234567890abcdef  "),
			expected:       "  0x1234567890abcdef  ",
			wantError:      false,
		},
		{
			name:           "very short private key",
			keyFileContent: []byte("abc"),
			expected:       "abc",
			wantError:      false,
		},
		{
			name:           "only spaces",
			keyFileContent: []byte("   "),
			expected:       "   ",
			wantError:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := loadPrivateKey(tt.keyFileContent)

			if tt.wantError {
				assert.Error(t, err)
				assert.Empty(t, result)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAggregatedSignedPriceToTemporalNumericValueInput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		update    types.AggregatedSignedPrice
		wantError bool
	}{
		{
			name: "valid positive price update",
			update: types.AggregatedSignedPrice{
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
			name: "valid negative price update",
			update: types.AggregatedSignedPrice{
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
			update: types.AggregatedSignedPrice{
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
			name: "invalid encoded asset ID",
			update: types.AggregatedSignedPrice{
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
			name: "invalid signature R",
			update: types.AggregatedSignedPrice{
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
		{
			name: "invalid signature S",
			update: types.AggregatedSignedPrice{
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
							S: "invalid hex",
							V: "0x1c",
						},
					},
				},
			},
			wantError: true,
		},
		{
			name: "invalid publisher merkle root",
			update: types.AggregatedSignedPrice{
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
			name: "invalid checksum",
			update: types.AggregatedSignedPrice{
				StorkSignedPrice: &types.StorkSignedPrice{
					EncodedAssetId:      "0x1234567890123456789012345678901234567890123456789012345678901234",
					QuantizedPrice:      "1000000000000000000",
					PublisherMerkleRoot: "0xe5ff773b0316059c04aa157898766731017610dcbeede7d7f169bfeaab7cc318",
					StorkCalculationAlg: types.StorkCalculationAlg{
						Checksum: "invalid hex",
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
			name: "invalid V signature",
			update: types.AggregatedSignedPrice{
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
							V: "invalid",
						},
					},
				},
			},
			wantError: true,
		},
		{
			name: "V signature too short",
			update: types.AggregatedSignedPrice{
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
							V: "0x",
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

			result, err := aggregatedSignedPriceToTemporalNumericValueInput(tt.update)

			if tt.wantError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)

			// Validate basic fields
			assert.Equal(t, uint64(tt.update.StorkSignedPrice.TimestampedSignature.TimestampNano), result.TemporalNumericValue.TimestampNs)
			assert.NotNil(t, result.TemporalNumericValue.QuantizedValue)

			// Validate V conversion (0x1c should convert to 28)
			assert.Equal(t, uint8(28), result.V)

			// Validate that hex strings are properly converted to hex encoded strings
			assert.NotEmpty(t, result.Id)
			assert.NotEmpty(t, result.PublisherMerkleRoot)
			assert.NotEmpty(t, result.ValueComputeAlgHash)
			assert.NotEmpty(t, result.R)
			assert.NotEmpty(t, result.S)
		})
	}
}
