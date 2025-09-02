package pusher

import (
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"strings"
)

var (
	ErrInputTooLong  = errors.New("input string too long for specified length")
	ErrNegativeInput = errors.New("input is negative")
	ErrInputTooLarge = errors.New("input is too large for int64")
)

// Pluralize returns an empty string if n is 1, otherwise returns "s".
func Pluralize(n int) string {
	if n == 1 {
		return ""
	}

	return "s"
}

// hexStringToBytes decodes a hex string to a byte slice.
func hexStringToBytes(input string) ([]byte, error) {
	// Remove the "0x" prefix unconditionally
	input = strings.TrimPrefix(input, "0x")

	bytes, err := hex.DecodeString(input)
	if err != nil {
		return nil, fmt.Errorf("failed to decode hex string: %w", err)
	}

	return bytes, nil
}

// hexStringToFixedBytes converts a hex string to a fixed-length byte array.
func hexStringToFixedBytes(input string, length int) ([]byte, error) {
	bytes, err := hexStringToBytes(input)
	if err != nil {
		return nil, err
	}

	if len(bytes) > length {
		return nil, ErrInputTooLong
	}

	// Create a fixed-size byte array and copy the input bytes into it
	result := make([]byte, length)
	copy(result[length-len(bytes):], bytes)

	return result, nil
}

// HexStringToByte20 converts a hex string to a [20]byte array.
func HexStringToByte20(input string) ([20]byte, error) {
	var result [20]byte

	//nolint:mnd // 20 is not magic here, but rather is clearly tied to the purpose of this function
	bytes, err := hexStringToFixedBytes(input, 20)
	if err != nil {
		return result, err
	}

	copy(result[:], bytes)

	return result, nil
}

// HexStringToByte32 converts a hex string to a [32]byte array.
func HexStringToByte32(input string) ([32]byte, error) {
	var result [32]byte

	//nolint:mnd // 32 is not magic here, but rather is clearly tied to the purpose of this function
	bytes, err := hexStringToFixedBytes(input, 32)
	if err != nil {
		return result, err
	}

	copy(result[:], bytes)

	return result, nil
}

// HexStringToByteArray converts a hex string to a byte array.
func HexStringToByteArray(hexString string) ([]byte, error) {
	hexString = strings.TrimPrefix(hexString, "0x")

	array, err := hex.DecodeString(hexString)
	if err != nil {
		return nil, fmt.Errorf("failed to decode hex string: %w", err)
	}

	return array, nil
}

// HexStringToInt32 converts a hex string to a [32]int array.
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

func SafeInt64ToUint64(input int64) (uint64, error) {
	if input < 0 {
		return 0, ErrNegativeInput
	}

	return uint64(input), nil
}

func SafeUint64ToInt64(input uint64) (int64, error) {
	if input > math.MaxInt64 {
		return 0, ErrInputTooLarge
	}

	return int64(input), nil
}
