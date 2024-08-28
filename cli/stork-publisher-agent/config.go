package stork_publisher_agent

import (
	"time"
)

type StorkPublisherAgentConfig struct {
	SignatureType                  SignatureType
	PrivateKey                     PrivateKey
	PublicKey                      PublisherKey
	ClockPeriod                    time.Duration
	DeltaCheckPeriod               time.Duration
	ChangeThresholdProportion      float64 // 0-1
	OracleId                       OracleId
	StorkRegistryBaseUrl           string
	StorkAuth                      AuthToken
	PullBasedWsUrl                 string
	PullBasedAuth                  AuthToken
	PullBasedWsSubscriptionRequest string
	PullBasedWsReconnectDelay      time.Duration
	SignEveryUpdate                bool
}

func NewStorkPublisherAgentConfig(
	signatureType SignatureType,
	privateKey PrivateKey,
	PublicKey PublisherKey,
	clockPeriod time.Duration,
	deltaPeriod time.Duration,
	changeThresholdPercentage float64,
	oracleId OracleId,
	storkRegistryBaseUrl string,
	storkAuth AuthToken,
	pullBasedWsUrl string,
	pullBasedAuth AuthToken,
	pullBasedWsSubscriptionRequest string,
	pullBasedWsReconnectDelay time.Duration,
	signEveryUpdate bool,
) *StorkPublisherAgentConfig {
	return &StorkPublisherAgentConfig{
		SignatureType:                  signatureType,
		PrivateKey:                     privateKey,
		PublicKey:                      PublicKey,
		ClockPeriod:                    clockPeriod,
		DeltaCheckPeriod:               deltaPeriod,
		ChangeThresholdProportion:      changeThresholdPercentage / 100.0,
		OracleId:                       oracleId,
		StorkRegistryBaseUrl:           storkRegistryBaseUrl,
		StorkAuth:                      storkAuth,
		PullBasedWsUrl:                 pullBasedWsUrl,
		PullBasedAuth:                  pullBasedAuth,
		PullBasedWsSubscriptionRequest: pullBasedWsSubscriptionRequest,
		PullBasedWsReconnectDelay:      pullBasedWsReconnectDelay,
		SignEveryUpdate:                signEveryUpdate,
	}
}
