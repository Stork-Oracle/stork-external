package contract_bindings_cosmwasm

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"

	cosmossdk_io_math "cosmossdk.io/math"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	http "github.com/cometbft/cometbft/rpc/client/http"
	sdkclient "github.com/cosmos/cosmos-sdk/client"
	sdkclient_tx "github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	keyring "github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

// Types
type Addr string

type Uint128 string

type Uint64 string

type Int128 string

type QueryMsg struct {
	GetLatestCanonicalTemporalNumericValueUnchecked *QueryMsg_GetLatestCanonicalTemporalNumericValueUnchecked `json:"get_latest_canonical_temporal_numeric_value_unchecked,omitempty"`
	GetSingleUpdateFee                              *QueryMsg_GetSingleUpdateFee                              `json:"get_single_update_fee,omitempty"`
}

type Coin struct {
	Amount Uint128 `json:"amount"`
	Denom  string  `json:"denom"`
}

func (c *Coin) toCosmosCoin() sdktypes.Coin {
	bi := new(big.Int)
	bi, ok := bi.SetString(string(c.Amount), 10)
	if !ok {
		panic("failed to convert Uint128 string to big.Int")
	}
	return sdktypes.NewCoin(c.Denom, cosmossdk_io_math.NewIntFromBigInt(bi))
}

// Response for the `get_single_update_fee` query containing the fee for a single update.
type GetSingleUpdateFeeResponse struct {
	Fee Coin `json:"fee"`
}

// A struct representing a timestamped value.
type TemporalNumericValue struct {
	// The quantized value.
	QuantizedValue Int128 `json:"quantized_value"`
	// The unix timestamp of the value in nanoseconds.
	TimestampNs Uint64 `json:"timestamp_ns"`
}

type ExecMsg struct {
	UpdateTemporalNumericValuesEvm *ExecMsg_UpdateTemporalNumericValuesEvm `json:"update_temporal_numeric_values_evm,omitempty"`
}

// Response for the `get_temporal_numeric_value` query containing the [`TemporalNumericValue`](./temporal_numeric_value.rs) for a given asset id.
type GetTemporalNumericValueResponse struct {
	TemporalNumericValue TemporalNumericValue `json:"temporal_numeric_value"`
}

type QueryAccountInfoRequest struct {
	Address string `json:"address"`
}

// The data structure for an update to a Stork feed. This is used in the `UpdateTemporalNumericValuesEvm` `ExecMsg" variant
type UpdateData struct {
	TemporalNumericValue TemporalNumericValue `json:"temporal_numeric_value"`
	V                    int                  `json:"v"`
	ValueComputeAlgHash  [32]int              `json:"value_compute_alg_hash"`
	Id                   [32]int              `json:"id"`
	PublisherMerkleRoot  [32]int              `json:"publisher_merkle_root"`
	R                    [32]int              `json:"r"`
	S                    [32]int              `json:"s"`
}

type QueryMsg_GetLatestCanonicalTemporalNumericValueUnchecked struct {
	Id [32]int `json:"id"`
}

type QueryMsg_GetSingleUpdateFee struct{}

type ExecMsg_UpdateTemporalNumericValuesEvm struct {
	UpdateData []UpdateData `json:"update_data"`
}

type StorkContract struct {
	ContractAddress string
	ChainPrefix     string
	singleUpdateFee Coin
	clientCtx       sdkclient.Context
	txf             sdkclient_tx.Factory
	marshaler       codec.Codec
}

func NewStorkContract(rpcUrl string, contractAddress string, mnemonic string, gasPrice float64, gasAdjustment float64, denom string, chainID string, chainPrefix string) (*StorkContract, error) {
	config := sdktypes.GetConfig()
	config.SetBech32PrefixForAccount(chainPrefix, chainPrefix+"pub")
	config.Seal()

	rpcClient, err := http.New(rpcUrl, "/websocket")
	if err != nil {
		return nil, err
	}
	hdPath := hd.NewFundraiserParams(0, sdktypes.CoinType, 0).String()
	derivedPrivKey, err := hd.Secp256k1.Derive()(mnemonic, "", hdPath)
	if err != nil {
		return nil, err
	}
	privKey := secp256k1.PrivKey{Key: derivedPrivKey[:32]}

	storkContract := &StorkContract{ContractAddress: contractAddress, ChainPrefix: chainPrefix}

	// set up execution context and factory
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	sdktypes.RegisterInterfaces(interfaceRegistry)
	authtypes.RegisterInterfaces(interfaceRegistry)
	cryptocodec.RegisterInterfaces(interfaceRegistry)
	wasmtypes.RegisterInterfaces(interfaceRegistry)

	marshaler := codec.NewProtoCodec(interfaceRegistry)
	storkContract.marshaler = marshaler
	txConfig := tx.NewTxConfig(marshaler, tx.DefaultSignModes)

	senderAddr := sdktypes.AccAddress(privKey.PubKey().Address())

	kr := keyring.NewInMemory(marshaler)
	keyName := privKey.PubKey().Address().String()
	err = kr.ImportPrivKeyHex(
		keyName,
		hex.EncodeToString(privKey.Key),
		"secp256k1",
	)
	if err != nil {
		return nil, fmt.Errorf("failed to import key: %w", err)
	}

	storkContract.clientCtx = sdkclient.Context{
		FromAddress:       senderAddr,
		ChainID:           chainID,
		FromName:          keyName,
		Client:            rpcClient,
		TxConfig:          txConfig,
		AccountRetriever:  authtypes.AccountRetriever{},
		InterfaceRegistry: interfaceRegistry,
		BroadcastMode:     "sync",
		Offline:           false,
		Keyring:           kr,
	}

	gasPriceStr := fmt.Sprintf("%.3f%s", gasPrice, denom)

	storkContract.txf = sdkclient_tx.Factory{}.
		WithChainID(chainID).
		WithGasPrices(gasPriceStr).
		WithGasAdjustment(gasAdjustment).
		WithSignMode(signing.SignMode_SIGN_MODE_DIRECT).
		WithTxConfig(storkContract.clientCtx.TxConfig).
		WithAccountRetriever(storkContract.clientCtx.AccountRetriever).
		WithKeybase(kr).
		WithFromName(keyName).
		WithSimulateAndExecute(true)

	singleUpdateFee, err := storkContract.GetSingleUpdateFee()
	if err != nil {
		return nil, err
	}
	storkContract.singleUpdateFee = singleUpdateFee.Fee

	return storkContract, nil
}

func (s *StorkContract) queryContract(rawQueryData []byte) ([]byte, error) {
	query := &wasmtypes.QuerySmartContractStateRequest{
		Address:   s.ContractAddress,
		QueryData: rawQueryData,
	}

	interfaceRegistry := codectypes.NewInterfaceRegistry()
	marshaler := codec.NewProtoCodec(interfaceRegistry)

	bz, err := marshaler.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query: %w", err)
	}

	result, err := s.clientCtx.Client.ABCIQuery(
		context.Background(),
		"/cosmwasm.wasm.v1.Query/SmartContractState",
		bz,
	)
	if err != nil {
		return nil, err
	}
	if result.Response.Code != 0 {
		return nil, fmt.Errorf("query failed with code %d", result.Response.Code)
	}

	var resp wasmtypes.QuerySmartContractStateResponse
	if err := marshaler.Unmarshal(result.Response.Value, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return resp.Data, nil
}

func (s *StorkContract) executeContract(rawExecData []byte, funds []sdktypes.Coin) (string, error) {

	senderBech32, err := sdktypes.Bech32ifyAddressBytes(s.ChainPrefix, s.clientCtx.FromAddress)
	if err != nil {
		return "", fmt.Errorf("failed to bech32ify address: %w", err)
	}

	msg := &wasmtypes.MsgExecuteContract{
		Sender:   senderBech32,
		Contract: s.ContractAddress,
		Msg:      rawExecData,
		Funds:    funds,
	}

	accMsg := &authtypes.QueryAccountRequest{
		Address: s.clientCtx.FromAddress.String(),
	}
	rawAccMsg, err := s.marshaler.Marshal(accMsg)
	if err != nil {
		return "", fmt.Errorf("failed to marshal account message: %w", err)
	}

	result, err := s.clientCtx.Client.ABCIQuery(
		context.Background(),
		"/cosmos.auth.v1beta1.Query/Account",
		rawAccMsg,
	)
	if err != nil {
		return "", fmt.Errorf("failed to query account: %w", err)
	}

	if result.Response.Value == nil {
		return "", fmt.Errorf("no account data returned for address %s", senderBech32)
	}

	var resp authtypes.QueryAccountResponse
	if err := s.marshaler.Unmarshal(result.Response.Value, &resp); err != nil {
		return "", fmt.Errorf("failed to unmarshal account: %w", err)
	}

	if resp.Account == nil {
		return "", fmt.Errorf("no account found for address %s", senderBech32)
	}

	var acc sdktypes.AccountI
	if err := s.clientCtx.InterfaceRegistry.UnpackAny(resp.Account, &acc); err != nil {
		return "", fmt.Errorf("failed to unpack account: %w", err)
	}

	if acc == nil {
		return "", fmt.Errorf("failed to decode account for address %s", senderBech32)
	}

	txf := s.txf.
		WithAccountNumber(acc.GetAccountNumber()).
		WithSequence(acc.GetSequence())

	_, adjusted, err := sdkclient_tx.CalculateGas(s.clientCtx, txf, msg)
	if err != nil {
		return "", fmt.Errorf("failed to calculate gas: %w", err)
	}
	txf = txf.WithGas(adjusted)

	tx, err := txf.BuildUnsignedTx(msg)
	if err != nil {
		return "", fmt.Errorf("failed to build unsigned transaction: %w", err)
	}

	if err = sdkclient_tx.Sign(s.clientCtx.CmdContext, txf, s.clientCtx.FromName, tx, true); err != nil {
		return "", fmt.Errorf("failed to sign transaction: %w", err)
	}

	encoder := s.clientCtx.TxConfig.TxEncoder()
	if encoder == nil {
		return "", fmt.Errorf("failed to encode transaction: tx json encoder is nil")
	}

	txBytes, err := encoder(tx.GetTx())
	if err != nil {
		return "", fmt.Errorf("failed to encode transaction: %w", err)
	}

	// broadcast to a CometBFT?
	res, err := s.clientCtx.BroadcastTx(txBytes)
	if err != nil {
		return "", fmt.Errorf("failed to broadcast transaction: %w", err)
	}
	if res.Code != 0 {
		return "", fmt.Errorf("transaction failed %s", res)
	}

	return res.TxHash, nil
}

func (s *StorkContract) GetSingleUpdateFee() (*GetSingleUpdateFeeResponse, error) {
	rawQueryData, err := json.Marshal(map[string]any{"get_single_update_fee": new(QueryMsg_GetSingleUpdateFee)})
	if err != nil {
		return nil, err
	}
	rawResponseData, err := s.queryContract(rawQueryData)
	if err != nil {
		return nil, err
	}
	var response GetSingleUpdateFeeResponse
	err = json.Unmarshal(rawResponseData, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

func (s *StorkContract) GetLatestCanonicalTemporalNumericValueUnchecked(id [32]int) (*GetTemporalNumericValueResponse, error) {
	rawQueryData, err := json.Marshal(map[string]any{"get_latest_canonical_temporal_numeric_value_unchecked": &QueryMsg_GetLatestCanonicalTemporalNumericValueUnchecked{Id: id}})
	if err != nil {
		return nil, err
	}
	rawResponseData, err := s.queryContract(rawQueryData)
	if err != nil {
		return nil, err
	}
	var response GetTemporalNumericValueResponse
	err = json.Unmarshal(rawResponseData, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	return &response, nil
}

func (s *StorkContract) UpdateTemporalNumericValuesEvm(updateData []UpdateData) (string, error) {
	rawExecData, err := json.Marshal(map[string]any{"update_temporal_numeric_values_evm": &ExecMsg_UpdateTemporalNumericValuesEvm{UpdateData: updateData}})
	if err != nil {
		return "", err
	}
	fee := s.singleUpdateFee.toCosmosCoin()
	fee.Amount = fee.Amount.MulRaw(int64(len(updateData)))
	txHash, err := s.executeContract(rawExecData, []sdktypes.Coin{fee})
	if err != nil {
		return "", err
	}
	return txHash, nil
}
