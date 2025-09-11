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

// This file allows access to pre-generated signed and valid websocket messages.
// These messages were collected from a local version of the stork aggregator with pubkey:
//
//     "0xC4A02e7D370402F4afC36032076B05e74FF81786"
//
// There are 5 assets:
//
//     - POSITIVE_ASSET_1
//     - POSITIVE_ASSET_2
//     - POSITIVE_ASSET_3
//     - POSITIVE_ASSET_4
//     - NEGATIVE_ASSET_1
//
// all with 50 collected messages.
// The assets that contain "POSITIVE" in the name have all positive prices.
// The assets that contain "NEGATIVE" in the name have all negative prices.

var (
	ErrFailedToGetPrices = errors.New("failed to get prices")
	ErrNoMorePrices      = errors.New("no more prices")
	ErrFailedToGetCaller = errors.New("failed to get caller")
	ErrNoPricesInFile    = errors.New("no prices in file")
)

// directory containing the pre-generated signed and valid websocket message json files.
const messagesDir = "testdata"

type CapturedAssetID string

type SampleAggregatedSignedPrices struct {
	positiveAsset1               []types.AggregatedSignedPrice
	positiveAsset2               []types.AggregatedSignedPrice
	positiveAsset3               []types.AggregatedSignedPrice
	positiveAsset4               []types.AggregatedSignedPrice
	negativeAsset1               []types.AggregatedSignedPrice
	positiveAsset1EncodedAssetID types.InternalEncodedAssetID
	positiveAsset2EncodedAssetID types.InternalEncodedAssetID
	positiveAsset3EncodedAssetID types.InternalEncodedAssetID
	positiveAsset4EncodedAssetID types.InternalEncodedAssetID
	negativeAsset1EncodedAssetID types.InternalEncodedAssetID
}

func LoadAggregatedSignedPrices() (*SampleAggregatedSignedPrices, error) {
	positiveAsset1Prices, err := loadWsMessages("POSITIVE_ASSET_1")
	if err != nil {
		return nil, fmt.Errorf("failed to read POSITIVE_ASSET_1.json: %w", err)
	}

	positiveAsset2Prices, err := loadWsMessages("POSITIVE_ASSET_2")
	if err != nil {
		return nil, fmt.Errorf("failed to read POSITIVE_ASSET_2.json: %w", err)
	}

	positiveAsset3Prices, err := loadWsMessages("POSITIVE_ASSET_3")
	if err != nil {
		return nil, fmt.Errorf("failed to read POSITIVE_ASSET_3.json: %w", err)
	}

	positiveAsset4Prices, err := loadWsMessages("POSITIVE_ASSET_4")
	if err != nil {
		return nil, fmt.Errorf("failed to read POSITIVE_ASSET_4.json: %w", err)
	}

	negativeAsset1Prices, err := loadWsMessages("NEGATIVE_ASSET_1")
	if err != nil {
		return nil, fmt.Errorf("failed to read NEGATIVE_ASSET_1.json: %w", err)
	}

	positiveAsset1EncodedAssetID, err := pusher.HexStringToByte32(
		string(positiveAsset1Prices[0].StorkSignedPrice.EncodedAssetID),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to convert POSITIVE_ASSET_1 encoded asset ID to byte32: %w", err)
	}

	positiveAsset2EncodedAssetID, err := pusher.HexStringToByte32(
		string(positiveAsset2Prices[0].StorkSignedPrice.EncodedAssetID),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to convert POSITIVE_ASSET_2 encoded asset ID to byte32: %w", err)
	}

	positiveAsset3EncodedAssetID, err := pusher.HexStringToByte32(
		string(positiveAsset3Prices[0].StorkSignedPrice.EncodedAssetID),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to convert POSITIVE_ASSET_3 encoded asset ID to byte32: %w", err)
	}

	positiveAsset4EncodedAssetID, err := pusher.HexStringToByte32(
		string(positiveAsset4Prices[0].StorkSignedPrice.EncodedAssetID),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to convert POSITIVE_ASSET_4 encoded asset ID to byte32: %w", err)
	}

	negativeAsset1EncodedAssetID, err := pusher.HexStringToByte32(
		string(negativeAsset1Prices[0].StorkSignedPrice.EncodedAssetID),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to convert NEGATIVE_ASSET_1 encoded asset ID to byte32: %w", err)
	}

	return &SampleAggregatedSignedPrices{
		positiveAsset1:               positiveAsset1Prices,
		positiveAsset2:               positiveAsset2Prices,
		positiveAsset3:               positiveAsset3Prices,
		positiveAsset4:               positiveAsset4Prices,
		negativeAsset1:               negativeAsset1Prices,
		positiveAsset1EncodedAssetID: positiveAsset1EncodedAssetID,
		positiveAsset2EncodedAssetID: positiveAsset2EncodedAssetID,
		positiveAsset3EncodedAssetID: positiveAsset3EncodedAssetID,
		positiveAsset4EncodedAssetID: positiveAsset4EncodedAssetID,
		negativeAsset1EncodedAssetID: negativeAsset1EncodedAssetID,
	}, nil
}

func (s *SampleAggregatedSignedPrices) NextPositiveAsset1() (*types.AggregatedSignedPrice, error) {
	if len(s.positiveAsset1) == 0 {
		return nil, ErrNoMorePrices
	}

	price := s.positiveAsset1[0]
	s.positiveAsset1 = s.positiveAsset1[1:]

	return &price, nil
}

func (s *SampleAggregatedSignedPrices) NextPositiveAsset2() (*types.AggregatedSignedPrice, error) {
	if len(s.positiveAsset2) == 0 {
		return nil, ErrNoMorePrices
	}

	price := s.positiveAsset2[0]
	s.positiveAsset2 = s.positiveAsset2[1:]

	return &price, nil
}

func (s *SampleAggregatedSignedPrices) NextPositiveAsset3() (*types.AggregatedSignedPrice, error) {
	if len(s.positiveAsset3) == 0 {
		return nil, ErrNoMorePrices
	}

	price := s.positiveAsset3[0]
	s.positiveAsset3 = s.positiveAsset3[1:]

	return &price, nil
}

func (s *SampleAggregatedSignedPrices) NextPositiveAsset4() (*types.AggregatedSignedPrice, error) {
	if len(s.positiveAsset4) == 0 {
		return nil, ErrNoMorePrices
	}

	price := s.positiveAsset4[0]
	s.positiveAsset4 = s.positiveAsset4[1:]

	return &price, nil
}

func (s *SampleAggregatedSignedPrices) NextNegativeAsset1() (*types.AggregatedSignedPrice, error) {
	if len(s.negativeAsset1) == 0 {
		return nil, ErrNoMorePrices
	}

	price := s.negativeAsset1[0]
	s.negativeAsset1 = s.negativeAsset1[1:]

	return &price, nil
}

func (s *SampleAggregatedSignedPrices) PositiveAsset1EncodedAssetID() types.InternalEncodedAssetID {
	return s.positiveAsset1EncodedAssetID
}

func (s *SampleAggregatedSignedPrices) PositiveAsset2EncodedAssetID() types.InternalEncodedAssetID {
	return s.positiveAsset2EncodedAssetID
}

func (s *SampleAggregatedSignedPrices) PositiveAsset3EncodedAssetID() types.InternalEncodedAssetID {
	return s.positiveAsset3EncodedAssetID
}

func (s *SampleAggregatedSignedPrices) PositiveAsset4EncodedAssetID() types.InternalEncodedAssetID {
	return s.positiveAsset4EncodedAssetID
}

func (s *SampleAggregatedSignedPrices) NegativeAsset1EncodedAssetID() types.InternalEncodedAssetID {
	return s.negativeAsset1EncodedAssetID
}

func (s *SampleAggregatedSignedPrices) AllEncodedAssetIDs() []types.InternalEncodedAssetID {
	return []types.InternalEncodedAssetID{
		s.positiveAsset1EncodedAssetID,
		s.positiveAsset2EncodedAssetID,
		s.positiveAsset3EncodedAssetID,
		s.positiveAsset4EncodedAssetID,
		s.negativeAsset1EncodedAssetID,
	}
}

func loadWsMessages(assetID string) ([]types.AggregatedSignedPrice, error) {
	pc, thisFile, line, ok := runtime.Caller(0)
	if !ok {
		return nil, ErrFailedToGetCaller
	}

	_ = pc
	_ = line

	file, err := os.ReadFile(filepath.Join(filepath.Dir(thisFile), messagesDir, assetID+".json"))
	if err != nil {
		return nil, fmt.Errorf("failed to read %s.json: %w", assetID, err)
	}

	var messages []types.OraclePricesMessage

	err = json.Unmarshal(file, &messages)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal %s.json: %w", assetID, err)
	}

	if len(messages) == 0 {
		return nil, ErrNoPricesInFile
	}

	prices := make([]types.AggregatedSignedPrice, len(messages))

	var price types.AggregatedSignedPrice
	for i, message := range messages {
		price, ok = message.Data[assetID]
		if !ok {
			return nil, ErrFailedToGetPrices
		}

		prices[i] = price
	}

	return prices, nil
}
