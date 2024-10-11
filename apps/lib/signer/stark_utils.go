package signer

import (
	"math/big"
	"strings"
)

func getPublisherPriceStarkXY(publishTimestamp int64, asset string, quantizedValue string) (xInt *big.Int, yInt *big.Int) {
	trimmedExternalAssetId, _ := strings.CutPrefix(asset, "0x")
	assetHexStr := trimmedExternalAssetId[:32]
	starkOracleNameStr := trimmedExternalAssetId[32:]
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
