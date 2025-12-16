package signer

/*
#include "signer_ffi.h"
*/
import "C"

import (
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/Stork-Oracle/stork-external/shared"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

func VerifyAuth(
	timestampNano int64,
	publicKey shared.PublisherKey,
	signatureType shared.SignatureType,
	signature string,
) error {
	strippedSignature := strip0x(signature)

	switch signatureType {
	case shared.EvmSignatureType:
		if len(strippedSignature) != 130 {
			return errors.New("invalid EVM signature length")
		}
		r := "0x" + strippedSignature[:64]
		s := "0x" + strippedSignature[64:128]
		v := "0x" + strippedSignature[128:]
		evmSignature := shared.EvmSignature{
			R: r,
			S: s,
			V: v,
		}
		err := VerifyEvmPublisherPrice(timestampNano, StorkAuthAssetId, StorkMagicNumber, publicKey, evmSignature)
		if err != nil {
			return fmt.Errorf("invalid evm auth signature: %w", err)
		}
		return nil
	case shared.StarkSignatureType:
		if len(strippedSignature) != 128 {
			return errors.New("invalid Stark signature length")
		}
		r := "0x" + strippedSignature[:64]
		s := "0x" + strippedSignature[64:128]
		starkSignature := shared.StarkSignature{
			R: r,
			S: s,
		}
		err := VerifyStarkPublisherPrice(
			timestampNano,
			StarkEncodedStorkAuthAssetId,
			StorkMagicNumber,
			publicKey,
			starkSignature,
		)
		if err != nil {
			return fmt.Errorf("invalid stark auth signature: %w", err)
		}
		return nil
	default:
		return fmt.Errorf("invalid signature type: %s", signatureType)
	}
}

func VerifyPublisherPrice(
	publishTimestampNano int64,
	externalAssetId string,
	quantizedValue string,
	publisherKey shared.PublisherKey,
	signatureType shared.SignatureType,
	signature any,
) error {
	switch signatureType {
	case shared.EvmSignatureType:
		return VerifyEvmPublisherPrice(publishTimestampNano, externalAssetId, quantizedValue, publisherKey, signature)
	case shared.StarkSignatureType:
		return VerifyStarkPublisherPrice(publishTimestampNano, externalAssetId, quantizedValue, publisherKey, signature)
	default:
		return fmt.Errorf("invalid signature type: %s", signatureType)
	}
}

func VerifyEvmPublisherPrice(
	publishTimestampNano int64,
	externalAssetId string,
	quantizedValue string,
	publisherKey shared.PublisherKey,
	signature any,
) error {
	evmSignature := signature.(shared.EvmSignature)
	publisherAddress := common.HexToAddress(string(publisherKey))
	payload := getPublisherEvmPricePayload(
		publishTimestampNano,
		quantizedValue,
		externalAssetId,
		publisherAddress,
	)
	isValid, err := VerifyEvmSignature(publisherAddress, payload, evmSignature)
	if err != nil {
		return fmt.Errorf("failed to verify signature: %w", err)
	}
	if !isValid {
		return fmt.Errorf("invalid publisher signature")
	}
	return nil
}

func VerifyEvmSignature(
	publisherAddress common.Address,
	payload [][]byte,
	signature shared.EvmSignature,
) (bool, error) {
	_, prefixedHash := getEvmHashes(payload)
	storkSignatureBytes, err := evmSignatureToBytes(signature)
	if err != nil {
		return false, fmt.Errorf("failed to convert signature to bytes: %v", err)
	}

	// get the public key that generated this signature and convert it to a public address
	foundPubKey, err := crypto.Ecrecover(prefixedHash.Bytes(), storkSignatureBytes)
	if err != nil {
		return false, fmt.Errorf("failed to recover publisher signature: %v", err)
	}
	pubKey, err := crypto.UnmarshalPubkey(foundPubKey)
	if err != nil {
		return false, fmt.Errorf("error unmarshalling public key: %v", err)
	}
	address := crypto.PubkeyToAddress(*pubKey)

	return address == publisherAddress, nil
}

func VerifyStarkPublisherPrice(
	publishTimestampNano int64,
	externalAssetId string,
	quantizedValue string,
	publisherKey shared.PublisherKey,
	signature any,
) error {
	xInt, yInt := getPublisherPriceStarkXY(publishTimestampNano, externalAssetId, quantizedValue)
	isValid := verifyStarkSignature(xInt, yInt, publisherKey, signature)
	if !isValid {
		return errors.New("invalid stark signature")
	} else {
		return nil
	}
}

func verifyStarkSignature(xInt *big.Int, yInt *big.Int, publicKey shared.PublisherKey, signature any) bool {
	starkSignature := signature.(shared.StarkSignature)
	publicKeyStr, _ := strings.CutPrefix(string(publicKey), "0x")
	pubKeyInt := new(big.Int)
	pubKeyInt.SetString(publicKeyStr, 16)

	rStr, _ := strings.CutPrefix(starkSignature.R, "0x")
	rInt := new(big.Int)
	rInt.SetString(rStr, 16)

	sStr, _ := strings.CutPrefix(starkSignature.S, "0x")
	sInt := new(big.Int)
	sInt.SetString(sStr, 16)

	xIntBuf, err := createStarkBufferFromBigIntAbs(xInt)
	if err != nil {
		return false
	}
	yIntBuf, err := createStarkBufferFromBigIntAbs(yInt)
	if err != nil {
		return false
	}
	pubKeyIntBuf, err := createStarkBufferFromBigIntAbs(pubKeyInt)
	if err != nil {
		return false
	}
	rIntBuf, err := createStarkBufferFromBigIntAbs(rInt)
	if err != nil {
		return false
	}
	sIntBuf, err := createStarkBufferFromBigIntAbs(sInt)
	if err != nil {
		return false
	}
	isValidInt := C.validate_stark_signature(
		xIntBuf,
		yIntBuf,
		pubKeyIntBuf,
		rIntBuf,
		sIntBuf,
	)
	isValid := isValidInt != 0

	return isValid
}
