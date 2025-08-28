package evm

import (
	"crypto/ecdsa"
	"testing"

	"github.com/Stork-Oracle/stork-external/apps/chain_pusher/pkg/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadPrivateKey(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		keyFileContent []byte
		expectedPubKey string // We'll verify by checking the derived public key address
		wantError      bool
	}{
		{
			name:           "valid private key",
			keyFileContent: []byte("ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"),
			expectedPubKey: "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266", // Known address for this private key
			wantError:      false,
		},
		{
			name:           "valid private key with newline",
			keyFileContent: []byte("ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80\n"),
			expectedPubKey: "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266",
			wantError:      false,
		},
		{
			name:           "valid private key with spaces and newlines",
			keyFileContent: []byte("  ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80  \n"),
			expectedPubKey: "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266",
			wantError:      false,
		},
		{
			name:           "valid private key with 0x prefix",
			keyFileContent: []byte("0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"),
			wantError:      true, // crypto.HexToECDSA doesn't accept 0x prefix
		},
		{
			name:           "invalid hex string",
			keyFileContent: []byte("invalid hex"),
			wantError:      true,
		},
		{
			name:           "too short private key",
			keyFileContent: []byte("1234"),
			wantError:      true,
		},
		{
			name:           "empty input",
			keyFileContent: []byte(""),
			wantError:      true,
		},
		{
			name:           "only whitespace",
			keyFileContent: []byte("   \n  \t  "),
			wantError:      true,
		},
		{
			name:           "private key with invalid characters",
			keyFileContent: []byte("ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff8g"), // 'g' is invalid hex
			wantError:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := loadPrivateKey(tt.keyFileContent)

			if tt.wantError {
				assert.Error(t, err)
				assert.Nil(t, result)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)
			assert.IsType(t, &ecdsa.PrivateKey{}, result)

			// Verify the private key by checking the derived address
			publicKey := result.Public()
			publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
			require.True(t, ok)
			address := crypto.PubkeyToAddress(*publicKeyECDSA)
			assert.Equal(t, tt.expectedPubKey, address.Hex())
		})
	}
}

func TestGetUpdatePayload(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		priceUpdates map[types.InternalEncodedAssetId]types.AggregatedSignedPrice
		wantError    bool
	}{
		{
			name: "valid single price update",
			priceUpdates: map[types.InternalEncodedAssetId]types.AggregatedSignedPrice{
				[32]byte{1, 2, 3, 4}: {
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
			},
			wantError: false,
		},
		{
			name: "valid negative price",
			priceUpdates: map[types.InternalEncodedAssetId]types.AggregatedSignedPrice{
				[32]byte{1, 2, 3, 4}: {
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
			},
			wantError: false,
		},
		{
			name: "valid price with V=0x1b",
			priceUpdates: map[types.InternalEncodedAssetId]types.AggregatedSignedPrice{
				[32]byte{5, 6, 7, 8}: {
					StorkSignedPrice: &types.StorkSignedPrice{
						EncodedAssetId:      "0xabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcd",
						QuantizedPrice:      "500000000000000000",
						PublisherMerkleRoot: "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
						StorkCalculationAlg: types.StorkCalculationAlg{
							Checksum: "0xfedcba0987654321fedcba0987654321fedcba0987654321fedcba0987654321",
						},
						TimestampedSignature: types.TimestampedSignature{
							TimestampNano: 1722632569208762200,
							Signature: types.EvmSignature{
								R: "0x1111111111111111111111111111111111111111111111111111111111111111",
								S: "0x2222222222222222222222222222222222222222222222222222222222222222",
								V: "0x1b",
							},
						},
					},
				},
			},
			wantError: false,
		},
		{
			name: "multiple price updates",
			priceUpdates: map[types.InternalEncodedAssetId]types.AggregatedSignedPrice{
				[32]byte{1, 2, 3, 4}: {
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
				[32]byte{5, 6, 7, 8}: {
					StorkSignedPrice: &types.StorkSignedPrice{
						EncodedAssetId:      "0x5678567890123456789012345678901234567890123456789012345678901234",
						QuantizedPrice:      "2000000000000000000",
						PublisherMerkleRoot: "0xe5ff773b0316059c04aa157898766731017610dcbeede7d7f169bfeaab7cc318",
						StorkCalculationAlg: types.StorkCalculationAlg{
							Checksum: "0x9be7e9f9ed459417d96112a7467bd0b27575a2c7847195c68f805b70ce1795ba",
						},
						TimestampedSignature: types.TimestampedSignature{
							TimestampNano: 1722632569208762118,
							Signature: types.EvmSignature{
								R: "0xb9b3c9f80a355bd0cd6f609fff4a4b15fa4e3b4632adabb74c020f5bcd240741",
								S: "0x16fab526529ac795108d201832cff8c2d2b1c710da6711fe9f7ab288a7149758",
								V: "0x1c",
							},
						},
					},
				},
			},
			wantError: false,
		},
		{
			name: "invalid encoded asset id",
			priceUpdates: map[types.InternalEncodedAssetId]types.AggregatedSignedPrice{
				[32]byte{1, 2, 3, 4}: {
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
			},
			wantError: true,
		},
		{
			name: "invalid signature R",
			priceUpdates: map[types.InternalEncodedAssetId]types.AggregatedSignedPrice{
				[32]byte{1, 2, 3, 4}: {
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
			},
			wantError: true,
		},
		{
			name: "invalid V value",
			priceUpdates: map[types.InternalEncodedAssetId]types.AggregatedSignedPrice{
				[32]byte{1, 2, 3, 4}: {
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
			},
			wantError: true,
		},
		{
			name:         "empty price updates",
			priceUpdates: map[types.InternalEncodedAssetId]types.AggregatedSignedPrice{},
			wantError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := getUpdatePayload(tt.priceUpdates)

			if tt.wantError {
				assert.Error(t, err)
				assert.Nil(t, result)
				return
			}

			require.NoError(t, err)
			assert.Len(t, result, len(tt.priceUpdates))

			// Validate each update in the result
			for _, update := range result {
				// Verify basic structure
				assert.NotNil(t, update.TemporalNumericValue.QuantizedValue)
				assert.Greater(t, update.TemporalNumericValue.TimestampNs, uint64(0))

				// Verify V conversion (0x1c should convert to 28)
				if len(tt.priceUpdates) > 0 {
					if tt.name == "valid price with V=0x1b" {
						assert.Equal(t, uint8(27), update.V)
					} else {
						assert.Equal(t, uint8(28), update.V)
					}
				}
			}
		})
	}
}

func TestGetVerifyPublishersPayloads(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		priceUpdates map[types.InternalEncodedAssetId]types.AggregatedSignedPrice
		wantError    bool
	}{
		{
			name: "valid price update with signatures",
			priceUpdates: map[types.InternalEncodedAssetId]types.AggregatedSignedPrice{
				[32]byte{1, 2, 3, 4}: {
					StorkSignedPrice: &types.StorkSignedPrice{
						PublisherMerkleRoot: "0xe5ff773b0316059c04aa157898766731017610dcbeede7d7f169bfeaab7cc318",
					},
					SignedPrices: []*types.PublisherSignedPrice{
						{
							PublisherKey:    "0x1234567890123456789012345678901234567890",
							ExternalAssetId: "BTCUSD",
							QuantizedPrice:  "1000000000000000000",
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
				},
			},
			wantError: false,
		},
		{
			name: "valid price update with V=0x1b",
			priceUpdates: map[types.InternalEncodedAssetId]types.AggregatedSignedPrice{
				[32]byte{5, 6, 7, 8}: {
					StorkSignedPrice: &types.StorkSignedPrice{
						PublisherMerkleRoot: "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
					},
					SignedPrices: []*types.PublisherSignedPrice{
						{
							PublisherKey:    "0xabcdabcdabcdabcdabcdabcdabcdabcdabcdabcd",
							ExternalAssetId: "ETHUSD",
							QuantizedPrice:  "500000000000000000",
							TimestampedSignature: types.TimestampedSignature{
								TimestampNano: 1722632569208762200,
								Signature: types.EvmSignature{
									R: "0x1111111111111111111111111111111111111111111111111111111111111111",
									S: "0x2222222222222222222222222222222222222222222222222222222222222222",
									V: "0x1b",
								},
							},
						},
					},
				},
			},
			wantError: false,
		},
		{
			name: "multiple signed prices",
			priceUpdates: map[types.InternalEncodedAssetId]types.AggregatedSignedPrice{
				[32]byte{1, 2, 3, 4}: {
					StorkSignedPrice: &types.StorkSignedPrice{
						PublisherMerkleRoot: "0xe5ff773b0316059c04aa157898766731017610dcbeede7d7f169bfeaab7cc318",
					},
					SignedPrices: []*types.PublisherSignedPrice{
						{
							PublisherKey:    "0x1234567890123456789012345678901234567890",
							ExternalAssetId: "BTCUSD",
							QuantizedPrice:  "1000000000000000000",
							TimestampedSignature: types.TimestampedSignature{
								TimestampNano: 1722632569208762117,
								Signature: types.EvmSignature{
									R: "0xb9b3c9f80a355bd0cd6f609fff4a4b15fa4e3b4632adabb74c020f5bcd240741",
									S: "0x16fab526529ac795108d201832cff8c2d2b1c710da6711fe9f7ab288a7149758",
									V: "0x1c",
								},
							},
						},
						{
							PublisherKey:    "0x5678567890123456789012345678901234567890",
							ExternalAssetId: "ETHUSD",
							QuantizedPrice:  "2000000000000000000",
							TimestampedSignature: types.TimestampedSignature{
								TimestampNano: 1722632569208762118,
								Signature: types.EvmSignature{
									R: "0xb9b3c9f80a355bd0cd6f609fff4a4b15fa4e3b4632adabb74c020f5bcd240741",
									S: "0x16fab526529ac795108d201832cff8c2d2b1c710da6711fe9f7ab288a7149758",
									V: "0x1c",
								},
							},
						},
					},
				},
			},
			wantError: false,
		},
		{
			name: "invalid merkle root",
			priceUpdates: map[types.InternalEncodedAssetId]types.AggregatedSignedPrice{
				[32]byte{1, 2, 3, 4}: {
					StorkSignedPrice: &types.StorkSignedPrice{
						PublisherMerkleRoot: "invalid hex",
					},
					SignedPrices: []*types.PublisherSignedPrice{
						{
							PublisherKey:    "0x1234567890123456789012345678901234567890",
							ExternalAssetId: "BTCUSD",
							QuantizedPrice:  "1000000000000000000",
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
				},
			},
			wantError: true,
		},
		{
			name: "invalid publisher key",
			priceUpdates: map[types.InternalEncodedAssetId]types.AggregatedSignedPrice{
				[32]byte{1, 2, 3, 4}: {
					StorkSignedPrice: &types.StorkSignedPrice{
						PublisherMerkleRoot: "0xe5ff773b0316059c04aa157898766731017610dcbeede7d7f169bfeaab7cc318",
					},
					SignedPrices: []*types.PublisherSignedPrice{
						{
							PublisherKey:    "invalid hex",
							ExternalAssetId: "BTCUSD",
							QuantizedPrice:  "1000000000000000000",
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
				},
			},
			wantError: true,
		},
		{
			name:         "empty price updates",
			priceUpdates: map[types.InternalEncodedAssetId]types.AggregatedSignedPrice{},
			wantError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := getVerifyPublishersPayloads(tt.priceUpdates)

			if tt.wantError {
				assert.Error(t, err)
				assert.Nil(t, result)
				return
			}

			require.NoError(t, err)
			assert.Len(t, result, len(tt.priceUpdates))

			// Validate each payload in the result
			for i, payload := range result {
				assert.NotNil(t, payload.merkleRoot)

				// Find corresponding price update to validate signature count
				var expectedSigCount int
				for _, priceUpdate := range tt.priceUpdates {
					expectedSigCount = len(priceUpdate.SignedPrices)
					break
				}

				if i < len(result) {
					assert.Len(t, payload.pubSigs, expectedSigCount)
				}

				// Validate publisher signatures structure
				for _, pubSig := range payload.pubSigs {
					assert.NotNil(t, pubSig.QuantizedValue)
					assert.Greater(t, pubSig.Timestamp, uint64(0))
					assert.NotEmpty(t, pubSig.AssetPairId)
				}
			}
		})
	}
}
