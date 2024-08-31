package stork_publisher_agent

/*
#cgo LDFLAGS: -L/app/rust/stork/target/aarch64-unknown-linux-gnu/release -L../../rust/stork/target/release -lstork
#cgo CFLAGS: -I/app/rust/stork/src -I../../rust/stork/src
#include "signing.h"
*/
import "C"
import (
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"math/big"
	"strings"
	"unsafe"

	"github.com/NethermindEth/juno/core/felt"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/rs/zerolog"
)

type Signer[T Signature] struct {
	config             StorkPublisherAgentConfig
	signatureType      SignatureType
	evmPrivateKey      *ecdsa.PrivateKey
	evmPublicAddress   common.Address
	starkOracleNameInt *big.Int
	starkPkBytes       []byte
	logger             zerolog.Logger
}

func NewSigner[T Signature](config StorkPublisherAgentConfig, signatureType SignatureType, logger zerolog.Logger) (*Signer[T], error) {
	switch signatureType {
	case EvmSignatureType:
		privateKey, err := convertHexToECDSA(config.EvmPrivateKey)
		if err != nil {
			return nil, fmt.Errorf("error converting private key to ECDSA: %v", err)
		}
		publicAddress := crypto.PubkeyToAddress(privateKey.PublicKey)
		signer := Signer[T]{
			config:           config,
			signatureType:    signatureType,
			evmPrivateKey:    privateKey,
			evmPublicAddress: publicAddress,
			logger:           logger,
		}
		return &signer, nil
	case StarkSignatureType:
		oracleNameHex := hex.EncodeToString([]byte(config.OracleId))
		oracleNameInt, _ := new(big.Int).SetString(oracleNameHex, 16)

		pkTrimmed := strings.TrimPrefix(string(config.StarkPrivateKey), "0x")
		if len(pkTrimmed)%2 != 0 {
			pkTrimmed = "0" + pkTrimmed
		}
		pkDecoded, err := hex.DecodeString(pkTrimmed)
		if err != nil {
			return nil, fmt.Errorf("error decoding private key: %v", err)
		}
		pkBytes := make([]byte, 32)
		copy(pkBytes[32-len(pkDecoded):], pkDecoded)

		signer := Signer[T]{
			config:             config,
			signatureType:      signatureType,
			starkPkBytes:       pkBytes,
			starkOracleNameInt: oracleNameInt,
			logger:             logger,
		}
		return &signer, nil
	default:
		return nil, fmt.Errorf("unknown signature type: %v", signatureType)
	}
}

func (s *Signer[T]) GetSignedPriceUpdate(priceUpdate PriceUpdate, triggerType TriggerType) SignedPriceUpdate[T] {
	quantizedPrice := FloatToQuantizedPrice(&priceUpdate.Value)
	var timestampedSignature TimestampedSignature[T]
	var publisherKey PublisherKey
	var externalAssetId string
	switch s.signatureType {
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
		publisherKey = PublisherKey(s.config.StarkPublicKey)
		externalAssetId = paddedAssetHex + s.starkOracleNameInt.Text(16)
	default:
		log.Fatalf("unknown signature type: %v", s.signatureType)
	}

	signedPriceUpdate := SignedPriceUpdate[T]{
		OracleId: s.config.OracleId,
		AssetId:  priceUpdate.Asset,
		Trigger:  triggerType,
		SignedPrice: SignedPrice[T]{
			PublisherKey:         publisherKey,
			ExternalAssetId:      externalAssetId,
			SignatureType:        s.signatureType,
			QuantizedPrice:       quantizedPrice,
			TimestampedSignature: timestampedSignature,
		},
	}

	return signedPriceUpdate
}

func convertHexToECDSA(privateKey EvmPrivateKey) (*ecdsa.PrivateKey, error) {
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

func createBufferFromBytes(buf []byte) *C.uint8_t {
	return (*C.uint8_t)(unsafe.Pointer(&buf[0]))
}

func createBufferFromBigInt(i *big.Int) *C.uint8_t {
	bytes := make([]byte, 32)
	i.FillBytes(bytes)
	return (*C.uint8_t)(unsafe.Pointer(&bytes[0]))
}

func (s *Signer[T]) SignStark(assetHexPadded string, quantizedPrice QuantizedPrice, timestampNs int64) (*TimestampedSignature[T], error) {
	assetInt, _ := new(big.Int).SetString(strip0x(assetHexPadded), 16)

	priceInt, _ := new(big.Int).SetString(string(quantizedPrice), 10)
	timestampInt := new(big.Int).SetInt64(timestampNs / 1_000_000_000)

	xInt := new(big.Int).Add(shiftLeft(assetInt, 40), s.starkOracleNameInt)
	yInt := new(big.Int).Add(shiftLeft(priceInt, 32), timestampInt)

	pedersonHashBuf := make([]byte, 32)
	sigRBuf := make([]byte, 32)
	sigSBuf := make([]byte, 32)

	hashAndSignStatus := C.hash_and_sign(
		createBufferFromBigInt(xInt),
		createBufferFromBigInt(yInt),
		createBufferFromBytes(s.starkPkBytes),
		createBufferFromBytes(pedersonHashBuf),
		createBufferFromBytes(sigRBuf),
		createBufferFromBytes(sigSBuf),
	)
	if hashAndSignStatus != 0 {
		return nil, errors.New(fmt.Sprintf("failed to hash and sign - response code %v", hashAndSignStatus))
	}

	pedersenHashFelt, err := bytesToFieldElement(pedersonHashBuf)
	if err != nil {
		return nil, err
	}
	sigRFelt, err := bytesToFieldElement(sigRBuf)
	if err != nil {
		return nil, err
	}
	sigSFelt, err := bytesToFieldElement(sigSBuf)
	if err != nil {
		return nil, err
	}

	var starkSignature any
	starkSignature = &StarkSignature{
		R: "0" + trimLeadingZeros(sigRFelt.String()),
		S: "0" + trimLeadingZeros(sigSFelt.String()),
	}
	signature := starkSignature.(T)
	// trim leading 0s
	msgHash := add0x(trimLeadingZeros(strip0x(pedersenHashFelt.String())))
	timestampedSignature := TimestampedSignature[T]{
		Signature: signature,
		Timestamp: timestampNs,
		MsgHash:   msgHash,
	}
	return &timestampedSignature, nil
}
