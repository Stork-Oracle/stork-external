package types

import (
	"context"
)

type ContractInteractor interface {
	ListenContractEvents(ctx context.Context, ch chan map[InternalEncodedAssetID]InternalTemporalNumericValue)
	PullValues(
		ctx context.Context,
		encodedAssetIDs []InternalEncodedAssetID,
	) (map[InternalEncodedAssetID]InternalTemporalNumericValue, error)
	BatchPushToContract(ctx context.Context, priceUpdates map[InternalEncodedAssetID]AggregatedSignedPrice) error
	GetWalletBalance(ctx context.Context) (float64, error)
	ConnectHTTP(ctx context.Context, url string) error
	ConnectWs(ctx context.Context, url string) error
}
