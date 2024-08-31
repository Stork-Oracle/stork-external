package stork_publisher_agent

import (
	"time"
)

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
	}
}
