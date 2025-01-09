package data_provider

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rs/zerolog"
)

type EthLogListenerDataSource struct {
	valueId      ValueId
	connectionId EthLogListenerConnectionId
	connector    EthLogListenerConnector
	logs         chan types.Log
	contract     *bind.BoundContract
	logger       zerolog.Logger
	sub          ethereum.Subscription
	ctx          context.Context
}

const (
	MaxQueryAttempts = 3
	BaseRetryDelay   = 500 * time.Millisecond
)

type EthLogListenerConnectionId struct {
	wsProviderUrl   string
	httpProviderUrl string
	abiFilename     string
	apiKeyEnvVar    string
}

func NewEthLogListenerDataSource(config DataProviderSourceConfig, connector EthLogListenerConnector) *EthLogListenerDataSource {
	connectionId := connector.GetEthLogListenerConnectionId(config)
	return &EthLogListenerDataSource{
		valueId:      config.Id,
		connectionId: connectionId,
		connector:    connector,
		logs:         make(chan types.Log),
		logger:       dataSourceLogger(connector.GetDataSourceId()),
	}
}

func (p *EthLogListenerDataSource) GetDataSourceId() DataSourceId {
	return p.connector.GetDataSourceId()
}

type EthLogListenerConnector interface {
	GetUpdateValue(contract *bind.BoundContract, valueId ValueId) (float64, error)
	GetDataSourceId() DataSourceId
	GetEthLogListenerConnectionId(config DataProviderSourceConfig) EthLogListenerConnectionId
	GetContractId() ethcommon.Address
}

func (p *EthLogListenerDataSource) reconnect() error {
	p.logger.Info().Msgf("About to connect dataSourceId %s with ws url %s", p.GetDataSourceId(), p.connectionId.wsProviderUrl)

	abiJson, err := resourcesFS.ReadFile("resources/abis/" + p.connectionId.abiFilename)
	if err != nil {
		p.logger.Fatal().Msgf("failed to read ABI file %s: %v", p.connectionId.abiFilename, err)
	}
	abi, err := abi.JSON(strings.NewReader(string(abiJson)))
	if err != nil {
		p.logger.Fatal().Msgf("failed to parse ABI file %s: %v", p.connectionId.abiFilename, err)
	}

	httpClient, err := ethclient.Dial(p.connectionId.httpProviderUrl)
	if err != nil {
		return fmt.Errorf("failed to connect to the http client: %v", err)
	}

	// Connect to WebSocket provider
	socketClient, err := ethclient.Dial(p.connectionId.wsProviderUrl)
	if err != nil {
		return fmt.Errorf("failed to connect to the websocket client: %v", err)
	}

	// Create a filter query
	address := p.connector.GetContractId()
	query := ethereum.FilterQuery{
		Addresses: []ethcommon.Address{address},
	}

	// Subscribe to the log filter
	p.ctx = context.Background()
	p.sub, err = socketClient.SubscribeFilterLogs(p.ctx, query, p.logs)
	if err != nil {
		return fmt.Errorf("failed to subscribe to filter logs: %v", err)
	}

	p.contract = bind.NewBoundContract(address, abi, httpClient, httpClient, httpClient)

	p.logger.Info().Msgf("Successfully subscribed to dataSourceId %s with ws url %s", p.GetDataSourceId(), p.connectionId.wsProviderUrl)
	return nil
}

func (p *EthLogListenerDataSource) GetUpdate() (DataSourceUpdateMap, error) {
	updates := make(DataSourceUpdateMap)

	updateTime := time.Now().UTC().UnixMilli()

	price, priceErr := p.connector.GetUpdateValue(p.contract, p.valueId)

	if priceErr != nil {
		return nil, fmt.Errorf("failed to get price from the contract: %v", priceErr)
	}

	updates[p.valueId] = DataSourceValueUpdate{
		Timestamp:    time.UnixMilli(updateTime),
		ValueId:      p.valueId,
		Value:        price,
		DataSourceId: p.connector.GetDataSourceId(),
	}

	return updates, nil
}

// readLoop reads messages from the WebSocket connection.
func (p *EthLogListenerDataSource) readLoop(updatesCh chan DataSourceUpdateMap) {
	updates, updatesErr := p.GetUpdate()
	if updatesErr != nil {
		p.logger.Error().Err(updatesErr).Msgf("failed to get updates for value id: %s", p.valueId)
	}
	updatesCh <- updates

	for {
		select {
		case err := <-p.sub.Err():
			p.logger.Warn().Msgf("Websocket error: %v", err)
			return
		case vLog := <-p.logs:
			// Handle the log entry
			address := vLog.Address
			updates, updatesErr := p.GetUpdate()
			if updatesErr != nil {
				if strings.Contains(updatesErr.Error(), "connection reset by peer") {
					p.logger.Warn().Msgf("Failed to get updates for address %s: %v", address, updatesErr)
				} else {
					p.logger.Error().Msgf("Failed to get updates for address %s: %v", address, updatesErr)
				}
			}
			updatesCh <- updates
		case <-p.ctx.Done():
			p.logger.Warn().Msg("Context canceled, exiting loop...")
			return
		}
	}
}

// handleFailure attempts to handleFailure after a delay.
func (p *EthLogListenerDataSource) handleFailure() {
	p.logger.Warn().Msg("Read loop failed")
	time.Sleep(time.Second)
}

// handleConnectFailure reacts to a connection failure.
func (p *EthLogListenerDataSource) handleConnectFailure(err error) {
	p.logger.Error().Err(err).Msg("Connection failed")
	time.Sleep(5 * time.Second)
}

func (p *EthLogListenerDataSource) Run(updatesCh chan DataSourceUpdateMap) {
	for {
		connectErr := p.reconnect()
		if connectErr != nil {
			p.handleConnectFailure(connectErr)
		} else {
			p.readLoop(updatesCh)
			p.handleFailure()
		}
	}
}
