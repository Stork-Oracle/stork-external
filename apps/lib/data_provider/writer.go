package data_provider

import (
	"fmt"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
)

const reconnectDuration = 5 * time.Second

type WebsocketWriter struct {
	updateCh chan DataSourceUpdateMap
	wsUrl    string
	verbose  bool
	logger   zerolog.Logger
}

func NewWebsocketWriter(wsUrl string, verbose bool) *WebsocketWriter {
	return &WebsocketWriter{
		wsUrl:   wsUrl,
		verbose: verbose,
		logger:  writerLogger(),
	}
}

func (w *WebsocketWriter) Run(updateCh chan DataSourceUpdateMap) {
	// always try to reconnect when we get disconnected
	for {
		err := w.runWriteLoop(updateCh)
		if err != nil {
			w.logger.Info().Err(err).Str("url", w.wsUrl).Msgf("Loop exited - waiting %s to resume", reconnectDuration)
			time.Sleep(reconnectDuration)
		}
	}
}

func (w *WebsocketWriter) runWriteLoop(updateCh chan DataSourceUpdateMap) error {
	var conn *websocket.Conn
	var err error
	if len(w.wsUrl) > 0 {
		conn, _, err = websocket.DefaultDialer.Dial(w.wsUrl, nil)
		if err != nil {
			return fmt.Errorf("error connecting to Websocket at %s: %v", w.wsUrl, err)
		}
	}

	for updateMap := range updateCh {
		for _, update := range updateMap {
			if conn != nil {
				err := conn.WriteJSON(update)
				if err != nil {
					return fmt.Errorf("error writing to Websocket at %s: %v", w.wsUrl, err)
				}
			}

			if w.verbose {
				w.logger.Debug().Msgf("Update received: %v", update)
			}
		}
	}

	return nil
}
