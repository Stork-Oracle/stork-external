package evm

import (
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"net/http"
	"time"

	"github.com/Stork-Oracle/stork-external/shared"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/rs/zerolog"
)

// Stork auth constants used when signing/verifying auth headers.
// "gmorkworld" in ascii (the literal magic number; matches the StarkSigner
// expectations).
const (
	StorkMagicNumber = "103109111114107119111114108100"
	StorkAuthAssetID = "STORKAUTH"
)

const (
	publicKeyHeader     = "X-Public-Key"
	timestampHeader     = "X-Timestamp"
	signatureHeader     = "X-Signature"
	signatureTypeHeader = "X-Signature-Type"
)

type Signer struct {
	privateKey       *ecdsa.PrivateKey
	publicKeyAddress common.Address
	logger           zerolog.Logger
}

func NewSigner(privateKeyStr PrivateKey, logger zerolog.Logger) (*Signer, error) {
	evmPrivateKey, err := convertHexToECDSA(privateKeyStr)
	if err != nil {
		return nil, err
	}

	publicKeyAddress := crypto.PubkeyToAddress(evmPrivateKey.PublicKey)
	return &Signer{
		privateKey:       evmPrivateKey,
		publicKeyAddress: publicKeyAddress,
		logger:           logger,
	}, nil
}

func (s *Signer) SignPublisherPrice(
	publishTimestamp int64,
	asset string,
	quantizedValue string,
) (*shared.TimestampedSignature[*shared.EvmSignature], string, error) {
	timestampBigInt := big.NewInt(publishTimestamp / 1_000_000_000)

	quantizedPriceBigInt := new(big.Int)
	quantizedPriceBigInt.SetString(string(quantizedValue), 10)
	quantizedPriceBytes := bigIntToTwosComplement32(quantizedPriceBigInt)

	publicAddress := crypto.PubkeyToAddress(s.privateKey.PublicKey)
	payload := [][]byte{
		publicAddress.Bytes(),
		[]byte(asset),
		common.LeftPadBytes(timestampBigInt.Bytes(), 32),
		quantizedPriceBytes,
	}

	msgHash, signature, err := signData(s.privateKey, payload)
	if err != nil {
		return nil, "", err
	}

	rsv, err := bytesToRsvSignature(signature)
	if err != nil {
		return nil, "", err
	}

	timestampedSignature := shared.TimestampedSignature[*shared.EvmSignature]{
		Signature:     rsv,
		TimestampNano: uint64(publishTimestamp),
		MsgHash:       msgHash,
	}
	return &timestampedSignature, asset, nil
}

func (s *Signer) GetPublisherKey() shared.PublisherKey {
	return shared.PublisherKey(s.publicKeyAddress.Hex())
}

func (s *Signer) GetSignatureType() shared.SignatureType {
	return shared.EvmSignatureType
}

type AuthSigner struct {
	signer *Signer
}

func NewAuthSigner(privateKeyStr PrivateKey, logger zerolog.Logger) (*AuthSigner, error) {
	s, err := NewSigner(privateKeyStr, logger)
	if err != nil {
		return nil, err
	}
	return &AuthSigner{signer: s}, nil
}

func (s *AuthSigner) SignAuth(publishTimestamp int64) (string, error) {
	timestampedSignature, _, err := s.signer.SignPublisherPrice(publishTimestamp, StorkAuthAssetID, StorkMagicNumber)
	if err != nil {
		return "", fmt.Errorf("failed to sign auth: %v", err)
	}
	signature := timestampedSignature.Signature
	rHex := fmt.Sprintf("%064s", strip0x(signature.R))
	sHex := fmt.Sprintf("%064s", strip0x(signature.S))
	vHex := fmt.Sprintf("%02s", strip0x(signature.V))
	return "0x" + rHex + sHex + vHex, nil
}

func (s *AuthSigner) GetAuthHeaders() (http.Header, error) {
	publicKey := s.signer.GetPublisherKey()
	signatureType := s.signer.GetSignatureType()
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
