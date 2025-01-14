package uniswap_v2

import (
	"embed"

	"github.com/Stork-Oracle/stork-external/apps/lib/data_provider/sources"
	"github.com/Stork-Oracle/stork-external/apps/lib/data_provider/types"
	"github.com/Stork-Oracle/stork-external/apps/lib/data_provider/utils"
)

var UniswapV2DataSourceId types.DataSourceId = types.DataSourceId(utils.GetCurrentDirName())

//go:embed resources
var resourcesFS embed.FS

type uniswapV2DataSourceFactory struct{}

func (f *uniswapV2DataSourceFactory) Build(sourceConfig types.DataProviderSourceConfig) types.DataSource {
	return newUniswapV2DataSource(sourceConfig)
}

func init() {
	sources.RegisterDataSourceFactory(UniswapV2DataSourceId, &uniswapV2DataSourceFactory{})
}

var _ types.DataSource = (*uniswapV2DataSource)(nil)
