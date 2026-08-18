// Package bindings provides a hand written client for the Stork Cairo contract.
//
// Starknet has no ABI code generator comparable to abigen, so the Cairo serde layout is
// implemented directly here: every value is encoded as a flat list of field elements, with `u256`
// split into (low, high) limbs and `i128` embedded into the field as `P - |v|` when negative.
package bindings

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"time"

	"github.com/NethermindEth/juno/core/felt"
	"github.com/NethermindEth/starknet.go/account"
	"github.com/NethermindEth/starknet.go/rpc"
	"github.com/NethermindEth/starknet.go/utils"
)

var (
	ErrNotConnected           = errors.New("contract is not connected to an HTTP RPC")
	ErrWsNotConnected         = errors.New("contract is not connected to a websocket RPC")
	ErrUnexpectedResultLength = errors.New("unexpected number of field elements in call result")
	ErrInvalidSignatureV      = errors.New("invalid signature v value, expected 27 or 28")
	ErrNoUpdates              = errors.New("no updates provided")
)

const (
	// feltsPerValue is the serialized width of a `TemporalNumericValue` (timestamp_ns, quantized_value).
	feltsPerValue = 2

	sigV27 = 27
	sigV28 = 28

	// rpcRequestTimeout bounds the hand-rolled JSON-RPC calls in invoke.go.
	rpcRequestTimeout = 30 * time.Second
)

// EncodedAssetID is the 32 byte Stork asset identifier.
type EncodedAssetID [32]byte

// TemporalNumericValue is the decoded form of the contract's value struct.
type TemporalNumericValue struct {
	TimestampNs    uint64
	QuantizedValue *big.Int
}

// UpdateData is one signed update, matching `TemporalNumericValueInput` in the Cairo contract.
type UpdateData struct {
	ID                   EncodedAssetID
	TemporalNumericValue TemporalNumericValue
	PublisherMerkleRoot  [32]byte
	ValueComputeAlgHash  [32]byte
	R                    [32]byte
	S                    [32]byte
	V                    uint8
}

// ValueUpdateEvent is a decoded `ValueUpdate` event.
type ValueUpdateEvent struct {
	ID    EncodedAssetID
	Value TemporalNumericValue
}

// StorkContract wraps the RPC providers and the account used to sign update transactions.
type StorkContract struct {
	contractAddress *felt.Felt
	accountAddress  *felt.Felt
	privateKey      *big.Int

	provider   *rpc.Provider
	wsClient   *rpc.WsProvider
	account    *account.Account
	rpcURL     string
	httpClient *http.Client

	// tip, when set, is used verbatim instead of asking the node to estimate one. Nodes that do
	// not serve tip estimation fail the whole transaction otherwise, and on a busy network an
	// explicit tip is how an operator bids for faster inclusion.
	tip rpc.U64

	// SpecVersionWarning is set when the node advertises a JSON-RPC spec version this client
	// does not implement. Calls are still attempted; the caller decides how loudly to complain.
	SpecVersionWarning error
}

// NewStorkContract creates a contract handle. It performs no I/O; call ConnectHTTP and
// ConnectWs to attach providers.
func NewStorkContract(
	contractAddress, accountAddress string, privateKey *big.Int, tip string,
) (*StorkContract, error) {
	contractFelt, err := utils.HexToFelt(contractAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to parse contract address: %w", err)
	}

	accountFelt, err := utils.HexToFelt(accountAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to parse account address: %w", err)
	}

	return &StorkContract{
		contractAddress:    contractFelt,
		accountAddress:     accountFelt,
		privateKey:         privateKey,
		tip:                rpc.U64(tip),
		rpcURL:             "",
		httpClient:         &http.Client{Timeout: rpcRequestTimeout},
		provider:           nil,
		wsClient:           nil,
		account:            nil,
		SpecVersionWarning: nil,
	}, nil
}

// ConnectHTTP dials the JSON-RPC endpoint and derives the signing account.
//
// A node advertising a JSON-RPC spec version the client library does not implement is reported as
// a warning rather than an error: starknet.go still returns a usable provider in that case, and
// nodes routinely run ahead of the library.
func (c *StorkContract) ConnectHTTP(ctx context.Context, url string) error {
	provider, err := rpc.NewProvider(ctx, url)
	if err != nil {
		if !errors.Is(err, rpc.ErrIncompatibleVersion) || provider == nil {
			return fmt.Errorf("failed to create starknet provider: %w", err)
		}

		c.SpecVersionWarning = err
	}

	publicKey := publicKeyFromPrivateKey(c.privateKey)

	keystore := account.NewMemKeystore()
	keystore.Put(publicKey, c.privateKey)

	acc, err := account.NewAccount(provider, c.accountAddress, publicKey, keystore, account.CairoV2)
	if err != nil {
		return fmt.Errorf("failed to create starknet account: %w", err)
	}

	c.provider = provider
	c.account = acc
	c.rpcURL = url

	return nil
}

// ConnectWs dials the websocket endpoint used for event subscriptions.
func (c *StorkContract) ConnectWs(ctx context.Context, url string) error {
	wsClient, err := rpc.NewWebsocketProvider(ctx, url)
	if err != nil {
		return fmt.Errorf("failed to create starknet websocket provider: %w", err)
	}

	c.wsClient = wsClient

	return nil
}

// Close releases the websocket connection, if any.
func (c *StorkContract) Close() {
	if c.wsClient != nil {
		c.wsClient.Close()
	}
}

// AccountAddress returns the address updates are sent from.
func (c *StorkContract) AccountAddress() *felt.Felt {
	return c.accountAddress
}

// GetMultipleTemporalNumericValuesUnchecked reads the latest value for each id in a single call.
//
// The contract returns a zeroed value rather than reverting for feeds it has never seen, so those
// are simply omitted from the returned map.
func (c *StorkContract) GetMultipleTemporalNumericValuesUnchecked(
	ctx context.Context,
	ids []EncodedAssetID,
) (map[EncodedAssetID]TemporalNumericValue, error) {
	if c.provider == nil {
		return nil, ErrNotConnected
	}

	calldata := make([]*felt.Felt, 0, 1+len(ids)*2)
	calldata = append(calldata, utils.Uint64ToFelt(uint64(len(ids))))

	for _, id := range ids {
		calldata = append(calldata, encodeU256(id[:])...)
	}

	result, err := c.provider.Call(ctx, rpc.FunctionCall{
		ContractAddress:    c.contractAddress,
		EntryPointSelector: utils.GetSelectorFromNameFelt("get_multiple_temporal_numeric_values_unchecked"),
		Calldata:           calldata,
	}, rpc.WithBlockTag(rpc.BlockTagLatest))
	if err != nil {
		return nil, fmt.Errorf("failed to call get_multiple_temporal_numeric_values_unchecked: %w", err)
	}

	values, err := decodeValueArray(result, len(ids))
	if err != nil {
		return nil, err
	}

	polled := make(map[EncodedAssetID]TemporalNumericValue, len(values))

	for i, value := range values {
		// A zero timestamp means the contract has no value for this feed yet.
		if value.TimestampNs == 0 {
			continue
		}

		polled[ids[i]] = value
	}

	return polled, nil
}

// UpdateTemporalNumericValuesV1 submits a batch of signed updates and returns the transaction hash.
func (c *StorkContract) UpdateTemporalNumericValuesV1(
	ctx context.Context,
	updates []UpdateData,
) (string, error) {
	if c.account == nil {
		return "", ErrNotConnected
	}

	if len(updates) == 0 {
		return "", ErrNoUpdates
	}

	calldata, err := encodeUpdateData(updates)
	if err != nil {
		return "", err
	}

	invokeCalldata, err := c.account.FmtCalldata([]rpc.FunctionCall{{
		ContractAddress:    c.contractAddress,
		EntryPointSelector: utils.GetSelectorFromNameFelt("update_temporal_numeric_values_v1"),
		Calldata:           calldata,
	}})
	if err != nil {
		return "", fmt.Errorf("failed to format calldata: %w", err)
	}

	txHash, err := c.sendInvokeV3(ctx, invokeCalldata)
	if err != nil {
		return "", fmt.Errorf("failed to send update_temporal_numeric_values_v1 transaction: %w", err)
	}

	return txHash, nil
}

// SubscribeValueUpdates streams `ValueUpdate` events from the contract until ctx is cancelled.
//
// The returned error channel surfaces subscription failures so the caller can reconnect.
func (c *StorkContract) SubscribeValueUpdates(
	ctx context.Context,
	ch chan<- ValueUpdateEvent,
) (<-chan error, error) {
	if c.wsClient == nil {
		return nil, ErrWsNotConnected
	}

	events := make(chan *rpc.EmittedEventWithFinalityStatus, eventBufferSize)

	sub, err := c.wsClient.SubscribeEvents(ctx, events, &rpc.EventSubscriptionInput{
		FromAddress: rpc.AddressList{c.contractAddress},
		// Only the ValueUpdate selector, in the first key position.
		Keys: [][]*felt.Felt{{utils.GetSelectorFromNameFelt("ValueUpdate")}},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe to contract events: %w", err)
	}

	errCh := make(chan error, 1)

	go func() {
		defer sub.Unsubscribe()
		defer close(errCh)

		for {
			select {
			case <-ctx.Done():
				return
			case subErr := <-sub.Err():
				if subErr != nil {
					errCh <- fmt.Errorf("event subscription failed: %w", subErr)
				}

				return
			case event := <-events:
				update, decodeErr := decodeValueUpdateEvent(event)
				if decodeErr != nil {
					continue
				}

				select {
				case ch <- update:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return errCh, nil
}

// GetBalance returns the raw balance of the pushing account in the given ERC20 token.
func (c *StorkContract) GetBalance(ctx context.Context, tokenAddress string) (*big.Int, error) {
	if c.provider == nil {
		return nil, ErrNotConnected
	}

	tokenFelt, err := utils.HexToFelt(tokenAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to parse fee token address: %w", err)
	}

	result, err := c.provider.Call(ctx, rpc.FunctionCall{
		ContractAddress:    tokenFelt,
		EntryPointSelector: utils.GetSelectorFromNameFelt("balanceOf"),
		Calldata:           []*felt.Felt{c.accountAddress},
	}, rpc.WithBlockTag(rpc.BlockTagLatest))
	if err != nil {
		return nil, fmt.Errorf("failed to call balanceOf: %w", err)
	}

	//nolint:mnd // A u256 is always two field elements.
	if len(result) != 2 {
		return nil, fmt.Errorf("%w: got %d, want 2", ErrUnexpectedResultLength, len(result))
	}

	return decodeU256(result[0], result[1]), nil
}
