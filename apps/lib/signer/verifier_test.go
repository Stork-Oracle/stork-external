package signer

import (
	"testing"

	"github.com/rs/zerolog"
)

func TestEvmVerifier_VerifyPublisherPrice(t *testing.T) {
	verifier := NewEvmVerifier(zerolog.Logger{})

	signature := EvmSignature{
		R: "0x14e378dcf486b15c157fb6af80fc275b895bd1cae818fc4597a6b4a1571a831e",
		S: "0x79b4823a159988c04576ff71bc3ca168a631ac666094b0f4157e59b2892f6490",
		V: "0x1b",
	}
	pubKey := PublisherKey("0x99e295e85cb07c16b7bb62a44df532a7f2620237")

	err := verifier.VerifyPublisherPrice(1710191092123456789, "BTCUSDMARK", "72147681412670819000000", pubKey, signature)
	if err != nil {
		t.Error(err)
	}

	// change the price slightly
	err = verifier.VerifyPublisherPrice(1710191092123456789, "BTCUSDMARK", "72147681412670719000000", pubKey, signature)
	if err == nil {
		t.Fail()
	}
}

func TestStarkVerifier_VerifyPublisherPrice(t *testing.T) {
	verifier := NewStarkVerifier(zerolog.Logger{})

	signature := StarkSignature{
		R: "0x60bbbb4142bca69a5278ecccb59964e3449e43915b02e5c729b9752a16309ac",
		S: "0x4cdbe54b985f6fb4495398f94554883ead7cbb983597dc7ea8b9e32dfe95c27",
	}
	pubKey := PublisherKey("0x418d3fd8219a2cf32a00d458f61802d17f01c5bcde5a4f82008ee4a7c8e9a06")

	err := verifier.VerifyPublisherPrice(1708940577123456789, "0x44594458555344000000000000000000637a6f7778", "3335950349880000000", pubKey, signature)
	if err != nil {
		t.Error(err)
	}

	// change the price slightly
	err = verifier.VerifyPublisherPrice(1708940577123456789, "0x44594458555344000000000000000000637a6f7778", "3335950348880000000", pubKey, signature)
	if err == nil {
		t.Fail()
	}
}
