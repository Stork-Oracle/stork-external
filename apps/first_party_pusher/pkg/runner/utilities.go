package runner

import (
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"os"
	"strings"

	chain_pusher_types "github.com/Stork-Oracle/stork-external/apps/chain_pusher/pkg/types"
	"github.com/ethereum/go-ethereum/crypto"
	"gopkg.in/yaml.v2"
)

func LoadAssetConfig(filename string) (*chain_pusher_types.AssetConfig, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read asset config file: %w", err)
	}

	var config chain_pusher_types.AssetConfig

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
