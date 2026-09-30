package evm

const (
	zeroByte   byte = 0x00
	ffByte     byte = 0xff
	ffRunMask  byte = 0x80
	maxZeroRun      = 128
	maxFFRun        = 32
)

// compressCalldata applies Solady LibZip's selective run-length encoding.
func compressCalldata(data []byte) []byte {
	compressed := make([]byte, 0, len(data))
	for i := 0; i < len(data); {
		value := data[i]

		var (
			limit   int
			control byte
		)

		switch value {
		case zeroByte:
			limit = maxZeroRun
		case ffByte:
			limit = maxFFRun
			control = ffRunMask
		default:
			compressed = append(compressed, value)
			i++

			continue
		}

		runLength := 1
		for i+runLength < len(data) && data[i+runLength] == value && runLength < limit {
			runLength++
		}

		compressed = append(compressed, zeroByte, control|byte(runLength-1))
		i += runLength
	}

	for i := 0; i < 4 && i < len(compressed); i++ {
		compressed[i] = ^compressed[i]
	}

	return compressed
}
