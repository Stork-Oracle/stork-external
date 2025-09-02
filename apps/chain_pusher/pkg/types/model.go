// Package types contains types and some helper functions for the chain_pusher app.
package types

import (
	"fmt"
	"math/big"
	"os"

	"gopkg.in/yaml.v2"
)

// AssetConfig is the type representation of the asset-config.yaml file.
type AssetConfig struct {
	Assets map[AssetID]AssetEntry `yaml:"assets"`
}

// AssetEntry is a single asset entry in the asset-config.yaml file.
type AssetEntry struct {
	AssetID                AssetID        `yaml:"asset_id"`
	EncodedAssetID         EncodedAssetID `yaml:"encoded_asset_id"`
	PercentChangeThreshold float64        `yaml:"percent_change_threshold"`
	FallbackPeriodSecs     uint64         `yaml:"fallback_period_sec"` //nolint:tagliatelle // Legacy
}

// LoadConfig loads the asset config from the given filename.
func LoadConfig(filename string) (*AssetConfig, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config AssetConfig

	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}

// Stork Aggregator Types.
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
	TimestampNano uint64       `json:"timestamp"` //nolint:tagliatelle // Legacy
	MsgHash       string       `json:"msg_hash"`
}

type StorkSignedPrice struct {
	PublicKey            string               `json:"public_key"`
	EncodedAssetID       EncodedAssetID       `json:"encoded_asset_id"`
	QuantizedPrice       QuantizedPrice       `json:"price"` //nolint:tagliatelle // Legacy
	TimestampedSignature TimestampedSignature `json:"timestamped_signature"`
	PublisherMerkleRoot  string               `json:"publisher_merkle_root"`
	StorkCalculationAlg  StorkCalculationAlg  `json:"calculation_alg"` //nolint:tagliatelle // Legacy
}

type PublisherSignedPrice struct {
	PublisherKey         PublisherKey         `json:"publisher_key"`
	ExternalAssetID      string               `json:"external_asset_id"`
	QuantizedPrice       QuantizedPrice       `json:"price"` //nolint:tagliatelle // Legacy
	TimestampedSignature TimestampedSignature `json:"timestamped_signature"`
}

type AggregatedSignedPrice struct {
	TimestampNano    uint64                  `json:"timestamp"` //nolint:tagliatelle // Legacy
	AssetID          AssetID                 `json:"asset_id"`
	StorkSignedPrice *StorkSignedPrice       `json:"stork_signed_price,omitempty"`
	SignedPrices     []*PublisherSignedPrice `json:"signed_prices"`
}

type OraclePricesMessage struct {
	Type    string                           `json:"type"`
	TraceID string                           `json:"trace_id"`
	Data    map[string]AggregatedSignedPrice `json:"data"`
}

// InternalEncodedAssetID is the internal (main) representation of the encoded asset ID.
type InternalEncodedAssetID [32]byte

// InternalTemporalNumericValue is the internal (main) representation of the temporal numeric value.
type InternalTemporalNumericValue struct {
	TimestampNs    uint64
	QuantizedValue *big.Int
}
