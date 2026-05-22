package evm

import (
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"github.com/Stork-Oracle/stork-external/shared"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

func getPublisherPricePayload(
	timestamp int64,
	quantizedPrice string,
	assetID string,
	publicAddress common.Address,
) [][]byte {
	timestampBigInt := big.NewInt(timestamp / 1_000_000_000)
	quantizedPriceBigInt := new(big.Int)
	quantizedPriceBigInt.SetString(quantizedPrice, 10)
	quantizedPriceBytes := bigIntToTwosComplement32(quantizedPriceBigInt)

	return [][]byte{
		publicAddress.Bytes(),
		[]byte(assetID),
		common.LeftPadBytes(timestampBigInt.Bytes(), 32),
		quantizedPriceBytes,
	}
}

func getHashes(payload [][]byte) (common.Hash, common.Hash) {
	payloadHash := crypto.Keccak256Hash(payload...)

	prefixedHash := crypto.Keccak256Hash(
		[]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%v", len(payloadHash))),
		payloadHash.Bytes(),
	)
	return payloadHash, prefixedHash
}

func signatureToBytes(sig shared.EvmSignature) ([]byte, error) {
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

func convertHexToECDSA(privateKey PrivateKey) (*ecdsa.PrivateKey, error) {
	privateKeyStr := strings.Replace(string(privateKey), "0x", "", 1)

	evmPrivateKey, err := crypto.HexToECDSA(privateKeyStr)
	if err != nil {
		return nil, err
	}

	return evmPrivateKey, nil
}

func signData(privateKey *ecdsa.PrivateKey, payload [][]byte) (string, []byte, error) {
	payloadHash, prefixedHash := getHashes(payload)
	signature, err := crypto.Sign(prefixedHash.Bytes(), privateKey)
	if err != nil {
		return "", nil, err
	}

	return payloadHash.String(), signature, nil
}

func bytesToRsvSignature(signature []byte) (*shared.EvmSignature, error) {
	r := hex.EncodeToString(signature[:32])
	s := hex.EncodeToString(signature[32:64])
	v := hex.EncodeToString([]byte{signature[64] + 27})

	return &shared.EvmSignature{R: "0x" + r, S: "0x" + s, V: "0x" + v}, nil
}

// bigIntToTwosComplement32 converts a big.Int to a 32-byte two's complement
// representation. Matches Ethereum's signed integer encoding where negative
// numbers are represented in two's complement.
func bigIntToTwosComplement32(x *big.Int) []byte {
	if x.Sign() >= 0 {
		return common.LeftPadBytes(x.Bytes(), 32)
	}

	absX := new(big.Int).Abs(x)

	mask := new(big.Int).Lsh(big.NewInt(1), 256)
	mask.Sub(mask, big.NewInt(1))

	inverted := new(big.Int).Sub(mask, absX)
	twosComp := new(big.Int).Add(inverted, big.NewInt(1))

	return twosComp.Bytes()
}

func strip0x(str string) string {
	return strings.TrimPrefix(str, "0x")
}
