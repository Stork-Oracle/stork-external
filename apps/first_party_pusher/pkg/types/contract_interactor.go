package types

import (
	"context"

	chain_pusher_types "github.com/Stork-Oracle/stork-external/apps/chain_pusher/pkg/types"
	publisher_agent "github.com/Stork-Oracle/stork-external/apps/publisher_agent/pkg"
	"github.com/Stork-Oracle/stork-external/shared"
)

type ContractInteractor[T shared.Signature] interface {
	PushSignedPriceUpdate(
		ctx context.Context,
		asset chain_pusher_types.AssetEntry,
		signedPriceUpdate publisher_agent.SignedPriceUpdate[T],
	) error
	Close()
}
