package uniswapv2

import (
	"math/big"
	"testing"
)

func TestCalculatePrice(t *testing.T) {
	tests := []struct {
		name               string
		result             []interface{}
		baseTokenIndex     int8
		quoteTokenIndex    int8
		baseTokenDecimals  int8
		quoteTokenDecimals int8
		expectedPrice      float64
		expectedError      bool
	}{
		{
			result: []interface{}{
				big.NewInt(1000000),
				big.NewInt(2000000000000000000),
			},
			baseTokenIndex:     0,
			quoteTokenIndex:    1,
			baseTokenDecimals:  6,
			quoteTokenDecimals: 18,
			expectedPrice:      2.0,
			expectedError:      false,
		},
		{
			result: []interface{}{
				big.NewInt(1000000),
				big.NewInt(2000000000000000000),
			},
			baseTokenIndex:     1,
			quoteTokenIndex:    0,
			baseTokenDecimals:  18,
			quoteTokenDecimals: 6,
			expectedPrice:      0.5,
			expectedError:      false,
		},
		{
			result: []interface{}{
				big.NewInt(1000000),
				big.NewInt(2000000000000000000),
			},
			baseTokenIndex:     0,
			quoteTokenIndex:    1,
			baseTokenDecimals:  6,
			quoteTokenDecimals: 17,
			expectedPrice:      20.0,
			expectedError:      false,
		},
		{
			result: []interface{}{
				big.NewInt(0),
				big.NewInt(2000000000000000000),
			},
			baseTokenIndex:     0,
			quoteTokenIndex:    1,
			baseTokenDecimals:  6,
			quoteTokenDecimals: 18,
			expectedPrice:      -1,
			expectedError:      true,
		},
		{
			result: []interface{}{
				big.NewInt(0),
				big.NewInt(2000000000000000000),
			},
			baseTokenIndex:     1,
			quoteTokenIndex:    0,
			baseTokenDecimals:  6,
			quoteTokenDecimals: 18,
			expectedPrice:      -1,
			expectedError:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			price, err := calculatePrice(tt.result, tt.baseTokenIndex, tt.quoteTokenIndex, tt.baseTokenDecimals, tt.quoteTokenDecimals)

			if tt.expectedError && err == nil {
				t.Errorf("expected error but got none")
			}

			if !tt.expectedError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if price != tt.expectedPrice {
				t.Errorf("expected price %v, but got %v", tt.expectedPrice, price)
			}
		})
	}
}
