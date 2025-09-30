//go:build integration

package evm

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/Stork-Oracle/stork-external/apps/chain_pusher/internal/testutil"
	"github.com/Stork-Oracle/stork-external/apps/chain_pusher/pkg/types"
	"github.com/caarlos0/env/v11"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/suite"
)

type InteractorTestConfig struct {
	RpcUrl          string `env:"EVM_RPC_URL" envDefault:"http://localhost:8545"`
	WsUrl           string `env:"EVM_WS_URL" envDefault:"ws://localhost:8545"`
	ContractAddress string `env:"EVM_CONTRACT_ADDRESS" envDefault:"0xe7f1725e7734ce288f8367e1bb143e90bb3f0512"`
	PrivateKey      string `env:"EVM_PRIVATE_KEY" envDefault:"ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"`
}
type InteractorTestSuite struct {
	suite.Suite
	config     InteractorTestConfig
	ctx        context.Context
	cancel     context.CancelFunc
	interactor *ContractInteractor
	logger     zerolog.Logger
	prices     *testutil.SampleAggregatedSignedPrices
	balance    float64
}

func (s *InteractorTestSuite) SetupSuite() {
	s.Require().NoError(env.Parse(&s.config))
	s.ctx, s.cancel = context.WithCancel(context.Background())

	fmt.Println("RpcUrl: ", s.config.RpcUrl)
	fmt.Println("WsUrl: ", s.config.WsUrl)
	fmt.Println("ContractAddress: ", s.config.ContractAddress)
	fmt.Println("PrivateKey: ", s.config.PrivateKey)

	s.logger = PusherLogger(s.config.RpcUrl, s.config.ContractAddress)

	var err error
	s.interactor, err = NewContractInteractor(s.config.ContractAddress, []byte(s.config.PrivateKey), false, s.logger, 0)
	s.Require().NoError(err)

	s.prices, err = testutil.LoadAggregatedSignedPrices()
	s.Require().NoError(err)
}

func (s *InteractorTestSuite) TearDownSuite() {
	s.cancel()
}

func TestInteractorTestSuite(t *testing.T) {
	suite.Run(t, new(InteractorTestSuite))
}

// Test_01_GetWalletBalance_Initial tests the initial balance of the wallet before any prices are pushed.
func (s *InteractorTestSuite) Test_01_GetWalletBalance_Initial() {
	balance, err := s.interactor.GetWalletBalance()
	s.Require().NoError(err)
	s.Require().Greater(balance, 0.0, "balance should be greater than 0 for testing")

	s.balance = balance
}

// Test_02_PullValues_Initial tests the behavior of pulling values from the contract before any prices are pushed.
func (s *InteractorTestSuite) Test_02_PullValues_Initial() {
	values, err := s.interactor.PullValues(s.prices.AllEncodedAssetIDs())
	s.Require().NoError(err)
	s.Require().NotNil(values)
	s.Require().Equal(0, len(values))
}

// Test_03_BatchPushToContract tests the behavior of batch pushing to the contract.
func (s *InteractorTestSuite) Test_03_BatchPushToContract_and_PullValues_Single_Asset_Positive() {
	positiveAsset1EncodedAssetID := s.prices.PositiveAsset1EncodedAssetID()
	s.Require().NotNil(positiveAsset1EncodedAssetID)

	priceUpdates := s.getPositiveAsset1PriceUpdate()
	s.Require().NotNil(priceUpdates)

	// Push the POSITIVE_ASSET_1 price to the contract
	err := s.interactor.BatchPushToContract(priceUpdates)
	s.Require().NoError(err)

	// Pull the POSITIVE_ASSET_1 price from the contract
	values, err := s.interactor.PullValues([]types.InternalEncodedAssetID{positiveAsset1EncodedAssetID})
	s.Require().NoError(err)
	s.Require().NotNil(values)
	s.Require().Equal(1, len(values))

	// Check the quantized value and timestamp
	s.Require().Equal(string(priceUpdates[positiveAsset1EncodedAssetID].StorkSignedPrice.QuantizedPrice), values[positiveAsset1EncodedAssetID].QuantizedValue.String())
	s.Require().Equal(priceUpdates[positiveAsset1EncodedAssetID].StorkSignedPrice.TimestampedSignature.TimestampNano, values[positiveAsset1EncodedAssetID].TimestampNs)
	s.Require().NoError(err)
}

func (s *InteractorTestSuite) Test_04_BatchPushToContract_and_PullValues_Single_Asset_Negative() {
	negativeAsset1EncodedAssetID := s.prices.NegativeAsset1EncodedAssetID()
	s.Require().NotNil(negativeAsset1EncodedAssetID)

	priceUpdates := s.getNegativeAsset1PriceUpdate()
	s.Require().NotNil(priceUpdates)

	// Push the prices to the contract
	err := s.interactor.BatchPushToContract(priceUpdates)
	s.Require().NoError(err)

	// Pull the prices from the contract
	values, err := s.interactor.PullValues([]types.InternalEncodedAssetID{negativeAsset1EncodedAssetID})
	s.Require().NoError(err)
	s.Require().NotNil(values)
	s.Require().Equal(1, len(values))

	// Check the quantized value and timestamp
	s.Require().Equal(string(priceUpdates[negativeAsset1EncodedAssetID].StorkSignedPrice.QuantizedPrice), values[negativeAsset1EncodedAssetID].QuantizedValue.String())
	s.Require().Equal(priceUpdates[negativeAsset1EncodedAssetID].StorkSignedPrice.TimestampedSignature.TimestampNano, values[negativeAsset1EncodedAssetID].TimestampNs)
	s.Require().NoError(err)
}

// Test_05_BatchPushToContract_and_PullValues_Multiple_Assets tests the behavior of batch pushing to the contract and pulling values from the contract.
func (s *InteractorTestSuite) Test_05_BatchPushToContract_and_PullValues_Multiple_Assets_Positive_Negative() {
	positiveAsset1EncodedAssetID := s.prices.PositiveAsset1EncodedAssetID()
	s.Require().NotNil(positiveAsset1EncodedAssetID)

	positiveAsset2EncodedAssetID := s.prices.PositiveAsset2EncodedAssetID()
	s.Require().NotNil(positiveAsset2EncodedAssetID)

	positiveAsset3EncodedAssetID := s.prices.PositiveAsset3EncodedAssetID()
	s.Require().NotNil(positiveAsset3EncodedAssetID)

	positiveAsset4EncodedAssetID := s.prices.PositiveAsset4EncodedAssetID()
	s.Require().NotNil(positiveAsset4EncodedAssetID)

	priceUpdates := s.getAllPriceUpdates()
	s.Require().NotNil(priceUpdates)

	// Push the prices to the contract
	err := s.interactor.BatchPushToContract(priceUpdates)
	s.Require().NoError(err)

	// Pull the prices from the contract
	values, err := s.interactor.PullValues(s.prices.AllEncodedAssetIDs())
	s.Require().NoError(err)
	s.Require().NotNil(values)
	s.Require().Equal(len(values), len(s.prices.AllEncodedAssetIDs()))
	s.Require().NoError(err)

	// Check the quantized value and timestamps
	// POSITIVE_ASSET_1
	s.Require().Equal(string(priceUpdates[positiveAsset1EncodedAssetID].StorkSignedPrice.QuantizedPrice), values[positiveAsset1EncodedAssetID].QuantizedValue.String())
	s.Require().Equal(priceUpdates[positiveAsset1EncodedAssetID].StorkSignedPrice.TimestampedSignature.TimestampNano, values[positiveAsset1EncodedAssetID].TimestampNs)

	// POSITIVE_ASSET_2
	s.Require().Equal(string(priceUpdates[positiveAsset2EncodedAssetID].StorkSignedPrice.QuantizedPrice), values[positiveAsset2EncodedAssetID].QuantizedValue.String())
	s.Require().Equal(priceUpdates[positiveAsset2EncodedAssetID].StorkSignedPrice.TimestampedSignature.TimestampNano, values[positiveAsset2EncodedAssetID].TimestampNs)

	// POSITIVE_ASSET_3
	s.Require().Equal(string(priceUpdates[positiveAsset3EncodedAssetID].StorkSignedPrice.QuantizedPrice), values[positiveAsset3EncodedAssetID].QuantizedValue.String())
	s.Require().Equal(priceUpdates[positiveAsset3EncodedAssetID].StorkSignedPrice.TimestampedSignature.TimestampNano, values[positiveAsset3EncodedAssetID].TimestampNs)

	// POSITIVE_ASSET_4
	s.Require().Equal(string(priceUpdates[positiveAsset4EncodedAssetID].StorkSignedPrice.QuantizedPrice), values[positiveAsset4EncodedAssetID].QuantizedValue.String())
	s.Require().Equal(priceUpdates[positiveAsset4EncodedAssetID].StorkSignedPrice.TimestampedSignature.TimestampNano, values[positiveAsset4EncodedAssetID].TimestampNs)
	s.Require().NoError(err)
}

// Test_06_ListenContractEvents tests the behavior of listening for contract events.
func (s *InteractorTestSuite) Test_06_ListenContractEvents() {
	ch := make(chan map[types.InternalEncodedAssetID]types.InternalTemporalNumericValue)

	listenCtx, listenCtxCancel := context.WithCancel(s.ctx)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		s.interactor.ListenContractEvents(listenCtx, ch)
	}()

	priceUpdates := s.getAllPriceUpdates()
	s.Require().NotNil(priceUpdates)

	err := s.interactor.BatchPushToContract(priceUpdates)
	s.Require().NoError(err)

	// Bool flags for each asset
	receivedPositiveAsset1 := false
	receivedPositiveAsset2 := false
	receivedPositiveAsset3 := false
	receivedPositiveAsset4 := false

	select {
	case update := <-ch:
		for encodedAssetID, value := range update {
			if encodedAssetID == s.prices.PositiveAsset1EncodedAssetID() {
				receivedPositiveAsset1 = true
				s.Require().Equal(string(priceUpdates[encodedAssetID].StorkSignedPrice.QuantizedPrice), value.QuantizedValue.String())
				s.Require().Equal(priceUpdates[encodedAssetID].StorkSignedPrice.TimestampedSignature.TimestampNano, value.TimestampNs)
			}
			if encodedAssetID == s.prices.PositiveAsset2EncodedAssetID() {
				receivedPositiveAsset2 = true
				s.Require().Equal(string(priceUpdates[encodedAssetID].StorkSignedPrice.QuantizedPrice), value.QuantizedValue.String())
				s.Require().Equal(priceUpdates[encodedAssetID].StorkSignedPrice.TimestampedSignature.TimestampNano, value.TimestampNs)
			}
			if encodedAssetID == s.prices.PositiveAsset3EncodedAssetID() {
				receivedPositiveAsset3 = true
				s.Require().Equal(string(priceUpdates[encodedAssetID].StorkSignedPrice.QuantizedPrice), value.QuantizedValue.String())
				s.Require().Equal(priceUpdates[encodedAssetID].StorkSignedPrice.TimestampedSignature.TimestampNano, value.TimestampNs)
			}
			if encodedAssetID == s.prices.PositiveAsset4EncodedAssetID() {
				receivedPositiveAsset4 = true
				s.Require().Equal(string(priceUpdates[encodedAssetID].StorkSignedPrice.QuantizedPrice), value.QuantizedValue.String())
				s.Require().Equal(priceUpdates[encodedAssetID].StorkSignedPrice.TimestampedSignature.TimestampNano, value.TimestampNs)
			}
			if receivedPositiveAsset1 && receivedPositiveAsset2 && receivedPositiveAsset3 && receivedPositiveAsset4 {
				break
			}
		}
	case <-time.After(5 * time.Second):
		s.Require().Fail("test timed out after 5 seconds, should have received all values")
	}
	listenCtxCancel()
	wg.Wait()
	close(ch)
}

// Test_07_GetWalletBalance_After_Push tests the behavior of getting the wallet balance after pushing to the contract.
// As this test runs last, we don't have to push, as we stored the initial balance in Test_01_GetWalletBalance_Initial,
// and have pushed in the previous tests.
func (s *InteractorTestSuite) Test_07_GetWalletBalance_After_Push() {
	balance, err := s.interactor.GetWalletBalance()
	s.Require().NoError(err)
	s.Require().Less(balance, s.balance, "balance should be less than initial balance")
}

// Helper functions

func (s *InteractorTestSuite) getPositiveAsset1PriceUpdate() map[types.InternalEncodedAssetID]types.AggregatedSignedPrice {
	priceUpdates := make(map[types.InternalEncodedAssetID]types.AggregatedSignedPrice)

	positiveAsset1Price, err := s.prices.NextPositiveAsset1()
	s.Require().NoError(err)
	s.Require().NotNil(positiveAsset1Price)

	priceUpdates[s.prices.PositiveAsset1EncodedAssetID()] = *positiveAsset1Price

	return priceUpdates
}

func (s *InteractorTestSuite) getPositiveAsset2PriceUpdate() map[types.InternalEncodedAssetID]types.AggregatedSignedPrice {
	priceUpdates := make(map[types.InternalEncodedAssetID]types.AggregatedSignedPrice)

	positiveAsset2Price, err := s.prices.NextPositiveAsset2()
	s.Require().NoError(err)
	s.Require().NotNil(positiveAsset2Price)

	priceUpdates[s.prices.PositiveAsset2EncodedAssetID()] = *positiveAsset2Price

	return priceUpdates
}

func (s *InteractorTestSuite) getPositiveAsset3PriceUpdate() map[types.InternalEncodedAssetID]types.AggregatedSignedPrice {
	priceUpdates := make(map[types.InternalEncodedAssetID]types.AggregatedSignedPrice)

	positiveAsset3Price, err := s.prices.NextPositiveAsset3()
	s.Require().NoError(err)
	s.Require().NotNil(positiveAsset3Price)

	priceUpdates[s.prices.PositiveAsset3EncodedAssetID()] = *positiveAsset3Price

	return priceUpdates
}

func (s *InteractorTestSuite) getPositiveAsset4PriceUpdate() map[types.InternalEncodedAssetID]types.AggregatedSignedPrice {
	priceUpdates := make(map[types.InternalEncodedAssetID]types.AggregatedSignedPrice)

	positiveAsset4Price, err := s.prices.NextPositiveAsset4()
	s.Require().NoError(err)
	s.Require().NotNil(positiveAsset4Price)

	priceUpdates[s.prices.PositiveAsset4EncodedAssetID()] = *positiveAsset4Price

	return priceUpdates
}

func (s *InteractorTestSuite) getNegativeAsset1PriceUpdate() map[types.InternalEncodedAssetID]types.AggregatedSignedPrice {
	priceUpdates := make(map[types.InternalEncodedAssetID]types.AggregatedSignedPrice)

	negativeAsset1Price, err := s.prices.NextNegativeAsset1()
	s.Require().NoError(err)
	s.Require().NotNil(negativeAsset1Price)

	priceUpdates[s.prices.NegativeAsset1EncodedAssetID()] = *negativeAsset1Price

	return priceUpdates
}

func (s *InteractorTestSuite) getAllPriceUpdates() map[types.InternalEncodedAssetID]types.AggregatedSignedPrice {
	priceUpdates := make(map[types.InternalEncodedAssetID]types.AggregatedSignedPrice)

	positiveAsset1Price := s.getPositiveAsset1PriceUpdate()
	s.Require().NotNil(positiveAsset1Price)

	positiveAsset2Price := s.getPositiveAsset2PriceUpdate()
	s.Require().NotNil(positiveAsset2Price)

	positiveAsset3Price := s.getPositiveAsset3PriceUpdate()
	s.Require().NotNil(positiveAsset3Price)

	positiveAsset4Price := s.getPositiveAsset4PriceUpdate()
	s.Require().NotNil(positiveAsset4Price)

	negativeAsset1Price := s.getNegativeAsset1PriceUpdate()
	s.Require().NotNil(negativeAsset1Price)

	// Merge the price updates
	for encodedAssetID, priceUpdate := range positiveAsset1Price {
		priceUpdates[encodedAssetID] = priceUpdate
	}
	for encodedAssetID, priceUpdate := range positiveAsset2Price {
		priceUpdates[encodedAssetID] = priceUpdate
	}
	for encodedAssetID, priceUpdate := range positiveAsset3Price {
		priceUpdates[encodedAssetID] = priceUpdate
	}
	for encodedAssetID, priceUpdate := range positiveAsset4Price {
		priceUpdates[encodedAssetID] = priceUpdate
	}
	for encodedAssetID, priceUpdate := range negativeAsset1Price {
		priceUpdates[encodedAssetID] = priceUpdate
	}
	return priceUpdates
}
