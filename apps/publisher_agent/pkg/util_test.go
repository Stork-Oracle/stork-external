package publisher_agent

import (
	"math/big"
	"testing"

	"github.com/Stork-Oracle/stork-external/shared"
	"github.com/stretchr/testify/assert"
)

func TestFloatToQuantizedPrice(t *testing.T) {
	bigFloat, _ := new(big.Float).SetString("72147.681412670819")
	quantizedPrice := FloatToQuantizedPrice(bigFloat)
	expectedQuantizedPrice := shared.QuantizedPrice("72147681412670819000000")
	assert.Equal(t, expectedQuantizedPrice, quantizedPrice)

	bigFloat, _ = new(big.Float).SetString("3.33595034988")
	quantizedPrice = FloatToQuantizedPrice(bigFloat)
	expectedQuantizedPrice = "3335950349880000000"
	assert.Equal(t, expectedQuantizedPrice, quantizedPrice)
}
