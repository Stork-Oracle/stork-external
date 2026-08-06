package evm

import (
	"bytes"
	"encoding/hex"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompressCalldata(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "selector literals", input: "11223344", expected: "eeddccbb"},
		{name: "zero run", input: "00000000", expected: "fffc"},
		{name: "ff run", input: "ffffffff", expected: "ff7c"},
		{name: "mixed", input: "12345678000001ff", expected: "edcba9870001010080"},
		{name: "split 129 zeros", input: strings.Repeat("00", 129), expected: "ff80ffff"},
		{name: "split 33 ff bytes", input: strings.Repeat("ff", 33), expected: "ff60ff7f"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			input, err := hex.DecodeString(tt.input)
			require.NoError(t, err)

			assert.Equal(t, tt.expected, hex.EncodeToString(compressCalldata(input)))
		})
	}
}

func TestCalldataCompressionRoundTrip(t *testing.T) {
	t.Parallel()

	deterministic := make([]byte, 1024)
	for i := range deterministic {
		switch {
		case i%17 == 0:
			deterministic[i] = 0xff
		case i%5 == 0:
			deterministic[i] = 0x00
		default:
			deterministic[i] = byte(i)
		}
	}

	tests := []struct {
		name  string
		input []byte
	}{
		{name: "empty", input: []byte{}},
		{name: "short", input: []byte{0x12, 0x00, 0xff}},
		{name: "mixed runs", input: []byte{0x12, 0x34, 0x56, 0x78, 0x00, 0x00, 0x01, 0xff}},
		{name: "deterministic 1 KiB", input: deterministic},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			compressed := compressCalldata(tt.input)
			assert.Equal(t, tt.input, decompressCalldataForTest(t, compressed))
		})
	}
}

func decompressCalldataForTest(t *testing.T, compressed []byte) []byte {
	t.Helper()

	encoded := bytes.Clone(compressed)
	for i := 0; i < 4 && i < len(encoded); i++ {
		encoded[i] = ^encoded[i]
	}

	decompressed := make([]byte, 0, len(encoded))
	for i := 0; i < len(encoded); {
		value := encoded[i]
		i++
		if value != 0x00 {
			decompressed = append(decompressed, value)
			continue
		}

		require.Less(t, i, len(encoded), "run marker must have a control byte")
		control := encoded[i]
		i++
		runLength := int(control&0x7f) + 1
		if control&0x80 != 0 {
			decompressed = append(decompressed, bytes.Repeat([]byte{0xff}, runLength)...)
		} else {
			decompressed = append(decompressed, make([]byte, runLength)...)
		}
	}

	return decompressed
}
