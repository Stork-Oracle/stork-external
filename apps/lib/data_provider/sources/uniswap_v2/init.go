package uniswap_v2

import (
	"embed"

	"github.com/Stork-Oracle/stork-external/apps/lib/data_provider/sources"
	"github.com/Stork-Oracle/stork-external/apps/lib/data_provider/types"
	"github.com/Stork-Oracle/stork-external/apps/lib/data_provider/utils"
	"github.com/xeipuuv/gojsonschema"
)

const UniswapV2DataSourceId types.DataSourceId = "UNISWAP_V2"

//go:embed resources
var resourcesFS embed.FS

type uniswapV2DataSourceFactory struct{}

func (f *uniswapV2DataSourceFactory) Build(sourceConfig types.DataProviderSourceConfig) types.DataSource {
	return newUniswapV2DataSource(sourceConfig)
}

func (f *uniswapV2DataSourceFactory) GetSchema() (*gojsonschema.Schema, error) {
	return utils.LoadSchema("resources/config_schema.json", resourcesFS)
}

func init() {
	sources.RegisterDataSourceFactory(UniswapV2DataSourceId, &uniswapV2DataSourceFactory{})
}

var _ types.DataSource = (*uniswapV2DataSource)(nil)
