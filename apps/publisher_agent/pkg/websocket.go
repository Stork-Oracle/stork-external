package publisher_agent

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net"
	"net/http"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/Stork-Oracle/stork-external/shared"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
)

const (
	FullQueueLogFrequency = time.Second * 10
	HandshakeTimeout      = time.Second * 10
	ReadBufferSize        = 1048576
	WriteBufferSize       = 1048576
	OutgoingWriteTimeout  = time.Second * 10
)

type connI interface {
	Close() error
	WriteControl(messageType int, data []byte, deadline time.Time) error
	NextWriter(messageType int) (io.WriteCloser, error)
	SetReadDeadline(t time.Time) error
	NextReader() (messageType int, r io.Reader, err error)
	SetWriteDeadline(t time.Time) error
}

type WebsocketConnection struct {
	conn    connI
	logger  zerolog.Logger
	onClose func()
	closed  chan struct{}
}

func NewWebsocketConnection(conn connI, logger zerolog.Logger, onClose func()) *WebsocketConnection {
	return &WebsocketConnection{
		conn:    conn,
		logger:  logger,
		onClose: onClose,
		closed:  make(chan struct{}),
	}
}

// Close notifies the reader and writer threads that the websocket is being closed,
// then closes the websocket connection and invokes the onClose callback.
func (ws *WebsocketConnection) Close() {
	select {
	case <-ws.closed: // already closed
	default:
		close(ws.closed)
		_ = ws.conn.Close()
		ws.onClose()
	}
}

func (ws *WebsocketConnection) IsClosed() bool {
	select {
	case <-ws.closed:
		return true
	default:
		return false
	}
}

type IncomingWebsocketConnection struct {
	WebsocketConnection
	logger zerolog.Logger
}

func NewIncomingWebsocketConnection(conn WebsocketConnection, logger zerolog.Logger) *IncomingWebsocketConnection {
	return &IncomingWebsocketConnection{
		WebsocketConnection: conn,
		logger:              logger,
	}
}

func (m *ValueUpdatePushWebsocket) getBigFloatValue() (*big.Float, error) {
	switch v := m.Value.(type) {
	case float64:
		bf := new(big.Float).SetFloat64(v)
		return bf, nil
	case string:
		if v == "" {
			return nil, fmt.Errorf("value cannot be an empty string")
		}

		bf, success := new(big.Float).SetString(v)
		if !success {
			return nil, errors.New("failed to convert string to float")
		}
		return bf, nil
	case big.Float:
		return &v, nil
	default:
		return nil, fmt.Errorf("unsupported type for value: %T", v)
	}
}

func convertToValueUpdate(valueUpdatePushWebsocket ValueUpdatePushWebsocket) (*ValueUpdate, error) {
	if bigFloatVal, err := valueUpdatePushWebsocket.getBigFloatValue(); err != nil {
		return nil, err
	} else {
		valueUpdate := ValueUpdate{
			PublishTimestampNano: valueUpdatePushWebsocket.PublishTimestampNano,
			Asset:                valueUpdatePushWebsocket.Asset,
			Value:                bigFloatVal,
			Metadata:             valueUpdatePushWebsocket.Metadata,
		}
		return &valueUpdate, nil
	}
}

func (iwc *IncomingWebsocketConnection) Reader(valueUpdateChannels []chan ValueUpdate) {
	logger := iwc.logger.With().Str("op", "reader").Logger()

	var lastDropLogTime time.Time

	// recover fatal errors
	defer func() {
		if r := recover(); r != nil {
			formatted := errors.New(fmt.Sprintf("%v", r))
			logger.Fatal().Stack().Err(formatted).Msg("restarting after panic")
		}
	}()

	err := readLoop(iwc.conn, nil, logger, func(wsMsgReader io.Reader) error {
		// parse the message
		var valueUpdateMsg WebsocketMessage[[]ValueUpdatePushWebsocket]
		err := json.NewDecoder(wsMsgReader).Decode(&valueUpdateMsg)
		if err != nil {
			iwc.logger.Error().Err(err).Msgf("Failed to parse incoming message")
			err := sendWebsocketResponse(iwc.conn, "failed to parse price update", iwc.logger, OutgoingWriteTimeout)
			if err != nil {
				iwc.logger.Error().Err(err).Msgf("Failed to send error message")
			}
		} else {
			if valueUpdateMsg.Type == "prices" {
				for _, valueUpdatePushWs := range valueUpdateMsg.Data {
					valueUpdate, err := convertToValueUpdate(valueUpdatePushWs)
					if err != nil {
						iwc.logger.Error().Err(err).Msgf("Failed to parse incoming message")
						err := sendWebsocketResponse(iwc.conn, "failed to parse price update", iwc.logger, OutgoingWriteTimeout)
						if err != nil {
							iwc.logger.Error().Err(err).Msgf("Failed to send error message")
						}
						break
					}
					for _, valueUpdateCh := range valueUpdateChannels {
						select {
						case valueUpdateCh <- *valueUpdate:
						default:
							if time.Since(lastDropLogTime) >= FullQueueLogFrequency {
								logger.Error().Msg("dropped incoming price update - too many updates")
								lastDropLogTime = time.Now()
							}
						}
					}
				}
			}
		}
		return nil
	})

	if err := iwc.conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, err.Error()), time.Now().Add(time.Second)); err != nil {
		if err.Error() != "websocket: close sent" {
			logger.Warn().Err(err).Msg("failed to send close message")
		}
	}

	iwc.Close()
}

type OutgoingWebsocketConnectionAssets[T shared.Signature] struct {
	assetIds     map[shared.AssetID]struct{}
	assetIdsLock sync.RWMutex
}

func NewOutgoingWebsocketConnectionAssets[T shared.Signature](
	assetIds map[shared.AssetID]struct{},
) *OutgoingWebsocketConnectionAssets[T] {
	return &OutgoingWebsocketConnectionAssets[T]{
		assetIds:     assetIds,
		assetIdsLock: sync.RWMutex{},
	}
}

func (a *OutgoingWebsocketConnectionAssets[T]) filterSignedPriceUpdateBatch(
	signedPriceUpdateBatch SignedPriceUpdateBatch[T],
) SignedPriceUpdateBatch[T] {
	filteredPriceUpdates := make(SignedPriceUpdateBatch[T])
	a.assetIdsLock.RLock()
	_, allAssets := a.assetIds[WildcardSubscriptionAsset]
	if allAssets {
		filteredPriceUpdates = signedPriceUpdateBatch
	} else {
		for asset, signedPriceUpdate := range signedPriceUpdateBatch {
			_, exists := a.assetIds[asset]
			if exists {
				filteredPriceUpdates[asset] = signedPriceUpdate
			}
		}
	}
	a.assetIdsLock.RUnlock()
	return filteredPriceUpdates
}

func (a *OutgoingWebsocketConnectionAssets[T]) UpdateAssets(assetIds map[shared.AssetID]struct{}) {
	a.assetIdsLock.Lock()
	a.assetIds = assetIds
	a.assetIdsLock.Unlock()
}

type OutgoingWebsocketConnection[T shared.Signature] struct {
	WebsocketConnection
	assets                   *OutgoingWebsocketConnectionAssets[T]
	removed                  bool
	logger                   zerolog.Logger
	signedPriceUpdateBatchCh chan SignedPriceUpdateBatch[T]
}

func NewOutgoingWebsocketConnection[T shared.Signature](
	conn WebsocketConnection,
	assets *OutgoingWebsocketConnectionAssets[T],
	logger zerolog.Logger,
) *OutgoingWebsocketConnection[T] {
	return &OutgoingWebsocketConnection[T]{
		WebsocketConnection:      conn,
		assets:                   assets,
		signedPriceUpdateBatchCh: make(chan SignedPriceUpdateBatch[T], 4096),
		logger:                   logger,
	}
}

func (owc *OutgoingWebsocketConnection[T]) Remove() {
	owc.logger.Warn().Msg("Removal requested for outgoing websocket connection")
	owc.removed = true
	owc.Close()
}

func (owc *OutgoingWebsocketConnection[T]) Writer() {
	logger := owc.logger.With().Str("op", "writer").Logger()

	// log fatal errors
	defer func() {
		if r := recover(); r != nil {
			formatted := errors.New(fmt.Sprintf("%v", r))
			logger.Fatal().Stack().Err(formatted).Msg("restarting after panic")
		}
	}()

	for {
		var err error
		select {
		// send out a price update
		case signedPriceUpdateBatch := <-owc.signedPriceUpdateBatchCh:
			if owc.IsClosed() {
				logger.Warn().Msg("attempted to send message on closed websocket")
				return
			}

			filteredPriceUpdates := owc.assets.filterSignedPriceUpdateBatch(signedPriceUpdateBatch)
			if len(filteredPriceUpdates) > 0 {
				err = SendWebsocketMsg(owc.conn, "signed_prices", filteredPriceUpdates, "", "", logger)
				if err != nil {
					logger.Warn().Err(err).Msg("failed to send signed prices")
				}
			}
		case _ = <-owc.closed:
			logger.Warn().Msg("Close() called, exiting write loop")
			return
		}

		if err != nil {
			logger.Warn().Err(err).Msg("failed to send out a message, exiting write loop")
			owc.Close()
			return
		}
	}
}

// readLoop is a generalized function for the use case of read looping a websocket connection while enforcing rate limits.
// The callback handles the actual websocket message bytes (only TextMessage type is allowed).
func readLoop(
	conn connI,
	readTimeout *time.Duration,
	logger zerolog.Logger,
	callback func(wsMsgReader io.Reader) error,
) error {
	for {
		// set read timeout
		if readTimeout != nil {
			deadline := time.Now().Add(*readTimeout)
			_ = conn.SetReadDeadline(deadline)
		}

		// wait for next message
		wsMsgType, dataReader, err := conn.NextReader()
		if err != nil {
			if netErr, ok := err.(net.Error); ok {
				if netErr.Timeout() {
					// explicitly log timeout errors to help with debugging publishers
					logger.Warn().Err(err).Float64("timeout_sec", readTimeout.Seconds()).
						Msg("timed out while waiting for next message from websocket connection, exiting read loop")
					return errors.New("timed out while waiting for next message")
				} else {
					// log general network errors
					logger.Warn().Err(err).Msg("network error on websocket connection, exiting read loop")
					return errors.New("network error, possible connection reset")
				}
			} else if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway, websocket.CloseNoStatusReceived, websocket.CloseAbnormalClosure) {
				logger.Info().Err(err).Msg("websocket connection closed, exiting read loop")
				return err
			} else if websocket.IsUnexpectedCloseError(err) {
				logger.Warn().Err(err).Msg("unexpected websocket connection close, exiting read loop")
				return err
			} else {
				logger.Error().Err(err).Msg("failed to read next message from websocket connection, exiting read loop")
				return errors.New("failed to read next message")
			}
		}

		// check for invalid message type
		if wsMsgType != websocket.TextMessage {
			logger.Warn().Int("ws_msg_type", wsMsgType).Msg("non-text websocket message received, exiting read loop")
			return errors.New("non-text websocket message received")
		}

		// handle the message
		if err = callback(dataReader); err != nil {
			logger.Warn().Err(err).Msg("bad message, exiting read loop")
			return err
		}
	}
}

func SendWebsocketMsg[T any](
	conn connI,
	msgType string,
	data T,
	traceId string,
	errMsg string,
	logger zerolog.Logger,
) error {
	// create a new websocket message
	msg := WebsocketMessage[T]{
		Type:    msgType,
		TraceID: traceId,
		Error:   errMsg,
		Data:    data,
	}

	return sendWebsocketResponse(conn, msg, logger, OutgoingWriteTimeout)
}

func sendWebsocketResponse[T any](
	conn connI,
	msg T,
	logger zerolog.Logger,
	writeTimeout time.Duration,
) error {
	if writeTimeout.Nanoseconds() > 0 {
		deadline := time.Now().Add(writeTimeout)
		_ = conn.SetWriteDeadline(deadline)
	}

	// a websocket connection can be closed at any time, so we need to handle this case in each part of the write process
	if dataWriter, err := conn.NextWriter(websocket.TextMessage); err != nil {
		if netErr, ok := err.(net.Error); ok {
			if netErr.Timeout() {
				logger.Warn().Err(err).Msg("Timed out while getting next websocket writer")
				return fmt.Errorf("timed out while getting next websocket writer: %v", netErr)
			} else {
				logger.Warn().Err(err).Msg("Network error while getting next websocket writer")
				return fmt.Errorf("network error while getting next websocket writer: %v", netErr)
			}
		} else if err.Error() == "websocket: close sent" || strings.Contains(err.Error(), "connection reset by peer") {
			logger.Warn().Err(err).Msg("websocket connection closed while getting next writer")
			return fmt.Errorf("websocket connection closed while getting next writer: %v", err)
		} else {
			logger.Error().Err(err).Msgf("failed to get next websocket writer. Err type: %T", err)
			return fmt.Errorf("failed to get next websocket writer. Err type: %T err: %v", err, err)
		}
	} else if err = json.NewEncoder(dataWriter).Encode(msg); err != nil {
		logger.Error().Err(err).Msgf("failed to serialize websocket message. Err type: %T", err)
		return fmt.Errorf("failed to serialize websocket message. Err type: %T: %v", err, err)
	} else if err = dataWriter.Close(); err != nil {
		if netErr, ok := err.(net.Error); ok {
			if netErr.Timeout() {
				logger.Warn().Err(err).Msg("timed out while flushing websocket message")
				return fmt.Errorf("timed out while flushing websocket message: %v", netErr)
			} else {
				logger.Warn().Err(err).Msg("network error while flushing websocket message")
				return fmt.Errorf("network error while flushing websocket message: %v", netErr)
			}
		} else if err.Error() == "websocket: close sent" || strings.Contains(err.Error(), "connection reset by peer") {
			logger.Warn().Err(err).Msg("websocket connection closed while flushing message")
			return fmt.Errorf("websocket connection closed while flushing websocket message: %v", err)
		} else {
			logger.Error().Err(err).Msgf("failed to flush websocket message. Err type: %T", err)
			return fmt.Errorf("failed to flush websocket message. Err type: %T: %v", err, err)
		}
	}
	return nil
}

func upgradeAndEnforceCompression(
	resp http.ResponseWriter,
	req *http.Request,
	enforceCompression bool,
	upgrader websocket.Upgrader,
	logger zerolog.Logger,
	authToken shared.AuthToken,
) (*websocket.Conn, error) {
	// all subscriber connections (except stork) must have the permessage-deflate extension to enable compression,
	// this cuts outgoing data size by ~75% per subscriber, huge aws egress cost savings.
	hasCompressionHeader := false
	if enforceCompression {
		for _, ext := range req.Header["Sec-Websocket-Extensions"] {
			if strings.Contains(ext, "permessage-deflate") {
				hasCompressionHeader = true
				break
			}
		}
		if !hasCompressionHeader {
			http.Error(
				resp,
				`{"type":"handshake","error":"missing permessage-deflate extension"}`,
				http.StatusBadRequest,
			)
			return nil, errors.New("compression not negotiated")
		}
	}

	// handshake
	if ws, err := upgrader.Upgrade(resp, req, nil); err != nil {
		logger.Warn().
			Str("token", string(authToken)).
			Object("request_headers", HttpHeaders(req.Header)).
			Msg("websocket handshake failed")
		// http response happens in Upgrade(...)
		return nil, err
	} else {
		isCompressed := !reflect.ValueOf(ws).Elem().FieldByName("newCompressionWriter").IsNil()
		logger.Debug().Bool("compressed", isCompressed).Object("request_headers", HttpHeaders(req.Header)).
			Object("response_headers", HttpHeaders(resp.Header())).Msg("websocket handshake completed")

		// one more check for compression just to be sure
		if enforceCompression && !isCompressed {
			_ = ws.WriteMessage(websocket.TextMessage, []byte(`{"type":"handshake","error":"compression not negotiated"}`))
			_ = ws.Close()
			return nil, errors.New("compression header present but not negotiated successfully")
		}

		ws.EnableWriteCompression(true)
		_ = ws.SetCompressionLevel(6)
		return ws, nil
	}
}

func getWsUpgrader() websocket.Upgrader {
	return websocket.Upgrader{
		HandshakeTimeout:  HandshakeTimeout,
		ReadBufferSize:    ReadBufferSize,
		WriteBufferSize:   WriteBufferSize,
		EnableCompression: true,
		Error: func(resp http.ResponseWriter, req *http.Request, status int, reason error) {
			http.Error(resp, fmt.Sprintf(`{"type":"handshake","error":"%s"}`, reason), status)
		},
	}
}

func HandleNewIncomingWsConnection(
	resp http.ResponseWriter,
	req *http.Request,
	logger zerolog.Logger,
	valueUpdateChannels []chan ValueUpdate,
) {
	conn, err := upgradeAndEnforceCompression(resp, req, false, getWsUpgrader(), logger, "")
	if err != nil {
		// debug log because err could be rate limit violation
		logger.Debug().
			Err(err).
			Object("request_headers", HttpHeaders(req.Header)).
			Msg("failed to complete publisher websocket handshake")
		return
	}

	connID := ConnectionID(uuid.New().String())

	logger.Debug().Str("conn_id", string(connID)).Msg("adding publisher websocket")

	websocketConn := *NewWebsocketConnection(
		conn,
		logger,
		func() {
			logger.Info().Str("conn_id", string(connID)).Msg("removing publisher websocket")
		},
	)
	incomingWebsocketConn := NewIncomingWebsocketConnection(websocketConn, logger)

	// kick off the reader thread for the publisher
	go incomingWebsocketConn.Reader(valueUpdateChannels)
}
