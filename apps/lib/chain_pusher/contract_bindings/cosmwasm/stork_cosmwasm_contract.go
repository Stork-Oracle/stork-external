package contract_bindings_cosmwasm

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"

	cosmossdk_io_math "cosmossdk.io/math"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	rpcclient "github.com/cometbft/cometbft/rpc/client"
	http "github.com/cometbft/cometbft/rpc/client/http"
	sdkclient "github.com/cosmos/cosmos-sdk/client"
	sdkclienttx "github.com/cosmos/cosmos-sdk/client/tx"
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

// type Coin struct {
// 	Amount Uint128 `json:"amount"`
// 	Denom  string  `json:"denom"`
// }

type QueryMsg struct {
	GetLatestCanonicalTemporalNumericValueUnchecked *QueryMsg_GetLatestCanonicalTemporalNumericValueUnchecked `json:"get_latest_canonical_temporal_numeric_value_unchecked,omitempty"`
	GetSingleUpdateFee                              *QueryMsg_GetSingleUpdateFee                              `json:"get_single_update_fee,omitempty"`
	GetStorkEvmPublicKey                            *QueryMsg_GetStorkEvmPublicKey                            `json:"get_stork_evm_public_key,omitempty"`
	GetOwner                                        *QueryMsg_GetOwner                                        `json:"get_owner,omitempty"`
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

// Response for the `get_stork_evm_public_key` query containing the EVM public key set in the Stork contract. This is typically the Stork Aggregator's public key
type GetStorkEvmPublicKeyResponse struct {
	StorkEvmPublicKey [20]int `json:"stork_evm_public_key"`
}

// Response for the `get_owner` query containing the address owner of the Stork contract.
type GetOwnerResponse struct {
	Owner Addr `json:"owner"`
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
	SetSingleUpdateFee             *ExecMsg_SetSingleUpdateFee             `json:"set_single_update_fee,omitempty"`
	SetStorkEvmPublicKey           *ExecMsg_SetStorkEvmPublicKey           `json:"set_stork_evm_public_key,omitempty"`
	SetOwner                       *ExecMsg_SetOwner                       `json:"set_owner,omitempty"`
}

// Response for the `get_temporal_numeric_value` query containing the [`TemporalNumericValue`](./temporal_numeric_value.rs) for a given asset id.
type GetTemporalNumericValueResponse struct {
	TemporalNumericValue TemporalNumericValue `json:"temporal_numeric_value"`
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

type QueryMsg_GetStorkEvmPublicKey struct{}

type QueryMsg_GetOwner struct{}

type ExecMsg_UpdateTemporalNumericValuesEvm struct {
	UpdateData []UpdateData `json:"update_data"`
}

type ExecMsg_SetSingleUpdateFee struct {
	Fee Coin `json:"fee"`
}

type ExecMsg_SetStorkEvmPublicKey struct {
	StorkEvmPublicKey [20]int `json:"stork_evm_public_key"`
}

type ExecMsg_SetOwner struct {
	Owner Addr `json:"owner"`
}

type StorkContract struct {
	Client          rpcclient.Client
	ContractAddress string
	Key             secp256k1.PrivKey
	SingleUpdateFee Coin
	GasPrice        float64
	GasAdjustment   float64
	Denom           string
	ChainID         string
	ChainPrefix     string
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

	storkContract := &StorkContract{Client: rpcClient, ContractAddress: contractAddress, Key: privKey}
	singleUpdateFee, err := storkContract.GetSingleUpdateFee()
	if err != nil {
		return nil, err
	}
	storkContract.SingleUpdateFee = singleUpdateFee.Fee
	storkContract.GasPrice = gasPrice
	storkContract.GasAdjustment = gasAdjustment
	storkContract.Denom = denom
	storkContract.ChainID = chainID
	storkContract.ChainPrefix = chainPrefix
	return storkContract, nil
}

func (s *StorkContract) queryContract(rawQueryData []byte) ([]byte, error) {
	query := &wasmtypes.QuerySmartContractStateRequest{
		Address:   s.ContractAddress,
		QueryData: rawQueryData,
	}

	// Create the protobuf-encoded query
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	marshaler := codec.NewProtoCodec(interfaceRegistry)

	bz, err := marshaler.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query: %w", err)
	}

	result, err := s.Client.ABCIQuery(
		context.Background(),
		"/cosmwasm.wasm.v1.Query/SmartContractState",
		bz,
	)
	if err != nil {
		return nil, err
	}

	// Parse the response
	var resp wasmtypes.QuerySmartContractStateResponse
	if err := marshaler.Unmarshal(result.Response.Value, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return resp.Data, nil
}

func (s *StorkContract) executeContract(rawExecData []byte, funds []sdktypes.Coin) (string, error) {
	interfaceRegistry := codectypes.NewInterfaceRegistry()

	// Register base interfaces
	sdktypes.RegisterInterfaces(interfaceRegistry)
	authtypes.RegisterInterfaces(interfaceRegistry)
	cryptocodec.RegisterInterfaces(interfaceRegistry)
	wasmtypes.RegisterInterfaces(interfaceRegistry)

	marshaler := codec.NewProtoCodec(interfaceRegistry)
	txConfig := tx.NewTxConfig(marshaler, tx.DefaultSignModes)

	// Convert the address to bech32 with the correct prefix
	senderAddr := sdktypes.AccAddress(s.Key.PubKey().Address())
	senderBech32, err := sdktypes.Bech32ifyAddressBytes(s.ChainPrefix, senderAddr)
	if err != nil {
		return "", fmt.Errorf("failed to bech32ify address: %w", err)
	}

	msg := &wasmtypes.MsgExecuteContract{
		Sender:   senderBech32,
		Contract: s.ContractAddress,
		Msg:      rawExecData,
		Funds:    funds,
	}

	kr := keyring.NewInMemory(marshaler)

	keyName := s.Key.PubKey().Address().String()
	err = kr.ImportPrivKeyHex(
		keyName,
		hex.EncodeToString(s.Key.Key),
		"secp256k1",
	)
	if err != nil {
		return "", fmt.Errorf("failed to import key: %w", err)
	}

	clientCtx := sdkclient.Context{
		FromAddress: sdktypes.AccAddress(s.Key.PubKey().Address()),
		ChainID:     s.ChainID,
		FromName:    keyName,
		// NodeURI:           s.Client.Remote(),
		Client:            s.Client,
		TxConfig:          txConfig,
		AccountRetriever:  authtypes.AccountRetriever{},
		InterfaceRegistry: interfaceRegistry,
		BroadcastMode:     "sync",
		Offline:           false,
		Keyring:           kr,
	}

	gasPrice := fmt.Sprintf("%.3f%s", s.GasPrice, s.Denom)

	// Query account info using RPC
	// result, err := s.Client.ABCIQuery(
	// 	context.Background(),
	// 	"/cosmos.auth.v1beta1.Query/Account",
	// 	[]byte(fmt.Sprintf(`{"address":"%s"}`, senderBech32)),
	// )
	// if err != nil {
	// 	return "", fmt.Errorf("failed to query account: %w", err)
	// }

	// if result.Response.Value == nil {
	// 	return "", fmt.Errorf("no account data returned for address %s", senderBech32)
	// }

	// var resp authtypes.QueryAccountResponse
	// if err := marshaler.Unmarshal(result.Response.Value, &resp); err != nil {
	// 	return "", fmt.Errorf("failed to unmarshal account: %w", err)
	// }

	// if resp.Account == nil {
	// 	return "", fmt.Errorf("no account found for address %s", senderBech32)
	// }

	// var acc sdktypes.AccountI
	// if err := interfaceRegistry.UnpackAny(resp.Account, &acc); err != nil {
	// 	return "", fmt.Errorf("failed to unpack account: %w", err)
	// }

	// if acc == nil {
	// 	return "", fmt.Errorf("failed to decode account for address %s", senderBech32)
	// }

	// Use the account number and sequence
	txf := sdkclienttx.Factory{}.
		WithChainID(s.ChainID).
		WithGasPrices(gasPrice).
		WithGasAdjustment(s.GasAdjustment).
		WithSignMode(signing.SignMode_SIGN_MODE_DIRECT).
		WithTxConfig(clientCtx.TxConfig).
		WithAccountRetriever(clientCtx.AccountRetriever).
		WithKeybase(clientCtx.Keyring)
		// WithAccountNumber(acc.GetAccountNumber()).
		// WithSequence(acc.GetSequence())

	if txf.SimulateAndExecute() || clientCtx.Simulate {
		_, adjusted, err := sdkclienttx.CalculateGas(clientCtx, txf, msg)
		if err != nil {
			return "", fmt.Errorf("failed to calculate gas: %w", err)
		}
		txf = txf.WithGas(adjusted)
	}

	txBuilder := txConfig.NewTxBuilder()
	txBuilder.SetMsgs(msg)
	txBuilder.SetFeeAmount(sdktypes.NewCoins(sdktypes.NewCoin(s.Denom, cosmossdk_io_math.NewInt(int64(txf.Gas())))))
	txBuilder.SetGasLimit(1000000)
	txBuilder.SetMemo("")
	txBuilder.SetTimeoutHeight(0)

	// tx, err := txf.BuildUnsignedTx(msg)
	// if err != nil {
	// 	return "", fmt.Errorf("failed to build unsigned transaction: %w", err)
	// }

	if err = sdkclienttx.Sign(clientCtx.CmdContext, txf, clientCtx.FromName, txBuilder, true); err != nil {
		return "", fmt.Errorf("failed to sign transaction: %w", err)
	}

	encoder := clientCtx.TxConfig.TxEncoder()
	if encoder == nil {
		return "", fmt.Errorf("failed to encode transaction: tx json encoder is nil")
	}

	txBytes, err := encoder(txBuilder.GetTx())
	if err != nil {
		return "", fmt.Errorf("failed to encode transaction: %w", err)
	}

	// broadcast to a CometBFT?
	// res, err := clientCtx.BroadcastTx(txBytes)
	// if err != nil {
	// 	return "", fmt.Errorf("failed to broadcast transaction: %w", err)
	// }
	// // txClient := tx.NewServiceClient(s.Client)
	// if res.Code != 0 {
	// 	return "", fmt.Errorf("transaction failed %s", res)
	// }

	res, err := s.Client.BroadcastTxSync(context.Background(), txBytes)
	if err != nil {
		return "", fmt.Errorf("failed to broadcast transaction: %w", err)
	}
	return "", fmt.Errorf("res: %v", res)

	return hex.EncodeToString(res.Hash), nil
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
		return nil, err
	}
	return &response, nil
}

func (s *StorkContract) UpdateTemporalNumericValuesEvm(updateData []UpdateData) (string, error) {
	rawExecData, err := json.Marshal(map[string]any{"update_temporal_numeric_values_evm": &ExecMsg_UpdateTemporalNumericValuesEvm{UpdateData: updateData}})
	if err != nil {
		return "", err
	}
	fee := s.SingleUpdateFee.toCosmosCoin()
	fee.Amount = fee.Amount.MulRaw(int64(len(updateData)))
	txHash, err := s.executeContract(rawExecData, []sdktypes.Coin{fee})
	if err != nil {
		return "", err
	}
	return txHash, nil
}
