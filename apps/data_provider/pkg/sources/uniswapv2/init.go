package uniswapv2

import (
	"github.com/Stork-Oracle/stork-external/apps/data_provider/pkg/sources"
	"github.com/Stork-Oracle/stork-external/apps/data_provider/pkg/types"
	"github.com/Stork-Oracle/stork-external/apps/data_provider/pkg/utils"
	"github.com/mitchellh/mapstructure"
)

var UniswapV2DataSourceID types.DataSourceID = types.DataSourceID(utils.GetCurrentDirName())

type uniswapV2DataSourceFactory struct{}

func (f *uniswapV2DataSourceFactory) Build(sourceConfig types.DataProviderSourceConfig) types.DataSource {
	return newUniswapV2DataSource(sourceConfig)
}

func init() {
	sources.RegisterDataSourceFactory(UniswapV2DataSourceID, &uniswapV2DataSourceFactory{})
}

var (
	_ types.DataSource        = (*uniswapV2DataSource)(nil)
	_ types.DataSourceFactory = (*uniswapV2DataSourceFactory)(nil)
)

func GetSourceSpecificConfig(sourceConfig types.DataProviderSourceConfig) (UniswapV2Config, error) {
	var config UniswapV2Config
	err := mapstructure.Decode(sourceConfig.Config, &config)
	return config, err
}
