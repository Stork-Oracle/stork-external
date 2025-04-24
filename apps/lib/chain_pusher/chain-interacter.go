package chain_pusher

type ContractInteractor interface {
	ListenContractEvents(ch chan map[InternalEncodedAssetId]InternalStorkStructsTemporalNumericValue)
	PullValues(encodedAssetIds []InternalEncodedAssetId) (map[InternalEncodedAssetId]InternalStorkStructsTemporalNumericValue, error)
	BatchPushToContract(priceUpdates map[InternalEncodedAssetId]AggregatedSignedPrice) error
}
