package chain_pusher

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
)

const (
	ReconnectInterval                 = 1 * time.Second
	ReconnectionAttemptErrorThreshold = 5
)

type StorkAggregatorWebsocketClient struct {
	logger       zerolog.Logger
	baseEndpoint string
	authToken    string
	assetIds     []AssetId

	conn           *websocket.Conn
	reconnAttempts int
}

func NewStorkAggregatorWebsocketClient(baseEndpoint, authToken string, assetIds []AssetId, logger zerolog.Logger) StorkAggregatorWebsocketClient {
	return StorkAggregatorWebsocketClient{
		logger:       logger.With().Str("component", "stork-ws").Logger(),
		baseEndpoint: baseEndpoint,
		authToken:    authToken,
		assetIds:     assetIds,
	}
}

func (p *StorkAggregatorWebsocketClient) Run(priceChan chan AggregatedSignedPrice) {
	for {
		p.connect()
		if p.conn != nil {
			p.readLoop(priceChan)
		}
		p.handleDisconnect()
	}
}

type SubscriberMessage struct {
	Type string    `json:"type"`
	Data []AssetId `json:"data"`
}

func (c *StorkAggregatorWebsocketClient) readLoop(priceChan chan AggregatedSignedPrice) {
	for {
		_, message, err := c.conn.ReadMessage()
		if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
			c.logger.Info().Msg("websocket closed")
			return
		} else if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure) {
			c.logger.Warn().Err(err).Msg("unexpected websocket close")
			return
		} else if err != nil {
			c.logger.Error().Err(err).Msg("failed to read websocket message")
			return
		} else if strings.Contains(string(message), `"type":"subscribe"`) {
			continue
		}

		var oracleMsg OraclePricesMessage
		if err := json.Unmarshal(message, &oracleMsg); err != nil {
			c.logger.Error().Err(err).Msg("failed to unmarshal message")
			continue
		}

		for _, data := range oracleMsg.Data {
			priceChan <- data
		}
	}
}

func (c *StorkAggregatorWebsocketClient) connect() {
	c.reconnAttempts++
	dialer := &websocket.Dialer{
		EnableCompression: true,
	}

	evmConn, _, err := dialer.Dial(fmt.Sprintf("%s/evm/subscribe", c.baseEndpoint), http.Header{
		"Authorization": []string{fmt.Sprintf("Basic %s", c.authToken)},
	})
	if err != nil {
		if c.reconnAttempts < ReconnectionAttemptErrorThreshold {
			c.logger.Warn().Err(err).Msg("failed to connect to websocket")
		} else {
			c.logger.Error().Err(err).Msgf("failed to connect to websocket after %d attempts", ReconnectionAttemptErrorThreshold)
		}
		return
	}
	c.logger.Info().Msg("websocket connected")

	subscribeMessage := SubscriberMessage{
		Type: "subscribe",
		Data: c.assetIds,
	}
	subscribeMessageBytes, err := json.Marshal(subscribeMessage)
	if err != nil {
		c.logger.Error().Err(err).Msg("failed to marshal subscribe message")
		return
	}

	err = evmConn.WriteMessage(websocket.TextMessage, subscribeMessageBytes)
	if err != nil {
		c.logger.Error().Err(err).Msg("failed to subscribe to feeds")
		return
	}
	c.logger.Info().Msgf("subscribed to %d feed%s", len(c.assetIds), pluralize(len(c.assetIds)))

	c.reconnAttempts = 0
	c.conn = evmConn
}

func (c *StorkAggregatorWebsocketClient) handleDisconnect() {
	c.logger.Info().Msg(fmt.Sprintf("websocket disconnected, reconnecting in %s", ReconnectInterval))
	time.Sleep(ReconnectInterval)
}
