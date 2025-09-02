package pusher

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPluralize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    int
		expected string
	}{
		{"singular", 1, ""},
		{"zero", 0, "s"},
		{"plural", 2, "s"},
		{"many", 100, "s"},
		{"negative", -1, "s"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := Pluralize(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestHexStringToBytes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		input     string
		expected  []byte
		wantError bool
	}{
		{"valid hex with prefix", "0x1234", []byte{0x12, 0x34}, false},
		{"valid hex without prefix", "1234", []byte{0x12, 0x34}, false},
		{"empty string", "", []byte{}, false},
		{"empty with prefix", "0x", []byte{}, false},
		{"single byte", "0xFF", []byte{0xFF}, false},
		{"invalid hex", "0xZZ", nil, true},
		{"odd length", "0x123", nil, true},
		{"long valid hex", "0x123456789ABCDEF0", []byte{0x12, 0x34, 0x56, 0x78, 0x9A, 0xBC, 0xDE, 0xF0}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := hexStringToBytes(tt.input)

			if tt.wantError {
				assert.Error(t, err)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestHexStringToFixedBytes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		input     string
		length    int
		expected  []byte
		wantError bool
	}{
		{"exact length", "0x1234", 2, []byte{0x12, 0x34}, false},
		{"shorter input padded", "0x12", 4, []byte{0x00, 0x00, 0x00, 0x12}, false},
		{"empty input", "0x", 4, []byte{0x00, 0x00, 0x00, 0x00}, false},
		{"input too long", "0x123456", 2, nil, true},
		{"invalid hex", "0xZZ", 2, nil, true},
		{"zero length", "", 0, []byte{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := hexStringToFixedBytes(tt.input, tt.length)

			if tt.wantError {
				assert.Error(t, err)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestHexStringToByte20(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		input     string
		expected  [20]byte
		wantError bool
	}{
		{
			name:  "valid 20-byte hex",
			input: "0x1234567890123456789012345678901234567890",
			expected: [20]byte{
				0x12, 0x34, 0x56, 0x78, 0x90, 0x12, 0x34, 0x56, 0x78, 0x90,
				0x12, 0x34, 0x56, 0x78, 0x90, 0x12, 0x34, 0x56, 0x78, 0x90,
			},
			wantError: false,
		},
		{
			name:  "shorter input padded",
			input: "0x1234",
			expected: [20]byte{
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x12, 0x34,
			},
			wantError: false,
		},
		{"input too long", "0x123456789012345678901234567890123456789012", [20]byte{}, true},
		{"invalid hex", "0xZZ", [20]byte{}, true},
		{"empty input", "", [20]byte{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := HexStringToByte20(tt.input)

			if tt.wantError {
				assert.Error(t, err)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestHexStringToByte32(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		input     string
		expected  [32]byte
		wantError bool
	}{
		{
			name:  "valid 32-byte hex",
			input: "0x1234567890123456789012345678901234567890123456789012345678901234",
			expected: [32]byte{
				0x12, 0x34, 0x56, 0x78, 0x90, 0x12, 0x34, 0x56, 0x78, 0x90,
				0x12, 0x34, 0x56, 0x78, 0x90, 0x12, 0x34, 0x56, 0x78, 0x90,
				0x12, 0x34, 0x56, 0x78, 0x90, 0x12, 0x34, 0x56, 0x78, 0x90, 0x12, 0x34,
			},
			wantError: false,
		},
		{
			name:  "shorter input padded",
			input: "0x1234",
			expected: [32]byte{
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x12, 0x34,
			},
			wantError: false,
		},
		{"input too long", "0x123456789012345678901234567890123456789012345678901234567890123456", [32]byte{}, true},
		{"invalid hex", "0xZZ", [32]byte{}, true},
		{"empty input", "", [32]byte{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := HexStringToByte32(tt.input)

			if tt.wantError {
				assert.Error(t, err)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestHexStringToByteArray(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		input     string
		expected  []byte
		wantError bool
	}{
		{"valid hex with prefix", "0x1234", []byte{0x12, 0x34}, false},
		{"valid hex without prefix", "1234", []byte{0x12, 0x34}, false},
		{"empty string", "", []byte{}, false},
		{"empty with prefix", "0x", []byte{}, false},
		{"single byte", "0xFF", []byte{0xFF}, false},
		{"invalid hex", "0xZZ", nil, true},
		{"odd length", "0x123", nil, true},
		{"all zeros", "0x0000", []byte{0x00, 0x00}, false},
		{"all ones", "0xFFFF", []byte{0xFF, 0xFF}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := HexStringToByteArray(tt.input)

			if tt.wantError {
				assert.Error(t, err)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestHexStringToInt32(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		input     string
		expected  [32]int
		wantError bool
	}{
		{
			name:  "valid 32 byte hex",
			input: "0x0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
			expected: [32]int{
				0x01,
				0x23,
				0x45,
				0x67,
				0x89,
				0xab,
				0xcd,
				0xef,
				0x01,
				0x23,
				0x45,
				0x67,
				0x89,
				0xab,
				0xcd,
				0xef,
				0x01,
				0x23,
				0x45,
				0x67,
				0x89,
				0xab,
				0xcd,
				0xef,
				0x01,
				0x23,
				0x45,
				0x67,
				0x89,
				0xab,
				0xcd,
				0xef,
			},
			wantError: false,
		},
		{
			name:  "valid hex without prefix",
			input: "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
			expected: [32]int{
				0x01,
				0x23,
				0x45,
				0x67,
				0x89,
				0xab,
				0xcd,
				0xef,
				0x01,
				0x23,
				0x45,
				0x67,
				0x89,
				0xab,
				0xcd,
				0xef,
				0x01,
				0x23,
				0x45,
				0x67,
				0x89,
				0xab,
				0xcd,
				0xef,
				0x01,
				0x23,
				0x45,
				0x67,
				0x89,
				0xab,
				0xcd,
				0xef,
			},
			wantError: false,
		},
		{
			name:  "all zeros",
			input: "0x0000000000000000000000000000000000000000000000000000000000000000",
			expected: [32]int{
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
			},
			wantError: false,
		},
		{
			name:  "all 255s",
			input: "0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
			expected: [32]int{
				255,
				255,
				255,
				255,
				255,
				255,
				255,
				255,
				255,
				255,
				255,
				255,
				255,
				255,
				255,
				255,
				255,
				255,
				255,
				255,
				255,
				255,
				255,
				255,
				255,
				255,
				255,
				255,
				255,
				255,
				255,
				255,
			},
			wantError: false,
		},
		{
			name:  "shorter than 32 bytes",
			input: "0x1234",
			expected: [32]int{
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0x12,
				0x34,
			},
			wantError: false,
		},
		{
			name:      "longer than 32 bytes",
			input:     "0x123456789012345678901234567890123456789012345678901234567890123456789012",
			expected:  [32]int{},
			wantError: true,
		},
		{
			name:      "invalid hex",
			input:     "0xZZ",
			expected:  [32]int{},
			wantError: true,
		},
		{
			name:      "odd length",
			input:     "0x123",
			expected:  [32]int{},
			wantError: true,
		},
		{
			name:  "empty string",
			input: "",
			expected: [32]int{
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
			},
			wantError: false,
		},
		{
			name:  "only prefix",
			input: "0x",
			expected: [32]int{
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := HexStringToInt32(tt.input)

			if tt.wantError {
				assert.Error(t, err)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSafeInt64ToUint64(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		input     int64
		expected  uint64
		wantError bool
	}{
		{
			name:      "positive input",
			input:     1,
			expected:  1,
			wantError: false,
		},
		{
			name:      "negative input",
			input:     -1,
			expected:  0,
			wantError: true,
		},
		{
			name:      "zero input",
			input:     0,
			expected:  0,
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := SafeInt64ToUint64(tt.input)

			if tt.wantError {
				assert.Error(t, err)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}
