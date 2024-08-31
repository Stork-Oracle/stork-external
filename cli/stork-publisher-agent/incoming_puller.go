package stork_publisher_agent

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
)

type IncomingWebsocketPuller struct {
	Auth                AuthToken
	Url                 string
	SubscriptionRequest string
	ReconnectDelay      time.Duration
	PriceUpdateChannels []chan PriceUpdate
	Logger              zerolog.Logger
}

func (p *IncomingWebsocketPuller) Run() {
	for {
		p.Logger.Debug().Msgf("Connecting to pull-based WebSocket with url %s", p.Url)

		var headers http.Header
		if len(p.Auth) > 0 {
			headers = http.Header{"Authorization": []string{"Basic " + string(p.Auth)}}
		}

		incomingWsConn, _, err := websocket.DefaultDialer.Dial(p.Url, headers)
		if err != nil {
			p.Logger.Error().Err(err).Msgf("Failed to connect to pull-based WebSocket: %v", err)
			break
		}

		_, messageBytes, err := incomingWsConn.ReadMessage()
		if err != nil {
			p.Logger.Error().Err(err).Msgf("Failed to read connection message from pull-based WebSocket: %v", err)
		}
		p.Logger.Debug().Msgf("Received connection message: %s", messageBytes)

		if len(p.SubscriptionRequest) > 0 {
			p.Logger.Debug().Msgf("Sending subscription request: %s", p.SubscriptionRequest)
			err = incomingWsConn.WriteMessage(websocket.TextMessage, []byte(p.SubscriptionRequest))
			_, subscriptionResponse, err := incomingWsConn.ReadMessage()
			p.Logger.Debug().Msgf("Received subscription response: %s", subscriptionResponse)
			if err != nil {
				p.Logger.Error().Err(err).Msg("Failed to send subscription request to pull-based WebSocket")
				break
			}
		}

		var lastDropLogTime time.Time

		for {
			_, messageBytes, err := incomingWsConn.ReadMessage()
			if err != nil {
				p.Logger.Error().Err(err).Msg("Failed to read from pull-based WebSocket")
				break
			}
			var message WebsocketMessage[[]PriceUpdate]
			err = json.Unmarshal(messageBytes, &message)
			if err != nil {
				p.Logger.Error().Err(err).Msgf("Failed to unmarshal message from pull-based WebSocket: %s", messageBytes)
				break
			}
			for _, priceUpdate := range message.Data {
				for _, priceUpdateCh := range p.PriceUpdateChannels {
					select {
					case priceUpdateCh <- priceUpdate:
					default:
						if time.Since(lastDropLogTime) >= StalePriceThreshold {
							p.Logger.Error().Msg("dropped incoming price update - too many updates")
							lastDropLogTime = time.Now()
						}
					}
				}
			}
		}

		p.Logger.Info().Msgf("Waiting %s to reconnect to pull-based WebSocket", p.ReconnectDelay)
		time.Sleep(p.ReconnectDelay)
	}
}
