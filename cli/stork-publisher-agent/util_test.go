package stork_publisher_agent

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFloatToQuantizedPrice(t *testing.T) {
	bigFloat, _, _ := big.ParseFloat("72147.681412670819", 10, 64, big.ToZero)
	quantizedPrice := FloatToQuantizedPrice(bigFloat)
	expectedQuantizedPrice := QuantizedPrice("72147681412670819000000")
	assert.Equal(t, expectedQuantizedPrice, quantizedPrice)

	bigFloat, _, _ = big.ParseFloat("3.33595034988", 10, 64, big.ToZero)
	quantizedPrice = FloatToQuantizedPrice(bigFloat)
	expectedQuantizedPrice = "3335950349880000000"
	assert.Equal(t, expectedQuantizedPrice, quantizedPrice)
}
