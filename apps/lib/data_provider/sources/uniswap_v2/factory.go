package uniswap_v2

import (
	"github.com/Stork-Oracle/stork-external/apps/lib/data_provider/sources"
	"github.com/Stork-Oracle/stork-external/apps/lib/data_provider/types"
	"github.com/Stork-Oracle/stork-external/apps/lib/data_provider/utils"
	"github.com/mitchellh/mapstructure"
)

var UniswapV2DataSourceId types.DataSourceId = types.DataSourceId(utils.GetCurrentDirName())

type uniswapV2DataSourceFactory struct{}

func (f *uniswapV2DataSourceFactory) Build(sourceConfig types.DataProviderSourceConfig) types.DataSource {
	return newUniswapV2DataSource(sourceConfig)
}

func init() {
	sources.RegisterDataSourceFactory(UniswapV2DataSourceId, &uniswapV2DataSourceFactory{})
}

func GetSourceSpecificConfig(sourceConfig types.DataProviderSourceConfig) (UniswapV2Config, error) {
	var config UniswapV2Config
	err := mapstructure.Decode(sourceConfig.Config, &config)
	return config, err
}

var _ types.DataSource = (*uniswapV2DataSource)(nil)
