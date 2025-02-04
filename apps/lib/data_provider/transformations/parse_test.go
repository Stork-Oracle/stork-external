package transformations

import (
	"testing"

	"github.com/Stork-Oracle/stork-external/apps/lib/data_provider/types"
	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name      string
		variables map[string]types.DataSourceValueUpdate
		formula   string
		expected  float64
	}{
		{
			name:     "simple addition",
			formula:  "1 + 2",
			expected: 3.0,
		},
		{
			name:     "simple multiplication",
			formula:  "2 * 3",
			expected: 6.0,
		},
		{
			name:     "complex expression",
			formula:  "1 + 2 * 3",
			expected: 7.0,
		},
		{
			name:     "multiple expressions",
			formula:  "1 + 2 * 3 + 4",
			expected: 11.0,
		},
		{
			name:     "order of operations",
			formula:  "((1 + 3) * 3) / 4",
			expected: 3.0,
		},
		{
			name:     "median",
			formula:  "median(1, 2, 3)",
			expected: 2.0,
		},
		{
			name:     "mean",
			formula:  "mean(1, 2, 3)",
			expected: 2.0,
		},
		{
			name:     "sum",
			formula:  "sum(1, 2, 3)",
			expected: 6.0,
		},
		{
			name:     "product",
			formula:  "product(1, 2, 3)",
			expected: 6.0,
		},
		{
			name:     "nested expression",
			formula:  "median(1, 2, 3) * mean(4, 5, 6)",
			expected: 10.0,
		},
		{
			name:    "nested expression with variables",
			formula: "median(x, y, z) * mean(4, 5, 6)",
			variables: map[string]types.DataSourceValueUpdate{
				"x": {
					Value: 1,
				},
				"y": {
					Value: 2,
				},
				"z": {
					Value: 3,
				},
			},
			expected: 10.0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expr, err := parse(test.formula)
			require.NoError(t, err)
			out := expr.Eval(test.variables)
			require.Equal(t, test.expected, out)
		})
	}
}

func TestGetDependencies(t *testing.T) {
	tests := []struct {
		name     string
		formula  string
		expected []string
	}{
		{
			name:     "simple addition",
			formula:  "1 + 2",
			expected: []string{},
		},
		{
			name:     "simple addition with variables",
			formula:  "x + y",
			expected: []string{"x", "y"},
		},
		{
			name:     "nested expression",
			formula:  "median(x, y, z) * mean(4, 5, 6)",
			expected: []string{"x", "y", "z"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expr, err := parse(test.formula)
			require.NoError(t, err)
			require.Equal(t, test.expected, expr.getDependencies())
		})
	}
}

func TestBuildTransformations(t *testing.T) {
	tests := []struct {
		name            string
		transformations []types.DataProviderTransformationConfig
		knownSources    map[types.ValueId]interface{}
		formula         string
		expectedSize    int
		expectedIndices []struct {
			Index int
			Id    types.ValueId
		}
	}{
		{
			name: "simple addition",
			transformations: []types.DataProviderTransformationConfig{
				{
					Id:      "vtest1",
					Formula: "s.test1 + 2",
				},
			},
			knownSources: map[types.ValueId]interface{}{
				"test1": nil,
			},
			expectedSize: 1,
			expectedIndices: []struct {
				Index int
				Id    types.ValueId
			}{
				{Index: 0, Id: "vtest1"},
			},
		},
		{
			name: "multiple transformations",
			transformations: []types.DataProviderTransformationConfig{
				{
					Id:      "vtest1",
					Formula: "s.test1 + 2",
				},
				{
					Id:      "vtest2",
					Formula: "s.test2 + 2",
				},
				{
					Id:      "vtest3",
					Formula: "median(t.vtest1, t.vtest2)",
				},
			},
			knownSources: map[types.ValueId]interface{}{
				"test1": nil,
				"test2": nil,
			},
			expectedSize: 3,
			expectedIndices: []struct {
				Index int
				Id    types.ValueId
			}{
				{Index: 2, Id: "vtest3"},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			transformations, err := BuildTransformations(test.transformations, test.knownSources)
			require.NoError(t, err)
			require.Equal(t, test.expectedSize, len(transformations))
			for _, expected := range test.expectedIndices {
				require.Equal(t, expected.Id, transformations[expected.Index].Id)
			}
		})
	}
	// 	transformations, err := BuildTransformations([]types.DataProviderTransformationConfig{
	// 	{
	// 		Id:      "vtest1",
	// 		Formula: "s.test1 + 2",
	// 	},
	// 	{
	// 		Id:      "vtest2",
	// 		Formula: "s.test2 + 2",
	// 	},
	// 	{
	// 		Id:      "vtest3",
	// 		Formula: "median(t.vtest1, t.vtest2)",
	// 	},
	// }, map[types.ValueId]interface{}{
	// 	"test1": nil,
	// 	"test2": nil,
	// })
	// require.NoError(t, err)
	// require.Equal(t, 3, len(transformations))
}
