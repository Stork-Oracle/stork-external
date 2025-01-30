package generate

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPascalToLower(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "pascal case",
			input:    "TestCase",
			expected: "testcase",
		},
		{
			name:     "one word",
			input:    "Test",
			expected: "test",
		},
		{
			name:     "multiple words",
			input:    "TestCaseExample",
			expected: "testcaseexample",
		},
		{
			name:     "with numbers",
			input:    "Test2Case",
			expected: "test2case",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := pascalToLower(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPascalToCamel(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "pascal case",
			input:    "TestCase",
			expected: "testCase",
		},
		{
			name:     "pascal case with acronym",
			input:    "TestCaseAPI",
			expected: "testCaseAPI",
		},
		{
			name:     "pascal case with start acronym",
			input:    "APITestCase",
			expected: "apiTestCase",
		},
		{
			name:     "one word",
			input:    "Test",
			expected: "test",
		},
		{
			name:     "multiple words",
			input:    "TestCaseExample",
			expected: "testCaseExample",
		},
		{
			name:     "with numbers",
			input:    "Test2Case",
			expected: "test2Case",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := pascalToCamel(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidatePascalCase(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "pascal case",
			input:    "TestCase",
			expected: true,
		},
		{
			name:     "pascal case with acronym",
			input:    "TestCaseAPI",
			expected: true,
		},
		{
			name:     "pascal case with start acronym",
			input:    "APITestCase",
			expected: true,
		},
		{
			name:     "invalid - empty",
			input:    "",
			expected: false,
		},
		{
			name:     "invalid - camel case",
			input:    "testCase",
			expected: false,
		},
		{
			name:     "invalid - spaces",
			input:    "Test Case",
			expected: false,
		},
		{
			name:     "valid - numbers",
			input:    "Test2Case",
			expected: true,
		},
		{
			name:     "invalid - starts with number",
			input:    "2TestCase",
			expected: false,
		},
		{
			name:     "invalid - contains special chars",
			input:    "Test_Case",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validatePascalCase(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
