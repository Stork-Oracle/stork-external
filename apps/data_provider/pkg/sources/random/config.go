package random

import "github.com/Stork-Oracle/stork-external/apps/data_provider/pkg/types"

type RandomConfig struct {
	DataSource      types.DataSourceID `json:"dataSource"`
	UpdateFrequency string             `json:"updateFrequency"`
	MinValue        float64            `json:"minValue"`
	MaxValue        float64            `json:"maxValue"`
}
