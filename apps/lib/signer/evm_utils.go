package signer

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

func getPublisherEvmPricePayload(
	timestamp int64,
	quantizedPrice string,
	assetId string,
	publicAddress common.Address,
) [][]byte {
	timestampBigInt := big.NewInt(timestamp / 1_000_000_000)
	quantizedPriceBigInt := new(big.Int)
	quantizedPriceBigInt.SetString(quantizedPrice, 10)

	return [][]byte{
		publicAddress.Bytes(),
		[]byte(assetId),
		common.LeftPadBytes(timestampBigInt.Bytes(), 32),
		common.LeftPadBytes(quantizedPriceBigInt.Bytes(), 32),
	}
}

func getEvmHashes(payload [][]byte) (common.Hash, common.Hash) {
	payloadHash := crypto.Keccak256Hash(payload...)

	prefixedHash := crypto.Keccak256Hash(
		[]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%v", len(payloadHash))),
		payloadHash.Bytes(),
	)
	return payloadHash, prefixedHash
}

func evmSignatureToBytes(sig EvmSignature) ([]byte, error) {
	cleanedR, _ := strings.CutPrefix(sig.R, "0x")
	rBytes, err := hex.DecodeString(cleanedR)
	if err != nil {
		return nil, err
	}

	cleanedS, _ := strings.CutPrefix(sig.S, "0x")
	sBytes, err := hex.DecodeString(cleanedS)
	if err != nil {
		return nil, err
	}

	cleanedV, _ := strings.CutPrefix(sig.V, "0x")
	vBytes, err := hex.DecodeString(cleanedV)
	if err != nil {
		return nil, err
	}
	vBytes[0] = vBytes[0] - 27

	combinedBytes := append(append(rBytes, sBytes...), vBytes...)
	return combinedBytes, nil
}
