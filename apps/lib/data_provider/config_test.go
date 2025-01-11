package data_provider

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidConfig(t *testing.T) {
	configStr := `
		{
		  "sources": [
			{
			  "id": "WETHUSDT",
			  "dataSource": "UNISWAP_V2",
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
			  "dataSource": "UNISWAP_V2",
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
			  "dataSource": "RANDOM_NUMBER",
			  "config": {
				"updateFrequency": "1s",
				"minValue": 2500,
				"maxValue": 3000
			  }
			}
		  ]
		}`

	err := validateConfig([]byte(configStr))
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
			  "dataSource": "RANDOM_NUMBER",
			  "config": {
				"updateFrequency": "1s",
				"minValue": 2500,
				"maxValue": 3000
			  }
			}
		  ]
		}`
	err := validateConfig([]byte(configStr))
	assert.ErrorContains(t, err, "Additional property extraField is not allowed")

	// missing field
	configStr = `{}`
	err = validateConfig([]byte(configStr))
	assert.ErrorContains(t, err, "sources is required")

	// empty source list
	configStr = `{
		"sources": []
	}`
	err = validateConfig([]byte(configStr))
	assert.ErrorContains(t, err, "sources: Array must have at least 1 items")

	// incorrect type
	configStr = `
		{
		  "sources": [
			{
			  "id": 17,
			  "dataSource": "RANDOM_NUMBER",
			  "config": {
				"updateFrequency": "1s",
				"minValue": 2500,
				"maxValue": 3000
			  }
			}
		  ]
		}`
	err = validateConfig([]byte(configStr))
	assert.ErrorContains(t, err, "Expected: string, given: integer")

	// invalid json
	configStr = `abcde`
	err = validateConfig([]byte(configStr))
	assert.ErrorContains(t, err, "failed to parse config JSON")
}

func TestInvalidSourceConfigs(t *testing.T) {
	// invalid value
	configStr := `
		{
		  "sources": [
			{
			  "id": "MY_RANDOM_VALUE",
			  "dataSource": "RANDOM_NUMBER",
			  "config": {
				"updateFrequency": "five_minutes",
				"minValue": 2500,
				"maxValue": 3000
			  }
			}
		  ]
		}`
	err := validateConfig([]byte(configStr))
	assert.ErrorContains(t, err, "updateFrequency: Does not match pattern")

	// unexpected field
	configStr = `
		{
		  "sources": [
			{
			  "id": "MY_RANDOM_VALUE",
			  "dataSource": "RANDOM_NUMBER",
			  "config": {
				"updateFrequency": "5s",
				"minValue": 2500,
				"maxValue": 3000,
				"extraSourceConfigField": 123
			  }
			}
		  ]
		}`
	err = validateConfig([]byte(configStr))
	assert.ErrorContains(t, err, "Additional property extraSourceConfigField is not allowed")

	// mismatched data source
	configStr = `
		{
		  "sources": [
			{
			  "id": "MY_RANDOM_VALUE",
			  "dataSource": "UNISWAP_V2",
			  "config": {
				"updateFrequency": "5s",
				"minValue": 2500,
				"maxValue": 3000
			  }
			}
		  ]
		}`
	err = validateConfig([]byte(configStr))
	assert.ErrorContains(t, err, "httpProviderUrl is required")

	// invalid data source
	configStr = `
		{
		  "sources": [
			{
			  "id": "MY_RANDOM_VALUE",
			  "dataSource": "FAKE_DATA_SOURCE",
			  "config": {
				"updateFrequency": "5s",
				"minValue": 2500,
				"maxValue": 3000
			  }
			}
		  ]
		}`
	err = validateConfig([]byte(configStr))
	assert.ErrorContains(t, err, "no factory registered for: FAKE_DATA_SOURCE")
}
