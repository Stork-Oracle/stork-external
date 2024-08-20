package stork_publisher_agent

import (
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
	"net/http"
	"sync"
)

type PublisherAgentRunner[T Signature] struct {
	config                  StorkPublisherAgentConfig
	logger                  zerolog.Logger
	priceUpdateCh           chan PriceUpdate
	signedPriceBatchCh      chan SignedPriceUpdateBatch[T]
	outgoingConnections     map[ConnectionId]*OutgoingWebsocketConnection[T]
	outgoingConnectionsLock sync.RWMutex
	incomingConnections     map[ConnectionId]*IncomingWebsocketConnection
	incomingConnectionsLock sync.RWMutex
	upgrader                websocket.Upgrader
	signer                  Signer[T]
}

func NewPublisherAgentRunner[T Signature](
	config StorkPublisherAgentConfig,
	logger zerolog.Logger,
) *PublisherAgentRunner[T] {
	signer, err := NewSigner[T](config)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to create signer")
	}
	return &PublisherAgentRunner[T]{
		config:                  config,
		logger:                  logger,
		priceUpdateCh:           make(chan PriceUpdate, 4096),
		signedPriceBatchCh:      make(chan SignedPriceUpdateBatch[T], 4096),
		outgoingConnections:     make(map[ConnectionId]*OutgoingWebsocketConnection[T]),
		outgoingConnectionsLock: sync.RWMutex{},
		incomingConnections:     make(map[ConnectionId]*IncomingWebsocketConnection),
		incomingConnectionsLock: sync.RWMutex{},
		upgrader:                GetWsUpgrader(),
		signer:                  *signer,
	}
}

func (r *PublisherAgentRunner[T]) Run() {
	processor := NewPriceUpdateProcessor[T](
		r.signer,
		r.config.ClockPeriod,
		r.config.DeltaCheckPeriod,
		r.config.ChangeThresholdProportion,
		r.priceUpdateCh,
		r.signedPriceBatchCh,
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

	processor.Run()

}

func (r *PublisherAgentRunner[T]) HandleNewSubscriberConnection(resp http.ResponseWriter, req *http.Request) {
	authToken := AuthToken("fake_auth")
	// complete the websocket handshake
	conn, err := upgradeAndEnforceCompression(resp, req, r.config.EnforceCompression, r.upgrader, r.logger, authToken)
	if err != nil {
		// debug log because err could be rate limit violation
		r.logger.Debug().Err(err).Object("request_headers", HttpHeaders(req.Header)).Msg("failed to complete subscriber websocket handshake")
		return
	}

	connId := ConnectionId(uuid.New().String())

	r.logger.Info().Str("auth_token", string(authToken)).Str("conn_id", string(connId)).Msg("adding subscriber websocket")

	websocketConn := *NewWebsocketConnection(
		conn,
		connId,
		r.logger,
		func() {
			r.logger.Info().Str("auth_token", string(authToken)).Str("conn_id", string(connId)).Msg("removing subscriber websocket")
			r.outgoingConnectionsLock.Lock()
			delete(r.outgoingConnections, connId)
			r.outgoingConnectionsLock.Unlock()
		},
	)
	outgoingWebsocketConn := NewOutgoingWebsocketConnection[T](websocketConn, r.logger)

	// add subscriber to list
	r.outgoingConnectionsLock.Lock()
	r.outgoingConnections[connId] = outgoingWebsocketConn
	r.outgoingConnectionsLock.Unlock()

	// kick off the reader and writer threads for the subscriber
	go outgoingWebsocketConn.Reader()
	go outgoingWebsocketConn.Writer()
}

func (r *PublisherAgentRunner[T]) HandleNewPublisherConnection(resp http.ResponseWriter, req *http.Request) {
	conn, err := upgradeAndEnforceCompression(resp, req, false, r.upgrader, r.logger, "")
	if err != nil {
		// debug log because err could be rate limit violation
		r.logger.Debug().Err(err).Object("request_headers", HttpHeaders(req.Header)).Msg("failed to complete publisher websocket handshake")
		return
	}

	connId := ConnectionId(uuid.New().String())

	r.logger.Info().Str("conn_id", string(connId)).Msg("adding publisher websocket")

	websocketConn := *NewWebsocketConnection(
		conn,
		connId,
		r.logger,
		func() {
			r.logger.Info().Str("conn_id", string(connId)).Msg("removing publisher websocket")
			r.incomingConnectionsLock.Lock()
			delete(r.incomingConnections, connId)
			r.incomingConnectionsLock.Unlock()
		},
	)
	incomingWebsocketConn := NewIncomingWebsocketConnection(websocketConn, r.logger)

	// add subscriber to list
	r.incomingConnectionsLock.Lock()
	r.incomingConnections[connId] = incomingWebsocketConn
	r.incomingConnectionsLock.Unlock()

	// kick off the reader thread for the publisher
	go incomingWebsocketConn.Reader(r.priceUpdateCh)
}
