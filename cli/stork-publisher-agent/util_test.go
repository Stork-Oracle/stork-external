package stork_publisher_agent

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFloatToQuantizedPrice(t *testing.T) {
	quantizedPrice := FloatToQuantizedPrice(72147.681412670819)
	expectedQuantizedPrice := QuantizedPrice("72147681412670819000000")
	assert.Equal(t, expectedQuantizedPrice, quantizedPrice)

	quantizedPrice = FloatToQuantizedPrice(3.33595034988)
	expectedQuantizedPrice = "3335950349880000000"
	assert.Equal(t, expectedQuantizedPrice, quantizedPrice)

	// todo: this case fails
	//quantizedPrice = FloatToQuantizedPrice(3.335950349883)
	//expectedQuantizedPrice = "3335950349883000000"
	//assert.Equal(t, expectedQuantizedPrice, quantizedPrice)
}
