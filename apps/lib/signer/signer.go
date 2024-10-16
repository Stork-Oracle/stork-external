package signer

/*
#cgo LDFLAGS: -L/app/rust/stork/target/aarch64-unknown-linux-gnu/release -L./rust/stork/target/release -lstork
#cgo CFLAGS: -I/app/rust/stork/src -I./rust/stork/src
#include "signing.h"
*/
import "C"
import (
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"unsafe"

	"github.com/NethermindEth/juno/core/felt"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/rs/zerolog"
)

type Signer[T Signature] interface {
	SignPublisherPrice(publishTimestamp int64, asset string, quantizedValue string) (timestampedSig *TimestampedSignature[T], encodedAssetId string, err error)
	GetPublisherKey() PublisherKey
	GetSignatureType() SignatureType
}

type EvmSigner struct {
	privateKey       *ecdsa.PrivateKey
	publicKeyAddress common.Address
	logger           zerolog.Logger
}

type StarkSigner struct {
	pkBytes       []byte
	publicKey     string
	oracleNameInt *big.Int
	logger        zerolog.Logger
}

func NewEvmSigner(privateKeyStr EvmPrivateKey, logger zerolog.Logger) (*EvmSigner, error) {
	evmPrivateKey, err := convertHexToECDSA(privateKeyStr)

	if err != nil {
		return nil, err
	}

	publicKeyAddress := crypto.PubkeyToAddress(evmPrivateKey.PublicKey)
	return &EvmSigner{
		privateKey:       evmPrivateKey,
		publicKeyAddress: publicKeyAddress,
		logger:           logger,
	}, nil
}

func NewStarkSigner(privateKeyStr StarkPrivateKey, publicKeyStr, oracleId string, logger zerolog.Logger) (*StarkSigner, error) {
	oracleNameHex := hex.EncodeToString([]byte(oracleId))
	oracleNameInt, _ := new(big.Int).SetString(oracleNameHex, 16)

	pkTrimmed := strings.TrimPrefix(string(privateKeyStr), "0x")
	if len(pkTrimmed)%2 != 0 {
		pkTrimmed = "0" + pkTrimmed
	}
	pkDecoded, err := hex.DecodeString(pkTrimmed)
	if err != nil {
		return nil, fmt.Errorf("error decoding private key: %v", err)
	}
	pkBytes := make([]byte, 32)
	copy(pkBytes[32-len(pkDecoded):], pkDecoded)

	return &StarkSigner{
		pkBytes:       pkBytes,
		publicKey:     publicKeyStr,
		oracleNameInt: oracleNameInt,
		logger:        logger,
	}, nil
}

func (s *EvmSigner) SignPublisherPrice(publishTimestamp int64, asset string, quantizedValue string) (timestampedSig *TimestampedSignature[*EvmSignature], encodedAssetId string, err error) {
	timestampBigInt := big.NewInt(publishTimestamp / 1_000_000_000)

	quantizedPriceBigInt := new(big.Int)
	quantizedPriceBigInt.SetString(string(quantizedValue), 10)

	publicAddress := crypto.PubkeyToAddress(s.privateKey.PublicKey)
	payload := [][]byte{
		publicAddress.Bytes(),
		[]byte(asset),
		common.LeftPadBytes(timestampBigInt.Bytes(), 32),
		common.LeftPadBytes(quantizedPriceBigInt.Bytes(), 32),
	}

	msgHash, signature, err := signData(s.privateKey, payload)
	if err != nil {
		return nil, "", err
	}

	rsv, err := bytesToRsvSignature(signature)
	if err != nil {
		return nil, "", err
	}

	timestampedSignature := TimestampedSignature[*EvmSignature]{
		Signature: rsv,
		Timestamp: publishTimestamp,
		MsgHash:   msgHash,
	}
	return &timestampedSignature, asset, nil
}

func (s *EvmSigner) GetPublisherKey() PublisherKey {
	return PublisherKey(s.publicKeyAddress.Hex())
}

func (s *EvmSigner) GetSignatureType() SignatureType {
	return EvmSignatureType
}

func (s *StarkSigner) SignPublisherPrice(publishTimestamp int64, asset string, quantizedValue string) (timestampedSig *TimestampedSignature[*StarkSignature], encodedAssetId string, err error) {
	// Convert asset to hex string
	assetHex := hex.EncodeToString([]byte(asset))
	assetHexPadded := assetHex
	if len(assetHex) < 34 {
		assetHexPadded = "0x" + assetHex + strings.Repeat("0", 32-len(assetHex))
	}

	assetInt, _ := new(big.Int).SetString(strip0x(assetHexPadded), 16)
	priceInt, _ := new(big.Int).SetString(quantizedValue, 10)
	timestampInt := new(big.Int).SetInt64(publishTimestamp / 1_000_000_000)

	xInt := new(big.Int).Add(shiftLeft(assetInt, 40), s.oracleNameInt)
	yInt := new(big.Int).Add(shiftLeft(priceInt, 32), timestampInt)

	pedersonHashBuf := make([]byte, 32)
	sigRBuf := make([]byte, 32)
	sigSBuf := make([]byte, 32)

	hashAndSignStatus := C.hash_and_sign(
		createBufferFromBigInt(xInt),
		createBufferFromBigInt(yInt),
		createBufferFromBytes(s.pkBytes),
		createBufferFromBytes(pedersonHashBuf),
		createBufferFromBytes(sigRBuf),
		createBufferFromBytes(sigSBuf),
	)
	if hashAndSignStatus != 0 {
		return nil, "", errors.New(fmt.Sprintf("failed to hash and sign - response code %v", hashAndSignStatus))
	}

	pedersenHashFelt, err := bytesToFieldElement(pedersonHashBuf)
	if err != nil {
		return nil, "", err
	}
	sigRFelt, err := bytesToFieldElement(sigRBuf)
	if err != nil {
		return nil, "", err
	}
	sigSFelt, err := bytesToFieldElement(sigSBuf)
	if err != nil {
		return nil, "", err
	}

	starkSignature := &StarkSignature{
		R: "0" + trimLeadingZeros(sigRFelt.String()),
		S: "0" + trimLeadingZeros(sigSFelt.String()),
	}
	// trim leading 0s
	msgHash := add0x(trimLeadingZeros(strip0x(pedersenHashFelt.String())))
	timestampedSignature := TimestampedSignature[*StarkSignature]{
		Signature: starkSignature,
		Timestamp: publishTimestamp,
		MsgHash:   msgHash,
	}
	externalAssetId := assetHexPadded + s.oracleNameInt.Text(16)

	return &timestampedSignature, externalAssetId, nil
}

func (s *StarkSigner) GetPublisherKey() PublisherKey {
	return PublisherKey(s.publicKey)
}

func (s *StarkSigner) GetSignatureType() SignatureType {
	return StarkSignatureType
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

func bytesToRsvSignature(signature []byte) (rsv *EvmSignature, err error) {
	r := hex.EncodeToString(signature[:32])
	s := hex.EncodeToString(signature[32:64])
	v := hex.EncodeToString([]byte{signature[64] + 27})

	return &EvmSignature{R: "0x" + r, S: "0x" + s, V: "0x" + v}, nil
}

func bytesToFieldElement(b []byte) (*felt.Felt, error) {
	element := new(fp.Element).SetBytes(b)
	return felt.NewFelt(element), nil
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
