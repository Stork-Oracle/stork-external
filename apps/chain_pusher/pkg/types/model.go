package types

import (
	"math/big"
	"os"

	"gopkg.in/yaml.v2"
)

// Config
type AssetConfig struct {
	Assets map[AssetID]AssetEntry `yaml:"assets"`
}

type AssetEntry struct {
	AssetID                AssetID        `yaml:"asset_id"`
	EncodedAssetID         EncodedAssetID `yaml:"encoded_asset_id"`
	PercentChangeThreshold float64        `yaml:"percent_change_threshold"`
	FallbackPeriodSecs     uint64         `yaml:"fallback_period_sec"`
}

func LoadConfig(filename string) (*AssetConfig, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var config AssetConfig
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// Stork Aggregator Types
type (
	AssetID        string
	EncodedAssetID string
	PublisherKey   string
	QuantizedPrice string
)

type EvmSignature struct {
	R string `json:"r"`
	S string `json:"s"`
	V string `json:"v"`
}

type StorkCalculationAlg struct {
	Type     string `json:"type"`
	Version  string `json:"version"`
	Checksum string `json:"checksum"`
}

type TimestampedSignature struct {
	Signature     EvmSignature `json:"signature"`
	TimestampNano uint64       `json:"timestamp"`
	MsgHash       string       `json:"msg_hash"`
}

type StorkSignedPrice struct {
	PublicKey            string               `json:"public_key"`
	EncodedAssetID       EncodedAssetID       `json:"encoded_asset_id"`
	QuantizedPrice       QuantizedPrice       `json:"price"`
	TimestampedSignature TimestampedSignature `json:"timestamped_signature"`
	PublisherMerkleRoot  string               `json:"publisher_merkle_root"`
	StorkCalculationAlg  StorkCalculationAlg  `json:"calculation_alg"`
}

type PublisherSignedPrice struct {
	PublisherKey         PublisherKey         `json:"publisher_key"`
	ExternalAssetID      string               `json:"external_asset_id"`
	QuantizedPrice       QuantizedPrice       `json:"price"`
	TimestampedSignature TimestampedSignature `json:"timestamped_signature"`
}

type AggregatedSignedPrice struct {
	TimestampNano    uint64                  `json:"timestamp"`
	AssetID          AssetID                 `json:"asset_id"`
	StorkSignedPrice *StorkSignedPrice       `json:"stork_signed_price,omitempty"`
	SignedPrices     []*PublisherSignedPrice `json:"signed_prices"`
}

type OraclePricesMessage struct {
	Type    string                           `json:"type"`
	TraceID string                           `json:"trace_id"`
	Data    map[string]AggregatedSignedPrice `json:"data"`
}

// Internal types
type InternalEncodedAssetID [32]byte

type InternalTemporalNumericValue struct {
	TimestampNs    uint64
	QuantizedValue *big.Int
}
