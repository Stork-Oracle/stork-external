package types

import (
	"context"
	"time"
)

type (
	DataSourceID string
	ValueID      string

	DataProviderConfig struct {
		Sources         []DataProviderSourceConfig         `json:"sources,omitempty"`
		Transformations []DataProviderTransformationConfig `json:"transformations,omitempty"`
	}

	DataProviderSourceConfig struct {
		ID     ValueID `json:"id"`
		Config any     `json:"config"`
	}

	DataProviderTransformationConfig struct {
		ID      ValueID `json:"id"`
		Formula string  `json:"formula"`
	}

	DataSource interface {
		RunDataSource(ctx context.Context, updatesCh chan DataSourceUpdateMap)
	}

	DataSourceFactory interface {
		Build(config DataProviderSourceConfig) DataSource
	}

	DataSourceValueUpdate struct {
		ValueID      ValueID
		DataSourceID DataSourceID
		Time         time.Time
		Value        float64
	}

	DataSourceUpdateMap map[ValueID]DataSourceValueUpdate

	ValueUpdate struct {
		PublishTimestampNano int64   `json:"t"`
		ValueID              ValueID `json:"a"`
		Value                string  `json:"v"`
	}

	ValueUpdateWebsocketMessage struct {
		Type string        `json:"type"`
		Data []ValueUpdate `json:"data"`
	}
)
