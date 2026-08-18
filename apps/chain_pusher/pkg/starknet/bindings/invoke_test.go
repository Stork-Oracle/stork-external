package bindings

import (
	"context"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/NethermindEth/juno/core/felt"
	"github.com/NethermindEth/starknet.go/rpc"
	"github.com/NethermindEth/starknet.go/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testNonce is an arbitrary nonce; these tests never touch a chain.
func testNonce(t *testing.T) *felt.Felt {
	t.Helper()

	nonce, err := utils.HexToFelt("0x7")
	require.NoError(t, err)

	return nonce
}

// newTestContract builds a contract pointed at a stub node, skipping ConnectHTTP so the tests do
// not need a live chain.
func newTestContract(t *testing.T, url string, tip string) *StorkContract {
	t.Helper()

	contract, err := NewStorkContract("0x1234", "0x5678", big.NewInt(1), tip)
	require.NoError(t, err)

	contract.rpcURL = url
	contract.httpClient = &http.Client{}

	return contract
}

// captureRequest stands up a node that records the request body and replies with `reply`.
func captureRequest(t *testing.T, reply string, captured *map[string]any) *httptest.Server {
	t.Helper()

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any

		err := json.NewDecoder(r.Body).Decode(&body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)

			return
		}

		*captured = body

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(reply))
	}))
}

// The whole reason invoke.go exists: the library serialises optional fields the pusher never sets,
// and nodes reject the result. Asserting the exact field set catches any extra field reappearing,
// not just the ones that broke it originally.
func TestInvokePayloadCarriesExactlyTheExpectedFields(t *testing.T) {
	t.Parallel()

	contract := newTestContract(t, "http://unused", "")

	raw, err := json.Marshal(contract.newPayload(nil, testNonce(t), nil, transactionVersionV3))
	require.NoError(t, err)

	var fields map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(raw, &fields))

	// BROADCASTED_INVOKE_TXN_V3 from the JSON-RPC spec, and nothing else.
	expected := []string{
		"type", "sender_address", "calldata", "version", "signature", "nonce",
		"resource_bounds", "tip", "paymaster_data", "account_deployment_data",
		"nonce_data_availability_mode", "fee_data_availability_mode",
	}

	got := make([]string, 0, len(fields))
	for name := range fields {
		got = append(got, name)
	}

	assert.ElementsMatch(t, expected, got, "the outgoing payload has an unexpected field set")
}

func TestResourceBoundsCarryAllThreeResources(t *testing.T) {
	t.Parallel()

	contract := newTestContract(t, "http://unused", "")

	raw, err := json.Marshal(contract.newPayload(nil, testNonce(t), nil, transactionVersionV3))
	require.NoError(t, err)

	var payload struct {
		ResourceBounds map[string]struct {
			MaxAmount       string `json:"max_amount"`
			MaxPricePerUnit string `json:"max_price_per_unit"`
		} `json:"resource_bounds"`
	}
	require.NoError(t, json.Unmarshal(raw, &payload))

	for _, name := range []string{"l1_gas", "l1_data_gas", "l2_gas"} {
		bound, ok := payload.ResourceBounds[name]
		require.True(t, ok, "resource_bounds is missing %q", name)
		assert.NotEmpty(t, bound.MaxAmount)
		assert.NotEmpty(t, bound.MaxPricePerUnit)
	}
}

func TestTipDefaultsToZeroAndIsOverridable(t *testing.T) {
	t.Parallel()

	assert.Equal(t, rpc.U64("0x0"), newTestContract(t, "http://unused", "").tipOrZero())
	assert.Equal(t, rpc.U64("0x5f5e100"), newTestContract(t, "http://unused", "0x5f5e100").tipOrZero())
}

func TestCallSurfacesNodeErrors(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(
			`{"jsonrpc":"2.0","id":1,"error":{"code":-32602,"message":"Invalid params","data":{"reason":"nope"}}}`,
		))
	}))
	defer server.Close()

	contract := newTestContract(t, server.URL, "")

	err := contract.call(context.Background(), "starknet_test", []any{}, nil)

	require.Error(t, err)
	// The node's reason has to survive: these are the errors that made the wire mismatch
	// diagnosable in the first place.
	assert.Contains(t, err.Error(), "Invalid params")
	assert.Contains(t, err.Error(), "nope")
	assert.Contains(t, err.Error(), "-32602")
}

func TestCallRejectsMalformedResponses(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`this is not json`))
	}))
	defer server.Close()

	contract := newTestContract(t, server.URL, "")

	err := contract.call(context.Background(), "starknet_test", []any{}, nil)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode")
}

func TestCallRespectsContextCancellation(t *testing.T) {
	t.Parallel()

	contract := newTestContract(t, "http://127.0.0.1:1", "")

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := contract.call(ctx, "starknet_test", []any{}, nil)
	require.Error(t, err)
}

func TestEstimateRejectsAnEmptyResult(t *testing.T) {
	t.Parallel()

	var captured map[string]any

	server := captureRequest(t, `{"jsonrpc":"2.0","id":1,"result":[]}`, &captured)
	defer server.Close()

	contract := newTestContract(t, server.URL, "")

	_, err := contract.estimateResourceBounds(context.Background(), nil, testNonce(t))
	require.ErrorIs(t, err, ErrEstimateFailed)
}

// The estimate runs before a signature exists, so it must use the query-bit version and ask the
// node to skip validation. Getting this wrong makes estimation fail with a signature error.
func TestEstimateUsesQueryVersionAndSkipsValidate(t *testing.T) {
	t.Parallel()

	var captured map[string]any

	server := captureRequest(t, `{"jsonrpc":"2.0","id":1,"result":[]}`, &captured)
	defer server.Close()

	contract := newTestContract(t, server.URL, "")
	_, _ = contract.estimateResourceBounds(context.Background(), nil, testNonce(t))

	require.Equal(t, "starknet_estimateFee", captured["method"])

	params, ok := captured["params"].([]any)
	require.True(t, ok)
	require.Len(t, params, 3)

	txns, ok := params[0].([]any)
	require.True(t, ok)
	require.Len(t, txns, 1)

	txn, ok := txns[0].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, transactionVersionV3Query, txn["version"])

	flags, ok := params[1].([]any)
	require.True(t, ok)
	assert.Equal(t, []any{"SKIP_VALIDATE"}, flags)
}

func TestAddInvokeSendsTheTransactionUnwrapped(t *testing.T) {
	t.Parallel()

	var captured map[string]any

	server := captureRequest(t, `{"jsonrpc":"2.0","id":1,"result":{"transaction_hash":"0xdead"}}`, &captured)
	defer server.Close()

	contract := newTestContract(t, server.URL, "")

	var result struct {
		TransactionHash string `json:"transaction_hash"`
	}

	err := contract.call(context.Background(), "starknet_addInvokeTransaction",
		[]any{contract.newPayload(nil, testNonce(t), nil, transactionVersionV3)}, &result)
	require.NoError(t, err)

	assert.Equal(t, "0xdead", result.TransactionHash)

	params, ok := captured["params"].([]any)
	require.True(t, ok)
	require.Len(t, params, 1)

	// The node wants the transaction itself, not nested under an "invoke_transaction" key.
	txn, ok := params[0].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "INVOKE", txn["type"])
	assert.NotContains(t, txn, "invoke_transaction")
}

func TestJsonRpcErrorFormatting(t *testing.T) {
	t.Parallel()

	withData := &jsonRpcError{Code: -32602, Message: "Invalid params", Data: json.RawMessage(`{"reason":"x"}`)}
	assert.Equal(t, `-32602 Invalid params: {"reason":"x"}`, withData.Error())

	withoutData := &jsonRpcError{Code: 40, Message: "Contract error"}
	assert.Equal(t, "40 Contract error", withoutData.Error())

	var target *jsonRpcError
	assert.ErrorAs(t, error(withData), &target)
}

func TestFeltsToHexHandlesEmpty(t *testing.T) {
	t.Parallel()

	assert.Empty(t, feltsToHex(nil))
	assert.NotNil(t, feltsToHex(nil), "must marshal as [] rather than null")
}
