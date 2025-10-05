package runner

import (
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"os"
	"strings"

	"github.com/Stork-Oracle/stork-external/apps/first_party_pusher/pkg/types"
	publisher_agent "github.com/Stork-Oracle/stork-external/apps/publisher_agent/pkg"
	"github.com/Stork-Oracle/stork-external/shared"
	"github.com/Stork-Oracle/stork-external/shared/signer"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/rs/zerolog"
	"gopkg.in/yaml.v2"
)

func LoadAssetConfig(filename string) (*types.AssetConfig, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read asset config file: %w", err)
	}

	var config types.AssetConfig

	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse asset config YAML: %w", err)
	}

	return &config, nil
}

func LoadPrivateKey(filename string) (*ecdsa.PrivateKey, error) {
	privateKeyHex := os.Getenv("PUSHER_PRIVATE_KEY")
	if privateKeyHex == "" {
		privateKeyRaw, err := os.ReadFile(filename)
		if err != nil {
			return nil, fmt.Errorf("failed to read private key file: %w", err)
		}

		privateKeyHex = strings.TrimSpace(string(privateKeyRaw))
	}

	privateKeyHex = strings.TrimPrefix(privateKeyHex, "0x")

	privateKeyBytes, err := hex.DecodeString(privateKeyHex)
	if err != nil {
		return nil, fmt.Errorf("failed to decode private key hex: %w", err)
	}

	privateKey, err := crypto.ToECDSA(privateKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to convert to ECDSA private key: %w", err)
	}

	return privateKey, nil
}

func CreateTestEvmSignedPriceUpdate(
	privateKey signer.EvmPrivateKey,
	publishTimestampNano int64,
	assetID shared.AssetID,
	quantizedPrice shared.QuantizedPrice,
	externalAssetID string,
	signatureType shared.SignatureType,
) (publisher_agent.SignedPriceUpdate[*shared.EvmSignature], error) {
	logger := zerolog.New(nil)

	signer, err := signer.NewEvmSigner(privateKey, logger)
	if err != nil {
		return publisher_agent.SignedPriceUpdate[*shared.EvmSignature]{}, err
	}

	timestampedSig, externalAssetID, err := signer.SignPublisherPrice(
		publishTimestampNano,
		string(assetID),
		string(quantizedPrice),
	)
	if err != nil {
		return publisher_agent.SignedPriceUpdate[*shared.EvmSignature]{}, err
	}

	return publisher_agent.SignedPriceUpdate[*shared.EvmSignature]{
		OracleID: "local",
		AssetID:  assetID,
		Trigger:  publisher_agent.UnspecifiedTriggerType,
		SignedPrice: publisher_agent.SignedPrice[*shared.EvmSignature]{
			PublisherKey:         shared.PublisherKey(signer.GetPublisherKey()),
			ExternalAssetID:      externalAssetID,
			SignatureType:        signatureType,
			QuantizedPrice:       quantizedPrice,
			TimestampedSignature: *timestampedSig,
			Metadata:             publisher_agent.Metadata{},
		},
	}, nil
}
