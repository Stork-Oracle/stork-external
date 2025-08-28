package pusher

import (
	"encoding/hex"
	"errors"
	"strings"
)

func Pluralize(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}

// stringToBytes decodes a hex string to a byte slice.
func hexStringToBytes(input string) ([]byte, error) {
	// Remove the "0x" prefix unconditionally
	input = strings.TrimPrefix(input, "0x")

	// Decode the hex string to bytes
	return hex.DecodeString(input)
}

// stringToFixedBytes converts a hex string to a fixed-length byte array.
func hexStringToFixedBytes(input string, length int) ([]byte, error) {
	bytes, err := hexStringToBytes(input)
	if err != nil {
		return nil, err
	}

	if len(bytes) > length {
		return nil, errors.New("input string too long for specified length")
	}

	// Create a fixed-size byte array and copy the input bytes into it
	result := make([]byte, length)
	copy(result[length-len(bytes):], bytes)

	return result, nil
}

// stringToByte20 converts a hex string to a [20]byte array.
func HexStringToByte20(input string) ([20]byte, error) {
	var result [20]byte

	bytes, err := hexStringToFixedBytes(input, 20)
	if err != nil {
		return result, err
	}

	copy(result[:], bytes)
	return result, nil
}

// stringToByte32 converts a hex string to a [32]byte array.
func HexStringToByte32(input string) ([32]byte, error) {
	var result [32]byte

	bytes, err := hexStringToFixedBytes(input, 32)
	if err != nil {
		return result, err
	}

	copy(result[:], bytes)
	return result, nil
}

func HexStringToByteArray(hexString string) ([]byte, error) {
	hexString = strings.TrimPrefix(hexString, "0x")
	return hex.DecodeString(hexString)
}

func HexStringToInt32(hexString string) ([32]int, error) {
	bytes, err := HexStringToByte32(hexString)
	if err != nil {
		return [32]int{}, err
	}
	var result [32]int
	for i, b := range bytes {
		result[i] = int(b)
	}
	return result, nil
}
