package pusher

import (
	"math/big"
	"os"
	"path/filepath"
	"testing"

	"github.com/Stork-Oracle/stork-external/apps/chain_pusher/pkg/types"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShouldUpdateAsset(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name               string
		currentTimestampNs uint64
		currentValueStr    string
		storkTimestampNano uint64
		storkPriceStr      string
		fallbackPeriodSecs uint64
		changeThreshold    float64
		expected           bool
	}{
		{
			name:               "update due to time threshold",
			currentTimestampNs: 1000000000000,         // 1000 seconds in nanoseconds
			currentValueStr:    "1000000000000000000", // 1.0
			storkTimestampNano: 1301000000000,         // 301 seconds later
			storkPriceStr:      "1000000000000000000", // Same price
			fallbackPeriodSecs: 300,                   // 300 seconds
			changeThreshold:    1.0,
			expected:           true,
		},
		{
			name:               "no update - within thresholds",
			currentTimestampNs: 1000000000000,
			currentValueStr:    "1000000000000000000", // 1.0
			storkTimestampNano: 1100000000000,         // 100 seconds later
			storkPriceStr:      "1005000000000000000", // 0.5% increase
			fallbackPeriodSecs: 300,
			changeThreshold:    1.0, // 1.0% threshold
			expected:           false,
		},
		{
			name:               "update due to price increase",
			currentTimestampNs: 1000000000000,
			currentValueStr:    "1000000000000000000", // 1.0
			storkTimestampNano: 1100000000000,         // 100 seconds later
			storkPriceStr:      "1020000000000000000", // 2.0% increase
			fallbackPeriodSecs: 300,
			changeThreshold:    1.0, // 1.0% threshold
			expected:           true,
		},
		{
			name:               "update due to price decrease",
			currentTimestampNs: 1000000000000,
			currentValueStr:    "1000000000000000000", // 1.0
			storkTimestampNano: 1100000000000,
			storkPriceStr:      "980000000000000000", // 2.0% decrease
			fallbackPeriodSecs: 300,
			changeThreshold:    1.0,
			expected:           true,
		},
		{
			name:               "update when current value is zero",
			currentTimestampNs: 1000000000000,
			currentValueStr:    "0", // 0.0
			storkTimestampNano: 1100000000000,
			storkPriceStr:      "1000000000000000000", // 1.0
			fallbackPeriodSecs: 300,
			changeThreshold:    1.0,
			expected:           true,
		},
		{
			name:               "no update when both values are zero",
			currentTimestampNs: 1000000000000,
			currentValueStr:    "0", // 0.0
			storkTimestampNano: 1100000000000,
			storkPriceStr:      "0", // 0.0
			fallbackPeriodSecs: 300,
			changeThreshold:    1.0,
			expected:           false,
		},
		{
			name:               "update with small threshold",
			currentTimestampNs: 1000000000000,
			currentValueStr:    "1000000000000000000", // 1.0
			storkTimestampNano: 1100000000000,
			storkPriceStr:      "1001000000000000000", // 0.1% increase
			fallbackPeriodSecs: 300,
			changeThreshold:    0.05, // 0.05% threshold
			expected:           true,
		},
		{
			name:               "no update with high threshold",
			currentTimestampNs: 1000000000000,
			currentValueStr:    "1000000000000000000", // 1.0
			storkTimestampNano: 1100000000000,
			storkPriceStr:      "1040000000000000000", // 4.0% increase
			fallbackPeriodSecs: 300,
			changeThreshold:    5.0, // 5.0% threshold
			expected:           false,
		},
		{
			name:               "same timestamp no update",
			currentTimestampNs: 1000000000000,
			currentValueStr:    "1000000000000000000",
			storkTimestampNano: 1000000000000,         // Exact same timestamp
			storkPriceStr:      "1000000000000000000", // 0.0% increase
			fallbackPeriodSecs: 300,
			changeThreshold:    1.0,
			expected:           false,
		},
		{
			name:               "zero fallback period triggers on time",
			currentTimestampNs: 1000000000000,
			currentValueStr:    "1000000000000000000",
			storkTimestampNano: 2000000000000,         // Much later
			storkPriceStr:      "1000000000000000000", // 0.0% increase
			fallbackPeriodSecs: 0,                     // No time-based updates
			changeThreshold:    1.0,
			expected:           true, // Any time difference > 0
		},
		{
			name:               "negative to more negative update",
			currentTimestampNs: 1000000000000,
			currentValueStr:    "-1000000000000000000", // -1.0
			storkTimestampNano: 1100000000000,
			storkPriceStr:      "-1200000000000000000", // -1.2 (20% more negative)
			fallbackPeriodSecs: 300,
			changeThreshold:    1.0,
			expected:           true,
		},
		{
			name:               "negative to less negative no update",
			currentTimestampNs: 1000000000000,
			currentValueStr:    "-1000000000000000000", // -1.0
			storkTimestampNano: 1100000000000,
			storkPriceStr:      "-1005000000000000000", // -1.005 (0.5% change)
			fallbackPeriodSecs: 300,
			changeThreshold:    1.0,
			expected:           false,
		},
		{
			name:               "negative to positive triggers",
			currentTimestampNs: 1000000000000,
			currentValueStr:    "-1000000000000000000", // -1.0
			storkTimestampNano: 1100000000000,
			storkPriceStr:      "1000000000000000000", // 1.0
			fallbackPeriodSecs: 300,
			changeThreshold:    1.0,
			expected:           true,
		},
		{
			name:               "positive to negative triggers",
			currentTimestampNs: 1000000000000,
			currentValueStr:    "1000000000000000000", // 1.0
			storkTimestampNano: 1100000000000,
			storkPriceStr:      "-1000000000000000000", // -1.0
			fallbackPeriodSecs: 300,
			changeThreshold:    1.0,
			expected:           true,
		},
		{
			name:               "zero threshold any change triggers",
			currentTimestampNs: 1000000000000,
			currentValueStr:    "1000000000000000000",
			storkTimestampNano: 1100000000000,
			storkPriceStr:      "1000000000000000001", // Tiny increase
			fallbackPeriodSecs: 300,
			changeThreshold:    0.0,
			expected:           true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create big.Int from string in test execution
			currentValue := new(big.Int)
			currentValue.SetString(tt.currentValueStr, 10)

			latestValue := types.InternalTemporalNumericValue{
				TimestampNs:    tt.currentTimestampNs,
				QuantizedValue: currentValue,
			}

			latestStorkPrice := types.AggregatedSignedPrice{
				TimestampNano: tt.storkTimestampNano,
				StorkSignedPrice: &types.StorkSignedPrice{
					QuantizedPrice: types.QuantizedPrice(tt.storkPriceStr),
				},
			}

			result := shouldUpdateAsset(latestValue, latestStorkPrice, tt.fallbackPeriodSecs, tt.changeThreshold)

			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestInitializeAssets(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		configContent   string
		expectedAssets  []types.AssetID
		expectedEncoded []types.InternalEncodedAssetID
		wantError       bool
		errorContains   string
	}{
		{
			name: "valid config with multiple assets",
			configContent: `assets:
  BTCUSD:
    asset_id: "BTCUSD"
    encoded_asset_id: "0x7404e3d104ea7841c3d9e6fd20adfe99b4ad586bc08d8f3bd3afef894cf184de"
    percent_change_threshold: 1.5
    fallback_period_sec: 300
  ETHUSD:
    asset_id: "ETHUSD"
    encoded_asset_id: "0x59102b37de83bdda9f38ac8254e596f0d9ac61d2035c07936675e87342817160"
    percent_change_threshold: 2.0
    fallback_period_sec: 600`,
			expectedAssets: []types.AssetID{"BTCUSD", "ETHUSD"},
			expectedEncoded: []types.InternalEncodedAssetID{
				{
					0x74,
					0x04,
					0xe3,
					0xd1,
					0x04,
					0xea,
					0x78,
					0x41,
					0xc3,
					0xd9,
					0xe6,
					0xfd,
					0x20,
					0xad,
					0xfe,
					0x99,
					0xb4,
					0xad,
					0x58,
					0x6b,
					0xc0,
					0x8d,
					0x8f,
					0x3b,
					0xd3,
					0xaf,
					0xef,
					0x89,
					0x4c,
					0xf1,
					0x84,
					0xde,
				},
				{
					0x59,
					0x10,
					0x2b,
					0x37,
					0xde,
					0x83,
					0xbd,
					0xda,
					0x9f,
					0x38,
					0xac,
					0x82,
					0x54,
					0xe5,
					0x96,
					0xf0,
					0xd9,
					0xac,
					0x61,
					0xd2,
					0x03,
					0x5c,
					0x07,
					0x93,
					0x66,
					0x75,
					0xe8,
					0x73,
					0x42,
					0x81,
					0x71,
					0x60,
				},
			},
			wantError: false,
		},
		{
			name: "valid config with single asset",
			configContent: `assets:
  SOLUSD:
    asset_id: "SOLUSD"
    encoded_asset_id: "0x1dcd89dfded9e8a9b0fa1745a8ebbacbb7c81e33d5abc81616633206d932e837"
    percent_change_threshold: 0.5
    fallback_period_sec: 120`,
			expectedAssets: []types.AssetID{"SOLUSD"},
			expectedEncoded: []types.InternalEncodedAssetID{
				{
					0x1d,
					0xcd,
					0x89,
					0xdf,
					0xde,
					0xd9,
					0xe8,
					0xa9,
					0xb0,
					0xfa,
					0x17,
					0x45,
					0xa8,
					0xeb,
					0xba,
					0xcb,
					0xb7,
					0xc8,
					0x1e,
					0x33,
					0xd5,
					0xab,
					0xc8,
					0x16,
					0x16,
					0x63,
					0x32,
					0x06,
					0xd9,
					0x32,
					0xe8,
					0x37,
				},
			},
			wantError: false,
		},
		{
			name:            "empty config",
			configContent:   `assets: {}`,
			expectedAssets:  []types.AssetID{},
			expectedEncoded: []types.InternalEncodedAssetID{},
			wantError:       false,
		},
		{
			name: "invalid hex encoded asset id",
			configContent: `assets:
  TEST:
    asset_id: "TEST"
    encoded_asset_id: "0xINVALID_HEX"
    percent_change_threshold: 1.0
    fallback_period_sec: 300`,
			expectedAssets:  []types.AssetID{},
			expectedEncoded: []types.InternalEncodedAssetID{},
			wantError:       true,
			errorContains:   "failed to decode hex string",
		},
		{
			name: "short hex encoded asset id",
			configContent: `assets:
  TEST:
    asset_id: "TEST"
    encoded_asset_id: "0x1234567890123456789012345678901234567890123456789012345678901234"
    percent_change_threshold: 1.0
    fallback_period_sec: 300`,
			expectedAssets: []types.AssetID{"TEST"},
			expectedEncoded: []types.InternalEncodedAssetID{
				{
					0x12,
					0x34,
					0x56,
					0x78,
					0x90,
					0x12,
					0x34,
					0x56,
					0x78,
					0x90,
					0x12,
					0x34,
					0x56,
					0x78,
					0x90,
					0x12,
					0x34,
					0x56,
					0x78,
					0x90,
					0x12,
					0x34,
					0x56,
					0x78,
					0x90,
					0x12,
					0x34,
					0x56,
					0x78,
					0x90,
					0x12,
					0x34,
				},
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create temporary config file
			tempDir := t.TempDir()
			configFile := filepath.Join(tempDir, "test_config.yaml")

			err := os.WriteFile(configFile, []byte(tt.configContent), 0o600)
			require.NoError(t, err)

			// Create pusher with test config
			logger := zerolog.Nop()
			pusher := &Pusher{
				assetConfigFile: configFile,
				logger:          &logger,
			}

			// Test the function
			priceConfig, assetIDs, encodedAssetIDs, err := pusher.initializeAssets()

			if tt.wantError {
				require.Error(t, err)

				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}

				return
			}

			require.NoError(t, err)
			require.NotNil(t, priceConfig)

			// Use expected pattern - compare exact results
			assert.ElementsMatch(t, tt.expectedAssets, assetIDs)
			assert.ElementsMatch(t, tt.expectedEncoded, encodedAssetIDs)
			assert.Len(t, priceConfig.Assets, len(tt.expectedAssets))
		})
	}
}

func TestInitializeAssets_FileNotFound(t *testing.T) {
	t.Parallel()

	logger := zerolog.Nop()
	pusher := &Pusher{
		assetConfigFile: "/nonexistent/path/config.yaml",
		logger:          &logger,
	}

	assetConfig, assetIDs, encodedAssetIDs, err := pusher.initializeAssets()
	_ = assetConfig
	_ = assetIDs
	_ = encodedAssetIDs

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load price config")
}

func TestHandleStorkUpdate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		valueUpdate     types.AggregatedSignedPrice
		initialMap      map[types.InternalEncodedAssetID]types.AggregatedSignedPrice
		expectedMapSize int
		wantError       bool
		expectedAssetID string
	}{
		{
			name: "successful update with valid hex",
			valueUpdate: types.AggregatedSignedPrice{
				AssetID:       "BTCUSD",
				TimestampNano: 1234567890,
				StorkSignedPrice: &types.StorkSignedPrice{
					EncodedAssetID: "0x7404e3d104ea7841c3d9e6fd20adfe99b4ad586bc08d8f3bd3afef894cf184de",
					QuantizedPrice: "1000000000000000000",
				},
			},
			initialMap:      make(map[types.InternalEncodedAssetID]types.AggregatedSignedPrice),
			expectedMapSize: 1,
			wantError:       false,
			expectedAssetID: "BTCUSD",
		},
		{
			name: "update existing asset",
			valueUpdate: types.AggregatedSignedPrice{
				AssetID:       "ETHUSD",
				TimestampNano: 9876543210,
				StorkSignedPrice: &types.StorkSignedPrice{
					EncodedAssetID: "0x59102b37de83bdda9f38ac8254e596f0d9ac61d2035c07936675e87342817160",
					QuantizedPrice: "2000000000000000000",
				},
			},
			initialMap: map[types.InternalEncodedAssetID]types.AggregatedSignedPrice{
				{0x59, 0x10, 0x2b, 0x37, 0xde, 0x83, 0xbd, 0xda, 0x9f, 0x38, 0xac, 0x82, 0x54, 0xe5, 0x96, 0xf0, 0xd9, 0xac, 0x61, 0xd2, 0x03, 0x5c, 0x07, 0x93, 0x66, 0x75, 0xe8, 0x73, 0x42, 0x81, 0x71, 0x60}: {
					AssetID:       "ETHUSD",
					TimestampNano: 1111111111,
					StorkSignedPrice: &types.StorkSignedPrice{
						EncodedAssetID: "0x59102b37de83bdda9f38ac8254e596f0d9ac61d2035c07936675e87342817160",
						QuantizedPrice: "1500000000000000000",
					},
				},
			},
			expectedMapSize: 1, // Same size because we're updating existing
			wantError:       false,
			expectedAssetID: "ETHUSD",
		},
		{
			name: "invalid hex encoded asset id",
			valueUpdate: types.AggregatedSignedPrice{
				AssetID:       "INVALID",
				TimestampNano: 1111111111,
				StorkSignedPrice: &types.StorkSignedPrice{
					EncodedAssetID: "INVALID_HEX",
					QuantizedPrice: "1000000000000000000",
				},
			},
			initialMap:      make(map[types.InternalEncodedAssetID]types.AggregatedSignedPrice),
			expectedMapSize: 0, // No change due to error
			wantError:       true,
		},
		{
			name: "hex without 0x prefix",
			valueUpdate: types.AggregatedSignedPrice{
				AssetID:       "SOLUSD",
				TimestampNano: 5555555555,
				StorkSignedPrice: &types.StorkSignedPrice{
					EncodedAssetID: "0x1dcd89dfded9e8a9b0fa1745a8ebbacbb7c81e33d5abc81616633206d932e837",
					QuantizedPrice: "3000000000000000000",
				},
			},
			initialMap:      make(map[types.InternalEncodedAssetID]types.AggregatedSignedPrice),
			expectedMapSize: 1,
			wantError:       false,
			expectedAssetID: "SOLUSD",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			logger := zerolog.Nop()
			pusher := &Pusher{
				logger: &logger,
			}

			// Use the provided initial map directly
			latestStorkValueMap := tt.initialMap
			if latestStorkValueMap == nil {
				latestStorkValueMap = make(map[types.InternalEncodedAssetID]types.AggregatedSignedPrice)
			}

			// Call the function
			pusher.handleStorkUpdate(tt.valueUpdate, latestStorkValueMap)

			// Verify results
			assert.Len(t, latestStorkValueMap, tt.expectedMapSize)

			if !tt.wantError {
				// Find the updated entry
				found := false

				for _, value := range latestStorkValueMap {
					if value.AssetID == types.AssetID(tt.expectedAssetID) {
						found = true

						assert.Equal(t, tt.valueUpdate.TimestampNano, value.TimestampNano)
						assert.Equal(
							t,
							tt.valueUpdate.StorkSignedPrice.QuantizedPrice,
							value.StorkSignedPrice.QuantizedPrice,
						)

						break
					}
				}

				assert.True(t, found, "Updated value should be in map")
			}
		})
	}
}

func TestHandleContractUpdate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                  string
		chainUpdate           map[types.InternalEncodedAssetID]types.InternalTemporalNumericValue
		initialMapSize        int
		expectedFinalMapSize  int
		expectedUpdatedAssets int
	}{
		{
			name: "single asset update",
			chainUpdate: map[types.InternalEncodedAssetID]types.InternalTemporalNumericValue{
				{0x12, 0x34}: {
					TimestampNs:    1234567890,
					QuantizedValue: big.NewInt(1000000000000000000),
				},
			},
			initialMapSize:        0,
			expectedFinalMapSize:  1,
			expectedUpdatedAssets: 1,
		},
		{
			name: "multiple asset updates",
			chainUpdate: map[types.InternalEncodedAssetID]types.InternalTemporalNumericValue{
				{0x12, 0x34}: {
					TimestampNs:    1234567890,
					QuantizedValue: big.NewInt(1000000000000000000),
				},
				{0xab, 0xcd}: {
					TimestampNs:    9876543210,
					QuantizedValue: big.NewInt(2000000000000000000),
				},
				{0xff, 0xee}: {
					TimestampNs:    5555555555,
					QuantizedValue: big.NewInt(3000000000000000000),
				},
			},
			initialMapSize:        0,
			expectedFinalMapSize:  3,
			expectedUpdatedAssets: 3,
		},
		{
			name: "update existing assets",
			chainUpdate: map[types.InternalEncodedAssetID]types.InternalTemporalNumericValue{
				{0x12, 0x34}: {
					TimestampNs:    9999999999,
					QuantizedValue: big.NewInt(5000000000000000000),
				},
			},
			initialMapSize:        2,
			expectedFinalMapSize:  2, // No new assets added
			expectedUpdatedAssets: 1,
		},
		{
			name:                  "empty update",
			chainUpdate:           map[types.InternalEncodedAssetID]types.InternalTemporalNumericValue{},
			initialMapSize:        1,
			expectedFinalMapSize:  1, // No change
			expectedUpdatedAssets: 0,
		},
		{
			name: "mixed new and existing assets",
			chainUpdate: map[types.InternalEncodedAssetID]types.InternalTemporalNumericValue{
				{0x12, 0x34}: { // Existing
					TimestampNs:    9999999999,
					QuantizedValue: big.NewInt(5000000000000000000),
				},
				{0x99, 0x88}: { // New
					TimestampNs:    1111111111,
					QuantizedValue: big.NewInt(6000000000000000000),
				},
			},
			initialMapSize:        1,
			expectedFinalMapSize:  2,
			expectedUpdatedAssets: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			logger := zerolog.Nop()
			pusher := &Pusher{
				logger: &logger,
			}

			// Initialize map with some data if needed
			latestContractValueMap := make(map[types.InternalEncodedAssetID]types.InternalTemporalNumericValue)

			if tt.initialMapSize > 0 {
				// Add some dummy entries
				for i := range tt.initialMapSize {
					key := [32]byte{}
					key[0] = byte(0x12)

					key[1] = byte(0x34)
					if i > 0 {
						key[1] = byte(0x34 + i) // Make keys unique
					}

					latestContractValueMap[key] = types.InternalTemporalNumericValue{
						//nolint:gosec // "i" being a valid uint64 is controlled by the test definitions.
						TimestampNs:    1000000000 + (uint64(i)),
						QuantizedValue: big.NewInt(int64(1000000000000000000 + i)),
					}
				}
			}

			initialSize := len(latestContractValueMap)
			assert.Equal(t, tt.initialMapSize, initialSize)

			// Call the function
			pusher.handleContractUpdate(tt.chainUpdate, latestContractValueMap)

			// Verify results
			assert.Len(t, latestContractValueMap, tt.expectedFinalMapSize)

			// Verify that all updates were applied
			for expectedAssetID, expectedValue := range tt.chainUpdate {
				actualValue, exists := latestContractValueMap[expectedAssetID]
				assert.True(t, exists, "Updated asset should exist in map")
				assert.Equal(t, expectedValue.TimestampNs, actualValue.TimestampNs)
				assert.Equal(t, expectedValue.QuantizedValue.String(), actualValue.QuantizedValue.String())
			}
		})
	}
}
