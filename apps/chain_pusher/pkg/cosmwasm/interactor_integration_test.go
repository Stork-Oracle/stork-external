//go:build integration

package cosmwasm

import (
	"context"
	"maps"
	"slices"
	"testing"
	"time"

	"github.com/Stork-Oracle/stork-external/apps/chain_pusher/internal/testutil"
	"github.com/Stork-Oracle/stork-external/apps/chain_pusher/pkg/cosmwasm/bindings"
	"github.com/Stork-Oracle/stork-external/apps/chain_pusher/pkg/pusher"
	"github.com/Stork-Oracle/stork-external/apps/chain_pusher/pkg/types"
	"github.com/caarlos0/env/v11"
	"github.com/stretchr/testify/suite"
)

const (
	// The local chain produces a block every 500ms, so pushes normally land well within this window.
	pushInclusionTimeout      = 30 * time.Second
	pushInclusionPollInterval = 250 * time.Millisecond
)

// InteractorTestConfig defaults match the local chain started by the cosmwasm-contract service in
// docker-compose.yml (see docker/scripts/cosmwasm-docker-entrypoint.sh).
type InteractorTestConfig struct {
	RpcUrl          string  `env:"COSMWASM_RPC_URL"          envDefault:"http://localhost:26657"`
	ContractAddress string  `env:"COSMWASM_CONTRACT_ADDRESS" envDefault:"wasm14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9s0phg4d"`
	Mnemonic        string  `env:"COSMWASM_MNEMONIC"         envDefault:"abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"`
	ChainID         string  `env:"COSMWASM_CHAIN_ID"         envDefault:"stork-local"`
	ChainPrefix     string  `env:"COSMWASM_CHAIN_PREFIX"     envDefault:"wasm"`
	Denom           string  `env:"COSMWASM_DENOM"            envDefault:"stake"`
	GasPrice        float64 `env:"COSMWASM_GAS_PRICE"        envDefault:"0.025"`
	GasAdjustment   float64 `env:"COSMWASM_GAS_ADJUSTMENT"   envDefault:"1.5"`
}

type InteractorTestSuite struct {
	suite.Suite

	config     InteractorTestConfig
	ctx        context.Context
	cancel     context.CancelFunc
	interactor *ContractInteractor
	prices     *testutil.SampleAggregatedSignedPrices
	balance    float64
}

func (s *InteractorTestSuite) SetupSuite() {
	s.Require().NoError(env.Parse(&s.config))
	s.ctx, s.cancel = context.WithCancel(context.Background())

	logger := PusherLogger(s.config.RpcUrl, s.config.ContractAddress)

	var err error

	s.interactor, err = NewContractInteractor(
		s.config.ContractAddress,
		[]byte(s.config.Mnemonic),
		logger,
		s.config.GasPrice,
		s.config.GasAdjustment,
		s.config.Denom,
		s.config.ChainID,
		s.config.ChainPrefix,
	)
	s.Require().NoError(err)

	// ConnectHTTP queries the contract's update fee, so the contract must already be instantiated.
	err = s.interactor.ConnectHTTP(s.ctx, s.config.RpcUrl)
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
	balance, err := s.interactor.GetWalletBalance(s.ctx)
	s.Require().NoError(err)
	s.Require().Greater(balance, 0.0, "balance should be greater than 0 for testing")

	s.balance = balance
}

// Test_02_PullValues_Initial tests pulling values before any prices are pushed. Unlike the EVM contract,
// the CosmWasm contract returns an error for feeds that have never been written, so PullValues returns
// an error alongside an empty map.
func (s *InteractorTestSuite) Test_02_PullValues_Initial() {
	values, err := s.interactor.PullValues(s.ctx, s.prices.AllEncodedAssetIDs())
	s.Require().ErrorIs(err, bindings.ErrQueryFailed)
	s.Require().Empty(values)
}

// Test_03_BatchPushToContract_and_PullValues_Single_Asset_Positive tests pushing a single positive price.
func (s *InteractorTestSuite) Test_03_BatchPushToContract_and_PullValues_Single_Asset_Positive() {
	priceUpdates := s.nextPriceUpdates(s.prices.NextPositiveAsset1)

	s.pushAndWaitForInclusion(priceUpdates)
	s.requireContractValues(priceUpdates)
}

// Test_04_BatchPushToContract_and_PullValues_Single_Asset_Negative tests pushing a single negative price.
func (s *InteractorTestSuite) Test_04_BatchPushToContract_and_PullValues_Single_Asset_Negative() {
	priceUpdates := s.nextPriceUpdates(s.prices.NextNegativeAsset1)

	s.pushAndWaitForInclusion(priceUpdates)
	s.requireContractValues(priceUpdates)
}

// Test_05_BatchPushToContract_and_PullValues_Multiple_Assets_Positive_Negative tests pushing positive and
// negative prices for several assets in one batch.
func (s *InteractorTestSuite) Test_05_BatchPushToContract_and_PullValues_Multiple_Assets_Positive_Negative() {
	priceUpdates := s.allNextPriceUpdates()

	s.pushAndWaitForInclusion(priceUpdates)
	s.requireContractValues(priceUpdates)
}

// Test_06_BatchPushToContract_Invalid_Signature tests that a push the contract rejects returns an error and
// leaves the stored value unchanged. Gas estimation simulates the transaction, which runs the contract, so the
// rejection is returned before anything is broadcast.
func (s *InteractorTestSuite) Test_06_BatchPushToContract_Invalid_Signature() {
	encodedAssetID := s.prices.PositiveAsset1EncodedAssetID()

	valuesBefore, err := s.interactor.PullValues(s.ctx, []types.InternalEncodedAssetID{encodedAssetID})
	s.Require().NoError(err)

	priceUpdates := s.nextPriceUpdates(s.prices.NextPositiveAsset1)

	// Changing the price invalidates the signature over it.
	tamperedSignedPrice := *priceUpdates[encodedAssetID].StorkSignedPrice
	tamperedSignedPrice.QuantizedPrice += "0"

	tamperedUpdate := priceUpdates[encodedAssetID]
	tamperedUpdate.StorkSignedPrice = &tamperedSignedPrice
	priceUpdates[encodedAssetID] = tamperedUpdate

	err = s.interactor.BatchPushToContract(s.ctx, priceUpdates)
	s.Require().ErrorContains(err, "Invalid signature")

	valuesAfter, err := s.interactor.PullValues(s.ctx, []types.InternalEncodedAssetID{encodedAssetID})
	s.Require().NoError(err)
	s.Require().Equal(valuesBefore, valuesAfter)
}

// Test_07_GetWalletBalance_After_Push tests the behavior of getting the wallet balance after pushing to the contract.
// As this test runs after the pushes, it compares against the initial balance stored in
// Test_01_GetWalletBalance_Initial.
func (s *InteractorTestSuite) Test_07_GetWalletBalance_After_Push() {
	balance, err := s.interactor.GetWalletBalance(s.ctx)
	s.Require().NoError(err)
	s.Require().Less(balance, s.balance, "balance should be less than initial balance")
}

// Test_08_PullValues_WithTimeout_ContextDeadlineExceeded tests that PullValues returns a context deadline
// exceeded error when called with a very short timeout.
func (s *InteractorTestSuite) Test_08_PullValues_WithTimeout_ContextDeadlineExceeded() {
	ctx, cancel := context.WithTimeout(s.ctx, 1*time.Nanosecond)
	defer cancel()

	// Make sure it times out
	time.Sleep(1 * time.Millisecond)

	_, err := s.interactor.PullValues(ctx, s.prices.AllEncodedAssetIDs())
	s.Require().ErrorIs(err, context.DeadlineExceeded)
}

// Test_09_BatchPushToContract_WithTimeout_ContextDeadlineExceeded tests that BatchPushToContract returns a
// context deadline exceeded error when called with a very short timeout. The deadline is hit by the account
// query that runs first; the Cosmos SDK's gas simulation and broadcast calls don't take a context.
func (s *InteractorTestSuite) Test_09_BatchPushToContract_WithTimeout_ContextDeadlineExceeded() {
	priceUpdates := s.allNextPriceUpdates()

	ctx, cancel := context.WithTimeout(s.ctx, 1*time.Nanosecond)
	defer cancel()

	// Make sure it times out
	time.Sleep(1 * time.Millisecond)

	err := s.interactor.BatchPushToContract(ctx, priceUpdates)
	s.Require().ErrorIs(err, context.DeadlineExceeded)
}

// Test_10_GetWalletBalance_WithTimeout_ContextDeadlineExceeded tests that GetWalletBalance returns a context
// deadline exceeded error when called with a very short timeout.
func (s *InteractorTestSuite) Test_10_GetWalletBalance_WithTimeout_ContextDeadlineExceeded() {
	ctx, cancel := context.WithTimeout(s.ctx, 1*time.Nanosecond)
	defer cancel()

	// Make sure it times out
	time.Sleep(1 * time.Millisecond)

	balance, err := s.interactor.GetWalletBalance(ctx)
	s.Require().ErrorIs(err, context.DeadlineExceeded)
	s.Require().Less(balance, 0.0, "balance should be invalid when there's an error")
}

// Helper functions

// nextPriceUpdates takes the next sample price from each of the given sources, keyed by encoded asset ID.
func (s *InteractorTestSuite) nextPriceUpdates(
	nextPrices ...func() (*types.AggregatedSignedPrice, error),
) map[types.InternalEncodedAssetID]types.AggregatedSignedPrice {
	priceUpdates := make(map[types.InternalEncodedAssetID]types.AggregatedSignedPrice, len(nextPrices))

	for _, nextPrice := range nextPrices {
		price, err := nextPrice()
		s.Require().NoError(err)

		encodedAssetID, err := pusher.HexStringToByte32(string(price.StorkSignedPrice.EncodedAssetID))
		s.Require().NoError(err)

		priceUpdates[encodedAssetID] = *price
	}

	return priceUpdates
}

// allNextPriceUpdates takes the next sample price for every asset.
func (s *InteractorTestSuite) allNextPriceUpdates() map[types.InternalEncodedAssetID]types.AggregatedSignedPrice {
	return s.nextPriceUpdates(
		s.prices.NextPositiveAsset1,
		s.prices.NextPositiveAsset2,
		s.prices.NextPositiveAsset3,
		s.prices.NextPositiveAsset4,
		s.prices.NextNegativeAsset1,
	)
}

// pushAndWaitForInclusion pushes the updates and waits until the contract returns them. BatchPushToContract
// broadcasts in sync mode, which returns once the transaction passes CheckTx rather than once it is in a block.
// The next push reads the account sequence from committed state, so pushing again before this transaction is
// included would be rejected for a sequence mismatch.
func (s *InteractorTestSuite) pushAndWaitForInclusion(
	priceUpdates map[types.InternalEncodedAssetID]types.AggregatedSignedPrice,
) {
	err := s.interactor.BatchPushToContract(s.ctx, priceUpdates)
	s.Require().NoError(err)

	encodedAssetIDs := slices.Collect(maps.Keys(priceUpdates))

	s.Require().Eventually(func() bool {
		values, pullErr := s.interactor.PullValues(s.ctx, encodedAssetIDs)
		if pullErr != nil {
			return false
		}

		for encodedAssetID, priceUpdate := range priceUpdates {
			if values[encodedAssetID].TimestampNs != priceUpdate.StorkSignedPrice.TimestampedSignature.TimestampNano {
				return false
			}
		}

		return true
	}, pushInclusionTimeout, pushInclusionPollInterval, "pushed values were not found on the contract")
}

// requireContractValues checks that the contract's latest values match the given updates.
func (s *InteractorTestSuite) requireContractValues(
	priceUpdates map[types.InternalEncodedAssetID]types.AggregatedSignedPrice,
) {
	values, err := s.interactor.PullValues(s.ctx, slices.Collect(maps.Keys(priceUpdates)))
	s.Require().NoError(err)
	s.Require().Len(values, len(priceUpdates))

	for encodedAssetID, priceUpdate := range priceUpdates {
		s.Require().
			Equal(string(priceUpdate.StorkSignedPrice.QuantizedPrice), values[encodedAssetID].QuantizedValue.String())
		s.Require().
			Equal(priceUpdate.StorkSignedPrice.TimestampedSignature.TimestampNano, values[encodedAssetID].TimestampNs)
	}
}
