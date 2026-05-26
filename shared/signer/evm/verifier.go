package evm

import (
	"errors"
	"fmt"

	"github.com/Stork-Oracle/stork-external/shared"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

func VerifyPublisherPrice(
	publishTimestampNano int64,
	externalAssetID string,
	quantizedValue string,
	publisherKey shared.PublisherKey,
	signature shared.EvmSignature,
) error {
	publisherAddress := common.HexToAddress(string(publisherKey))
	payload := getPublisherPricePayload(
		publishTimestampNano,
		quantizedValue,
		externalAssetID,
		publisherAddress,
	)
	isValid, err := VerifySignature(publisherAddress, payload, signature)
	if err != nil {
		return fmt.Errorf("failed to verify signature: %w", err)
	}
	if !isValid {
		return errors.New("invalid publisher signature")
	}
	return nil
}

func VerifySignature(
	publisherAddress common.Address,
	payload [][]byte,
	signature shared.EvmSignature,
) (bool, error) {
	_, prefixedHash := getHashes(payload)
	storkSignatureBytes, err := signatureToBytes(signature)
	if err != nil {
		return false, fmt.Errorf("failed to convert signature to bytes: %v", err)
	}

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

// VerifyAuth verifies an EVM-signed Stork auth header. The signature is a
// 0x-prefixed hex string concatenating R||S||V (32, 32, 1 bytes).
func VerifyAuth(
	timestampNano int64,
	publicKey shared.PublisherKey,
	signature string,
) error {
	stripped := strip0x(signature)
	if len(stripped) != 130 {
		return errors.New("invalid EVM signature length")
	}
	evmSignature := shared.EvmSignature{
		R: "0x" + stripped[:64],
		S: "0x" + stripped[64:128],
		V: "0x" + stripped[128:],
	}
	if err := VerifyPublisherPrice(timestampNano, StorkAuthAssetID, StorkMagicNumber, publicKey, evmSignature); err != nil {
		return fmt.Errorf("invalid evm auth signature: %w", err)
	}
	return nil
}
