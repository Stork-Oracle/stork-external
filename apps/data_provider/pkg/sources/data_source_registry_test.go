package sources

import (
	"testing"

	"github.com/Stork-Oracle/stork-external/apps/data_provider/pkg/types"
	"github.com/stretchr/testify/assert"
)

type fakeDataSourceFactory struct{}

func (*fakeDataSourceFactory) Build(config types.DataProviderSourceConfig) types.DataSource {
	return nil
}

func TestDuplicateDataSourceId(t *testing.T) {
	fakeDataSourceId := types.DataSourceID("fake_data_source_1")
	err := tryRegisterDataSourceFactory(fakeDataSourceId, nil)
	assert.NoError(t, err)
	err = tryRegisterDataSourceFactory(fakeDataSourceId, nil)
	assert.ErrorContains(t, err, "DataSourceFactory already registered for: fake_data_source_1")
}

func TestDuplicateValueId(t *testing.T) {
	fakeDataSourceId := "fake_data_source_2"

	RegisterDataSourceFactory(types.DataSourceID(fakeDataSourceId), &fakeDataSourceFactory{})
	_, _, err := BuildDataSources([]types.DataProviderSourceConfig{
		{
			ID: "fake1",
			Config: map[string]any{
				"dataSource": fakeDataSourceId,
			},
		},
		{
			ID: "fake1",
			Config: map[string]any{
				"dataSource": fakeDataSourceId,
			},
		},
	})

	assert.ErrorContains(t, err, "duplicate value id in config: fake1")
}
