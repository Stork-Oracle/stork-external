package publisher_agent

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/Stork-Oracle/stork-external/apps/lib/signer"
	"github.com/rs/zerolog"
)

type RegistryClient struct {
	baseUrl   string
	authToken AuthToken
	logger    zerolog.Logger
}

func NewRegistryClient(baseUrl string, authToken AuthToken, logger zerolog.Logger) *RegistryClient {
	return &RegistryClient{
		baseUrl:   baseUrl,
		authToken: authToken,
		logger:    logger,
	}
}

func (c *RegistryClient) GetBrokersForPublisher(publisherKey signer.PublisherKey) (map[BrokerPublishUrl]map[AssetId]struct{}, error) {
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
		// check if the registry returned an error
		var errorResponse RegistryErrorResponse
		err = json.Unmarshal(response, &errorResponse)
		if err == nil {
			if errorResponse.Error == "Unauthorized" {
				return nil, fmt.Errorf("not authorized to query Stork Registry - check your configured StorkAuth")
			} else {
				return nil, fmt.Errorf("failed to query the Stork Registry: %s", errorResponse.Error)
			}
		}
		return nil, fmt.Errorf("failed to unmarshal response from Stork Registry: %s", string(response))
	}

	brokerMap := make(map[BrokerPublishUrl]map[AssetId]struct{})

	if len(brokers) == 0 {
		c.logger.Warn().Msgf("no stork registry broker found for public key (%s) - reach out to Stork to make sure you're whitelisted", publisherKey)
		return brokerMap, nil
	}

	// combine all configs into a single asset map per url
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
