package signer

import (
	"testing"
)

func TestEvmVerifier_VerifyPublisherPrice(t *testing.T) {
	signature := EvmSignature{
		R: "0x14e378dcf486b15c157fb6af80fc275b895bd1cae818fc4597a6b4a1571a831e",
		S: "0x79b4823a159988c04576ff71bc3ca168a631ac666094b0f4157e59b2892f6490",
		V: "0x1b",
	}

	err := VerifyEvmPublisherPrice(
		1710191092123456789,
		"BTCUSDMARK",
		"72147681412670819000000",
		"0x99e295e85cb07c16b7bb62a44df532a7f2620237",
		signature,
	)
	if err != nil {
		t.Error(err)
	}

	// changing the price causes the signature to be invalid
	err = VerifyEvmPublisherPrice(
		1710191092123456789,
		"BTCUSDMARK",
		"72147681412670719000000",
		"0x99e295e85cb07c16b7bb62a44df532a7f2620237",
		signature,
	)
	if err == nil {
		t.Fail()
	}

	// changing the asset id causes the signature to be invalid
	err = VerifyEvmPublisherPrice(
		1710191092123456789,
		"BTCUSDMARK2",
		"72147681412670819000000",
		"0x99e295e85cb07c16b7bb62a44df532a7f2620237",
		signature,
	)
	if err == nil {
		t.Fail()
	}

	// changing the timestamp causes the signature to be invalid
	err = VerifyEvmPublisherPrice(
		1710191192123456789,
		"BTCUSDMARK",
		"72147681412670819000000",
		"0x99e295e85cb07c16b7bb62a44df532a7f2620237",
		signature,
	)
	if err == nil {
		t.Fail()
	}

	// changing the publisher key causes the signature to be invalid
	err = VerifyEvmPublisherPrice(
		1710191092123456789,
		"BTCUSDMARK",
		"72147681412670819000000",
		"0x98e295e85cb07c16b7bb62a44df532a7f2620237",
		signature,
	)
	if err == nil {
		t.Fail()
	}
}

func TestStarkVerifier_VerifyPublisherPrice(t *testing.T) {
	signature := StarkSignature{
		R: "0x60bbbb4142bca69a5278ecccb59964e3449e43915b02e5c729b9752a16309ac",
		S: "0x4cdbe54b985f6fb4495398f94554883ead7cbb983597dc7ea8b9e32dfe95c27",
	}

	err := VerifyStarkPublisherPrice(
		1708940577123456789,
		"0x44594458555344000000000000000000637a6f7778",
		"3335950349880000000",
		"0x418d3fd8219a2cf32a00d458f61802d17f01c5bcde5a4f82008ee4a7c8e9a06",
		signature,
	)
	if err != nil {
		t.Error(err)
	}

	// changing the price causes the signature to be invalid
	err = VerifyStarkPublisherPrice(
		1708940577123456789,
		"0x44594458555344000000000000000000637a6f7778",
		"3335950348880000000",
		"0x418d3fd8219a2cf32a00d458f61802d17f01c5bcde5a4f82008ee4a7c8e9a06",
		signature,
	)
	if err == nil {
		t.Fail()
	}

	// changing the external asset id causes the signature to be invalid
	err = VerifyStarkPublisherPrice(
		1708940577123456789,
		"0x44594458555344000001000000000000637a6f7778",
		"3335950349880000000",
		"0x418d3fd8219a2cf32a00d458f61802d17f01c5bcde5a4f82008ee4a7c8e9a06",
		signature,
	)
	if err == nil {
		t.Fail()
	}

	// changing the timestamp causes the signature to be invalid
	err = VerifyStarkPublisherPrice(
		1708940576123456789,
		"0x44594458555344000000000000000000637a6f7778",
		"3335950349880000000",
		"0x418d3fd8219a2cf32a00d458f61802d17f01c5bcde5a4f82008ee4a7c8e9a06",
		signature,
	)
	if err == nil {
		t.Fail()
	}

	// changing the publisher key causes the signature to be invalid
	err = VerifyStarkPublisherPrice(
		1708940577123456789,
		"0x44594458555344000000000000000000637a6f7778",
		"3335950349880000000",
		"0x419d3fd8219a2cf32a00d458f61802d17f01c5bcde5a4f82008ee4a7c8e9a06",
		signature,
	)
	if err == nil {
		t.Fail()
	}

}

func TestStarkVerifier_VerifyPublisherPriceLongAssetId(t *testing.T) {
	signature := StarkSignature{
		R: "0x518f9a20f62381dc341e83e8715d36dfb0f7e1f3cf8efd2231f3b1a6b843685",
		S: "0x434cfdd6adfe376c86a5a28320212be79c04c36f3d7fe432db53b215a07cef4",
	}
	pubKey := PublisherKey("0x2798bbe74d340f938e8151b4af9992481dbb952ed359e2c46cf23021d6befd8")

	err := VerifyStarkPublisherPrice(
		1729023715673877869,
		"444a5457494e594553555344545741503438307073727631",
		"574709288691000000",
		pubKey,
		signature,
	)
	if err != nil {
		t.Error(err)
	}
}
