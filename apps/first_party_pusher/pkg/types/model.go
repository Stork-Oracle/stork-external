package types

import (
	chain_pusher_types "github.com/Stork-Oracle/stork-external/apps/chain_pusher/pkg/types"
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

type ContractUpdate struct {
	Pubkey                 common.Address
	LatestContractValueMap map[string]chain_pusher_types.InternalTemporalNumericValue
}
