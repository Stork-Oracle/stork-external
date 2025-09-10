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
	s.interactor, err = NewContractInteractor(s.config.RpcUrl, s.config.WsUrl, s.config.ContractAddress, []byte(s.config.PrivateKey), false, s.logger, 0)
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
func (s *InteractorTestSuite) Test_03_BatchPushToContract_and_PullValues_Single_Asset() {
	btcUsdEncodedAssetID := s.prices.BtcUsdEncodedAssetID()
	s.Require().NotNil(btcUsdEncodedAssetID)

	priceUpdates := s.getBtcUsdPriceUpdate()
	s.Require().NotNil(priceUpdates)

	// Push the BTCUSD price to the contract
	err := s.interactor.BatchPushToContract(priceUpdates)
	s.Require().NoError(err)

	// Pull the BTCUSD price from the contract
	values, err := s.interactor.PullValues([]types.InternalEncodedAssetID{btcUsdEncodedAssetID})
	s.Require().NoError(err)
	s.Require().NotNil(values)
	s.Require().Equal(1, len(values))

	// Check the quantized value and timestamp
	s.Require().Equal(string(priceUpdates[btcUsdEncodedAssetID].StorkSignedPrice.QuantizedPrice), values[btcUsdEncodedAssetID].QuantizedValue.String())
	s.Require().Equal(priceUpdates[btcUsdEncodedAssetID].StorkSignedPrice.TimestampedSignature.TimestampNano, values[btcUsdEncodedAssetID].TimestampNs)
	s.Require().NoError(err)
}

// Test_04_BatchPushToContract_and_PullValues_Multiple_Assets tests the behavior of batch pushing to the contract and pulling values from the contract.
func (s *InteractorTestSuite) Test_04_BatchPushToContract_and_PullValues_Multiple_Assets() {
	btcUsdEncodedAssetID := s.prices.BtcUsdEncodedAssetID()
	s.Require().NotNil(btcUsdEncodedAssetID)

	ethUsdEncodedAssetID := s.prices.EthUsdEncodedAssetID()
	s.Require().NotNil(ethUsdEncodedAssetID)

	solUsdEncodedAssetID := s.prices.SolUsdEncodedAssetID()
	s.Require().NotNil(solUsdEncodedAssetID)

	suiUsdEncodedAssetID := s.prices.SuiUsdEncodedAssetID()
	s.Require().NotNil(suiUsdEncodedAssetID)

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
	// BTCUSD
	s.Require().Equal(string(priceUpdates[btcUsdEncodedAssetID].StorkSignedPrice.QuantizedPrice), values[btcUsdEncodedAssetID].QuantizedValue.String())
	s.Require().Equal(priceUpdates[btcUsdEncodedAssetID].StorkSignedPrice.TimestampedSignature.TimestampNano, values[btcUsdEncodedAssetID].TimestampNs)

	// ETHUSD
	s.Require().Equal(string(priceUpdates[ethUsdEncodedAssetID].StorkSignedPrice.QuantizedPrice), values[ethUsdEncodedAssetID].QuantizedValue.String())
	s.Require().Equal(priceUpdates[ethUsdEncodedAssetID].StorkSignedPrice.TimestampedSignature.TimestampNano, values[ethUsdEncodedAssetID].TimestampNs)

	// SOLUSD
	s.Require().Equal(string(priceUpdates[solUsdEncodedAssetID].StorkSignedPrice.QuantizedPrice), values[solUsdEncodedAssetID].QuantizedValue.String())
	s.Require().Equal(priceUpdates[solUsdEncodedAssetID].StorkSignedPrice.TimestampedSignature.TimestampNano, values[solUsdEncodedAssetID].TimestampNs)

	// SUIUSD
	s.Require().Equal(string(priceUpdates[suiUsdEncodedAssetID].StorkSignedPrice.QuantizedPrice), values[suiUsdEncodedAssetID].QuantizedValue.String())
	s.Require().Equal(priceUpdates[suiUsdEncodedAssetID].StorkSignedPrice.TimestampedSignature.TimestampNano, values[suiUsdEncodedAssetID].TimestampNs)
	s.Require().NoError(err)
}

// Test_05_ListenContractEvents tests the behavior of listening for contract events.
func (s *InteractorTestSuite) Test_05_ListenContractEvents() {
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
	receivedBtcUsd := false
	receivedEthUsd := false
	receivedSolUsd := false
	receivedSuiUsd := false

	select {
	case update := <-ch:
		for encodedAssetID, value := range update {
			if encodedAssetID == s.prices.BtcUsdEncodedAssetID() {
				receivedBtcUsd = true
				s.Require().Equal(string(priceUpdates[encodedAssetID].StorkSignedPrice.QuantizedPrice), value.QuantizedValue.String())
				s.Require().Equal(priceUpdates[encodedAssetID].StorkSignedPrice.TimestampedSignature.TimestampNano, value.TimestampNs)
			}
			if encodedAssetID == s.prices.EthUsdEncodedAssetID() {
				receivedEthUsd = true
				s.Require().Equal(string(priceUpdates[encodedAssetID].StorkSignedPrice.QuantizedPrice), value.QuantizedValue.String())
				s.Require().Equal(priceUpdates[encodedAssetID].StorkSignedPrice.TimestampedSignature.TimestampNano, value.TimestampNs)
			}
			if encodedAssetID == s.prices.SolUsdEncodedAssetID() {
				receivedSolUsd = true
				s.Require().Equal(string(priceUpdates[encodedAssetID].StorkSignedPrice.QuantizedPrice), value.QuantizedValue.String())
				s.Require().Equal(priceUpdates[encodedAssetID].StorkSignedPrice.TimestampedSignature.TimestampNano, value.TimestampNs)
			}
			if encodedAssetID == s.prices.SuiUsdEncodedAssetID() {
				receivedSuiUsd = true
				s.Require().Equal(string(priceUpdates[encodedAssetID].StorkSignedPrice.QuantizedPrice), value.QuantizedValue.String())
				s.Require().Equal(priceUpdates[encodedAssetID].StorkSignedPrice.TimestampedSignature.TimestampNano, value.TimestampNs)
			}
			if receivedBtcUsd && receivedEthUsd && receivedSolUsd && receivedSuiUsd {
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

func (s *InteractorTestSuite) Test_06_GetWalletBalance_After_Push() {
	balance, err := s.interactor.GetWalletBalance()
	s.Require().NoError(err)
	s.Require().Less(balance, s.balance, "balance should be less than initial balance")
}

// Helper functions

func (s *InteractorTestSuite) getBtcUsdPriceUpdate() map[types.InternalEncodedAssetID]types.AggregatedSignedPrice {
	priceUpdates := make(map[types.InternalEncodedAssetID]types.AggregatedSignedPrice)

	btcUsdPrice, err := s.prices.NextBtcUsd()
	s.Require().NoError(err)
	s.Require().NotNil(btcUsdPrice)

	priceUpdates[s.prices.BtcUsdEncodedAssetID()] = *btcUsdPrice

	return priceUpdates
}

func (s *InteractorTestSuite) getEthUsdPriceUpdate() map[types.InternalEncodedAssetID]types.AggregatedSignedPrice {
	priceUpdates := make(map[types.InternalEncodedAssetID]types.AggregatedSignedPrice)

	ethUsdPrice, err := s.prices.NextEthUsd()
	s.Require().NoError(err)
	s.Require().NotNil(ethUsdPrice)

	priceUpdates[s.prices.EthUsdEncodedAssetID()] = *ethUsdPrice

	return priceUpdates
}

func (s *InteractorTestSuite) getSolUsdPriceUpdate() map[types.InternalEncodedAssetID]types.AggregatedSignedPrice {
	priceUpdates := make(map[types.InternalEncodedAssetID]types.AggregatedSignedPrice)

	solUsdPrice, err := s.prices.NextSolUsd()
	s.Require().NoError(err)
	s.Require().NotNil(solUsdPrice)

	priceUpdates[s.prices.SolUsdEncodedAssetID()] = *solUsdPrice

	return priceUpdates
}

func (s *InteractorTestSuite) getSuiUsdPriceUpdate() map[types.InternalEncodedAssetID]types.AggregatedSignedPrice {
	priceUpdates := make(map[types.InternalEncodedAssetID]types.AggregatedSignedPrice)

	suiUsdPrice, err := s.prices.NextSuiUsd()
	s.Require().NoError(err)
	s.Require().NotNil(suiUsdPrice)

	priceUpdates[s.prices.SuiUsdEncodedAssetID()] = *suiUsdPrice

	return priceUpdates
}

func (s *InteractorTestSuite) getAllPriceUpdates() map[types.InternalEncodedAssetID]types.AggregatedSignedPrice {
	priceUpdates := make(map[types.InternalEncodedAssetID]types.AggregatedSignedPrice)

	btcUsdPrice := s.getBtcUsdPriceUpdate()
	s.Require().NotNil(btcUsdPrice)

	ethUsdPrice := s.getEthUsdPriceUpdate()
	s.Require().NotNil(ethUsdPrice)

	solUsdPrice := s.getSolUsdPriceUpdate()
	s.Require().NotNil(solUsdPrice)

	suiUsdPrice := s.getSuiUsdPriceUpdate()
	s.Require().NotNil(suiUsdPrice)

	// Merge the price updates
	for encodedAssetID, priceUpdate := range btcUsdPrice {
		priceUpdates[encodedAssetID] = priceUpdate
	}
	for encodedAssetID, priceUpdate := range ethUsdPrice {
		priceUpdates[encodedAssetID] = priceUpdate
	}
	for encodedAssetID, priceUpdate := range solUsdPrice {
		priceUpdates[encodedAssetID] = priceUpdate
	}
	for encodedAssetID, priceUpdate := range suiUsdPrice {
		priceUpdates[encodedAssetID] = priceUpdate
	}

	return priceUpdates
}
