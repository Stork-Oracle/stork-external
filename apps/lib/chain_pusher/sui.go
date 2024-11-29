type SuiContractInteracter struct {
	logger              zerolog.Logger
	contract            *contract.StorkSuiContract
	keyPair             *sui.KeyPair
	pollingFrequencySec int
}

func NewSuiContractInteracter(rpcUrl, contractAddr, keyFile string, pollingFreqSec int, logger zerolog.Logger) *SuiContractInteracter {

}

func (sci *SuiContractInteracter) ListenContractEvents(ch chan map[InternalEncodedAssetId]InternalStorkStructsTemporalNumericValue) {

}

func (sci *SuiContractInteracter) PullValues(encodedAssetIds []InternalEncodedAssetId) (map[InternalEncodedAssetId]InternalStorkStructsTemporalNumericValue, error) {

}

func (sci *SuiContractInteracter) BatchPushToContract(priceUpdates map[InternalEncodedAssetId]AggregatedSignedPrice) error {

}
