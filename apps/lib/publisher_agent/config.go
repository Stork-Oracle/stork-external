package publisher_agent

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"time"

	"github.com/Stork-Oracle/stork-external/apps/lib/signer"
)

var Hex32Regex = regexp.MustCompile(`^0x[0-9a-fA-F]+$`)
var StorkAuthRegex = regexp.MustCompile(`^[A-Za-z0-9+/]+={0,2}$`)

const DefaultClockUpdatePeriod = "500ms"
const DefaultDeltaUpdatePeriod = "10ms"
const DefaultChangeThresholdPercent = 0.1
const DefaultStorkRegistryRefreshInterval = "10m"
const DefaultPublisherMetadataRefreshInterval = "1h"
const DefaultStorkRegistryBaseUrl = "https://rest.jp.stork-oracle.network"
const DefaultPublisherMetadataBaseUrl = "https://rest.jp.stork-oracle.network"
const DefaultBrokerReconnectDelay = "5s"
const DefaultPullBasedReconnectDelay = "5s"
const DefaultPullBasedReadTimeout = "10s"

type Config struct {
	SignatureTypes                   []signer.SignatureType
	ClockPeriod                      string
	DeltaCheckPeriod                 string
	ChangeThresholdPercent           float64 // 0-100
	StorkRegistryBaseUrl             string
	StorkRegistryRefreshInterval     string
	BrokerReconnectDelay             string
	PublisherMetadataRefreshInterval string
	PublisherMetadataBaseUrl         string
	PullBasedWsUrl                   string
	PullBasedWsSubscriptionRequest   string
	PullBasedWsReconnectDelay        string
	PullBasedWsReadTimeout           string
	SignEveryUpdate                  bool
	IncomingWsPort                   int
}

type Keys struct {
	EvmPrivateKey   signer.EvmPrivateKey
	EvmPublicKey    signer.EvmPublisherKey
	StarkPrivateKey signer.StarkPrivateKey
	StarkPublicKey  signer.StarkPublisherKey
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

	keysFileData, keysFileReadErr := readFile(keysFilePath)

	var configFile Config
	err = json.Unmarshal(configFileData, &configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config file: %w", err)
	}

	var keys Keys

	// only deserialize keysFileData if keysFilePath was successfully read
	if keysFileReadErr == nil {
		err = json.Unmarshal(keysFileData, &keys)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal keys file: %w", err)
		}
	}

	// overwrite keys with any set env vars
	loadKeysFromEnv(&keys)

	// validate config and keys file
	if configFile.SignatureTypes == nil || len(configFile.SignatureTypes) == 0 {
		return nil, fmt.Errorf("must specify at least one signatureType")
	}

	for _, signatureType := range configFile.SignatureTypes {
		switch signatureType {
		case EvmSignatureType:
			if !Hex32Regex.MatchString(string(keys.EvmPrivateKey)) {
				return nil, errors.New("must pass a valid EVM private key")
			}
			if !Hex32Regex.MatchString(string(keys.EvmPublicKey)) {
				return nil, errors.New("must pass a valid EVM public key")
			}
		case StarkSignatureType:
			if !Hex32Regex.MatchString(string(keys.StarkPrivateKey)) {
				return nil, errors.New("must pass a valid Stark private key")
			}
			if !Hex32Regex.MatchString(string(keys.StarkPublicKey)) {
				return nil, errors.New("must pass a valid Stark public key")
			}
		default:
			return nil, fmt.Errorf("invalid signature type: %s", signatureType)
		}
	}

	if !StorkAuthRegex.MatchString(string(keys.StorkAuth)) {
		return nil, errors.New("stork auth token must a non-empty base64 string")
	}

	if len(keys.OracleId) != 5 {
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

	publisherMetadataUpdateIntervalStr := configFile.PublisherMetadataRefreshInterval
	if len(publisherMetadataUpdateIntervalStr) == 0 {
		publisherMetadataUpdateIntervalStr = DefaultPublisherMetadataRefreshInterval
	}
	publisherMetadataUpdateDuration, err := time.ParseDuration(publisherMetadataUpdateIntervalStr)
	if err != nil {
		return nil, fmt.Errorf("invalid publisher metadata update duration: %s", publisherMetadataUpdateIntervalStr)
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

	pullBasedWsReadTimeoutStr := configFile.PullBasedWsReadTimeout
	if len(pullBasedWsReadTimeoutStr) == 0 {
		pullBasedWsReadTimeoutStr = DefaultPullBasedReadTimeout
	}
	pullBasedWsReadTimeout, err := time.ParseDuration(pullBasedWsReadTimeoutStr)
	if err != nil {
		return nil, fmt.Errorf("invalid pull-based websocket read timeout: %s", pullBasedWsReadTimeoutStr)
	}

	storkRegistryBaseUrl := configFile.StorkRegistryBaseUrl
	if len(storkRegistryBaseUrl) == 0 {
		storkRegistryBaseUrl = DefaultStorkRegistryBaseUrl
	}

	publisherMetadataBaseUrl := configFile.PublisherMetadataBaseUrl
	if len(publisherMetadataBaseUrl) == 0 {
		publisherMetadataBaseUrl = DefaultPublisherMetadataBaseUrl
	}

	config := NewStorkPublisherAgentConfig(
		configFile.SignatureTypes,
		keys.EvmPrivateKey,
		keys.EvmPublicKey,
		keys.StarkPrivateKey,
		keys.StarkPublicKey,
		clockUpdatePeriod,
		deltaUpdatePeriod,
		changeThresholdPercent,
		keys.OracleId,
		storkRegistryBaseUrl,
		storkRegistryRefreshDuration,
		brokerReconnectDelayDuration,
		publisherMetadataBaseUrl,
		publisherMetadataUpdateDuration,
		keys.StorkAuth,
		configFile.PullBasedWsUrl,
		keys.PullBasedAuth,
		configFile.PullBasedWsSubscriptionRequest,
		pullBasedReconnectDuration,
		pullBasedWsReadTimeout,
		configFile.SignEveryUpdate,
		configFile.IncomingWsPort,
	)

	return config, nil
}

// this overwrites any existing values in keys
func loadKeysFromEnv(keys *Keys) {
	evmPrivateKey := os.Getenv("STORK_EVM_PRIVATE_KEY")
	if evmPrivateKey != "" {
		keys.EvmPrivateKey = signer.EvmPrivateKey(evmPrivateKey)
	}
	evmPublicKey := os.Getenv("STORK_EVM_PUBLIC_KEY")
	if evmPublicKey != "" {
		keys.EvmPublicKey = signer.EvmPublisherKey(evmPublicKey)
	}
	starkPrivateKey := os.Getenv("STORK_STARK_PRIVATE_KEY")
	if starkPrivateKey != "" {
		keys.StarkPrivateKey = signer.StarkPrivateKey(starkPrivateKey)
	}
	starkPublicKey := os.Getenv("STORK_STARK_PUBLIC_KEY")
	if starkPublicKey != "" {
		keys.StarkPublicKey = signer.StarkPublisherKey(starkPublicKey)
	}
	oracleId := os.Getenv("STORK_ORACLE_ID")
	if oracleId != "" {
		keys.OracleId = OracleId(oracleId)
	}
	storkAuth := os.Getenv("STORK_AUTH")
	if storkAuth != "" {
		keys.StorkAuth = AuthToken(storkAuth)
	}
	pullBasedAuth := os.Getenv("STORK_PULL_BASED_AUTH")
	if pullBasedAuth != "" {
		keys.PullBasedAuth = AuthToken(pullBasedAuth)
	}
}

type StorkPublisherAgentConfig struct {
	SignatureTypes                  []signer.SignatureType
	EvmPrivateKey                   signer.EvmPrivateKey
	EvmPublicKey                    signer.EvmPublisherKey
	StarkPrivateKey                 signer.StarkPrivateKey
	StarkPublicKey                  signer.StarkPublisherKey
	ClockPeriod                     time.Duration
	DeltaCheckPeriod                time.Duration
	ChangeThresholdProportion       float64 // 0-1
	OracleId                        OracleId
	StorkRegistryBaseUrl            string
	StorkAuth                       AuthToken
	StorkRegistryRefreshInterval    time.Duration
	BrokerReconnectDelay            time.Duration
	PublisherMetadataBaseUrl        string
	PublisherMetadataUpdateInterval time.Duration
	PullBasedWsUrl                  string
	PullBasedAuth                   AuthToken
	PullBasedWsSubscriptionRequest  string
	PullBasedWsReconnectDelay       time.Duration
	PullBasedWsReadTimeout          time.Duration
	SignEveryUpdate                 bool
	IncomingWsPort                  int
}

func NewStorkPublisherAgentConfig(
	signatureTypes []signer.SignatureType,
	evmPrivateKey signer.EvmPrivateKey,
	evmPublisherKey signer.EvmPublisherKey,
	starkPrivateKey signer.StarkPrivateKey,
	starkPublisherKey signer.StarkPublisherKey,
	clockPeriod time.Duration,
	deltaPeriod time.Duration,
	changeThresholdPercentage float64,
	oracleId OracleId,
	storkRegistryBaseUrl string,
	storkRegistryRefreshInterval time.Duration,
	brokerReconnectDelay time.Duration,
	publisherMetadataBaseUrl string,
	publisherMetadataUpdateInterval time.Duration,
	storkAuth AuthToken,
	pullBasedWsUrl string,
	pullBasedAuth AuthToken,
	pullBasedWsSubscriptionRequest string,
	pullBasedWsReconnectDelay time.Duration,
	pullBasedWsReadTimeout time.Duration,
	signEveryUpdate bool,
	incomingWsPort int,
) *StorkPublisherAgentConfig {
	return &StorkPublisherAgentConfig{
		SignatureTypes:                  signatureTypes,
		EvmPrivateKey:                   evmPrivateKey,
		EvmPublicKey:                    evmPublisherKey,
		StarkPrivateKey:                 starkPrivateKey,
		StarkPublicKey:                  starkPublisherKey,
		ClockPeriod:                     clockPeriod,
		DeltaCheckPeriod:                deltaPeriod,
		ChangeThresholdProportion:       changeThresholdPercentage / 100.0,
		OracleId:                        oracleId,
		StorkRegistryBaseUrl:            storkRegistryBaseUrl,
		StorkRegistryRefreshInterval:    storkRegistryRefreshInterval,
		BrokerReconnectDelay:            brokerReconnectDelay,
		PublisherMetadataBaseUrl:        publisherMetadataBaseUrl,
		PublisherMetadataUpdateInterval: publisherMetadataUpdateInterval,
		StorkAuth:                       storkAuth,
		PullBasedWsUrl:                  pullBasedWsUrl,
		PullBasedAuth:                   pullBasedAuth,
		PullBasedWsSubscriptionRequest:  pullBasedWsSubscriptionRequest,
		PullBasedWsReconnectDelay:       pullBasedWsReconnectDelay,
		PullBasedWsReadTimeout:          pullBasedWsReadTimeout,
		SignEveryUpdate:                 signEveryUpdate,
		IncomingWsPort:                  incomingWsPort,
	}
}
