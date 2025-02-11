package configs

import (
	"testing"

	"github.com/Stork-Oracle/stork-external/apps/lib/data_provider/types"
	"github.com/stretchr/testify/assert"
)

func TestValidConfig(t *testing.T) {
	configStr := `
		{
		  "sources": [
			{
			  "id": "WETHUSDT",
			  "config": {
			  	"dataSource": "uniswapv2",
				"updateFrequency": "5s",
				"contractAddress": "0x0d4a11d5eeaac28ec3f61d100daf4d40471f1852",
				"httpProviderUrl": "https://ethereum-rpc.publicnode.com",
				"baseTokenIndex": 0,
				"baseTokenDecimals": 18,
				"quoteTokenIndex": 1,
				"quoteTokenDecimals": 6
			  }
			},
			{
			  "id": "PEPEWETH",
			  "config": {
			  	"dataSource": "uniswapv2",
				"updateFrequency": "5s",
				"contractAddress": "0xa43fe16908251ee70ef74718545e4fe6c5ccec9f",
				"httpProviderUrl": "https://ethereum-rpc.publicnode.com",
				"baseTokenIndex": 0,
				"baseTokenDecimals": 18,
				"quoteTokenIndex": 1,
				"quoteTokenDecimals": 18
			  }
			},
			{
			  "id": "MY_RANDOM_VALUE",
			  "config": {
			  	"dataSource": "random",
				"updateFrequency": "1s",
				"minValue": 2500,
				"maxValue": 3000
			  }
			}
		  ]
		}`

	schema, err := LoadConfigFromBytes([]byte(configStr))
	assert.NoError(t, err)

	assert.Equal(t, 3, len(schema.Sources))

	config1 := schema.Sources[0]
	assert.Equal(t, types.ValueId("WETHUSDT"), config1.Id)
	assert.NotNil(t, config1.Config)

	config2 := schema.Sources[1]
	assert.Equal(t, types.ValueId("PEPEWETH"), config2.Id)
	assert.NotNil(t, config2.Config)

	config3 := schema.Sources[2]
	assert.Equal(t, types.ValueId("MY_RANDOM_VALUE"), config3.Id)
	assert.NotNil(t, config3.Config)
}

func TestInvalidConfigs(t *testing.T) {
	// unexpected field
	configStr := `
		{
	      "extraField": "",
		  "sources": [
			{
			  "id": "MY_RANDOM_VALUE",
			  "config": {
			  	"dataSource": "random",
				"updateFrequency": "1s",
				"minValue": 2500,
				"maxValue": 3000
			  }
			}
		  ]
		}`
	_, err := LoadConfigFromBytes([]byte(configStr))
	assert.ErrorContains(t, err, "Additional property extraField is not allowed")

	// missing field
	configStr = `{}`
	_, err = LoadConfigFromBytes([]byte(configStr))
	assert.ErrorContains(t, err, "sources is required")

	// empty source list
	configStr = `{
		"sources": []
	}`
	_, err = LoadConfigFromBytes([]byte(configStr))
	assert.ErrorContains(t, err, "sources: Array must have at least 1 items")

	// incorrect type
	configStr = `
		{
		  "sources": [
			{
			  "id": 17,
			  "config": {
			  	"dataSource": "random",
				"updateFrequency": "1s",
				"minValue": 2500,
				"maxValue": 3000
			  }
			}
		  ]
		}`
	_, err = LoadConfigFromBytes([]byte(configStr))
	assert.ErrorContains(t, err, "Expected: string, given: integer")

	// invalid json
	configStr = `abcde`
	_, err = LoadConfigFromBytes([]byte(configStr))
	assert.ErrorContains(t, err, "failed to parse config JSON")

	// invalid value
	configStr = `
		{
		  "sources": [
			{
			  "id": "MY_RANDOM_VALUE",
			  "config": {
			    "dataSource": "random",
				"updateFrequency": "five_minutes",
				"minValue": 2500,
				"maxValue": 3000
			  }
			}
		  ]
		}`
	_, err = LoadConfigFromBytes([]byte(configStr))
	assert.ErrorContains(t, err, "updateFrequency: Does not match pattern")

	// unexpected field
	configStr = `
		{
		  "sources": [
			{
			  "id": "MY_RANDOM_VALUE",
			  "config": {
			  	"dataSource": "random",
				"updateFrequency": "5s",
				"minValue": 2500,
				"maxValue": 3000,
				"extraSourceConfigField": 123
			  }
			}
		  ]
		}`
	_, err = LoadConfigFromBytes([]byte(configStr))
	assert.ErrorContains(t, err, "Additional property extraSourceConfigField is not allowed")

	// invalid data source
	configStr = `
		{
		  "sources": [
			{
			  "id": "MY_RANDOM_VALUE",
			  "config": {
			  	"dataSource": "fake_data_source",
				"updateFrequency": "5s",
				"minValue": 2500,
				"maxValue": 3000
			  }
			}
		  ]
		}`
	_, err = LoadConfigFromBytes([]byte(configStr))
	assert.ErrorContains(t, err, "sources.0.config.dataSource does not match")
}
