//go:build integration
// +build integration

// This file contains integration tests for this data source.

package boringvaultevm

import (
	"context"
	"testing"
	"time"

	"github.com/Stork-Oracle/stork-external/apps/data_provider/pkg/types"
	"github.com/stretchr/testify/assert"
)

// This test will hit real external data sources. It's meant to be run manually so you can manually examine the results.
func TestBoringVaultEvmDataSource_RunDataSource(t *testing.T) {
	config := types.DataProviderSourceConfig{
		Id: "MY_TEST_VALUE_ID",
		Config: BoringVaultEvmConfig{
			DataSource:      BoringVaultEvmDataSourceId,
			UpdateFrequency: "1s",
			HttpProviderUrl: "https://rpc.plumenetwork.xyz",
			ContractAddress: "0xbB2fAA1e1D6183EE3c4177476ce0d70CBd55A388",
		},
	}

	dataSource := newBoringVaultEvmDataSource(config)
	updateCh := make(chan types.DataSourceUpdateMap)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go dataSource.RunDataSource(ctx, updateCh)

	// print a few messages, fail if none come through within the timeout period
	numMessages := 10
	timeoutDuration := 10 * time.Second

	for i := 0; i < numMessages; i++ {
		select {
		case result := <-updateCh:
			t.Logf("received update: %v", result)
		case <-time.After(timeoutDuration):
			assert.Fail(t, "didn't receive update from data source in time")
		}
	}
}
