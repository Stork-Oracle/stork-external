package data_provider

import "time"

type (
	DataSourceId string
	ValueId      string

	DataProviderSourceConfig struct {
		Id           ValueId      `json:"id"`
		DataSourceId DataSourceId `json:"dataSource"`
		Config       any          `json:"config"`
	}

	DataProviderConfig struct {
		WsUrl   string                     `json:"wsUrl,omitempty"`
		Verbose bool                       `json:"verbose,omitempty"`
		Sources []DataProviderSourceConfig `json:"sources,omitempty"`
	}

	DataSourceValueUpdate struct {
		ValueId      ValueId
		DataSourceId DataSourceId
		Timestamp    time.Time
		Value        float64
	}

	DataSourceUpdateMap map[ValueId]DataSourceValueUpdate

	ValueUpdate struct {
		PublishTimestamp int64   `json:"t"`
		ValueId          ValueId `json:"a"`
		Value            float64 `json:"p"`
	}
)
