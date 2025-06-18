package publisher_agent

import (
	"fmt"
	"time"

	"github.com/DataDog/datadog-go/statsd"
	"github.com/Stork-Oracle/stork-external/apps/lib/signer"
	"github.com/rs/zerolog"
)

const (
	datadogRate          = 1
	datadogMonitorPeriod = 10 * time.Second

	publisherKeyTagFormat  = "publisher_key:%s"
	SignatureTypeTagFormat = "signature_type:%s"

	SignedUpdateCountMetric       = "stork.publisher_agent.signer.signed.update.count"
	LatestSignedUpdateAgeNsMetric = "stork.publisher_agent.signer.signed.update.age.latest.ns"
	SignQueueSizeMetric           = "stork.publisher_agent.signer.queue.size"

	OutgoingUpdateCountMetric = "stork.publisher_agent.outgoing.update.count"
	SentUpdateCountMetric     = "stork.publisher_agent.sent.update.count"
	SentMessageCountMetric    = "stork.publisher_agent.sent.message.count"

	IncomingPullerUpdateCountMetric = "stork.publisher_agent.incoming.puller.update.count"
	IncomingPusherUpdateCountMetric = "stork.publisher_agent.incoming.pusher.update.count"
)

func getPublisherKeyTag(publisherKey signer.PublisherKey) string {
	return fmt.Sprintf(publisherKeyTagFormat, publisherKey)
}

func getSignatureTypeTag(signatureType signer.SignatureType) string {
	return fmt.Sprintf(SignatureTypeTagFormat, signatureType)
}

func reportDatadogCount(metric string, value int64, tags []string, datadogClient *statsd.Client, logger zerolog.Logger) {
	if datadogClient == nil {
		logger.Info().Msg("datadog client not initialized")
		return
	}

	err := datadogClient.Count(metric, value, tags, datadogRate)
	if err != nil {
		logger.Warn().Err(err).Msg("failed to send datadog metric")
	}
}

func reportDatadogGauge(metric string, value float64, tags []string, datadogClient *statsd.Client, logger zerolog.Logger) {
	if datadogClient == nil {
		return
	}

	err := datadogClient.Gauge(metric, value, tags, datadogRate)
	if err != nil {
		logger.Warn().Err(err).Msg("failed to send datadog metric")
	}
}
