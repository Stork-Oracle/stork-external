package chain_pusher

import (
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/crypto"
)

func pluralize(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}

// stringToBytes decodes a hex string to a byte slice.
func stringToBytes(input string) ([]byte, error) {
	// Remove the "0x" prefix unconditionally
	input = strings.TrimPrefix(input, "0x")

	// Decode the hex string to bytes
	return hex.DecodeString(input)
}

// stringToFixedBytes converts a hex string to a fixed-length byte array.
func stringToFixedBytes(input string, length int) ([]byte, error) {
	bytes, err := stringToBytes(input)
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
func stringToByte20(input string) ([20]byte, error) {
	var result [20]byte

	bytes, err := stringToFixedBytes(input, 20)
	if err != nil {
		return result, err
	}

	copy(result[:], bytes)
	return result, nil
}

// stringToByte32 converts a hex string to a [32]byte array.
func stringToByte32(input string) ([32]byte, error) {
	var result [32]byte

	bytes, err := stringToFixedBytes(input, 32)
	if err != nil {
		return result, err
	}

	copy(result[:], bytes)
	return result, nil
}

// For simplicity, this function assumes the mnemonic file contains the private key directly
func loadPrivateKey(mnemonicFile string) (*ecdsa.PrivateKey, error) {
	data, err := os.ReadFile(mnemonicFile)
	if err != nil {
		return nil, err
	}
	// Remove any trailing newline characters
	dataString := strings.TrimSpace(string(data))

	privateKey, err := crypto.HexToECDSA(dataString)
	if err != nil {
		return nil, err
	}
	return privateKey, nil
}
