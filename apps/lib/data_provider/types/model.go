package types

import (
	"time"

	"github.com/xeipuuv/gojsonschema"
)

type (
	DataSourceId string
	ValueId      string

	DataProviderConfig struct {
		Sources []DataProviderSourceConfig `json:"sources,omitempty"`
	}

	DataProviderSourceConfig struct {
		Id           ValueId      `json:"id"`
		DataSourceId DataSourceId `json:"dataSource"`
		Config       any          `json:"config"`
	}

	DataSource interface {
		RunDataSource(updatesCh chan DataSourceUpdateMap)
	}

	DataSourceFactory interface {
		Build(config DataProviderSourceConfig) DataSource
		GetSchema() (*gojsonschema.Schema, error)
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
