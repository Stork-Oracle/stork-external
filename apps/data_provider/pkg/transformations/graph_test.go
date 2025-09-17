package transformations

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/Stork-Oracle/stork-external/apps/data_provider/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildTransformations(t *testing.T) {
	tests := []struct {
		name            string
		transformations []types.DataProviderTransformationConfig
		knownSources    map[types.ValueID]any
		expectedOrder   []types.ValueID
	}{
		{
			name: "simple addition",
			transformations: []types.DataProviderTransformationConfig{
				{
					ID:      "vtest1",
					Formula: "test1 + 2",
				},
			},
			knownSources: map[types.ValueID]any{
				"test1": nil,
			},
			expectedOrder: []types.ValueID{"test1", "vtest1"},
		},
		{
			name: "multiple transformations",
			transformations: []types.DataProviderTransformationConfig{
				{
					ID:      "vtest1",
					Formula: "test1 + 2",
				},
				{
					ID:      "vtest3",
					Formula: "median(vtest1, 5)",
				},
			},
			knownSources: map[types.ValueID]any{
				"test1": nil,
			},
			expectedOrder: []types.ValueID{"test1", "vtest1", "vtest3"},
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
		knownSources    map[types.ValueID]any
		expectedError   string
	}{
		{
			name: "duplicate value ids",
			transformations: []types.DataProviderTransformationConfig{
				{
					ID:      "test1",
					Formula: "test1 + 2",
				},
			},
			knownSources: map[types.ValueID]any{
				"test1": nil,
			},
			expectedError: "duplicate value id: test1",
		},
		{
			name: "unknown value ids",
			transformations: []types.DataProviderTransformationConfig{
				{
					ID:      "t1",
					Formula: "test2 + 2",
				},
			},
			knownSources: map[types.ValueID]any{
				"test1": nil,
			},
			expectedError: "no such source or transformation id: test2",
		},
		{
			name: "circular dependencies",
			transformations: []types.DataProviderTransformationConfig{
				{
					ID:      "t1",
					Formula: "t2 + 2",
				},
				{
					ID:      "t2",
					Formula: "t1 + 2",
				},
			},
			knownSources: map[types.ValueID]any{
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
			ID:      "simple_var",
			Formula: "test1 + 2",
		},
		{
			ID:      "multiple_input",
			Formula: "product(test1, test2)",
		},
		{
			ID:      "var_of_vars",
			Formula: "multiple_input + 1",
		},
	}

	knownSourceIds := map[types.ValueID]any{
		"test1": nil,
		"test2": nil,
		"test3": nil,
	}

	transformationGraph, err := BuildTransformationGraph(transformationConfigs, knownSourceIds)
	assert.NoError(t, err)

	// source update that's enough info for some but not all affected transformations
	outgoingUpdates := transformationGraph.ProcessSourceUpdates(map[types.ValueID]types.DataSourceValueUpdate{
		"test1": {
			ValueID:      "test1",
			DataSourceID: "",
			Time:         time.Now(),
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
	outgoingUpdates = transformationGraph.ProcessSourceUpdates(map[types.ValueID]types.DataSourceValueUpdate{
		"test2": {
			ValueID:      "test2",
			DataSourceID: "",
			Time:         time.Now(),
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
	outgoingUpdates = transformationGraph.ProcessSourceUpdates(map[types.ValueID]types.DataSourceValueUpdate{
		"test1": {
			ValueID:      "test1",
			DataSourceID: "",
			Time:         time.Now(),
			Value:        20.0,
		},
		"test2": {
			ValueID:      "test2",
			DataSourceID: "",
			Time:         time.Now(),
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
	outgoingUpdates = transformationGraph.ProcessSourceUpdates(map[types.ValueID]types.DataSourceValueUpdate{
		"test3": {
			ValueID:      "test3",
			DataSourceID: "",
			Time:         time.Now(),
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
			ID:      "test_transform_a_1",
			Formula: "test_source_a + 1",
		},
		{
			ID:      "test_transform_b_1",
			Formula: "test_source_b - 1",
		},
	}

	for i := 2; i < 100; i++ {
		transformationConfigs = append(transformationConfigs, types.DataProviderTransformationConfig{
			ID:      types.ValueID(fmt.Sprintf("test_transform_a_%s", strconv.Itoa(i))),
			Formula: fmt.Sprintf("test_transform_a_%s + 1", strconv.Itoa(i-1)),
		})

		transformationConfigs = append(transformationConfigs, types.DataProviderTransformationConfig{
			ID:      types.ValueID(fmt.Sprintf("test_transform_b_%s", strconv.Itoa(i))),
			Formula: fmt.Sprintf("test_transform_b_%s - 1", strconv.Itoa(i-1)),
		})
	}

	knownSourceIds := map[types.ValueID]any{
		"test_source_a": nil,
		"test_source_b": nil,
	}

	transformationGraph, err := BuildTransformationGraph(transformationConfigs, knownSourceIds)
	assert.NoError(b, err)

	now := time.Now()
	for n := 0; n < b.N; n++ {
		// update only one of the two chains
		transformationGraph.ProcessSourceUpdates(map[types.ValueID]types.DataSourceValueUpdate{
			"test_source_a": {
				ValueID:      "test_source_a",
				DataSourceID: "",
				Time:         now,
				Value:        10.0,
			},
		})
	}
}
