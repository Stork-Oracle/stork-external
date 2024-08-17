package main

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
}

func NewPublisherAgentRunner[T Signature](
	config StorkPublisherAgentConfig,
	logger zerolog.Logger,
	upgrader websocket.Upgrader,
) *PublisherAgentRunner[T] {
	return &PublisherAgentRunner[T]{
		config:                  config,
		logger:                  logger,
		priceUpdateCh:           make(chan PriceUpdate, 4096),
		signedPriceBatchCh:      make(chan SignedPriceUpdateBatch[T], 4096),
		outgoingConnections:     make(map[ConnectionId]*OutgoingWebsocketConnection[T]),
		outgoingConnectionsLock: sync.RWMutex{},
		incomingConnections:     make(map[ConnectionId]*IncomingWebsocketConnection),
		incomingConnectionsLock: sync.RWMutex{},
		upgrader:                upgrader,
	}
}

func (r *PublisherAgentRunner[T]) Run() {
	signer := Signer[T]{}
	processor := NewPriceUpdateProcessor[T](
		signer,
		r.config.clockPeriod,
		r.config.deltaCheckPeriod,
		r.config.changeThresholdProportion,
		r.priceUpdateCh,
		r.signedPriceBatchCh,
	)

	processor.Run()
}

func (r *PublisherAgentRunner[T]) HandleNewSubscriberConnection(resp http.ResponseWriter, req *http.Request) {
	authToken := AuthToken("fake_auth")
	// complete the websocket handshake
	conn, err := upgradeAndEnforceCompression(resp, req, r.config.enforceCompression, r.upgrader, r.logger, authToken)
	if err != nil {
		// debug log because err could be rate limit violation
		r.logger.Debug().Err(err).Object("request_headers", HttpHeaders(req.Header)).Msg("failed to complete subscriber websocket handshake")
		return
	}

	connId := ConnectionId(uuid.New().String())

	r.logger.Info().Str("auth_token", string(authToken)).Str("conn_id", string(connId)).Msg("adding subscriber websocket")

	outgoingWebsocketConn := OutgoingWebsocketConnection[T]{
		// todo: use a constructor here
		WebsocketConnection: *NewWebsocketConnection(
			conn,
			connId,
			r.logger,
			func() {
				r.logger.Info().Str("auth_token", string(authToken)).Str("conn_id", string(connId)).Msg("removing subscriber websocket")
				r.outgoingConnectionsLock.Lock()
				delete(r.outgoingConnections, connId)
				r.outgoingConnectionsLock.Unlock()
			},
		),
		logger:              r.logger,
		outgoingResponsesCh: make(chan any),
		subscriptionTracker: *NewSubscriptionTracker(),
	}

	// add subscriber to list
	r.outgoingConnectionsLock.Lock()
	r.outgoingConnections[connId] = &outgoingWebsocketConn
	r.outgoingConnectionsLock.Unlock()

	// kick off the reader and writer threads for the subscriber
	go outgoingWebsocketConn.Reader()
	go outgoingWebsocketConn.Writer(r.signedPriceBatchCh)
}

func (r *PublisherAgentRunner[T]) HandleNewPublisherConnection(resp http.ResponseWriter, req *http.Request) {
	conn, err := upgradeAndEnforceCompression(resp, req, r.config.enforceCompression, r.upgrader, r.logger, "")
	if err != nil {
		// debug log because err could be rate limit violation
		r.logger.Debug().Err(err).Object("request_headers", HttpHeaders(req.Header)).Msg("failed to complete publisher websocket handshake")
		return
	}

	connId := ConnectionId(uuid.New().String())

	r.logger.Info().Str("conn_id", string(connId)).Msg("adding subscriber websocket")

	incomingWebsocketConn := IncomingWebsocketConnection{
		WebsocketConnection: *NewWebsocketConnection(
			conn,
			connId,
			r.logger,
			func() {
				r.logger.Info().Str("conn_id", string(connId)).Msg("removing subscriber websocket")
				r.incomingConnectionsLock.Lock()
				delete(r.incomingConnections, connId)
				r.incomingConnectionsLock.Unlock()
			},
		),
		logger: r.logger,
	}

	// add subscriber to list
	r.incomingConnectionsLock.Lock()
	r.incomingConnections[connId] = &incomingWebsocketConn
	r.incomingConnectionsLock.Unlock()

	// kick off the reader thread for the publisher
	go incomingWebsocketConn.Reader(r.priceUpdateCh)
}
