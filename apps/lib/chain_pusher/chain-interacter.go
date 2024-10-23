package chain_pusher

type ChainInteracter interface {
	PushToContract(updates map[InternalEncodedAssetId]InternalStorkStructsTemporalNumericValue) error
	ListenContractEvents(ch chan map[InternalEncodedAssetId]InternalStorkStructsTemporalNumericValue)
	PullValues(encodedAssetIds []InternalEncodedAssetId) (map[InternalEncodedAssetId]InternalStorkStructsTemporalNumericValue, error)
	BatchPushToContract(priceUpdates map[InternalEncodedAssetId]AggregatedSignedPrice) error
}
