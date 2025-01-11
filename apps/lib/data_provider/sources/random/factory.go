package random

import (
	"embed"

	"github.com/Stork-Oracle/stork-external/apps/lib/data_provider/sources"
	"github.com/Stork-Oracle/stork-external/apps/lib/data_provider/types"
	"github.com/Stork-Oracle/stork-external/apps/lib/data_provider/utils"
	"github.com/xeipuuv/gojsonschema"
)

const RandomDataSourceId types.DataSourceId = "RANDOM_NUMBER"

//go:embed resources
var resourcesFS embed.FS

type randomDataSourceFactory struct{}

func (f *randomDataSourceFactory) Build(sourceConfig types.DataProviderSourceConfig) types.DataSource {
	return newRandomDataSource(sourceConfig)
}

func (f *randomDataSourceFactory) GetSchema() (*gojsonschema.Schema, error) {
	return utils.LoadSchema("resources/config_schema.json", resourcesFS)
}

func init() {
	sources.RegisterDataSourceFactory(RandomDataSourceId, &randomDataSourceFactory{})
}

var _ types.DataSource = (*randomDataSource)(nil)
