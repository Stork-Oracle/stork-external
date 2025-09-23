package types

import (
	"crypto/ecdsa"
	"math/big"
	"time"

	chain_pusher_types "github.com/Stork-Oracle/stork-external/apps/chain_pusher/pkg/types"
	publisher_agent "github.com/Stork-Oracle/stork-external/apps/publisher_agent/pkg"
	"github.com/Stork-Oracle/stork-external/shared"
)

type FirstPartyConfig struct {
	WebsocketPort   string
	ChainRpcUrl     string
	ChainWsUrl      string
	ContractAddress string
	AssetConfig     *chain_pusher_types.AssetConfig
	PrivateKey      *ecdsa.PrivateKey
	GasLimit        uint64
}

type AssetPushState[T shared.Signature] struct {
	AssetID                  shared.AssetID
	Config                   chain_pusher_types.AssetEntry
	LastPrice                *big.Float
	LastPushTime             time.Time
	PendingSignedPriceUpdate *publisher_agent.SignedPriceUpdate[T]
	NextPushTime             time.Time
}
