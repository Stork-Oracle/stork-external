package publisher_agent

import (
	"fmt"
	"time"

	"github.com/Stork-Oracle/stork_external/lib/signer"
)

const evmPrivateKey = "0x8b558d5fc31eb64bb51d44b4b28658180e96764d5d5ac68e6d124f86f576d9de"
const evmPublicKey = "0x99e295e85cb07c16b7bb62a44df532a7f2620237"
const starkPublicKey = "0x66253bdeb3c1a235cf4376611e3a14474e2c00fd2fb225f9a388faae7fb095a"
const starkPrivateKey = "0x418d3fd8219a2cf32a00d458f61802d17f01c5bcde5a4f82008ee4a7c8e9a06"
const storkAuth = "fakeauth"

const assetId = "fakeAsset"

const pushWsPort = 5216
const localRegistryPort = 8080
const brokerPort = 5200

func GetDeltaOnlyTestConfig() *StorkPublisherAgentConfig {
	return NewStorkPublisherAgentConfig(
		[]signer.SignatureType{EvmSignatureType},
		evmPrivateKey,
		evmPublicKey,
		starkPrivateKey,
		starkPublicKey,
		time.Duration(0),
		10*time.Millisecond,
		DefaultChangeThresholdPercent,
		"czowx",
		fmt.Sprintf("http://localhost:%v", localRegistryPort),
		time.Duration(0),
		time.Duration(0),
		storkAuth,
		"",
		"",
		"",
		time.Duration(0),
		time.Duration(0),
		false,
		pushWsPort,
	)
}
