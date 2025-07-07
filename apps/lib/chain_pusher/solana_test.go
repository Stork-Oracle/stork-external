package chain_pusher

import (
	"bytes"
	"reflect"
	"testing"

	bin "github.com/gagliardetto/binary"
	"github.com/rs/zerolog"
)

func TestHexStringToByteArray(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []byte
		wantErr bool
	}{
		{
			name:    "valid hex string",
			input:   "0x1234",
			want:    []byte{0x12, 0x34},
			wantErr: false,
		},
		{
			name:    "valid hex string without prefix",
			input:   "1234",
			want:    []byte{0x12, 0x34},
			wantErr: false,
		},
		{
			name:    "invalid hex string",
			input:   "0xZZ",
			want:    []byte{},
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   "",
			want:    []byte{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := hexStringToByteArray(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("hexStringToByteArray() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("hexStringToByteArray() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBatchPriceUpdates(t *testing.T) {
	tests := []struct {
		name         string
		batchSize    int
		priceUpdates map[InternalEncodedAssetId]AggregatedSignedPrice
		wantBatches  int
	}{
		{
			name:         "empty updates",
			batchSize:    2,
			priceUpdates: map[InternalEncodedAssetId]AggregatedSignedPrice{},
			wantBatches:  0,
		},
		{
			name:      "single update",
			batchSize: 2,
			priceUpdates: map[InternalEncodedAssetId]AggregatedSignedPrice{
				{1}: {},
			},
			wantBatches: 1,
		},
		{
			name:      "exact batch size",
			batchSize: 2,
			priceUpdates: map[InternalEncodedAssetId]AggregatedSignedPrice{
				{1}: {},
				{2}: {},
			},
			wantBatches: 1,
		},
		{
			name:      "multiple batches",
			batchSize: 2,
			priceUpdates: map[InternalEncodedAssetId]AggregatedSignedPrice{
				{1}: {},
				{2}: {},
				{3}: {},
			},
			wantBatches: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sci := &SolanaContractInteractor{
				batchSize: tt.batchSize,
			}
			batches := sci.batchPriceUpdates(tt.priceUpdates)
			if len(batches) != tt.wantBatches {
				t.Errorf("batchPriceUpdates() got %v batches, want %v", len(batches), tt.wantBatches)
			}

			// Verify each batch size is correct
			for i, batch := range batches {
				if i == len(batches)-1 {
					// Last batch might be smaller
					if len(batch) > tt.batchSize {
						t.Errorf("Last batch size %v exceeds max batch size %v", len(batch), tt.batchSize)
					}
				} else {
					if len(batch) != tt.batchSize {
						t.Errorf("Batch %v size = %v, want %v", i, len(batch), tt.batchSize)
					}
				}
			}
		})
	}
}

func TestQuantizedPriceToInt128(t *testing.T) {
	tests := []struct {
		name           string
		quantizedPrice QuantizedPrice
		want           bin.Int128
	}{
		{
			name:           "zero value",
			quantizedPrice: "0",
			want:           bin.Int128{Lo: 0, Hi: 0},
		},
		{
			name:           "small number",
			quantizedPrice: "1000000000000000000", // 1 with 18 decimals
			want:           bin.Int128{Lo: 1000000000000000000, Hi: 0},
		},
		{
			name:           "large number",
			quantizedPrice: "115792089237316195423570985008687907853269984665640564039457584007913129639935",
			want:           bin.Int128{Lo: 18446744073709551615, Hi: 18446744073709551615}, // max uint128
		},
	}

	sci := &SolanaContractInteractor{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sci.quantizedPriceToInt128(tt.quantizedPrice)
			if got != tt.want {
				t.Errorf("quantizedPriceToInt128() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPriceUpdateToTemporalNumericValueEvmInput(t *testing.T) {
	logger := zerolog.New(nil)
	sci := &SolanaContractInteractor{
		logger: logger,
	}

	tests := []struct {
		name        string
		priceUpdate AggregatedSignedPrice
		assetId     InternalEncodedAssetId
		treasuryId  uint8
		wantErr     bool
	}{
		{
			name: "valid input",
			priceUpdate: AggregatedSignedPrice{
				StorkSignedPrice: &StorkSignedPrice{
					QuantizedPrice:      "1000000000000000000",
					PublisherMerkleRoot: "0xe5ff773b0316059c04aa157898766731017610dcbeede7d7f169bfeaab7cc318",
					StorkCalculationAlg: StorkCalculationAlg{
						Checksum: "0x9be7e9f9ed459417d96112a7467bd0b27575a2c7847195c68f805b70ce1795ba",
					},
					TimestampedSignature: TimestampedSignature{
						TimestampNano: 1722632569208762117,
						Signature: EvmSignature{
							R: "0xb9b3c9f80a355bd0cd6f609fff4a4b15fa4e3b4632adabb74c020f5bcd240741",
							S: "0x16fab526529ac795108d201832cff8c2d2b1c710da6711fe9f7ab288a7149758",
							V: "0x1c",
						},
					},
				},
			},
			assetId:    [32]byte{1, 2, 3, 4}, // example asset ID
			treasuryId: 1,
			wantErr:    false,
		},
		{
			name: "invalid hex in merkle root",
			priceUpdate: AggregatedSignedPrice{
				StorkSignedPrice: &StorkSignedPrice{
					QuantizedPrice:      "1000000000000000000",
					PublisherMerkleRoot: "invalid hex",
					StorkCalculationAlg: StorkCalculationAlg{
						Checksum: "0x9be7e9f9ed459417d96112a7467bd0b27575a2c7847195c68f805b70ce1795ba",
					},
					TimestampedSignature: TimestampedSignature{
						TimestampNano: 1722632569208762117,
						Signature: EvmSignature{
							R: "0xb9b3c9f80a355bd0cd6f609fff4a4b15fa4e3b4632adabb74c020f5bcd240741",
							S: "0x16fab526529ac795108d201832cff8c2d2b1c710da6711fe9f7ab288a7149758",
							V: "0x1c",
						},
					},
				},
			},
			assetId:    [32]byte{1, 2, 3, 4},
			treasuryId: 1,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := sci.priceUpdateToTemporalNumericValueEvmInput(tt.priceUpdate, tt.assetId, tt.treasuryId)
			if (err != nil) != tt.wantErr {
				t.Errorf("priceUpdateToTemporalNumericValueEvmInput() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			if !bytes.Equal(tt.assetId[:], got.Id[:]) {
				t.Errorf("AssetId = %v, want %v", got.Id, tt.assetId)
			}
			if got.TemporalNumericValue.TimestampNs != uint64(tt.priceUpdate.StorkSignedPrice.TimestampedSignature.TimestampNano) {
				t.Errorf("TimestampNs = %v, want %v", got.TemporalNumericValue.TimestampNs, tt.priceUpdate.StorkSignedPrice.TimestampedSignature.TimestampNano)
			}
			if got.TreasuryId != tt.treasuryId {
				t.Errorf("TreasuryId = %v, want %v", got.TreasuryId, tt.treasuryId)
			}

			// Verify the signature components were properly converted
			rBytes, _ := hexStringToByteArray(tt.priceUpdate.StorkSignedPrice.TimestampedSignature.Signature.R)
			sBytes, _ := hexStringToByteArray(tt.priceUpdate.StorkSignedPrice.TimestampedSignature.Signature.S)
			if !bytes.Equal(rBytes, got.R[:len(rBytes)]) {
				t.Errorf("R = %v, want %v", got.R[:len(rBytes)], rBytes)
			}
			if !bytes.Equal(sBytes, got.S[:len(sBytes)]) {
				t.Errorf("S = %v, want %v", got.S[:len(sBytes)], sBytes)
			}
		})
	}
}
