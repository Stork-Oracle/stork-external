package chain_pusher

import (
	"reflect"
	"testing"
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
				InternalEncodedAssetId{1}: {},
			},
			wantBatches: 1,
		},
		{
			name:      "exact batch size",
			batchSize: 2,
			priceUpdates: map[InternalEncodedAssetId]AggregatedSignedPrice{
				InternalEncodedAssetId{1}: {},
				InternalEncodedAssetId{2}: {},
			},
			wantBatches: 1,
		},
		{
			name:      "multiple batches",
			batchSize: 2,
			priceUpdates: map[InternalEncodedAssetId]AggregatedSignedPrice{
				InternalEncodedAssetId{1}: {},
				InternalEncodedAssetId{2}: {},
				InternalEncodedAssetId{3}: {},
			},
			wantBatches: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sci := &SolanaContractInteracter{
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
