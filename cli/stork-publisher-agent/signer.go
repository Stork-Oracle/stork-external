package stork_publisher_agent

import (
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	junocrypto "github.com/NethermindEth/juno/core/crypto"
	"github.com/NethermindEth/juno/core/felt"
	"github.com/NethermindEth/starknet.go/curve"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"log"
	"math/big"
	"strings"
)

type Signer[T Signature] struct {
	config              StorkPublisherAgentConfig
	evmPrivateKey       *ecdsa.PrivateKey
	evmPublicAddress    common.Address
	starkOracleNameInt  *big.Int
	starkPrivateKeyFelt *felt.Felt
}

func NewSigner[T Signature](config StorkPublisherAgentConfig) (*Signer[T], error) {
	privateKey, err := convertHexToECDSA(config.PrivateKey)

	switch config.SignatureType {
	case EvmSignatureType:
		if err != nil {
			return nil, fmt.Errorf("error converting private key to ECDSA: %v", err)
		}
		publicAddress := crypto.PubkeyToAddress(privateKey.PublicKey)
		signer := Signer[T]{
			config:           config,
			evmPrivateKey:    privateKey,
			evmPublicAddress: publicAddress,
		}
		return &signer, nil
	case StarkSignatureType:
		oracleNameHex := hex.EncodeToString([]byte(config.OracleId))
		oracleNameInt, _ := new(big.Int).SetString(oracleNameHex, 16)

		pkTrimmed := strings.TrimPrefix(string(config.PrivateKey), "0x")
		if len(pkTrimmed)%2 != 0 {
			pkTrimmed = "0" + pkTrimmed
		}
		pkBytes, err := hex.DecodeString(pkTrimmed)
		if err != nil {
			return nil, fmt.Errorf("error decoding private key: %v", err)
		}
		pkFelt, err := bytesToFieldElement(pkBytes)
		if err != nil {
			return nil, fmt.Errorf("error converting pk to a field element: %v", err)
		}

		signer := Signer[T]{
			config:              config,
			starkPrivateKeyFelt: pkFelt,
			starkOracleNameInt:  oracleNameInt,
		}
		return &signer, nil
	default:
		return nil, fmt.Errorf("unknown signature type: %v", config.SignatureType)
	}
}

func (s *Signer[T]) GetSignedPriceUpdate(priceUpdate PriceUpdate, triggerType TriggerType) SignedPriceUpdate[T] {
	quantizedPrice := FloatToQuantizedPrice(priceUpdate.Price)
	var timestampedSignature TimestampedSignature[T]
	var publisherKey PublisherKey
	var externalAssetId string
	switch s.config.SignatureType {
	case EvmSignatureType:
		timestampedSignatureRef, err := s.SignEvm(priceUpdate.Asset, quantizedPrice, priceUpdate.PublishTimestamp)
		if err != nil {
			panic(err)
		}
		timestampedSignature = *timestampedSignatureRef
		publisherKey = PublisherKey(s.evmPublicAddress.Hex())
		externalAssetId = string(priceUpdate.Asset)
	case StarkSignatureType:
		// Convert asset to hex string
		assetHex := hex.EncodeToString([]byte(priceUpdate.Asset))
		paddedAssetHex := assetHex
		if len(assetHex) < 34 {
			paddedAssetHex = "0x" + assetHex + strings.Repeat("0", 32-len(assetHex))
		}

		timestampedSignatureRef, err := s.SignStark(paddedAssetHex, quantizedPrice, priceUpdate.PublishTimestamp)
		if err != nil {
			panic(err)
		}
		timestampedSignature = *timestampedSignatureRef
		publisherKey = s.config.PublicKey
		externalAssetId = paddedAssetHex + s.starkOracleNameInt.Text(16)
	default:
		log.Fatalf("unknown signature type: %v", s.config.SignatureType)
	}

	signedPriceUpdate := SignedPriceUpdate[T]{
		OracleId: s.config.OracleId,
		AssetId:  priceUpdate.Asset,
		Trigger:  triggerType,
		SignedPrice: SignedPrice[T]{
			PublisherKey:         publisherKey,
			ExternalAssetId:      externalAssetId,
			SignatureType:        s.config.SignatureType,
			QuantizedPrice:       quantizedPrice,
			TimestampedSignature: timestampedSignature,
		},
	}

	return signedPriceUpdate
}

func convertHexToECDSA(privateKey PrivateKey) (*ecdsa.PrivateKey, error) {
	privateKeyStr := strings.Replace(string(privateKey), "0x", "", 1)

	// Create a new ecdsa.PrivateKey object
	evmPrivateKey, err := crypto.HexToECDSA(privateKeyStr)
	if err != nil {
		return nil, err
	}

	return evmPrivateKey, nil
}

// getHashes returns the keccak hash of the payload and the keccak hash of the prefixed data hash
func getHashes(payload [][]byte) (common.Hash, common.Hash) {
	payloadHash := crypto.Keccak256Hash(payload...)

	prefixedHash := crypto.Keccak256Hash(
		[]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%v", len(payloadHash))),
		payloadHash.Bytes(),
	)
	return payloadHash, prefixedHash
}

func signData(privateKey *ecdsa.PrivateKey, payload [][]byte) (string, []byte, error) {
	payloadHash, prefixedHash := getHashes(payload)
	signature, err := crypto.Sign(prefixedHash.Bytes(), privateKey)

	if err != nil {
		return "", nil, err
	}

	return payloadHash.String(), signature, nil
}

func bytesToSignature[T Signature](signature []byte) (rsv T, err error) {
	r := hex.EncodeToString(signature[:32])
	s := hex.EncodeToString(signature[32:64])
	v := hex.EncodeToString([]byte{signature[64] + 27})

	var result any

	var zeroValue T

	switch (any)(zeroValue).(type) {
	case *EvmSignature:
		result = &EvmSignature{R: "0x" + r, S: "0x" + s, V: "0x" + v}
	default:
		return rsv, errors.New(fmt.Sprintf("invalid type for T: %T", zeroValue))
	}

	return result.(T), nil
}

func (s *Signer[T]) SignEvm(assetId AssetId, quantizedPrice QuantizedPrice, timestampNs int64) (*TimestampedSignature[T], error) {
	timestampBigInt := big.NewInt(timestampNs / 1_000_000_000)

	quantizedPriceBigInt := new(big.Int)
	quantizedPriceBigInt.SetString(string(quantizedPrice), 10)

	publicAddress := crypto.PubkeyToAddress(s.evmPrivateKey.PublicKey)
	payload := [][]byte{
		publicAddress.Bytes(),
		[]byte(assetId),
		common.LeftPadBytes(timestampBigInt.Bytes(), 32),
		common.LeftPadBytes(quantizedPriceBigInt.Bytes(), 32),
	}

	msgHash, signature, err := signData(s.evmPrivateKey, payload)
	if err != nil {
		return nil, err
	}

	rsv, err := bytesToSignature[T](signature)
	if err != nil {
		return nil, err
	}

	timestampedSignature := TimestampedSignature[T]{
		Signature: rsv,
		Timestamp: timestampNs,
		MsgHash:   msgHash,
	}
	return &timestampedSignature, nil
}

func bytesToFieldElement(b []byte) (*felt.Felt, error) {
	element := new(fp.Element).SetBytes(b)
	return felt.NewFelt(element), nil
}

func shiftLeft(num *big.Int, shift int) *big.Int {
	return new(big.Int).Lsh(num, uint(shift))
}

func trimLeadingZeros(str string) string {
	return strings.TrimLeft(str, "0")
}

func strip0x(str string) string {
	return strings.TrimPrefix(str, "0x")
}

func add0x(str string) string {
	return "0x" + str
}

func (s *Signer[T]) SignStark(assetHexPadded string, quantizedPrice QuantizedPrice, timestampNs int64) (*TimestampedSignature[T], error) {

	assetInt, _ := new(big.Int).SetString(strip0x(assetHexPadded), 16)

	priceInt, _ := new(big.Int).SetString(string(quantizedPrice), 10)
	timestampInt := new(big.Int).SetInt64(timestampNs / 1_000_000_000)

	xInt := new(big.Int).Add(shiftLeft(assetInt, 40), s.starkOracleNameInt)
	yInt := new(big.Int).Add(shiftLeft(priceInt, 32), timestampInt)

	xFE, err := bytesToFieldElement(xInt.Bytes())
	if err != nil {
		return nil, fmt.Errorf("error converting x to a field element: %v", err)
	}

	yFE, err := bytesToFieldElement(yInt.Bytes())
	if err != nil {
		return nil, fmt.Errorf("error converting y to a field element: %v", err)
	}

	pedersenHash := junocrypto.Pedersen(xFE, yFE)

	xFelt, yFelt, err := curve.Curve.SignFelt(pedersenHash, s.starkPrivateKeyFelt)
	if err != nil {
		return nil, fmt.Errorf("error signing the pedersen hash: %v", err)
	}

	var starkSignature any
	starkSignature = &StarkSignature{
		R: add0x(trimLeadingZeros(xFelt.Text(16))),
		S: add0x(trimLeadingZeros(yFelt.Text(16))),
	}
	signature := starkSignature.(T)
	// trim leading 0s
	msgHash := add0x(trimLeadingZeros(strip0x(pedersenHash.String())))
	timestampedSignature := TimestampedSignature[T]{
		Signature: signature,
		Timestamp: timestampNs,
		MsgHash:   msgHash,
	}
	return &timestampedSignature, nil
}
