// Code initially generated by gen.go.
// This file defines the configuration for the data source.

package boringvaultevm

import "github.com/Stork-Oracle/stork-external/apps/lib/data_provider/types"

type BoringVaultEvmConfig struct {
	DataSource      types.DataSourceId `json:"dataSource"` // required for all Data Provider Sources
	UpdateFrequency string             `json:"updateFrequency"`
	HttpProviderUrl string             `json:"httpProviderUrl"`
	ContractAddress string             `json:"contractAddress"`
}
