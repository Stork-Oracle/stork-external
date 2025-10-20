package types

import (
	chain_pusher_types "github.com/Stork-Oracle/stork-external/apps/chain_pusher/pkg/types"
	"github.com/Stork-Oracle/stork-external/shared"
	"github.com/ethereum/go-ethereum/common"
)

type FirstPartyConfig struct {
	WebsocketPort   string
	ChainRpcUrl     string
	ChainWsUrl      string
	ContractAddress string
	AssetConfig     *AssetConfig
	GasLimit        uint64
}

// AssetConfig is the type representation of the asset-config.yaml file.
type AssetConfig struct {
	Assets map[shared.AssetID]AssetEntry `yaml:"assets"`
}

// AssetEntry is a single asset entry in the asset-config.yaml file.
type AssetEntry struct {
	AssetID                shared.AssetID      `yaml:"asset_id"`
	PercentChangeThreshold float64             `yaml:"percent_change_threshold"`
	FallbackPeriodSecs     uint64              `yaml:"fallback_period_sec"` //nolint:tagliatelle // Legacy
	PublicKey              shared.PublisherKey `yaml:"public_key"`
	Historical             bool                `yaml:"historical"`
}

// PublisherAssetPair contains the data needed for indexing the contract state.
type PublisherAssetPair struct {
	Address common.Address
	AssetID shared.AssetID
}

type ContractUpdate struct {
	Pubkey           common.Address
	ContractValueMap map[shared.AssetID]chain_pusher_types.InternalTemporalNumericValue
}
