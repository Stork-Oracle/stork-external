package evm

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBigIntBytesToTwosComplement(t *testing.T) {
	// negative
	intString := "-17725899000000"
	intBigInt := new(big.Int)
	intBigInt.SetString(intString, 10)

	twosComplement := bigIntToTwosComplement32(intBigInt)
	assert.Equal(
		t,
		"ffffffffffffffffffffffffffffffffffffffffffffffffffffefe0de163740",
		hex.EncodeToString(twosComplement),
	)

	// positive
	intString = "12500000000000"
	intBigInt.SetString(intString, 10)

	twosComplement = bigIntToTwosComplement32(intBigInt)
	assert.Equal(
		t,
		"00000000000000000000000000000000000000000000000000000b5e620f4800",
		hex.EncodeToString(twosComplement),
	)
}
