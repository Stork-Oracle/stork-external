package config

import (
	"testing"

	"github.com/Stork-Oracle/stork-external/apps/data_provider/pkg/configs"
	"github.com/Stork-Oracle/stork-external/apps/data_provider/pkg/sources/random"
	"github.com/Stork-Oracle/stork-external/apps/data_provider/pkg/types"
	"github.com/Stork-Oracle/stork-external/apps/data_provider/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestValidRandomConfig(t *testing.T) {
	validRandomConfig := `
	{
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
	}
`
	config, err := configs.LoadConfigFromBytes([]byte(validRandomConfig))
	assert.NoError(t, err)

	assert.Equal(t, 1, len(config.Sources))

	sourceConfig := config.Sources[0]
	assert.Equal(t, types.ValueID("MY_RANDOM_VALUE"), sourceConfig.ID)

	dataSourceID, err := utils.GetDataSourceID(sourceConfig.Config)
	assert.NoError(t, err)
	assert.Equal(t, random.RandomDataSourceID, dataSourceID)

	randomConfig, err := random.GetSourceSpecificConfig(sourceConfig)
	assert.NoError(t, err)

	assert.Equal(t, types.DataSourceID("random"), randomConfig.DataSource)
	assert.Equal(t, "1s", randomConfig.UpdateFrequency)
	assert.Equal(t, 2500.0, randomConfig.MinValue)
	assert.Equal(t, 3000.0, randomConfig.MaxValue)
}
