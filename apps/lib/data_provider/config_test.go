package data_provider

import (
	"testing"

	"github.com/Stork-Oracle/stork-external/apps/lib/data_provider/utils"
	"github.com/stretchr/testify/assert"
)

func TestValidConfig(t *testing.T) {
	configStr := `
		{
		  "sources": [
			{
			  "id": "WETHUSDT",
			  "dataSource": "uniswap_v2",
			  "config": {
				"updateFrequency": "5s",
				"contractAddress": "0x0d4a11d5eeaac28ec3f61d100daf4d40471f1852",
				"httpProviderUrl": "https://eth-mainnet.g.alchemy.com/v2/",
				"providerApiKeyEnvVar": "ALCHEMY_API_KEY",
				"baseTokenIndex": 0,
				"baseTokenDecimals": 18,
				"quoteTokenIndex": 1,
				"quoteTokenDecimals": 6
			  }
			},
			{
			  "id": "PEPEWETH",
			  "dataSource": "uniswap_v2",
			  "config": {
				"updateFrequency": "5s",
				"contractAddress": "0xa43fe16908251ee70ef74718545e4fe6c5ccec9f",
				"httpProviderUrl": "https://eth-mainnet.g.alchemy.com/v2/",
				"providerApiKeyEnvVar": "ALCHEMY_API_KEY",
				"baseTokenIndex": 0,
				"baseTokenDecimals": 18,
				"quoteTokenIndex": 1,
				"quoteTokenDecimals": 18
			  }
			},
			{
			  "id": "MY_RANDOM_VALUE",
			  "dataSource": "random",
			  "config": {
				"updateFrequency": "1s",
				"minValue": 2500,
				"maxValue": 3000
			  }
			}
		  ]
		}`

	schema, err := utils.LoadSchema(resourcesFS)
	assert.NoError(t, err)
	err = utils.ValidateConfig([]byte(configStr), schema)
	assert.NoError(t, err)
}

func TestInvalidTopLevelConfigs(t *testing.T) {
	// unexpected field
	configStr := `
		{
	      "extraField": "",
		  "sources": [
			{
			  "id": "MY_RANDOM_VALUE",
			  "dataSource": "random",
			  "config": {
				"updateFrequency": "1s",
				"minValue": 2500,
				"maxValue": 3000
			  }
			}
		  ]
		}`
	schema, err := utils.LoadSchema(resourcesFS)
	assert.NoError(t, err)
	err = utils.ValidateConfig([]byte(configStr), schema)
	assert.ErrorContains(t, err, "Additional property extraField is not allowed")

	// missing field
	configStr = `{}`
	err = utils.ValidateConfig([]byte(configStr), schema)
	assert.ErrorContains(t, err, "sources is required")

	// empty source list
	configStr = `{
		"sources": []
	}`
	err = utils.ValidateConfig([]byte(configStr), schema)
	assert.ErrorContains(t, err, "sources: Array must have at least 1 items")

	// incorrect type
	configStr = `
		{
		  "sources": [
			{
			  "id": 17,
			  "dataSource": "random",
			  "config": {
				"updateFrequency": "1s",
				"minValue": 2500,
				"maxValue": 3000
			  }
			}
		  ]
		}`
	err = utils.ValidateConfig([]byte(configStr), schema)
	assert.ErrorContains(t, err, "Expected: string, given: integer")

	// invalid json
	configStr = `abcde`
	err = utils.ValidateConfig([]byte(configStr), schema)
	assert.ErrorContains(t, err, "failed to parse config JSON")

	// invalid value
	configStr = `
		{
		  "sources": [
			{
			  "id": "MY_RANDOM_VALUE",
			  "dataSource": "random",
			  "config": {
				"updateFrequency": "five_minutes",
				"minValue": 2500,
				"maxValue": 3000
			  }
			}
		  ]
		}`
	err = utils.ValidateConfig([]byte(configStr), schema)
	assert.ErrorContains(t, err, "updateFrequency: Does not match pattern")

	// unexpected field
	configStr = `
		{
		  "sources": [
			{
			  "id": "MY_RANDOM_VALUE",
			  "dataSource": "random",
			  "config": {
				"updateFrequency": "5s",
				"minValue": 2500,
				"maxValue": 3000,
				"extraSourceConfigField": 123
			  }
			}
		  ]
		}`
	err = utils.ValidateConfig([]byte(configStr), schema)
	assert.ErrorContains(t, err, "Additional property extraSourceConfigField is not allowed")

	// invalid data source
	configStr = `
		{
		  "sources": [
			{
			  "id": "MY_RANDOM_VALUE",
			  "dataSource": "fake_data_source",
			  "config": {
				"updateFrequency": "5s",
				"minValue": 2500,
				"maxValue": 3000
			  }
			}
		  ]
		}`
	err = utils.ValidateConfig([]byte(configStr), schema)
	assert.ErrorContains(t, err, "sources.0.dataSource must be one of the following")
}
