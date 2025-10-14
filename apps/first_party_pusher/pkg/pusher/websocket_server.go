package pusher

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	publisher_agent "github.com/Stork-Oracle/stork-external/apps/publisher_agent/pkg"
	"github.com/Stork-Oracle/stork-external/shared"
)

const (
	HandshakeTimeout = 10 * time.Second
	ReadBufferSize   = 1048576
	WriteBufferSize  = 1048576
	WriteTimeout     = 10 * time.Second
)

type WebsocketServer[T shared.Signature] struct {
	port                string
	upgrader            websocket.Upgrader
	signedPriceUpdateCh chan publisher_agent.SignedPriceUpdate[T]
	logger              zerolog.Logger
	server              *http.Server
	mutex               sync.RWMutex
	connections         map[*websocket.Conn]bool
}

func NewWebsocketServer[T shared.Signature](port string, signedPriceUpdateCh chan publisher_agent.SignedPriceUpdate[T]) *WebsocketServer[T] {
	return &WebsocketServer[T]{
		port: port,
		upgrader: websocket.Upgrader{
			HandshakeTimeout:  HandshakeTimeout,
			ReadBufferSize:    ReadBufferSize,
			WriteBufferSize:   WriteBufferSize,
			EnableCompression: true,
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins for first party chain pusher
			},
		},
		signedPriceUpdateCh: signedPriceUpdateCh,
		logger:              log.With().Str("component", "websocket_server").Logger(),
		connections:         make(map[*websocket.Conn]bool),
	}
}

func (ws *WebsocketServer[T]) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", ws.handleWebsocket)
	mux.HandleFunc("/health", ws.handleHealth)

	ws.server = &http.Server{
		Addr:    ":" + ws.port,
		Handler: mux,
	}

	ws.logger.Info().Str("port", ws.port).Msg("Starting WebSocket server")

	return ws.server.ListenAndServe()
}

func (ws *WebsocketServer[T]) Stop() error {
	if ws.server != nil {
		return ws.server.Close()
	}

	return nil
}

func (ws *WebsocketServer[T]) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)

	_, err := w.Write([]byte("OK"))
	if err != nil {
		ws.logger.Error().Err(err).Msg("Failed to write health response")
	}
}

func (ws *WebsocketServer[T]) handleWebsocket(w http.ResponseWriter, r *http.Request) {
	conn, err := ws.upgrader.Upgrade(w, r, nil)
	if err != nil {
		ws.logger.Error().Err(err).Msg("Failed to upgrade WebSocket connection")

		return
	}

	ws.mutex.Lock()
	ws.connections[conn] = true
	ws.mutex.Unlock()

	ws.logger.Info().Msg("New WebSocket connection established")

	defer func() {
		ws.mutex.Lock()
		delete(ws.connections, conn)
		ws.mutex.Unlock()

		err = conn.Close()
		if err != nil {
			ws.logger.Error().Err(err).Msg("Failed to close WebSocket connection")

			return
		}

		ws.logger.Info().Msg("WebSocket connection closed")
	}()

	ws.handleConnection(conn)
}

func (ws *WebsocketServer[T]) handleConnection(conn *websocket.Conn) {
	for {
		messageType, data, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				ws.logger.Error().Err(err).Msg("WebSocket read error")
			} else {
				ws.logger.Debug().Err(err).Msg("WebSocket connection closed by client")
			}

			return
		}

		if messageType != websocket.TextMessage {
			ws.logger.Warn().Int("message_type", messageType).Msg("Received non-text message")

			continue
		}

		err = ws.processMessage(data)
		if err != nil {
			ws.logger.Error().Err(err).Msg("Failed to process message")
		}
	}
}

func (ws *WebsocketServer[T]) processMessage(data []byte) error {
	var msg publisher_agent.WebsocketMessage[publisher_agent.SignedPriceUpdateBatch[T]]

	err := json.Unmarshal(data, &msg)
	if err != nil {
		return fmt.Errorf("failed to unmarshal message: %w", err)
	}

	if msg.Type != "signed_prices" {
		return fmt.Errorf("unsupported message type: %s", msg.Type)
	}

	for assetID, signedPriceUpdate := range msg.Data {
		select {
		case ws.signedPriceUpdateCh <- signedPriceUpdate:
			ws.logger.Debug().
				Str("asset", string(assetID)).
				Str("price", string(signedPriceUpdate.SignedPrice.QuantizedPrice)).
				Uint64("timestamp", signedPriceUpdate.SignedPrice.TimestampedSignature.TimestampNano).
				Msg("Received signed price update")
		default:
			ws.logger.Warn().Str("asset", string(assetID)).Msg("Signed price update channel full, dropping message")
		}
	}

	return nil
}

func isConnectionClosedError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()

	return strings.Contains(errStr, "connection reset by peer") ||
		strings.Contains(errStr, "broken pipe") ||
		strings.Contains(errStr, "websocket: close sent")
}

func (ws *WebsocketServer[T]) GetConnectionCount() int {
	ws.mutex.RLock()
	defer ws.mutex.RUnlock()

	return len(ws.connections)
}
