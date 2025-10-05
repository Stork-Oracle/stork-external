package types

import (
	"context"

	publisher_agent "github.com/Stork-Oracle/stork-external/apps/publisher_agent/pkg"
	"github.com/Stork-Oracle/stork-external/shared"
	"github.com/ethereum/go-ethereum/common"
)

type ContractInteractor[T shared.Signature] interface {
	CheckPublisherUser(
		pubKey common.Address,
	) (bool, error)
	PullValues(
		pubKeyAssetIDPairs map[common.Address][]shared.AssetID,
		assetIDtoEncodedAssetID map[shared.AssetID]shared.EncodedAssetID,
	) ([]ContractUpdate, error)
	ListenContractEvents(
		ctx context.Context,
		ch chan ContractUpdate,
		pubKeyAssetIDPairs map[common.Address][]shared.AssetID,
	)
	BatchPushToContract(
		signedPriceUpdatesByAssetEntry map[AssetEntry]publisher_agent.SignedPriceUpdate[T],
	) error
	Close()
}
