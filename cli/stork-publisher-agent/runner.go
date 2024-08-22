package stork_publisher_agent

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
	"net/http"
	"sync"
	"time"
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
		r.config.SignEveryUpdate,
		r.priceUpdateCh,
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

	if len(r.config.PullBasedWsUrl) > 0 {
		go r.RunPullBasedIncomingConnection(
			r.config.PullBasedWsUrl,
			r.config.PullBasedAuth,
			r.config.PullBasedWsSubscriptionRequest,
			r.config.PullBasedWsReconnectDelay,
		)
	}

	processor.Run()
}

func (r *PublisherAgentRunner[T]) RunPullBasedIncomingConnection(url string, auth string, subscriptionRequest string, reconnectDelay time.Duration) {
	for {
		r.logger.Info().Msgf("Connecting to pull-based WebSocket with url %s", url)

		var headers http.Header
		if len(auth) > 0 {
			headers = http.Header{"Authorization": []string{"Basic " + auth}}
		}

		incomingWsConn, _, err := websocket.DefaultDialer.Dial(url, headers)
		if err != nil {
			r.logger.Error().Err(err).Msgf("Failed to connect to pull-based WebSocket: %v", err)
			break
		}

		_, messageBytes, err := incomingWsConn.ReadMessage()
		if err != nil {
			r.logger.Error().Err(err).Msgf("Failed to read connection message from pull-based WebSocket: %v", err)
		}
		r.logger.Info().Msgf("Received connection message: %s", messageBytes)

		if len(subscriptionRequest) > 0 {
			r.logger.Info().Msgf("Sending subscription request: %s", subscriptionRequest)
			err = incomingWsConn.WriteMessage(websocket.TextMessage, []byte(subscriptionRequest))
			_, subscriptionResponse, err := incomingWsConn.ReadMessage()
			r.logger.Info().Msgf("Received subscription response: %s", subscriptionResponse)
			if err != nil {
				r.logger.Error().Err(err).Msg("Failed to send subscription request to pull-based WebSocket")
				break
			}
		}

		var lastDropLogTime time.Time

		for {
			_, messageBytes, err := incomingWsConn.ReadMessage()
			if err != nil {
				r.logger.Error().Err(err).Msg("Failed to read from pull-based WebSocket")
				break
			}
			var message WebsocketMessage[[]PriceUpdate]
			err = json.Unmarshal(messageBytes, &message)
			if err != nil {
				r.logger.Error().Err(err).Msgf("Failed to unmarshal message from pull-based WebSocket: %s", messageBytes)
				break
			}
			for _, priceUpdate := range message.Data {
				select {
				case r.priceUpdateCh <- priceUpdate:
				default:
					if time.Since(lastDropLogTime) >= time.Second*10 {
						r.logger.Error().Msg("dropped incoming price update - too many updates")
						lastDropLogTime = time.Now()
					}
				}
			}
		}

		r.logger.Info().Msgf("Waiting %s to reconnect to pull-based WebSocket", reconnectDelay)
		time.Sleep(reconnectDelay)
	}

}

func (r *PublisherAgentRunner[T]) HandleNewOutgoingConnection(resp http.ResponseWriter, req *http.Request) {
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

func (r *PublisherAgentRunner[T]) HandleNewIncomingWsConnection(resp http.ResponseWriter, req *http.Request) {
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
