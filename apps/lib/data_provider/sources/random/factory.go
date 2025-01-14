package random

import (
	"embed"

	"github.com/Stork-Oracle/stork-external/apps/lib/data_provider/sources"
	"github.com/Stork-Oracle/stork-external/apps/lib/data_provider/types"
	"github.com/Stork-Oracle/stork-external/apps/lib/data_provider/utils"
)

var RandomDataSourceId types.DataSourceId = types.DataSourceId(utils.GetCurrentDirName())

//go:embed resources
var resourcesFS embed.FS

type randomDataSourceFactory struct{}

func (f *randomDataSourceFactory) Build(sourceConfig types.DataProviderSourceConfig) types.DataSource {
	return newRandomDataSource(sourceConfig)
}

func init() {
	sources.RegisterDataSourceFactory(RandomDataSourceId, &randomDataSourceFactory{})
}

var _ types.DataSource = (*randomDataSource)(nil)
