package types

import (
	"context"
	"time"
)

type (
	DataSourceId string
	ValueId      string

	DataProviderConfig struct {
		Sources         []DataProviderSourceConfig         `json:"sources,omitempty"`
		Transformations []DataProviderTransformationConfig `json:"transformations,omitempty"`
	}

	DataProviderSourceConfig struct {
		Id     ValueId `json:"id"`
		Config any     `json:"config"`
	}

	DataProviderTransformationConfig struct {
		Id      ValueId `json:"id"`
		Formula string  `json:"formula"`
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
		Time         time.Time
		Value        float64
	}

	DataSourceUpdateMap map[ValueId]DataSourceValueUpdate

	ValueUpdate struct {
		PublishTimestampNano int64   `json:"t"`
		ValueId              ValueId `json:"a"`
		Value                string  `json:"v"`
	}

	ValueUpdateWebsocketMessage struct {
		Type string        `json:"type"`
		Data []ValueUpdate `json:"data"`
	}
)
