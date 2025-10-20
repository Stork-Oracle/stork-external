//go:build integration

package first_party_evm

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/Stork-Oracle/stork-external/apps/first_party_pusher/pkg/types"
	publisher_agent "github.com/Stork-Oracle/stork-external/apps/publisher_agent/pkg"
	"github.com/Stork-Oracle/stork-external/shared"
	"github.com/caarlos0/env/v11"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/suite"
)

type FirstPartyInteractorTestConfig struct {
	RpcUrl          string `env:"RPC_URL" envDefault:"http://localhost:8545"`
	WsUrl           string `env:"WS_URL" envDefault:"ws://localhost:8545"`
	ContractAddress string `env:"CONTRACT_ADDRESS" envDefault:"0xe7f1725e7734ce288f8367e1bb143e90bb3f0512"`
	TestPublicKey   string `env:"PUBLISHER_EVM_PUBLIC_KEY" envDefault:"0x99e295e85cb07C16B7BB62A44dF532A7F2620237"` // This must be the same as what is used in the evm-contract setup
	TestPrivateKey  string `env:"PRIVATE_KEY" envDefault:"0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"`
}

type FirstPartyInteractorTestSuite struct {
	suite.Suite
	config     FirstPartyInteractorTestConfig
	ctx        context.Context
	cancel     context.CancelFunc
	interactor *ContractInteractor
	logger     zerolog.Logger
	privateKey *ecdsa.PrivateKey
	pubKey     common.Address
}

func (s *FirstPartyInteractorTestSuite) SetupSuite() {
	s.Require().NoError(env.Parse(&s.config))
	s.ctx, s.cancel = context.WithCancel(context.Background())

	fmt.Println("RpcUrl: ", s.config.RpcUrl)
	fmt.Println("WsUrl: ", s.config.WsUrl)
	fmt.Println("ContractAddress: ", s.config.ContractAddress)
	fmt.Println("TestPrivateKey: ", s.config.TestPrivateKey)
	fmt.Println("TestPublicKey: ", s.config.TestPublicKey)

	s.logger = zerolog.New(zerolog.NewConsoleWriter()).With().
		Str("component", "first_party_interactor_test").
		Timestamp().
		Logger()

	var err error
	s.config.TestPrivateKey = strings.TrimPrefix(s.config.TestPrivateKey, "0x")
	s.privateKey, err = crypto.HexToECDSA(s.config.TestPrivateKey)
	s.Require().NoError(err)

	s.pubKey = common.HexToAddress(s.config.TestPublicKey)

	s.interactor, err = NewContractInteractor(
		s.config.RpcUrl,
		s.config.WsUrl,
		s.config.ContractAddress,
		s.privateKey,
		0,
		s.logger,
	)
	s.Require().NoError(err)
}

func (s *FirstPartyInteractorTestSuite) TearDownSuite() {
	s.cancel()
	if s.interactor != nil {
		s.interactor.Close()
	}
}

func TestFirstPartyInteractorTestSuite(t *testing.T) {
	suite.Run(t, new(FirstPartyInteractorTestSuite))
}

// Test_01_CheckPublisherUser tests checking if a publisher user exists.
func (s *FirstPartyInteractorTestSuite) Test_01_CheckPublisherUser() {
	// Publisher is expected to be registered in the contract setup.
	exists, err := s.interactor.CheckPublisherUser(s.pubKey)
	s.Require().NoError(err)
	s.logger.Info().Bool("exists", exists).Str("pubKey", s.pubKey.Hex()).Msg("Publisher user check")
}

// Test_02_PullValues_Initial tests pulling values before any updates
func (s *FirstPartyInteractorTestSuite) Test_02_PullValues_Initial() {
	pubKeyAssetIDPairs := map[common.Address][]shared.AssetID{
		s.pubKey: {"ETHUSD", "BTCUSD"},
	}

	values, err := s.interactor.PullValues(pubKeyAssetIDPairs)
	s.Require().NoError(err) // first party evm interactor does not return error in current implementation
	s.Require().Equal(len(values), len(pubKeyAssetIDPairs))
	s.Require().Empty(values[0].ContractValueMap) // contract is empty
}

// Test_03_BatchPushToContract_Single tests pushing a single update.
func (s *FirstPartyInteractorTestSuite) Test_03_BatchPushToContract_Single() {
	update := map[types.AssetEntry]publisher_agent.SignedPriceUpdate[*shared.EvmSignature]{
		{
			AssetID:        "ETHUSD",
			PublicKey:      shared.PublisherKey(s.pubKey.Hex()),
			Historical:     false,
		}: {
			OracleID: "local",
			AssetID:  "ETHUSD",
			Trigger:  publisher_agent.ClockTriggerType,
			SignedPrice: publisher_agent.SignedPrice[*shared.EvmSignature]{
				PublisherKey:    shared.PublisherKey(s.pubKey.Hex()),
				ExternalAssetID: "ETHUSD",
				SignatureType:   shared.EvmSignatureType,
				QuantizedPrice:  "1000000000000000000",
				TimestampedSignature: shared.TimestampedSignature[*shared.EvmSignature]{
					TimestampNano: 1680210934000000000,
					Signature: &shared.EvmSignature{
						R: "0x12068d27663c3139ee274ea287cc6193f6c4f6ccacac5f7c5d55604f6326c8f9",
						S: "0x2d8718269cf35cfa73db573ffabdfaa2c8a370c4b814cf47d6994f60599244ea",
						V: "0x1b",
					},
				},
			},
		},
	}

	err := s.interactor.BatchPushToContract(update)
	s.Require().NoError(err)

	time.Sleep(1 * time.Second)

	pubKeyAssetIDPairs := map[common.Address][]shared.AssetID{
		s.pubKey: {"ETHUSD"},
	}

	values, err := s.interactor.PullValues(pubKeyAssetIDPairs)
	s.Require().NoError(err)
	s.Require().NotEmpty(values)
	s.Require().Equal(1, len(values))
	s.Require().Equal(1, len(values[0].ContractValueMap))

	value, exists := values[0].ContractValueMap["ETHUSD"]
	s.Require().True(exists, "value should exist")
	s.Require().Equal("1000000000000000000", value.QuantizedValue.String())
	s.Require().Equal(uint64(1680210934000000000), value.TimestampNs)
}

// Test_04_BatchPushToContract_Multiple tests pushing multiple updates.
func (s *FirstPartyInteractorTestSuite) Test_04_BatchPushToContract_Multiple() {
	update := map[types.AssetEntry]publisher_agent.SignedPriceUpdate[*shared.EvmSignature]{
		{
			AssetID:        "ETHUSD",
			PublicKey:      shared.PublisherKey(s.pubKey.Hex()),
			Historical:     false,
		}: {
			OracleID: "local",
			AssetID:  "ETHUSD",
			Trigger:  publisher_agent.ClockTriggerType,
			SignedPrice: publisher_agent.SignedPrice[*shared.EvmSignature]{
				PublisherKey:    shared.PublisherKey(s.pubKey.Hex()),
				ExternalAssetID: "ETHUSD",
				SignatureType:   shared.EvmSignatureType,
				QuantizedPrice:  "1100000000000000000",
				TimestampedSignature: shared.TimestampedSignature[*shared.EvmSignature]{
					TimestampNano: 1680210935000000000,
					Signature: &shared.EvmSignature{
						R: "0xf8b94614c595a0fe989a377787c3ec257fd9fdfd1b1c171b21b550745d8c1952",
						S: "0x42e7576183888271fa19bb3b526b915b789578e2132529dd891b1928f11c9d25",
						V: "0x1c",
					},
				},
			},
		},
		{
			AssetID:        "BTCUSD",
			PublicKey:      shared.PublisherKey(s.pubKey.Hex()),
			Historical:     false,
		}: {
			OracleID: "local",
			AssetID:  "BTCUSD",
			Trigger:  publisher_agent.ClockTriggerType,
			SignedPrice: publisher_agent.SignedPrice[*shared.EvmSignature]{
				PublisherKey:    shared.PublisherKey(s.pubKey.Hex()),
				ExternalAssetID: "BTCUSD",
				SignatureType:   shared.EvmSignatureType,
				QuantizedPrice:  "2000000000000000000",
				TimestampedSignature: shared.TimestampedSignature[*shared.EvmSignature]{
					TimestampNano: 1680210935000000000,
					Signature: &shared.EvmSignature{
						R: "0xecb9a4c1200cd3d4f46d5e5d8e3e675e73459c99ea853f85a3819ded85970d81",
						S: "0x33988bea5a1c51d5f2f60173ec153a28552bd86a1c7e8c712b77baec5d472e31",
						V: "0x1c",
					},
				},
			},
		},
	}

	err := s.interactor.BatchPushToContract(update)
	s.Require().NoError(err)

	time.Sleep(1 * time.Second)

	pubKeyAssetIDPairs := map[common.Address][]shared.AssetID{
		s.pubKey: {"ETHUSD", "BTCUSD"},
	}

	values, err := s.interactor.PullValues(pubKeyAssetIDPairs)
	s.Require().NoError(err)
	s.Require().NotEmpty(values)
	s.Require().Equal(1, len(values))
	s.Require().Equal(2, len(values[0].ContractValueMap))

	value1, exists := values[0].ContractValueMap["ETHUSD"]
	s.Require().True(exists, "ETHUSD value should exist")
	s.Require().Equal("1100000000000000000", value1.QuantizedValue.String())
	s.Require().Equal(uint64(1680210935000000000), value1.TimestampNs)

	value2, exists := values[0].ContractValueMap["BTCUSD"]
	s.Require().True(exists, "BTCUSD value should exist")
	s.Require().Equal("2000000000000000000", value2.QuantizedValue.String())
	s.Require().Equal(uint64(1680210935000000000), value2.TimestampNs)
}

// Test_05_ListenContractEvents tests listening for contract events.
func (s *FirstPartyInteractorTestSuite) Test_05_ListenContractEvents() {
	ch := make(chan types.ContractUpdate)
	defer close(ch)

	listenCtx, listenCtxCancel := context.WithCancel(s.ctx)
	defer listenCtxCancel()

	pubKeyAssetIDPairs := map[common.Address][]shared.AssetID{
		s.pubKey: {"ETHUSD", "BTCUSD"},
	}

	go func() {
		s.interactor.ListenContractEvents(listenCtx, ch, pubKeyAssetIDPairs)
	}()

	time.Sleep(1 * time.Second)

	update := map[types.AssetEntry]publisher_agent.SignedPriceUpdate[*shared.EvmSignature]{
		{
			AssetID:        "ETHUSD",
			PublicKey:      shared.PublisherKey(s.pubKey.Hex()),
			Historical:     false,
		}: {
			OracleID: "local",
			AssetID:  "ETHUSD",
			Trigger:  publisher_agent.ClockTriggerType,
			SignedPrice: publisher_agent.SignedPrice[*shared.EvmSignature]{
				PublisherKey:    shared.PublisherKey(s.pubKey.Hex()),
				ExternalAssetID: "ETHUSD",
				SignatureType:   shared.EvmSignatureType,
				QuantizedPrice:  "1200000000000000000",
				TimestampedSignature: shared.TimestampedSignature[*shared.EvmSignature]{
					TimestampNano: 1680210936000000000,
					Signature: &shared.EvmSignature{
						R: "0xbf6ea4d88cb8127b4f318ab718fcdd5e0c83f48269b7abfdfe23f0daf35c79ee",
						S: "0x0a5d9a4ef69f050981f5f0178e4733754f520631611b2fcaca6fff78deccd21e",
						V: "0x1c",
					},
				},
			},
		},
		{
			AssetID:        "BTCUSD",
			PublicKey:      shared.PublisherKey(s.pubKey.Hex()),
			Historical:     false,
		}: {
			OracleID: "local",
			AssetID:  "BTCUSD",
			Trigger:  publisher_agent.ClockTriggerType,
			SignedPrice: publisher_agent.SignedPrice[*shared.EvmSignature]{
				PublisherKey:    shared.PublisherKey(s.pubKey.Hex()),
				ExternalAssetID: "BTCUSD",
				SignatureType:   shared.EvmSignatureType,
				QuantizedPrice:  "2100000000000000000",
				TimestampedSignature: shared.TimestampedSignature[*shared.EvmSignature]{
					TimestampNano: 1680210936000000000,
					Signature: &shared.EvmSignature{
						R: "0x14b51db5e4674b19f6299c34ddd2a20322860ef0688e46482a9d573670645028",
						S: "0x5ffea38d560430fc7ac5a9fc35dee2cbfc95caeee744be5729d7716e4644395e",
						V: "0x1b",
					},
				},
			},
		},
	}
	err := s.interactor.BatchPushToContract(update)
	s.Require().NoError(err)

	expectedValues := map[shared.AssetID]struct {
		quantizedValue string
		timestamp      uint64
	}{
		"ETHUSD": {
			quantizedValue: "1200000000000000000",
			timestamp:      1680210936000000000,
		},
		"BTCUSD": {
			quantizedValue: "2100000000000000000",
			timestamp:      1680210936000000000,
		},
	}

	receivedAssets := make(map[shared.AssetID]bool)
	timeout := time.After(5 * time.Second)

	for len(receivedAssets) < len(expectedValues) {
		select {
		case receivedUpdate := <-ch:
			s.logger.Info().
				Str("pubkey", receivedUpdate.Pubkey.Hex()).
				Int("num_values", len(receivedUpdate.ContractValueMap)).
				Msg("Received contract event")

			s.Require().NotEmpty(receivedUpdate.ContractValueMap)
			s.Require().Equal(s.pubKey, receivedUpdate.Pubkey)
			s.Require().Equal(1, len(receivedUpdate.ContractValueMap))

			// Validate each asset in the update
			for assetID, value := range receivedUpdate.ContractValueMap {
				expected, exists := expectedValues[assetID]
				s.Require().True(exists, "Received unexpected asset: %s", assetID)
				s.Require().Equal(expected.quantizedValue, value.QuantizedValue.String())
				s.Require().Equal(expected.timestamp, value.TimestampNs)
				receivedAssets[assetID] = true
			}

		case <-timeout:
			s.Require().Fail("test timed out after 5 seconds, should have received all contract events. Got %d/%d events", len(receivedAssets), len(expectedValues))
		}
	}

	s.Require().Equal(len(expectedValues), len(receivedAssets), "Should have received all expected assets")
}
