package random

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
		"updateFrequency": "1s",
		"minValue": 2500,
		"maxValue": 3000
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
		"updateFrequency": "1s",
		"maxValue": 3000
	}
	`

	err = utils.ValidateConfig([]byte(configStr), schema)
	assert.ErrorContains(t, err, "minValue is required")
}

func TestExtraKey(t *testing.T) {
	schema, err := utils.LoadSchema(resourcesFS)
	assert.NoError(t, err)

	configStr := `
	{
		"updateFrequency": "1s",
		"minValue": 2500,
		"maxValue": 3000,
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
		"updateFrequency": "10",
		"minValue": 2500,
		"maxValue": 3000
	}
	`

	err = utils.ValidateConfig([]byte(configStr), schema)
	assert.ErrorContains(t, err, "updateFrequency: Does not match pattern")
}

func TestLoadConfig(t *testing.T) {
	configStr := `
	{
		"updateFrequency": "1s",
		"minValue": 2500,
		"maxValue": 3000
	}
	`
	var config randomConfig
	err := json.Unmarshal([]byte(configStr), &config)
	assert.NoError(t, err)

	assert.Equal(t, "1s", config.UpdateFrequency)
	assert.Equal(t, 2500.0, config.MinValue)
	assert.Equal(t, 3000.0, config.MaxValue)
}
