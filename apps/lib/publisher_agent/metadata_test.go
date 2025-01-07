package publisher_agent

import (
	"encoding/json"
	"testing"

	"github.com/Stork-Oracle/stork-external/apps/lib/signer"
	"github.com/stretchr/testify/assert"
)

func TestGetRedactedConfigJson(t *testing.T) {
	myEvmPrivateKey := signer.EvmPrivateKey("my_evm_private_key")
	myEvmPublicKey := signer.EvmPublisherKey("my_evm_public_key")
	myStarkPrivateKey := signer.StarkPrivateKey("my_stark_private_key")
	myStarkPublicKey := signer.StarkPublisherKey("my_stark_public_key")
	myPullBasedAuth := AuthToken("my_pull_based_auth")

	config := StorkPublisherAgentConfig{
		EvmPrivateKey:   myEvmPrivateKey,
		EvmPublicKey:    myEvmPublicKey,
		StarkPrivateKey: myStarkPrivateKey,
		StarkPublicKey:  myStarkPublicKey,
		PullBasedAuth:   myPullBasedAuth,
	}

	redactedConfigJson := getRedactedConfigJson(config)

	// original config isn't mutatated
	assert.Equal(t, myEvmPrivateKey, config.EvmPrivateKey)
	assert.Equal(t, myStarkPrivateKey, config.StarkPrivateKey)
	assert.Equal(t, myPullBasedAuth, config.PullBasedAuth)

	// output json doesn't include secrets
	var redactedConfig StorkPublisherAgentConfig
	json.Unmarshal([]byte(redactedConfigJson), &redactedConfig)
	assert.Equal(t, signer.EvmPrivateKey(""), redactedConfig.EvmPrivateKey)
	assert.Equal(t, signer.StarkPrivateKey(""), redactedConfig.StarkPrivateKey)
	assert.Equal(t, AuthToken(""), redactedConfig.PullBasedAuth)

	// output json includes non-secret fields
	assert.Equal(t, myEvmPublicKey, redactedConfig.EvmPublicKey)
	assert.Equal(t, myStarkPublicKey, redactedConfig.StarkPublicKey)
}
