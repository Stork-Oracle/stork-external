package stork_publisher_agent

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
)

type PublisherAgentRunner[T Signature] struct {
	config                  StorkPublisherAgentConfig
	signatureType           SignatureType
	logger                  zerolog.Logger
	ValueUpdateCh           chan ValueUpdate
	signedPriceBatchCh      chan SignedPriceUpdateBatch[T]
	registryClient          *RegistryClient
	brokerMap               map[BrokerPublishUrl]map[AssetId]struct{}
	outgoingConnections     map[BrokerPublishUrl]*OutgoingWebsocketConnection[T]
	outgoingConnectionsLock sync.RWMutex
	signer                  Signer[T]
}

func NewPublisherAgentRunner[T Signature](
	config StorkPublisherAgentConfig,
	signatureType SignatureType,
	logger zerolog.Logger,
) *PublisherAgentRunner[T] {
	signer, err := NewSigner[T](config, signatureType, logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to create signer")
	}

	registryClient := NewRegistryClient(
		config.StorkRegistryBaseUrl,
		config.StorkAuth,
	)
	return &PublisherAgentRunner[T]{
		config:                  config,
		signatureType:           signatureType,
		logger:                  logger,
		ValueUpdateCh:           make(chan ValueUpdate, 4096),
		signedPriceBatchCh:      make(chan SignedPriceUpdateBatch[T], 4096),
		registryClient:          registryClient,
		brokerMap:               make(map[BrokerPublishUrl]map[AssetId]struct{}),
		outgoingConnections:     make(map[BrokerPublishUrl]*OutgoingWebsocketConnection[T]),
		outgoingConnectionsLock: sync.RWMutex{},
		signer:                  *signer,
	}
}

func (r *PublisherAgentRunner[T]) getPublisherKey() PublisherKey {
	var publicKey PublisherKey
	switch r.signatureType {
	case EvmSignatureType:
		publicKey = PublisherKey(r.config.EvmPublicKey)
	case StarkSignatureType:
		publicKey = PublisherKey(r.config.StarkPublicKey)
	default:
		panic("unknown signature type: " + r.signatureType)
	}
	return publicKey
}

func (r *PublisherAgentRunner[T]) UpdateBrokerConnections() {
	r.logger.Debug().Msg("Running broker connection updater")

	// query Stork Registry for brokers

	newBrokerMap, err := r.registryClient.GetBrokersForPublisher(r.getPublisherKey())
	if err != nil {
		r.logger.Error().Err(err).Msg("failed to get broker connections from Stork Registry")
		return
	}

	// add or update desired connections
	for brokerUrl, newAssetIdMap := range newBrokerMap {
		_, exists := r.brokerMap[brokerUrl]
		if exists {
			// update connection
			r.outgoingConnections[brokerUrl].UpdateAssets(newAssetIdMap)
		} else {
			// create connection
			go r.RunOutgoingConnection(brokerUrl, newAssetIdMap, r.getPublisherKey())
		}
		r.brokerMap[brokerUrl] = newAssetIdMap
	}

	// remove undesired connections
	for url, _ := range r.brokerMap {
		_, exists := newBrokerMap[url]
		if !exists {
			r.outgoingConnections[url].Remove()
			delete(r.brokerMap, url)
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
			for _, outgoingConnection := range r.outgoingConnections {
				outgoingConnection.signedPriceUpdateBatchCh <- signedPriceUpdateBatch
			}
			r.outgoingConnectionsLock.RUnlock()
		}
	}(r.signedPriceBatchCh)

	go r.RunBrokerConnectionUpdater()

	processor.Run()
}

func (r *PublisherAgentRunner[T]) RunOutgoingConnection(url BrokerPublishUrl, assetIds map[AssetId]struct{}, publicKey PublisherKey) {
	for {
		r.logger.Debug().Msgf("Connecting to receiver WebSocket with url %s", url)

		nowNs := time.Now().UnixNano()
		_, signature, err := r.signer.GetConnectionSignature(nowNs, publicKey)
		if err != nil {
			r.logger.Error().Err(err).Msg("failed to sign connection")
		}
		headers := http.Header{
			"X-PUBLIC-KEY": []string{string(publicKey)},
			"X-TIMESTAMP":  []string{strconv.FormatInt(nowNs, 10)},
			"X-SIG-TYPE":   []string{string(r.signatureType)},
			"X-SIGNATURE":  []string{*signature},
		}

		conn, _, err := websocket.DefaultDialer.Dial(string(url), headers)
		if err != nil {
			r.logger.Error().Err(err).Msgf("Failed to connect to outgoing WebSocket: %v", err)
			break
		}

		r.logger.Debug().Str("broker_url", string(url)).Msg("adding receiver websocket")

		websocketConn := *NewWebsocketConnection(
			conn,
			r.logger,
			func() {
				r.logger.Info().Str("broker_url", string(url)).Msg("removing receiver websocket")
				r.outgoingConnectionsLock.Lock()
				delete(r.outgoingConnections, url)
				r.outgoingConnectionsLock.Unlock()
			},
		)
		outgoingWebsocketConn := NewOutgoingWebsocketConnection[T](websocketConn, assetIds, r.logger)

		// add subscriber to list
		r.outgoingConnectionsLock.Lock()
		r.outgoingConnections[url] = outgoingWebsocketConn
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
