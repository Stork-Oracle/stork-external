package config

import (
	"testing"

	"github.com/Stork-Oracle/stork-external/apps/lib/data_provider/configs"
	"github.com/Stork-Oracle/stork-external/apps/lib/data_provider/sources/uniswap_v2"
	"github.com/Stork-Oracle/stork-external/apps/lib/data_provider/types"
	"github.com/stretchr/testify/assert"
)

func TestValidUniswapV2Config(t *testing.T) {
	configStr := `
		{
		  "sources": [
			{
			  "id": "WETHUSDT",
			  "dataSource": "uniswap_v2",
			  "configs": {
				"updateFrequency": "5s",
				"contractAddress": "0x0d4a11d5eeaac28ec3f61d100daf4d40471f1852",
				"httpProviderUrl": "https://eth-mainnet.g.alchemy.com/v2/",
				"providerApiKeyEnvVar": "ALCHEMY_API_KEY",
				"baseTokenIndex": 1,
				"baseTokenDecimals": 18,
				"quoteTokenIndex": 2,
				"quoteTokenDecimals": 6
			  }
			}
		  ]
		}`

	config, err := configs.LoadConfigFromBytes([]byte(configStr))
	assert.NoError(t, err)

	assert.Equal(t, 1, len(config.Sources))

	sourceConfig := config.Sources[0]
	assert.Equal(t, types.ValueId("WETHUSDT"), sourceConfig.Id)
	assert.Equal(t, types.DataSourceId("uniswap_v2"), sourceConfig.DataSourceId)

	uniswapConfig, err := uniswap_v2.GetSourceSpecificConfig(sourceConfig)
	assert.NoError(t, err)

	assert.Equal(t, "5s", uniswapConfig.UpdateFrequency)
	assert.Equal(t, "0x0d4a11d5eeaac28ec3f61d100daf4d40471f1852", uniswapConfig.ContractAddress)
	assert.Equal(t, "https://eth-mainnet.g.alchemy.com/v2/", uniswapConfig.HttpProviderUrl)
	assert.Equal(t, "ALCHEMY_API_KEY", uniswapConfig.ProviderApiKeyEnvVar)
	assert.Equal(t, int8(1), uniswapConfig.BaseTokenIndex)
	assert.Equal(t, int8(18), uniswapConfig.BaseTokenDecimals)
	assert.Equal(t, int8(2), uniswapConfig.QuoteTokenIndex)
	assert.Equal(t, int8(6), uniswapConfig.QuoteTokenDecimals)

}
