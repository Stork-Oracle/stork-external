package types

import (
	"crypto/ecdsa"
	"math/big"
	"time"

	publisher_agent "github.com/Stork-Oracle/stork-external/apps/publisher_agent/pkg"
	"github.com/Stork-Oracle/stork-external/shared"
)

// TODO: copy chain_pusher/pkg/types/model.go structs
type AssetPushConfig struct {
	AssetID                shared.AssetID        `yaml:"asset_id"`
	EncodedAssetID         shared.EncodedAssetID `yaml:"encoded_asset_id"`
	PushIntervalSec        int                   `yaml:"push_interval_sec"`
	PercentChangeThreshold float64               `yaml:"percent_change_threshold"`
}

type AssetConfigFile struct {
	Assets map[shared.AssetID]AssetPushConfig `yaml:"assets"`
}

type FirstPartyConfig struct {
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
