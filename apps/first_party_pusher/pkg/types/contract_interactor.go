package types

import (
	"context"

	publisher_agent "github.com/Stork-Oracle/stork-external/apps/publisher_agent/pkg"
	"github.com/Stork-Oracle/stork-external/shared"
)

type ContractInteractor interface {
	PushSignedPriceUpdate(
		ctx context.Context,
		asset AssetPushConfig,
		signedPriceUpdate publisher_agent.SignedPriceUpdate[*shared.EvmSignature],
	) error
	Close()
}
