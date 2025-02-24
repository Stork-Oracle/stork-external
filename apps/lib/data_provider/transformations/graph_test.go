package transformations

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/Stork-Oracle/stork-external/apps/lib/data_provider/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildTransformations(t *testing.T) {
	tests := []struct {
		name            string
		transformations []types.DataProviderTransformationConfig
		knownSources    map[types.ValueId]interface{}
		expectedOrder   []types.ValueId
	}{
		{
			name: "simple addition",
			transformations: []types.DataProviderTransformationConfig{
				{
					Id:      "vtest1",
					Formula: "test1 + 2",
				},
			},
			knownSources: map[types.ValueId]interface{}{
				"test1": nil,
			},
			expectedOrder: []types.ValueId{"test1", "vtest1"},
		},
		{
			name: "multiple transformations",
			transformations: []types.DataProviderTransformationConfig{
				{
					Id:      "vtest1",
					Formula: "test1 + 2",
				},
				{
					Id:      "vtest3",
					Formula: "median(vtest1, 5)",
				},
			},
			knownSources: map[types.ValueId]interface{}{
				"test1": nil,
			},
			expectedOrder: []types.ValueId{"test1", "vtest1", "vtest3"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			transformationGraph, err := BuildTransformationGraph(test.transformations, test.knownSources)
			require.NoError(t, err)
			require.Equal(t, len(test.expectedOrder), len(transformationGraph.orderedNodes))
			for i, expected := range test.expectedOrder {
				actualValueId := transformationGraph.nodeToValueId[transformationGraph.orderedNodes[i]]
				require.Equal(t, expected, actualValueId)
			}
		})
	}
}

func TestInvalidConfigs(t *testing.T) {
	tests := []struct {
		name            string
		transformations []types.DataProviderTransformationConfig
		knownSources    map[types.ValueId]interface{}
		expectedError   string
	}{
		{
			name: "duplicate value ids",
			transformations: []types.DataProviderTransformationConfig{
				{
					Id:      "test1",
					Formula: "test1 + 2",
				},
			},
			knownSources: map[types.ValueId]interface{}{
				"test1": nil,
			},
			expectedError: "duplicate value id: test1",
		},
		{
			name: "unknown value ids",
			transformations: []types.DataProviderTransformationConfig{
				{
					Id:      "t1",
					Formula: "test2 + 2",
				},
			},
			knownSources: map[types.ValueId]interface{}{
				"test1": nil,
			},
			expectedError: "no such source or transformation id: test2",
		},
		{
			name: "circular dependencies",
			transformations: []types.DataProviderTransformationConfig{
				{
					Id:      "t1",
					Formula: "t2 + 2",
				},
				{
					Id:      "t2",
					Formula: "t1 + 2",
				},
			},
			knownSources: map[types.ValueId]interface{}{
				"test1": nil,
			},
			expectedError: "could not linearize price id graph - there may be circular dependencies",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := BuildTransformationGraph(test.transformations, test.knownSources)
			require.ErrorContains(t, err, test.expectedError)
		})
	}
}

func TestProcessSourceUpdates(t *testing.T) {
	transformationConfigs := []types.DataProviderTransformationConfig{
		{
			Id:      "simple_var",
			Formula: "test1 + 2",
		},
		{
			Id:      "multiple_input",
			Formula: "product(test1, test2)",
		},
		{
			Id:      "var_of_vars",
			Formula: "multiple_input + 1",
		},
	}

	knownSourceIds := map[types.ValueId]interface{}{
		"test1": nil,
		"test2": nil,
		"test3": nil,
	}

	transformationGraph, err := BuildTransformationGraph(transformationConfigs, knownSourceIds)
	assert.NoError(t, err)

	// source update that's enough info for some but not all affected transformations
	outgoingUpdates := transformationGraph.ProcessSourceUpdates(map[types.ValueId]types.DataSourceValueUpdate{
		"test1": {
			ValueId:      "test1",
			DataSourceId: "",
			Timestamp:    time.Now(),
			Value:        10.0,
		},
	})

	assert.Len(t, outgoingUpdates, 2)
	update, exists := outgoingUpdates["test1"]
	assert.True(t, exists)
	assert.Equal(t, 10.0, update.Value)

	update, exists = outgoingUpdates["simple_var"]
	assert.True(t, exists)
	assert.Equal(t, 12.0, update.Value)

	// source update that's enough info for all affected transformations
	outgoingUpdates = transformationGraph.ProcessSourceUpdates(map[types.ValueId]types.DataSourceValueUpdate{
		"test2": {
			ValueId:      "test2",
			DataSourceId: "",
			Timestamp:    time.Now(),
			Value:        100.0,
		},
	})

	assert.Len(t, outgoingUpdates, 3)
	update, exists = outgoingUpdates["test2"]
	assert.True(t, exists)
	assert.Equal(t, 100.0, update.Value)

	update, exists = outgoingUpdates["multiple_input"]
	assert.True(t, exists)
	assert.Equal(t, 1000.0, update.Value)

	update, exists = outgoingUpdates["var_of_vars"]
	assert.True(t, exists)
	assert.Equal(t, 1001.0, update.Value)

	// multiple source updates at once trigger a consistent batch of updates
	outgoingUpdates = transformationGraph.ProcessSourceUpdates(map[types.ValueId]types.DataSourceValueUpdate{
		"test1": {
			ValueId:      "test1",
			DataSourceId: "",
			Timestamp:    time.Now(),
			Value:        20.0,
		},
		"test2": {
			ValueId:      "test2",
			DataSourceId: "",
			Timestamp:    time.Now(),
			Value:        200.0,
		},
	})

	assert.Len(t, outgoingUpdates, 5)
	update, exists = outgoingUpdates["test1"]
	assert.True(t, exists)
	assert.Equal(t, 20.0, update.Value)

	update, exists = outgoingUpdates["test2"]
	assert.True(t, exists)
	assert.Equal(t, 200.0, update.Value)

	update, exists = outgoingUpdates["simple_var"]
	assert.True(t, exists)
	assert.Equal(t, 22.0, update.Value)

	update, exists = outgoingUpdates["multiple_input"]
	assert.True(t, exists)
	assert.Equal(t, 4000.0, update.Value)

	update, exists = outgoingUpdates["var_of_vars"]
	assert.True(t, exists)
	assert.Equal(t, 4001.0, update.Value)

	// source update with no transformations
	outgoingUpdates = transformationGraph.ProcessSourceUpdates(map[types.ValueId]types.DataSourceValueUpdate{
		"test3": {
			ValueId:      "test3",
			DataSourceId: "",
			Timestamp:    time.Now(),
			Value:        1.0,
		},
	})

	assert.Len(t, outgoingUpdates, 1)
	update, exists = outgoingUpdates["test3"]
	assert.True(t, exists)
	assert.Equal(t, 1.0, update.Value)
}

func BenchmarkTransformationGraph_ProcessSourceUpdates(b *testing.B) {
	// create two long independent dependency chains
	transformationConfigs := []types.DataProviderTransformationConfig{
		{
			Id:      "test_transform_a_1",
			Formula: "test_source_a + 1",
		},
		{
			Id:      "test_transform_b_1",
			Formula: "test_source_b - 1",
		},
	}

	for i := 2; i < 100; i++ {
		transformationConfigs = append(transformationConfigs, types.DataProviderTransformationConfig{
			Id:      types.ValueId(fmt.Sprintf("test_transform_a_%s", strconv.Itoa(i))),
			Formula: fmt.Sprintf("test_transform_a_%s + 1", strconv.Itoa(i-1)),
		})

		transformationConfigs = append(transformationConfigs, types.DataProviderTransformationConfig{
			Id:      types.ValueId(fmt.Sprintf("test_transform_b_%s", strconv.Itoa(i))),
			Formula: fmt.Sprintf("test_transform_b_%s - 1", strconv.Itoa(i-1)),
		})
	}

	knownSourceIds := map[types.ValueId]interface{}{
		"test_source_a": nil,
		"test_source_b": nil,
	}

	transformationGraph, err := BuildTransformationGraph(transformationConfigs, knownSourceIds)
	assert.NoError(b, err)

	now := time.Now()
	for n := 0; n < b.N; n++ {
		// update only one of the two chains
		transformationGraph.ProcessSourceUpdates(map[types.ValueId]types.DataSourceValueUpdate{
			"test_source_a": {
				ValueId:      "test_source_a",
				DataSourceId: "",
				Timestamp:    now,
				Value:        10.0,
			},
		})
	}
}
