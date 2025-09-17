package self_serve_evm

import (
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"
	"os"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"gopkg.in/yaml.v2"

	publisher_agent "github.com/Stork-Oracle/stork-external/apps/publisher_agent/pkg"
	"github.com/Stork-Oracle/stork-external/shared"
)

type AssetPushConfig struct {
	AssetID                shared.AssetID        `yaml:"asset_id"`
	EncodedAssetID         shared.EncodedAssetID `yaml:"encoded_asset_id"`
	PushIntervalSec        int                   `yaml:"push_interval_sec"`
	PercentChangeThreshold float64               `yaml:"percent_change_threshold"`
}

type AssetConfigFile struct {
	Assets map[shared.AssetID]AssetPushConfig `yaml:"assets"`
}

type EvmSelfServeConfig struct {
	WebsocketPort   string
	ChainRpcUrl     string
	ChainWsUrl      string
	ContractAddress string
	AssetConfig     *AssetConfigFile
	PrivateKey      *ecdsa.PrivateKey
	GasLimit        uint64
}

type AssetPushState struct {
	AssetID                  shared.AssetID
	Config                   AssetPushConfig
	LastPrice                *big.Float
	LastPushTime             time.Time
	PendingSignedPriceUpdate *publisher_agent.SignedPriceUpdate[*shared.EvmSignature]
	NextPushTime             time.Time
}

func LoadAssetConfig(filename string) (*AssetConfigFile, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read asset config file: %w", err)
	}

	var config AssetConfigFile
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

func GenerateNonce() *big.Int {
	nonce, _ := rand.Int(rand.Reader, big.NewInt(1<<62))
	return nonce
}
