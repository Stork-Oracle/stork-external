// Package testutil provides utilities for testing.
package testutil

import (
	"math/big"

	"github.com/Stork-Oracle/stork-external/apps/chain_pusher/pkg/types"
)

type EvmSignatureBytes struct {
	R [32]byte
	S [32]byte
	V byte
}

// TimestampedSignatureBytes is a mirror of the TimestampedSignature struct, but with changes:
// bytes instead of hex strings, big.Ints instead of int strings, and some fields removed.
type TimestampedSignatureBytes struct {
	Signature     EvmSignatureBytes
	TimestampNano uint64
	MsgHash       string
}

// StorkSignedPriceBytes is a mirror of the StorkSignedPrice struct, but with changes:
// bytes instead of hex strings, big.Ints instead of int strings, and some fields removed.
type StorkSignedPriceBytes struct {
	EncodedAssetID       [32]byte
	QuantizedPrice       *big.Int
	TimestampedSignature TimestampedSignatureBytes
	PublisherMerkleRoot  [32]byte
	StorkCalculationAlg  [32]byte
}

// AggregatedSignedPriceBytes is a mirror of the AggregatedSignedPrice struct, but with changes:
// bytes instead of hex strings, big.Ints instead of int strings, and some fields removed.
type AggregatedSignedPriceBytes struct {
	TimestampNano    uint64
	StorkSignedPrice *StorkSignedPriceBytes
}

type PriceCase struct {
	Name       string
	Price      types.AggregatedSignedPrice
	PriceBytes *AggregatedSignedPriceBytes
	WantError  bool
}

// defaultPrice returns a default price and it's bytes representation.
func defaultPrice() (types.AggregatedSignedPrice, AggregatedSignedPriceBytes) {
	return types.AggregatedSignedPrice{
			StorkSignedPrice: &types.StorkSignedPrice{
				EncodedAssetID: types.EncodedAssetID(
					"0x1234567890123456789012345678901234567890123456789012345678901234",
				),
				QuantizedPrice:      types.QuantizedPrice("1000000000000000000"),
				PublisherMerkleRoot: "0xe5ff773b0316059c04aa157898766731017610dcbeede7d7f169bfeaab7cc318",
				StorkCalculationAlg: types.StorkCalculationAlg{
					Checksum: "0x9be7e9f9ed459417d96112a7467bd0b27575a2c7847195c68f805b70ce1795ba",
				},
				TimestampedSignature: types.TimestampedSignature{
					TimestampNano: uint64(1722632569208762117),
					Signature: types.EvmSignature{
						R: "0xb9b3c9f80a355bd0cd6f609fff4a4b15fa4e3b4632adabb74c020f5bcd240741",
						S: "0x16fab526529ac795108d201832cff8c2d2b1c710da6711fe9f7ab288a7149758",
						V: "0x1c",
					},
				},
			},
		}, AggregatedSignedPriceBytes{
			StorkSignedPrice: &StorkSignedPriceBytes{
				EncodedAssetID: [32]byte{
					0x12,
					0x34,
					0x56,
					0x78,
					0x90,
					0x12,
					0x34,
					0x56,
					0x78,
					0x90,
					0x12,
					0x34,
					0x56,
					0x78,
					0x90,
					0x12,
					0x34,
					0x56,
					0x78,
					0x90,
					0x12,
					0x34,
					0x56,
					0x78,
					0x90,
					0x12,
					0x34,
					0x56,
					0x78,
					0x90,
					0x12,
					0x34,
				},
				QuantizedPrice: big.NewInt(1000000000000000000),
				TimestampedSignature: TimestampedSignatureBytes{
					TimestampNano: uint64(1722632569208762117),
					Signature: EvmSignatureBytes{
						R: [32]byte{
							0xb9,
							0xb3,
							0xc9,
							0xf8,
							0x0a,
							0x35,
							0x5b,
							0xd0,
							0xcd,
							0x6f,
							0x60,
							0x9f,
							0xff,
							0x4a,
							0x4b,
							0x15,
							0xfa,
							0x4e,
							0x3b,
							0x46,
							0x32,
							0xad,
							0xab,
							0xb7,
							0x4c,
							0x02,
							0x0f,
							0x5b,
							0xcd,
							0x24,
							0x07,
							0x41,
						},
						S: [32]byte{
							0x16,
							0xfa,
							0xb5,
							0x26,
							0x52,
							0x9a,
							0xc7,
							0x95,
							0x10,
							0x8d,
							0x20,
							0x18,
							0x32,
							0xcf,
							0xf8,
							0xc2,
							0xd2,
							0xb1,
							0xc7,
							0x10,
							0xda,
							0x67,
							0x11,
							0xfe,
							0x9f,
							0x7a,
							0xb2,
							0x88,
							0xa7,
							0x14,
							0x97,
							0x58,
						},
						V: byte(28),
					},
				},
				PublisherMerkleRoot: [32]byte{
					0xe5,
					0xff,
					0x77,
					0x3b,
					0x03,
					0x16,
					0x05,
					0x9c,
					0x04,
					0xaa,
					0x15,
					0x78,
					0x98,
					0x76,
					0x67,
					0x31,
					0x01,
					0x76,
					0x10,
					0xdc,
					0xbe,
					0xed,
					0xe7,
					0xd7,
					0xf1,
					0x69,
					0xbf,
					0xea,
					0xab,
					0x7c,
					0xc3,
					0x18,
				},
				StorkCalculationAlg: [32]byte{
					0x9b,
					0xe7,
					0xe9,
					0xf9,
					0xed,
					0x45,
					0x94,
					0x17,
					0xd9,
					0x61,
					0x12,
					0xa7,
					0x46,
					0x7b,
					0xd0,
					0xb2,
					0x75,
					0x75,
					0xa2,
					0xc7,
					0x84,
					0x71,
					0x95,
					0xc6,
					0x8f,
					0x80,
					0x5b,
					0x70,
					0xce,
					0x17,
					0x95,
					0xba,
				},
			},
		}
}

// validPositivePriceCase is a test case where the price is positive.
func validPositivePriceCase() PriceCase {
	defaultPrice, defaultPriceBytes := defaultPrice()

	return PriceCase{
		Name:       "valid positive price",
		Price:      defaultPrice,
		PriceBytes: &defaultPriceBytes,
		WantError:  false,
	}
}

// validNegativePriceCase is a test case where the price is negative.
func validNegativePriceCase() PriceCase {
	negativeQuantizedPrice := "-1000000000000000000"
	negativeQuantizedPriceBigInt := big.NewInt(-1000000000000000000)

	defaultPrice, defaultPriceBytes := defaultPrice()
	defaultPrice.StorkSignedPrice.QuantizedPrice = types.QuantizedPrice(negativeQuantizedPrice)
	defaultPriceBytes.StorkSignedPrice.QuantizedPrice = negativeQuantizedPriceBigInt

	return PriceCase{
		Name:       "valid negative price",
		Price:      defaultPrice,
		PriceBytes: &defaultPriceBytes,
		WantError:  false,
	}
}

// validZeroPriceCase is a test case where the price is zero.
func validZeroPriceCase() PriceCase {
	zeroQuantizedPrice := "0"
	zeroQuantizedPriceBigInt := big.NewInt(0)

	defaultPrice, defaultPriceBytes := defaultPrice()
	defaultPrice.StorkSignedPrice.QuantizedPrice = types.QuantizedPrice(zeroQuantizedPrice)
	defaultPriceBytes.StorkSignedPrice.QuantizedPrice = zeroQuantizedPriceBigInt

	return PriceCase{
		Name:       "valid zero price",
		Price:      defaultPrice,
		PriceBytes: &defaultPriceBytes,
		WantError:  false,
	}
}

// invalidVSignatureCase is a test case where the V signature is invalid.
func invalidVSignatureCase() PriceCase {
	invalidV := "invalid"

	defaultPrice, _ := defaultPrice()
	defaultPrice.StorkSignedPrice.TimestampedSignature.Signature.V = invalidV

	return PriceCase{
		Name:       "invalid V signature",
		Price:      defaultPrice,
		PriceBytes: nil,
		WantError:  true,
	}
}

// invalidEncodedAssetIDCase is a test case where the encoded asset ID is invalid.
func invalidEncodedAssetIDCase() PriceCase {
	invalidEncodedAssetID := "invalid"

	defaultPrice, _ := defaultPrice()
	defaultPrice.StorkSignedPrice.EncodedAssetID = types.EncodedAssetID(invalidEncodedAssetID)

	return PriceCase{
		Name:       "invalid encoded asset ID",
		Price:      defaultPrice,
		PriceBytes: nil,
		WantError:  true,
	}
}

// StandardPriceCase returns a slice of standard price cases that should be tested against in interactor tests.
func StandardPriceCase() []PriceCase {
	return []PriceCase{
		validPositivePriceCase(),
		validNegativePriceCase(),
		validZeroPriceCase(),
		invalidVSignatureCase(),
		invalidEncodedAssetIDCase(),
	}
}
