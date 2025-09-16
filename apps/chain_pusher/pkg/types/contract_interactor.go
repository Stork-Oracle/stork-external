package types

import (
	"context"
	"math"

	"github.com/rs/zerolog"
)

type ContractInteractor interface {
	ListenContractEvents(ctx context.Context, ch chan map[InternalEncodedAssetID]InternalTemporalNumericValue)
	PullValues(
		encodedAssetIDs []InternalEncodedAssetID,
	) (map[InternalEncodedAssetID]InternalTemporalNumericValue, error)
	BatchPushToContract(priceUpdates map[InternalEncodedAssetID]AggregatedSignedPrice) error
	GetWalletBalance() (float64, error)
	ConnectRest(url string) error
	ConnectWs(url string) error
}

type FallbackContractInteractor struct {
	contractInteractor ContractInteractor
	restRpcUrls        []string
	wsRpcUrls          []string
	logger             zerolog.Logger
}

func NewFallbackContractInteractor(
	interactor ContractInteractor,
	restRpcUrls []string,
	wsRpcUrls []string,
	logger zerolog.Logger,
) *FallbackContractInteractor {
	return &FallbackContractInteractor{
		contractInteractor: interactor,
		restRpcUrls:        restRpcUrls,
		wsRpcUrls:          wsRpcUrls,
		logger:             logger,
	}
}

func (f *FallbackContractInteractor) ConnectRest(_ string) error {
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

func (f *FallbackContractInteractor) ConnectWs(_ string) error {
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

func (f *FallbackContractInteractor) ListenContractEvents(ctx context.Context, ch chan map[InternalEncodedAssetID]InternalTemporalNumericValue) {
	f.contractInteractor.ListenContractEvents(ctx, ch)
}

func (f *FallbackContractInteractor) PullValues(encodedAssetIds []InternalEncodedAssetID) (map[InternalEncodedAssetID]InternalTemporalNumericValue, error) {
	var err error
	var result map[InternalEncodedAssetID]InternalTemporalNumericValue
	for _, restRpcUrl := range f.wsRpcUrls {
		err = f.contractInteractor.ConnectRest(restRpcUrl)
		if err != nil {
			f.logger.Error().Err(err).Str("rpcUrl", restRpcUrl).Msgf("error connecting to rpc rest client, will attempt fallback")
			continue
		}

		result, err = f.contractInteractor.PullValues(encodedAssetIds)
		if err == nil {
			return result, nil
		}
		f.logger.Error().Err(err).Str("rpcUrl", restRpcUrl).Msgf("error pulling values from RPC, will attempt fallback")
	}
	f.logger.Error().Err(err).Msg("failed to pull values from all supplied rest rpc urls!")
	return nil, err
}

func (f *FallbackContractInteractor) BatchPushToContract(priceUpdates map[InternalEncodedAssetID]AggregatedSignedPrice) error {
	var err error
	for _, restRpcUrl := range f.wsRpcUrls {
		err = f.contractInteractor.ConnectRest(restRpcUrl)
		if err != nil {
			f.logger.Error().Err(err).Str("rpcUrl", restRpcUrl).Msgf("error connecting to rpc rest client, will attempt fallback")
			continue
		}

		err = f.contractInteractor.BatchPushToContract(priceUpdates)
		if err == nil {
			return nil
		}
		f.logger.Error().Err(err).Str("rpcUrl", restRpcUrl).Msgf("error running batch push with RPC, will attempt fallback")
	}
	f.logger.Error().Err(err).Msg("failed to batch push with all supplied rest rpc urls!")
	return err
}

func (f *FallbackContractInteractor) GetWalletBalance() (float64, error) {
	var err error
	var result float64
	for _, restRpcUrl := range f.wsRpcUrls {
		err = f.contractInteractor.ConnectRest(restRpcUrl)
		if err != nil {
			f.logger.Error().Err(err).Str("rpcUrl", restRpcUrl).Msgf("error connecting to rpc rest client, will attempt fallback")
			continue
		}

		result, err = f.contractInteractor.GetWalletBalance()
		if err == nil {
			return result, nil
		}
		f.logger.Error().Err(err).Str("rpcUrl", restRpcUrl).Msgf("error pulling wallet balance from RPC, will attempt fallback")
	}
	f.logger.Error().Err(err).Msg("failed to pull wallet balance from all supplied rest rpc urls!")
	return math.NaN(), err
}
