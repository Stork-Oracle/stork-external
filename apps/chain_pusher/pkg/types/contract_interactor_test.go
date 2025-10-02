package types

import (
	"bytes"
	"errors"
	"fmt"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestRunWithFallback(t *testing.T) {
	badRpc := "https://badrpc.com/"
	goodRpc := "https://goodrpc.com/"

	interactor := MockContractInteractor{}
	interactor.On("ConnectHTTP", goodRpc).Return(nil)
	interactor.On("ConnectHTTP", badRpc).Return(errors.New("bad rpc"))

	var buf bytes.Buffer
	logger := zerolog.New(&buf)

	// one rpc url where connections and functions succeed
	fallbackContractInteractor := NewFallbackContractInteractor(
		&interactor,
		[]string{
			goodRpc,
		},
		[]string{},
		logger,
	)

	result, err := fallbackContractInteractor.runWithFallback(
		"successfulFunc",
		func() (any, error) {
			return 0, nil
		},
	)
	assert.Equal(t, 0, result)
	assert.NoError(t, err)
	assert.Equal(t, "", buf.String())

	// one rpc url where connections succeed but functions fail
	result, err = fallbackContractInteractor.runWithFallback(
		"failedFunc",
		func() (any, error) {
			return nil, fmt.Errorf("rpc down")
		},
	)
	assert.Nil(t, result)
	assert.ErrorContains(t, err, "failed with all supplied rpc urls. Last error: rpc down")
	assert.Contains(t, buf.String(), "error calling contract function on primary http rpc url, will attempt to fallback")

	// one rpc where connections fail
	buf.Reset()
	fallbackContractInteractor = NewFallbackContractInteractor(
		&interactor,
		[]string{
			badRpc,
		},
		[]string{},
		logger,
	)
	result, err = fallbackContractInteractor.runWithFallback(
		"failedFunc",
		func() (any, error) {
			return nil, fmt.Errorf("rpc down")
		},
	)
	assert.Nil(t, result)
	assert.ErrorContains(t, err, "failed with all supplied rpc urls. Last error")
	assert.Contains(t, buf.String(), "error connecting to primary rpc http client, will attempt to fallback")

	// no fallback if first rpc connect + functions succeed
	buf.Reset()
	fallbackContractInteractor = NewFallbackContractInteractor(
		&interactor,
		[]string{
			goodRpc,
			badRpc,
		},
		[]string{},
		logger,
	)
	result, err = fallbackContractInteractor.runWithFallback(
		"successfulFunc",
		func() (any, error) {
			return 0, nil
		},
	)
	assert.Equal(t, 0, result)
	assert.NoError(t, err)
	assert.Equal(t, "", buf.String())

	// fallback if first rpc connect fails
	buf.Reset()
	fallbackContractInteractor = NewFallbackContractInteractor(
		&interactor,
		[]string{
			badRpc,
			goodRpc,
		},
		[]string{},
		logger,
	)
	result, err = fallbackContractInteractor.runWithFallback(
		"successfulFunc",
		func() (any, error) {
			return 0, nil
		},
	)
	assert.Equal(t, 0, result)
	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "error connecting to primary rpc http client, will attempt to fallback")
	assert.Contains(t, buf.String(), "successfully connected to fallback http rpc url")
	assert.Contains(t, buf.String(), "successfully called contract function on fallback http rpc url")

	// fallback if first rpc connect succeeds but first rpc function fails
	buf.Reset()
	fallbackContractInteractor = NewFallbackContractInteractor(
		&interactor,
		[]string{
			goodRpc,
			goodRpc,
		},
		[]string{},
		logger,
	)
	isFirstCall := true
	result, err = fallbackContractInteractor.runWithFallback(
		"failOnFirstCallFunc",
		func() (any, error) {
			if isFirstCall {
				isFirstCall = false
				return nil, errors.New("rpc down")
			}
			return 0, nil
		},
	)
	assert.Equal(t, 0, result)
	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "error calling contract function on primary http rpc url, will attempt to fallback")
	assert.Contains(t, buf.String(), "successfully connected to fallback http rpc url")
	assert.Contains(t, buf.String(), "successfully called contract function on fallback http rpc url")
}
