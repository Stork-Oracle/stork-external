package stork_publisher_agent

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"time"
)

var Hex32Regex = regexp.MustCompile(`^0x[0-9a-fA-F]+$`)

const DefaultClockUpdatePeriod = "500ms"
const DefaultDeltaUpdatePeriod = "10ms"
const DefaultChangeThresholdPercent = 0.1
const DefaultStorkRegistryRefreshInterval = "10m"
const DefaultStorkRegistryBaseUrl = "https://rest.jp.stork-oracle.network"
const DefaultBrokerReconnectDelay = "5s"
const DefaultPullBasedReconnectDelay = "5s"

type ConfigFile struct {
	SignatureTypes                 []SignatureType
	ClockPeriod                    string
	DeltaCheckPeriod               string
	ChangeThresholdPercent         float64 // 0-100
	StorkRegistryBaseUrl           string
	StorkRegistryRefreshInterval   string
	BrokerReconnectDelay           string
	PullBasedWsUrl                 string
	PullBasedWsSubscriptionRequest string
	PullBasedWsReconnectDelay      string
	SignEveryUpdate                bool
	IncomingWsPort                 int
}

type KeysFile struct {
	EvmPrivateKey   EvmPrivateKey
	EvmPublicKey    EvmPublisherKey
	StarkPrivateKey StarkPrivateKey
	StarkPublicKey  StarkPublisherKey
	OracleId        OracleId
	StorkAuth       AuthToken
	PullBasedAuth   AuthToken
}

func readFile(path string) ([]byte, error) {
	file, err := os.Open(path)
	defer file.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return data, nil
}

func LoadConfig(configFilePath string, keysFilePath string) (*StorkPublisherAgentConfig, error) {
	configFileData, err := readFile(configFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	keyFileData, err := readFile(keysFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read keys file: %w", err)
	}

	var configFile ConfigFile
	err = json.Unmarshal(configFileData, &configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config file: %w", err)
	}

	var keysFile KeysFile
	err = json.Unmarshal(keyFileData, &keysFile)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal keys file: %w", err)
	}

	// validate config and key file
	if configFile.SignatureTypes == nil || len(configFile.SignatureTypes) == 0 {
		return nil, fmt.Errorf("must specify at least one signatureType")
	}
	for _, signatureType := range configFile.SignatureTypes {
		switch signatureType {
		case EvmSignatureType:
			if !Hex32Regex.MatchString(string(keysFile.EvmPrivateKey)) {
				return nil, errors.New("must pass a valid EVM private key")
			}
			if !Hex32Regex.MatchString(string(keysFile.EvmPublicKey)) {
				return nil, errors.New("must pass a valid EVM public key")
			}
		case StarkSignatureType:
			if !Hex32Regex.MatchString(string(keysFile.StarkPrivateKey)) {
				return nil, errors.New("must pass a valid Stark private key")
			}
			if !Hex32Regex.MatchString(string(keysFile.StarkPublicKey)) {
				return nil, errors.New("must pass a valid Stark public key")
			}
		default:
			return nil, fmt.Errorf("invalid signature type: %s", signatureType)
		}
	}

	if len(keysFile.OracleId) != 5 {
		return nil, errors.New("oracle id length must be 5")
	}

	clockPeriodStr := configFile.ClockPeriod
	if len(clockPeriodStr) == 0 {
		clockPeriodStr = DefaultClockUpdatePeriod
	}
	clockUpdatePeriod, err := time.ParseDuration(clockPeriodStr)
	if err != nil {
		return nil, fmt.Errorf("invalid clock update period: %s", clockPeriodStr)
	}

	deltaCheckPeriodStr := configFile.DeltaCheckPeriod
	if len(deltaCheckPeriodStr) == 0 {
		deltaCheckPeriodStr = DefaultDeltaUpdatePeriod
	}
	deltaUpdatePeriod, err := time.ParseDuration(deltaCheckPeriodStr)
	if err != nil {
		return nil, fmt.Errorf("invalid delta update period: %s", deltaCheckPeriodStr)
	}
	if deltaUpdatePeriod.Nanoseconds() == 0 {
		return nil, errors.New("delta update period must be positive")
	}

	storkRegistryRefreshIntervalStr := configFile.StorkRegistryRefreshInterval
	if len(storkRegistryRefreshIntervalStr) == 0 {
		storkRegistryRefreshIntervalStr = DefaultStorkRegistryRefreshInterval
	}
	storkRegistryRefreshDuration, err := time.ParseDuration(storkRegistryRefreshIntervalStr)
	if err != nil {
		return nil, fmt.Errorf("invalid stork registry refresh duration: %s", storkRegistryRefreshIntervalStr)
	}
	if storkRegistryRefreshDuration.Nanoseconds() == 0 {
		return nil, errors.New("stork registry refresh duration must be positive")
	}

	brokerReconnectDelayStr := configFile.BrokerReconnectDelay
	if len(brokerReconnectDelayStr) == 0 {
		brokerReconnectDelayStr = DefaultBrokerReconnectDelay
	}
	brokerReconnectDelayDuration, err := time.ParseDuration(brokerReconnectDelayStr)
	if err != nil {
		return nil, fmt.Errorf("invalid broker reconnect duration: %s", brokerReconnectDelayStr)
	}
	if brokerReconnectDelayDuration.Nanoseconds() == 0 {
		return nil, errors.New("broker reconnect duration must be positive")
	}

	changeThresholdPercent := configFile.ChangeThresholdPercent
	if changeThresholdPercent == 0 {
		changeThresholdPercent = DefaultChangeThresholdPercent
	}
	if changeThresholdPercent <= 0 {
		return nil, errors.New("change threshold percent must be positive")
	}

	if configFile.IncomingWsPort > 65535 {
		return nil, errors.New("incoming ws port must be between 0 and 65535")
	}

	if configFile.IncomingWsPort == 0 && len(configFile.PullBasedWsUrl) == 0 {
		return nil, errors.New("must specify an incoming ws url to pull from or a port to expose for our incoming ws")
	}

	pullBasedReconnectDelayStr := configFile.PullBasedWsReconnectDelay
	if len(pullBasedReconnectDelayStr) == 0 {
		pullBasedReconnectDelayStr = DefaultPullBasedReconnectDelay
	}
	pullBasedReconnectDuration, err := time.ParseDuration(pullBasedReconnectDelayStr)
	if err != nil {
		return nil, fmt.Errorf("invalid pull-based websocket reconnect period: %s", pullBasedReconnectDelayStr)
	}

	storkRegistryBaseUrl := configFile.StorkRegistryBaseUrl
	if len(storkRegistryBaseUrl) == 0 {
		storkRegistryBaseUrl = DefaultStorkRegistryBaseUrl
	}

	config := NewStorkPublisherAgentConfig(
		configFile.SignatureTypes,
		keysFile.EvmPrivateKey,
		keysFile.EvmPublicKey,
		keysFile.StarkPrivateKey,
		keysFile.StarkPublicKey,
		clockUpdatePeriod,
		deltaUpdatePeriod,
		changeThresholdPercent,
		keysFile.OracleId,
		storkRegistryBaseUrl,
		storkRegistryRefreshDuration,
		brokerReconnectDelayDuration,
		keysFile.StorkAuth,
		configFile.PullBasedWsUrl,
		keysFile.PullBasedAuth,
		configFile.PullBasedWsSubscriptionRequest,
		pullBasedReconnectDuration,
		configFile.SignEveryUpdate,
		configFile.IncomingWsPort,
	)

	return config, nil
}

type StorkPublisherAgentConfig struct {
	SignatureTypes                 []SignatureType
	EvmPrivateKey                  EvmPrivateKey
	EvmPublicKey                   EvmPublisherKey
	StarkPrivateKey                StarkPrivateKey
	StarkPublicKey                 StarkPublisherKey
	ClockPeriod                    time.Duration
	DeltaCheckPeriod               time.Duration
	ChangeThresholdProportion      float64 // 0-1
	OracleId                       OracleId
	StorkRegistryBaseUrl           string
	StorkAuth                      AuthToken
	StorkRegistryRefreshInterval   time.Duration
	BrokerReconnectDelay           time.Duration
	PullBasedWsUrl                 string
	PullBasedAuth                  AuthToken
	PullBasedWsSubscriptionRequest string
	PullBasedWsReconnectDelay      time.Duration
	SignEveryUpdate                bool
	IncomingWsPort                 int
}

func NewStorkPublisherAgentConfig(
	signatureTypes []SignatureType,
	evmPrivateKey EvmPrivateKey,
	evmPublisherKey EvmPublisherKey,
	starkPrivateKey StarkPrivateKey,
	starkPublisherKey StarkPublisherKey,
	clockPeriod time.Duration,
	deltaPeriod time.Duration,
	changeThresholdPercentage float64,
	oracleId OracleId,
	storkRegistryBaseUrl string,
	storkRegistryRefreshInterval time.Duration,
	brokerReconnectDelay time.Duration,
	storkAuth AuthToken,
	pullBasedWsUrl string,
	pullBasedAuth AuthToken,
	pullBasedWsSubscriptionRequest string,
	pullBasedWsReconnectDelay time.Duration,
	signEveryUpdate bool,
	incomingWsPort int,
) *StorkPublisherAgentConfig {
	return &StorkPublisherAgentConfig{
		SignatureTypes:                 signatureTypes,
		EvmPrivateKey:                  evmPrivateKey,
		EvmPublicKey:                   evmPublisherKey,
		StarkPrivateKey:                starkPrivateKey,
		StarkPublicKey:                 starkPublisherKey,
		ClockPeriod:                    clockPeriod,
		DeltaCheckPeriod:               deltaPeriod,
		ChangeThresholdProportion:      changeThresholdPercentage / 100.0,
		OracleId:                       oracleId,
		StorkRegistryBaseUrl:           storkRegistryBaseUrl,
		StorkRegistryRefreshInterval:   storkRegistryRefreshInterval,
		BrokerReconnectDelay:           brokerReconnectDelay,
		StorkAuth:                      storkAuth,
		PullBasedWsUrl:                 pullBasedWsUrl,
		PullBasedAuth:                  pullBasedAuth,
		PullBasedWsSubscriptionRequest: pullBasedWsSubscriptionRequest,
		PullBasedWsReconnectDelay:      pullBasedWsReconnectDelay,
		SignEveryUpdate:                signEveryUpdate,
		IncomingWsPort:                 incomingWsPort,
	}
}
