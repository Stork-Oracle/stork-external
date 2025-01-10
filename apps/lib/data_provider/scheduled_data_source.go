package data_provider

import (
	"time"

	"github.com/rs/zerolog"
)

// Implement this interface to allow pulling on a schedule
type scheduledDataSourceConnector interface {
	GetUpdate() (DataSourceUpdateMap, error)
	GetUpdateFrequency() time.Duration
	GetDataSourceId() DataSourceId
}

type scheduledDataSource struct {
	connector scheduledDataSourceConnector
	logger    zerolog.Logger
}

func newScheduledDataSource(connector scheduledDataSourceConnector) *scheduledDataSource {
	return &scheduledDataSource{
		connector: connector,
		logger:    DataSourceLogger(connector.GetDataSourceId()),
	}
}

func (s *scheduledDataSource) Run(updatesCh chan DataSourceUpdateMap) {
	s.emitUpdate(updatesCh)
	for range time.NewTicker(s.connector.GetUpdateFrequency()).C {
		s.emitUpdate(updatesCh)
	}
}

func (s *scheduledDataSource) emitUpdate(updatesCh chan DataSourceUpdateMap) {
	update, err := s.connector.GetUpdate()
	if err != nil {
		s.logger.Error().Err(err).Msg("error pulling update from data source")
	} else {
		updatesCh <- update
	}
}

func (s *scheduledDataSource) GetDataSourceId() DataSourceId {
	return s.connector.GetDataSourceId()
}
