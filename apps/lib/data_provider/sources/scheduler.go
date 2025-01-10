package sources

import (
	"time"

	"github.com/Stork-Oracle/stork-external/apps/lib/data_provider"
)

type Scheduler struct {
	updateFrequency time.Duration
	getUpdate       func() (data_provider.DataSourceUpdateMap, error)
}

func NewScheduler(updateFrequency time.Duration, getUpdate func() (data_provider.DataSourceUpdateMap, error)) *Scheduler {
	return &Scheduler{
		updateFrequency: updateFrequency,
		getUpdate:       getUpdate,
	}
}

func (s *Scheduler) Run(updatesCh chan data_provider.DataSourceUpdateMap) {
	s.emitUpdate(updatesCh)
	for range time.NewTicker(s.updateFrequency).C {
		s.emitUpdate(updatesCh)
	}
}

func (s *Scheduler) emitUpdate(updatesCh chan data_provider.DataSourceUpdateMap) {
	update, err := s.getUpdate()
	if err != nil {
		panic(err) // TODO: handle error
	}
	updatesCh <- update
}
