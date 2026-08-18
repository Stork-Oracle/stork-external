//go:build integration

package starknet

import (
	"context"
	"encoding/json"
	"errors"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/NethermindEth/juno/core/felt"
	"github.com/NethermindEth/starknet.go/account"
	"github.com/NethermindEth/starknet.go/contracts"
	"github.com/NethermindEth/starknet.go/curve"
	"github.com/NethermindEth/starknet.go/hash"
	"github.com/NethermindEth/starknet.go/rpc"
	"github.com/NethermindEth/starknet.go/utils"
	"github.com/Stork-Oracle/stork-external/apps/chain_pusher/internal/testutil"
	"github.com/Stork-Oracle/stork-external/apps/chain_pusher/pkg/types"
	"github.com/caarlos0/env/v11"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// This exercises the Go bindings against the real Cairo contract, which is the only way to confirm
// that the hand written serde in bindings/serde.go agrees with what the contract expects.
//
// The whole loop is scripted; prefer that over running these by hand:
//
//	make starknet-e2e
//
// The tests skip when no Starknet node is reachable, so `make integration-test` stays green in
// environments that do not bring one up.
const (
	// The key that signed the fixtures in internal/testutil/testdata.
	storkPublicKey = "0xC4A02e7D370402F4afC36032076B05e74FF81786"

	contractArtifactDir = "../../../../chains/starknet/contracts/target/dev"
	sierraArtifact      = "stork_Stork.contract_class.json"
	casmArtifact        = "stork_Stork.compiled_contract_class.json"

	validTimePeriodSeconds = 3600

	nodeProbeTimeout = 3 * time.Second

	// Messages per asset in internal/testutil/testdata. The tests walk forward until one is fresh
	// enough for the target contract, which on a long-lived deployment can be most of the way in.
	fixtureCount = 50
)

// InteractorTestConfig points the tests at a node. The defaults match `starknet-devnet --seed 42`.
//
//nolint:tagliatelle // Env vars are namespaced by chain, matching the other pushers.
type InteractorTestConfig struct {
	RpcUrl          string `env:"STARKNET_RPC_URL"         envDefault:"http://127.0.0.1:5050"`
	AccountAddress  string `env:"STARKNET_ACCOUNT_ADDRESS" envDefault:"0x34ba56f92265f0868c57d3fe72ecab144fc96f97954bbbc4252cef8e8a979ba"`
	PrivateKey      string `env:"STARKNET_PRIVATE_KEY"     envDefault:"0xb137668388dbe9acdfa3bc734cc2c469"`
	ContractAddress string `env:"STORK_CONTRACT_ADDRESS"`
	WsUrl           string `env:"STARKNET_WS_URL"`
	// Empty means the node estimates the tip. Set it when the node does not serve estimation.
	Tip string `env:"STARKNET_TIP"`
}

// requireNode loads the config and skips the test unless a node is actually reachable.
func requireNode(t *testing.T) InteractorTestConfig {
	t.Helper()

	var config InteractorTestConfig
	require.NoError(t, env.Parse(&config))

	ctx, cancel := context.WithTimeout(context.Background(), nodeProbeTimeout)
	defer cancel()

	if _, err := rpc.NewProvider(ctx, config.RpcUrl); err != nil &&
		!errors.Is(err, rpc.ErrIncompatibleVersion) {
		t.Skipf("no Starknet node at %s, run 'make starknet-e2e' to exercise this: %v",
			config.RpcUrl, err)
	}

	return config
}

func TestIntegrationPushAndPull(t *testing.T) {
	config := requireNode(t)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	contractAddress := deployStorkContract(ctx, t, config)
	t.Logf("deployed Stork contract at %s", contractAddress)

	keyFile := []byte(config.PrivateKey)
	logger := zerolog.New(zerolog.NewTestWriter(t))

	interactor, err := NewContractInteractor(contractAddress, config.AccountAddress, "", config.Tip, keyFile, logger)
	require.NoError(t, err)
	require.NoError(t, interactor.ConnectHTTP(ctx, config.RpcUrl))

	prices, err := testutil.LoadAggregatedSignedPrices()
	require.NoError(t, err)

	// The contract keeps state across runs, so walk the fixtures forward until a batch is fresh
	// enough to apply rather than assuming the first message of each asset is.
	var (
		updates map[types.InternalEncodedAssetID]types.AggregatedSignedPrice
		lastErr error
	)

	for range fixtureCount {
		positive1, priceErr := prices.NextPositiveAsset1()
		require.NoError(t, priceErr)

		positive2, priceErr := prices.NextPositiveAsset2()
		require.NoError(t, priceErr)

		negative1, priceErr := prices.NextNegativeAsset1()
		require.NoError(t, priceErr)

		batch := map[types.InternalEncodedAssetID]types.AggregatedSignedPrice{
			prices.PositiveAsset1EncodedAssetID(): *positive1,
			prices.PositiveAsset2EncodedAssetID(): *positive2,
			prices.NegativeAsset1EncodedAssetID(): *negative1,
		}

		// A batch the contract accepts proves the calldata encoding round trips through Cairo's
		// serde *and* that the reconstructed message hash matches what Stork signed.
		lastErr = interactor.BatchPushToContract(ctx, batch)
		if lastErr == nil {
			updates = batch

			break
		}
	}

	require.NotNil(t, updates, "no fixture batch applied; last push error: %v", lastErr)

	ids := []types.InternalEncodedAssetID{
		prices.PositiveAsset1EncodedAssetID(),
		prices.PositiveAsset2EncodedAssetID(),
		prices.NegativeAsset1EncodedAssetID(),
		// A feed that was never written must not poison the batch read.
		{0xde, 0xad, 0xbe, 0xef},
	}

	// A real network needs a block before the write is visible, and the contract may already hold
	// values from an earlier run, so wait for the timestamps just pushed rather than for any value.
	var pulled map[types.InternalEncodedAssetID]types.InternalTemporalNumericValue

	require.Eventually(t, func() bool {
		var pullErr error

		pulled, pullErr = interactor.PullValues(ctx, ids)
		if pullErr != nil {
			return false
		}

		for id, price := range updates {
			value, ok := pulled[id]
			if !ok || value.TimestampNs != price.StorkSignedPrice.TimestampedSignature.TimestampNano {
				return false
			}
		}

		return true
	}, 3*time.Minute, 5*time.Second, "pushed values did not become readable on chain")

	assert.Len(t, pulled, 3, "the unwritten feed should be omitted, not error")

	for id, price := range updates {
		value, ok := pulled[id]
		require.True(t, ok, "missing value for %x", id)

		assert.Equal(t, price.StorkSignedPrice.TimestampedSignature.TimestampNano, value.TimestampNs)
		assert.Equal(t, string(price.StorkSignedPrice.QuantizedPrice), value.QuantizedValue.String())
	}

	// Specifically confirm the sign survived the round trip through the field encoding.
	negativeValue := pulled[prices.NegativeAsset1EncodedAssetID()]
	assert.Negative(
		t, negativeValue.QuantizedValue.Sign(),
		"negative price came back as %s", negativeValue.QuantizedValue,
	)
}

func TestIntegrationRepushedBatchIsRejected(t *testing.T) {
	config := requireNode(t)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	contractAddress := deployStorkContract(ctx, t, config)

	logger := zerolog.New(zerolog.NewTestWriter(t))
	interactor, err := NewContractInteractor(
		contractAddress, config.AccountAddress, "", config.Tip, []byte(config.PrivateKey), logger,
	)
	require.NoError(t, err)
	require.NoError(t, interactor.ConnectHTTP(ctx, config.RpcUrl))

	prices, err := testutil.LoadAggregatedSignedPrices()
	require.NoError(t, err)

	// The contract may already hold values from an earlier test, so walk forward until a push
	// lands rather than assuming the first fixture is fresh.
	var (
		applied map[types.InternalEncodedAssetID]types.AggregatedSignedPrice
		lastErr error
	)

	for range fixtureCount {
		price, priceErr := prices.NextPositiveAsset1()
		require.NoError(t, priceErr)

		batch := map[types.InternalEncodedAssetID]types.AggregatedSignedPrice{
			prices.PositiveAsset1EncodedAssetID(): *price,
		}
		lastErr = interactor.BatchPushToContract(ctx, batch)
		if lastErr == nil {
			applied = batch

			break
		}
	}

	require.NotNil(t, applied, "no fixture applied; last push error: %v", lastErr)

	// Re-pushing what was just applied has nothing fresh in it, so the contract rejects the whole
	// batch rather than charging for a no-op.
	require.Error(t, interactor.BatchPushToContract(ctx, applied))
}

// deployStorkContract declares and deploys a fresh Stork contract on devnet, returning its address.
func deployStorkContract(ctx context.Context, t *testing.T, config InteractorTestConfig) string {
	t.Helper()

	// Deployment is the CLI's job (chains/starknet/cli), and declaring from Go exercises code the
	// pusher never runs. Point the test at an already deployed contract when one is available.
	if config.ContractAddress != "" {
		return config.ContractAddress
	}

	acc := devnetAccount(ctx, t, config)

	sierraPath := filepath.Join(contractArtifactDir, sierraArtifact)
	casmPath := filepath.Join(contractArtifactDir, casmArtifact)

	if _, err := os.Stat(sierraPath); err != nil {
		t.Skipf("contract artifacts not built, run scarb build in chains/starknet/contracts: %v", err)
	}

	sierraBytes, err := os.ReadFile(sierraPath)
	require.NoError(t, err)

	var contractClass contracts.ContractClass
	require.NoError(t, json.Unmarshal(sierraBytes, &contractClass))

	casmClass, err := contracts.UnmarshalCasmClass(casmPath)
	require.NoError(t, err)

	classHash := hash.ClassHash(&contractClass)

	// Declaring a class that already exists is not a failure; devnet keeps state across the tests
	// in a single run, so the second deploy reuses the class declared by the first.
	if _, err := acc.Provider.Class(ctx, rpc.WithBlockTag(rpc.BlockTagLatest), classHash); err != nil {
		declareResp, declareErr := acc.BuildAndSendDeclareTxn(ctx, casmClass, &contractClass, deployTxnOptions())
		require.NoError(t, declareErr)

		_, err = acc.WaitForTransactionReceipt(ctx, declareResp.Hash, time.Second)
		require.NoError(t, err)

		classHash = declareResp.ClassHash
	}

	owner, err := utils.HexToFelt(config.AccountAddress)
	require.NoError(t, err)

	storkKey, err := utils.HexToFelt(storkPublicKey)
	require.NoError(t, err)

	constructorCalldata := []*felt.Felt{
		owner,
		storkKey,
		utils.Uint64ToFelt(validTimePeriodSeconds),
		utils.Uint64ToFelt(0), // single_update_fee low
		utils.Uint64ToFelt(0), // single_update_fee high
		utils.Uint64ToFelt(0), // fee_token (unused when the fee is zero)
	}

	deployResp, salt, err := acc.DeployContractWithUDC(ctx, classHash, constructorCalldata, deployTxnOptions(), nil)
	require.NoError(t, err)

	_, err = acc.WaitForTransactionReceipt(ctx, deployResp.Hash, time.Second)
	require.NoError(t, err)

	return contracts.PrecomputeAddress(&felt.Zero, salt, classHash, constructorCalldata).String()
}

// deployTxnOptions pads the fee estimate generously; devnet's estimates are not always enough to
// cover its own minimum transaction fee.
func deployTxnOptions() *account.TxnOptions {
	return &account.TxnOptions{FeeMultiplier: 2}
}

func devnetAccount(ctx context.Context, t *testing.T, config InteractorTestConfig) *account.Account {
	t.Helper()

	provider, err := rpc.NewProvider(ctx, config.RpcUrl)
	if err != nil {
		require.ErrorIs(t, err, rpc.ErrIncompatibleVersion)
		require.NotNil(t, provider)
	}

	privateKey, ok := new(big.Int).SetString(strings.TrimPrefix(config.PrivateKey, "0x"), 16)
	require.True(t, ok)

	address, err := utils.HexToFelt(config.AccountAddress)
	require.NoError(t, err)

	x, _ := curve.PrivateKeyToPoint(privateKey)
	publicKey := utils.BigIntToFelt(x).String()

	ks := account.NewMemKeystore()
	ks.Put(publicKey, privateKey)

	acc, err := account.NewAccount(provider, address, publicKey, ks, account.CairoV2)
	require.NoError(t, err)

	return acc
}

// The websocket path is how the pusher avoids waiting on the poll interval, and it is the one
// piece of the interactor that polling would silently paper over if it were broken.
func TestIntegrationListenContractEvents(t *testing.T) {
	config := requireNode(t)

	if config.WsUrl == "" {
		t.Skip("STARKNET_WS_URL is not set, skipping the websocket subscription test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	contractAddress := deployStorkContract(ctx, t, config)

	logger := zerolog.New(zerolog.NewTestWriter(t))
	interactor, err := NewContractInteractor(
		contractAddress, config.AccountAddress, "", config.Tip, []byte(config.PrivateKey), logger,
	)
	require.NoError(t, err)
	require.NoError(t, interactor.ConnectHTTP(ctx, config.RpcUrl))
	require.NoError(t, interactor.ConnectWs(ctx, config.WsUrl))

	events := make(chan map[types.InternalEncodedAssetID]types.InternalTemporalNumericValue, 8)
	go interactor.ListenContractEvents(ctx, events)

	// Give the subscription a moment to be established before generating an event.
	time.Sleep(2 * time.Second)

	prices, err := testutil.LoadAggregatedSignedPrices()
	require.NoError(t, err)

	var pushed types.AggregatedSignedPrice

	for range fixtureCount {
		price, priceErr := prices.NextPositiveAsset1()
		require.NoError(t, priceErr)

		batch := map[types.InternalEncodedAssetID]types.AggregatedSignedPrice{
			prices.PositiveAsset1EncodedAssetID(): *price,
		}
		if interactor.BatchPushToContract(ctx, batch) == nil {
			pushed = *price

			break
		}
	}

	require.NotNil(t, pushed.StorkSignedPrice, "no fixture was fresh enough to apply")

	// A subscription can replay a backlog, so scan for the update just pushed rather than
	// assuming it arrives first.
	wanted := pushed.StorkSignedPrice.TimestampedSignature.TimestampNano
	deadline := time.After(90 * time.Second)

	for {
		select {
		case update := <-events:
			value, ok := update[prices.PositiveAsset1EncodedAssetID()]
			if !ok || value.TimestampNs != wanted {
				continue
			}

			assert.Equal(
				t,
				string(pushed.StorkSignedPrice.QuantizedPrice),
				value.QuantizedValue.String(),
				"event carried the wrong value",
			)

			return
		case <-deadline:
			t.Fatal("no matching ValueUpdate event arrived over the websocket subscription")
		}
	}
}
