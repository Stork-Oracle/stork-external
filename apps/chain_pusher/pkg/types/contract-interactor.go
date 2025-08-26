package types

import "context"

type ContractInteractor interface {
	ListenContractEvents(ctx context.Context, ch chan map[InternalEncodedAssetId]InternalTemporalNumericValue)
	PullValues(encodedAssetIds []InternalEncodedAssetId) (map[InternalEncodedAssetId]InternalTemporalNumericValue, error)
	BatchPushToContract(priceUpdates map[InternalEncodedAssetId]AggregatedSignedPrice) error
	GetWalletBalance() (float64, error)
}

type MockContractInteractor struct {
}

func (m *MockContractInteractor) ListenContractEvents(
	ctx context.Context,
	ch chan map[InternalEncodedAssetId]InternalTemporalNumericValue,
) {
	// Do nothing
}

func (m *MockContractInteractor) PullValues(
	encodedAssetIds []InternalEncodedAssetId,
) (map[InternalEncodedAssetId]InternalTemporalNumericValue, error) {
	// Do nothing
	return nil, nil
}

func (m *MockContractInteractor) BatchPushToContract(
	priceUpdates map[InternalEncodedAssetId]AggregatedSignedPrice,
) error {
	// Do nothing
	return nil
}

func (m *MockContractInteractor) GetWalletBalance() (float64, error) {
	return 0, nil
}
