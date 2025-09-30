package types

import (
	"context"
	"fmt"
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
	ConnectHTTP(url string) error
	ConnectWs(url string) error
}

type FallbackContractInteractor struct {
	contractInteractor        ContractInteractor
	httpRpcUrls               []string
	wsRpcUrls                 []string
	firstHTTPRpcUrlSuccessful bool
	logger                    zerolog.Logger
}

func NewFallbackContractInteractor(
	interactor ContractInteractor,
	httpRpcUrls []string,
	wsRpcUrls []string,
	logger zerolog.Logger,
) *FallbackContractInteractor {
	return &FallbackContractInteractor{
		contractInteractor:        interactor,
		httpRpcUrls:               httpRpcUrls,
		wsRpcUrls:                 wsRpcUrls,
		logger:                    logger,
		firstHTTPRpcUrlSuccessful: false,
	}
}

func (f *FallbackContractInteractor) ConnectHTTP(httpRpcUrl string) error {
	f.logger.Info().Msgf("attempting connection to HTTO rpc url %s", httpRpcUrl)
	err := f.contractInteractor.ConnectHTTP(httpRpcUrl)
	if err != nil {
		return fmt.Errorf("failed to connect to HTTP rpc url %s", httpRpcUrl)
	}

	return nil
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

func (f *FallbackContractInteractor) ListenContractEvents(
	ctx context.Context,
	ch chan map[InternalEncodedAssetID]InternalTemporalNumericValue,
) {
	f.contractInteractor.ListenContractEvents(ctx, ch)
}

func (f *FallbackContractInteractor) PullValues(
	encodedAssetIDs []InternalEncodedAssetID,
) (map[InternalEncodedAssetID]InternalTemporalNumericValue, error) {
	result, err := f.runWithFallback(
		"pullValues",
		func() (any, error) {
			return f.contractInteractor.PullValues(encodedAssetIDs)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to pull values from all supplied HTTP rpc urls: %w", err)
	}

	values, success := result.(map[InternalEncodedAssetID]InternalTemporalNumericValue)
	if !success {
		return nil, fmt.Errorf("could not convert result to values: %w", err)
	}

	return values, nil
}

func (f *FallbackContractInteractor) BatchPushToContract(priceUpdates map[InternalEncodedAssetID]AggregatedSignedPrice) error {
	_, err := f.runWithFallback(
		"pushBatch",
		func() (any, error) {
			err := f.contractInteractor.BatchPushToContract(priceUpdates)
			return nil, err
		},
	)
	if err != nil {
		return fmt.Errorf("failed to push batch from all supplied HTTP rpc urls: %w", err)
	}

	return nil
}

func (f *FallbackContractInteractor) GetWalletBalance() (float64, error) {
	result, err := f.runWithFallback(
		"pullValues",
		func() (any, error) {
			return f.contractInteractor.GetWalletBalance()
		},
	)
	if err != nil {
		return math.NaN(), fmt.Errorf("failed to pull wallet balance from all supplied HTTP rpc urls: %w", err)
	}

	balance, success := result.(float64)

	if !success {
		return math.NaN(), fmt.Errorf("could not convert result to float: %w", err)
	}

	return balance, nil
}

func (f *FallbackContractInteractor) runWithFallback(contractFuncName string, contractFunc func() (any, error)) (any, error) {
	var err error

	var result any

	for idx, httpRpcUrl := range f.wsRpcUrls {
		if idx > 0 || !f.firstHTTPRpcUrlSuccessful {
			err = f.contractInteractor.ConnectHTTP(httpRpcUrl)
			if err != nil {
				f.logger.Error().Err(err).
					Str("httpRpcUrl", httpRpcUrl).
					Msgf("error connecting to rpc http client, will attempt fallback")

				f.firstHTTPRpcUrlSuccessful = false

				continue
			}
		}

		result, err = contractFunc()
		if err == nil {
			if idx == 0 {
				f.firstHTTPRpcUrlSuccessful = true
			}
			return result, nil
		}

		f.firstHTTPRpcUrlSuccessful = false
		f.logger.Error().Err(err).
			Str("rpcUrl", httpRpcUrl).
			Str("contractFunction", contractFuncName).
			Msgf("error calling contract function on rpc, will attempt fallback")
	}

	f.logger.Error().Err(err).
		Str("contractFunction", contractFuncName).
		Msg("failed to pull values from all supplied http rpc urls!")

	return nil, err
}
