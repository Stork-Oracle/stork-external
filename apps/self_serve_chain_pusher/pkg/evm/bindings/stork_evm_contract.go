// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package bindings

import (
	"errors"
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
	_ = abi.ConvertType
)

// SelfServeStorkStructsPublisherTemporalNumericValueInput is an auto generated low-level Go binding around an user-defined struct.
type SelfServeStorkStructsPublisherTemporalNumericValueInput struct {
	TemporalNumericValue SelfServeStorkStructsTemporalNumericValue
	PubKey               common.Address
	AssetPairId          string
	R                    [32]byte
	S                    [32]byte
	V                    uint8
}

// SelfServeStorkStructsPublisherUser is an auto generated low-level Go binding around an user-defined struct.
type SelfServeStorkStructsPublisherUser struct {
	PubKey          common.Address
	SingleUpdateFee *big.Int
}

// SelfServeStorkStructsTemporalNumericValue is an auto generated low-level Go binding around an user-defined struct.
type SelfServeStorkStructsTemporalNumericValue struct {
	TimestampNs    uint64
	QuantizedValue *big.Int
}

// SelfServeStorkContractMetaData contains all meta data concerning the SelfServeStorkContract contract.
var SelfServeStorkContractMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"name\":\"ECDSAInvalidSignature\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"length\",\"type\":\"uint256\"}],\"name\":\"ECDSAInvalidSignatureLength\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"s\",\"type\":\"bytes32\"}],\"name\":\"ECDSAInvalidSignatureS\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InsufficientFee\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidSignature\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NoFreshUpdate\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotFound\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"pubKey\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"string\",\"name\":\"assetId\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"timestampNs\",\"type\":\"uint64\"},{\"indexed\":false,\"internalType\":\"int192\",\"name\":\"quantizedValue\",\"type\":\"int192\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"roundId\",\"type\":\"uint256\"}],\"name\":\"HistoricalValueStored\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"pubKey\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"singleUpdateFee\",\"type\":\"uint256\"}],\"name\":\"PublisherUserAdded\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"pubKey\",\"type\":\"address\"}],\"name\":\"PublisherUserRemoved\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"pubKey\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"string\",\"name\":\"assetId\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"timestampNs\",\"type\":\"uint64\"},{\"indexed\":false,\"internalType\":\"int192\",\"name\":\"quantizedValue\",\"type\":\"int192\"}],\"name\":\"ValueUpdate\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"pubKey\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"singleUpdateFee\",\"type\":\"uint256\"}],\"name\":\"createPublisherUser\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"pubKey\",\"type\":\"address\"}],\"name\":\"deletePublisherUser\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"pubKey\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"assetPairId\",\"type\":\"string\"}],\"name\":\"getCurrentRoundId\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"pubKey\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"assetPairId\",\"type\":\"string\"}],\"name\":\"getHistoricalRecordsCount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"pubKey\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"assetPairId\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"roundId\",\"type\":\"uint256\"}],\"name\":\"getHistoricalTemporalNumericValue\",\"outputs\":[{\"components\":[{\"internalType\":\"uint64\",\"name\":\"timestampNs\",\"type\":\"uint64\"},{\"internalType\":\"int192\",\"name\":\"quantizedValue\",\"type\":\"int192\"}],\"internalType\":\"structSelfServeStorkStructs.TemporalNumericValue\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"pubKey\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"assetPairId\",\"type\":\"string\"}],\"name\":\"getLatestTemporalNumericValue\",\"outputs\":[{\"components\":[{\"internalType\":\"uint64\",\"name\":\"timestampNs\",\"type\":\"uint64\"},{\"internalType\":\"int192\",\"name\":\"quantizedValue\",\"type\":\"int192\"}],\"internalType\":\"structSelfServeStorkStructs.TemporalNumericValue\",\"name\":\"value\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"pubKey\",\"type\":\"address\"}],\"name\":\"getPublisherUser\",\"outputs\":[{\"components\":[{\"internalType\":\"address\",\"name\":\"pubKey\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"singleUpdateFee\",\"type\":\"uint256\"}],\"internalType\":\"structSelfServeStorkStructs.PublisherUser\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"components\":[{\"internalType\":\"uint64\",\"name\":\"timestampNs\",\"type\":\"uint64\"},{\"internalType\":\"int192\",\"name\":\"quantizedValue\",\"type\":\"int192\"}],\"internalType\":\"structSelfServeStorkStructs.TemporalNumericValue\",\"name\":\"temporalNumericValue\",\"type\":\"tuple\"},{\"internalType\":\"address\",\"name\":\"pubKey\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"assetPairId\",\"type\":\"string\"},{\"internalType\":\"bytes32\",\"name\":\"r\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"s\",\"type\":\"bytes32\"},{\"internalType\":\"uint8\",\"name\":\"v\",\"type\":\"uint8\"}],\"internalType\":\"structSelfServeStorkStructs.PublisherTemporalNumericValueInput[]\",\"name\":\"updateData\",\"type\":\"tuple[]\"},{\"internalType\":\"bool\",\"name\":\"storeHistoric\",\"type\":\"bool\"}],\"name\":\"updateTemporalNumericValues\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"oraclePubKey\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"assetPairId\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"},{\"internalType\":\"int256\",\"name\":\"value\",\"type\":\"int256\"},{\"internalType\":\"bytes32\",\"name\":\"r\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"s\",\"type\":\"bytes32\"},{\"internalType\":\"uint8\",\"name\":\"v\",\"type\":\"uint8\"}],\"name\":\"verifyPublisherSignatureV1\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"pure\",\"type\":\"function\"}]",
}

// SelfServeStorkContractABI is the input ABI used to generate the binding from.
// Deprecated: Use SelfServeStorkContractMetaData.ABI instead.
var SelfServeStorkContractABI = SelfServeStorkContractMetaData.ABI

// SelfServeStorkContract is an auto generated Go binding around an Ethereum contract.
type SelfServeStorkContract struct {
	SelfServeStorkContractCaller     // Read-only binding to the contract
	SelfServeStorkContractTransactor // Write-only binding to the contract
	SelfServeStorkContractFilterer   // Log filterer for contract events
}

// SelfServeStorkContractCaller is an auto generated read-only Go binding around an Ethereum contract.
type SelfServeStorkContractCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SelfServeStorkContractTransactor is an auto generated write-only Go binding around an Ethereum contract.
type SelfServeStorkContractTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SelfServeStorkContractFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type SelfServeStorkContractFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SelfServeStorkContractSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type SelfServeStorkContractSession struct {
	Contract     *SelfServeStorkContract // Generic contract binding to set the session for
	CallOpts     bind.CallOpts           // Call options to use throughout this session
	TransactOpts bind.TransactOpts       // Transaction auth options to use throughout this session
}

// SelfServeStorkContractCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type SelfServeStorkContractCallerSession struct {
	Contract *SelfServeStorkContractCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts                 // Call options to use throughout this session
}

// SelfServeStorkContractTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type SelfServeStorkContractTransactorSession struct {
	Contract     *SelfServeStorkContractTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts                 // Transaction auth options to use throughout this session
}

// SelfServeStorkContractRaw is an auto generated low-level Go binding around an Ethereum contract.
type SelfServeStorkContractRaw struct {
	Contract *SelfServeStorkContract // Generic contract binding to access the raw methods on
}

// SelfServeStorkContractCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type SelfServeStorkContractCallerRaw struct {
	Contract *SelfServeStorkContractCaller // Generic read-only contract binding to access the raw methods on
}

// SelfServeStorkContractTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type SelfServeStorkContractTransactorRaw struct {
	Contract *SelfServeStorkContractTransactor // Generic write-only contract binding to access the raw methods on
}

// NewSelfServeStorkContract creates a new instance of SelfServeStorkContract, bound to a specific deployed contract.
func NewSelfServeStorkContract(address common.Address, backend bind.ContractBackend) (*SelfServeStorkContract, error) {
	contract, err := bindSelfServeStorkContract(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &SelfServeStorkContract{SelfServeStorkContractCaller: SelfServeStorkContractCaller{contract: contract}, SelfServeStorkContractTransactor: SelfServeStorkContractTransactor{contract: contract}, SelfServeStorkContractFilterer: SelfServeStorkContractFilterer{contract: contract}}, nil
}

// NewSelfServeStorkContractCaller creates a new read-only instance of SelfServeStorkContract, bound to a specific deployed contract.
func NewSelfServeStorkContractCaller(address common.Address, caller bind.ContractCaller) (*SelfServeStorkContractCaller, error) {
	contract, err := bindSelfServeStorkContract(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &SelfServeStorkContractCaller{contract: contract}, nil
}

// NewSelfServeStorkContractTransactor creates a new write-only instance of SelfServeStorkContract, bound to a specific deployed contract.
func NewSelfServeStorkContractTransactor(address common.Address, transactor bind.ContractTransactor) (*SelfServeStorkContractTransactor, error) {
	contract, err := bindSelfServeStorkContract(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &SelfServeStorkContractTransactor{contract: contract}, nil
}

// NewSelfServeStorkContractFilterer creates a new log filterer instance of SelfServeStorkContract, bound to a specific deployed contract.
func NewSelfServeStorkContractFilterer(address common.Address, filterer bind.ContractFilterer) (*SelfServeStorkContractFilterer, error) {
	contract, err := bindSelfServeStorkContract(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &SelfServeStorkContractFilterer{contract: contract}, nil
}

// bindSelfServeStorkContract binds a generic wrapper to an already deployed contract.
func bindSelfServeStorkContract(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := SelfServeStorkContractMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_SelfServeStorkContract *SelfServeStorkContractRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _SelfServeStorkContract.Contract.SelfServeStorkContractCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_SelfServeStorkContract *SelfServeStorkContractRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SelfServeStorkContract.Contract.SelfServeStorkContractTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_SelfServeStorkContract *SelfServeStorkContractRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _SelfServeStorkContract.Contract.SelfServeStorkContractTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_SelfServeStorkContract *SelfServeStorkContractCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _SelfServeStorkContract.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_SelfServeStorkContract *SelfServeStorkContractTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SelfServeStorkContract.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_SelfServeStorkContract *SelfServeStorkContractTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _SelfServeStorkContract.Contract.contract.Transact(opts, method, params...)
}

// GetCurrentRoundId is a free data retrieval call binding the contract method 0xb7f16567.
//
// Solidity: function getCurrentRoundId(address pubKey, string assetPairId) view returns(uint256)
func (_SelfServeStorkContract *SelfServeStorkContractCaller) GetCurrentRoundId(opts *bind.CallOpts, pubKey common.Address, assetPairId string) (*big.Int, error) {
	var out []interface{}
	err := _SelfServeStorkContract.contract.Call(opts, &out, "getCurrentRoundId", pubKey, assetPairId)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetCurrentRoundId is a free data retrieval call binding the contract method 0xb7f16567.
//
// Solidity: function getCurrentRoundId(address pubKey, string assetPairId) view returns(uint256)
func (_SelfServeStorkContract *SelfServeStorkContractSession) GetCurrentRoundId(pubKey common.Address, assetPairId string) (*big.Int, error) {
	return _SelfServeStorkContract.Contract.GetCurrentRoundId(&_SelfServeStorkContract.CallOpts, pubKey, assetPairId)
}

// GetCurrentRoundId is a free data retrieval call binding the contract method 0xb7f16567.
//
// Solidity: function getCurrentRoundId(address pubKey, string assetPairId) view returns(uint256)
func (_SelfServeStorkContract *SelfServeStorkContractCallerSession) GetCurrentRoundId(pubKey common.Address, assetPairId string) (*big.Int, error) {
	return _SelfServeStorkContract.Contract.GetCurrentRoundId(&_SelfServeStorkContract.CallOpts, pubKey, assetPairId)
}

// GetHistoricalRecordsCount is a free data retrieval call binding the contract method 0x543b2fc1.
//
// Solidity: function getHistoricalRecordsCount(address pubKey, string assetPairId) view returns(uint256)
func (_SelfServeStorkContract *SelfServeStorkContractCaller) GetHistoricalRecordsCount(opts *bind.CallOpts, pubKey common.Address, assetPairId string) (*big.Int, error) {
	var out []interface{}
	err := _SelfServeStorkContract.contract.Call(opts, &out, "getHistoricalRecordsCount", pubKey, assetPairId)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetHistoricalRecordsCount is a free data retrieval call binding the contract method 0x543b2fc1.
//
// Solidity: function getHistoricalRecordsCount(address pubKey, string assetPairId) view returns(uint256)
func (_SelfServeStorkContract *SelfServeStorkContractSession) GetHistoricalRecordsCount(pubKey common.Address, assetPairId string) (*big.Int, error) {
	return _SelfServeStorkContract.Contract.GetHistoricalRecordsCount(&_SelfServeStorkContract.CallOpts, pubKey, assetPairId)
}

// GetHistoricalRecordsCount is a free data retrieval call binding the contract method 0x543b2fc1.
//
// Solidity: function getHistoricalRecordsCount(address pubKey, string assetPairId) view returns(uint256)
func (_SelfServeStorkContract *SelfServeStorkContractCallerSession) GetHistoricalRecordsCount(pubKey common.Address, assetPairId string) (*big.Int, error) {
	return _SelfServeStorkContract.Contract.GetHistoricalRecordsCount(&_SelfServeStorkContract.CallOpts, pubKey, assetPairId)
}

// GetHistoricalTemporalNumericValue is a free data retrieval call binding the contract method 0x2651ad06.
//
// Solidity: function getHistoricalTemporalNumericValue(address pubKey, string assetPairId, uint256 roundId) view returns((uint64,int192))
func (_SelfServeStorkContract *SelfServeStorkContractCaller) GetHistoricalTemporalNumericValue(opts *bind.CallOpts, pubKey common.Address, assetPairId string, roundId *big.Int) (SelfServeStorkStructsTemporalNumericValue, error) {
	var out []interface{}
	err := _SelfServeStorkContract.contract.Call(opts, &out, "getHistoricalTemporalNumericValue", pubKey, assetPairId, roundId)

	if err != nil {
		return *new(SelfServeStorkStructsTemporalNumericValue), err
	}

	out0 := *abi.ConvertType(out[0], new(SelfServeStorkStructsTemporalNumericValue)).(*SelfServeStorkStructsTemporalNumericValue)

	return out0, err

}

// GetHistoricalTemporalNumericValue is a free data retrieval call binding the contract method 0x2651ad06.
//
// Solidity: function getHistoricalTemporalNumericValue(address pubKey, string assetPairId, uint256 roundId) view returns((uint64,int192))
func (_SelfServeStorkContract *SelfServeStorkContractSession) GetHistoricalTemporalNumericValue(pubKey common.Address, assetPairId string, roundId *big.Int) (SelfServeStorkStructsTemporalNumericValue, error) {
	return _SelfServeStorkContract.Contract.GetHistoricalTemporalNumericValue(&_SelfServeStorkContract.CallOpts, pubKey, assetPairId, roundId)
}

// GetHistoricalTemporalNumericValue is a free data retrieval call binding the contract method 0x2651ad06.
//
// Solidity: function getHistoricalTemporalNumericValue(address pubKey, string assetPairId, uint256 roundId) view returns((uint64,int192))
func (_SelfServeStorkContract *SelfServeStorkContractCallerSession) GetHistoricalTemporalNumericValue(pubKey common.Address, assetPairId string, roundId *big.Int) (SelfServeStorkStructsTemporalNumericValue, error) {
	return _SelfServeStorkContract.Contract.GetHistoricalTemporalNumericValue(&_SelfServeStorkContract.CallOpts, pubKey, assetPairId, roundId)
}

// GetLatestTemporalNumericValue is a free data retrieval call binding the contract method 0xea419887.
//
// Solidity: function getLatestTemporalNumericValue(address pubKey, string assetPairId) view returns((uint64,int192) value)
func (_SelfServeStorkContract *SelfServeStorkContractCaller) GetLatestTemporalNumericValue(opts *bind.CallOpts, pubKey common.Address, assetPairId string) (SelfServeStorkStructsTemporalNumericValue, error) {
	var out []interface{}
	err := _SelfServeStorkContract.contract.Call(opts, &out, "getLatestTemporalNumericValue", pubKey, assetPairId)

	if err != nil {
		return *new(SelfServeStorkStructsTemporalNumericValue), err
	}

	out0 := *abi.ConvertType(out[0], new(SelfServeStorkStructsTemporalNumericValue)).(*SelfServeStorkStructsTemporalNumericValue)

	return out0, err

}

// GetLatestTemporalNumericValue is a free data retrieval call binding the contract method 0xea419887.
//
// Solidity: function getLatestTemporalNumericValue(address pubKey, string assetPairId) view returns((uint64,int192) value)
func (_SelfServeStorkContract *SelfServeStorkContractSession) GetLatestTemporalNumericValue(pubKey common.Address, assetPairId string) (SelfServeStorkStructsTemporalNumericValue, error) {
	return _SelfServeStorkContract.Contract.GetLatestTemporalNumericValue(&_SelfServeStorkContract.CallOpts, pubKey, assetPairId)
}

// GetLatestTemporalNumericValue is a free data retrieval call binding the contract method 0xea419887.
//
// Solidity: function getLatestTemporalNumericValue(address pubKey, string assetPairId) view returns((uint64,int192) value)
func (_SelfServeStorkContract *SelfServeStorkContractCallerSession) GetLatestTemporalNumericValue(pubKey common.Address, assetPairId string) (SelfServeStorkStructsTemporalNumericValue, error) {
	return _SelfServeStorkContract.Contract.GetLatestTemporalNumericValue(&_SelfServeStorkContract.CallOpts, pubKey, assetPairId)
}

// GetPublisherUser is a free data retrieval call binding the contract method 0x3d57a294.
//
// Solidity: function getPublisherUser(address pubKey) view returns((address,uint256))
func (_SelfServeStorkContract *SelfServeStorkContractCaller) GetPublisherUser(opts *bind.CallOpts, pubKey common.Address) (SelfServeStorkStructsPublisherUser, error) {
	var out []interface{}
	err := _SelfServeStorkContract.contract.Call(opts, &out, "getPublisherUser", pubKey)

	if err != nil {
		return *new(SelfServeStorkStructsPublisherUser), err
	}

	out0 := *abi.ConvertType(out[0], new(SelfServeStorkStructsPublisherUser)).(*SelfServeStorkStructsPublisherUser)

	return out0, err

}

// GetPublisherUser is a free data retrieval call binding the contract method 0x3d57a294.
//
// Solidity: function getPublisherUser(address pubKey) view returns((address,uint256))
func (_SelfServeStorkContract *SelfServeStorkContractSession) GetPublisherUser(pubKey common.Address) (SelfServeStorkStructsPublisherUser, error) {
	return _SelfServeStorkContract.Contract.GetPublisherUser(&_SelfServeStorkContract.CallOpts, pubKey)
}

// GetPublisherUser is a free data retrieval call binding the contract method 0x3d57a294.
//
// Solidity: function getPublisherUser(address pubKey) view returns((address,uint256))
func (_SelfServeStorkContract *SelfServeStorkContractCallerSession) GetPublisherUser(pubKey common.Address) (SelfServeStorkStructsPublisherUser, error) {
	return _SelfServeStorkContract.Contract.GetPublisherUser(&_SelfServeStorkContract.CallOpts, pubKey)
}

// VerifyPublisherSignatureV1 is a free data retrieval call binding the contract method 0x9bccd2d5.
//
// Solidity: function verifyPublisherSignatureV1(address oraclePubKey, string assetPairId, uint256 timestamp, int256 value, bytes32 r, bytes32 s, uint8 v) pure returns(bool)
func (_SelfServeStorkContract *SelfServeStorkContractCaller) VerifyPublisherSignatureV1(opts *bind.CallOpts, oraclePubKey common.Address, assetPairId string, timestamp *big.Int, value *big.Int, r [32]byte, s [32]byte, v uint8) (bool, error) {
	var out []interface{}
	err := _SelfServeStorkContract.contract.Call(opts, &out, "verifyPublisherSignatureV1", oraclePubKey, assetPairId, timestamp, value, r, s, v)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// VerifyPublisherSignatureV1 is a free data retrieval call binding the contract method 0x9bccd2d5.
//
// Solidity: function verifyPublisherSignatureV1(address oraclePubKey, string assetPairId, uint256 timestamp, int256 value, bytes32 r, bytes32 s, uint8 v) pure returns(bool)
func (_SelfServeStorkContract *SelfServeStorkContractSession) VerifyPublisherSignatureV1(oraclePubKey common.Address, assetPairId string, timestamp *big.Int, value *big.Int, r [32]byte, s [32]byte, v uint8) (bool, error) {
	return _SelfServeStorkContract.Contract.VerifyPublisherSignatureV1(&_SelfServeStorkContract.CallOpts, oraclePubKey, assetPairId, timestamp, value, r, s, v)
}

// VerifyPublisherSignatureV1 is a free data retrieval call binding the contract method 0x9bccd2d5.
//
// Solidity: function verifyPublisherSignatureV1(address oraclePubKey, string assetPairId, uint256 timestamp, int256 value, bytes32 r, bytes32 s, uint8 v) pure returns(bool)
func (_SelfServeStorkContract *SelfServeStorkContractCallerSession) VerifyPublisherSignatureV1(oraclePubKey common.Address, assetPairId string, timestamp *big.Int, value *big.Int, r [32]byte, s [32]byte, v uint8) (bool, error) {
	return _SelfServeStorkContract.Contract.VerifyPublisherSignatureV1(&_SelfServeStorkContract.CallOpts, oraclePubKey, assetPairId, timestamp, value, r, s, v)
}

// CreatePublisherUser is a paid mutator transaction binding the contract method 0x714fbcad.
//
// Solidity: function createPublisherUser(address pubKey, uint256 singleUpdateFee) returns()
func (_SelfServeStorkContract *SelfServeStorkContractTransactor) CreatePublisherUser(opts *bind.TransactOpts, pubKey common.Address, singleUpdateFee *big.Int) (*types.Transaction, error) {
	return _SelfServeStorkContract.contract.Transact(opts, "createPublisherUser", pubKey, singleUpdateFee)
}

// CreatePublisherUser is a paid mutator transaction binding the contract method 0x714fbcad.
//
// Solidity: function createPublisherUser(address pubKey, uint256 singleUpdateFee) returns()
func (_SelfServeStorkContract *SelfServeStorkContractSession) CreatePublisherUser(pubKey common.Address, singleUpdateFee *big.Int) (*types.Transaction, error) {
	return _SelfServeStorkContract.Contract.CreatePublisherUser(&_SelfServeStorkContract.TransactOpts, pubKey, singleUpdateFee)
}

// CreatePublisherUser is a paid mutator transaction binding the contract method 0x714fbcad.
//
// Solidity: function createPublisherUser(address pubKey, uint256 singleUpdateFee) returns()
func (_SelfServeStorkContract *SelfServeStorkContractTransactorSession) CreatePublisherUser(pubKey common.Address, singleUpdateFee *big.Int) (*types.Transaction, error) {
	return _SelfServeStorkContract.Contract.CreatePublisherUser(&_SelfServeStorkContract.TransactOpts, pubKey, singleUpdateFee)
}

// DeletePublisherUser is a paid mutator transaction binding the contract method 0x488487a9.
//
// Solidity: function deletePublisherUser(address pubKey) returns()
func (_SelfServeStorkContract *SelfServeStorkContractTransactor) DeletePublisherUser(opts *bind.TransactOpts, pubKey common.Address) (*types.Transaction, error) {
	return _SelfServeStorkContract.contract.Transact(opts, "deletePublisherUser", pubKey)
}

// DeletePublisherUser is a paid mutator transaction binding the contract method 0x488487a9.
//
// Solidity: function deletePublisherUser(address pubKey) returns()
func (_SelfServeStorkContract *SelfServeStorkContractSession) DeletePublisherUser(pubKey common.Address) (*types.Transaction, error) {
	return _SelfServeStorkContract.Contract.DeletePublisherUser(&_SelfServeStorkContract.TransactOpts, pubKey)
}

// DeletePublisherUser is a paid mutator transaction binding the contract method 0x488487a9.
//
// Solidity: function deletePublisherUser(address pubKey) returns()
func (_SelfServeStorkContract *SelfServeStorkContractTransactorSession) DeletePublisherUser(pubKey common.Address) (*types.Transaction, error) {
	return _SelfServeStorkContract.Contract.DeletePublisherUser(&_SelfServeStorkContract.TransactOpts, pubKey)
}

// UpdateTemporalNumericValues is a paid mutator transaction binding the contract method 0x25ae409f.
//
// Solidity: function updateTemporalNumericValues(((uint64,int192),address,string,bytes32,bytes32,uint8)[] updateData, bool storeHistoric) payable returns()
func (_SelfServeStorkContract *SelfServeStorkContractTransactor) UpdateTemporalNumericValues(opts *bind.TransactOpts, updateData []SelfServeStorkStructsPublisherTemporalNumericValueInput, storeHistoric bool) (*types.Transaction, error) {
	return _SelfServeStorkContract.contract.Transact(opts, "updateTemporalNumericValues", updateData, storeHistoric)
}

// UpdateTemporalNumericValues is a paid mutator transaction binding the contract method 0x25ae409f.
//
// Solidity: function updateTemporalNumericValues(((uint64,int192),address,string,bytes32,bytes32,uint8)[] updateData, bool storeHistoric) payable returns()
func (_SelfServeStorkContract *SelfServeStorkContractSession) UpdateTemporalNumericValues(updateData []SelfServeStorkStructsPublisherTemporalNumericValueInput, storeHistoric bool) (*types.Transaction, error) {
	return _SelfServeStorkContract.Contract.UpdateTemporalNumericValues(&_SelfServeStorkContract.TransactOpts, updateData, storeHistoric)
}

// UpdateTemporalNumericValues is a paid mutator transaction binding the contract method 0x25ae409f.
//
// Solidity: function updateTemporalNumericValues(((uint64,int192),address,string,bytes32,bytes32,uint8)[] updateData, bool storeHistoric) payable returns()
func (_SelfServeStorkContract *SelfServeStorkContractTransactorSession) UpdateTemporalNumericValues(updateData []SelfServeStorkStructsPublisherTemporalNumericValueInput, storeHistoric bool) (*types.Transaction, error) {
	return _SelfServeStorkContract.Contract.UpdateTemporalNumericValues(&_SelfServeStorkContract.TransactOpts, updateData, storeHistoric)
}

// SelfServeStorkContractHistoricalValueStoredIterator is returned from FilterHistoricalValueStored and is used to iterate over the raw logs and unpacked data for HistoricalValueStored events raised by the SelfServeStorkContract contract.
type SelfServeStorkContractHistoricalValueStoredIterator struct {
	Event *SelfServeStorkContractHistoricalValueStored // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *SelfServeStorkContractHistoricalValueStoredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SelfServeStorkContractHistoricalValueStored)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(SelfServeStorkContractHistoricalValueStored)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *SelfServeStorkContractHistoricalValueStoredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SelfServeStorkContractHistoricalValueStoredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SelfServeStorkContractHistoricalValueStored represents a HistoricalValueStored event raised by the SelfServeStorkContract contract.
type SelfServeStorkContractHistoricalValueStored struct {
	PubKey         common.Address
	AssetId        common.Hash
	TimestampNs    uint64
	QuantizedValue *big.Int
	RoundId        *big.Int
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterHistoricalValueStored is a free log retrieval operation binding the contract event 0xfbb8f1bb7c4b5b8719a0357ba08c979650b113503c79ae1415088c0d6e429a8c.
//
// Solidity: event HistoricalValueStored(address indexed pubKey, string indexed assetId, uint64 timestampNs, int192 quantizedValue, uint256 roundId)
func (_SelfServeStorkContract *SelfServeStorkContractFilterer) FilterHistoricalValueStored(opts *bind.FilterOpts, pubKey []common.Address, assetId []string) (*SelfServeStorkContractHistoricalValueStoredIterator, error) {

	var pubKeyRule []interface{}
	for _, pubKeyItem := range pubKey {
		pubKeyRule = append(pubKeyRule, pubKeyItem)
	}
	var assetIdRule []interface{}
	for _, assetIdItem := range assetId {
		assetIdRule = append(assetIdRule, assetIdItem)
	}

	logs, sub, err := _SelfServeStorkContract.contract.FilterLogs(opts, "HistoricalValueStored", pubKeyRule, assetIdRule)
	if err != nil {
		return nil, err
	}
	return &SelfServeStorkContractHistoricalValueStoredIterator{contract: _SelfServeStorkContract.contract, event: "HistoricalValueStored", logs: logs, sub: sub}, nil
}

// WatchHistoricalValueStored is a free log subscription operation binding the contract event 0xfbb8f1bb7c4b5b8719a0357ba08c979650b113503c79ae1415088c0d6e429a8c.
//
// Solidity: event HistoricalValueStored(address indexed pubKey, string indexed assetId, uint64 timestampNs, int192 quantizedValue, uint256 roundId)
func (_SelfServeStorkContract *SelfServeStorkContractFilterer) WatchHistoricalValueStored(opts *bind.WatchOpts, sink chan<- *SelfServeStorkContractHistoricalValueStored, pubKey []common.Address, assetId []string) (event.Subscription, error) {

	var pubKeyRule []interface{}
	for _, pubKeyItem := range pubKey {
		pubKeyRule = append(pubKeyRule, pubKeyItem)
	}
	var assetIdRule []interface{}
	for _, assetIdItem := range assetId {
		assetIdRule = append(assetIdRule, assetIdItem)
	}

	logs, sub, err := _SelfServeStorkContract.contract.WatchLogs(opts, "HistoricalValueStored", pubKeyRule, assetIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SelfServeStorkContractHistoricalValueStored)
				if err := _SelfServeStorkContract.contract.UnpackLog(event, "HistoricalValueStored", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseHistoricalValueStored is a log parse operation binding the contract event 0xfbb8f1bb7c4b5b8719a0357ba08c979650b113503c79ae1415088c0d6e429a8c.
//
// Solidity: event HistoricalValueStored(address indexed pubKey, string indexed assetId, uint64 timestampNs, int192 quantizedValue, uint256 roundId)
func (_SelfServeStorkContract *SelfServeStorkContractFilterer) ParseHistoricalValueStored(log types.Log) (*SelfServeStorkContractHistoricalValueStored, error) {
	event := new(SelfServeStorkContractHistoricalValueStored)
	if err := _SelfServeStorkContract.contract.UnpackLog(event, "HistoricalValueStored", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SelfServeStorkContractPublisherUserAddedIterator is returned from FilterPublisherUserAdded and is used to iterate over the raw logs and unpacked data for PublisherUserAdded events raised by the SelfServeStorkContract contract.
type SelfServeStorkContractPublisherUserAddedIterator struct {
	Event *SelfServeStorkContractPublisherUserAdded // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *SelfServeStorkContractPublisherUserAddedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SelfServeStorkContractPublisherUserAdded)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(SelfServeStorkContractPublisherUserAdded)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *SelfServeStorkContractPublisherUserAddedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SelfServeStorkContractPublisherUserAddedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SelfServeStorkContractPublisherUserAdded represents a PublisherUserAdded event raised by the SelfServeStorkContract contract.
type SelfServeStorkContractPublisherUserAdded struct {
	PubKey          common.Address
	SingleUpdateFee *big.Int
	Raw             types.Log // Blockchain specific contextual infos
}

// FilterPublisherUserAdded is a free log retrieval operation binding the contract event 0x99d8764b5ae5766702a4ec20b7a7c9afdb8827250795c8937679c91f91696ffd.
//
// Solidity: event PublisherUserAdded(address indexed pubKey, uint256 singleUpdateFee)
func (_SelfServeStorkContract *SelfServeStorkContractFilterer) FilterPublisherUserAdded(opts *bind.FilterOpts, pubKey []common.Address) (*SelfServeStorkContractPublisherUserAddedIterator, error) {

	var pubKeyRule []interface{}
	for _, pubKeyItem := range pubKey {
		pubKeyRule = append(pubKeyRule, pubKeyItem)
	}

	logs, sub, err := _SelfServeStorkContract.contract.FilterLogs(opts, "PublisherUserAdded", pubKeyRule)
	if err != nil {
		return nil, err
	}
	return &SelfServeStorkContractPublisherUserAddedIterator{contract: _SelfServeStorkContract.contract, event: "PublisherUserAdded", logs: logs, sub: sub}, nil
}

// WatchPublisherUserAdded is a free log subscription operation binding the contract event 0x99d8764b5ae5766702a4ec20b7a7c9afdb8827250795c8937679c91f91696ffd.
//
// Solidity: event PublisherUserAdded(address indexed pubKey, uint256 singleUpdateFee)
func (_SelfServeStorkContract *SelfServeStorkContractFilterer) WatchPublisherUserAdded(opts *bind.WatchOpts, sink chan<- *SelfServeStorkContractPublisherUserAdded, pubKey []common.Address) (event.Subscription, error) {

	var pubKeyRule []interface{}
	for _, pubKeyItem := range pubKey {
		pubKeyRule = append(pubKeyRule, pubKeyItem)
	}

	logs, sub, err := _SelfServeStorkContract.contract.WatchLogs(opts, "PublisherUserAdded", pubKeyRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SelfServeStorkContractPublisherUserAdded)
				if err := _SelfServeStorkContract.contract.UnpackLog(event, "PublisherUserAdded", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParsePublisherUserAdded is a log parse operation binding the contract event 0x99d8764b5ae5766702a4ec20b7a7c9afdb8827250795c8937679c91f91696ffd.
//
// Solidity: event PublisherUserAdded(address indexed pubKey, uint256 singleUpdateFee)
func (_SelfServeStorkContract *SelfServeStorkContractFilterer) ParsePublisherUserAdded(log types.Log) (*SelfServeStorkContractPublisherUserAdded, error) {
	event := new(SelfServeStorkContractPublisherUserAdded)
	if err := _SelfServeStorkContract.contract.UnpackLog(event, "PublisherUserAdded", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SelfServeStorkContractPublisherUserRemovedIterator is returned from FilterPublisherUserRemoved and is used to iterate over the raw logs and unpacked data for PublisherUserRemoved events raised by the SelfServeStorkContract contract.
type SelfServeStorkContractPublisherUserRemovedIterator struct {
	Event *SelfServeStorkContractPublisherUserRemoved // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *SelfServeStorkContractPublisherUserRemovedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SelfServeStorkContractPublisherUserRemoved)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(SelfServeStorkContractPublisherUserRemoved)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *SelfServeStorkContractPublisherUserRemovedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SelfServeStorkContractPublisherUserRemovedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SelfServeStorkContractPublisherUserRemoved represents a PublisherUserRemoved event raised by the SelfServeStorkContract contract.
type SelfServeStorkContractPublisherUserRemoved struct {
	PubKey common.Address
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterPublisherUserRemoved is a free log retrieval operation binding the contract event 0xde37507a2967feaf95649773534bddc03d153af2b26b9ea17cdb9b62c5bbf18e.
//
// Solidity: event PublisherUserRemoved(address indexed pubKey)
func (_SelfServeStorkContract *SelfServeStorkContractFilterer) FilterPublisherUserRemoved(opts *bind.FilterOpts, pubKey []common.Address) (*SelfServeStorkContractPublisherUserRemovedIterator, error) {

	var pubKeyRule []interface{}
	for _, pubKeyItem := range pubKey {
		pubKeyRule = append(pubKeyRule, pubKeyItem)
	}

	logs, sub, err := _SelfServeStorkContract.contract.FilterLogs(opts, "PublisherUserRemoved", pubKeyRule)
	if err != nil {
		return nil, err
	}
	return &SelfServeStorkContractPublisherUserRemovedIterator{contract: _SelfServeStorkContract.contract, event: "PublisherUserRemoved", logs: logs, sub: sub}, nil
}

// WatchPublisherUserRemoved is a free log subscription operation binding the contract event 0xde37507a2967feaf95649773534bddc03d153af2b26b9ea17cdb9b62c5bbf18e.
//
// Solidity: event PublisherUserRemoved(address indexed pubKey)
func (_SelfServeStorkContract *SelfServeStorkContractFilterer) WatchPublisherUserRemoved(opts *bind.WatchOpts, sink chan<- *SelfServeStorkContractPublisherUserRemoved, pubKey []common.Address) (event.Subscription, error) {

	var pubKeyRule []interface{}
	for _, pubKeyItem := range pubKey {
		pubKeyRule = append(pubKeyRule, pubKeyItem)
	}

	logs, sub, err := _SelfServeStorkContract.contract.WatchLogs(opts, "PublisherUserRemoved", pubKeyRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SelfServeStorkContractPublisherUserRemoved)
				if err := _SelfServeStorkContract.contract.UnpackLog(event, "PublisherUserRemoved", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParsePublisherUserRemoved is a log parse operation binding the contract event 0xde37507a2967feaf95649773534bddc03d153af2b26b9ea17cdb9b62c5bbf18e.
//
// Solidity: event PublisherUserRemoved(address indexed pubKey)
func (_SelfServeStorkContract *SelfServeStorkContractFilterer) ParsePublisherUserRemoved(log types.Log) (*SelfServeStorkContractPublisherUserRemoved, error) {
	event := new(SelfServeStorkContractPublisherUserRemoved)
	if err := _SelfServeStorkContract.contract.UnpackLog(event, "PublisherUserRemoved", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SelfServeStorkContractValueUpdateIterator is returned from FilterValueUpdate and is used to iterate over the raw logs and unpacked data for ValueUpdate events raised by the SelfServeStorkContract contract.
type SelfServeStorkContractValueUpdateIterator struct {
	Event *SelfServeStorkContractValueUpdate // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *SelfServeStorkContractValueUpdateIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SelfServeStorkContractValueUpdate)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(SelfServeStorkContractValueUpdate)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *SelfServeStorkContractValueUpdateIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SelfServeStorkContractValueUpdateIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SelfServeStorkContractValueUpdate represents a ValueUpdate event raised by the SelfServeStorkContract contract.
type SelfServeStorkContractValueUpdate struct {
	PubKey         common.Address
	AssetId        common.Hash
	TimestampNs    uint64
	QuantizedValue *big.Int
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterValueUpdate is a free log retrieval operation binding the contract event 0x0596010914c581c18e615a5b0d097d78728eb775f2412d8bedd4dbc808f4f855.
//
// Solidity: event ValueUpdate(address indexed pubKey, string indexed assetId, uint64 timestampNs, int192 quantizedValue)
func (_SelfServeStorkContract *SelfServeStorkContractFilterer) FilterValueUpdate(opts *bind.FilterOpts, pubKey []common.Address, assetId []string) (*SelfServeStorkContractValueUpdateIterator, error) {

	var pubKeyRule []interface{}
	for _, pubKeyItem := range pubKey {
		pubKeyRule = append(pubKeyRule, pubKeyItem)
	}
	var assetIdRule []interface{}
	for _, assetIdItem := range assetId {
		assetIdRule = append(assetIdRule, assetIdItem)
	}

	logs, sub, err := _SelfServeStorkContract.contract.FilterLogs(opts, "ValueUpdate", pubKeyRule, assetIdRule)
	if err != nil {
		return nil, err
	}
	return &SelfServeStorkContractValueUpdateIterator{contract: _SelfServeStorkContract.contract, event: "ValueUpdate", logs: logs, sub: sub}, nil
}

// WatchValueUpdate is a free log subscription operation binding the contract event 0x0596010914c581c18e615a5b0d097d78728eb775f2412d8bedd4dbc808f4f855.
//
// Solidity: event ValueUpdate(address indexed pubKey, string indexed assetId, uint64 timestampNs, int192 quantizedValue)
func (_SelfServeStorkContract *SelfServeStorkContractFilterer) WatchValueUpdate(opts *bind.WatchOpts, sink chan<- *SelfServeStorkContractValueUpdate, pubKey []common.Address, assetId []string) (event.Subscription, error) {

	var pubKeyRule []interface{}
	for _, pubKeyItem := range pubKey {
		pubKeyRule = append(pubKeyRule, pubKeyItem)
	}
	var assetIdRule []interface{}
	for _, assetIdItem := range assetId {
		assetIdRule = append(assetIdRule, assetIdItem)
	}

	logs, sub, err := _SelfServeStorkContract.contract.WatchLogs(opts, "ValueUpdate", pubKeyRule, assetIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SelfServeStorkContractValueUpdate)
				if err := _SelfServeStorkContract.contract.UnpackLog(event, "ValueUpdate", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseValueUpdate is a log parse operation binding the contract event 0x0596010914c581c18e615a5b0d097d78728eb775f2412d8bedd4dbc808f4f855.
//
// Solidity: event ValueUpdate(address indexed pubKey, string indexed assetId, uint64 timestampNs, int192 quantizedValue)
func (_SelfServeStorkContract *SelfServeStorkContractFilterer) ParseValueUpdate(log types.Log) (*SelfServeStorkContractValueUpdate, error) {
	event := new(SelfServeStorkContractValueUpdate)
	if err := _SelfServeStorkContract.contract.UnpackLog(event, "ValueUpdate", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
