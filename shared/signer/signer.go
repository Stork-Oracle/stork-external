package signer

/*
#include "signer_ffi.h"
*/
import "C"

import (
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"time"
	"unsafe"

	"github.com/NethermindEth/juno/core/felt"
	"github.com/Stork-Oracle/stork-external/shared"
	"github.com/Stork-Oracle/stork-external/shared/signer/evm"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
	"github.com/rs/zerolog"
)

var ErrBigIntTooLarge = errors.New("big int is larger than the modulus of the Stark curve")

// Stork auth constants. Magic number / asset id constants are re-exported from
// the evm subpackage to keep a single source of truth.
const (
	StorkMagicNumber             = evm.StorkMagicNumber
	StorkAuthAssetId             = evm.StorkAuthAssetID
	StarkEncodedStorkAuthAssetId = "0x53544f524b41555448000000000000007361757468"
	StorkAuthOracleId            = "sauth"
)

const (
	publicKeyHeader     = "X-Public-Key"
	timestampHeader     = "X-Timestamp"
	signatureHeader     = "X-Signature"
	signatureTypeHeader = "X-Signature-Type"
)

type Signer[T shared.Signature] interface {
	SignPublisherPrice(
		publishTimestamp int64,
		asset string,
		quantizedValue string,
	) (timestampedSig *shared.TimestampedSignature[T], encodedAssetId string, err error)
	GetPublisherKey() shared.PublisherKey
	GetSignatureType() shared.SignatureType
}

// Re-exports of the EVM signer types/constructors from the evm subpackage.
// External consumers that previously imported these symbols from this package
// continue to work unchanged.
type (
	EvmSigner     = evm.Signer
	EvmAuthSigner = evm.AuthSigner
)

var (
	NewEvmSigner     = evm.NewSigner
	NewEvmAuthSigner = evm.NewAuthSigner
)

type StarkSigner struct {
	pkBytes       []byte
	publicKey     string
	oracleNameInt *big.Int
	logger        zerolog.Logger
}

func NewStarkSigner(
	privateKeyStr StarkPrivateKey,
	publicKeyStr, oracleId string,
	logger zerolog.Logger,
) (*StarkSigner, error) {
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

func (s *StarkSigner) SignPublisherPrice(
	publishTimestamp int64,
	asset string,
	quantizedValue string,
) (timestampedSig *shared.TimestampedSignature[*shared.StarkSignature], encodedAssetId string, err error) {
	// Convert asset to hex string
	assetHex := hex.EncodeToString([]byte(asset))
	assetHexPadded := assetHex
	if len(assetHex) < 34 {
		assetHexPadded = assetHex + strings.Repeat("0", 32-len(assetHex))
	}

	assetInt, _ := new(big.Int).SetString(assetHexPadded, 16)
	priceInt, _ := new(big.Int).SetString(quantizedValue, 10)
	timestampInt := new(big.Int).SetInt64(publishTimestamp / 1_000_000_000)

	xInt := new(big.Int).Add(shiftLeft(assetInt, 40), s.oracleNameInt)
	yInt := new(big.Int).Add(shiftLeft(priceInt, 32), timestampInt)

	pedersonHashBuf := make([]byte, 32)
	sigRBuf := make([]byte, 32)
	sigSBuf := make([]byte, 32)

	xIntBuf, err := createStarkBufferFromBigIntAbs(xInt)
	if err != nil {
		return nil, "", err
	}
	yIntBuf, err := createStarkBufferFromBigIntAbs(yInt)
	if err != nil {
		return nil, "", err
	}
	hashAndSignStatus := C.hash_and_sign(
		xIntBuf,
		yIntBuf,
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

	starkSignature := &shared.StarkSignature{
		R: "0" + trimLeadingZeros(sigRFelt.String()),
		S: "0" + trimLeadingZeros(sigSFelt.String()),
	}
	msgHash := add0x(trimLeadingZeros(strip0x(pedersenHashFelt.String())))
	timestampedSignature := shared.TimestampedSignature[*shared.StarkSignature]{
		Signature:     starkSignature,
		TimestampNano: uint64(publishTimestamp),
		MsgHash:       msgHash,
	}
	externalAssetId := "0x" + assetHexPadded + s.oracleNameInt.Text(16)

	return &timestampedSignature, externalAssetId, nil
}

func (s *StarkSigner) GetPublisherKey() shared.PublisherKey {
	return shared.PublisherKey(s.publicKey)
}

func (s *StarkSigner) GetSignatureType() shared.SignatureType {
	return shared.StarkSignatureType
}

type StorkAuthSigner interface {
	SignAuth(publishTimestamp int64) (string, error)
	GetAuthHeaders() (http.Header, error)
}

type StarkAuthSigner struct {
	starkSigner *StarkSigner
}

func NewStarkAuthSigner(
	privateKeyStr StarkPrivateKey,
	publicKeyStr string,
	logger zerolog.Logger,
) (*StarkAuthSigner, error) {
	starkSigner, err := NewStarkSigner(privateKeyStr, publicKeyStr, StorkAuthOracleId, logger)
	if err != nil {
		return nil, err
	}
	return &StarkAuthSigner{starkSigner: starkSigner}, nil
}

func (s *StarkAuthSigner) SignAuth(publishTimestamp int64) (string, error) {
	timestampedSignature, _, err := s.starkSigner.SignPublisherPrice(
		publishTimestamp,
		StorkAuthAssetId,
		StorkMagicNumber,
	)
	if err != nil {
		return "", fmt.Errorf("failed to sign auth: %v", err)
	}
	signature := timestampedSignature.Signature
	rHex := fmt.Sprintf("%064s", strip0x(signature.R))
	sHex := fmt.Sprintf("%064s", strip0x(signature.S))
	return "0x" + rHex + sHex, nil
}

func (s *StarkAuthSigner) GetAuthHeaders() (http.Header, error) {
	publicKey := s.starkSigner.GetPublisherKey()
	signatureType := s.starkSigner.GetSignatureType()
	timestamp := time.Now().UnixNano()
	signatureString, err := s.SignAuth(timestamp)
	if err != nil {
		return nil, fmt.Errorf("failed to sign auth header: %v", err)
	}
	header := http.Header{}
	header.Set(publicKeyHeader, string(publicKey))
	header.Set(timestampHeader, fmt.Sprintf("%d", timestamp))
	header.Set(signatureHeader, signatureString)
	header.Set(signatureTypeHeader, string(signatureType))
	return header, nil
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

// createStarkBufferFromBigIntAbs creates a 32-byte buffer from a big.Int and
// returns a pointer to the buffer as a C uint8_t. The big.Int must be less
// than the modulus of the Stark curve (251-bit prime); the absolute value is
// used regardless of sign.
func createStarkBufferFromBigIntAbs(i *big.Int) (*C.uint8_t, error) {
	absI := new(big.Int).Abs(i)
	if absI.Cmp(fp.Modulus()) >= 0 {
		return nil, ErrBigIntTooLarge
	}
	bytes := make([]byte, 32)
	absI.FillBytes(bytes)
	return (*C.uint8_t)(unsafe.Pointer(&bytes[0])), nil
}
