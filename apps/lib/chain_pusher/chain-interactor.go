package chain_pusher

import (
	"context"

	"github.com/rs/zerolog"
)

type ContractInteractor interface {
	ListenContractEvents(ctx context.Context, ch chan map[InternalEncodedAssetId]InternalTemporalNumericValue)
	PullValues(encodedAssetIds []InternalEncodedAssetId) (map[InternalEncodedAssetId]InternalTemporalNumericValue, error)
	BatchPushToContract(priceUpdates map[InternalEncodedAssetId]AggregatedSignedPrice) error
	GetWalletBalance() (float64, error)
	ConnectRest(restRpcUrl string) error
	ConnectWs(wsRpcUrl string) error
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

type FallbackContractInteractor struct {
	contractInteractor *ContractInteractor
	restRpcUrls        []string
	wsRpcUrls          []string
	logger             zerolog.Logger
}

func (f *FallbackContractInteractor) ConnectRest(restRpcUrl string) error {
	var err error
	for _, restRpcUrl := range f.restRpcUrls {
		f.logger.Info().Msgf("attempting connection to Rest rpc url %s", restRpcUrl)
		err = f.contractInteractor.ConnectRest(restRpcUrl)
		if err == nil {
			return nil
		}
		f.logger.Error().Err(err).Str("rpcUrl", restRpcUrl).Msgf("error connecting to Rest RPC, will attempt fallback")
	}
	f.logger.Error().Err(err).Msg("failed to connect to all supplied rest rpc urls!")
	return err
}

func (f *FallbackContractInteractor) ConnectWs(wsRpcUrl string) error {
	var err error
	for _, wsRpcUrl := range f.wsRpcUrls {
		f.logger.Info().Msgf("attempting connection to WS rpc url %s", wsRpcUrl)
		err = f.contractInteractor.ConnectWs(wsRpcUrl)
		if err == nil {
			return nil
		}
		f.logger.Error().Err(err).Str("rpcUrl", wsRpcUrl).Msgf("error connecting to WS RPC, will attempt fallback")
	}
	f.logger.Error().Err(err).Msg("failed to connect to all supplied ws rpc urls!")
	return err
}

func (f *FallbackContractInteractor) ListenContractEvents(ctx context.Context, ch chan map[InternalEncodedAssetId]InternalTemporalNumericValue) {
	f.ContractInteractor.ListenContractEvents(ctx, ch)
}

func (f *FallbackContractInteractor) PullValues(encodedAssetIds []InternalEncodedAssetId) (map[InternalEncodedAssetId]InternalTemporalNumericValue, error) {
	//TODO implement me
	panic("implement me")
}

func (f *FallbackContractInteractor) BatchPushToContract(priceUpdates map[InternalEncodedAssetId]AggregatedSignedPrice) error {
	//TODO implement me
	panic("implement me")
}

func (f *FallbackContractInteractor) GetWalletBalance() (float64, error) {
	//TODO implement me
	panic("implement me")
}
