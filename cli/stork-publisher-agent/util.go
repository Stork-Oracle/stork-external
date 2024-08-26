package stork_publisher_agent

import (
	"fmt"
	"math/big"
)

func FloatToQuantizedPrice(f float64) QuantizedPrice {
	// convert to string first to avoid rounding error
	strValue := fmt.Sprintf("%.18f", f)
	bigFloatValue, _, _ := big.ParseFloat(strValue, 10, 0, big.ToZero)
	multiplier := new(big.Float).SetInt64(1e18)
	bigFloatValue.Mul(bigFloatValue, multiplier)
	result := new(big.Int)
	bigFloatValue.Int(result)
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
