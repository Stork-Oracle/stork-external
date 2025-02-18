package data_provider

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Stork-Oracle/stork-external/apps/lib/data_provider/types"
	"github.com/Stork-Oracle/stork-external/apps/lib/data_provider/utils"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
)

const reconnectDuration = 5 * time.Second

type Writer interface {
	Run(updateCh chan types.DataSourceUpdateMap)
}

func GetWriter(outputAddress string) (Writer, error) {
	if strings.HasPrefix(outputAddress, "ws://") || len(outputAddress) == 0 {
		return NewWebsocketWriter(outputAddress), nil
	}
	return nil, fmt.Errorf("invalid output address: %s", outputAddress)
}

type WebsocketWriter struct {
	updateCh chan types.DataSourceUpdateMap
	wsUrl    string
	logger   zerolog.Logger
}

func NewWebsocketWriter(wsUrl string) *WebsocketWriter {
	return &WebsocketWriter{
		wsUrl:  wsUrl,
		logger: utils.WriterLogger(),
	}
}

func (w *WebsocketWriter) Run(updateCh chan types.DataSourceUpdateMap) {
	// always try to reconnect when we get disconnected
	for {
		err := w.runWriteLoop(updateCh)
		if err != nil {
			w.logger.Info().Err(err).Str("url", w.wsUrl).Msgf("Loop exited - waiting %s to resume", reconnectDuration)
			time.Sleep(reconnectDuration)
		}
	}
}

func (w *WebsocketWriter) runWriteLoop(updateCh chan types.DataSourceUpdateMap) error {
	var conn *websocket.Conn
	var err error
	if len(w.wsUrl) > 0 {
		conn, _, err = websocket.DefaultDialer.Dial(w.wsUrl, nil)
		if err != nil {
			return fmt.Errorf("error connecting to Websocket at %s: %v", w.wsUrl, err)
		}
	}

	for updateMap := range updateCh {
		valueUpdates := make([]types.ValueUpdate, 0)
		for _, update := range updateMap {
			valueUpdate := types.ValueUpdate{
				PublishTimestamp: update.Timestamp.UnixNano(),
				ValueId:          update.ValueId,
				Value:            fmt.Sprintf(`%.18f`, update.Value),
			}
			valueUpdates = append(valueUpdates, valueUpdate)
		}

		valueUpdateWebsocketMessage := types.ValueUpdateWebsocketMessage{
			Type: "prices",
			Data: valueUpdates,
		}
		wsMessageBytes, err := json.Marshal(valueUpdateWebsocketMessage)
		if err != nil {
			w.logger.Error().Msgf("Error marshalling update %v: %v", valueUpdateWebsocketMessage, err)
		}

		w.logger.Debug().Msgf("Update: %s", string(wsMessageBytes))

		if conn != nil {
			err := conn.WriteMessage(websocket.TextMessage, wsMessageBytes)
			if err != nil {
				return fmt.Errorf("error writing to websocket at %s: %v", w.wsUrl, err)
			}
		}
	}

	return nil
}
