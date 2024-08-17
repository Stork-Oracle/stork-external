package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
	"net/http"
	"time"
)

func main() {
	zerolog.TimeFieldFormat = time.RFC3339Nano
	zerolog.DurationFieldUnit = time.Nanosecond
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack

	mainLogger := MainLogger()

	mainLogger.Info().Msg("initializing")

	upgrader := websocket.Upgrader{
		HandshakeTimeout:  time.Second * 10,
		ReadBufferSize:    1048576,
		WriteBufferSize:   1048576,
		EnableCompression: true,
		Error: func(resp http.ResponseWriter, req *http.Request, status int, reason error) {
			http.Error(resp, fmt.Sprintf(`{"type":"handshake","error":"%s"}`, reason), status)
		},
	}

	config := StorkPublisherAgentConfig{
		signatureType:             EvmSignatureType,
		clockPeriod:               500 * time.Millisecond,
		deltaCheckPeriod:          10 * time.Millisecond,
		changeThresholdProportion: 0.01,
		oracleId:                  OracleId("nwpub"),
		publisherKey:              "0xabcde",
		httpPort:                  5215,
	}

	switch config.signatureType {
	case EvmSignatureType:
		runner := *NewPublisherAgentRunner[*EvmSignature](config, mainLogger, upgrader)
		go runner.Run()

		http.HandleFunc("/evm/subscribe", runner.HandleNewSubscriberConnection)
		http.HandleFunc("/evm/publish", runner.HandleNewPublisherConnection)

		mainLogger.Warn().Msg("starting http server")
		err := http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", config.httpPort), nil)
		mainLogger.Fatal().Err(err).Msg("http server failed, process exiting")

	case StarkSignatureType:
	}
}
