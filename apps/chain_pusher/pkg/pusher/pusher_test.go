package pusher

import (
	"math/big"
	"testing"

	"github.com/Stork-Oracle/stork-external/apps/chain_pusher/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestShouldUpdateAsset(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name               string
		currentTimestampNs uint64
		currentValueStr    string
		storkTimestampNano int64
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
