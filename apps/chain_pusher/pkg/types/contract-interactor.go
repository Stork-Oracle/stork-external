package types

import "context"

type ContractInteractor interface {
	ListenContractEvents(ctx context.Context, ch chan map[InternalEncodedAssetID]InternalTemporalNumericValue)
	PullValues(encodedAssetIDs []InternalEncodedAssetID) (map[InternalEncodedAssetID]InternalTemporalNumericValue, error)
	BatchPushToContract(priceUpdates map[InternalEncodedAssetID]AggregatedSignedPrice) error
	GetWalletBalance() (float64, error)
}

type MockContractInteractor struct {
}

func (m *MockContractInteractor) ListenContractEvents(
	ctx context.Context,
	ch chan map[InternalEncodedAssetID]InternalTemporalNumericValue,
) {
	// Do nothing
}

func (m *MockContractInteractor) PullValues(
	encodedAssetIDs []InternalEncodedAssetID,
) (map[InternalEncodedAssetID]InternalTemporalNumericValue, error) {
	// Do nothing
	return nil, nil
}

func (m *MockContractInteractor) BatchPushToContract(
	priceUpdates map[InternalEncodedAssetID]AggregatedSignedPrice,
) error {
	// Do nothing
	return nil
}

func (m *MockContractInteractor) GetWalletBalance() (float64, error) {
	return 0, nil
}
