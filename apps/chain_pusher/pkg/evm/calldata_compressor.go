package evm

const (
	maxZeroRun = 128
	maxFFRun   = 32
)

// compressCalldata applies Solady LibZip's selective run-length encoding.
func compressCalldata(data []byte) []byte {
	compressed := make([]byte, 0, len(data))
	for i := 0; i < len(data); {
		value := data[i]
		limit := 0
		control := byte(0)

		switch value {
		case 0x00:
			limit = maxZeroRun
		case 0xff:
			limit = maxFFRun
			control = 0x80
		default:
			compressed = append(compressed, value)
			i++

			continue
		}

		runLength := 1
		for i+runLength < len(data) && data[i+runLength] == value && runLength < limit {
			runLength++
		}

		compressed = append(compressed, 0x00, control|byte(runLength-1))
		i += runLength
	}

	for i := 0; i < 4 && i < len(compressed); i++ {
		compressed[i] = ^compressed[i]
	}

	return compressed
}
