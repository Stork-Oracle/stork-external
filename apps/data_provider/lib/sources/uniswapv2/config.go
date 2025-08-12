package uniswapv2

import "github.com/Stork-Oracle/stork-external/apps/data_provider/lib/types"

type UniswapV2Config struct {
	DataSource         types.DataSourceId `json:"dataSource"`
	UpdateFrequency    string             `json:"updateFrequency"`
	HttpProviderUrl    string             `json:"httpProviderUrl"`
	ContractAddress    string             `json:"contractAddress"`
	BaseTokenIndex     int8               `json:"baseTokenIndex"`
	QuoteTokenIndex    int8               `json:"quoteTokenIndex"`
	BaseTokenDecimals  int8               `json:"baseTokenDecimals,omitempty"`
	QuoteTokenDecimals int8               `json:"quoteTokenDecimals,omitempty"`
}
