//go:build integration
// +build integration

package uniswapv2

import (
	"context"
	"testing"
	"time"

	"github.com/Stork-Oracle/stork-external/apps/lib/data_provider/types"
	"github.com/stretchr/testify/assert"
)

func TestUniswapDataSource_RunDataSource(t *testing.T) {
	config := types.DataProviderSourceConfig{
		Id: "WETHUSDT",
		Config: UniswapV2Config{
			DataSource:         UniswapV2DataSourceId,
			UpdateFrequency:    "5s",
			HttpProviderUrl:    "https://ethereum-rpc.publicnode.com",
			ContractAddress:    "0x0d4a11d5eeaac28ec3f61d100daf4d40471f1852",
			BaseTokenIndex:     0,
			QuoteTokenIndex:    1,
			BaseTokenDecimals:  18,
			QuoteTokenDecimals: 6,
		},
	}
	dataSource := newUniswapV2DataSource(config)
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
