package bindings

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/NethermindEth/juno/core/felt"
	"github.com/NethermindEth/starknet.go/hash"
	"github.com/NethermindEth/starknet.go/rpc"
	"github.com/NethermindEth/starknet.go/utils"
)

// This file builds and submits the invoke transaction directly rather than going through
// starknet.go's account.BuildAndSendInvokeTxn.
//
// The library's BroadcastInvokeTxnV3 declares two optional fields without `omitempty`, so every
// transaction it builds carries them as nulls, and one of them is typed as an array where the spec
// calls for a base64 string. Nodes reject the result:
//
//	json: cannot unmarshal array into Go struct field BroadcastedTransaction.proof of type core.Base64
//
// The payloads are therefore serialized here instead, carrying only the fields a Stork update
// actually needs. Everything security sensitive - the transaction hash and the signature - still
// comes from starknet.go.

var ErrEstimateFailed = errors.New("fee estimation returned no result")

const (
	// transactionVersionV3 is the invoke version, and the query-bit variant used for estimation.
	transactionVersionV3      = "0x3"
	transactionVersionV3Query = "0x100000000000000000000000000000003"

	// feeMultiplier pads the estimate to absorb price movement between estimating and landing.
	feeMultiplier = 1.5

	// dataAvailabilityModeL1 is the only mode currently supported for fees and nonces.
	dataAvailabilityModeL1 = "L1"
)

// zeroBound is the placeholder used while estimating, before real bounds are known.
//
//nolint:gochecknoglobals // Effectively a constant; structs cannot be const.
var zeroBound = resourceBound{MaxAmount: "0x0", MaxPricePerUnit: "0x0"}

// resourceBound is one entry of a V3 transaction's resource_bounds.
type resourceBound struct {
	MaxAmount       string `json:"max_amount"`
	MaxPricePerUnit string `json:"max_price_per_unit"`
}

type resourceBoundsPayload struct {
	L1Gas     resourceBound `json:"l1_gas"`
	L1DataGas resourceBound `json:"l1_data_gas"`
	L2Gas     resourceBound `json:"l2_gas"`
}

// invokeV3Payload mirrors BROADCASTED_INVOKE_TXN_V3 from the JSON-RPC spec, carrying only the
// fields a Stork update needs. The optional fields the pusher never sets are left out entirely
// rather than sent as nulls.
//
//nolint:tagliatelle // Field names are fixed by the JSON-RPC spec.
type invokeV3Payload struct {
	Type                  string                `json:"type"`
	SenderAddress         string                `json:"sender_address"`
	Calldata              []string              `json:"calldata"`
	Version               string                `json:"version"`
	Signature             []string              `json:"signature"`
	Nonce                 string                `json:"nonce"`
	ResourceBounds        resourceBoundsPayload `json:"resource_bounds"`
	Tip                   string                `json:"tip"`
	PaymasterData         []string              `json:"paymaster_data"`
	AccountDeploymentData []string              `json:"account_deployment_data"`
	NonceDAMode           string                `json:"nonce_data_availability_mode"`
	FeeDAMode             string                `json:"fee_data_availability_mode"`
}

//nolint:tagliatelle // "jsonrpc" is fixed by the JSON-RPC 2.0 spec.
type jsonRpcRequest struct {
	JSONRpc string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  any    `json:"params"`
	ID      int    `json:"id"`
}

type jsonRpcError struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}

func (e *jsonRpcError) Error() string {
	if len(e.Data) > 0 {
		return fmt.Sprintf("%d %s: %s", e.Code, e.Message, string(e.Data))
	}

	return fmt.Sprintf("%d %s", e.Code, e.Message)
}

type jsonRpcResponse struct {
	Result json.RawMessage `json:"result"`
	Error  *jsonRpcError   `json:"error"`
}

// call performs a single JSON-RPC request against the node.
func (c *StorkContract) call(ctx context.Context, method string, params any, out any) error {
	body, err := json.Marshal(jsonRpcRequest{JSONRpc: "2.0", Method: method, Params: params, ID: 1})
	if err != nil {
		return fmt.Errorf("failed to encode %s request: %w", method, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.rpcURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to build %s request: %w", method, err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send %s request: %w", method, err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read %s response: %w", method, err)
	}

	var decoded jsonRpcResponse

	err = json.Unmarshal(raw, &decoded)
	if err != nil {
		return fmt.Errorf("failed to decode %s response %q: %w", method, string(raw), err)
	}

	if decoded.Error != nil {
		return fmt.Errorf("%s failed: %w", method, decoded.Error)
	}

	if out != nil {
		err = json.Unmarshal(decoded.Result, out)
		if err != nil {
			return fmt.Errorf("failed to decode %s result: %w", method, err)
		}
	}

	return nil
}

// sendInvokeV3 builds, signs and submits a V3 invoke transaction, returning its hash.
func (c *StorkContract) sendInvokeV3(ctx context.Context, calldata []*felt.Felt) (string, error) {
	nonce, err := c.account.Nonce(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to read account nonce: %w", err)
	}

	bounds, err := c.estimateResourceBounds(ctx, calldata, nonce)
	if err != nil {
		return "", err
	}

	txn := rpc.InvokeTxnV3{
		Type:                  rpc.TransactionTypeInvoke,
		SenderAddress:         c.accountAddress,
		Calldata:              calldata,
		Version:               rpc.TransactionV3,
		Signature:             []*felt.Felt{},
		Nonce:                 nonce,
		ResourceBounds:        bounds,
		Tip:                   c.tipOrZero(),
		PayMasterData:         []*felt.Felt{},
		AccountDeploymentData: []*felt.Felt{},
		NonceDataMode:         rpc.DAModeL1,
		FeeMode:               rpc.DAModeL1,
	}

	txnHash, err := hash.TransactionHashInvokeV3(&txn, c.account.ChainID)
	if err != nil {
		return "", fmt.Errorf("failed to compute transaction hash: %w", err)
	}

	signature, err := c.account.Sign(ctx, txnHash)
	if err != nil {
		return "", fmt.Errorf("failed to sign transaction: %w", err)
	}

	payload := c.newPayload(calldata, nonce, bounds, transactionVersionV3)
	payload.Signature = feltsToHex(signature)

	var result struct {
		TransactionHash string `json:"transaction_hash"`
	}

	err = c.call(ctx, "starknet_addInvokeTransaction", []any{payload}, &result)
	if err != nil {
		return "", err
	}

	return result.TransactionHash, nil
}

// estimateResourceBounds asks the node what the call will cost and pads the answer.
//
// The estimate is run with the query-bit version and SKIP_VALIDATE so it does not need a valid
// signature, which would otherwise have to be computed against bounds that are not known yet.
func (c *StorkContract) estimateResourceBounds(
	ctx context.Context, calldata []*felt.Felt, nonce *felt.Felt,
) (*rpc.ResourceBoundsMapping, error) {
	payload := c.newPayload(calldata, nonce, nil, transactionVersionV3Query)

	var estimates []rpc.FeeEstimation

	err := c.call(ctx, "starknet_estimateFee",
		[]any{[]any{payload}, []string{"SKIP_VALIDATE"}, "pre_confirmed"}, &estimates)
	if err != nil {
		return nil, err
	}

	if len(estimates) == 0 {
		return nil, ErrEstimateFailed
	}

	return utils.FeeEstToResBoundsMap(estimates[0], feeMultiplier), nil
}

// newPayload assembles the parts of the transaction that are the same for estimation and submission.
func (c *StorkContract) newPayload(
	calldata []*felt.Felt, nonce *felt.Felt, bounds *rpc.ResourceBoundsMapping, version string,
) invokeV3Payload {
	payload := invokeV3Payload{
		Type:                  "INVOKE",
		SenderAddress:         c.accountAddress.String(),
		Calldata:              feltsToHex(calldata),
		Version:               version,
		Signature:             []string{},
		Nonce:                 nonce.String(),
		ResourceBounds:        resourceBoundsPayload{L1Gas: zeroBound, L1DataGas: zeroBound, L2Gas: zeroBound},
		Tip:                   string(c.tipOrZero()),
		PaymasterData:         []string{},
		AccountDeploymentData: []string{},
		NonceDAMode:           dataAvailabilityModeL1,
		FeeDAMode:             dataAvailabilityModeL1,
	}

	if bounds != nil {
		payload.ResourceBounds = resourceBoundsPayload{
			L1Gas:     toResourceBound(bounds.L1Gas),
			L1DataGas: toResourceBound(bounds.L1DataGas),
			L2Gas:     toResourceBound(bounds.L2Gas),
		}
	}

	return payload
}

// tipOrZero defaults the tip to zero, which nodes accept when they are not congested.
func (c *StorkContract) tipOrZero() rpc.U64 {
	if c.tip == "" {
		return "0x0"
	}

	return c.tip
}

func toResourceBound(bound rpc.ResourceBounds) resourceBound {
	return resourceBound{
		MaxAmount:       string(bound.MaxAmount),
		MaxPricePerUnit: string(bound.MaxPricePerUnit),
	}
}

func feltsToHex(felts []*felt.Felt) []string {
	out := make([]string, 0, len(felts))
	for _, f := range felts {
		out = append(out, f.String())
	}

	return out
}
