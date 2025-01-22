package config

import (
	"testing"

	"github.com/Stork-Oracle/stork-external/apps/lib/data_provider/configs"
	"github.com/Stork-Oracle/stork-external/apps/lib/data_provider/sources/uniswapv2"
	"github.com/Stork-Oracle/stork-external/apps/lib/data_provider/types"
	"github.com/Stork-Oracle/stork-external/apps/lib/data_provider/utils"
	"github.com/stretchr/testify/assert"
)

func TestValidUniswapV2Config(t *testing.T) {
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
			}
		  ]
		}`

	config, err := configs.LoadConfigFromBytes([]byte(configStr))
	assert.NoError(t, err)

	assert.Equal(t, 1, len(config.Sources))

	sourceConfig := config.Sources[0]
	assert.Equal(t, types.ValueId("WETHUSDT"), sourceConfig.Id)

	dataSourceId, err := utils.GetDataSourceId(sourceConfig.Config)
	assert.NoError(t, err)
	assert.Equal(t, uniswapv2.UniswapV2DataSourceId, dataSourceId)

	uniswapConfig, err := uniswapv2.GetSourceSpecificConfig(sourceConfig)
	assert.NoError(t, err)

	assert.Equal(t, types.DataSourceId("uniswapv2"), uniswapConfig.DataSource)
	assert.Equal(t, "5s", uniswapConfig.UpdateFrequency)
	assert.Equal(t, "0x0d4a11d5eeaac28ec3f61d100daf4d40471f1852", uniswapConfig.ContractAddress)
	assert.Equal(t, "https://ethereum-rpc.publicnode.com", uniswapConfig.HttpProviderUrl)
	assert.Equal(t, int8(0), uniswapConfig.BaseTokenIndex)
	assert.Equal(t, int8(18), uniswapConfig.BaseTokenDecimals)
	assert.Equal(t, int8(1), uniswapConfig.QuoteTokenIndex)
	assert.Equal(t, int8(6), uniswapConfig.QuoteTokenDecimals)
}
