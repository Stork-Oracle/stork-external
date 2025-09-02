package pusher

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Stork-Oracle/stork-external/apps/chain_pusher/pkg/types"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
)

const (
	// ReconnectInterval is the interval at which the client will attempt to reconnect to the websocket.
	ReconnectInterval = 1 * time.Second
	// ReconnectionAttemptErrorThreshold is the number of attempts at which the client will give up and log an error.
	ReconnectionAttemptErrorThreshold = 5
)

// StorkAggregatorWebsocketClient is a client for the Stork aggregator websocket.
type StorkAggregatorWebsocketClient struct {
	logger       zerolog.Logger
	baseEndpoint string
	authToken    string
	assetIDs     []types.AssetID
	// Default values for conn and reconnAttempts. Call connect() to set them properly.
	conn           *websocket.Conn
	reconnAttempts int
}

// NewStorkAggregatorWebsocketClient creates a new StorkAggregatorWebsocketClient with the given parameters.
func NewStorkAggregatorWebsocketClient(
	baseEndpoint, authToken string,
	assetIDs []types.AssetID,
	logger *zerolog.Logger,
) StorkAggregatorWebsocketClient {
	return StorkAggregatorWebsocketClient{
		logger:       logger.With().Str("component", "stork-ws").Logger(),
		baseEndpoint: baseEndpoint,
		authToken:    authToken,
		assetIDs:     assetIDs,
		// Default values for conn and reconnAttempts. Call connect() to set them properly.
		conn:           nil,
		reconnAttempts: 0,
	}
}

// Run connects to the Stork aggregator websocket and reads prices from the channel.
func (c *StorkAggregatorWebsocketClient) Run(priceChan chan types.AggregatedSignedPrice) {
	for {
		c.connect()

		if c.conn != nil {
			c.readLoop(priceChan)
		}

		c.handleDisconnect()
	}
}

// SubscriberMessage is a message to subscribe to one or more feeds.
type SubscriberMessage struct {
	Type string          `json:"type"`
	Data []types.AssetID `json:"data"`
}

func (c *StorkAggregatorWebsocketClient) readLoop(priceChan chan types.AggregatedSignedPrice) {
	for {
		_, message, err := c.conn.ReadMessage()
		if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
			c.logger.Info().Msg("websocket closed")

			return
		}

		if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure) {
			c.logger.Warn().Err(err).Msg("unexpected websocket close")

			return
		}

		if err != nil {
			c.logger.Error().Err(err).Msg("failed to read websocket message")

			return
		}

		if strings.Contains(string(message), `"type":"subscribe"`) {
			continue
		}

		var oracleMsg types.OraclePricesMessage

		err = json.Unmarshal(message, &oracleMsg)
		if err != nil {
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

	evmConn, resp, err := dialer.Dial(c.baseEndpoint+"/evm/subscribe", http.Header{
		"Authorization": []string{"Basic " + c.authToken},
	})
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}

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
		Data: c.assetIDs,
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

	c.logger.Info().Msgf("subscribed to %d feed%s", len(c.assetIDs), Pluralize(len(c.assetIDs)))

	c.reconnAttempts = 0
	c.conn = evmConn
}

func (c *StorkAggregatorWebsocketClient) handleDisconnect() {
	c.logger.Info().Msg(fmt.Sprintf("websocket disconnected, reconnecting in %s", ReconnectInterval))
	time.Sleep(ReconnectInterval)
}
