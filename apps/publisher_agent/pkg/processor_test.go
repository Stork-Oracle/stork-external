package publisher_agent

import (
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/Stork-Oracle/stork-external/shared"
	"github.com/Stork-Oracle/stork-external/shared/signer"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

const (
	evmPrivateKey   = "0x8b558d5fc31eb64bb51d44b4b28658180e96764d5d5ac68e6d124f86f576d9de"
	evmPublicKey    = "0x99e295e85cb07c16b7bb62a44df532a7f2620237"
	starkPublicKey  = "0x66253bdeb3c1a235cf4376611e3a14474e2c00fd2fb225f9a388faae7fb095a"
	starkPrivateKey = "0x418d3fd8219a2cf32a00d458f61802d17f01c5bcde5a4f82008ee4a7c8e9a06"
	storkAuth       = "fake_auth"
)

const assetID = "fakeAsset"

func getNextSignedOutput[T shared.Signature](
	ch chan SignedPriceUpdateBatch[T],
	timeout time.Duration,
) (SignedPriceUpdateBatch[T], bool) {
	select {
	case value := <-ch:
		return value, true
	case <-time.After(timeout):
		return nil, false
	}
}

func TestDeltaOnly(t *testing.T) {
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	config := NewStorkPublisherAgentConfig(
		[]shared.SignatureType{shared.EvmSignatureType},
		evmPublicKey,
		starkPublicKey,
		time.Duration(0),
		10*time.Millisecond,
		DefaultChangeThresholdPercent,
		"czowx",
		DefaultStorkRegistryBaseUrl,
		time.Duration(0),
		time.Duration(0),
		"",
		time.Duration(0),
		"",
		"",
		time.Duration(0),
		time.Duration(0),
		false,
		0,
		[]BrokerConnectionConfig{},
	)

	evmSigner, err := signer.NewEvmSigner(evmPrivateKey, logger)
	if err != nil {
		t.Fatalf("NewSigner[*signer.EvmSignature]: %v", err)
	}

	inputCh := make(chan ValueUpdate)
	outputCh := make(chan SignedPriceUpdateBatch[*shared.EvmSignature])
	processor := NewPriceUpdateProcessor[*shared.EvmSignature](
		evmSigner,
		config.OracleID,
		1,
		config.ClockPeriod,
		config.DeltaCheckPeriod,
		DefaultChangeThresholdPercent,
		false,
		inputCh,
		outputCh,
		logger,
	)

	go processor.Run()

	// initial update gets sent out
	initialUpdate := ValueUpdate{
		PublishTimestampNano: 10000000,
		Asset:                assetID,
		Value:                big.NewFloat(1.0),
	}
	inputCh <- initialUpdate
	firstResult, success := getNextSignedOutput(outputCh, time.Second)
	if !success {
		t.Fatalf("getNextSignedOutput timed out")
	}
	assert.Equal(t, uint64(10000000), firstResult[assetID].SignedPrice.TimestampedSignature.TimestampNano)
	assert.Equal(t, shared.QuantizedPrice("1000000000000000000"), firstResult[assetID].SignedPrice.QuantizedPrice)

	// subsequent updates with no change don't get sent out
	noDeltaUpdate := ValueUpdate{
		PublishTimestampNano: 20000000,
		Asset:                assetID,
		Value:                big.NewFloat(1.0),
	}
	inputCh <- noDeltaUpdate
	_, success = getNextSignedOutput(outputCh, time.Second)
	if success {
		t.Fatalf("getNextSignedOutput should have timed out")
	}

	// updates with a large change get sent out
	changeUpdate := ValueUpdate{
		PublishTimestampNano: 30000000,
		Asset:                assetID,
		Value:                big.NewFloat(2.0),
	}
	inputCh <- changeUpdate
	nextResult, success := getNextSignedOutput(outputCh, time.Second)
	if !success {
		t.Fatalf("getNextSignedOutput timed out")
	}
	assert.Equal(t, uint64(30000000), nextResult[assetID].SignedPrice.TimestampedSignature.TimestampNano)
	assert.Equal(t, shared.QuantizedPrice("2000000000000000000"), nextResult[assetID].SignedPrice.QuantizedPrice)
}

func TestZeroPrice(t *testing.T) {
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	config := NewStorkPublisherAgentConfig(
		[]shared.SignatureType{shared.EvmSignatureType},
		evmPublicKey,
		starkPublicKey,
		time.Duration(0),
		10*time.Millisecond,
		DefaultChangeThresholdPercent,
		"czowx",
		DefaultStorkRegistryBaseUrl,
		time.Duration(0),
		time.Duration(0),
		"",
		time.Duration(0),
		"",
		"",
		time.Duration(0),
		time.Duration(0),
		false,
		0,
		[]BrokerConnectionConfig{},
	)

	evmSigner, err := signer.NewEvmSigner(evmPrivateKey, logger)
	if err != nil {
		t.Fatalf("NewSigner[*signer.EvmSignature]: %v", err)
	}

	inputCh := make(chan ValueUpdate)
	outputCh := make(chan SignedPriceUpdateBatch[*shared.EvmSignature])
	processor := NewPriceUpdateProcessor[*shared.EvmSignature](
		evmSigner,
		config.OracleID,
		1,
		config.ClockPeriod,
		config.DeltaCheckPeriod,
		DefaultChangeThresholdPercent,
		false,
		inputCh,
		outputCh,
		logger,
	)

	go processor.Run()

	// initial zero update gets sent out
	zeroUpdate := ValueUpdate{
		PublishTimestampNano: 10000000,
		Asset:                assetID,
		Value:                big.NewFloat(0.0),
	}

	inputCh <- zeroUpdate
	firstResult, success := getNextSignedOutput(outputCh, time.Second)
	if !success {
		t.Fatalf("getNextSignedOutput timed out")
	}
	assert.Equal(t, uint64(10000000), firstResult[assetID].SignedPrice.TimestampedSignature.TimestampNano)
	assert.Equal(t, shared.QuantizedPrice("0"), firstResult[assetID].SignedPrice.QuantizedPrice)

	// subsequent zero updates have no delta, so nothing is sent out
	subsequentZeroUpdate := ValueUpdate{
		PublishTimestampNano: 20000000,
		Asset:                assetID,
		Value:                big.NewFloat(0.0),
	}

	inputCh <- subsequentZeroUpdate
	_, success = getNextSignedOutput(outputCh, time.Second)
	if success {
		t.Fatalf("getNextSignedOutput should have timed out")
	}

	// any nonzero value has infinite delta, so gets sent out
	nonzeroUpdate := ValueUpdate{
		PublishTimestampNano: 30000000,
		Asset:                assetID,
		Value:                big.NewFloat(1.0),
	}
	inputCh <- nonzeroUpdate
	nextResult, success := getNextSignedOutput(outputCh, time.Second)
	if !success {
		t.Fatalf("getNextSignedOutput timed out")
	}
	assert.Equal(t, uint64(30000000), nextResult[assetID].SignedPrice.TimestampedSignature.TimestampNano)
	assert.Equal(t, shared.QuantizedPrice("1000000000000000000"), nextResult[assetID].SignedPrice.QuantizedPrice)

	// updating price back to zero gets sent out
	returnToZeroUpdate := ValueUpdate{
		PublishTimestampNano: 40000000,
		Asset:                assetID,
		Value:                big.NewFloat(0.0),
	}

	inputCh <- returnToZeroUpdate
	returnToZeroResult, success := getNextSignedOutput(outputCh, time.Second)
	if !success {
		t.Fatalf("getNextSignedOutput timed out")
	}
	assert.Equal(t, uint64(40000000), returnToZeroResult[assetID].SignedPrice.TimestampedSignature.TimestampNano)
	assert.Equal(t, shared.QuantizedPrice("0"), returnToZeroResult[assetID].SignedPrice.QuantizedPrice)
}
