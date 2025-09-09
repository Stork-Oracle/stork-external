package testutil

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/Stork-Oracle/stork-external/apps/chain_pusher/pkg/pusher"
	"github.com/Stork-Oracle/stork-external/apps/chain_pusher/pkg/types"
)

// This file allows access to pre-generated valid websocket message for integration testing.
// These messages were collected from the production websocket at wss://api.jp.stork-oracle.network/evm/subscribe.
// There are 4 assets: BTCUSD, ETHUSD, SOLUSD, and SUIUSD, all with 50 collected messages.

var (
	ErrFailedToGetPrices = errors.New("failed to get prices")
	ErrNoMorePrices      = errors.New("no more prices")
)

// directory containing the pre-generated websocket message json files
const (
	messagesDir          = "testdata"
	BtcUsdEncodedAssetID = "0x7404e3d104ea7841c3d9e6fd20adfe99b4ad586bc08d8f3bd3afef894cf184de"
	EthUsdEncodedAssetID = "0x59102b37de83bdda9f38ac8254e596f0d9ac61d2035c07936675e87342817160"
	SolUsdEncodedAssetID = "0x1dcd89dfded9e8a9b0fa1745a8ebbacbb7c81e33d5abc81616633206d932e837"
	SuiUsdEncodedAssetID = "0xa24cc95a4f3d70a0a2f7ac652b67a4a73791631ff06b4ee7f729097311169b81"
)

type CapturedAssetID string

type SampleAggregatedSignedPrices struct {
	btcUsd               []types.AggregatedSignedPrice
	ethUsd               []types.AggregatedSignedPrice
	solUsd               []types.AggregatedSignedPrice
	suiUsd               []types.AggregatedSignedPrice
	btcUsdEncodedAssetID types.InternalEncodedAssetID
	ethUsdEncodedAssetID types.InternalEncodedAssetID
	solUsdEncodedAssetID types.InternalEncodedAssetID
	suiUsdEncodedAssetID types.InternalEncodedAssetID
}

func LoadAggregatedSignedPrices() (*SampleAggregatedSignedPrices, error) {
	// BTCUSD
	btcPrices, err := loadWsMessages("BTCUSD")
	if err != nil {
		return nil, fmt.Errorf("failed to read BTCUSD.json: %w", err)
	}

	ethPrices, err := loadWsMessages("ETHUSD")
	if err != nil {
		return nil, fmt.Errorf("failed to read ETHUSD.json: %w", err)
	}

	solPrices, err := loadWsMessages("SOLUSD")
	if err != nil {
		return nil, fmt.Errorf("failed to read SOLUSD.json: %w", err)
	}

	suiPrices, err := loadWsMessages("SUIUSD")
	if err != nil {
		return nil, fmt.Errorf("failed to read SUIUSD.json: %w", err)
	}

	btcUsdEncodedAssetID, err := pusher.HexStringToByte32(BtcUsdEncodedAssetID)
	if err != nil {
		return nil, fmt.Errorf("failed to convert BTCUSD encoded asset ID to byte32: %w", err)
	}

	ethUsdEncodedAssetID, err := pusher.HexStringToByte32(EthUsdEncodedAssetID)
	if err != nil {
		return nil, fmt.Errorf("failed to convert ETHUSD encoded asset ID to byte32: %w", err)
	}

	solUsdEncodedAssetID, err := pusher.HexStringToByte32(SolUsdEncodedAssetID)
	if err != nil {
		return nil, fmt.Errorf("failed to convert SOLUSD encoded asset ID to byte32: %w", err)
	}

	suiUsdEncodedAssetID, err := pusher.HexStringToByte32(SuiUsdEncodedAssetID)
	if err != nil {
		return nil, fmt.Errorf("failed to convert SUIUSD encoded asset ID to byte32: %w", err)
	}

	return &SampleAggregatedSignedPrices{
		btcUsd:               btcPrices,
		ethUsd:               ethPrices,
		solUsd:               solPrices,
		suiUsd:               suiPrices,
		btcUsdEncodedAssetID: btcUsdEncodedAssetID,
		ethUsdEncodedAssetID: ethUsdEncodedAssetID,
		solUsdEncodedAssetID: solUsdEncodedAssetID,
		suiUsdEncodedAssetID: suiUsdEncodedAssetID,
	}, nil
}

func (s *SampleAggregatedSignedPrices) NextBtcUsd() (*types.AggregatedSignedPrice, error) {
	if len(s.btcUsd) == 0 {
		return nil, ErrNoMorePrices
	}

	price := s.btcUsd[0]
	s.btcUsd = s.btcUsd[1:]

	return &price, nil
}

func (s *SampleAggregatedSignedPrices) NextEthUsd() (*types.AggregatedSignedPrice, error) {
	if len(s.ethUsd) == 0 {
		return nil, ErrNoMorePrices
	}

	price := s.ethUsd[0]
	s.ethUsd = s.ethUsd[1:]

	return &price, nil
}

func (s *SampleAggregatedSignedPrices) NextSolUsd() (*types.AggregatedSignedPrice, error) {
	if len(s.solUsd) == 0 {
		return nil, ErrNoMorePrices
	}

	price := s.solUsd[0]
	s.solUsd = s.solUsd[1:]

	return &price, nil
}

func (s *SampleAggregatedSignedPrices) NextSuiUsd() (*types.AggregatedSignedPrice, error) {
	if len(s.suiUsd) == 0 {
		return nil, ErrNoMorePrices
	}

	price := s.suiUsd[0]
	s.suiUsd = s.suiUsd[1:]

	return &price, nil
}

func (s *SampleAggregatedSignedPrices) BtcUsdEncodedAssetID() types.InternalEncodedAssetID {
	return s.btcUsdEncodedAssetID
}

func (s *SampleAggregatedSignedPrices) EthUsdEncodedAssetID() types.InternalEncodedAssetID {
	return s.ethUsdEncodedAssetID
}

func (s *SampleAggregatedSignedPrices) SolUsdEncodedAssetID() types.InternalEncodedAssetID {
	return s.solUsdEncodedAssetID
}

func (s *SampleAggregatedSignedPrices) SuiUsdEncodedAssetID() types.InternalEncodedAssetID {
	return s.suiUsdEncodedAssetID
}

func (s *SampleAggregatedSignedPrices) AllEncodedAssetIDs() []types.InternalEncodedAssetID {
	return []types.InternalEncodedAssetID{s.btcUsdEncodedAssetID, s.ethUsdEncodedAssetID, s.solUsdEncodedAssetID, s.suiUsdEncodedAssetID}
}

func loadWsMessages(assetID string) ([]types.AggregatedSignedPrice, error) {
	_, thisFile, _, _ := runtime.Caller(0)
	file, err := os.ReadFile(filepath.Join(filepath.Dir(thisFile), messagesDir, assetID+".json"))
	if err != nil {
		return nil, fmt.Errorf("failed to read %s.json: %w", assetID, err)
	}

	var messages []types.OraclePricesMessage

	err = json.Unmarshal(file, &messages)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal %s.json: %w", assetID, err)
	}

	prices := make([]types.AggregatedSignedPrice, len(messages))
	for i, message := range messages {
		price, ok := message.Data[assetID]
		if !ok {
			return nil, ErrFailedToGetPrices
		}

		prices[i] = price
	}

	return prices, nil
}
