package stork_publisher_agent

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfig(t *testing.T) {
	config, err := LoadConfig("./resources/push_config.json", "./resources/example_keys.json")
	assert.NoError(t, err)
	assert.NotNil(t, config)

	assert.Equal(t, []SignatureType{"evm", "stark"}, config.SignatureTypes)
	assert.Equal(t, EvmPrivateKey("0x8b558d5fc31eb64bb51d44b4b28658180e96764d5d5ac68e6d124f86f576d9de"), config.EvmPrivateKey)
	assert.Equal(t, EvmPublisherKey("0x99e295e85cb07c16b7bb62a44df532a7f2620237"), config.EvmPublicKey)
	assert.Equal(t, StarkPrivateKey("0x66253bdeb3c1a235cf4376611e3a14474e2c00fd2fb225f9a388faae7fb095a"), config.StarkPrivateKey)
	assert.Equal(t, StarkPublisherKey("0x418d3fd8219a2cf32a00d458f61802d17f01c5bcde5a4f82008ee4a7c8e9a06"), config.StarkPublicKey)

	assert.Equal(t, "", config.PullBasedWsUrl)
	assert.Equal(t, 5216, config.IncomingWsPort)
}
