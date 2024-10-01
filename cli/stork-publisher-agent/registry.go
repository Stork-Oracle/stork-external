package stork_publisher_agent

import (
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/rs/zerolog"
)

type RegistryClient[T Signature] struct {
	baseUrl string
	signer  Signer[T]
	logger  zerolog.Logger
}

func NewRegistryClient[T Signature](baseUrl string, signer Signer[T], logger zerolog.Logger) *RegistryClient[T] {
	return &RegistryClient[T]{
		baseUrl: baseUrl,
		signer:  signer,
		logger:  logger,
	}
}

func (c *RegistryClient[T]) GetBrokersForPublisher(publisherKey PublisherKey) (map[BrokerPublishUrl]map[AssetId]struct{}, error) {
	brokerEndpoint := c.baseUrl + "/v1/registry/brokers"
	queryParams := url.Values{}
	queryParams.Add("publisher_key", string(publisherKey))

	nowNs := time.Now().UnixNano()
	connectionHeaders, err := c.signer.getConnectionHeaders(nowNs, publisherKey)
	if err != nil {
		c.logger.Error().Err(err).Msg("failed to generate connection headers")
	}

	response, err := RestQuery("GET", brokerEndpoint, queryParams, nil, connectionHeaders)
	if err != nil {
		return nil, fmt.Errorf("failed to get broker list: %v", err)
	}
	var brokers []BrokerConnectionConfig
	err = json.Unmarshal(response, &brokers)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal broker list: %v", err)
	}

	if len(brokers) == 0 {
		c.logger.Warn().Msgf("broker list is empty - make sure your publisher key (%s) is correct and that Stork has registered it", publisherKey)
		emptyMap := make(map[BrokerPublishUrl]map[AssetId]struct{})
		return emptyMap, nil
	}

	// combine all configs into a single asset map per url
	brokerMap := make(map[BrokerPublishUrl]map[AssetId]struct{})
	for _, broker := range brokers {
		assetIds, exists := brokerMap[broker.PublishUrl]

		if !exists {
			assetIds = make(map[AssetId]struct{})
			brokerMap[broker.PublishUrl] = assetIds
		}

		for _, asset := range broker.AssetIds {
			assetIds[asset] = struct{}{}
		}
	}

	return brokerMap, nil
}
