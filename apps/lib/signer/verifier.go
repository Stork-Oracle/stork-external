package signer

/*
#cgo LDFLAGS: -L/app/rust/stork/target/aarch64-unknown-linux-gnu/release -L./rust/stork/target/release -lstork
#cgo CFLAGS: -I/app/rust/stork/src -I./rust/stork/src
#include "signing.h"
*/
import "C"
import (
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

func VerifyEvmPublisherPrice(publishTimestamp int64, externalAssetId string, quantizedValue string, publisherKey PublisherKey, signature EvmSignature) error {
	publisherAddress := common.HexToAddress(string(publisherKey))
	payload := getPublisherEvmPricePayload(
		publishTimestamp,
		quantizedValue,
		externalAssetId,
		publisherAddress,
	)
	isValid, err := verifyEvmSignature(publisherAddress, payload, signature)
	if err != nil {
		return fmt.Errorf("failed to verify signature: %w", err)
	}
	if !isValid {
		return fmt.Errorf("invalid publisher signature")
	}
	return nil
}

func verifyEvmSignature(publisherAddress common.Address, payload [][]byte, signature EvmSignature) (bool, error) {
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

func VerifyStarkPublisherPrice(publishTimestamp int64, externalAssetId string, quantizedValue string, publisherKey PublisherKey, signature StarkSignature) error {
	xInt, yInt := getPublisherPriceStarkXY(publishTimestamp, externalAssetId, quantizedValue)
	isValid := verifyStarkSignature(xInt, yInt, publisherKey, signature)
	if !isValid {
		return errors.New("invalid stark signature")
	} else {
		return nil
	}
}

func verifyStarkSignature(xInt *big.Int, yInt *big.Int, publicKey PublisherKey, signature StarkSignature) bool {
	publicKeyStr, _ := strings.CutPrefix(string(publicKey), "0x")
	pubKeyInt := new(big.Int)
	pubKeyInt.SetString(publicKeyStr, 16)

	rStr, _ := strings.CutPrefix(signature.R, "0x")
	rInt := new(big.Int)
	rInt.SetString(rStr, 16)

	sStr, _ := strings.CutPrefix(signature.S, "0x")
	sInt := new(big.Int)
	sInt.SetString(sStr, 16)

	isValidInt := C.validate_stark_signature(
		createBufferFromBigInt(xInt),
		createBufferFromBigInt(yInt),
		createBufferFromBigInt(pubKeyInt),
		createBufferFromBigInt(rInt),
		createBufferFromBigInt(sInt),
	)
	isValid := isValidInt != 0

	return isValid
}
