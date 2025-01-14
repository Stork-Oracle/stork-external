package config

import (
	"testing"

	"github.com/Stork-Oracle/stork-external/apps/lib/data_provider/configs"
	"github.com/Stork-Oracle/stork-external/apps/lib/data_provider/sources/random"
	"github.com/Stork-Oracle/stork-external/apps/lib/data_provider/types"
	"github.com/stretchr/testify/assert"
)

func TestValidRandomConfig(t *testing.T) {
	configStr := `
		{
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

	config, err := configs.LoadConfigFromBytes([]byte(configStr))
	assert.NoError(t, err)

	assert.Equal(t, 1, len(config.Sources))

	sourceConfig := config.Sources[0]
	assert.Equal(t, types.ValueId("MY_RANDOM_VALUE"), sourceConfig.Id)
	assert.Equal(t, types.DataSourceId("random"), sourceConfig.DataSourceId)

	uniswapConfig, err := random.GetSourceSpecificConfig(sourceConfig)
	assert.NoError(t, err)

	assert.Equal(t, "1s", uniswapConfig.UpdateFrequency)
	assert.Equal(t, 2500.0, uniswapConfig.MinValue)
	assert.Equal(t, 3000.0, uniswapConfig.MaxValue)
}
