package stork_publisher_agent

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"log"
	"math/big"
	"strings"
)

type Signer[T Signature] struct {
	config           StorkPublisherAgentConfig
	evmPrivateKey    *ecdsa.PrivateKey
	evmPublicAddress common.Address
}

func NewSigner[T Signature](config StorkPublisherAgentConfig) (*Signer[T], error) {
	switch config.SignatureType {
	case EvmSignatureType:
		evmPrivateKey, err := convertHexToECDSA(config.PrivateKey)
		if err != nil {
			return nil, fmt.Errorf("error converting private key to ECDSA: %v", err)
		}
		publicAddress := crypto.PubkeyToAddress(evmPrivateKey.PublicKey)
		signer := Signer[T]{
			config:           config,
			evmPrivateKey:    evmPrivateKey,
			evmPublicAddress: publicAddress,
		}
		return &signer, nil
	default:
		return nil, fmt.Errorf("unknown signature type: %v", config.SignatureType)
	}
}

func convertHexToECDSA(privateKey PrivateKey) (*ecdsa.PrivateKey, error) {
	privateKeyStr := string(privateKey)
	if strings.HasPrefix(privateKeyStr, "0x") {
		privateKeyStr = strings.Replace(privateKeyStr, "0x", "", 1)
	}
	privateKeyBytes, err := hex.DecodeString(privateKeyStr)
	if err != nil {
		return nil, err
	}

	// Create a new ecdsa.PrivateKey object
	evmPrivateKey := new(ecdsa.PrivateKey)
	evmPrivateKey.D = new(big.Int).SetBytes(privateKeyBytes)
	evmPrivateKey.PublicKey.Curve = elliptic.P256()
	evmPrivateKey.PublicKey.X, evmPrivateKey.PublicKey.Y = evmPrivateKey.PublicKey.Curve.ScalarBaseMult(privateKeyBytes)

	return evmPrivateKey, nil
}

func (s *Signer[T]) GetSignedPriceUpdate(priceUpdate PriceUpdate, triggerType TriggerType) SignedPriceUpdate[T] {
	quantizedPrice := FloatToQuantizedPrice(priceUpdate.Price)
	var timestampedSignature TimestampedSignature[T]
	var publisherKey PublisherKey
	switch s.config.SignatureType {
	case EvmSignatureType:
		timestampedSignatureRef, err := s.SignEvm(priceUpdate.Asset, quantizedPrice, priceUpdate.PublishTimestamp)
		if err != nil {
			panic(err)
		}
		timestampedSignature = *timestampedSignatureRef
		publisherKey = PublisherKey(s.evmPublicAddress.Hex())
	case StarkSignatureType:
		// todo: implement
		panic("stark signing not yet implemented")
	default:
		log.Fatalf("unknown signature type: %v", s.config.SignatureType)
	}

	signedPriceUpdate := SignedPriceUpdate[T]{
		OracleId: s.config.OracleId,
		AssetId:  priceUpdate.Asset,
		Trigger:  triggerType,
		SignedPrice: SignedPrice[T]{
			PublisherKey:         publisherKey,
			ExternalAssetId:      string(priceUpdate.Asset),
			SignatureType:        s.config.SignatureType,
			QuantizedPrice:       quantizedPrice,
			TimestampedSignature: timestampedSignature,
		},
	}

	return signedPriceUpdate
}

func (s *Signer[T]) SignEvm(assetId AssetId, quantizedPrice QuantizedPrice, timestampNs int64) (*TimestampedSignature[T], error) {
	timestampBigInt := big.NewInt(timestampNs / 1_000_000_000)

	quantizedPriceBigInt := new(big.Int)
	quantizedPriceBigInt.SetString(string(quantizedPrice), 10)
	hashedAsset := crypto.Keccak256([]byte(assetId)) // strings in golang are utf-8 encoded by default

	publicAddress := crypto.PubkeyToAddress(s.evmPrivateKey.PublicKey)
	payload := [][]byte{
		publicAddress.Bytes(),
		hashedAsset,
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
	case *StarkSignature:
		result = &StarkSignature{R: "0x" + r, S: "0x" + s}
	default:
		return rsv, errors.New(fmt.Sprintf("invalid type for T: %T", zeroValue))
	}

	return result.(T), nil
}
