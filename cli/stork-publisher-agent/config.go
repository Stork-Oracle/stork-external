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
	EnforceCompression             bool
	PullBasedWsUrl                 string
	PullBasedAuth                  string
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
	enforceCompression bool,
	pullBasedWsUrl string,
	pullBasedAuth string,
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
		EnforceCompression:             enforceCompression,
		PullBasedWsUrl:                 pullBasedWsUrl,
		PullBasedAuth:                  pullBasedAuth,
		PullBasedWsSubscriptionRequest: pullBasedWsSubscriptionRequest,
		PullBasedWsReconnectDelay:      pullBasedWsReconnectDelay,
		SignEveryUpdate:                signEveryUpdate,
	}
}
