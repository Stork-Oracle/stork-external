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
	AssetConfig     *chain_pusher_types.AssetConfig
	GasLimit        uint64
}

type PublisherAssetPair struct {
	Address        common.Address
	EncodedAssetID shared.EncodedAssetID
}

type ContractUpdate struct {
	Pubkey                 common.Address
	LatestContractValueMap map[string]chain_pusher_types.InternalTemporalNumericValue
}
