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
	"github.com/Stork-Oracle/stork-external/shared/signer/evm"
)

// Re-exports of EVM verifier functions from the evm subpackage.
var (
	VerifyEvmPublisherPrice = evm.VerifyPublisherPrice
	VerifyEvmSignature      = evm.VerifySignature
)

func VerifyAuth(
	timestampNano int64,
	publicKey shared.PublisherKey,
	signatureType shared.SignatureType,
	signature string,
) error {
	switch signatureType {
	case shared.EvmSignatureType:
		return evm.VerifyAuth(timestampNano, publicKey, signature)
	case shared.StarkSignatureType:
		strippedSignature := strip0x(signature)
		if len(strippedSignature) != 128 {
			return errors.New("invalid Stark signature length")
		}
		r := "0x" + strippedSignature[:64]
		s := "0x" + strippedSignature[64:128]
		starkSignature := shared.StarkSignature{R: r, S: s}
		if err := VerifyStarkPublisherPrice(
			timestampNano,
			StarkEncodedStorkAuthAssetId,
			StorkMagicNumber,
			publicKey,
			starkSignature,
		); err != nil {
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
		evmSig, ok := signature.(shared.EvmSignature)
		if !ok {
			return errors.New("expected shared.EvmSignature for EVM signature type")
		}
		return evm.VerifyPublisherPrice(publishTimestampNano, externalAssetId, quantizedValue, publisherKey, evmSig)
	case shared.StarkSignatureType:
		starkSig, ok := signature.(shared.StarkSignature)
		if !ok {
			return errors.New("expected shared.StarkSignature for Stark signature type")
		}
		return VerifyStarkPublisherPrice(publishTimestampNano, externalAssetId, quantizedValue, publisherKey, starkSig)
	default:
		return fmt.Errorf("invalid signature type: %s", signatureType)
	}
}

func VerifyStarkPublisherPrice(
	publishTimestampNano int64,
	externalAssetId string,
	quantizedValue string,
	publisherKey shared.PublisherKey,
	signature shared.StarkSignature,
) error {
	xInt, yInt := getPublisherPriceStarkXY(publishTimestampNano, externalAssetId, quantizedValue)
	if !verifyStarkSignature(xInt, yInt, publisherKey, signature) {
		return errors.New("invalid stark signature")
	}
	return nil
}

func verifyStarkSignature(xInt, yInt *big.Int, publicKey shared.PublisherKey, signature shared.StarkSignature) bool {
	publicKeyStr, _ := strings.CutPrefix(string(publicKey), "0x")
	pubKeyInt, _ := new(big.Int).SetString(publicKeyStr, 16)

	rStr, _ := strings.CutPrefix(signature.R, "0x")
	rInt, _ := new(big.Int).SetString(rStr, 16)

	sStr, _ := strings.CutPrefix(signature.S, "0x")
	sInt, _ := new(big.Int).SetString(sStr, 16)

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
	return C.validate_stark_signature(xIntBuf, yIntBuf, pubKeyIntBuf, rIntBuf, sIntBuf) != 0
}
