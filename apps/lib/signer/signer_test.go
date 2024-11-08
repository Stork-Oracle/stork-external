package signer

import (
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestSigner_SignPublisherPrice_Evm(t *testing.T) {
	signer, err := NewEvmSigner("0x8b558d5fc31eb64bb51d44b4b28658180e96764d5d5ac68e6d124f86f576d9de", zerolog.Logger{})
	if err != nil {
		t.Fatalf("error creating signer: %v", err)
	}
	expectedTimestampedSig := &TimestampedSignature[*EvmSignature]{
		Timestamp: 1710191092123456789,
		MsgHash:   "0x4a8e2a9c736a3a2e315facf28ba95e126e37b57646481078e4f0809262c6560b",
		Signature: &EvmSignature{
			R: "0x14e378dcf486b15c157fb6af80fc275b895bd1cae818fc4597a6b4a1571a831e",
			S: "0x79b4823a159988c04576ff71bc3ca168a631ac666094b0f4157e59b2892f6490",
			V: "0x1b",
		},
	}
	signedPriceUpdate, assetId, err := signer.SignPublisherPrice(1710191092123456789, "BTCUSDMARK", "72147681412670819000000")
	assert.NoError(t, err)
	assert.Equal(t, "BTCUSDMARK", assetId)
	assert.Equal(t, expectedTimestampedSig, signedPriceUpdate)
}

func TestSigner_SignPublisherPrice_Stark(t *testing.T) {
	signer, err := NewStarkSigner("0x66253bdeb3c1a235cf4376611e3a14474e2c00fd2fb225f9a388faae7fb095a", "0x418d3fd8219a2cf32a00d458f61802d17f01c5bcde5a4f82008ee4a7c8e9a06", "czowx", zerolog.Logger{})
	if err != nil {
		t.Fatalf("error creating signer: %v", err)
	}
	expectedTimestampedSig := &TimestampedSignature[*StarkSignature]{
		Timestamp: 1708940577123456789,
		MsgHash:   "0x7cc1cf795d076cfff8b5920adb2dcc0d13813ed4519220a36d693e6084abe1c",
		Signature: &StarkSignature{
			R: "0x60bbbb4142bca69a5278ecccb59964e3449e43915b02e5c729b9752a16309ac",
			S: "0x4cdbe54b985f6fb4495398f94554883ead7cbb983597dc7ea8b9e32dfe95c27",
		},
	}

	signedPriceUpdate, assetId, err := signer.SignPublisherPrice(1708940577123456789, "DYDXUSD", "3335950349880000000")
	assert.NoError(t, err)
	assert.Equal(t, "0x44594458555344000000000000000000637a6f7778", assetId)
	assert.Equal(t, expectedTimestampedSig, signedPriceUpdate)
}

func TestSigner_SignAuth_Evm(t *testing.T) {
	signer, err := NewEvmAuthSigner("0x8b558d5fc31eb64bb51d44b4b28658180e96764d5d5ac68e6d124f86f576d9de", zerolog.Logger{})
	if err != nil {
		t.Fatalf("error creating signer: %v", err)
	}
	expectedAuthSignature := "{\"r\":\"0x2bde80c32c372aaf187b793d188ac13f7f1c92ec0121dc99b57ebfbfda74cecf\",\"s\":\"0x06d37333f3b56864090d77b7fe3efb815ced8270bfb47cbc3f806d957063bf3a\",\"v\":\"0x1b\"}"
	signedAuth, err := signer.SignAuth(1710191092123456789)
	assert.NoError(t, err)
	assert.Equal(t, expectedAuthSignature, signedAuth)
}

func TestSigner_SignAuth_Stark(t *testing.T) {
	signer, err := NewStarkAuthSigner("0x66253bdeb3c1a235cf4376611e3a14474e2c00fd2fb225f9a388faae7fb095a", "0x418d3fd8219a2cf32a00d458f61802d17f01c5bcde5a4f82008ee4a7c8e9a06", zerolog.Logger{})
	if err != nil {
		t.Fatalf("error creating signer: %v", err)
	}
	expectedAuthSignature := "{\"r\":\"0x6d317d0c403d4bb822db27843f7cca56f5922863ced48b380e6c4494c7d23a7\",\"s\":\"0x296da7fd09ed7e436a91d5667fa7d5f0f969d739231c2ba1fa00aa364b2dfe2\"}"
	signedPriceUpdate, err := signer.SignAuth(1708940577123456789)
	assert.NoError(t, err)
	assert.Equal(t, expectedAuthSignature, signedPriceUpdate)
}

func BenchmarkSigner_SignPublisherPrice_Evm(b *testing.B) {
	signer, err := NewEvmSigner("0x8b558d5fc31eb64bb51d44b4b28658180e96764d5d5ac68e6d124f86f576d9de", zerolog.Logger{})
	if err != nil {
		b.Fatalf("error creating signer: %v", err)
	}
	for i := 0; i < b.N; i++ {
		signer.SignPublisherPrice(1710191092123456789, "BTCUSDMARK", "72147681412670819000000")
	}
}

func BenchmarkSigner_SignPublisherPrice_Stark(b *testing.B) {
	signer, err := NewStarkSigner("0x66253bdeb3c1a235cf4376611e3a14474e2c00fd2fb225f9a388faae7fb095a", "0x418d3fd8219a2cf32a00d458f61802d17f01c5bcde5a4f82008ee4a7c8e9a06", "czowx", zerolog.Logger{})
	if err != nil {
		b.Fatalf("error creating signer: %v", err)
	}
	for i := 0; i < b.N; i++ {
		signer.SignPublisherPrice(1708940577123456789, "DYDXUSD", "3335950349880000000")
	}
}
