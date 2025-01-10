package data_provider

import (
	"time"
)

type (
	DataSourceId string
	ValueId      string

	DataProviderSourceConfig struct {
		Id           ValueId      `json:"id"`
		DataSourceId DataSourceId `json:"dataSource"`
		Config       any          `json:"config"`
	}

	DataProviderConfig struct {
		Sources []DataProviderSourceConfig `json:"sources,omitempty"`
	}

	DataSourceValueUpdate struct {
		ValueId   ValueId
		Timestamp time.Time
		Value     float64
	}

	DataSourceUpdateMap map[ValueId]DataSourceValueUpdate

	ValueUpdate struct {
		PublishTimestamp int64   `json:"t"`
		ValueId          ValueId `json:"a"`
		Value            string  `json:"v"`
	}

	ValueUpdateWebsocketMessage struct {
		Type string        `json:"type"`
		Data []ValueUpdate `json:"data"`
	}
)
