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

// FirstPartyStorkStructsPublisherTemporalNumericValueInput is an auto generated low-level Go binding around an user-defined struct.
type FirstPartyStorkStructsPublisherTemporalNumericValueInput struct {
	TemporalNumericValue FirstPartyStorkStructsTemporalNumericValue
	PubKey               common.Address
	AssetPairId          string
	StoreHistorical      bool
	R                    [32]byte
	S                    [32]byte
	V                    uint8
}

// FirstPartyStorkStructsPublisherUser is an auto generated low-level Go binding around an user-defined struct.
type FirstPartyStorkStructsPublisherUser struct {
	PubKey          common.Address
	SingleUpdateFee *big.Int
}

// FirstPartyStorkStructsTemporalNumericValue is an auto generated low-level Go binding around an user-defined struct.
type FirstPartyStorkStructsTemporalNumericValue struct {
	TimestampNs    uint64
	QuantizedValue *big.Int
}

// FirstPartyStorkContractMetaData contains all meta data concerning the FirstPartyStorkContract contract.
var FirstPartyStorkContractMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"name\":\"ECDSAInvalidSignature\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"length\",\"type\":\"uint256\"}],\"name\":\"ECDSAInvalidSignatureLength\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"s\",\"type\":\"bytes32\"}],\"name\":\"ECDSAInvalidSignatureS\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InsufficientFee\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidSignature\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NoFreshUpdate\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotFound\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"pubKey\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"string\",\"name\":\"assetId\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"timestampNs\",\"type\":\"uint64\"},{\"indexed\":false,\"internalType\":\"int192\",\"name\":\"quantizedValue\",\"type\":\"int192\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"roundId\",\"type\":\"uint256\"}],\"name\":\"HistoricalValueStored\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"pubKey\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"singleUpdateFee\",\"type\":\"uint256\"}],\"name\":\"PublisherUserAdded\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"pubKey\",\"type\":\"address\"}],\"name\":\"PublisherUserRemoved\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"pubKey\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"string\",\"name\":\"assetId\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"timestampNs\",\"type\":\"uint64\"},{\"indexed\":false,\"internalType\":\"int192\",\"name\":\"quantizedValue\",\"type\":\"int192\"}],\"name\":\"ValueUpdate\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"pubKey\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"singleUpdateFee\",\"type\":\"uint256\"}],\"name\":\"createPublisherUser\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"pubKey\",\"type\":\"address\"}],\"name\":\"deletePublisherUser\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"pubKey\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"assetPairId\",\"type\":\"string\"}],\"name\":\"getCurrentRoundId\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"pubKey\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"assetPairId\",\"type\":\"string\"}],\"name\":\"getHistoricalRecordsCount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"pubKey\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"assetPairId\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"roundId\",\"type\":\"uint256\"}],\"name\":\"getHistoricalTemporalNumericValue\",\"outputs\":[{\"components\":[{\"internalType\":\"uint64\",\"name\":\"timestampNs\",\"type\":\"uint64\"},{\"internalType\":\"int192\",\"name\":\"quantizedValue\",\"type\":\"int192\"}],\"internalType\":\"structFirstPartyStorkStructs.TemporalNumericValue\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"pubKey\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"assetPairId\",\"type\":\"string\"}],\"name\":\"getLatestTemporalNumericValue\",\"outputs\":[{\"components\":[{\"internalType\":\"uint64\",\"name\":\"timestampNs\",\"type\":\"uint64\"},{\"internalType\":\"int192\",\"name\":\"quantizedValue\",\"type\":\"int192\"}],\"internalType\":\"structFirstPartyStorkStructs.TemporalNumericValue\",\"name\":\"value\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"pubKey\",\"type\":\"address\"}],\"name\":\"getPublisherUser\",\"outputs\":[{\"components\":[{\"internalType\":\"address\",\"name\":\"pubKey\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"singleUpdateFee\",\"type\":\"uint256\"}],\"internalType\":\"structFirstPartyStorkStructs.PublisherUser\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"pubKey\",\"type\":\"address\"}],\"name\":\"getSingleUpdateFee\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"pubKey\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"totalNumUpdates\",\"type\":\"uint256\"}],\"name\":\"getUpdateFeeV1\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"components\":[{\"internalType\":\"uint64\",\"name\":\"timestampNs\",\"type\":\"uint64\"},{\"internalType\":\"int192\",\"name\":\"quantizedValue\",\"type\":\"int192\"}],\"internalType\":\"structFirstPartyStorkStructs.TemporalNumericValue\",\"name\":\"temporalNumericValue\",\"type\":\"tuple\"},{\"internalType\":\"address\",\"name\":\"pubKey\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"assetPairId\",\"type\":\"string\"},{\"internalType\":\"bool\",\"name\":\"storeHistorical\",\"type\":\"bool\"},{\"internalType\":\"bytes32\",\"name\":\"r\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"s\",\"type\":\"bytes32\"},{\"internalType\":\"uint8\",\"name\":\"v\",\"type\":\"uint8\"}],\"internalType\":\"structFirstPartyStorkStructs.PublisherTemporalNumericValueInput[]\",\"name\":\"updateData\",\"type\":\"tuple[]\"}],\"name\":\"updateTemporalNumericValues\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"publisherPubKey\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"assetPairId\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"},{\"internalType\":\"int256\",\"name\":\"value\",\"type\":\"int256\"},{\"internalType\":\"bytes32\",\"name\":\"r\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"s\",\"type\":\"bytes32\"},{\"internalType\":\"uint8\",\"name\":\"v\",\"type\":\"uint8\"}],\"name\":\"verifyPublisherSignatureV1\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"pure\",\"type\":\"function\"}]",
}

// FirstPartyStorkContractABI is the input ABI used to generate the binding from.
// Deprecated: Use FirstPartyStorkContractMetaData.ABI instead.
var FirstPartyStorkContractABI = FirstPartyStorkContractMetaData.ABI

// FirstPartyStorkContract is an auto generated Go binding around an Ethereum contract.
type FirstPartyStorkContract struct {
	FirstPartyStorkContractCaller     // Read-only binding to the contract
	FirstPartyStorkContractTransactor // Write-only binding to the contract
	FirstPartyStorkContractFilterer   // Log filterer for contract events
}

// FirstPartyStorkContractCaller is an auto generated read-only Go binding around an Ethereum contract.
type FirstPartyStorkContractCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// FirstPartyStorkContractTransactor is an auto generated write-only Go binding around an Ethereum contract.
type FirstPartyStorkContractTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// FirstPartyStorkContractFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type FirstPartyStorkContractFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// FirstPartyStorkContractSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type FirstPartyStorkContractSession struct {
	Contract     *FirstPartyStorkContract // Generic contract binding to set the session for
	CallOpts     bind.CallOpts            // Call options to use throughout this session
	TransactOpts bind.TransactOpts        // Transaction auth options to use throughout this session
}

// FirstPartyStorkContractCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type FirstPartyStorkContractCallerSession struct {
	Contract *FirstPartyStorkContractCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts                  // Call options to use throughout this session
}

// FirstPartyStorkContractTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type FirstPartyStorkContractTransactorSession struct {
	Contract     *FirstPartyStorkContractTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts                  // Transaction auth options to use throughout this session
}

// FirstPartyStorkContractRaw is an auto generated low-level Go binding around an Ethereum contract.
type FirstPartyStorkContractRaw struct {
	Contract *FirstPartyStorkContract // Generic contract binding to access the raw methods on
}

// FirstPartyStorkContractCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type FirstPartyStorkContractCallerRaw struct {
	Contract *FirstPartyStorkContractCaller // Generic read-only contract binding to access the raw methods on
}

// FirstPartyStorkContractTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type FirstPartyStorkContractTransactorRaw struct {
	Contract *FirstPartyStorkContractTransactor // Generic write-only contract binding to access the raw methods on
}

// NewFirstPartyStorkContract creates a new instance of FirstPartyStorkContract, bound to a specific deployed contract.
func NewFirstPartyStorkContract(address common.Address, backend bind.ContractBackend) (*FirstPartyStorkContract, error) {
	contract, err := bindFirstPartyStorkContract(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &FirstPartyStorkContract{FirstPartyStorkContractCaller: FirstPartyStorkContractCaller{contract: contract}, FirstPartyStorkContractTransactor: FirstPartyStorkContractTransactor{contract: contract}, FirstPartyStorkContractFilterer: FirstPartyStorkContractFilterer{contract: contract}}, nil
}

// NewFirstPartyStorkContractCaller creates a new read-only instance of FirstPartyStorkContract, bound to a specific deployed contract.
func NewFirstPartyStorkContractCaller(address common.Address, caller bind.ContractCaller) (*FirstPartyStorkContractCaller, error) {
	contract, err := bindFirstPartyStorkContract(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &FirstPartyStorkContractCaller{contract: contract}, nil
}

// NewFirstPartyStorkContractTransactor creates a new write-only instance of FirstPartyStorkContract, bound to a specific deployed contract.
func NewFirstPartyStorkContractTransactor(address common.Address, transactor bind.ContractTransactor) (*FirstPartyStorkContractTransactor, error) {
	contract, err := bindFirstPartyStorkContract(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &FirstPartyStorkContractTransactor{contract: contract}, nil
}

// NewFirstPartyStorkContractFilterer creates a new log filterer instance of FirstPartyStorkContract, bound to a specific deployed contract.
func NewFirstPartyStorkContractFilterer(address common.Address, filterer bind.ContractFilterer) (*FirstPartyStorkContractFilterer, error) {
	contract, err := bindFirstPartyStorkContract(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &FirstPartyStorkContractFilterer{contract: contract}, nil
}

// bindFirstPartyStorkContract binds a generic wrapper to an already deployed contract.
func bindFirstPartyStorkContract(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := FirstPartyStorkContractMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_FirstPartyStorkContract *FirstPartyStorkContractRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _FirstPartyStorkContract.Contract.FirstPartyStorkContractCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_FirstPartyStorkContract *FirstPartyStorkContractRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _FirstPartyStorkContract.Contract.FirstPartyStorkContractTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_FirstPartyStorkContract *FirstPartyStorkContractRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _FirstPartyStorkContract.Contract.FirstPartyStorkContractTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_FirstPartyStorkContract *FirstPartyStorkContractCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _FirstPartyStorkContract.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_FirstPartyStorkContract *FirstPartyStorkContractTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _FirstPartyStorkContract.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_FirstPartyStorkContract *FirstPartyStorkContractTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _FirstPartyStorkContract.Contract.contract.Transact(opts, method, params...)
}

// GetCurrentRoundId is a free data retrieval call binding the contract method 0xb7f16567.
//
// Solidity: function getCurrentRoundId(address pubKey, string assetPairId) view returns(uint256)
func (_FirstPartyStorkContract *FirstPartyStorkContractCaller) GetCurrentRoundId(opts *bind.CallOpts, pubKey common.Address, assetPairId string) (*big.Int, error) {
	var out []interface{}
	err := _FirstPartyStorkContract.contract.Call(opts, &out, "getCurrentRoundId", pubKey, assetPairId)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetCurrentRoundId is a free data retrieval call binding the contract method 0xb7f16567.
//
// Solidity: function getCurrentRoundId(address pubKey, string assetPairId) view returns(uint256)
func (_FirstPartyStorkContract *FirstPartyStorkContractSession) GetCurrentRoundId(pubKey common.Address, assetPairId string) (*big.Int, error) {
	return _FirstPartyStorkContract.Contract.GetCurrentRoundId(&_FirstPartyStorkContract.CallOpts, pubKey, assetPairId)
}

// GetCurrentRoundId is a free data retrieval call binding the contract method 0xb7f16567.
//
// Solidity: function getCurrentRoundId(address pubKey, string assetPairId) view returns(uint256)
func (_FirstPartyStorkContract *FirstPartyStorkContractCallerSession) GetCurrentRoundId(pubKey common.Address, assetPairId string) (*big.Int, error) {
	return _FirstPartyStorkContract.Contract.GetCurrentRoundId(&_FirstPartyStorkContract.CallOpts, pubKey, assetPairId)
}

// GetHistoricalRecordsCount is a free data retrieval call binding the contract method 0x543b2fc1.
//
// Solidity: function getHistoricalRecordsCount(address pubKey, string assetPairId) view returns(uint256)
func (_FirstPartyStorkContract *FirstPartyStorkContractCaller) GetHistoricalRecordsCount(opts *bind.CallOpts, pubKey common.Address, assetPairId string) (*big.Int, error) {
	var out []interface{}
	err := _FirstPartyStorkContract.contract.Call(opts, &out, "getHistoricalRecordsCount", pubKey, assetPairId)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetHistoricalRecordsCount is a free data retrieval call binding the contract method 0x543b2fc1.
//
// Solidity: function getHistoricalRecordsCount(address pubKey, string assetPairId) view returns(uint256)
func (_FirstPartyStorkContract *FirstPartyStorkContractSession) GetHistoricalRecordsCount(pubKey common.Address, assetPairId string) (*big.Int, error) {
	return _FirstPartyStorkContract.Contract.GetHistoricalRecordsCount(&_FirstPartyStorkContract.CallOpts, pubKey, assetPairId)
}

// GetHistoricalRecordsCount is a free data retrieval call binding the contract method 0x543b2fc1.
//
// Solidity: function getHistoricalRecordsCount(address pubKey, string assetPairId) view returns(uint256)
func (_FirstPartyStorkContract *FirstPartyStorkContractCallerSession) GetHistoricalRecordsCount(pubKey common.Address, assetPairId string) (*big.Int, error) {
	return _FirstPartyStorkContract.Contract.GetHistoricalRecordsCount(&_FirstPartyStorkContract.CallOpts, pubKey, assetPairId)
}

// GetHistoricalTemporalNumericValue is a free data retrieval call binding the contract method 0x2651ad06.
//
// Solidity: function getHistoricalTemporalNumericValue(address pubKey, string assetPairId, uint256 roundId) view returns((uint64,int192))
func (_FirstPartyStorkContract *FirstPartyStorkContractCaller) GetHistoricalTemporalNumericValue(opts *bind.CallOpts, pubKey common.Address, assetPairId string, roundId *big.Int) (FirstPartyStorkStructsTemporalNumericValue, error) {
	var out []interface{}
	err := _FirstPartyStorkContract.contract.Call(opts, &out, "getHistoricalTemporalNumericValue", pubKey, assetPairId, roundId)

	if err != nil {
		return *new(FirstPartyStorkStructsTemporalNumericValue), err
	}

	out0 := *abi.ConvertType(out[0], new(FirstPartyStorkStructsTemporalNumericValue)).(*FirstPartyStorkStructsTemporalNumericValue)

	return out0, err

}

// GetHistoricalTemporalNumericValue is a free data retrieval call binding the contract method 0x2651ad06.
//
// Solidity: function getHistoricalTemporalNumericValue(address pubKey, string assetPairId, uint256 roundId) view returns((uint64,int192))
func (_FirstPartyStorkContract *FirstPartyStorkContractSession) GetHistoricalTemporalNumericValue(pubKey common.Address, assetPairId string, roundId *big.Int) (FirstPartyStorkStructsTemporalNumericValue, error) {
	return _FirstPartyStorkContract.Contract.GetHistoricalTemporalNumericValue(&_FirstPartyStorkContract.CallOpts, pubKey, assetPairId, roundId)
}

// GetHistoricalTemporalNumericValue is a free data retrieval call binding the contract method 0x2651ad06.
//
// Solidity: function getHistoricalTemporalNumericValue(address pubKey, string assetPairId, uint256 roundId) view returns((uint64,int192))
func (_FirstPartyStorkContract *FirstPartyStorkContractCallerSession) GetHistoricalTemporalNumericValue(pubKey common.Address, assetPairId string, roundId *big.Int) (FirstPartyStorkStructsTemporalNumericValue, error) {
	return _FirstPartyStorkContract.Contract.GetHistoricalTemporalNumericValue(&_FirstPartyStorkContract.CallOpts, pubKey, assetPairId, roundId)
}

// GetLatestTemporalNumericValue is a free data retrieval call binding the contract method 0xea419887.
//
// Solidity: function getLatestTemporalNumericValue(address pubKey, string assetPairId) view returns((uint64,int192) value)
func (_FirstPartyStorkContract *FirstPartyStorkContractCaller) GetLatestTemporalNumericValue(opts *bind.CallOpts, pubKey common.Address, assetPairId string) (FirstPartyStorkStructsTemporalNumericValue, error) {
	var out []interface{}
	err := _FirstPartyStorkContract.contract.Call(opts, &out, "getLatestTemporalNumericValue", pubKey, assetPairId)

	if err != nil {
		return *new(FirstPartyStorkStructsTemporalNumericValue), err
	}

	out0 := *abi.ConvertType(out[0], new(FirstPartyStorkStructsTemporalNumericValue)).(*FirstPartyStorkStructsTemporalNumericValue)

	return out0, err

}

// GetLatestTemporalNumericValue is a free data retrieval call binding the contract method 0xea419887.
//
// Solidity: function getLatestTemporalNumericValue(address pubKey, string assetPairId) view returns((uint64,int192) value)
func (_FirstPartyStorkContract *FirstPartyStorkContractSession) GetLatestTemporalNumericValue(pubKey common.Address, assetPairId string) (FirstPartyStorkStructsTemporalNumericValue, error) {
	return _FirstPartyStorkContract.Contract.GetLatestTemporalNumericValue(&_FirstPartyStorkContract.CallOpts, pubKey, assetPairId)
}

// GetLatestTemporalNumericValue is a free data retrieval call binding the contract method 0xea419887.
//
// Solidity: function getLatestTemporalNumericValue(address pubKey, string assetPairId) view returns((uint64,int192) value)
func (_FirstPartyStorkContract *FirstPartyStorkContractCallerSession) GetLatestTemporalNumericValue(pubKey common.Address, assetPairId string) (FirstPartyStorkStructsTemporalNumericValue, error) {
	return _FirstPartyStorkContract.Contract.GetLatestTemporalNumericValue(&_FirstPartyStorkContract.CallOpts, pubKey, assetPairId)
}

// GetPublisherUser is a free data retrieval call binding the contract method 0x3d57a294.
//
// Solidity: function getPublisherUser(address pubKey) view returns((address,uint256))
func (_FirstPartyStorkContract *FirstPartyStorkContractCaller) GetPublisherUser(opts *bind.CallOpts, pubKey common.Address) (FirstPartyStorkStructsPublisherUser, error) {
	var out []interface{}
	err := _FirstPartyStorkContract.contract.Call(opts, &out, "getPublisherUser", pubKey)

	if err != nil {
		return *new(FirstPartyStorkStructsPublisherUser), err
	}

	out0 := *abi.ConvertType(out[0], new(FirstPartyStorkStructsPublisherUser)).(*FirstPartyStorkStructsPublisherUser)

	return out0, err

}

// GetPublisherUser is a free data retrieval call binding the contract method 0x3d57a294.
//
// Solidity: function getPublisherUser(address pubKey) view returns((address,uint256))
func (_FirstPartyStorkContract *FirstPartyStorkContractSession) GetPublisherUser(pubKey common.Address) (FirstPartyStorkStructsPublisherUser, error) {
	return _FirstPartyStorkContract.Contract.GetPublisherUser(&_FirstPartyStorkContract.CallOpts, pubKey)
}

// GetPublisherUser is a free data retrieval call binding the contract method 0x3d57a294.
//
// Solidity: function getPublisherUser(address pubKey) view returns((address,uint256))
func (_FirstPartyStorkContract *FirstPartyStorkContractCallerSession) GetPublisherUser(pubKey common.Address) (FirstPartyStorkStructsPublisherUser, error) {
	return _FirstPartyStorkContract.Contract.GetPublisherUser(&_FirstPartyStorkContract.CallOpts, pubKey)
}

// GetSingleUpdateFee is a free data retrieval call binding the contract method 0x44bae290.
//
// Solidity: function getSingleUpdateFee(address pubKey) view returns(uint256)
func (_FirstPartyStorkContract *FirstPartyStorkContractCaller) GetSingleUpdateFee(opts *bind.CallOpts, pubKey common.Address) (*big.Int, error) {
	var out []interface{}
	err := _FirstPartyStorkContract.contract.Call(opts, &out, "getSingleUpdateFee", pubKey)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetSingleUpdateFee is a free data retrieval call binding the contract method 0x44bae290.
//
// Solidity: function getSingleUpdateFee(address pubKey) view returns(uint256)
func (_FirstPartyStorkContract *FirstPartyStorkContractSession) GetSingleUpdateFee(pubKey common.Address) (*big.Int, error) {
	return _FirstPartyStorkContract.Contract.GetSingleUpdateFee(&_FirstPartyStorkContract.CallOpts, pubKey)
}

// GetSingleUpdateFee is a free data retrieval call binding the contract method 0x44bae290.
//
// Solidity: function getSingleUpdateFee(address pubKey) view returns(uint256)
func (_FirstPartyStorkContract *FirstPartyStorkContractCallerSession) GetSingleUpdateFee(pubKey common.Address) (*big.Int, error) {
	return _FirstPartyStorkContract.Contract.GetSingleUpdateFee(&_FirstPartyStorkContract.CallOpts, pubKey)
}

// GetUpdateFeeV1 is a free data retrieval call binding the contract method 0xe11ddbc3.
//
// Solidity: function getUpdateFeeV1(address pubKey, uint256 totalNumUpdates) view returns(uint256)
func (_FirstPartyStorkContract *FirstPartyStorkContractCaller) GetUpdateFeeV1(opts *bind.CallOpts, pubKey common.Address, totalNumUpdates *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _FirstPartyStorkContract.contract.Call(opts, &out, "getUpdateFeeV1", pubKey, totalNumUpdates)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetUpdateFeeV1 is a free data retrieval call binding the contract method 0xe11ddbc3.
//
// Solidity: function getUpdateFeeV1(address pubKey, uint256 totalNumUpdates) view returns(uint256)
func (_FirstPartyStorkContract *FirstPartyStorkContractSession) GetUpdateFeeV1(pubKey common.Address, totalNumUpdates *big.Int) (*big.Int, error) {
	return _FirstPartyStorkContract.Contract.GetUpdateFeeV1(&_FirstPartyStorkContract.CallOpts, pubKey, totalNumUpdates)
}

// GetUpdateFeeV1 is a free data retrieval call binding the contract method 0xe11ddbc3.
//
// Solidity: function getUpdateFeeV1(address pubKey, uint256 totalNumUpdates) view returns(uint256)
func (_FirstPartyStorkContract *FirstPartyStorkContractCallerSession) GetUpdateFeeV1(pubKey common.Address, totalNumUpdates *big.Int) (*big.Int, error) {
	return _FirstPartyStorkContract.Contract.GetUpdateFeeV1(&_FirstPartyStorkContract.CallOpts, pubKey, totalNumUpdates)
}

// VerifyPublisherSignatureV1 is a free data retrieval call binding the contract method 0x9bccd2d5.
//
// Solidity: function verifyPublisherSignatureV1(address publisherPubKey, string assetPairId, uint256 timestamp, int256 value, bytes32 r, bytes32 s, uint8 v) pure returns(bool)
func (_FirstPartyStorkContract *FirstPartyStorkContractCaller) VerifyPublisherSignatureV1(opts *bind.CallOpts, publisherPubKey common.Address, assetPairId string, timestamp *big.Int, value *big.Int, r [32]byte, s [32]byte, v uint8) (bool, error) {
	var out []interface{}
	err := _FirstPartyStorkContract.contract.Call(opts, &out, "verifyPublisherSignatureV1", publisherPubKey, assetPairId, timestamp, value, r, s, v)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// VerifyPublisherSignatureV1 is a free data retrieval call binding the contract method 0x9bccd2d5.
//
// Solidity: function verifyPublisherSignatureV1(address publisherPubKey, string assetPairId, uint256 timestamp, int256 value, bytes32 r, bytes32 s, uint8 v) pure returns(bool)
func (_FirstPartyStorkContract *FirstPartyStorkContractSession) VerifyPublisherSignatureV1(publisherPubKey common.Address, assetPairId string, timestamp *big.Int, value *big.Int, r [32]byte, s [32]byte, v uint8) (bool, error) {
	return _FirstPartyStorkContract.Contract.VerifyPublisherSignatureV1(&_FirstPartyStorkContract.CallOpts, publisherPubKey, assetPairId, timestamp, value, r, s, v)
}

// VerifyPublisherSignatureV1 is a free data retrieval call binding the contract method 0x9bccd2d5.
//
// Solidity: function verifyPublisherSignatureV1(address publisherPubKey, string assetPairId, uint256 timestamp, int256 value, bytes32 r, bytes32 s, uint8 v) pure returns(bool)
func (_FirstPartyStorkContract *FirstPartyStorkContractCallerSession) VerifyPublisherSignatureV1(publisherPubKey common.Address, assetPairId string, timestamp *big.Int, value *big.Int, r [32]byte, s [32]byte, v uint8) (bool, error) {
	return _FirstPartyStorkContract.Contract.VerifyPublisherSignatureV1(&_FirstPartyStorkContract.CallOpts, publisherPubKey, assetPairId, timestamp, value, r, s, v)
}

// CreatePublisherUser is a paid mutator transaction binding the contract method 0x714fbcad.
//
// Solidity: function createPublisherUser(address pubKey, uint256 singleUpdateFee) returns()
func (_FirstPartyStorkContract *FirstPartyStorkContractTransactor) CreatePublisherUser(opts *bind.TransactOpts, pubKey common.Address, singleUpdateFee *big.Int) (*types.Transaction, error) {
	return _FirstPartyStorkContract.contract.Transact(opts, "createPublisherUser", pubKey, singleUpdateFee)
}

// CreatePublisherUser is a paid mutator transaction binding the contract method 0x714fbcad.
//
// Solidity: function createPublisherUser(address pubKey, uint256 singleUpdateFee) returns()
func (_FirstPartyStorkContract *FirstPartyStorkContractSession) CreatePublisherUser(pubKey common.Address, singleUpdateFee *big.Int) (*types.Transaction, error) {
	return _FirstPartyStorkContract.Contract.CreatePublisherUser(&_FirstPartyStorkContract.TransactOpts, pubKey, singleUpdateFee)
}

// CreatePublisherUser is a paid mutator transaction binding the contract method 0x714fbcad.
//
// Solidity: function createPublisherUser(address pubKey, uint256 singleUpdateFee) returns()
func (_FirstPartyStorkContract *FirstPartyStorkContractTransactorSession) CreatePublisherUser(pubKey common.Address, singleUpdateFee *big.Int) (*types.Transaction, error) {
	return _FirstPartyStorkContract.Contract.CreatePublisherUser(&_FirstPartyStorkContract.TransactOpts, pubKey, singleUpdateFee)
}

// DeletePublisherUser is a paid mutator transaction binding the contract method 0x488487a9.
//
// Solidity: function deletePublisherUser(address pubKey) returns()
func (_FirstPartyStorkContract *FirstPartyStorkContractTransactor) DeletePublisherUser(opts *bind.TransactOpts, pubKey common.Address) (*types.Transaction, error) {
	return _FirstPartyStorkContract.contract.Transact(opts, "deletePublisherUser", pubKey)
}

// DeletePublisherUser is a paid mutator transaction binding the contract method 0x488487a9.
//
// Solidity: function deletePublisherUser(address pubKey) returns()
func (_FirstPartyStorkContract *FirstPartyStorkContractSession) DeletePublisherUser(pubKey common.Address) (*types.Transaction, error) {
	return _FirstPartyStorkContract.Contract.DeletePublisherUser(&_FirstPartyStorkContract.TransactOpts, pubKey)
}

// DeletePublisherUser is a paid mutator transaction binding the contract method 0x488487a9.
//
// Solidity: function deletePublisherUser(address pubKey) returns()
func (_FirstPartyStorkContract *FirstPartyStorkContractTransactorSession) DeletePublisherUser(pubKey common.Address) (*types.Transaction, error) {
	return _FirstPartyStorkContract.Contract.DeletePublisherUser(&_FirstPartyStorkContract.TransactOpts, pubKey)
}

// UpdateTemporalNumericValues is a paid mutator transaction binding the contract method 0x38a3f02f.
//
// Solidity: function updateTemporalNumericValues(((uint64,int192),address,string,bool,bytes32,bytes32,uint8)[] updateData) payable returns()
func (_FirstPartyStorkContract *FirstPartyStorkContractTransactor) UpdateTemporalNumericValues(opts *bind.TransactOpts, updateData []FirstPartyStorkStructsPublisherTemporalNumericValueInput) (*types.Transaction, error) {
	return _FirstPartyStorkContract.contract.Transact(opts, "updateTemporalNumericValues", updateData)
}

// UpdateTemporalNumericValues is a paid mutator transaction binding the contract method 0x38a3f02f.
//
// Solidity: function updateTemporalNumericValues(((uint64,int192),address,string,bool,bytes32,bytes32,uint8)[] updateData) payable returns()
func (_FirstPartyStorkContract *FirstPartyStorkContractSession) UpdateTemporalNumericValues(updateData []FirstPartyStorkStructsPublisherTemporalNumericValueInput) (*types.Transaction, error) {
	return _FirstPartyStorkContract.Contract.UpdateTemporalNumericValues(&_FirstPartyStorkContract.TransactOpts, updateData)
}

// UpdateTemporalNumericValues is a paid mutator transaction binding the contract method 0x38a3f02f.
//
// Solidity: function updateTemporalNumericValues(((uint64,int192),address,string,bool,bytes32,bytes32,uint8)[] updateData) payable returns()
func (_FirstPartyStorkContract *FirstPartyStorkContractTransactorSession) UpdateTemporalNumericValues(updateData []FirstPartyStorkStructsPublisherTemporalNumericValueInput) (*types.Transaction, error) {
	return _FirstPartyStorkContract.Contract.UpdateTemporalNumericValues(&_FirstPartyStorkContract.TransactOpts, updateData)
}

// FirstPartyStorkContractHistoricalValueStoredIterator is returned from FilterHistoricalValueStored and is used to iterate over the raw logs and unpacked data for HistoricalValueStored events raised by the FirstPartyStorkContract contract.
type FirstPartyStorkContractHistoricalValueStoredIterator struct {
	Event *FirstPartyStorkContractHistoricalValueStored // Event containing the contract specifics and raw log

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
func (it *FirstPartyStorkContractHistoricalValueStoredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(FirstPartyStorkContractHistoricalValueStored)
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
		it.Event = new(FirstPartyStorkContractHistoricalValueStored)
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
func (it *FirstPartyStorkContractHistoricalValueStoredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *FirstPartyStorkContractHistoricalValueStoredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// FirstPartyStorkContractHistoricalValueStored represents a HistoricalValueStored event raised by the FirstPartyStorkContract contract.
type FirstPartyStorkContractHistoricalValueStored struct {
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
func (_FirstPartyStorkContract *FirstPartyStorkContractFilterer) FilterHistoricalValueStored(opts *bind.FilterOpts, pubKey []common.Address, assetId []string) (*FirstPartyStorkContractHistoricalValueStoredIterator, error) {

	var pubKeyRule []interface{}
	for _, pubKeyItem := range pubKey {
		pubKeyRule = append(pubKeyRule, pubKeyItem)
	}
	var assetIdRule []interface{}
	for _, assetIdItem := range assetId {
		assetIdRule = append(assetIdRule, assetIdItem)
	}

	logs, sub, err := _FirstPartyStorkContract.contract.FilterLogs(opts, "HistoricalValueStored", pubKeyRule, assetIdRule)
	if err != nil {
		return nil, err
	}
	return &FirstPartyStorkContractHistoricalValueStoredIterator{contract: _FirstPartyStorkContract.contract, event: "HistoricalValueStored", logs: logs, sub: sub}, nil
}

// WatchHistoricalValueStored is a free log subscription operation binding the contract event 0xfbb8f1bb7c4b5b8719a0357ba08c979650b113503c79ae1415088c0d6e429a8c.
//
// Solidity: event HistoricalValueStored(address indexed pubKey, string indexed assetId, uint64 timestampNs, int192 quantizedValue, uint256 roundId)
func (_FirstPartyStorkContract *FirstPartyStorkContractFilterer) WatchHistoricalValueStored(opts *bind.WatchOpts, sink chan<- *FirstPartyStorkContractHistoricalValueStored, pubKey []common.Address, assetId []string) (event.Subscription, error) {

	var pubKeyRule []interface{}
	for _, pubKeyItem := range pubKey {
		pubKeyRule = append(pubKeyRule, pubKeyItem)
	}
	var assetIdRule []interface{}
	for _, assetIdItem := range assetId {
		assetIdRule = append(assetIdRule, assetIdItem)
	}

	logs, sub, err := _FirstPartyStorkContract.contract.WatchLogs(opts, "HistoricalValueStored", pubKeyRule, assetIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(FirstPartyStorkContractHistoricalValueStored)
				if err := _FirstPartyStorkContract.contract.UnpackLog(event, "HistoricalValueStored", log); err != nil {
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
func (_FirstPartyStorkContract *FirstPartyStorkContractFilterer) ParseHistoricalValueStored(log types.Log) (*FirstPartyStorkContractHistoricalValueStored, error) {
	event := new(FirstPartyStorkContractHistoricalValueStored)
	if err := _FirstPartyStorkContract.contract.UnpackLog(event, "HistoricalValueStored", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// FirstPartyStorkContractPublisherUserAddedIterator is returned from FilterPublisherUserAdded and is used to iterate over the raw logs and unpacked data for PublisherUserAdded events raised by the FirstPartyStorkContract contract.
type FirstPartyStorkContractPublisherUserAddedIterator struct {
	Event *FirstPartyStorkContractPublisherUserAdded // Event containing the contract specifics and raw log

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
func (it *FirstPartyStorkContractPublisherUserAddedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(FirstPartyStorkContractPublisherUserAdded)
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
		it.Event = new(FirstPartyStorkContractPublisherUserAdded)
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
func (it *FirstPartyStorkContractPublisherUserAddedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *FirstPartyStorkContractPublisherUserAddedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// FirstPartyStorkContractPublisherUserAdded represents a PublisherUserAdded event raised by the FirstPartyStorkContract contract.
type FirstPartyStorkContractPublisherUserAdded struct {
	PubKey          common.Address
	SingleUpdateFee *big.Int
	Raw             types.Log // Blockchain specific contextual infos
}

// FilterPublisherUserAdded is a free log retrieval operation binding the contract event 0x99d8764b5ae5766702a4ec20b7a7c9afdb8827250795c8937679c91f91696ffd.
//
// Solidity: event PublisherUserAdded(address indexed pubKey, uint256 singleUpdateFee)
func (_FirstPartyStorkContract *FirstPartyStorkContractFilterer) FilterPublisherUserAdded(opts *bind.FilterOpts, pubKey []common.Address) (*FirstPartyStorkContractPublisherUserAddedIterator, error) {

	var pubKeyRule []interface{}
	for _, pubKeyItem := range pubKey {
		pubKeyRule = append(pubKeyRule, pubKeyItem)
	}

	logs, sub, err := _FirstPartyStorkContract.contract.FilterLogs(opts, "PublisherUserAdded", pubKeyRule)
	if err != nil {
		return nil, err
	}
	return &FirstPartyStorkContractPublisherUserAddedIterator{contract: _FirstPartyStorkContract.contract, event: "PublisherUserAdded", logs: logs, sub: sub}, nil
}

// WatchPublisherUserAdded is a free log subscription operation binding the contract event 0x99d8764b5ae5766702a4ec20b7a7c9afdb8827250795c8937679c91f91696ffd.
//
// Solidity: event PublisherUserAdded(address indexed pubKey, uint256 singleUpdateFee)
func (_FirstPartyStorkContract *FirstPartyStorkContractFilterer) WatchPublisherUserAdded(opts *bind.WatchOpts, sink chan<- *FirstPartyStorkContractPublisherUserAdded, pubKey []common.Address) (event.Subscription, error) {

	var pubKeyRule []interface{}
	for _, pubKeyItem := range pubKey {
		pubKeyRule = append(pubKeyRule, pubKeyItem)
	}

	logs, sub, err := _FirstPartyStorkContract.contract.WatchLogs(opts, "PublisherUserAdded", pubKeyRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(FirstPartyStorkContractPublisherUserAdded)
				if err := _FirstPartyStorkContract.contract.UnpackLog(event, "PublisherUserAdded", log); err != nil {
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
func (_FirstPartyStorkContract *FirstPartyStorkContractFilterer) ParsePublisherUserAdded(log types.Log) (*FirstPartyStorkContractPublisherUserAdded, error) {
	event := new(FirstPartyStorkContractPublisherUserAdded)
	if err := _FirstPartyStorkContract.contract.UnpackLog(event, "PublisherUserAdded", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// FirstPartyStorkContractPublisherUserRemovedIterator is returned from FilterPublisherUserRemoved and is used to iterate over the raw logs and unpacked data for PublisherUserRemoved events raised by the FirstPartyStorkContract contract.
type FirstPartyStorkContractPublisherUserRemovedIterator struct {
	Event *FirstPartyStorkContractPublisherUserRemoved // Event containing the contract specifics and raw log

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
func (it *FirstPartyStorkContractPublisherUserRemovedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(FirstPartyStorkContractPublisherUserRemoved)
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
		it.Event = new(FirstPartyStorkContractPublisherUserRemoved)
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
func (it *FirstPartyStorkContractPublisherUserRemovedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *FirstPartyStorkContractPublisherUserRemovedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// FirstPartyStorkContractPublisherUserRemoved represents a PublisherUserRemoved event raised by the FirstPartyStorkContract contract.
type FirstPartyStorkContractPublisherUserRemoved struct {
	PubKey common.Address
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterPublisherUserRemoved is a free log retrieval operation binding the contract event 0xde37507a2967feaf95649773534bddc03d153af2b26b9ea17cdb9b62c5bbf18e.
//
// Solidity: event PublisherUserRemoved(address indexed pubKey)
func (_FirstPartyStorkContract *FirstPartyStorkContractFilterer) FilterPublisherUserRemoved(opts *bind.FilterOpts, pubKey []common.Address) (*FirstPartyStorkContractPublisherUserRemovedIterator, error) {

	var pubKeyRule []interface{}
	for _, pubKeyItem := range pubKey {
		pubKeyRule = append(pubKeyRule, pubKeyItem)
	}

	logs, sub, err := _FirstPartyStorkContract.contract.FilterLogs(opts, "PublisherUserRemoved", pubKeyRule)
	if err != nil {
		return nil, err
	}
	return &FirstPartyStorkContractPublisherUserRemovedIterator{contract: _FirstPartyStorkContract.contract, event: "PublisherUserRemoved", logs: logs, sub: sub}, nil
}

// WatchPublisherUserRemoved is a free log subscription operation binding the contract event 0xde37507a2967feaf95649773534bddc03d153af2b26b9ea17cdb9b62c5bbf18e.
//
// Solidity: event PublisherUserRemoved(address indexed pubKey)
func (_FirstPartyStorkContract *FirstPartyStorkContractFilterer) WatchPublisherUserRemoved(opts *bind.WatchOpts, sink chan<- *FirstPartyStorkContractPublisherUserRemoved, pubKey []common.Address) (event.Subscription, error) {

	var pubKeyRule []interface{}
	for _, pubKeyItem := range pubKey {
		pubKeyRule = append(pubKeyRule, pubKeyItem)
	}

	logs, sub, err := _FirstPartyStorkContract.contract.WatchLogs(opts, "PublisherUserRemoved", pubKeyRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(FirstPartyStorkContractPublisherUserRemoved)
				if err := _FirstPartyStorkContract.contract.UnpackLog(event, "PublisherUserRemoved", log); err != nil {
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
func (_FirstPartyStorkContract *FirstPartyStorkContractFilterer) ParsePublisherUserRemoved(log types.Log) (*FirstPartyStorkContractPublisherUserRemoved, error) {
	event := new(FirstPartyStorkContractPublisherUserRemoved)
	if err := _FirstPartyStorkContract.contract.UnpackLog(event, "PublisherUserRemoved", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// FirstPartyStorkContractValueUpdateIterator is returned from FilterValueUpdate and is used to iterate over the raw logs and unpacked data for ValueUpdate events raised by the FirstPartyStorkContract contract.
type FirstPartyStorkContractValueUpdateIterator struct {
	Event *FirstPartyStorkContractValueUpdate // Event containing the contract specifics and raw log

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
func (it *FirstPartyStorkContractValueUpdateIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(FirstPartyStorkContractValueUpdate)
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
		it.Event = new(FirstPartyStorkContractValueUpdate)
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
func (it *FirstPartyStorkContractValueUpdateIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *FirstPartyStorkContractValueUpdateIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// FirstPartyStorkContractValueUpdate represents a ValueUpdate event raised by the FirstPartyStorkContract contract.
type FirstPartyStorkContractValueUpdate struct {
	PubKey         common.Address
	AssetId        common.Hash
	TimestampNs    uint64
	QuantizedValue *big.Int
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterValueUpdate is a free log retrieval operation binding the contract event 0x0596010914c581c18e615a5b0d097d78728eb775f2412d8bedd4dbc808f4f855.
//
// Solidity: event ValueUpdate(address indexed pubKey, string indexed assetId, uint64 timestampNs, int192 quantizedValue)
func (_FirstPartyStorkContract *FirstPartyStorkContractFilterer) FilterValueUpdate(opts *bind.FilterOpts, pubKey []common.Address, assetId []string) (*FirstPartyStorkContractValueUpdateIterator, error) {

	var pubKeyRule []interface{}
	for _, pubKeyItem := range pubKey {
		pubKeyRule = append(pubKeyRule, pubKeyItem)
	}
	var assetIdRule []interface{}
	for _, assetIdItem := range assetId {
		assetIdRule = append(assetIdRule, assetIdItem)
	}

	logs, sub, err := _FirstPartyStorkContract.contract.FilterLogs(opts, "ValueUpdate", pubKeyRule, assetIdRule)
	if err != nil {
		return nil, err
	}
	return &FirstPartyStorkContractValueUpdateIterator{contract: _FirstPartyStorkContract.contract, event: "ValueUpdate", logs: logs, sub: sub}, nil
}

// WatchValueUpdate is a free log subscription operation binding the contract event 0x0596010914c581c18e615a5b0d097d78728eb775f2412d8bedd4dbc808f4f855.
//
// Solidity: event ValueUpdate(address indexed pubKey, string indexed assetId, uint64 timestampNs, int192 quantizedValue)
func (_FirstPartyStorkContract *FirstPartyStorkContractFilterer) WatchValueUpdate(opts *bind.WatchOpts, sink chan<- *FirstPartyStorkContractValueUpdate, pubKey []common.Address, assetId []string) (event.Subscription, error) {

	var pubKeyRule []interface{}
	for _, pubKeyItem := range pubKey {
		pubKeyRule = append(pubKeyRule, pubKeyItem)
	}
	var assetIdRule []interface{}
	for _, assetIdItem := range assetId {
		assetIdRule = append(assetIdRule, assetIdItem)
	}

	logs, sub, err := _FirstPartyStorkContract.contract.WatchLogs(opts, "ValueUpdate", pubKeyRule, assetIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(FirstPartyStorkContractValueUpdate)
				if err := _FirstPartyStorkContract.contract.UnpackLog(event, "ValueUpdate", log); err != nil {
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
func (_FirstPartyStorkContract *FirstPartyStorkContractFilterer) ParseValueUpdate(log types.Log) (*FirstPartyStorkContractValueUpdate, error) {
	event := new(FirstPartyStorkContractValueUpdate)
	if err := _FirstPartyStorkContract.contract.UnpackLog(event, "ValueUpdate", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
