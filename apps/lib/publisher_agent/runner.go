package publisher_agent

import (
	"net/http"
	"sync"
	"time"

	"github.com/Stork-Oracle/stork-external/apps/lib/signer"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
)

type PublisherAgentRunner[T signer.Signature] struct {
	config                      StorkPublisherAgentConfig
	signatureType               signer.SignatureType
	logger                      zerolog.Logger
	ValueUpdateCh               chan ValueUpdate
	signedPriceBatchCh          chan SignedPriceUpdateBatch[T]
	registryClient              *RegistryClient
	assetsByBroker              map[BrokerPublishUrl]map[AssetId]struct{}
	outgoingConnectionsByBroker map[BrokerPublishUrl]*OutgoingWebsocketConnection[T]
	outgoingConnectionsLock     sync.RWMutex
	signer                      signer.Signer[T]
}

func NewPublisherAgentRunner[T signer.Signature](
	config StorkPublisherAgentConfig,
	signer signer.Signer[T],
	signatureType signer.SignatureType,
	logger zerolog.Logger,
) *PublisherAgentRunner[T] {
	registryClient := NewRegistryClient(
		config.StorkRegistryBaseUrl,
		config.StorkAuth,
		logger,
	)
	return &PublisherAgentRunner[T]{
		config:                      config,
		signatureType:               signatureType,
		logger:                      logger,
		ValueUpdateCh:               make(chan ValueUpdate, 4096),
		signedPriceBatchCh:          make(chan SignedPriceUpdateBatch[T], 4096),
		registryClient:              registryClient,
		assetsByBroker:              make(map[BrokerPublishUrl]map[AssetId]struct{}),
		outgoingConnectionsByBroker: make(map[BrokerPublishUrl]*OutgoingWebsocketConnection[T]),
		outgoingConnectionsLock:     sync.RWMutex{},
		signer:                      signer,
	}
}

func (r *PublisherAgentRunner[T]) UpdateBrokerConnections() {
	r.logger.Debug().Msg("Running broker connection updater")

	// query Stork Registry for brokers

	newBrokerMap, err := r.registryClient.GetBrokersForPublisher(r.signer.GetPublisherKey())
	if err != nil {
		r.logger.Error().Err(err).Msg("failed to get broker connections from Stork Registry")
		return
	}

	// add or update desired connections
	for brokerUrl, newAssetIdMap := range newBrokerMap {
		outgoingConnection, outgoingConnectionExists := r.outgoingConnectionsByBroker[brokerUrl]
		if outgoingConnectionExists {
			// update connection
			outgoingConnection.UpdateAssets(newAssetIdMap)
		} else {
			// create connection
			go r.RunOutgoingConnection(brokerUrl, newAssetIdMap, r.config.StorkAuth)
		}
		r.assetsByBroker[brokerUrl] = newAssetIdMap
	}

	// remove undesired connections
	for url, _ := range r.assetsByBroker {
		_, exists := newBrokerMap[url]
		if !exists {
			r.outgoingConnectionsByBroker[url].Remove()
			delete(r.assetsByBroker, url)
		}
	}

	r.logger.Debug().Msg("Broker connection updater finished")
}

func (r *PublisherAgentRunner[T]) RunBrokerConnectionUpdater() {
	r.UpdateBrokerConnections()
	for range time.Tick(r.config.StorkRegistryRefreshInterval) {
		r.UpdateBrokerConnections()
	}
}

func (r *PublisherAgentRunner[T]) Run() {
	processor := NewPriceUpdateProcessor[T](
		r.signer,
		r.config.OracleId,
		len(r.config.SignatureTypes),
		r.config.ClockPeriod,
		r.config.DeltaCheckPeriod,
		r.config.ChangeThresholdProportion,
		r.config.SignEveryUpdate,
		r.ValueUpdateCh,
		r.signedPriceBatchCh,
		r.logger,
	)

	// fan out the signed update to all subscriber websockets
	go func(signedPriceBatchCh chan SignedPriceUpdateBatch[T]) {
		for signedPriceUpdateBatch := range signedPriceBatchCh {
			r.outgoingConnectionsLock.RLock()
			for _, outgoingConnection := range r.outgoingConnectionsByBroker {
				outgoingConnection.signedPriceUpdateBatchCh <- signedPriceUpdateBatch
			}
			r.outgoingConnectionsLock.RUnlock()
		}
	}(r.signedPriceBatchCh)

	go r.RunBrokerConnectionUpdater()

	processor.Run()
}

func (r *PublisherAgentRunner[T]) RunOutgoingConnection(url BrokerPublishUrl, assetIds map[AssetId]struct{}, authToken AuthToken) {
	for {
		r.logger.Debug().Msgf("Connecting to receiver WebSocket with url %s", url)

		var headers http.Header
		if len(authToken) > 0 {
			headers = http.Header{"Authorization": []string{"Basic " + string(authToken)}}
		}

		conn, _, err := websocket.DefaultDialer.Dial(string(url), headers)
		if err != nil {
			r.logger.Error().Err(err).Msgf("Failed to connect to outgoing WebSocket: %v", err)
			time.Sleep(r.config.BrokerReconnectDelay)
			continue
		}

		r.logger.Debug().Str("broker_url", string(url)).Msg("adding receiver websocket")

		websocketConn := *NewWebsocketConnection(
			conn,
			r.logger,
			func() {
				r.logger.Info().Str("broker_url", string(url)).Msg("removing receiver websocket")
				r.outgoingConnectionsLock.Lock()
				delete(r.outgoingConnectionsByBroker, url)
				r.outgoingConnectionsLock.Unlock()
			},
		)
		outgoingWebsocketConn := NewOutgoingWebsocketConnection[T](websocketConn, assetIds, r.logger)

		// add subscriber to list
		r.outgoingConnectionsLock.Lock()
		r.outgoingConnectionsByBroker[url] = outgoingWebsocketConn
		r.outgoingConnectionsLock.Unlock()

		// read until a failure happens or the connection is closed
		outgoingWebsocketConn.Writer()

		if outgoingWebsocketConn.removed {
			r.logger.Info().Msg("Outgoing websocket was removed - not reconnecting")
			return
		} else {
			r.logger.Warn().Msgf("Outgoing websocket writer thread failed - reconnecting after %s", r.config.BrokerReconnectDelay)
			time.Sleep(r.config.BrokerReconnectDelay)
		}
	}
}
