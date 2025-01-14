package random

import "github.com/Stork-Oracle/stork-external/apps/lib/data_provider/types"

type RandomConfig struct {
	DataSource      types.DataSourceId `json:"dataSource"`
	UpdateFrequency string             `json:"updateFrequency"`
	MinValue        float64            `json:"minValue"`
	MaxValue        float64            `json:"maxValue"`
}
