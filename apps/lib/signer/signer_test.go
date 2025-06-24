package signer

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestSigner_SignPublisherPrice_Evm(t *testing.T) {
	signer, err := NewEvmSigner("0x28253097630cca4158c909efa1af971e7aa759eb3d966cdb34e50f5ca1916ac7", zerolog.Logger{})
	if err != nil {
		t.Fatalf("error creating signer: %v", err)
	}
	expectedTimestampedSig := &TimestampedSignature[*EvmSignature]{
		Timestamp: 1720730544719000064,
		MsgHash:   "0x94796ac50f614eaec60734ffb48577da56f6cb4d9cf4ce5c61c222f1d3693be1",
		Signature: &EvmSignature{
			R: "0x8ac298121624afad3057ec39bd5d7d08dbccd98453b67add7d871d94a18c3302",
			S: "0x3cf613d9bee0cbc01073ac7b23ca3e86eb34bc2bd5748f07cab984377b4291b3",
			V: "0x1c",
		},
	}
	signedPriceUpdate, assetId, err := signer.SignPublisherPrice(1720730544719000064, "BTCUSD", "60000000000000000000000")
	assert.NoError(t, err)
	assert.Equal(t, "BTCUSD", assetId)
	assert.Equal(t, expectedTimestampedSig, signedPriceUpdate)

	// negative test
	expectedTimestampedSig = &TimestampedSignature[*EvmSignature]{
		Timestamp: 1720730544719000064,
		MsgHash:   "0x2aa596404bdb22d180d4a6d297a7781aa9590300ac66124f59ece77c25acad4e",
		Signature: &EvmSignature{
			R: "0xf7f78a5074adc80dccc6a5abfbf47b993ff4ee50b6e09c8db08a0d99b37b9637",
			S: "0x5b057e5d67bb77eab748e47653bdf9b34225a7de1f1af333e953bc79f6991212",
			V: "0x1c",
		},
	}

	signedNegativePriceUpdate, assetId, err := signer.SignPublisherPrice(1720730544719000064, "BTCUSD", "-60000000000000000000000")
	assert.NoError(t, err)
	assert.NotEqual(t, signedPriceUpdate.Signature, signedNegativePriceUpdate.Signature)
	assert.Equal(t, "BTCUSD", assetId)
	assert.Equal(t, expectedTimestampedSig, signedNegativePriceUpdate)
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

	// negative
	// WARNING: While this test demonstrates that negative and positive prices yield different signatures,
	// this is not a guarantee that payloads with negative prices are signed correctly.
	// Testing against starknet should be done with negative prices in order to confirm that the stark signer can
	// properly handle negative prices.
	expectedNegativeTimestampedSig := &TimestampedSignature[*StarkSignature]{
		Timestamp: 1708940577123456789,
		MsgHash:   "0x223b3bf417894341325c99275acb14714f3f94caf7386f434dafd496443eb1",
		Signature: &StarkSignature{
			R: "0x9dffaea089d280d45180cbbddde9336a4e2c926234ae4d58ae9be8878821e6",
			S: "0x6777f741610f8ebe69707ab12bda9c6efc03cf6aafe919b187d226ac8ece6b8",
		},
	}
	signedNegativePriceUpdate, assetId, err := signer.SignPublisherPrice(1708940577123456789, "DYDXUSD", "-3335950349880000000")
	assert.NoError(t, err)
	assert.Equal(t, "0x44594458555344000000000000000000637a6f7778", assetId)
	assert.NotEqual(t, signedPriceUpdate.Signature, signedNegativePriceUpdate.Signature)
	assert.Equal(t, expectedNegativeTimestampedSig, signedNegativePriceUpdate)

	// long asset name
	expectedTimestampedSig = &TimestampedSignature[*StarkSignature]{
		Timestamp: 1708940577123456789,
		MsgHash:   "0x7acab52851a7b006dbf5d350f8dda7438f843204a3612030b7b0178ff93b37b",
		Signature: &StarkSignature{
			R: "0x3fbe61ab618ed32e4d7a9cb3e9c9be8f4a64128eba6ddd12cd6058bdae546c4",
			S: "0x31a930c2989244043c86b138ea75ba2bbb18f51012c6b00fe8e4d93ce03c030",
		},
	}
	signedPriceUpdate, assetId, err = signer.SignPublisherPrice(1708940577123456789, "DJTWINYESUSDTWAP480", "3335950349880000000")
	assert.NoError(t, err)
	assert.Equal(t, "0x444a5457494e59455355534454574150343830637a6f7778", assetId)
	assert.Equal(t, expectedTimestampedSig, signedPriceUpdate)
}

func TestSigner_SignAuth_Evm(t *testing.T) {
	signer, err := NewEvmAuthSigner("0x8b558d5fc31eb64bb51d44b4b28658180e96764d5d5ac68e6d124f86f576d9de", zerolog.Logger{})
	if err != nil {
		t.Fatalf("error creating signer: %v", err)
	}
	signatureStr, err := signer.SignAuth(1710191092123456789)
	assert.NoError(t, err)
	assert.Equal(t, "0x2bde80c32c372aaf187b793d188ac13f7f1c92ec0121dc99b57ebfbfda74cecf06d37333f3b56864090d77b7fe3efb815ced8270bfb47cbc3f806d957063bf3a1b", signatureStr)
}

func TestSigner_SignAuth_Stark(t *testing.T) {
	signer, err := NewStarkAuthSigner("0x66253bdeb3c1a235cf4376611e3a14474e2c00fd2fb225f9a388faae7fb095a", "0x418d3fd8219a2cf32a00d458f61802d17f01c5bcde5a4f82008ee4a7c8e9a06", zerolog.Logger{})
	if err != nil {
		t.Fatalf("error creating signer: %v", err)
	}
	signatureStr, err := signer.SignAuth(1708940577123456789)
	assert.NoError(t, err)
	assert.Equal(t, "0x06d317d0c403d4bb822db27843f7cca56f5922863ced48b380e6c4494c7d23a70296da7fd09ed7e436a91d5667fa7d5f0f969d739231c2ba1fa00aa364b2dfe2", signatureStr)
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

func TestBigIntBytesToTwosComplement(t *testing.T) {
	// negative
	intString := "-17725899000000"
	intBigInt := new(big.Int)
	intBigInt.SetString(intString, 10)

	twosComplement := bigIntToTwosComplement32(intBigInt)
	assert.Equal(t, "ffffffffffffffffffffffffffffffffffffffffffffffffffffefe0de163740", hex.EncodeToString(twosComplement))

	// positive
	intString = "12500000000000"
	intBigInt.SetString(intString, 10)

	twosComplement = bigIntToTwosComplement32(intBigInt)
	assert.Equal(t, "00000000000000000000000000000000000000000000000000000b5e620f4800", hex.EncodeToString(twosComplement))
}
