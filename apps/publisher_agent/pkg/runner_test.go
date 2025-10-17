package publisher_agent

import (
	"testing"

	"github.com/Stork-Oracle/stork-external/shared"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestMergeBrokers_UnionWhenOverlapping(t *testing.T) {
	t.Parallel()

	runner := &PublisherAgentRunner[*shared.EvmSignature]{
		logger: zerolog.Nop(),
	}

	sharedBrokerUrl := BrokerPublishUrl("wss://shared-broker.example.com")
	asset1 := shared.AssetID("BTCUSD")
	asset2 := shared.AssetID("ETHUSD")
	asset3 := shared.AssetID("SOLUSD")

	registryBrokers := map[BrokerPublishUrl]map[shared.AssetID]struct{}{
		sharedBrokerUrl: {
			asset2: struct{}{},
			asset3: struct{}{},
		},
	}

	seededBrokers := map[BrokerPublishUrl]map[shared.AssetID]struct{}{
		sharedBrokerUrl: {
			asset1: struct{}{},
			asset2: struct{}{},
		},
	}

	result := runner.mergeBrokers(registryBrokers, seededBrokers)

	assert.Contains(t, result, sharedBrokerUrl, "Shared broker should exist")
	assert.Len(t, result[sharedBrokerUrl], 3, "Should have exactly 3 assets")
	assert.Contains(t, result[sharedBrokerUrl], asset1, "Should have BTCUSD from seeded")
	assert.Contains(t, result[sharedBrokerUrl], asset2, "Should have ETHUSD (overlapping)")
	assert.Contains(t, result[sharedBrokerUrl], asset3, "Should have SOLUSD from registry")
}

func TestMergeBrokers_NilRegistry(t *testing.T) {
	t.Parallel()

	runner := &PublisherAgentRunner[*shared.EvmSignature]{
		logger: zerolog.Nop(),
	}

	broker1 := BrokerPublishUrl("wss://broker1.example.com")
	asset1 := shared.AssetID("BTCUSD")

	seededBrokers := map[BrokerPublishUrl]map[shared.AssetID]struct{}{
		broker1: {
			asset1: struct{}{},
		},
	}

	result := runner.mergeBrokers(nil, seededBrokers)

	assert.Contains(t, result, broker1, "Should have broker1")
	assert.Len(t, result[broker1], 1, "Should have exactly 1 asset")
	assert.Contains(t, result[broker1], asset1, "Should have BTCUSD from seeded")
}
