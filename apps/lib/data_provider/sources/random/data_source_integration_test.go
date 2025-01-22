//go:build integration
// +build integration

package random

import (
	"testing"
	"time"

	"github.com/Stork-Oracle/stork-external/apps/lib/data_provider/types"
	"github.com/stretchr/testify/assert"
)

func TestRandomDataSource_RunDataSource(t *testing.T) {
	config := types.DataProviderSourceConfig{
		Id: "TEST_RANDOM",
		Config: RandomConfig{
			DataSource:      RandomDataSourceId,
			UpdateFrequency: "50ms",
			MinValue:        101.0,
			MaxValue:        105.0,
		},
	}
	dataSource := newRandomDataSource(config)
	updateCh := make(chan types.DataSourceUpdateMap)
	go dataSource.RunDataSource(updateCh)

	// print a few messages, fail if none come through within the timeout period
	numMessages := 10
	timeoutDuration := 100 * time.Millisecond
	for i := 0; i < numMessages; i++ {
		select {
		case result := <-updateCh:
			t.Logf("received update: %v", result)
		case <-time.After(timeoutDuration):
			assert.Fail(t, "didn't receive update from data source in time")
		}
	}
}
