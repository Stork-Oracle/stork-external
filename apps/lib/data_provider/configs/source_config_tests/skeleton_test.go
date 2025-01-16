package config

import (
	"testing"

	"github.com/Stork-Oracle/stork-external/apps/lib/data_provider/configs"
	"github.com/Stork-Oracle/stork-external/apps/lib/data_provider/sources/skeleton"
	"github.com/Stork-Oracle/stork-external/apps/lib/data_provider/utils"
	"github.com/stretchr/testify/assert"
)

func TestValidSkeletonConfig(t *testing.T) {

	// TODO: set this to a valid config string using a feed from your new source
	validConfig := `
	{
	  "sources": [
		{
		  "id": "MY_VALUE",
		  "config": {
			"dataSource": "skeleton"
		  }
		}
	  ]
	}`

	config, err := configs.LoadConfigFromBytes([]byte(validConfig))
	assert.NoError(t, err)

	assert.Equal(t, 1, len(config.Sources))

	sourceConfig := config.Sources[0]

	dataSourceId, err := utils.GetDataSourceId(sourceConfig.Config)
	assert.NoError(t, err)
	assert.Equal(t, skeleton.SkeletonDataSourceId, dataSourceId)

	sourceSpecificConfig, err := skeleton.GetSourceSpecificConfig(sourceConfig)
	assert.NoError(t, err)
	assert.NotNil(t, sourceSpecificConfig)

	// TODO: write some asserts to check that the fields on sourceSpecificConfig have the values you'd expect
	t.Fatalf("implement me")
}
