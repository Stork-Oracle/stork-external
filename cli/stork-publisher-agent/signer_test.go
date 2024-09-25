package stork_publisher_agent

import (
	"math/big"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestSigner_GetSignedPriceUpdate_Evm(t *testing.T) {
	config := NewStorkPublisherAgentConfig(
		[]SignatureType{EvmSignatureType},
		"0x8b558d5fc31eb64bb51d44b4b28658180e96764d5d5ac68e6d124f86f576d9de",
		"0x99e295e85cb07c16b7bb62a44df532a7f2620237",
		"",
		"",
		time.Duration(0),
		time.Duration(0),
		0.0,
		"faked",
		"",
		time.Duration(0),
		time.Duration(0),
		"",
		"",
		"",
		"",
		time.Duration(0),
		time.Duration(0),
		false,
		0,
	)
	signer, err := NewSigner[*EvmSignature](*config, EvmSignatureType, zerolog.Logger{})
	if err != nil {
		t.Fatalf("error creating signer: %v", err)
	}
	value, _ := new(big.Float).SetString("72147.681412670819")
	valueUpdate := ValueUpdate{
		PublishTimestamp: 1710191092123456789,
		Asset:            "BTCUSDMARK",
		Value:            value,
	}
	expectedSignedPriceUpdate := SignedPriceUpdate[*EvmSignature]{
		OracleId: "faked",
		AssetId:  "BTCUSDMARK",
		Trigger:  ClockTriggerType,
		SignedPrice: SignedPrice[*EvmSignature]{
			PublisherKey:    "0x99e295e85cb07C16B7BB62A44dF532A7F2620237",
			ExternalAssetId: "BTCUSDMARK",
			SignatureType:   EvmSignatureType,
			QuantizedPrice:  "72147681412670819000000",
			TimestampedSignature: TimestampedSignature[*EvmSignature]{
				Timestamp: 1710191092123456789,
				MsgHash:   "0x4a8e2a9c736a3a2e315facf28ba95e126e37b57646481078e4f0809262c6560b",
				Signature: &EvmSignature{
					R: "0x14e378dcf486b15c157fb6af80fc275b895bd1cae818fc4597a6b4a1571a831e",
					S: "0x79b4823a159988c04576ff71bc3ca168a631ac666094b0f4157e59b2892f6490",
					V: "0x1b",
				},
			},
		},
	}
	signedPriceUpdate := signer.GetSignedPriceUpdate(valueUpdate, ClockTriggerType)

	assert.Equal(t, expectedSignedPriceUpdate, signedPriceUpdate)
}

func TestSigner_SignStark(t *testing.T) {
	config := NewStorkPublisherAgentConfig(
		[]SignatureType{StarkSignatureType},
		"",
		"",
		"0x66253bdeb3c1a235cf4376611e3a14474e2c00fd2fb225f9a388faae7fb095a",
		"0x418d3fd8219a2cf32a00d458f61802d17f01c5bcde5a4f82008ee4a7c8e9a06",
		time.Duration(0),
		time.Duration(0),
		0.0,
		"czowx",
		"",
		time.Duration(0),
		time.Duration(0),
		"",
		"",
		"",
		"",
		time.Duration(0),
		time.Duration(0),
		false,
		0,
	)

	signer, err := NewSigner[*StarkSignature](*config, StarkSignatureType, zerolog.Logger{})
	if err != nil {
		t.Fatalf("error creating signer: %v", err)
	}
	value, success := new(big.Float).SetString("3.33595034988")
	if !success {
		t.Fatalf("failed to parse float value")
	}
	valueUpdate := ValueUpdate{
		PublishTimestamp: 1708940577123456789,
		Asset:            "DYDXUSD",
		Value:            value,
	}
	expectedSignedPriceUpdate := SignedPriceUpdate[*StarkSignature]{
		OracleId: "czowx",
		AssetId:  "DYDXUSD",
		Trigger:  ClockTriggerType,
		SignedPrice: SignedPrice[*StarkSignature]{
			PublisherKey:    "0x418d3fd8219a2cf32a00d458f61802d17f01c5bcde5a4f82008ee4a7c8e9a06",
			SignatureType:   StarkSignatureType,
			ExternalAssetId: "0x44594458555344000000000000000000637a6f7778",
			QuantizedPrice:  "3335950349880000000",
			TimestampedSignature: TimestampedSignature[*StarkSignature]{
				Timestamp: 1708940577123456789,
				MsgHash:   "0x7cc1cf795d076cfff8b5920adb2dcc0d13813ed4519220a36d693e6084abe1c",
				Signature: &StarkSignature{
					R: "0x60bbbb4142bca69a5278ecccb59964e3449e43915b02e5c729b9752a16309ac",
					S: "0x4cdbe54b985f6fb4495398f94554883ead7cbb983597dc7ea8b9e32dfe95c27",
				},
			},
		},
	}

	signedPriceUpdate := signer.GetSignedPriceUpdate(valueUpdate, ClockTriggerType)
	assert.Equal(t, expectedSignedPriceUpdate, signedPriceUpdate)
}

func TestSigner_GetConnectionSignature_Stark(t *testing.T) {
	config := NewStorkPublisherAgentConfig(
		[]SignatureType{StarkSignatureType},
		"",
		"",
		"0x66253bdeb3c1a235cf4376611e3a14474e2c00fd2fb225f9a388faae7fb095a",
		"0x418d3fd8219a2cf32a00d458f61802d17f01c5bcde5a4f82008ee4a7c8e9a06",
		time.Duration(0),
		time.Duration(0),
		0.0,
		"czowx",
		"",
		time.Duration(0),
		time.Duration(0),
		"",
		"",
		"",
		"",
		time.Duration(0),
		time.Duration(0),
		false,
		0,
	)

	signer, err := NewSigner[*StarkSignature](*config, StarkSignatureType, zerolog.Logger{})
	if err != nil {
		t.Fatalf("error creating signer: %v", err)
	}

	msgHash, signature, err := signer.GetConnectionSignature(1727220712123123123, PublisherKey(config.StarkPublicKey))

	assert.Equal(t, "0x5178587ea35ba813ac6b04af0c79f533cb4fd68a7f3e491ed6f41cab70bb0ab", *msgHash)
	assert.Equal(t, "05413511ef95430d2cd6c65ed8d5d3086ac50416247948e171b335054afe597d060329a905f765d740d9fcfbbe833d301893d67e91f8b7534d08fa809f3b12bb", *signature)
}

func TestSigner_GetConnectionSignature_Evm(t *testing.T) {
	config := NewStorkPublisherAgentConfig(
		[]SignatureType{EvmSignatureType},
		"0x8b558d5fc31eb64bb51d44b4b28658180e96764d5d5ac68e6d124f86f576d9de",
		"0x99e295e85cb07c16b7bb62a44df532a7f2620237",
		"",
		"",
		time.Duration(0),
		time.Duration(0),
		0.0,
		"faked",
		"",
		time.Duration(0),
		time.Duration(0),
		"",
		"",
		"",
		"",
		time.Duration(0),
		time.Duration(0),
		false,
		0,
	)

	signer, err := NewSigner[*StarkSignature](*config, EvmSignatureType, zerolog.Logger{})
	if err != nil {
		t.Fatalf("error creating signer: %v", err)
	}

	msgHash, signature, err := signer.GetConnectionSignature(1727220712123123123, PublisherKey(config.EvmPublicKey))

	assert.Equal(t, "0xaa8a109b87b30e8dc780e05385ec76bd315310e4cc72220cba8ec97c41253685", *msgHash)
	assert.Equal(t, "052970fda7d9c8cd2e3a11bf01944e1552e21378530ffebdbafc10acb366f4da59fc04d6d2f4801640db1020f0f2e4cc95c71cd9ad933aa2a139b862eee3f9d400", *signature)
}

func BenchmarkSigner_SignEvm(b *testing.B) {
	config := NewStorkPublisherAgentConfig(
		[]SignatureType{EvmSignatureType},
		"0x8b558d5fc31eb64bb51d44b4b28658180e96764d5d5ac68e6d124f86f576d9de",
		"0x99e295e85cb07c16b7bb62a44df532a7f2620237",
		"",
		"",
		time.Duration(0),
		time.Duration(0),
		0.0,
		"faked",
		"",
		time.Duration(0),
		time.Duration(0),
		"",
		"",
		"",
		"",
		time.Duration(0),
		time.Duration(0),
		false,
		0,
	)
	signer, err := NewSigner[*EvmSignature](*config, EvmSignatureType, zerolog.Logger{})
	if err != nil {
		b.Fatalf("error creating signer: %v", err)
	}
	value, success := new(big.Float).SetString("72147.681412670819")
	if !success {
		b.Fatalf("failed to parse float value")
	}
	valueUpdate := ValueUpdate{
		PublishTimestamp: 1710191092123456789,
		Asset:            "BTCUSDMARK",
		Value:            value,
	}
	for i := 0; i < b.N; i++ {
		signer.GetSignedPriceUpdate(valueUpdate, ClockTriggerType)
	}
}

func BenchmarkSigner_SignStark(b *testing.B) {
	config := NewStorkPublisherAgentConfig(
		[]SignatureType{StarkSignatureType},
		"",
		"",
		"0x66253bdeb3c1a235cf4376611e3a14474e2c00fd2fb225f9a388faae7fb095a",
		"0x418d3fd8219a2cf32a00d458f61802d17f01c5bcde5a4f82008ee4a7c8e9a06",
		time.Duration(0),
		time.Duration(0),
		0.0,
		"czowx",
		"",
		time.Duration(0),
		time.Duration(0),
		"",
		"",
		"",
		"",
		time.Duration(0),
		time.Duration(0),
		false,
		0,
	)

	signer, err := NewSigner[*StarkSignature](*config, StarkSignatureType, zerolog.Logger{})
	if err != nil {
		b.Fatalf("error creating signer: %v", err)
	}
	value, _ := new(big.Float).SetString("3.33595034988")
	valueUpdate := ValueUpdate{
		PublishTimestamp: 1708940577123456789,
		Asset:            "DYDXUSD",
		Value:            value,
	}
	for i := 0; i < b.N; i++ {
		signer.GetSignedPriceUpdate(valueUpdate, ClockTriggerType)
	}
}
