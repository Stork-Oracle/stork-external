package runner

import (
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"
	"os"
	"strings"

	"github.com/Stork-Oracle/stork-external/apps/self_serve_chain_pusher/pkg/types"
	"github.com/ethereum/go-ethereum/crypto"
	"gopkg.in/yaml.v2"
)

func LoadAssetConfig(filename string) (*types.AssetConfigFile, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read asset config file: %w", err)
	}

	var config types.AssetConfigFile

	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse asset config YAML: %w", err)
	}

	return &config, nil
}

func LoadPrivateKey(filename string) (*ecdsa.PrivateKey, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key file: %w", err)
	}

	privateKeyHex := strings.TrimSpace(string(data))
	privateKeyHex = strings.TrimPrefix(privateKeyHex, "0x") // TODO: idk if we do this regularly

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

func GenerateNonce() *big.Int {
	nonce, _ := rand.Int(rand.Reader, big.NewInt(1<<62))

	return nonce
}
