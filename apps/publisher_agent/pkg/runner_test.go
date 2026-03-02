package publisher_agent

import (
	"errors"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/Stork-Oracle/stork-external/shared"
	"github.com/Stork-Oracle/stork-external/shared/signer"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
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

func TestRemoveBrokerConnections(t *testing.T) {
	brokerPublishUrl1 := BrokerPublishUrl("wss://broker1.example.com")
	brokerPublishUrl2 := BrokerPublishUrl("wss://broker2.example.com")

	registry := NewMockRegistryClientI(t)
	evmSigner, err := signer.NewEvmSigner("0x8b558d5fc31eb64bb51d44b4b28658180e96764d5d5ac68e6d124f86f576d9de", zerolog.Logger{})
	require.NoError(t, err)
	evmAuthSigner, err := signer.NewEvmAuthSigner("0x8b558d5fc31eb64bb51d44b4b28658180e96764d5d5ac68e6d124f86f576d9de", zerolog.Logger{})
	require.NoError(t, err)

	conn := NewMockConnI(t)
	conn.On("Close").Return(nil).Once()

	wsConnectFn := func(urlStr string, requestHeader http.Header) (connI, error) {
		return conn, nil
	}

	runner := &PublisherAgentRunner[*shared.EvmSignature]{
		registryClient:              registry,
		signer:                      evmSigner,
		storkAuthSigner:             evmAuthSigner,
		seededBrokers:               make(map[BrokerPublishUrl]map[shared.AssetID]struct{}),
		assetsByBroker:              make(map[BrokerPublishUrl]map[shared.AssetID]struct{}),
		outgoingConnectionsByBroker: make(map[BrokerPublishUrl]*OutgoingWebsocketConnection[*shared.EvmSignature]),
		outgoingConnectionsLock:     sync.RWMutex{},
		wsConnectFn:                 wsConnectFn,
		logger:                      zerolog.Logger{},
	}

	registry.On("GetBrokersForPublisher", mock.Anything).Return(map[BrokerPublishUrl]map[shared.AssetID]struct{}{
		brokerPublishUrl1: {"BTCUSD": {}},
		brokerPublishUrl2: {"BTCUSD": {}},
	}, nil).Once()
	runner.UpdateBrokerConnections()
	time.Sleep(100 * time.Millisecond)

	registry.On("GetBrokersForPublisher", mock.Anything).Return(map[BrokerPublishUrl]map[shared.AssetID]struct{}{
		brokerPublishUrl1: {"BTCUSD": {}},
	}, nil).Once()

	done := make(chan struct{})
	go func() {
		runner.UpdateBrokerConnections()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout waiting for update broker connections")
	}

}

func TestRemoveClosedConnection(t *testing.T) {
	brokerPublishUrl1 := BrokerPublishUrl("wss://broker1.example.com")

	registry := NewMockRegistryClientI(t)
	evmSigner, err := signer.NewEvmSigner("0x8b558d5fc31eb64bb51d44b4b28658180e96764d5d5ac68e6d124f86f576d9de", zerolog.Logger{})
	require.NoError(t, err)
	evmAuthSigner, err := signer.NewEvmAuthSigner("0x8b558d5fc31eb64bb51d44b4b28658180e96764d5d5ac68e6d124f86f576d9de", zerolog.Logger{})
	require.NoError(t, err)

	conn := NewMockConnI(t)

	wsConnectFn := func(urlStr string, requestHeader http.Header) (connI, error) {
		return conn, nil
	}

	runner := &PublisherAgentRunner[*shared.EvmSignature]{
		registryClient:              registry,
		signer:                      evmSigner,
		storkAuthSigner:             evmAuthSigner,
		seededBrokers:               make(map[BrokerPublishUrl]map[shared.AssetID]struct{}),
		assetsByBroker:              make(map[BrokerPublishUrl]map[shared.AssetID]struct{}),
		outgoingConnectionsByBroker: make(map[BrokerPublishUrl]*OutgoingWebsocketConnection[*shared.EvmSignature]),
		outgoingConnectionsLock:     sync.RWMutex{},
		wsConnectFn:                 wsConnectFn,
		logger:                      zerolog.Logger{},
	}

	registry.On("GetBrokersForPublisher", mock.Anything).Return(map[BrokerPublishUrl]map[shared.AssetID]struct{}{
		brokerPublishUrl1: {"BTCUSD": {}},
	}, nil).Once()
	runner.UpdateBrokerConnections()
	time.Sleep(100 * time.Millisecond)

	// close connection (disconnected by stork)
	runner.outgoingConnectionsByBroker[brokerPublishUrl1].onClose()

	// removing the only connection means that we get an auth error when hitting the broker
	registry.On("GetBrokersForPublisher", mock.Anything).Return(nil, errors.New("fake error")).Once()

	runner.UpdateBrokerConnections()
}
