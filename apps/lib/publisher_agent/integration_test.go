package publisher_agent

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/Stork-Oracle/stork_external/lib/signer"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
)

func runPublisherAgent(config *StorkPublisherAgentConfig, logger zerolog.Logger) {
	err := RunPublisherAgent(config, logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("publisher agent failed")
	}
}

func getIncomingPushConnection() *websocket.Conn {
	dialer := websocket.DefaultDialer
	subUrl := fmt.Sprintf("ws://localhost:%v/publish", pushWsPort)
	conn, _, err := dialer.Dial(subUrl, http.Header{})
	if err != nil {
		panic(err)
	}

	return conn
}

type LocalBroker[T signer.Signature] struct {
	port        int
	readTimeout time.Duration
	outputCh    chan WebsocketMessage[SignedPriceUpdateBatch[T]]
	logger      zerolog.Logger
}

func (lb *LocalBroker[T]) reader(conn *websocket.Conn) {
	for {
		if lb.readTimeout > 0 {
			deadline := time.Now().Add(lb.readTimeout)
			_ = conn.SetReadDeadline(deadline)
		}
		_, message, err := conn.ReadMessage()
		if err != nil {
			continue
		}
		var websocketMessage WebsocketMessage[SignedPriceUpdateBatch[T]]
		err = json.Unmarshal(message, &websocketMessage)
		if err != nil {
			lb.logger.Error().Err(err).Msgf("local broker failed to unmarshal")
			lb.logger.Info().Msgf("response was: %s", string(message))
			continue
		}
		lb.outputCh <- websocketMessage
	}

}

func (lb *LocalBroker[T]) handleLocalBrokerConnection(resp http.ResponseWriter, req *http.Request) {
	conn, err := upgradeAndEnforceCompression(resp, req, false, getWsUpgrader(), lb.logger, "")
	if err != nil {
		// debug log because err could be rate limit violation
		lb.logger.Debug().Err(err).Object("request_headers", HttpHeaders(req.Header)).Msg("failed to complete publisher websocket handshake")
		return
	}
	go lb.reader(conn)
}

func runLocalBroker[T signer.Signature](port int, signatureType signer.SignatureType, readTimeout time.Duration, outputCh chan WebsocketMessage[SignedPriceUpdateBatch[T]], logger zerolog.Logger) {
	localBroker := LocalBroker[T]{
		port:        port,
		outputCh:    outputCh,
		logger:      logger,
		readTimeout: readTimeout,
	}

	endpoint := fmt.Sprintf("/%s/publish", signatureType)
	http.HandleFunc(endpoint, func(resp http.ResponseWriter, req *http.Request) {
		localBroker.handleLocalBrokerConnection(
			resp,
			req,
		)
	})

	logger.Info().Msgf("starting local broker http server on port %d", port)
	err := http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", port), nil)
	if err != nil {
		logger.Fatal().Err(err).Msg("local broker websocket server failed")
	}
}

func runStorkRegistry(brokerPort int, registryPort int, sigType signer.SignatureType, logger zerolog.Logger) {
	http.HandleFunc("/v1/registry/brokers", func(resp http.ResponseWriter, req *http.Request) {
		response := fmt.Sprintf("[{\"publish_url\":\"ws://localhost:%v/%s/publish\",\"asset_ids\":[\"*\"]}]", brokerPort, sigType)
		_, _ = io.WriteString(resp, response)
		logger.Info().Msgf("local stork registry returning response: %s", response)
	})

	_ = http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", registryPort), nil)
}

func TestPushWs(t *testing.T) {
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	outputCh := make(chan WebsocketMessage[SignedPriceUpdateBatch[*signer.EvmSignature]])
	go runLocalBroker(brokerPort, EvmSignatureType, time.Duration(0), outputCh, logger)

	go runStorkRegistry(brokerPort, localRegistryPort, EvmSignatureType, logger)

	deltaOnlyConfig := GetDeltaOnlyTestConfig()
	go runPublisherAgent(deltaOnlyConfig, logger)

	conn := getIncomingPushConnection()

	go func(conn *websocket.Conn) {
		i := 0
		for {
			pushUpdate := ValueUpdatePushWebsocket{
				PublishTimestamp: time.Now().UnixNano(),
				Asset:            assetId,
				Value:            i,
			}
			message := WebsocketMessage[[]ValueUpdatePushWebsocket]{
				Type: "prices",
				Data: []ValueUpdatePushWebsocket{pushUpdate},
			}
			err := conn.WriteJSON(message)
			logger.Info().Msgf("Sent update %v to push websocket", pushUpdate)
			if err != nil {
				logger.Error().Err(err).Msgf("failed to send push update")
			}
			i++
			time.Sleep(1 * time.Second)
		}
	}(conn)

	for output := range outputCh {
		logger.Info().Msgf("Broker received message: %v", output)
	}
}
