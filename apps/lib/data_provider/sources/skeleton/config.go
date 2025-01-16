package skeleton

import "github.com/Stork-Oracle/stork-external/apps/lib/data_provider/types"

type SkeletonConfig struct {
	DataSource types.DataSourceId `json:"dataSource"` // required for all Data Provider Sources
	// TODO: Add any additional config parameters needed to pull a particular data feed
}
