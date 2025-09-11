package sources

import (
	"context"
	"time"

	"github.com/Stork-Oracle/stork-external/apps/data_provider/pkg/types"
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

func (s *Scheduler) RunScheduler(ctx context.Context, updatesCh chan types.DataSourceUpdateMap) {
	s.emitUpdate(updatesCh)
	ticker := time.NewTicker(s.updateFrequency)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.emitUpdate(updatesCh)
		case <-ctx.Done():
			return
		}
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
