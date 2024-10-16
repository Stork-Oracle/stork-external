package signer

import (
	"math/big"
	"strings"
)

const starkOracleNameHexLength = 10

func getPublisherPriceStarkXY(publishTimestamp int64, asset string, quantizedValue string) (xInt *big.Int, yInt *big.Int) {
	trimmedExternalAssetId, _ := strings.CutPrefix(asset, "0x")

	assetLength := len(trimmedExternalAssetId) - starkOracleNameHexLength
	assetHexStr := trimmedExternalAssetId[:assetLength]
	starkOracleNameStr := trimmedExternalAssetId[assetLength:]
	assetInt := new(big.Int)
	assetInt.SetString(assetHexStr, 16)
	starkOracleNameInt := new(big.Int)
	starkOracleNameInt.SetString(starkOracleNameStr, 16)
	priceInt, _ := new(big.Int).SetString(quantizedValue, 10)
	timestampInt := new(big.Int).SetInt64(publishTimestamp / 1_000_000_000)

	xInt = new(big.Int).Add(shiftLeft(assetInt, 40), starkOracleNameInt)
	yInt = new(big.Int).Add(shiftLeft(priceInt, 32), timestampInt)
	return xInt, yInt
}

func shiftLeft(num *big.Int, shift int) *big.Int {
	return new(big.Int).Lsh(num, uint(shift))
}
