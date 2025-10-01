package types

import (
	"context"

	chain_pusher_types "github.com/Stork-Oracle/stork-external/apps/chain_pusher/pkg/types"
	publisher_agent "github.com/Stork-Oracle/stork-external/apps/publisher_agent/pkg"
	"github.com/Stork-Oracle/stork-external/shared"
	"github.com/ethereum/go-ethereum/common"
)

type ContractInteractor[T shared.Signature] interface {
	CheckPublisherUser(
		pubKey common.Address,
	) (bool, error)
	PullValues(
		pubKeyAssetIDPairs map[common.Address][]string,
	) ([]ContractUpdate, error)
	ListenContractEvents(
		ctx context.Context,
		ch chan ContractUpdate,
		pubKeyAssetIDPairs map[common.Address][]string,
	)
	BatchPushToContract(
		signedPriceUpdatesByAssetEntry map[chain_pusher_types.AssetEntry]publisher_agent.SignedPriceUpdate[T],
	) error
	Close()
}
