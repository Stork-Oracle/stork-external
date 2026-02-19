package publisher_agent

import (
	"net/http"
	"sync"
	"time"

	"github.com/Stork-Oracle/stork-external/shared"
	"github.com/Stork-Oracle/stork-external/shared/signer"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
)

type PublisherAgentRunner[T shared.Signature] struct {
	config                      StorkPublisherAgentConfig
	signatureType               shared.SignatureType
	logger                      zerolog.Logger
	ValueUpdateCh               chan ValueUpdate
	signedPriceBatchCh          chan SignedPriceUpdateBatch[T]
	registryClient              RegistryClientI
	seededBrokers               map[BrokerPublishUrl]map[shared.AssetID]struct{}
	assetsByBroker              map[BrokerPublishUrl]map[shared.AssetID]struct{}
	outgoingConnectionsByBroker map[BrokerPublishUrl]*OutgoingWebsocketConnection[T]
	wsConnectFn                 func(urlStr string, requestHeader http.Header) (connI, error)
	outgoingConnectionsLock     sync.RWMutex
	signer                      signer.Signer[T]
	storkAuthSigner             signer.StorkAuthSigner
	publisherMetadataReporter   *PublisherMetadataReporter
}

func NewPublisherAgentRunner[T shared.Signature](
	config StorkPublisherAgentConfig,
	signer signer.Signer[T],
	storkAuthSigner signer.StorkAuthSigner,
	signatureType shared.SignatureType,
	logger zerolog.Logger,
) *PublisherAgentRunner[T] {
	registryClient := NewRegistryClient(
		config.StorkRegistryBaseUrl,
		storkAuthSigner,
		logger,
	)

	publisherMetadataReporter := NewPublisherMetadataReporter(
		signer.GetPublisherKey(),
		signatureType,
		config.PublisherMetadataUpdateInterval,
		config.PublisherMetadataBaseUrl,
		storkAuthSigner,
		logger,
		config,
	)

	seededBrokers := make(map[BrokerPublishUrl]map[shared.AssetID]struct{})
	for _, broker := range config.SeededBrokers {
		assetIDs := make(map[shared.AssetID]struct{})
		for _, asset := range broker.AssetIDs {
			assetIDs[asset] = struct{}{}
		}

		seededBrokers[broker.PublishUrl] = assetIDs
	}

	wsConnectFn := func(urlStr string, requestHeader http.Header) (connI, error) {
		conn, _, err := websocket.DefaultDialer.Dial(urlStr, requestHeader)
		return conn, err
	}

	return &PublisherAgentRunner[T]{
		config:                      config,
		signatureType:               signatureType,
		logger:                      logger,
		ValueUpdateCh:               make(chan ValueUpdate, 4096),
		signedPriceBatchCh:          make(chan SignedPriceUpdateBatch[T], 4096),
		registryClient:              registryClient,
		seededBrokers:               seededBrokers,
		assetsByBroker:              make(map[BrokerPublishUrl]map[shared.AssetID]struct{}),
		outgoingConnectionsByBroker: make(map[BrokerPublishUrl]*OutgoingWebsocketConnection[T]),
		outgoingConnectionsLock:     sync.RWMutex{},
		wsConnectFn:                 wsConnectFn,
		signer:                      signer,
		storkAuthSigner:             storkAuthSigner,
		publisherMetadataReporter:   publisherMetadataReporter,
	}
}

func (r *PublisherAgentRunner[T]) UpdateBrokerConnections() {
	r.logger.Debug().Msg("Running broker connection updater")

	// query Stork Registry for brokers

	publicKey := r.signer.GetPublisherKey()
	registryBrokers, err := r.registryClient.GetBrokersForPublisher(publicKey)
	if err != nil {
		r.logger.Error().Err(err).Msg("failed to get broker connections from Stork Registry")
	}

	newBrokerMap := r.mergeBrokers(registryBrokers, r.seededBrokers)

	// add or update desired connections
	r.outgoingConnectionsLock.RLock()

	for brokerUrl, newAssetIDMap := range newBrokerMap {
		outgoingConnection, outgoingConnectionExists := r.outgoingConnectionsByBroker[brokerUrl]
		if outgoingConnectionExists {
			// update connection
			outgoingConnection.assets.UpdateAssets(newAssetIDMap)
		} else {
			// create connection
			go r.RunOutgoingConnection(brokerUrl, newAssetIDMap)
		}
		r.assetsByBroker[brokerUrl] = newAssetIDMap
	}

	r.outgoingConnectionsLock.RUnlock()

	// remove undesired connections
	for url := range r.assetsByBroker {
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
	processor := NewPriceUpdateProcessor(
		r.signer,
		r.config.OracleID,
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

	if r.config.PublisherMetadataUpdateInterval.Nanoseconds() > 0 {
		go r.publisherMetadataReporter.Run()
	}

	processor.Run()
}

func (r *PublisherAgentRunner[T]) RunOutgoingConnection(url BrokerPublishUrl, assetIds map[shared.AssetID]struct{}) {
	assets := NewOutgoingWebsocketConnectionAssets[T](assetIds)

	for {
		r.logger.Debug().Msgf("Connecting to receiver WebSocket with url %s", url)

		headers, err := r.storkAuthSigner.GetAuthHeaders()
		if err != nil {
			r.logger.Error().Err(err).Msgf("Failed to get auth headers for outgoing WebSocket: %v", err)
		}

		conn, err := r.wsConnectFn(string(url), headers)
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
		outgoingWebsocketConn := NewOutgoingWebsocketConnection(websocketConn, assets, r.logger)

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

func (r *PublisherAgentRunner[T]) mergeBrokers(
	registryBrokers map[BrokerPublishUrl]map[shared.AssetID]struct{},
	seededBrokers map[BrokerPublishUrl]map[shared.AssetID]struct{},
) map[BrokerPublishUrl]map[shared.AssetID]struct{} {
	if registryBrokers == nil {
		return seededBrokers
	}

	// merge seeded brokers with registry brokers
	for brokerUrl, assetIDs := range seededBrokers {
		_, exists := registryBrokers[brokerUrl]
		if !exists {
			registryBrokers[brokerUrl] = assetIDs
		} else {
			for assetID := range assetIDs {
				registryBrokers[brokerUrl][assetID] = struct{}{}
			}
		}
	}

	return registryBrokers
}
