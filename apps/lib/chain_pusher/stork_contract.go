// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package chain_pusher

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

// StorkStructsPublisherSignature is an auto generated low-level Go binding around an user-defined struct.
type StorkStructsPublisherSignature struct {
	PubKey         common.Address
	AssetPairId    string
	Timestamp      uint64
	QuantizedValue *big.Int
	R              [32]byte
	S              [32]byte
	V              uint8
}

// StorkStructsTemporalNumericValue is an auto generated low-level Go binding around an user-defined struct.
type StorkStructsTemporalNumericValue struct {
	TimestampNs    uint64
	QuantizedValue *big.Int
}

// StorkStructsTemporalNumericValueInput is an auto generated low-level Go binding around an user-defined struct.
type StorkStructsTemporalNumericValueInput struct {
	TemporalNumericValue StorkStructsTemporalNumericValue
	Id                   [32]byte
	PublisherMerkleRoot  [32]byte
	ValueComputeAlgHash  [32]byte
	R                    [32]byte
	S                    [32]byte
	V                    uint8
}

// StorkContractMetaData contains all meta data concerning the StorkContract contract.
var StorkContractMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"address\",\"name\":\"target\",\"type\":\"address\"}],\"name\":\"AddressEmptyCode\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"implementation\",\"type\":\"address\"}],\"name\":\"ERC1967InvalidImplementation\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"ERC1967NonPayable\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"FailedInnerCall\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InsufficientFee\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidInitialization\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidSignature\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NoFreshUpdate\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotFound\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotInitializing\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"OwnableInvalidOwner\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"OwnableUnauthorizedAccount\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"StaleValue\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"UUPSUnauthorizedCallContext\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"slot\",\"type\":\"bytes32\"}],\"name\":\"UUPSUnsupportedProxiableUUID\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"version\",\"type\":\"uint64\"}],\"name\":\"Initialized\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"implementation\",\"type\":\"address\"}],\"name\":\"Upgraded\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"id\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"timestampNs\",\"type\":\"uint64\"},{\"indexed\":false,\"internalType\":\"int192\",\"name\":\"quantizedValue\",\"type\":\"int192\"}],\"name\":\"ValueUpdate\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"UPGRADE_INTERFACE_VERSION\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"id\",\"type\":\"bytes32\"}],\"name\":\"getTemporalNumericValueV1\",\"outputs\":[{\"components\":[{\"internalType\":\"uint64\",\"name\":\"timestampNs\",\"type\":\"uint64\"},{\"internalType\":\"int192\",\"name\":\"quantizedValue\",\"type\":\"int192\"}],\"internalType\":\"structStorkStructs.TemporalNumericValue\",\"name\":\"value\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"components\":[{\"internalType\":\"uint64\",\"name\":\"timestampNs\",\"type\":\"uint64\"},{\"internalType\":\"int192\",\"name\":\"quantizedValue\",\"type\":\"int192\"}],\"internalType\":\"structStorkStructs.TemporalNumericValue\",\"name\":\"temporalNumericValue\",\"type\":\"tuple\"},{\"internalType\":\"bytes32\",\"name\":\"id\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"publisherMerkleRoot\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"valueComputeAlgHash\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"r\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"s\",\"type\":\"bytes32\"},{\"internalType\":\"uint8\",\"name\":\"v\",\"type\":\"uint8\"}],\"internalType\":\"structStorkStructs.TemporalNumericValueInput[]\",\"name\":\"updateData\",\"type\":\"tuple[]\"}],\"name\":\"getUpdateFeeV1\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"feeAmount\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"initialOwner\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"storkPublicKey\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"validTimePeriodSeconds\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"singleUpdateFeeInWei\",\"type\":\"uint256\"}],\"name\":\"initialize\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"proxiableUUID\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"singleUpdateFeeInWei\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"storkPublicKey\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"maxStorkPerBlock\",\"type\":\"uint256\"}],\"name\":\"updateSingleUpdateFeeInWei\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"storkPublicKey\",\"type\":\"address\"}],\"name\":\"updateStorkPublicKey\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"components\":[{\"internalType\":\"uint64\",\"name\":\"timestampNs\",\"type\":\"uint64\"},{\"internalType\":\"int192\",\"name\":\"quantizedValue\",\"type\":\"int192\"}],\"internalType\":\"structStorkStructs.TemporalNumericValue\",\"name\":\"temporalNumericValue\",\"type\":\"tuple\"},{\"internalType\":\"bytes32\",\"name\":\"id\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"publisherMerkleRoot\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"valueComputeAlgHash\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"r\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"s\",\"type\":\"bytes32\"},{\"internalType\":\"uint8\",\"name\":\"v\",\"type\":\"uint8\"}],\"internalType\":\"structStorkStructs.TemporalNumericValueInput[]\",\"name\":\"updateData\",\"type\":\"tuple[]\"}],\"name\":\"updateTemporalNumericValuesV1\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"validTimePeriodSeconds\",\"type\":\"uint256\"}],\"name\":\"updateValidTimePeriodSeconds\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newImplementation\",\"type\":\"address\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"}],\"name\":\"upgradeToAndCall\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"validTimePeriodSeconds\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32[]\",\"name\":\"leaves\",\"type\":\"bytes32[]\"},{\"internalType\":\"bytes32\",\"name\":\"root\",\"type\":\"bytes32\"}],\"name\":\"verifyMerkleRoot\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"oraclePubKey\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"assetPairId\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"},{\"internalType\":\"bytes32\",\"name\":\"r\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"s\",\"type\":\"bytes32\"},{\"internalType\":\"uint8\",\"name\":\"v\",\"type\":\"uint8\"}],\"name\":\"verifyPublisherSignatureV1\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"address\",\"name\":\"pubKey\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"assetPairId\",\"type\":\"string\"},{\"internalType\":\"uint64\",\"name\":\"timestamp\",\"type\":\"uint64\"},{\"internalType\":\"uint256\",\"name\":\"quantizedValue\",\"type\":\"uint256\"},{\"internalType\":\"bytes32\",\"name\":\"r\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"s\",\"type\":\"bytes32\"},{\"internalType\":\"uint8\",\"name\":\"v\",\"type\":\"uint8\"}],\"internalType\":\"structStorkStructs.PublisherSignature[]\",\"name\":\"signatures\",\"type\":\"tuple[]\"},{\"internalType\":\"bytes32\",\"name\":\"merkleRoot\",\"type\":\"bytes32\"}],\"name\":\"verifyPublisherSignaturesV1\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"storkPubKey\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"id\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"recvTime\",\"type\":\"uint256\"},{\"internalType\":\"int256\",\"name\":\"quantizedValue\",\"type\":\"int256\"},{\"internalType\":\"bytes32\",\"name\":\"publisherMerkleRoot\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"valueComputeAlgHash\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"r\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"s\",\"type\":\"bytes32\"},{\"internalType\":\"uint8\",\"name\":\"v\",\"type\":\"uint8\"}],\"name\":\"verifyStorkSignatureV1\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"version\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"pure\",\"type\":\"function\"}]",
}

// StorkContractABI is the input ABI used to generate the binding from.
// Deprecated: Use StorkContractMetaData.ABI instead.
var StorkContractABI = StorkContractMetaData.ABI

// StorkContract is an auto generated Go binding around an Ethereum contract.
type StorkContract struct {
	StorkContractCaller     // Read-only binding to the contract
	StorkContractTransactor // Write-only binding to the contract
	StorkContractFilterer   // Log filterer for contract events
}

// StorkContractCaller is an auto generated read-only Go binding around an Ethereum contract.
type StorkContractCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// StorkContractTransactor is an auto generated write-only Go binding around an Ethereum contract.
type StorkContractTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// StorkContractFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type StorkContractFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// StorkContractSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type StorkContractSession struct {
	Contract     *StorkContract    // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// StorkContractCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type StorkContractCallerSession struct {
	Contract *StorkContractCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts        // Call options to use throughout this session
}

// StorkContractTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type StorkContractTransactorSession struct {
	Contract     *StorkContractTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts        // Transaction auth options to use throughout this session
}

// StorkContractRaw is an auto generated low-level Go binding around an Ethereum contract.
type StorkContractRaw struct {
	Contract *StorkContract // Generic contract binding to access the raw methods on
}

// StorkContractCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type StorkContractCallerRaw struct {
	Contract *StorkContractCaller // Generic read-only contract binding to access the raw methods on
}

// StorkContractTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type StorkContractTransactorRaw struct {
	Contract *StorkContractTransactor // Generic write-only contract binding to access the raw methods on
}

// NewStorkContract creates a new instance of StorkContract, bound to a specific deployed contract.
func NewStorkContract(address common.Address, backend bind.ContractBackend) (*StorkContract, error) {
	contract, err := bindStorkContract(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &StorkContract{StorkContractCaller: StorkContractCaller{contract: contract}, StorkContractTransactor: StorkContractTransactor{contract: contract}, StorkContractFilterer: StorkContractFilterer{contract: contract}}, nil
}

// NewStorkContractCaller creates a new read-only instance of StorkContract, bound to a specific deployed contract.
func NewStorkContractCaller(address common.Address, caller bind.ContractCaller) (*StorkContractCaller, error) {
	contract, err := bindStorkContract(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &StorkContractCaller{contract: contract}, nil
}

// NewStorkContractTransactor creates a new write-only instance of StorkContract, bound to a specific deployed contract.
func NewStorkContractTransactor(address common.Address, transactor bind.ContractTransactor) (*StorkContractTransactor, error) {
	contract, err := bindStorkContract(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &StorkContractTransactor{contract: contract}, nil
}

// NewStorkContractFilterer creates a new log filterer instance of StorkContract, bound to a specific deployed contract.
func NewStorkContractFilterer(address common.Address, filterer bind.ContractFilterer) (*StorkContractFilterer, error) {
	contract, err := bindStorkContract(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &StorkContractFilterer{contract: contract}, nil
}

// bindStorkContract binds a generic wrapper to an already deployed contract.
func bindStorkContract(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := StorkContractMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_StorkContract *StorkContractRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _StorkContract.Contract.StorkContractCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_StorkContract *StorkContractRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _StorkContract.Contract.StorkContractTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_StorkContract *StorkContractRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _StorkContract.Contract.StorkContractTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_StorkContract *StorkContractCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _StorkContract.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_StorkContract *StorkContractTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _StorkContract.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_StorkContract *StorkContractTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _StorkContract.Contract.contract.Transact(opts, method, params...)
}

// UPGRADEINTERFACEVERSION is a free data retrieval call binding the contract method 0xad3cb1cc.
//
// Solidity: function UPGRADE_INTERFACE_VERSION() view returns(string)
func (_StorkContract *StorkContractCaller) UPGRADEINTERFACEVERSION(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _StorkContract.contract.Call(opts, &out, "UPGRADE_INTERFACE_VERSION")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// UPGRADEINTERFACEVERSION is a free data retrieval call binding the contract method 0xad3cb1cc.
//
// Solidity: function UPGRADE_INTERFACE_VERSION() view returns(string)
func (_StorkContract *StorkContractSession) UPGRADEINTERFACEVERSION() (string, error) {
	return _StorkContract.Contract.UPGRADEINTERFACEVERSION(&_StorkContract.CallOpts)
}

// UPGRADEINTERFACEVERSION is a free data retrieval call binding the contract method 0xad3cb1cc.
//
// Solidity: function UPGRADE_INTERFACE_VERSION() view returns(string)
func (_StorkContract *StorkContractCallerSession) UPGRADEINTERFACEVERSION() (string, error) {
	return _StorkContract.Contract.UPGRADEINTERFACEVERSION(&_StorkContract.CallOpts)
}

// GetTemporalNumericValueV1 is a free data retrieval call binding the contract method 0x19af7a40.
//
// Solidity: function getTemporalNumericValueV1(bytes32 id) view returns((uint64,int192) value)
func (_StorkContract *StorkContractCaller) GetTemporalNumericValueV1(opts *bind.CallOpts, id [32]byte) (StorkStructsTemporalNumericValue, error) {
	var out []interface{}
	err := _StorkContract.contract.Call(opts, &out, "getTemporalNumericValueV1", id)

	if err != nil {
		return *new(StorkStructsTemporalNumericValue), err
	}

	out0 := *abi.ConvertType(out[0], new(StorkStructsTemporalNumericValue)).(*StorkStructsTemporalNumericValue)

	return out0, err

}

// GetTemporalNumericValueV1 is a free data retrieval call binding the contract method 0x19af7a40.
//
// Solidity: function getTemporalNumericValueV1(bytes32 id) view returns((uint64,int192) value)
func (_StorkContract *StorkContractSession) GetTemporalNumericValueV1(id [32]byte) (StorkStructsTemporalNumericValue, error) {
	return _StorkContract.Contract.GetTemporalNumericValueV1(&_StorkContract.CallOpts, id)
}

// GetTemporalNumericValueV1 is a free data retrieval call binding the contract method 0x19af7a40.
//
// Solidity: function getTemporalNumericValueV1(bytes32 id) view returns((uint64,int192) value)
func (_StorkContract *StorkContractCallerSession) GetTemporalNumericValueV1(id [32]byte) (StorkStructsTemporalNumericValue, error) {
	return _StorkContract.Contract.GetTemporalNumericValueV1(&_StorkContract.CallOpts, id)
}

// GetUpdateFeeV1 is a free data retrieval call binding the contract method 0xb2255ba3.
//
// Solidity: function getUpdateFeeV1(((uint64,int192),bytes32,bytes32,bytes32,bytes32,bytes32,uint8)[] updateData) view returns(uint256 feeAmount)
func (_StorkContract *StorkContractCaller) GetUpdateFeeV1(opts *bind.CallOpts, updateData []StorkStructsTemporalNumericValueInput) (*big.Int, error) {
	var out []interface{}
	err := _StorkContract.contract.Call(opts, &out, "getUpdateFeeV1", updateData)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetUpdateFeeV1 is a free data retrieval call binding the contract method 0xb2255ba3.
//
// Solidity: function getUpdateFeeV1(((uint64,int192),bytes32,bytes32,bytes32,bytes32,bytes32,uint8)[] updateData) view returns(uint256 feeAmount)
func (_StorkContract *StorkContractSession) GetUpdateFeeV1(updateData []StorkStructsTemporalNumericValueInput) (*big.Int, error) {
	return _StorkContract.Contract.GetUpdateFeeV1(&_StorkContract.CallOpts, updateData)
}

// GetUpdateFeeV1 is a free data retrieval call binding the contract method 0xb2255ba3.
//
// Solidity: function getUpdateFeeV1(((uint64,int192),bytes32,bytes32,bytes32,bytes32,bytes32,uint8)[] updateData) view returns(uint256 feeAmount)
func (_StorkContract *StorkContractCallerSession) GetUpdateFeeV1(updateData []StorkStructsTemporalNumericValueInput) (*big.Int, error) {
	return _StorkContract.Contract.GetUpdateFeeV1(&_StorkContract.CallOpts, updateData)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_StorkContract *StorkContractCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _StorkContract.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_StorkContract *StorkContractSession) Owner() (common.Address, error) {
	return _StorkContract.Contract.Owner(&_StorkContract.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_StorkContract *StorkContractCallerSession) Owner() (common.Address, error) {
	return _StorkContract.Contract.Owner(&_StorkContract.CallOpts)
}

// ProxiableUUID is a free data retrieval call binding the contract method 0x52d1902d.
//
// Solidity: function proxiableUUID() view returns(bytes32)
func (_StorkContract *StorkContractCaller) ProxiableUUID(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _StorkContract.contract.Call(opts, &out, "proxiableUUID")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// ProxiableUUID is a free data retrieval call binding the contract method 0x52d1902d.
//
// Solidity: function proxiableUUID() view returns(bytes32)
func (_StorkContract *StorkContractSession) ProxiableUUID() ([32]byte, error) {
	return _StorkContract.Contract.ProxiableUUID(&_StorkContract.CallOpts)
}

// ProxiableUUID is a free data retrieval call binding the contract method 0x52d1902d.
//
// Solidity: function proxiableUUID() view returns(bytes32)
func (_StorkContract *StorkContractCallerSession) ProxiableUUID() ([32]byte, error) {
	return _StorkContract.Contract.ProxiableUUID(&_StorkContract.CallOpts)
}

// SingleUpdateFeeInWei is a free data retrieval call binding the contract method 0x48b6404d.
//
// Solidity: function singleUpdateFeeInWei() view returns(uint256)
func (_StorkContract *StorkContractCaller) SingleUpdateFeeInWei(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _StorkContract.contract.Call(opts, &out, "singleUpdateFeeInWei")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// SingleUpdateFeeInWei is a free data retrieval call binding the contract method 0x48b6404d.
//
// Solidity: function singleUpdateFeeInWei() view returns(uint256)
func (_StorkContract *StorkContractSession) SingleUpdateFeeInWei() (*big.Int, error) {
	return _StorkContract.Contract.SingleUpdateFeeInWei(&_StorkContract.CallOpts)
}

// SingleUpdateFeeInWei is a free data retrieval call binding the contract method 0x48b6404d.
//
// Solidity: function singleUpdateFeeInWei() view returns(uint256)
func (_StorkContract *StorkContractCallerSession) SingleUpdateFeeInWei() (*big.Int, error) {
	return _StorkContract.Contract.SingleUpdateFeeInWei(&_StorkContract.CallOpts)
}

// StorkPublicKey is a free data retrieval call binding the contract method 0x8eeae4a7.
//
// Solidity: function storkPublicKey() view returns(address)
func (_StorkContract *StorkContractCaller) StorkPublicKey(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _StorkContract.contract.Call(opts, &out, "storkPublicKey")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// StorkPublicKey is a free data retrieval call binding the contract method 0x8eeae4a7.
//
// Solidity: function storkPublicKey() view returns(address)
func (_StorkContract *StorkContractSession) StorkPublicKey() (common.Address, error) {
	return _StorkContract.Contract.StorkPublicKey(&_StorkContract.CallOpts)
}

// StorkPublicKey is a free data retrieval call binding the contract method 0x8eeae4a7.
//
// Solidity: function storkPublicKey() view returns(address)
func (_StorkContract *StorkContractCallerSession) StorkPublicKey() (common.Address, error) {
	return _StorkContract.Contract.StorkPublicKey(&_StorkContract.CallOpts)
}

// ValidTimePeriodSeconds is a free data retrieval call binding the contract method 0xcb718a9b.
//
// Solidity: function validTimePeriodSeconds() view returns(uint256)
func (_StorkContract *StorkContractCaller) ValidTimePeriodSeconds(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _StorkContract.contract.Call(opts, &out, "validTimePeriodSeconds")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// ValidTimePeriodSeconds is a free data retrieval call binding the contract method 0xcb718a9b.
//
// Solidity: function validTimePeriodSeconds() view returns(uint256)
func (_StorkContract *StorkContractSession) ValidTimePeriodSeconds() (*big.Int, error) {
	return _StorkContract.Contract.ValidTimePeriodSeconds(&_StorkContract.CallOpts)
}

// ValidTimePeriodSeconds is a free data retrieval call binding the contract method 0xcb718a9b.
//
// Solidity: function validTimePeriodSeconds() view returns(uint256)
func (_StorkContract *StorkContractCallerSession) ValidTimePeriodSeconds() (*big.Int, error) {
	return _StorkContract.Contract.ValidTimePeriodSeconds(&_StorkContract.CallOpts)
}

// VerifyMerkleRoot is a free data retrieval call binding the contract method 0x44ecc82c.
//
// Solidity: function verifyMerkleRoot(bytes32[] leaves, bytes32 root) pure returns(bool)
func (_StorkContract *StorkContractCaller) VerifyMerkleRoot(opts *bind.CallOpts, leaves [][32]byte, root [32]byte) (bool, error) {
	var out []interface{}
	err := _StorkContract.contract.Call(opts, &out, "verifyMerkleRoot", leaves, root)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// VerifyMerkleRoot is a free data retrieval call binding the contract method 0x44ecc82c.
//
// Solidity: function verifyMerkleRoot(bytes32[] leaves, bytes32 root) pure returns(bool)
func (_StorkContract *StorkContractSession) VerifyMerkleRoot(leaves [][32]byte, root [32]byte) (bool, error) {
	return _StorkContract.Contract.VerifyMerkleRoot(&_StorkContract.CallOpts, leaves, root)
}

// VerifyMerkleRoot is a free data retrieval call binding the contract method 0x44ecc82c.
//
// Solidity: function verifyMerkleRoot(bytes32[] leaves, bytes32 root) pure returns(bool)
func (_StorkContract *StorkContractCallerSession) VerifyMerkleRoot(leaves [][32]byte, root [32]byte) (bool, error) {
	return _StorkContract.Contract.VerifyMerkleRoot(&_StorkContract.CallOpts, leaves, root)
}

// VerifyPublisherSignatureV1 is a free data retrieval call binding the contract method 0xd83cfe2c.
//
// Solidity: function verifyPublisherSignatureV1(address oraclePubKey, string assetPairId, uint256 timestamp, uint256 value, bytes32 r, bytes32 s, uint8 v) pure returns(bool)
func (_StorkContract *StorkContractCaller) VerifyPublisherSignatureV1(opts *bind.CallOpts, oraclePubKey common.Address, assetPairId string, timestamp *big.Int, value *big.Int, r [32]byte, s [32]byte, v uint8) (bool, error) {
	var out []interface{}
	err := _StorkContract.contract.Call(opts, &out, "verifyPublisherSignatureV1", oraclePubKey, assetPairId, timestamp, value, r, s, v)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// VerifyPublisherSignatureV1 is a free data retrieval call binding the contract method 0xd83cfe2c.
//
// Solidity: function verifyPublisherSignatureV1(address oraclePubKey, string assetPairId, uint256 timestamp, uint256 value, bytes32 r, bytes32 s, uint8 v) pure returns(bool)
func (_StorkContract *StorkContractSession) VerifyPublisherSignatureV1(oraclePubKey common.Address, assetPairId string, timestamp *big.Int, value *big.Int, r [32]byte, s [32]byte, v uint8) (bool, error) {
	return _StorkContract.Contract.VerifyPublisherSignatureV1(&_StorkContract.CallOpts, oraclePubKey, assetPairId, timestamp, value, r, s, v)
}

// VerifyPublisherSignatureV1 is a free data retrieval call binding the contract method 0xd83cfe2c.
//
// Solidity: function verifyPublisherSignatureV1(address oraclePubKey, string assetPairId, uint256 timestamp, uint256 value, bytes32 r, bytes32 s, uint8 v) pure returns(bool)
func (_StorkContract *StorkContractCallerSession) VerifyPublisherSignatureV1(oraclePubKey common.Address, assetPairId string, timestamp *big.Int, value *big.Int, r [32]byte, s [32]byte, v uint8) (bool, error) {
	return _StorkContract.Contract.VerifyPublisherSignatureV1(&_StorkContract.CallOpts, oraclePubKey, assetPairId, timestamp, value, r, s, v)
}

// VerifyPublisherSignaturesV1 is a free data retrieval call binding the contract method 0x1519e36c.
//
// Solidity: function verifyPublisherSignaturesV1((address,string,uint64,uint256,bytes32,bytes32,uint8)[] signatures, bytes32 merkleRoot) pure returns(bool)
func (_StorkContract *StorkContractCaller) VerifyPublisherSignaturesV1(opts *bind.CallOpts, signatures []StorkStructsPublisherSignature, merkleRoot [32]byte) (bool, error) {
	var out []interface{}
	err := _StorkContract.contract.Call(opts, &out, "verifyPublisherSignaturesV1", signatures, merkleRoot)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// VerifyPublisherSignaturesV1 is a free data retrieval call binding the contract method 0x1519e36c.
//
// Solidity: function verifyPublisherSignaturesV1((address,string,uint64,uint256,bytes32,bytes32,uint8)[] signatures, bytes32 merkleRoot) pure returns(bool)
func (_StorkContract *StorkContractSession) VerifyPublisherSignaturesV1(signatures []StorkStructsPublisherSignature, merkleRoot [32]byte) (bool, error) {
	return _StorkContract.Contract.VerifyPublisherSignaturesV1(&_StorkContract.CallOpts, signatures, merkleRoot)
}

// VerifyPublisherSignaturesV1 is a free data retrieval call binding the contract method 0x1519e36c.
//
// Solidity: function verifyPublisherSignaturesV1((address,string,uint64,uint256,bytes32,bytes32,uint8)[] signatures, bytes32 merkleRoot) pure returns(bool)
func (_StorkContract *StorkContractCallerSession) VerifyPublisherSignaturesV1(signatures []StorkStructsPublisherSignature, merkleRoot [32]byte) (bool, error) {
	return _StorkContract.Contract.VerifyPublisherSignaturesV1(&_StorkContract.CallOpts, signatures, merkleRoot)
}

// VerifyStorkSignatureV1 is a free data retrieval call binding the contract method 0x2a6cd210.
//
// Solidity: function verifyStorkSignatureV1(address storkPubKey, bytes32 id, uint256 recvTime, int256 quantizedValue, bytes32 publisherMerkleRoot, bytes32 valueComputeAlgHash, bytes32 r, bytes32 s, uint8 v) pure returns(bool)
func (_StorkContract *StorkContractCaller) VerifyStorkSignatureV1(opts *bind.CallOpts, storkPubKey common.Address, id [32]byte, recvTime *big.Int, quantizedValue *big.Int, publisherMerkleRoot [32]byte, valueComputeAlgHash [32]byte, r [32]byte, s [32]byte, v uint8) (bool, error) {
	var out []interface{}
	err := _StorkContract.contract.Call(opts, &out, "verifyStorkSignatureV1", storkPubKey, id, recvTime, quantizedValue, publisherMerkleRoot, valueComputeAlgHash, r, s, v)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// VerifyStorkSignatureV1 is a free data retrieval call binding the contract method 0x2a6cd210.
//
// Solidity: function verifyStorkSignatureV1(address storkPubKey, bytes32 id, uint256 recvTime, int256 quantizedValue, bytes32 publisherMerkleRoot, bytes32 valueComputeAlgHash, bytes32 r, bytes32 s, uint8 v) pure returns(bool)
func (_StorkContract *StorkContractSession) VerifyStorkSignatureV1(storkPubKey common.Address, id [32]byte, recvTime *big.Int, quantizedValue *big.Int, publisherMerkleRoot [32]byte, valueComputeAlgHash [32]byte, r [32]byte, s [32]byte, v uint8) (bool, error) {
	return _StorkContract.Contract.VerifyStorkSignatureV1(&_StorkContract.CallOpts, storkPubKey, id, recvTime, quantizedValue, publisherMerkleRoot, valueComputeAlgHash, r, s, v)
}

// VerifyStorkSignatureV1 is a free data retrieval call binding the contract method 0x2a6cd210.
//
// Solidity: function verifyStorkSignatureV1(address storkPubKey, bytes32 id, uint256 recvTime, int256 quantizedValue, bytes32 publisherMerkleRoot, bytes32 valueComputeAlgHash, bytes32 r, bytes32 s, uint8 v) pure returns(bool)
func (_StorkContract *StorkContractCallerSession) VerifyStorkSignatureV1(storkPubKey common.Address, id [32]byte, recvTime *big.Int, quantizedValue *big.Int, publisherMerkleRoot [32]byte, valueComputeAlgHash [32]byte, r [32]byte, s [32]byte, v uint8) (bool, error) {
	return _StorkContract.Contract.VerifyStorkSignatureV1(&_StorkContract.CallOpts, storkPubKey, id, recvTime, quantizedValue, publisherMerkleRoot, valueComputeAlgHash, r, s, v)
}

// Version is a free data retrieval call binding the contract method 0x54fd4d50.
//
// Solidity: function version() pure returns(string)
func (_StorkContract *StorkContractCaller) Version(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _StorkContract.contract.Call(opts, &out, "version")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Version is a free data retrieval call binding the contract method 0x54fd4d50.
//
// Solidity: function version() pure returns(string)
func (_StorkContract *StorkContractSession) Version() (string, error) {
	return _StorkContract.Contract.Version(&_StorkContract.CallOpts)
}

// Version is a free data retrieval call binding the contract method 0x54fd4d50.
//
// Solidity: function version() pure returns(string)
func (_StorkContract *StorkContractCallerSession) Version() (string, error) {
	return _StorkContract.Contract.Version(&_StorkContract.CallOpts)
}

// Initialize is a paid mutator transaction binding the contract method 0xeb990c59.
//
// Solidity: function initialize(address initialOwner, address storkPublicKey, uint256 validTimePeriodSeconds, uint256 singleUpdateFeeInWei) returns()
func (_StorkContract *StorkContractTransactor) Initialize(opts *bind.TransactOpts, initialOwner common.Address, storkPublicKey common.Address, validTimePeriodSeconds *big.Int, singleUpdateFeeInWei *big.Int) (*types.Transaction, error) {
	return _StorkContract.contract.Transact(opts, "initialize", initialOwner, storkPublicKey, validTimePeriodSeconds, singleUpdateFeeInWei)
}

// Initialize is a paid mutator transaction binding the contract method 0xeb990c59.
//
// Solidity: function initialize(address initialOwner, address storkPublicKey, uint256 validTimePeriodSeconds, uint256 singleUpdateFeeInWei) returns()
func (_StorkContract *StorkContractSession) Initialize(initialOwner common.Address, storkPublicKey common.Address, validTimePeriodSeconds *big.Int, singleUpdateFeeInWei *big.Int) (*types.Transaction, error) {
	return _StorkContract.Contract.Initialize(&_StorkContract.TransactOpts, initialOwner, storkPublicKey, validTimePeriodSeconds, singleUpdateFeeInWei)
}

// Initialize is a paid mutator transaction binding the contract method 0xeb990c59.
//
// Solidity: function initialize(address initialOwner, address storkPublicKey, uint256 validTimePeriodSeconds, uint256 singleUpdateFeeInWei) returns()
func (_StorkContract *StorkContractTransactorSession) Initialize(initialOwner common.Address, storkPublicKey common.Address, validTimePeriodSeconds *big.Int, singleUpdateFeeInWei *big.Int) (*types.Transaction, error) {
	return _StorkContract.Contract.Initialize(&_StorkContract.TransactOpts, initialOwner, storkPublicKey, validTimePeriodSeconds, singleUpdateFeeInWei)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_StorkContract *StorkContractTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _StorkContract.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_StorkContract *StorkContractSession) RenounceOwnership() (*types.Transaction, error) {
	return _StorkContract.Contract.RenounceOwnership(&_StorkContract.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_StorkContract *StorkContractTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _StorkContract.Contract.RenounceOwnership(&_StorkContract.TransactOpts)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_StorkContract *StorkContractTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _StorkContract.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_StorkContract *StorkContractSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _StorkContract.Contract.TransferOwnership(&_StorkContract.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_StorkContract *StorkContractTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _StorkContract.Contract.TransferOwnership(&_StorkContract.TransactOpts, newOwner)
}

// UpdateSingleUpdateFeeInWei is a paid mutator transaction binding the contract method 0x4268340a.
//
// Solidity: function updateSingleUpdateFeeInWei(uint256 maxStorkPerBlock) returns()
func (_StorkContract *StorkContractTransactor) UpdateSingleUpdateFeeInWei(opts *bind.TransactOpts, maxStorkPerBlock *big.Int) (*types.Transaction, error) {
	return _StorkContract.contract.Transact(opts, "updateSingleUpdateFeeInWei", maxStorkPerBlock)
}

// UpdateSingleUpdateFeeInWei is a paid mutator transaction binding the contract method 0x4268340a.
//
// Solidity: function updateSingleUpdateFeeInWei(uint256 maxStorkPerBlock) returns()
func (_StorkContract *StorkContractSession) UpdateSingleUpdateFeeInWei(maxStorkPerBlock *big.Int) (*types.Transaction, error) {
	return _StorkContract.Contract.UpdateSingleUpdateFeeInWei(&_StorkContract.TransactOpts, maxStorkPerBlock)
}

// UpdateSingleUpdateFeeInWei is a paid mutator transaction binding the contract method 0x4268340a.
//
// Solidity: function updateSingleUpdateFeeInWei(uint256 maxStorkPerBlock) returns()
func (_StorkContract *StorkContractTransactorSession) UpdateSingleUpdateFeeInWei(maxStorkPerBlock *big.Int) (*types.Transaction, error) {
	return _StorkContract.Contract.UpdateSingleUpdateFeeInWei(&_StorkContract.TransactOpts, maxStorkPerBlock)
}

// UpdateStorkPublicKey is a paid mutator transaction binding the contract method 0x11992f3b.
//
// Solidity: function updateStorkPublicKey(address storkPublicKey) returns()
func (_StorkContract *StorkContractTransactor) UpdateStorkPublicKey(opts *bind.TransactOpts, storkPublicKey common.Address) (*types.Transaction, error) {
	return _StorkContract.contract.Transact(opts, "updateStorkPublicKey", storkPublicKey)
}

// UpdateStorkPublicKey is a paid mutator transaction binding the contract method 0x11992f3b.
//
// Solidity: function updateStorkPublicKey(address storkPublicKey) returns()
func (_StorkContract *StorkContractSession) UpdateStorkPublicKey(storkPublicKey common.Address) (*types.Transaction, error) {
	return _StorkContract.Contract.UpdateStorkPublicKey(&_StorkContract.TransactOpts, storkPublicKey)
}

// UpdateStorkPublicKey is a paid mutator transaction binding the contract method 0x11992f3b.
//
// Solidity: function updateStorkPublicKey(address storkPublicKey) returns()
func (_StorkContract *StorkContractTransactorSession) UpdateStorkPublicKey(storkPublicKey common.Address) (*types.Transaction, error) {
	return _StorkContract.Contract.UpdateStorkPublicKey(&_StorkContract.TransactOpts, storkPublicKey)
}

// UpdateTemporalNumericValuesV1 is a paid mutator transaction binding the contract method 0x41bd64ba.
//
// Solidity: function updateTemporalNumericValuesV1(((uint64,int192),bytes32,bytes32,bytes32,bytes32,bytes32,uint8)[] updateData) payable returns()
func (_StorkContract *StorkContractTransactor) UpdateTemporalNumericValuesV1(opts *bind.TransactOpts, updateData []StorkStructsTemporalNumericValueInput) (*types.Transaction, error) {
	return _StorkContract.contract.Transact(opts, "updateTemporalNumericValuesV1", updateData)
}

// UpdateTemporalNumericValuesV1 is a paid mutator transaction binding the contract method 0x41bd64ba.
//
// Solidity: function updateTemporalNumericValuesV1(((uint64,int192),bytes32,bytes32,bytes32,bytes32,bytes32,uint8)[] updateData) payable returns()
func (_StorkContract *StorkContractSession) UpdateTemporalNumericValuesV1(updateData []StorkStructsTemporalNumericValueInput) (*types.Transaction, error) {
	return _StorkContract.Contract.UpdateTemporalNumericValuesV1(&_StorkContract.TransactOpts, updateData)
}

// UpdateTemporalNumericValuesV1 is a paid mutator transaction binding the contract method 0x41bd64ba.
//
// Solidity: function updateTemporalNumericValuesV1(((uint64,int192),bytes32,bytes32,bytes32,bytes32,bytes32,uint8)[] updateData) payable returns()
func (_StorkContract *StorkContractTransactorSession) UpdateTemporalNumericValuesV1(updateData []StorkStructsTemporalNumericValueInput) (*types.Transaction, error) {
	return _StorkContract.Contract.UpdateTemporalNumericValuesV1(&_StorkContract.TransactOpts, updateData)
}

// UpdateValidTimePeriodSeconds is a paid mutator transaction binding the contract method 0x785bdda0.
//
// Solidity: function updateValidTimePeriodSeconds(uint256 validTimePeriodSeconds) returns()
func (_StorkContract *StorkContractTransactor) UpdateValidTimePeriodSeconds(opts *bind.TransactOpts, validTimePeriodSeconds *big.Int) (*types.Transaction, error) {
	return _StorkContract.contract.Transact(opts, "updateValidTimePeriodSeconds", validTimePeriodSeconds)
}

// UpdateValidTimePeriodSeconds is a paid mutator transaction binding the contract method 0x785bdda0.
//
// Solidity: function updateValidTimePeriodSeconds(uint256 validTimePeriodSeconds) returns()
func (_StorkContract *StorkContractSession) UpdateValidTimePeriodSeconds(validTimePeriodSeconds *big.Int) (*types.Transaction, error) {
	return _StorkContract.Contract.UpdateValidTimePeriodSeconds(&_StorkContract.TransactOpts, validTimePeriodSeconds)
}

// UpdateValidTimePeriodSeconds is a paid mutator transaction binding the contract method 0x785bdda0.
//
// Solidity: function updateValidTimePeriodSeconds(uint256 validTimePeriodSeconds) returns()
func (_StorkContract *StorkContractTransactorSession) UpdateValidTimePeriodSeconds(validTimePeriodSeconds *big.Int) (*types.Transaction, error) {
	return _StorkContract.Contract.UpdateValidTimePeriodSeconds(&_StorkContract.TransactOpts, validTimePeriodSeconds)
}

// UpgradeToAndCall is a paid mutator transaction binding the contract method 0x4f1ef286.
//
// Solidity: function upgradeToAndCall(address newImplementation, bytes data) payable returns()
func (_StorkContract *StorkContractTransactor) UpgradeToAndCall(opts *bind.TransactOpts, newImplementation common.Address, data []byte) (*types.Transaction, error) {
	return _StorkContract.contract.Transact(opts, "upgradeToAndCall", newImplementation, data)
}

// UpgradeToAndCall is a paid mutator transaction binding the contract method 0x4f1ef286.
//
// Solidity: function upgradeToAndCall(address newImplementation, bytes data) payable returns()
func (_StorkContract *StorkContractSession) UpgradeToAndCall(newImplementation common.Address, data []byte) (*types.Transaction, error) {
	return _StorkContract.Contract.UpgradeToAndCall(&_StorkContract.TransactOpts, newImplementation, data)
}

// UpgradeToAndCall is a paid mutator transaction binding the contract method 0x4f1ef286.
//
// Solidity: function upgradeToAndCall(address newImplementation, bytes data) payable returns()
func (_StorkContract *StorkContractTransactorSession) UpgradeToAndCall(newImplementation common.Address, data []byte) (*types.Transaction, error) {
	return _StorkContract.Contract.UpgradeToAndCall(&_StorkContract.TransactOpts, newImplementation, data)
}

// StorkContractInitializedIterator is returned from FilterInitialized and is used to iterate over the raw logs and unpacked data for Initialized events raised by the StorkContract contract.
type StorkContractInitializedIterator struct {
	Event *StorkContractInitialized // Event containing the contract specifics and raw log

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
func (it *StorkContractInitializedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(StorkContractInitialized)
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
		it.Event = new(StorkContractInitialized)
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
func (it *StorkContractInitializedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *StorkContractInitializedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// StorkContractInitialized represents a Initialized event raised by the StorkContract contract.
type StorkContractInitialized struct {
	Version uint64
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterInitialized is a free log retrieval operation binding the contract event 0xc7f505b2f371ae2175ee4913f4499e1f2633a7b5936321eed1cdaeb6115181d2.
//
// Solidity: event Initialized(uint64 version)
func (_StorkContract *StorkContractFilterer) FilterInitialized(opts *bind.FilterOpts) (*StorkContractInitializedIterator, error) {

	logs, sub, err := _StorkContract.contract.FilterLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return &StorkContractInitializedIterator{contract: _StorkContract.contract, event: "Initialized", logs: logs, sub: sub}, nil
}

// WatchInitialized is a free log subscription operation binding the contract event 0xc7f505b2f371ae2175ee4913f4499e1f2633a7b5936321eed1cdaeb6115181d2.
//
// Solidity: event Initialized(uint64 version)
func (_StorkContract *StorkContractFilterer) WatchInitialized(opts *bind.WatchOpts, sink chan<- *StorkContractInitialized) (event.Subscription, error) {

	logs, sub, err := _StorkContract.contract.WatchLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(StorkContractInitialized)
				if err := _StorkContract.contract.UnpackLog(event, "Initialized", log); err != nil {
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

// ParseInitialized is a log parse operation binding the contract event 0xc7f505b2f371ae2175ee4913f4499e1f2633a7b5936321eed1cdaeb6115181d2.
//
// Solidity: event Initialized(uint64 version)
func (_StorkContract *StorkContractFilterer) ParseInitialized(log types.Log) (*StorkContractInitialized, error) {
	event := new(StorkContractInitialized)
	if err := _StorkContract.contract.UnpackLog(event, "Initialized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// StorkContractOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the StorkContract contract.
type StorkContractOwnershipTransferredIterator struct {
	Event *StorkContractOwnershipTransferred // Event containing the contract specifics and raw log

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
func (it *StorkContractOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(StorkContractOwnershipTransferred)
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
		it.Event = new(StorkContractOwnershipTransferred)
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
func (it *StorkContractOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *StorkContractOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// StorkContractOwnershipTransferred represents a OwnershipTransferred event raised by the StorkContract contract.
type StorkContractOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_StorkContract *StorkContractFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*StorkContractOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _StorkContract.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &StorkContractOwnershipTransferredIterator{contract: _StorkContract.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_StorkContract *StorkContractFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *StorkContractOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _StorkContract.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(StorkContractOwnershipTransferred)
				if err := _StorkContract.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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

// ParseOwnershipTransferred is a log parse operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_StorkContract *StorkContractFilterer) ParseOwnershipTransferred(log types.Log) (*StorkContractOwnershipTransferred, error) {
	event := new(StorkContractOwnershipTransferred)
	if err := _StorkContract.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// StorkContractUpgradedIterator is returned from FilterUpgraded and is used to iterate over the raw logs and unpacked data for Upgraded events raised by the StorkContract contract.
type StorkContractUpgradedIterator struct {
	Event *StorkContractUpgraded // Event containing the contract specifics and raw log

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
func (it *StorkContractUpgradedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(StorkContractUpgraded)
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
		it.Event = new(StorkContractUpgraded)
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
func (it *StorkContractUpgradedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *StorkContractUpgradedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// StorkContractUpgraded represents a Upgraded event raised by the StorkContract contract.
type StorkContractUpgraded struct {
	Implementation common.Address
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterUpgraded is a free log retrieval operation binding the contract event 0xbc7cd75a20ee27fd9adebab32041f755214dbc6bffa90cc0225b39da2e5c2d3b.
//
// Solidity: event Upgraded(address indexed implementation)
func (_StorkContract *StorkContractFilterer) FilterUpgraded(opts *bind.FilterOpts, implementation []common.Address) (*StorkContractUpgradedIterator, error) {

	var implementationRule []interface{}
	for _, implementationItem := range implementation {
		implementationRule = append(implementationRule, implementationItem)
	}

	logs, sub, err := _StorkContract.contract.FilterLogs(opts, "Upgraded", implementationRule)
	if err != nil {
		return nil, err
	}
	return &StorkContractUpgradedIterator{contract: _StorkContract.contract, event: "Upgraded", logs: logs, sub: sub}, nil
}

// WatchUpgraded is a free log subscription operation binding the contract event 0xbc7cd75a20ee27fd9adebab32041f755214dbc6bffa90cc0225b39da2e5c2d3b.
//
// Solidity: event Upgraded(address indexed implementation)
func (_StorkContract *StorkContractFilterer) WatchUpgraded(opts *bind.WatchOpts, sink chan<- *StorkContractUpgraded, implementation []common.Address) (event.Subscription, error) {

	var implementationRule []interface{}
	for _, implementationItem := range implementation {
		implementationRule = append(implementationRule, implementationItem)
	}

	logs, sub, err := _StorkContract.contract.WatchLogs(opts, "Upgraded", implementationRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(StorkContractUpgraded)
				if err := _StorkContract.contract.UnpackLog(event, "Upgraded", log); err != nil {
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

// ParseUpgraded is a log parse operation binding the contract event 0xbc7cd75a20ee27fd9adebab32041f755214dbc6bffa90cc0225b39da2e5c2d3b.
//
// Solidity: event Upgraded(address indexed implementation)
func (_StorkContract *StorkContractFilterer) ParseUpgraded(log types.Log) (*StorkContractUpgraded, error) {
	event := new(StorkContractUpgraded)
	if err := _StorkContract.contract.UnpackLog(event, "Upgraded", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// StorkContractValueUpdateIterator is returned from FilterValueUpdate and is used to iterate over the raw logs and unpacked data for ValueUpdate events raised by the StorkContract contract.
type StorkContractValueUpdateIterator struct {
	Event *StorkContractValueUpdate // Event containing the contract specifics and raw log

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
func (it *StorkContractValueUpdateIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(StorkContractValueUpdate)
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
		it.Event = new(StorkContractValueUpdate)
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
func (it *StorkContractValueUpdateIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *StorkContractValueUpdateIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// StorkContractValueUpdate represents a ValueUpdate event raised by the StorkContract contract.
type StorkContractValueUpdate struct {
	Id             [32]byte
	TimestampNs    uint64
	QuantizedValue *big.Int
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterValueUpdate is a free log retrieval operation binding the contract event 0xe24720f45cb74f2d55f1deebb6098f50f10b511dab8a7d47c4819a08dcd0b895.
//
// Solidity: event ValueUpdate(bytes32 indexed id, uint64 timestampNs, int192 quantizedValue)
func (_StorkContract *StorkContractFilterer) FilterValueUpdate(opts *bind.FilterOpts, id [][32]byte) (*StorkContractValueUpdateIterator, error) {

	var idRule []interface{}
	for _, idItem := range id {
		idRule = append(idRule, idItem)
	}

	logs, sub, err := _StorkContract.contract.FilterLogs(opts, "ValueUpdate", idRule)
	if err != nil {
		return nil, err
	}
	return &StorkContractValueUpdateIterator{contract: _StorkContract.contract, event: "ValueUpdate", logs: logs, sub: sub}, nil
}

// WatchValueUpdate is a free log subscription operation binding the contract event 0xe24720f45cb74f2d55f1deebb6098f50f10b511dab8a7d47c4819a08dcd0b895.
//
// Solidity: event ValueUpdate(bytes32 indexed id, uint64 timestampNs, int192 quantizedValue)
func (_StorkContract *StorkContractFilterer) WatchValueUpdate(opts *bind.WatchOpts, sink chan<- *StorkContractValueUpdate, id [][32]byte) (event.Subscription, error) {

	var idRule []interface{}
	for _, idItem := range id {
		idRule = append(idRule, idItem)
	}

	logs, sub, err := _StorkContract.contract.WatchLogs(opts, "ValueUpdate", idRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(StorkContractValueUpdate)
				if err := _StorkContract.contract.UnpackLog(event, "ValueUpdate", log); err != nil {
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

// ParseValueUpdate is a log parse operation binding the contract event 0xe24720f45cb74f2d55f1deebb6098f50f10b511dab8a7d47c4819a08dcd0b895.
//
// Solidity: event ValueUpdate(bytes32 indexed id, uint64 timestampNs, int192 quantizedValue)
func (_StorkContract *StorkContractFilterer) ParseValueUpdate(log types.Log) (*StorkContractValueUpdate, error) {
	event := new(StorkContractValueUpdate)
	if err := _StorkContract.contract.UnpackLog(event, "ValueUpdate", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
