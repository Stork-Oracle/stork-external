package publisher_agent

import (
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/Stork-Oracle/stork-external/apps/lib/signer"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

const evmPrivateKey = "0x8b558d5fc31eb64bb51d44b4b28658180e96764d5d5ac68e6d124f86f576d9de"
const evmPublicKey = "0x99e295e85cb07c16b7bb62a44df532a7f2620237"
const starkPublicKey = "0x66253bdeb3c1a235cf4376611e3a14474e2c00fd2fb225f9a388faae7fb095a"
const starkPrivateKey = "0x418d3fd8219a2cf32a00d458f61802d17f01c5bcde5a4f82008ee4a7c8e9a06"
const storkAuth = "fake_auth"

const assetId = "fakeAsset"

func getNextSignedOutput[T signer.Signature](ch chan SignedPriceUpdateBatch[T], timeout time.Duration) (SignedPriceUpdateBatch[T], bool) {
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
		[]signer.SignatureType{EvmSignatureType},
		evmPrivateKey,
		evmPublicKey,
		starkPrivateKey,
		starkPublicKey,
		time.Duration(0),
		10*time.Millisecond,
		DefaultChangeThresholdPercent,
		"czowx",
		DefaultStorkRegistryBaseUrl,
		time.Duration(0),
		time.Duration(0),
		storkAuth,
		"",
		"",
		"",
		time.Duration(0),
		time.Duration(0),
		false,
		0,
	)

	evmSigner, err := signer.NewEvmSigner(evmPrivateKey, logger)
	if err != nil {
		t.Fatalf("NewSigner[*signer.EvmSignature]: %v", err)
	}

	inputCh := make(chan ValueUpdate)
	outputCh := make(chan SignedPriceUpdateBatch[*signer.EvmSignature])
	processor := NewPriceUpdateProcessor[*signer.EvmSignature](
		evmSigner,
		config.OracleId,
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
		PublishTimestamp: 10000000,
		Asset:            assetId,
		Value:            big.NewFloat(1.0),
	}
	inputCh <- initialUpdate
	firstResult, success := getNextSignedOutput(outputCh, time.Second)
	if !success {
		t.Fatalf("getNextSignedOutput timed out")
	}
	assert.Equal(t, int64(10000000), firstResult[assetId].SignedPrice.TimestampedSignature.Timestamp)
	assert.Equal(t, QuantizedPrice("1000000000000000000"), firstResult[assetId].SignedPrice.QuantizedPrice)

	// subsequent updates with no change don't get sent out
	noDeltaUpdate := ValueUpdate{
		PublishTimestamp: 20000000,
		Asset:            assetId,
		Value:            big.NewFloat(1.0),
	}
	inputCh <- noDeltaUpdate
	_, success = getNextSignedOutput(outputCh, time.Second)
	if success {
		t.Fatalf("getNextSignedOutput should have timed out")
	}

	// updates with a large change get sent out
	changeUpdate := ValueUpdate{
		PublishTimestamp: 30000000,
		Asset:            assetId,
		Value:            big.NewFloat(2.0),
	}
	inputCh <- changeUpdate
	nextResult, success := getNextSignedOutput(outputCh, time.Second)
	if !success {
		t.Fatalf("getNextSignedOutput timed out")
	}
	assert.Equal(t, int64(30000000), nextResult[assetId].SignedPrice.TimestampedSignature.Timestamp)
	assert.Equal(t, QuantizedPrice("2000000000000000000"), nextResult[assetId].SignedPrice.QuantizedPrice)
}

func TestZeroPrice(t *testing.T) {
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	config := NewStorkPublisherAgentConfig(
		[]signer.SignatureType{EvmSignatureType},
		evmPrivateKey,
		evmPublicKey,
		starkPrivateKey,
		starkPublicKey,
		time.Duration(0),
		10*time.Millisecond,
		DefaultChangeThresholdPercent,
		"czowx",
		DefaultStorkRegistryBaseUrl,
		time.Duration(0),
		time.Duration(0),
		storkAuth,
		"",
		"",
		"",
		time.Duration(0),
		time.Duration(0),
		false,
		0,
	)

	evmSigner, err := signer.NewEvmSigner(evmPrivateKey, logger)
	if err != nil {
		t.Fatalf("NewSigner[*signer.EvmSignature]: %v", err)
	}

	inputCh := make(chan ValueUpdate)
	outputCh := make(chan SignedPriceUpdateBatch[*signer.EvmSignature])
	processor := NewPriceUpdateProcessor[*signer.EvmSignature](
		evmSigner,
		config.OracleId,
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
		PublishTimestamp: 10000000,
		Asset:            assetId,
		Value:            big.NewFloat(0.0),
	}

	inputCh <- zeroUpdate
	firstResult, success := getNextSignedOutput(outputCh, time.Second)
	if !success {
		t.Fatalf("getNextSignedOutput timed out")
	}
	assert.Equal(t, int64(10000000), firstResult[assetId].SignedPrice.TimestampedSignature.Timestamp)
	assert.Equal(t, QuantizedPrice("0"), firstResult[assetId].SignedPrice.QuantizedPrice)

	// subsequent zero updates have no delta, so nothing is sent out
	subsequentZeroUpdate := ValueUpdate{
		PublishTimestamp: 20000000,
		Asset:            assetId,
		Value:            big.NewFloat(0.0),
	}

	inputCh <- subsequentZeroUpdate
	_, success = getNextSignedOutput(outputCh, time.Second)
	if success {
		t.Fatalf("getNextSignedOutput should have timed out")
	}

	// any nonzero value has infinite delta, so gets sent out
	nonzeroUpdate := ValueUpdate{
		PublishTimestamp: 30000000,
		Asset:            assetId,
		Value:            big.NewFloat(1.0),
	}
	inputCh <- nonzeroUpdate
	nextResult, success := getNextSignedOutput(outputCh, time.Second)
	if !success {
		t.Fatalf("getNextSignedOutput timed out")
	}
	assert.Equal(t, int64(30000000), nextResult[assetId].SignedPrice.TimestampedSignature.Timestamp)
	assert.Equal(t, QuantizedPrice("1000000000000000000"), nextResult[assetId].SignedPrice.QuantizedPrice)

	// updating price back to zero gets sent out
	returnToZeroUpdate := ValueUpdate{
		PublishTimestamp: 40000000,
		Asset:            assetId,
		Value:            big.NewFloat(0.0),
	}

	inputCh <- returnToZeroUpdate
	returnToZeroResult, success := getNextSignedOutput(outputCh, time.Second)
	if !success {
		t.Fatalf("getNextSignedOutput timed out")
	}
	assert.Equal(t, int64(40000000), returnToZeroResult[assetId].SignedPrice.TimestampedSignature.Timestamp)
	assert.Equal(t, QuantizedPrice("0"), returnToZeroResult[assetId].SignedPrice.QuantizedPrice)
}
