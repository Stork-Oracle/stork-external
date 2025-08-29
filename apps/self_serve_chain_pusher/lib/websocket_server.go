package self_serve_chain_pusher

import (
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	publisher_agent "github.com/Stork-Oracle/stork-external/apps/publisher_agent/lib"
	"github.com/Stork-Oracle/stork-external/shared/signer"
)

const (
	HandshakeTimeout = 10 * time.Second
	ReadBufferSize   = 1048576
	WriteBufferSize  = 1048576
	WriteTimeout     = 10 * time.Second
)


type WebsocketServer struct {
	port           string
	upgrader       websocket.Upgrader
	signedPriceUpdateCh  chan publisher_agent.SignedPriceUpdate[*signer.EvmSignature]
	logger         zerolog.Logger
	server         *http.Server
	mutex          sync.RWMutex
	connections    map[*websocket.Conn]bool
}

func NewWebsocketServer(port string, signedPriceUpdateCh chan publisher_agent.SignedPriceUpdate[*signer.EvmSignature]) *WebsocketServer {
	return &WebsocketServer{
		port:          port,
		signedPriceUpdateCh: signedPriceUpdateCh,
		logger:        log.With().Str("component", "websocket_server").Logger(),
		connections:   make(map[*websocket.Conn]bool),
		upgrader: websocket.Upgrader{
			HandshakeTimeout:  HandshakeTimeout,
			ReadBufferSize:    ReadBufferSize,
			WriteBufferSize:   WriteBufferSize,
			EnableCompression: true,
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins for self-serve
			},
		},
	}
}

func (ws *WebsocketServer) Start() error {
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

func (ws *WebsocketServer) Stop() error {
	if ws.server != nil {
		return ws.server.Close()
	}
	return nil
}

func (ws *WebsocketServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (ws *WebsocketServer) handleWebsocket(w http.ResponseWriter, r *http.Request) {
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
		conn.Close()
		ws.logger.Info().Msg("WebSocket connection closed")
	}()

	ws.handleConnection(conn)
}

func (ws *WebsocketServer) handleConnection(conn *websocket.Conn) {
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

		if err := ws.processMessage(data); err != nil {
			ws.logger.Error().Err(err).Msg("Failed to process message")
			ws.sendErrorResponse(conn, err.Error())
		}
	}
}

func (ws *WebsocketServer) processMessage(data []byte) error {
	var msg publisher_agent.WebsocketMessage[publisher_agent.SignedPriceUpdateBatch[*signer.EvmSignature]]
	if err := json.Unmarshal(data, &msg); err != nil {
		return fmt.Errorf("failed to unmarshal message: %w", err)
	}

	if msg.Type != "signed_prices" {
		return fmt.Errorf("unsupported message type: %s", msg.Type)
	}

	for assetId, signedPriceUpdate := range msg.Data {
		select {
		case ws.signedPriceUpdateCh <- signedPriceUpdate:
			ws.logger.Debug().
				Str("asset", string(assetId)).
				Str("price", string(signedPriceUpdate.SignedPrice.QuantizedPrice)).
				Int64("timestamp", signedPriceUpdate.SignedPrice.TimestampedSignature.TimestampNano).
				Msg("Received signed price update")
		default:
			ws.logger.Warn().Str("asset", string(assetId)).Msg("Signed price update channel full, dropping message")
		}
	}

	return nil
}


func (ws *WebsocketServer) getBigFloatValue(value any) (*big.Float, error) {
	switch v := value.(type) {
	case float64:
		return new(big.Float).SetFloat64(v), nil
	case string:
		if v == "" {
			return nil, fmt.Errorf("value cannot be an empty string")
		}
		bf, success := new(big.Float).SetString(v)
		if !success {
			return nil, fmt.Errorf("failed to convert string to float")
		}
		return bf, nil
	case big.Float:
		return &v, nil
	default:
		return nil, fmt.Errorf("unsupported type for value: %T", v)
	}
}

func (ws *WebsocketServer) sendErrorResponse(conn *websocket.Conn, errMsg string) {
	response := publisher_agent.WebsocketMessage[interface{}]{
		Type:  "error",
		Error: errMsg,
	}

	conn.SetWriteDeadline(time.Now().Add(WriteTimeout))
	err := conn.WriteJSON(response)
	if err != nil && !isConnectionClosedError(err) {
		ws.logger.Error().Err(err).Msg("Failed to send error response")
	}
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

func (ws *WebsocketServer) GetConnectionCount() int {
	ws.mutex.RLock()
	defer ws.mutex.RUnlock()
	return len(ws.connections)
}