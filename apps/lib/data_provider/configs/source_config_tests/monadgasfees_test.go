// Code initially generated by gen.go.
// This file tests correctly loading and parsing the new source config.
package config

import (
	"testing"

	"github.com/Stork-Oracle/stork-external/apps/lib/data_provider/configs"
	"github.com/Stork-Oracle/stork-external/apps/lib/data_provider/sources/monadgasfees"
	"github.com/Stork-Oracle/stork-external/apps/lib/data_provider/utils"
	"github.com/stretchr/testify/assert"
)

func TestValidMonadGasFeesConfig(t *testing.T) {

	// TODO: set this to a valid config string using a feed from your new source
	validConfig := `
	{
	  "sources": [
		{
		  "id": "MY_VALUE",
		  "config": {
			"dataSource": "monadgasfees"
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
	assert.Equal(t, monadgasfees.MonadGasFeesDataSourceId, dataSourceId)

	sourceSpecificConfig, err := monadgasfees.GetSourceSpecificConfig(sourceConfig)
	assert.NoError(t, err)
	assert.NotNil(t, sourceSpecificConfig)

	// TODO: write some asserts to check that the fields on sourceSpecificConfig have the values you'd expect
	t.Fatalf("implement me")
}
