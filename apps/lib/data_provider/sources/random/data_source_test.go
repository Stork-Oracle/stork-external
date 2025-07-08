package random

import (
	"testing"
	"time"

	"github.com/Stork-Oracle/stork-external/apps/lib/data_provider/types"
	"github.com/stretchr/testify/assert"
)

func TestRandomDataSource_getUpdate(t *testing.T) {
	minValue := 101.0
	maxValue := 105.0
	valueId := types.ValueId("TEST_RANDOM")

	config := types.DataProviderSourceConfig{
		Id: valueId,
		Config: RandomConfig{
			DataSource:      RandomDataSourceId,
			UpdateFrequency: "50ms",
			MinValue:        minValue,
			MaxValue:        maxValue,
		},
	}

	now := time.Now()
	dataSource := newRandomDataSource(config)

	// update has a valid value, valid timestamp, and expected data source id and value id
	updateMap1, err := dataSource.getUpdate()
	assert.NoError(t, err)
	update1, exists := updateMap1[valueId]
	assert.True(t, exists)
	assert.Equal(t, RandomDataSourceId, update1.DataSourceId)
	assert.Equal(t, valueId, update1.ValueId)
	time1 := update1.Time
	assert.Greater(t, time1, now)
	value1 := update1.Value
	assert.GreaterOrEqual(t, value1, minValue)
	assert.LessOrEqual(t, value1, maxValue)

	// second update has greater timestamp than first and not exactly the same value
	updateMap2, err := dataSource.getUpdate()
	assert.NoError(t, err)
	update2, exists := updateMap2[valueId]
	assert.True(t, exists)
	assert.Equal(t, RandomDataSourceId, update2.DataSourceId)
	assert.Equal(t, valueId, update2.ValueId)
	time2 := update2.Time
	assert.Greater(t, time2, time1)
	value2 := update2.Value
	assert.GreaterOrEqual(t, value2, minValue)
	assert.LessOrEqual(t, value2, maxValue)
	assert.NotEqual(t, value1, value2)
}
