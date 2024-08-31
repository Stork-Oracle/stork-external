package stork_publisher_agent

import (
	"math/big"
)

func FloatToQuantizedPrice(f *big.Float) QuantizedPrice {
	multiplier := new(big.Float).SetInt64(1e18)
	f.Mul(f, multiplier)
	result := new(big.Int)
	f.Int(result)
	return StringifyQuantizedPrice(result)
}

func StringifyQuantizedPrice(price *big.Int) QuantizedPrice {
	// Convert the big.Int to a string
	valStr := price.String()

	if len(valStr) > 6 {
		// zero out last 6 digits
		valStr = valStr[:len(valStr)-6] + "000000"
	}

	return QuantizedPrice(valStr)
}
