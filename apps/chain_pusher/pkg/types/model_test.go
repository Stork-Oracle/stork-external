package types

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		fileContent string
		fileName    string
		expected    *AssetConfig
		wantError   bool
	}{
		{
			name: "valid config",
			fileContent: `assets:
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
			fileName: "test_config.yaml",
			expected: &AssetConfig{
				Assets: map[AssetID]AssetEntry{
					"BTCUSD": {
						AssetID:                "BTCUSD",
						EncodedAssetID:         "0x7404e3d104ea7841c3d9e6fd20adfe99b4ad586bc08d8f3bd3afef894cf184de",
						PercentChangeThreshold: 1.5,
						FallbackPeriodSecs:     300,
					},
					"ETHUSD": {
						AssetID:                "ETHUSD",
						EncodedAssetID:         "0x59102b37de83bdda9f38ac8254e596f0d9ac61d2035c07936675e87342817160",
						PercentChangeThreshold: 2.0,
						FallbackPeriodSecs:     600,
					},
				},
			},
			wantError: false,
		},
		{
			name:        "empty config",
			fileContent: `assets: {}`,
			fileName:    "empty_config.yaml",
			expected: &AssetConfig{
				Assets: map[AssetID]AssetEntry{},
			},
			wantError: false,
		},
		{
			name: "single asset",
			fileContent: `assets:
  SOLUSD:
    asset_id: "SOLUSD"
    encoded_asset_id: "0x1dcd89dfded9e8a9b0fa1745a8ebbacbb7c81e33d5abc81616633206d932e837"
    percent_change_threshold: 0.5
    fallback_period_sec: 120`,
			fileName: "single_config.yaml",
			expected: &AssetConfig{
				Assets: map[AssetID]AssetEntry{
					"SOLUSD": {
						AssetID:                "SOLUSD",
						EncodedAssetID:         "0x1dcd89dfded9e8a9b0fa1745a8ebbacbb7c81e33d5abc81616633206d932e837",
						PercentChangeThreshold: 0.5,
						FallbackPeriodSecs:     120,
					},
				},
			},
			wantError: false,
		},
		{
			name:        "invalid yaml",
			fileContent: `assets:\n  invalid: yaml: content: [`,
			fileName:    "invalid_config.yaml",
			expected:    nil,
			wantError:   true,
		},
		{
			name:        "malformed structure",
			fileContent: `not_assets: {}`,
			fileName:    "malformed_config.yaml",
			expected: &AssetConfig{
				Assets: nil,
			},
			wantError: false,
		},
		{
			name: "zero values",
			fileContent: `assets:
  TEST:
    asset_id: "TEST"
    encoded_asset_id: "0x0000000000000000000000000000000000000000000000000000000000000000"
    percent_change_threshold: 0
    fallback_period_sec: 0`,
			fileName: "zero_values_config.yaml",
			expected: &AssetConfig{
				Assets: map[AssetID]AssetEntry{
					"TEST": {
						AssetID:                "TEST",
						EncodedAssetID:         "0x0000000000000000000000000000000000000000000000000000000000000000",
						PercentChangeThreshold: 0,
						FallbackPeriodSecs:     0,
					},
				},
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create temporary directory
			tempDir, err := os.MkdirTemp("", "config_test")
			require.NoError(t, err)
			defer os.RemoveAll(tempDir)

			// Write test file
			filePath := filepath.Join(tempDir, tt.fileName)
			err = os.WriteFile(filePath, []byte(tt.fileContent), 0644)
			require.NoError(t, err)

			// Test LoadConfig
			result, err := LoadConfig(filePath)

			if tt.wantError {
				assert.Error(t, err)
				assert.Equal(t, tt.expected, result)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestLoadConfig_FileNotFound(t *testing.T) {
	t.Parallel()

	// if this test is failing, ensure that this file doesn't exist.
	result, err := LoadConfig("nonexistent_file.yaml")

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestLoadConfig_EmptyFilename(t *testing.T) {
	t.Parallel()

	result, err := LoadConfig("")

	assert.Error(t, err)
	assert.Nil(t, result)
}
