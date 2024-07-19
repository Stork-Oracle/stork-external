package main

import (
	"crypto/ecdsa"
	"encoding/hex"
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

func stringToByte32(input string) ([32]byte, error) {
	var result [32]byte

	// Remove the "0x" prefix unconditionally
	input = strings.TrimPrefix(input, "0x")

	// Decode the hex string to bytes
	bytes, err := hex.DecodeString(input)
	if err != nil {
		return result, err
	}

	// Check the length and copy to the result
	if len(bytes) > 32 {
		copy(result[:], bytes[len(bytes)-32:])
	} else {
		copy(result[32-len(bytes):], bytes)
	}

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
