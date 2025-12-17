// These bindings are not generated.
// Instead, this file contains utility functions for interacting with the Initia MiniMove Stork contract.

package bindings

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/aptos-labs/serde-reflection/serde-generate/runtime/golang/serde"
	http "github.com/cometbft/cometbft/rpc/client/http"
	sdkclient "github.com/cosmos/cosmos-sdk/client"
	sdkclient_tx "github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	keyring "github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	proto "github.com/cosmos/gogoproto/proto"
	initiacodec "github.com/initia-labs/initia/crypto/codec"
	ethsecp256k1 "github.com/initia-labs/initia/crypto/ethsecp256k1"
	initiahd "github.com/initia-labs/initia/crypto/hd"
	initiakeyring "github.com/initia-labs/initia/crypto/keyring"
	vmtypes "github.com/initia-labs/movevm/types"
)

// Initia MiniMove constants from github.com/initia-labs/minimove/app/const.go
// These are standard for all MiniMove chains.
const (
	AccountAddressPrefix = "init"
	CoinType             = uint32(60) // SLIP44 coin type for Ethereum-compatible chains
)

var (
	ErrNilTxEncoder          = errors.New("tx encoder is nil")
	ErrFailedToDecodeAccount = errors.New("failed to decode account")
	ErrNoAccountFound        = errors.New("no account found")
	ErrQueryFailed           = errors.New("query failed")
	ErrTxFailed              = errors.New("transaction failed")
	ErrFeedNotFound          = errors.New("feed not found")
	ErrFailedToParseBigInt   = errors.New("failed to parse big int")
	ErrNoUpdatesProvided     = errors.New("no updates provided")
	ErrEmptyResponse         = errors.New("empty response from view function")
)

// ------------
// The following types and functions were copied and pasted from:
//  - https://github.com/initia-labs/initia/blob/main/x/move/types/query.pb.go.
//  - https://github.com/initia-labs/initia/blob/main/x/move/types/tx.pb.go.
// This was done to avoid having to link initias libmovevm library (https://github.com/initia-labs/movevm).
// libmovevm is required to compile the package containing these types but not actually used in the types.
// ------------

// QueryViewRequest is the request type for the QueryView
// RPC method.
type QueryViewRequest struct {
	// Address is the owner address of the module to query
	Address string `json:"address,omitempty" protobuf:"bytes,1,opt,name=address,proto3"`
	// ModuleName is the module name of the entry function to query
	ModuleName string `json:"module_name,omitempty" protobuf:"bytes,2,opt,name=module_name,json=moduleName,proto3"`
	// FunctionName is the name of a function to query
	FunctionName string `json:"function_name,omitempty" protobuf:"bytes,3,opt,name=function_name,json=functionName,proto3"`
	// TypeArgs is the type arguments of a function to execute
	// ex) "0x1::BasicCoin::Initia", "bool", "u8", "u64"
	TypeArgs []string `json:"type_args,omitempty" protobuf:"bytes,4,rep,name=type_args,json=typeArgs,proto3"`
	// Args is the arguments of a function to execute
	// - number: little endian
	// - string: base64 bytes
	Args [][]byte `json:"args,omitempty" protobuf:"bytes,5,rep,name=args,proto3"`
}

func (*QueryViewRequest) ProtoMessage()    {}
func (m *QueryViewRequest) Reset()         { *m = QueryViewRequest{} }
func (m *QueryViewRequest) String() string { return proto.CompactTextString(m) }

// VMEvent is the event emitted from vm.
type VMEvent struct {
	TypeTag string `json:"type_tag,omitempty" protobuf:"bytes,1,opt,name=type_tag,json=typeTag,proto3"`
	Data    string `json:"data,omitempty"     protobuf:"bytes,2,opt,name=data,proto3"`
}

// QueryViewResponse is the response type for the
// QueryView RPC method.
type QueryViewResponse struct {
	Data    string    `json:"data,omitempty"     protobuf:"bytes,1,opt,name=data,proto3"`
	Events  []VMEvent `json:"events"             protobuf:"bytes,2,rep,name=events,proto3"`
	GasUsed uint64    `json:"gas_used,omitempty" protobuf:"varint,3,opt,name=gas_used,json=gasUsed,proto3"`
}

func (m *QueryViewResponse) Reset()         { *m = QueryViewResponse{} }
func (m *QueryViewResponse) String() string { return proto.CompactTextString(m) }
func (*QueryViewResponse) ProtoMessage()    {}

// MsgExecute is the message to execute the given module function.
type MsgExecute struct {
	// Sender is the that actor that signed the messages
	Sender string `json:"sender,omitempty" protobuf:"bytes,1,opt,name=sender,proto3"`
	// ModuleAddr is the address of the module deployer
	//nolint:lll // copied code.
	ModuleAddress string `json:"module_address,omitempty" protobuf:"bytes,2,opt,name=module_address,json=moduleAddress,proto3"`
	// ModuleName is the name of module to execute
	ModuleName string `json:"module_name,omitempty" protobuf:"bytes,3,opt,name=module_name,json=moduleName,proto3"`
	// FunctionName is the name of a function to execute
	FunctionName string `json:"function_name,omitempty" protobuf:"bytes,4,opt,name=function_name,json=functionName,proto3"`
	// TypeArgs is the type arguments of a function to execute
	// ex) "0x1::BasicCoin::Initia", "bool", "u8", "u64"
	TypeArgs []string `json:"type_args,omitempty" protobuf:"bytes,5,rep,name=type_args,json=typeArgs,proto3"`
	// Args is the arguments of a function to execute
	// - number: little endian
	// - string: base64 bytes
	Args [][]byte `json:"args,omitempty" protobuf:"bytes,6,rep,name=args,proto3"`
}

func (m *MsgExecute) Reset()         { *m = MsgExecute{} }
func (m *MsgExecute) String() string { return proto.CompactTextString(m) }
func (*MsgExecute) ProtoMessage()    {}

// ------------
// End copied types
// ------------

// I128 represents a signed 128-bit integer with magnitude and sign.
type I128 struct {
	Magnitude *big.Int `json:"magnitude"`
	Negative  bool     `json:"negative"`
}

// UnmarshalJSON implements custom JSON unmarshaling for I128.
func (i *I128) UnmarshalJSON(data []byte) error {
	var raw struct {
		Magnitude string `json:"magnitude"`
		Negative  bool   `json:"negative"`
	}

	err := json.Unmarshal(data, &raw)
	if err != nil {
		return fmt.Errorf("failed to unmarshal i128: %w", err)
	}

	magnitude := new(big.Int)
	//nolint:mnd // base number.
	_, ok := magnitude.SetString(raw.Magnitude, 10)
	if !ok {
		return ErrFailedToParseBigInt
	}

	i.Magnitude = magnitude
	i.Negative = raw.Negative

	return nil
}

// TemporalNumericValue represents a timestamped value.
type TemporalNumericValue struct {
	TimestampNs    uint64 `json:"timestamp_ns,string"`
	QuantizedValue I128   `json:"quantized_value"`
}

// UpdateData is a data structure for an update to a Stork feed.
type UpdateData struct {
	ID                              []byte
	TemporalNumericValueTimestampNs uint64
	TemporalNumericValueMagnitude   *big.Int
	TemporalNumericValueNegative    bool
	PublisherMerkleRoot             []byte
	ValueComputeAlgHash             []byte
	R                               []byte
	S                               []byte
	V                               byte
}

type StorkContract struct {
	ContractAddress string
	ChainPrefix     string
	clientCtx       sdkclient.Context
	txf             sdkclient_tx.Factory
	marshaler       codec.Codec
}

func NewStorkContract(
	rpcUrl string,
	contractAddress string,
	mnemonic string,
	gasPrice float64,
	gasAdjustment float64,
	denom string,
	chainID string,
) (*StorkContract, error) {
	config := sdktypes.GetConfig()
	config.SetBech32PrefixForAccount(AccountAddressPrefix, AccountAddressPrefix+"pub")
	config.Seal()

	// Note: rpcUrl should be a Tendermint RPC endpoint (e.g., https://rpc.testnet.initia.xyz)
	// not a REST endpoint, despite the parameter name
	rpcClient, err := http.New(rpcUrl, "/websocket")
	if err != nil {
		return nil, fmt.Errorf("failed to create rpc http client: %w", err)
	}

	// Initia uses Ethereum-style derivation (coinType 60)
	hdPath := hd.NewFundraiserParams(0, CoinType, 0).String()

	// Use Initia's ethsecp256k1 algorithm for key derivation
	ethAlgo := initiahd.EthSecp256k1

	derivedPrivKey, err := ethAlgo.Derive()(mnemonic, "", hdPath)
	if err != nil {
		return nil, fmt.Errorf("failed to derive private key: %w", err)
	}

	privKey := ethsecp256k1.PrivKey{Key: derivedPrivKey}

	storkContract := &StorkContract{ContractAddress: contractAddress, ChainPrefix: AccountAddressPrefix}

	proto.RegisterType((*MsgExecute)(nil), "initia.move.v1.MsgExecute")
	// set up execution context and factory
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	sdktypes.RegisterInterfaces(interfaceRegistry)
	authtypes.RegisterInterfaces(interfaceRegistry)
	cryptocodec.RegisterInterfaces(interfaceRegistry)
	initiacodec.RegisterInterfaces(interfaceRegistry)

	interfaceRegistry.RegisterImplementations((*sdktypes.Msg)(nil), &MsgExecute{})

	marshaler := codec.NewProtoCodec(interfaceRegistry)
	storkContract.marshaler = marshaler
	txConfig := tx.NewTxConfig(marshaler, tx.DefaultSignModes)

	senderAddr := sdktypes.AccAddress(privKey.PubKey().Address())

	kr := keyring.NewInMemory(marshaler, initiakeyring.EthSecp256k1Option())
	keyName := privKey.PubKey().Address().String()

	err = kr.ImportPrivKeyHex(
		keyName,
		hex.EncodeToString(privKey.Key),
		ethsecp256k1.KeyType,
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

	return storkContract, nil
}

// GetTemporalNumericValueUnchecked queries the latest temporal numeric value for an asset.
func (s *StorkContract) GetTemporalNumericValueUnchecked(
	// TODO: pass ctx context.Context, 
	assetID []byte,
) (*TemporalNumericValue, error) {
	// Serialize the asset ID parameter using Initia's BCS serializer
	encodedArg, err := vmtypes.SerializeBytes(assetID)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize asset ID: %w", err)
	}

	// TODO: pass ctx
	// result, err := s.viewFunction(ctx, "stork", "get_temporal_numeric_value_unchecked", []string{}, [][]byte{encodedArg})
	result, err := s.viewFunction("stork", "get_temporal_numeric_value_unchecked", []string{}, [][]byte{encodedArg})
	if err != nil {
		if strings.Contains(err.Error(), "temporal_numeric_value_feed_registry, code=0") {
			return nil, ErrFeedNotFound
		}

		return nil, fmt.Errorf("failed to query temporal numeric value: %w", err)
	}

	if len(result) == 0 {
		return nil, ErrEmptyResponse
	}

	// Marshal the result back to JSON and unmarshal into our struct
	jsonBytes, err := json.Marshal(result[0])
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response: %w", err)
	}

	var value TemporalNumericValue

	err = json.Unmarshal(jsonBytes, &value)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal temporal numeric value: %w", err)
	}

	return &value, nil
}

// UpdateMultipleTemporalNumericValuesEvm updates multiple feeds with EVM signatures.
//
//nolint:funlen // permissible complexity for this function due to lack of nesting.
func (s *StorkContract) UpdateMultipleTemporalNumericValuesEvm(
	// TODO: pass ctx context.Context
	updateData []UpdateData,
) (string, error) {
	if len(updateData) == 0 {
		return "", ErrNoUpdatesProvided
	}

	// Prepare vectors for BCS serialization
	ids := make([][]byte, len(updateData))
	timestamps := make([]uint64, len(updateData))
	magnitudes := make([]*big.Int, len(updateData))
	negatives := make([]bool, len(updateData))
	merkleRoots := make([][]byte, len(updateData))
	algHashes := make([][]byte, len(updateData))
	rs := make([][]byte, len(updateData))
	ss := make([][]byte, len(updateData))
	vs := make([]byte, len(updateData))

	for i, data := range updateData {
		ids[i] = data.ID
		timestamps[i] = data.TemporalNumericValueTimestampNs
		magnitudes[i] = data.TemporalNumericValueMagnitude
		negatives[i] = data.TemporalNumericValueNegative
		merkleRoots[i] = data.PublisherMerkleRoot
		algHashes[i] = data.ValueComputeAlgHash
		rs[i] = data.R
		ss[i] = data.S
		vs[i] = data.V
	}

	// Serialize each argument using Initia's BCS serializers
	idsBytes, err := vmtypes.SerializeBytesVector(ids)
	if err != nil {
		return "", fmt.Errorf("failed to serialize ids: %w", err)
	}

	timestampsBytes, err := vmtypes.SerializeUint64Vector(timestamps)
	if err != nil {
		return "", fmt.Errorf("failed to serialize timestamps: %w", err)
	}

	// For u128 vector, we need to serialize manually as there's no direct function
	magnitudesBytes, err := serializeU128Vector(magnitudes)
	if err != nil {
		return "", fmt.Errorf("failed to serialize magnitudes: %w", err)
	}

	// For bool vector, we need to serialize manually
	negativesBytes, err := serializeBoolVector(negatives)
	if err != nil {
		return "", fmt.Errorf("failed to serialize negatives: %w", err)
	}

	merkleRootsBytes, err := vmtypes.SerializeBytesVector(merkleRoots)
	if err != nil {
		return "", fmt.Errorf("failed to serialize merkle roots: %w", err)
	}

	algHashesBytes, err := vmtypes.SerializeBytesVector(algHashes)
	if err != nil {
		return "", fmt.Errorf("failed to serialize alg hashes: %w", err)
	}

	rsBytes, err := vmtypes.SerializeBytesVector(rs)
	if err != nil {
		return "", fmt.Errorf("failed to serialize rs: %w", err)
	}

	ssBytes, err := vmtypes.SerializeBytesVector(ss)
	if err != nil {
		return "", fmt.Errorf("failed to serialize ss: %w", err)
	}

	// For u8 vector (vs), use SerializeBytes
	vsBytes, err := vmtypes.SerializeBytes(vs)
	if err != nil {
		return "", fmt.Errorf("failed to serialize vs: %w", err)
	}

	// Build the MsgExecute args as [][]byte
	args := [][]byte{
		idsBytes,
		timestampsBytes,
		magnitudesBytes,
		negativesBytes,
		merkleRootsBytes,
		algHashesBytes,
		rsBytes,
		ssBytes,
		vsBytes,
	}

	txHash, err := s.executeContract(
		// TODO: pass ctx,
		"stork",
		"update_multiple_temporal_numeric_values_evm",
		[]string{},
		args,
	)
	if err != nil {
		return "", fmt.Errorf("failed to execute contract: %w", err)
	}

	return txHash, nil
}

func (s *StorkContract) viewFunction(
	// TODO: pass ctx context.Context,
	moduleName string,
	functionName string,
	typeArgs []string,
	args [][]byte,
) ([]interface{}, error) {
	request := &QueryViewRequest{
		Address:      s.ContractAddress,
		ModuleName:   moduleName,
		FunctionName: functionName,
		TypeArgs:     typeArgs,
		Args:         args,
	}

	bz, err := s.marshaler.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query: %w", err)
	}

	result, err := s.clientCtx.Client.ABCIQuery(
		// TODO: pass ctx context.Context
		context.Background(),
		"/initia.move.v1.Query/View",
		bz,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query contract: %w", err)
	}

	if result.Response.Code != 0 {
		//nolint:err113 // This is effectively wrapping an error.
		return nil, fmt.Errorf("query failed with code %d: %s", result.Response.Code, result.Response.Log)
	}

	var resp QueryViewResponse

	err = s.marshaler.Unmarshal(result.Response.Value, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Parse JSON response
	// The response is wrapped in an array by the Move VM
	var jsonResult []interface{}

	err = json.Unmarshal([]byte(resp.Data), &jsonResult)
	if err != nil {
		// If it's not an array, try parsing as a single value and wrap it
		var singleResult interface{}

		err2 := json.Unmarshal([]byte(resp.Data), &singleResult)
		if err2 == nil {
			return []interface{}{singleResult}, nil
		}

		return nil, fmt.Errorf("failed to unmarshal JSON response: %w (data: %s)", err, resp.Data)
	}

	return jsonResult, nil
}

//nolint:cyclop,funlen // permissible complexity and funlen for this function due to lack of nesting.
func (s *StorkContract) executeContract(
	// TODO: pass ctx context.Context
	moduleName string,
	functionName string,
	typeArgs []string,
	args [][]byte,
) (string, error) {
	senderBech32, err := sdktypes.Bech32ifyAddressBytes(s.ChainPrefix, s.clientCtx.FromAddress)
	if err != nil {
		return "", fmt.Errorf("failed to bech32ify address: %w", err)
	}

	msg := &MsgExecute{
		Sender:        senderBech32,
		ModuleAddress: s.ContractAddress,
		ModuleName:    moduleName,
		FunctionName:  functionName,
		TypeArgs:      typeArgs,
		Args:          args,
	}

	accMsg := &authtypes.QueryAccountRequest{
		Address: s.clientCtx.FromAddress.String(),
	}

	rawAccMsg, err := s.marshaler.Marshal(accMsg)
	if err != nil {
		return "", fmt.Errorf("failed to marshal account message: %w", err)
	}

	result, err := s.clientCtx.Client.ABCIQuery(
		// TODO: pass ctx context.Context
		context.Background(),
		"/cosmos.auth.v1beta1.Query/Account",
		rawAccMsg,
	)
	if err != nil {
		return "", fmt.Errorf("failed to query account: %w", err)
	}

	if result.Response.Value == nil {
		return "", ErrNoAccountFound
	}

	var resp authtypes.QueryAccountResponse

	err = s.marshaler.Unmarshal(result.Response.Value, &resp)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal account: %w", err)
	}

	if resp.Account == nil {
		return "", ErrNoAccountFound
	}

	var acc sdktypes.AccountI

	err = s.clientCtx.InterfaceRegistry.UnpackAny(resp.Account, &acc)
	if err != nil {
		return "", fmt.Errorf("failed to unpack account: %w", err)
	}

	if acc == nil {
		return "", ErrFailedToDecodeAccount
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

	// TODO: pass ctx
	// err = sdkclient_tx.Sign(ctx, txf, s.clientCtx.FromName, tx, true)
	err = sdkclient_tx.Sign(s.clientCtx.CmdContext, txf, s.clientCtx.FromName, tx, true)
	if err != nil {
		return "", fmt.Errorf("failed to sign transaction: %w", err)
	}

	encoder := s.clientCtx.TxConfig.TxEncoder()
	if encoder == nil {
		return "", ErrNilTxEncoder
	}

	txBytes, err := encoder(tx.GetTx())
	if err != nil {
		return "", fmt.Errorf("failed to encode transaction: %w", err)
	}

	// broadcast to a CometBFT node
	res, err := s.clientCtx.BroadcastTx(txBytes)
	if err != nil {
		return "", fmt.Errorf("failed to broadcast transaction: %w", err)
	}

	if res.Code != 0 {
		//nolint:err113 // This is effectively wrapping an error.
		return "", fmt.Errorf("transaction failed with code %d: %s", res.Code, res.RawLog)
	}

	return res.TxHash, nil
}

// serializeU128Vector serializes a vector of u128 values using BCS.
func serializeU128Vector(values []*big.Int) ([]byte, error) {
	s := vmtypes.NewSerializer()
	// Serialize the length of the vector
	err := s.SerializeLen(uint64(len(values)))
	if err != nil {
		return nil, fmt.Errorf("failed to serialize length: %w", err)
	}
	// Serialize each u128 value
	for _, v := range values {
		// Convert big.Int to serde.Uint128
		u128 := bigIntToU128(v)

		err = s.SerializeU128(u128)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize u128: %w", err)
		}
	}

	return s.GetBytes(), nil
}

// bigIntToU128 converts a *big.Int to serde.Uint128.
//
//nolint:mnd // magic numbers for bit operations.
func bigIntToU128(v *big.Int) serde.Uint128 {
	// Get the bytes in big endian
	bytes := v.Bytes()

	// Pad to 16 bytes if needed
	if len(bytes) < 16 {
		padded := make([]byte, 16)
		copy(padded[16-len(bytes):], bytes)
		bytes = padded
	}

	// Extract high (first 8 bytes) and low (last 8 bytes)
	high := uint64(bytes[0])<<56 | uint64(bytes[1])<<48 | uint64(bytes[2])<<40 | uint64(bytes[3])<<32 |
		uint64(bytes[4])<<24 | uint64(bytes[5])<<16 | uint64(bytes[6])<<8 | uint64(bytes[7])
	low := uint64(bytes[8])<<56 | uint64(bytes[9])<<48 | uint64(bytes[10])<<40 | uint64(bytes[11])<<32 |
		uint64(bytes[12])<<24 | uint64(bytes[13])<<16 | uint64(bytes[14])<<8 | uint64(bytes[15])

	return serde.Uint128{High: high, Low: low}
}

// serializeBoolVector serializes a vector of bool values using BCS.
func serializeBoolVector(values []bool) ([]byte, error) {
	s := vmtypes.NewSerializer()
	// Serialize the length of the vector
	err := s.SerializeLen(uint64(len(values)))
	if err != nil {
		return nil, fmt.Errorf("failed to serialize length: %w", err)
	}
	// Serialize each bool value
	for _, v := range values {
		err = s.SerializeBool(v)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize bool: %w", err)
		}
	}

	return s.GetBytes(), nil
}
