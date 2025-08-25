package self_serve_chain_pusher

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
)

type AssetPushConfig struct {
	AssetId                 string  `yaml:"asset_id"`
	EncodedAssetId          string  `yaml:"encoded_asset_id"`
	PushIntervalSec         int     `yaml:"push_interval_sec"`
	PercentChangeThreshold  float64 `yaml:"percent_change_threshold"`
}

type AssetConfigFile struct {
	Assets map[string]AssetPushConfig `yaml:"assets"`
}

type EvmSelfServeConfig struct {
	WebsocketPort   string
	ChainRpcUrl     string
	ChainWsUrl      string
	ContractAddress string
	AssetConfig     *AssetConfigFile
	PrivateKey      *ecdsa.PrivateKey
	GasLimit        uint64
	LimitPerSecond  float64
	BurstLimit      int
}

type AssetPushState struct {
	AssetId              string
	Config               AssetPushConfig
	LastPrice            *big.Float
	LastPushTime         time.Time
	PendingValue         *ValueUpdate
	NextPushTime         time.Time
}

type ValueUpdate struct {
	Asset                string
	Value                *big.Float
	PublishTimestampNano int64
	Metadata             map[string]any
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