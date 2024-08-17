package main

import "time"

type StorkPublisherAgentConfig struct {
	signatureType             SignatureType
	clockPeriod               time.Duration
	deltaCheckPeriod          time.Duration
	changeThresholdProportion float64 // 0-1
	oracleId                  OracleId
	publisherKey              PublisherKey
	httpPort                  int
	enforceCompression        bool
}
