package stork_publisher_agent

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type RegistryClient struct {
	baseUrl   string
	authToken AuthToken
}

func NewRegistryClient(baseUrl string, authToken AuthToken) *RegistryClient {
	return &RegistryClient{
		baseUrl:   baseUrl,
		authToken: authToken,
	}
}

func (c *RegistryClient) GetBrokersForPublisher(publisherKey PublisherKey) (map[BrokerPublishUrl]map[AssetId]struct{}, error) {
	brokerEndpoint := c.baseUrl + "/v1/registry/brokers"
	queryParams := url.Values{}
	queryParams.Add("publisher_key", string(publisherKey))
	authHeader := http.Header{"Authorization": []string{"Basic " + string(c.authToken)}}
	response, err := RestQuery("GET", brokerEndpoint, queryParams, nil, authHeader)
	if err != nil {
		return nil, fmt.Errorf("failed to get broker list: %v", err)
	}
	var brokers []BrokerConnectionConfig
	err = json.Unmarshal(response, &brokers)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal broker list: %v", err)
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
