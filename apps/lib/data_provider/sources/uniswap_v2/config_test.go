package uniswap_v2

import (
	"encoding/json"
	"testing"

	"github.com/Stork-Oracle/stork-external/apps/lib/data_provider/utils"
	"github.com/stretchr/testify/assert"
)

func TestLoadConfigSchema(t *testing.T) {
	_, err := utils.LoadSchema(resourcesFS)
	assert.NoError(t, err)
}

func TestValidConfig(t *testing.T) {
	schema, err := utils.LoadSchema(resourcesFS)
	assert.NoError(t, err)

	configStr := `
	{
		"updateFrequency": "5s",
		"contractAddress": "0x0d4a11d5eeaac28ec3f61d100daf4d40471f1852",
		"httpProviderUrl": "https://eth-mainnet.g.alchemy.com/v2/",
		"providerApiKeyEnvVar": "ALCHEMY_API_KEY",
		"baseTokenIndex": 0,
		"baseTokenDecimals": 18,
		"quoteTokenIndex": 1,
		"quoteTokenDecimals": 6
  	}
	`
	err = utils.ValidateConfig([]byte(configStr), schema)
	assert.NoError(t, err)

	// missing some optional fields
	configStr = `
	{
		"updateFrequency": "5s",
		"contractAddress": "0x0d4a11d5eeaac28ec3f61d100daf4d40471f1852",
		"httpProviderUrl": "https://eth-mainnet.g.alchemy.com/v2/",
		"baseTokenIndex": 0,
		"quoteTokenIndex": 1
  	}
	`
	err = utils.ValidateConfig([]byte(configStr), schema)
	assert.NoError(t, err)
}

func TestMissingKey(t *testing.T) {
	schema, err := utils.LoadSchema(resourcesFS)
	assert.NoError(t, err)

	configStr := `
	{
		"updateFrequency": "5s",
		"contractAddress": "0x0d4a11d5eeaac28ec3f61d100daf4d40471f1852",
		"httpProviderUrl": "https://eth-mainnet.g.alchemy.com/v2/",
		"providerApiKeyEnvVar": "ALCHEMY_API_KEY",
		"baseTokenDecimals": 18,
		"quoteTokenIndex": 1,
		"quoteTokenDecimals": 6
  	}
	`

	err = utils.ValidateConfig([]byte(configStr), schema)
	assert.ErrorContains(t, err, "baseTokenIndex is required")
}

func TestExtraKey(t *testing.T) {
	schema, err := utils.LoadSchema(resourcesFS)
	assert.NoError(t, err)

	configStr := `
	{
		"updateFrequency": "5s",
		"contractAddress": "0x0d4a11d5eeaac28ec3f61d100daf4d40471f1852",
		"httpProviderUrl": "https://eth-mainnet.g.alchemy.com/v2/",
		"providerApiKeyEnvVar": "ALCHEMY_API_KEY",
		"baseTokenIndex": 0,
		"baseTokenDecimals": 18,
		"quoteTokenIndex": 1,
		"quoteTokenDecimals": 6,
		"extraField": 123
  	}
	`

	err = utils.ValidateConfig([]byte(configStr), schema)
	assert.ErrorContains(t, err, "Additional property extraField is not allowed")
}

func TestBadFrequency(t *testing.T) {
	schema, err := utils.LoadSchema(resourcesFS)
	assert.NoError(t, err)

	configStr := `
	{
		"updateFrequency": "5",
		"contractAddress": "0x0d4a11d5eeaac28ec3f61d100daf4d40471f1852",
		"httpProviderUrl": "https://eth-mainnet.g.alchemy.com/v2/",
		"providerApiKeyEnvVar": "ALCHEMY_API_KEY",
		"baseTokenIndex": 0,
		"baseTokenDecimals": 18,
		"quoteTokenIndex": 1,
		"quoteTokenDecimals": 6
  	}
	`

	err = utils.ValidateConfig([]byte(configStr), schema)
	assert.ErrorContains(t, err, "updateFrequency: Does not match pattern")
}

func TestBadContractAddress(t *testing.T) {
	schema, err := utils.LoadSchema(resourcesFS)
	assert.NoError(t, err)

	configStr := `
	{
		"updateFrequency": "5s",
		"contractAddress": "0d4a11d5eeaac28ec3f61d100daf4d40471f1852",
		"httpProviderUrl": "https://eth-mainnet.g.alchemy.com/v2/",
		"providerApiKeyEnvVar": "ALCHEMY_API_KEY",
		"baseTokenIndex": 0,
		"baseTokenDecimals": 18,
		"quoteTokenIndex": 1,
		"quoteTokenDecimals": 6
  	}
	`

	err = utils.ValidateConfig([]byte(configStr), schema)
	assert.ErrorContains(t, err, "contractAddress: Does not match pattern")
}

func TestLoadConfig(t *testing.T) {
	configStr := `
	{
		"updateFrequency": "5s",
		"contractAddress": "0x0d4a11d5eeaac28ec3f61d100daf4d40471f1852",
		"httpProviderUrl": "https://eth-mainnet.g.alchemy.com/v2/",
		"providerApiKeyEnvVar": "ALCHEMY_API_KEY",
		"baseTokenIndex": 0,
		"baseTokenDecimals": 18,
		"quoteTokenIndex": 1,
		"quoteTokenDecimals": 6
  	}
	`
	var config uniswapV2Config
	err := json.Unmarshal([]byte(configStr), &config)
	assert.NoError(t, err)

	assert.Equal(t, "5s", config.UpdateFrequency)
	assert.Equal(t, "0x0d4a11d5eeaac28ec3f61d100daf4d40471f1852", config.ContractAddress)
	assert.Equal(t, "https://eth-mainnet.g.alchemy.com/v2/", config.HttpProviderUrl)
	assert.Equal(t, "ALCHEMY_API_KEY", config.ProviderApiKeyEnvVar)
	assert.Equal(t, int8(0), config.BaseTokenIndex)
	assert.Equal(t, int8(18), config.BaseTokenDecimals)
	assert.Equal(t, int8(1), config.QuoteTokenIndex)
	assert.Equal(t, int8(6), config.QuoteTokenDecimals)
}
