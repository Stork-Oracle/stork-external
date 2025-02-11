package types

import (
	"context"
	"time"
)

type (
	DataSourceId string
	ValueId      string

	DataProviderConfig struct {
		Sources []DataProviderSourceConfig `json:"sources,omitempty"`
	}

	DataProviderSourceConfig struct {
		Id     ValueId `json:"id"`
		Config any     `json:"config"`
	}

	DataSource interface {
		RunDataSource(ctx context.Context, updatesCh chan DataSourceUpdateMap)
	}

	DataSourceFactory interface {
		Build(config DataProviderSourceConfig) DataSource
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
		Value            string  `json:"v"`
	}

	ValueUpdateWebsocketMessage struct {
		Type string        `json:"type"`
		Data []ValueUpdate `json:"data"`
	}
)
