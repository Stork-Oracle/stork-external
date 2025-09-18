// Package types contains types and some helper functions for the chain_pusher app.
package types

import (
	"fmt"
	"math/big"
	"os"

	"github.com/Stork-Oracle/stork-external/shared"
	"gopkg.in/yaml.v2"
)

// AssetConfig is the type representation of the asset-config.yaml file.
type AssetConfig struct {
	Assets map[shared.AssetID]AssetEntry `yaml:"assets"`
}

// AssetEntry is a single asset entry in the asset-config.yaml file.
type AssetEntry struct {
	AssetID                shared.AssetID        `yaml:"asset_id"`
	EncodedAssetID         shared.EncodedAssetID `yaml:"encoded_asset_id"`
	PercentChangeThreshold float64               `yaml:"percent_change_threshold"`
	FallbackPeriodSecs     uint64                `yaml:"fallback_period_sec"` //nolint:tagliatelle // Legacy
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

type StorkCalculationAlg struct {
	Type     string `json:"type"`
	Version  string `json:"version"`
	Checksum string `json:"checksum"`
}

//nolint:tagliatelle // Legacy
type StorkSignedPrice struct {
	PublicKey            string                                            `json:"public_key"`
	EncodedAssetID       shared.EncodedAssetID                             `json:"encoded_asset_id"`
	QuantizedPrice       shared.QuantizedPrice                             `json:"price"`
	TimestampedSignature shared.TimestampedSignature[*shared.EvmSignature] `json:"timestamped_signature"`
	PublisherMerkleRoot  string                                            `json:"publisher_merkle_root"`
	StorkCalculationAlg  StorkCalculationAlg                               `json:"calculation_alg"`
}

type PublisherSignedPrice struct {
	PublisherKey         shared.PublisherKey                               `json:"publisher_key"`
	ExternalAssetID      string                                            `json:"external_asset_id"`
	QuantizedPrice       shared.QuantizedPrice                             `json:"price"` //nolint:tagliatelle // Legacy
	TimestampedSignature shared.TimestampedSignature[*shared.EvmSignature] `json:"timestamped_signature"`
}

type AggregatedSignedPrice struct {
	TimestampNano    uint64                  `json:"timestamp"` //nolint:tagliatelle // Legacy
	AssetID          shared.AssetID          `json:"asset_id"`
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
