package sources

import (
	"time"

	"github.com/Stork-Oracle/stork-external/apps/lib/data_provider/types"
	"github.com/rs/zerolog"
)

type Scheduler struct {
	updateFrequency time.Duration
	getUpdate       func() (types.DataSourceUpdateMap, error)
	handleErr       func(error)
}

func NewScheduler(
	updateFrequency time.Duration,
	getUpdate func() (types.DataSourceUpdateMap, error),
	handleErr func(error),
) *Scheduler {
	return &Scheduler{
		updateFrequency: updateFrequency,
		getUpdate:       getUpdate,
		handleErr:       handleErr,
	}
}

func (s *Scheduler) RunScheduler(updatesCh chan types.DataSourceUpdateMap) {
	s.emitUpdate(updatesCh)
	for range time.NewTicker(s.updateFrequency).C {
		s.emitUpdate(updatesCh)
	}
}

func (s *Scheduler) emitUpdate(updatesCh chan types.DataSourceUpdateMap) {
	update, err := s.getUpdate()
	if err != nil {
		s.handleErr(err)
	}
	updatesCh <- update
}

func GetErrorLogHandler(logger zerolog.Logger, level zerolog.Level) func(error) {
	return func(err error) {
		logger.WithLevel(level).Err(err).Msg("Failed to get scheduled update")
	}
}
