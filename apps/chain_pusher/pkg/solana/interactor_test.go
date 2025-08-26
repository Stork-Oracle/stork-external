package solana

import (
	"testing"

	"github.com/Stork-Oracle/stork-external/apps/chain_pusher/pkg/pusher"
	"github.com/Stork-Oracle/stork-external/apps/chain_pusher/pkg/types"
	bin "github.com/gagliardetto/binary"
	"github.com/rs/zerolog"
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
			priceUpdates := make(map[types.InternalEncodedAssetId]types.AggregatedSignedPrice)
			for i := 0; i < tt.numUpdates; i++ {
				var assetId types.InternalEncodedAssetId
				assetId[0] = byte(i + 1) // Create unique asset IDs
				priceUpdates[assetId] = types.AggregatedSignedPrice{}
			}

			sci := &ContractInteractor{
				batchSize: tt.batchSize,
			}

			batches := sci.batchPriceUpdates(priceUpdates)

			assert.Equal(t, tt.expectedBatches, len(batches))

			// Verify each batch size is correct
			for i, batch := range batches {
				if i == len(batches)-1 {
					// Last batch might be smaller
					assert.LessOrEqual(t, len(batch), tt.batchSize)
				} else {
					assert.Equal(t, tt.batchSize, len(batch))
				}
			}
		})
	}
}

func TestQuantizedPriceToInt128(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		quantizedPrice string
		expected       bin.Int128
	}{
		{"zero value", "0", bin.Int128{Lo: 0, Hi: 0}},
		{"small number", "1000000000000000000", bin.Int128{Lo: 1000000000000000000, Hi: 0}},
		{"large number", "115792089237316195423570985008687907853269984665640564039457584007913129639935", bin.Int128{Lo: 18446744073709551615, Hi: 18446744073709551615}},
	}

	sci := &ContractInteractor{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := sci.quantizedPriceToInt128(types.QuantizedPrice(tt.quantizedPrice))
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPriceUpdateToTemporalNumericValueEvmInput(t *testing.T) {
	t.Parallel()

	logger := zerolog.New(nil)
	sci := &ContractInteractor{
		logger: logger,
	}

	tests := []struct {
		name        string
		priceUpdate types.AggregatedSignedPrice
		assetId     types.InternalEncodedAssetId
		treasuryId  uint8
		wantError   bool
	}{
		{
			name: "valid input",
			priceUpdate: types.AggregatedSignedPrice{
				StorkSignedPrice: &types.StorkSignedPrice{
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
			assetId:    [32]byte{1, 2, 3, 4},
			treasuryId: 1,
			wantError:  false,
		},
		{
			name: "invalid hex in merkle root",
			priceUpdate: types.AggregatedSignedPrice{
				StorkSignedPrice: &types.StorkSignedPrice{
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
			assetId:    [32]byte{1, 2, 3, 4},
			treasuryId: 1,
			wantError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := sci.priceUpdateToTemporalNumericValueEvmInput(tt.priceUpdate, tt.assetId, tt.treasuryId)

			if tt.wantError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)

			// Validate basic fields
			assert.Equal(t, tt.assetId[:], result.Id[:])
			assert.Equal(t, uint64(tt.priceUpdate.StorkSignedPrice.TimestampedSignature.TimestampNano), result.TemporalNumericValue.TimestampNs)
			assert.Equal(t, tt.treasuryId, result.TreasuryId)

			// Verify signature components were properly converted
			rBytes, err := pusher.HexStringToByteArray(tt.priceUpdate.StorkSignedPrice.TimestampedSignature.Signature.R)
			require.NoError(t, err)
			sBytes, err := pusher.HexStringToByteArray(tt.priceUpdate.StorkSignedPrice.TimestampedSignature.Signature.S)
			require.NoError(t, err)

			assert.Equal(t, rBytes, result.R[:len(rBytes)])
			assert.Equal(t, sBytes, result.S[:len(sBytes)])
		})
	}
}
