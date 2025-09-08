package publisher_agent

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/Stork-Oracle/stork-external/shared/signer"
	"github.com/stretchr/testify/assert"
)

func TestLoadConfig(t *testing.T) {
	config, secrets, err := LoadConfig("./resources/push_config.json", "./resources/example_keys.json")
	assert.NoError(t, err)
	assert.NotNil(t, config)

	assert.Equal(t, []signer.SignatureType{"evm", "stark"}, config.SignatureTypes)
	assert.Equal(
		t,
		signer.EvmPrivateKey("0x8b558d5fc31eb64bb51d44b4b28658180e96764d5d5ac68e6d124f86f576d9de"),
		secrets.EvmPrivateKey,
	)
	assert.Equal(t, signer.EvmPublisherKey("0x99e295e85cb07c16b7bb62a44df532a7f2620237"), config.EvmPublicKey)
	assert.Equal(
		t,
		signer.StarkPrivateKey("0x66253bdeb3c1a235cf4376611e3a14474e2c00fd2fb225f9a388faae7fb095a"),
		secrets.StarkPrivateKey,
	)
	assert.Equal(
		t,
		signer.StarkPublisherKey("0x418d3fd8219a2cf32a00d458f61802d17f01c5bcde5a4f82008ee4a7c8e9a06"),
		config.StarkPublicKey,
	)

	assert.Equal(t, "", config.PullBasedWsUrl)
	assert.Equal(t, 5216, config.IncomingWsPort)
}

func TestLoadKeys(t *testing.T) {
	// load keys from file
	keysFileData, err := readFile("./resources/example_keys.json")
	assert.NoError(t, err)
	assert.NotNil(t, keysFileData)

	var keys Keys
	err = json.Unmarshal(keysFileData, &keys)
	assert.NoError(t, err)
	assert.Equal(
		t,
		signer.EvmPrivateKey("0x8b558d5fc31eb64bb51d44b4b28658180e96764d5d5ac68e6d124f86f576d9de"),
		keys.EvmPrivateKey,
	)
	assert.Equal(t, signer.EvmPublisherKey("0x99e295e85cb07c16b7bb62a44df532a7f2620237"), keys.EvmPublicKey)
	assert.Equal(
		t,
		signer.StarkPrivateKey("0x66253bdeb3c1a235cf4376611e3a14474e2c00fd2fb225f9a388faae7fb095a"),
		keys.StarkPrivateKey,
	)
	assert.Equal(
		t,
		signer.StarkPublisherKey("0x418d3fd8219a2cf32a00d458f61802d17f01c5bcde5a4f82008ee4a7c8e9a06"),
		keys.StarkPublicKey,
	)

	envEvmPrivateKey := "0x8b558d5fc31eb64bb51d44b4b28658180e96764d5d5ac68e6d124f86f576d9d4"
	envEvmPublicKey := "0x99e295e85cb07c16b7bb62a44df532a7f2620234"
	envStarkPrivateKey := "0x66253bdeb3c1a235cf4376611e3a14474e2c00fd2fb225f9a388faae7fb0954"
	envStarkPublicKey := "0x418d3fd8219a2cf32a00d458f61802d17f01c5bcde5a4f82008ee4a7c8e9a04"

	os.Setenv("STORK_EVM_PRIVATE_KEY", envEvmPrivateKey)
	os.Setenv("STORK_EVM_PUBLIC_KEY", envEvmPublicKey)
	os.Setenv("STORK_STARK_PRIVATE_KEY", envStarkPrivateKey)
	os.Setenv("STORK_STARK_PUBLIC_KEY", envStarkPublicKey)

	// ensure env keys aren't the same as file keys
	assert.NotEqual(t, keys.EvmPrivateKey, signer.EvmPrivateKey(envEvmPrivateKey))
	assert.NotEqual(t, keys.EvmPublicKey, signer.EvmPrivateKey(envEvmPublicKey))
	assert.NotEqual(t, keys.StarkPrivateKey, signer.EvmPrivateKey(envStarkPrivateKey))
	assert.NotEqual(t, keys.StarkPublicKey, signer.EvmPrivateKey(envStarkPublicKey))

	keys.updateFromEnvVars()

	// ensure all keys were overwritten
	assert.Equal(t, signer.EvmPrivateKey(envEvmPrivateKey), keys.EvmPrivateKey)
	assert.Equal(t, signer.EvmPublisherKey(envEvmPublicKey), keys.EvmPublicKey)
	assert.Equal(t, signer.StarkPrivateKey(envStarkPrivateKey), keys.StarkPrivateKey)
	assert.Equal(t, signer.StarkPublisherKey(envStarkPublicKey), keys.StarkPublicKey)
}
