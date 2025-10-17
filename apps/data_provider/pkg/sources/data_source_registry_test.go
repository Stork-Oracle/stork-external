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

func TestDuplicateDataSourceID(t *testing.T) {
	fakeDataSourceID := types.DataSourceID("fake_data_source_1")
	err := tryRegisterDataSourceFactory(fakeDataSourceID, nil)
	assert.NoError(t, err)
	err = tryRegisterDataSourceFactory(fakeDataSourceID, nil)
	assert.ErrorContains(t, err, "DataSourceFactory already registered for: fake_data_source_1")
}

func TestDuplicateValueID(t *testing.T) {
	fakeDataSourceID := "fake_data_source_2"

	RegisterDataSourceFactory(types.DataSourceID(fakeDataSourceID), &fakeDataSourceFactory{})
	_, _, err := BuildDataSources([]types.DataProviderSourceConfig{
		{
			ID: "fake1",
			Config: map[string]any{
				"dataSource": fakeDataSourceID,
			},
		},
		{
			ID: "fake1",
			Config: map[string]any{
				"dataSource": fakeDataSourceID,
			},
		},
	})

	assert.ErrorContains(t, err, "duplicate value id in config: fake1")
}
