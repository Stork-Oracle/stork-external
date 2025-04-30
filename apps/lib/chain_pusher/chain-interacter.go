package chain_pusher

import "context"

type ContractInteractor interface {
	ListenContractEvents(ctx context.Context, ch chan map[InternalEncodedAssetId]InternalStorkStructsTemporalNumericValue)
	PullValues(encodedAssetIds []InternalEncodedAssetId) (map[InternalEncodedAssetId]InternalStorkStructsTemporalNumericValue, error)
	BatchPushToContract(priceUpdates map[InternalEncodedAssetId]AggregatedSignedPrice) error
}

type MockContractInteractor struct {
	ContractInteractor
}

func (m *MockContractInteractor) ListenContractEvents(
	ctx context.Context,
	ch chan map[InternalEncodedAssetId]InternalStorkStructsTemporalNumericValue,
) {
	// Do nothing
}

func (m *MockContractInteractor) PullValues(
	encodedAssetIds []InternalEncodedAssetId,
) (map[InternalEncodedAssetId]InternalStorkStructsTemporalNumericValue, error) {
	// Do nothing
	return nil, nil
}

func (m *MockContractInteractor) BatchPushToContract(
	priceUpdates map[InternalEncodedAssetId]AggregatedSignedPrice,
) error {
	// Do nothing
	return nil
}
