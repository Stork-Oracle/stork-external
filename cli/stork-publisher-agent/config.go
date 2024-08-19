package stork_publisher_agent

import (
	"time"
)

type StorkPublisherAgentConfig struct {
	SignatureType             SignatureType
	PrivateKey                PrivateKey
	ClockPeriod               time.Duration
	DeltaCheckPeriod          time.Duration
	ChangeThresholdProportion float64 // 0-1
	OracleId                  OracleId
	EnforceCompression        bool
}

func NewStorkPublisherAgentConfig(
	signatureType SignatureType,
	privateKey PrivateKey,
	clockPeriod time.Duration,
	deltaPeriod time.Duration,
	changeThresholdPercentage float64,
	oracleId OracleId,
	enforceCompression bool,
) *StorkPublisherAgentConfig {
	return &StorkPublisherAgentConfig{
		SignatureType:             signatureType,
		PrivateKey:                privateKey,
		ClockPeriod:               clockPeriod,
		DeltaCheckPeriod:          deltaPeriod,
		ChangeThresholdProportion: changeThresholdPercentage / 100.0,
		OracleId:                  oracleId,
		EnforceCompression:        enforceCompression,
	}
}
