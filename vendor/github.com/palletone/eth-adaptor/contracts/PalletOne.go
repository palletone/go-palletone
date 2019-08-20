// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package PalletOne

import (
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// ITokenABI is the input ABI used to generate the binding from.
const ITokenABI = "[{\"constant\":true,\"inputs\":[],\"name\":\"name\",\"outputs\":[{\"name\":\"\",\"type\":\"string\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_spender\",\"type\":\"address\"},{\"name\":\"_value\",\"type\":\"uint256\"}],\"name\":\"approve\",\"outputs\":[{\"name\":\"success\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"totalSupply\",\"outputs\":[{\"name\":\"supply\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_from\",\"type\":\"address\"},{\"name\":\"_to\",\"type\":\"address\"},{\"name\":\"_value\",\"type\":\"uint256\"}],\"name\":\"transferFrom\",\"outputs\":[{\"name\":\"success\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"decimals\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_owner\",\"type\":\"address\"}],\"name\":\"balanceOf\",\"outputs\":[{\"name\":\"balance\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_to\",\"type\":\"address\"},{\"name\":\"_value\",\"type\":\"uint256\"}],\"name\":\"transfer\",\"outputs\":[{\"name\":\"success\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_owner\",\"type\":\"address\"},{\"name\":\"_spender\",\"type\":\"address\"}],\"name\":\"allowance\",\"outputs\":[{\"name\":\"remaining\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"_from\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"_to\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"_value\",\"type\":\"uint256\"}],\"name\":\"Transfer\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"_owner\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"_spender\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"_value\",\"type\":\"uint256\"}],\"name\":\"Approval\",\"type\":\"event\"}]"

// ITokenBin is the compiled bytecode used for deploying new contracts.
const ITokenBin = `0x608060405234801561001057600080fd5b50610311806100206000396000f30060806040526004361061008d5763ffffffff7c010000000000000000000000000000000000000000000000000000000060003504166306fdde038114610092578063095ea7b31461011c57806318160ddd1461016157806323b872dd14610188578063313ce567146101bf57806370a08231146101d4578063a9059cbb1461011c578063dd62ed3e14610202575b600080fd5b34801561009e57600080fd5b506100a7610236565b6040805160208082528351818301528351919283929083019185019080838360005b838110156100e15781810151838201526020016100c9565b50505050905090810190601f16801561010e5780820380516001836020036101000a031916815260200191505b509250505060405180910390f35b34801561012857600080fd5b5061014d73ffffffffffffffffffffffffffffffffffffffff600435166024356102c3565b604080519115158252519081900360200190f35b34801561016d57600080fd5b506101766102cb565b60408051918252519081900360200190f35b34801561019457600080fd5b5061014d73ffffffffffffffffffffffffffffffffffffffff600435811690602435166044356102d0565b3480156101cb57600080fd5b506101766102d9565b3480156101e057600080fd5b5061017673ffffffffffffffffffffffffffffffffffffffff600435166102df565b34801561020e57600080fd5b5061017673ffffffffffffffffffffffffffffffffffffffff600435811690602435166102c3565b60018054604080516020600284861615610100026000190190941693909304601f810184900484028201840190925281815292918301828280156102bb5780601f10610290576101008083540402835291602001916102bb565b820191906000526020600020905b81548152906001019060200180831161029e57829003601f168201915b505050505081565b600092915050565b600090565b60009392505050565b60005481565b506000905600a165627a7a7230582083adeb3b1ae323d13f2531e6d80f0e40870de73f1ecf2a882dabd007c736b1860029`

// DeployIToken deploys a new Ethereum contract, binding an instance of IToken to it.
func DeployIToken(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *IToken, error) {
	parsed, err := abi.JSON(strings.NewReader(ITokenABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(ITokenBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &IToken{ITokenCaller: ITokenCaller{contract: contract}, ITokenTransactor: ITokenTransactor{contract: contract}, ITokenFilterer: ITokenFilterer{contract: contract}}, nil
}

// IToken is an auto generated Go binding around an Ethereum contract.
type IToken struct {
	ITokenCaller     // Read-only binding to the contract
	ITokenTransactor // Write-only binding to the contract
	ITokenFilterer   // Log filterer for contract events
}

// ITokenCaller is an auto generated read-only Go binding around an Ethereum contract.
type ITokenCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ITokenTransactor is an auto generated write-only Go binding around an Ethereum contract.
type ITokenTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ITokenFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type ITokenFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ITokenSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type ITokenSession struct {
	Contract     *IToken           // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// ITokenCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type ITokenCallerSession struct {
	Contract *ITokenCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts // Call options to use throughout this session
}

// ITokenTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type ITokenTransactorSession struct {
	Contract     *ITokenTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// ITokenRaw is an auto generated low-level Go binding around an Ethereum contract.
type ITokenRaw struct {
	Contract *IToken // Generic contract binding to access the raw methods on
}

// ITokenCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type ITokenCallerRaw struct {
	Contract *ITokenCaller // Generic read-only contract binding to access the raw methods on
}

// ITokenTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type ITokenTransactorRaw struct {
	Contract *ITokenTransactor // Generic write-only contract binding to access the raw methods on
}

// NewIToken creates a new instance of IToken, bound to a specific deployed contract.
func NewIToken(address common.Address, backend bind.ContractBackend) (*IToken, error) {
	contract, err := bindIToken(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &IToken{ITokenCaller: ITokenCaller{contract: contract}, ITokenTransactor: ITokenTransactor{contract: contract}, ITokenFilterer: ITokenFilterer{contract: contract}}, nil
}

// NewITokenCaller creates a new read-only instance of IToken, bound to a specific deployed contract.
func NewITokenCaller(address common.Address, caller bind.ContractCaller) (*ITokenCaller, error) {
	contract, err := bindIToken(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ITokenCaller{contract: contract}, nil
}

// NewITokenTransactor creates a new write-only instance of IToken, bound to a specific deployed contract.
func NewITokenTransactor(address common.Address, transactor bind.ContractTransactor) (*ITokenTransactor, error) {
	contract, err := bindIToken(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ITokenTransactor{contract: contract}, nil
}

// NewITokenFilterer creates a new log filterer instance of IToken, bound to a specific deployed contract.
func NewITokenFilterer(address common.Address, filterer bind.ContractFilterer) (*ITokenFilterer, error) {
	contract, err := bindIToken(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ITokenFilterer{contract: contract}, nil
}

// bindIToken binds a generic wrapper to an already deployed contract.
func bindIToken(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(ITokenABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_IToken *ITokenRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _IToken.Contract.ITokenCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_IToken *ITokenRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _IToken.Contract.ITokenTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_IToken *ITokenRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _IToken.Contract.ITokenTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_IToken *ITokenCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _IToken.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_IToken *ITokenTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _IToken.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_IToken *ITokenTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _IToken.Contract.contract.Transact(opts, method, params...)
}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(_owner address, _spender address) constant returns(remaining uint256)
func (_IToken *ITokenCaller) Allowance(opts *bind.CallOpts, _owner common.Address, _spender common.Address) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _IToken.contract.Call(opts, out, "allowance", _owner, _spender)
	return *ret0, err
}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(_owner address, _spender address) constant returns(remaining uint256)
func (_IToken *ITokenSession) Allowance(_owner common.Address, _spender common.Address) (*big.Int, error) {
	return _IToken.Contract.Allowance(&_IToken.CallOpts, _owner, _spender)
}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(_owner address, _spender address) constant returns(remaining uint256)
func (_IToken *ITokenCallerSession) Allowance(_owner common.Address, _spender common.Address) (*big.Int, error) {
	return _IToken.Contract.Allowance(&_IToken.CallOpts, _owner, _spender)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(_owner address) constant returns(balance uint256)
func (_IToken *ITokenCaller) BalanceOf(opts *bind.CallOpts, _owner common.Address) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _IToken.contract.Call(opts, out, "balanceOf", _owner)
	return *ret0, err
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(_owner address) constant returns(balance uint256)
func (_IToken *ITokenSession) BalanceOf(_owner common.Address) (*big.Int, error) {
	return _IToken.Contract.BalanceOf(&_IToken.CallOpts, _owner)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(_owner address) constant returns(balance uint256)
func (_IToken *ITokenCallerSession) BalanceOf(_owner common.Address) (*big.Int, error) {
	return _IToken.Contract.BalanceOf(&_IToken.CallOpts, _owner)
}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() constant returns(uint256)
func (_IToken *ITokenCaller) Decimals(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _IToken.contract.Call(opts, out, "decimals")
	return *ret0, err
}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() constant returns(uint256)
func (_IToken *ITokenSession) Decimals() (*big.Int, error) {
	return _IToken.Contract.Decimals(&_IToken.CallOpts)
}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() constant returns(uint256)
func (_IToken *ITokenCallerSession) Decimals() (*big.Int, error) {
	return _IToken.Contract.Decimals(&_IToken.CallOpts)
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() constant returns(string)
func (_IToken *ITokenCaller) Name(opts *bind.CallOpts) (string, error) {
	var (
		ret0 = new(string)
	)
	out := ret0
	err := _IToken.contract.Call(opts, out, "name")
	return *ret0, err
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() constant returns(string)
func (_IToken *ITokenSession) Name() (string, error) {
	return _IToken.Contract.Name(&_IToken.CallOpts)
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() constant returns(string)
func (_IToken *ITokenCallerSession) Name() (string, error) {
	return _IToken.Contract.Name(&_IToken.CallOpts)
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() constant returns(supply uint256)
func (_IToken *ITokenCaller) TotalSupply(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _IToken.contract.Call(opts, out, "totalSupply")
	return *ret0, err
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() constant returns(supply uint256)
func (_IToken *ITokenSession) TotalSupply() (*big.Int, error) {
	return _IToken.Contract.TotalSupply(&_IToken.CallOpts)
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() constant returns(supply uint256)
func (_IToken *ITokenCallerSession) TotalSupply() (*big.Int, error) {
	return _IToken.Contract.TotalSupply(&_IToken.CallOpts)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(_spender address, _value uint256) returns(success bool)
func (_IToken *ITokenTransactor) Approve(opts *bind.TransactOpts, _spender common.Address, _value *big.Int) (*types.Transaction, error) {
	return _IToken.contract.Transact(opts, "approve", _spender, _value)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(_spender address, _value uint256) returns(success bool)
func (_IToken *ITokenSession) Approve(_spender common.Address, _value *big.Int) (*types.Transaction, error) {
	return _IToken.Contract.Approve(&_IToken.TransactOpts, _spender, _value)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(_spender address, _value uint256) returns(success bool)
func (_IToken *ITokenTransactorSession) Approve(_spender common.Address, _value *big.Int) (*types.Transaction, error) {
	return _IToken.Contract.Approve(&_IToken.TransactOpts, _spender, _value)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(_to address, _value uint256) returns(success bool)
func (_IToken *ITokenTransactor) Transfer(opts *bind.TransactOpts, _to common.Address, _value *big.Int) (*types.Transaction, error) {
	return _IToken.contract.Transact(opts, "transfer", _to, _value)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(_to address, _value uint256) returns(success bool)
func (_IToken *ITokenSession) Transfer(_to common.Address, _value *big.Int) (*types.Transaction, error) {
	return _IToken.Contract.Transfer(&_IToken.TransactOpts, _to, _value)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(_to address, _value uint256) returns(success bool)
func (_IToken *ITokenTransactorSession) Transfer(_to common.Address, _value *big.Int) (*types.Transaction, error) {
	return _IToken.Contract.Transfer(&_IToken.TransactOpts, _to, _value)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(_from address, _to address, _value uint256) returns(success bool)
func (_IToken *ITokenTransactor) TransferFrom(opts *bind.TransactOpts, _from common.Address, _to common.Address, _value *big.Int) (*types.Transaction, error) {
	return _IToken.contract.Transact(opts, "transferFrom", _from, _to, _value)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(_from address, _to address, _value uint256) returns(success bool)
func (_IToken *ITokenSession) TransferFrom(_from common.Address, _to common.Address, _value *big.Int) (*types.Transaction, error) {
	return _IToken.Contract.TransferFrom(&_IToken.TransactOpts, _from, _to, _value)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(_from address, _to address, _value uint256) returns(success bool)
func (_IToken *ITokenTransactorSession) TransferFrom(_from common.Address, _to common.Address, _value *big.Int) (*types.Transaction, error) {
	return _IToken.Contract.TransferFrom(&_IToken.TransactOpts, _from, _to, _value)
}

// ITokenApprovalIterator is returned from FilterApproval and is used to iterate over the raw logs and unpacked data for Approval events raised by the IToken contract.
type ITokenApprovalIterator struct {
	Event *ITokenApproval // Event containing the contract specifics and raw log

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
func (it *ITokenApprovalIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ITokenApproval)
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
		it.Event = new(ITokenApproval)
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
func (it *ITokenApprovalIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ITokenApprovalIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ITokenApproval represents a Approval event raised by the IToken contract.
type ITokenApproval struct {
	Owner   common.Address
	Spender common.Address
	Value   *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterApproval is a free log retrieval operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: e Approval(_owner indexed address, _spender indexed address, _value uint256)
func (_IToken *ITokenFilterer) FilterApproval(opts *bind.FilterOpts, _owner []common.Address, _spender []common.Address) (*ITokenApprovalIterator, error) {

	var _ownerRule []interface{}
	for _, _ownerItem := range _owner {
		_ownerRule = append(_ownerRule, _ownerItem)
	}
	var _spenderRule []interface{}
	for _, _spenderItem := range _spender {
		_spenderRule = append(_spenderRule, _spenderItem)
	}

	logs, sub, err := _IToken.contract.FilterLogs(opts, "Approval", _ownerRule, _spenderRule)
	if err != nil {
		return nil, err
	}
	return &ITokenApprovalIterator{contract: _IToken.contract, event: "Approval", logs: logs, sub: sub}, nil
}

// WatchApproval is a free log subscription operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: e Approval(_owner indexed address, _spender indexed address, _value uint256)
func (_IToken *ITokenFilterer) WatchApproval(opts *bind.WatchOpts, sink chan<- *ITokenApproval, _owner []common.Address, _spender []common.Address) (event.Subscription, error) {

	var _ownerRule []interface{}
	for _, _ownerItem := range _owner {
		_ownerRule = append(_ownerRule, _ownerItem)
	}
	var _spenderRule []interface{}
	for _, _spenderItem := range _spender {
		_spenderRule = append(_spenderRule, _spenderItem)
	}

	logs, sub, err := _IToken.contract.WatchLogs(opts, "Approval", _ownerRule, _spenderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ITokenApproval)
				if err := _IToken.contract.UnpackLog(event, "Approval", log); err != nil {
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

// ITokenTransferIterator is returned from FilterTransfer and is used to iterate over the raw logs and unpacked data for Transfer events raised by the IToken contract.
type ITokenTransferIterator struct {
	Event *ITokenTransfer // Event containing the contract specifics and raw log

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
func (it *ITokenTransferIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ITokenTransfer)
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
		it.Event = new(ITokenTransfer)
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
func (it *ITokenTransferIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ITokenTransferIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ITokenTransfer represents a Transfer event raised by the IToken contract.
type ITokenTransfer struct {
	From  common.Address
	To    common.Address
	Value *big.Int
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterTransfer is a free log retrieval operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: e Transfer(_from indexed address, _to indexed address, _value uint256)
func (_IToken *ITokenFilterer) FilterTransfer(opts *bind.FilterOpts, _from []common.Address, _to []common.Address) (*ITokenTransferIterator, error) {

	var _fromRule []interface{}
	for _, _fromItem := range _from {
		_fromRule = append(_fromRule, _fromItem)
	}
	var _toRule []interface{}
	for _, _toItem := range _to {
		_toRule = append(_toRule, _toItem)
	}

	logs, sub, err := _IToken.contract.FilterLogs(opts, "Transfer", _fromRule, _toRule)
	if err != nil {
		return nil, err
	}
	return &ITokenTransferIterator{contract: _IToken.contract, event: "Transfer", logs: logs, sub: sub}, nil
}

// WatchTransfer is a free log subscription operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: e Transfer(_from indexed address, _to indexed address, _value uint256)
func (_IToken *ITokenFilterer) WatchTransfer(opts *bind.WatchOpts, sink chan<- *ITokenTransfer, _from []common.Address, _to []common.Address) (event.Subscription, error) {

	var _fromRule []interface{}
	for _, _fromItem := range _from {
		_fromRule = append(_fromRule, _fromItem)
	}
	var _toRule []interface{}
	for _, _toItem := range _to {
		_toRule = append(_toRule, _toItem)
	}

	logs, sub, err := _IToken.contract.WatchLogs(opts, "Transfer", _fromRule, _toRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ITokenTransfer)
				if err := _IToken.contract.UnpackLog(event, "Transfer", log); err != nil {
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

// LSafeMathABI is the input ABI used to generate the binding from.
const LSafeMathABI = "[]"

// LSafeMathBin is the compiled bytecode used for deploying new contracts.
const LSafeMathBin = `0x604c602c600b82828239805160001a60731460008114601c57601e565bfe5b5030600052607381538281f30073000000000000000000000000000000000000000030146080604052600080fd00a165627a7a72305820e74d020d9213419188e1392adeeefd19090601f95ee04f42eb6d7a38ba136b630029`

// DeployLSafeMath deploys a new Ethereum contract, binding an instance of LSafeMath to it.
func DeployLSafeMath(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *LSafeMath, error) {
	parsed, err := abi.JSON(strings.NewReader(LSafeMathABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(LSafeMathBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &LSafeMath{LSafeMathCaller: LSafeMathCaller{contract: contract}, LSafeMathTransactor: LSafeMathTransactor{contract: contract}, LSafeMathFilterer: LSafeMathFilterer{contract: contract}}, nil
}

// LSafeMath is an auto generated Go binding around an Ethereum contract.
type LSafeMath struct {
	LSafeMathCaller     // Read-only binding to the contract
	LSafeMathTransactor // Write-only binding to the contract
	LSafeMathFilterer   // Log filterer for contract events
}

// LSafeMathCaller is an auto generated read-only Go binding around an Ethereum contract.
type LSafeMathCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// LSafeMathTransactor is an auto generated write-only Go binding around an Ethereum contract.
type LSafeMathTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// LSafeMathFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type LSafeMathFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// LSafeMathSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type LSafeMathSession struct {
	Contract     *LSafeMath        // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// LSafeMathCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type LSafeMathCallerSession struct {
	Contract *LSafeMathCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts    // Call options to use throughout this session
}

// LSafeMathTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type LSafeMathTransactorSession struct {
	Contract     *LSafeMathTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts    // Transaction auth options to use throughout this session
}

// LSafeMathRaw is an auto generated low-level Go binding around an Ethereum contract.
type LSafeMathRaw struct {
	Contract *LSafeMath // Generic contract binding to access the raw methods on
}

// LSafeMathCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type LSafeMathCallerRaw struct {
	Contract *LSafeMathCaller // Generic read-only contract binding to access the raw methods on
}

// LSafeMathTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type LSafeMathTransactorRaw struct {
	Contract *LSafeMathTransactor // Generic write-only contract binding to access the raw methods on
}

// NewLSafeMath creates a new instance of LSafeMath, bound to a specific deployed contract.
func NewLSafeMath(address common.Address, backend bind.ContractBackend) (*LSafeMath, error) {
	contract, err := bindLSafeMath(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &LSafeMath{LSafeMathCaller: LSafeMathCaller{contract: contract}, LSafeMathTransactor: LSafeMathTransactor{contract: contract}, LSafeMathFilterer: LSafeMathFilterer{contract: contract}}, nil
}

// NewLSafeMathCaller creates a new read-only instance of LSafeMath, bound to a specific deployed contract.
func NewLSafeMathCaller(address common.Address, caller bind.ContractCaller) (*LSafeMathCaller, error) {
	contract, err := bindLSafeMath(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &LSafeMathCaller{contract: contract}, nil
}

// NewLSafeMathTransactor creates a new write-only instance of LSafeMath, bound to a specific deployed contract.
func NewLSafeMathTransactor(address common.Address, transactor bind.ContractTransactor) (*LSafeMathTransactor, error) {
	contract, err := bindLSafeMath(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &LSafeMathTransactor{contract: contract}, nil
}

// NewLSafeMathFilterer creates a new log filterer instance of LSafeMath, bound to a specific deployed contract.
func NewLSafeMathFilterer(address common.Address, filterer bind.ContractFilterer) (*LSafeMathFilterer, error) {
	contract, err := bindLSafeMath(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &LSafeMathFilterer{contract: contract}, nil
}

// bindLSafeMath binds a generic wrapper to an already deployed contract.
func bindLSafeMath(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(LSafeMathABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_LSafeMath *LSafeMathRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _LSafeMath.Contract.LSafeMathCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_LSafeMath *LSafeMathRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _LSafeMath.Contract.LSafeMathTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_LSafeMath *LSafeMathRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _LSafeMath.Contract.LSafeMathTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_LSafeMath *LSafeMathCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _LSafeMath.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_LSafeMath *LSafeMathTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _LSafeMath.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_LSafeMath *LSafeMathTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _LSafeMath.Contract.contract.Transact(opts, method, params...)
}

// PalletOneABI is the input ABI used to generate the binding from.
const PalletOneABI = "[{\"constant\":false,\"inputs\":[{\"name\":\"token\",\"type\":\"address\"},{\"name\":\"redeem\",\"type\":\"bytes\"},{\"name\":\"recver\",\"type\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\"},{\"name\":\"nonece\",\"type\":\"uint256\"},{\"name\":\"sigstr1\",\"type\":\"bytes\"},{\"name\":\"sigstr2\",\"type\":\"bytes\"}],\"name\":\"withdrawtoken\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"redeem\",\"type\":\"bytes\"},{\"name\":\"recver\",\"type\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\"},{\"name\":\"nonece\",\"type\":\"uint256\"},{\"name\":\"sigstr1\",\"type\":\"bytes\"},{\"name\":\"sigstr2\",\"type\":\"bytes\"}],\"name\":\"withdraw\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"suicideto\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"redeem\",\"type\":\"bytes\"}],\"name\":\"deposit\",\"outputs\":[],\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"address\"},{\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"tokens\",\"outputs\":[{\"name\":\"balance\",\"type\":\"uint256\"},{\"name\":\"nonece\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"token\",\"type\":\"address\"},{\"name\":\"redeem\",\"type\":\"bytes\"},{\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"deposittoken\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"my_eth_bal\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"addr\",\"type\":\"address\"},{\"name\":\"redeem\",\"type\":\"bytes\"}],\"name\":\"getmultisig\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"},{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"admin\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"name\":\"admin_\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"fallback\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"token\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"user\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"amount\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"redeem\",\"type\":\"bytes\"}],\"name\":\"Deposit\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"token\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"user\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"redeem\",\"type\":\"bytes\"},{\"indexed\":false,\"name\":\"recver\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"amount\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"confirmvalue\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"state\",\"type\":\"string\"}],\"name\":\"Withdraw\",\"type\":\"event\"}]"

// PalletOneBin is the compiled bytecode used for deploying new contracts.
const PalletOneBin = `0x608060405234801561001057600080fd5b50604051602080611868833981016040525160008054600160a060020a03909216600160a060020a0319909216919091179055611816806100526000396000f3006080604052600436106100985763ffffffff7c010000000000000000000000000000000000000000000000000000000060003504166333890eca81146100aa5780638c2e0321146101aa5780638e644ec31461029a57806398b1e06a146102bb578063a9964d1c14610307578063c0c6cf4e14610344578063c8fc638a146103ad578063e7a64ff2146103d4578063f851a4401461043b575b3480156100a457600080fd5b50600080fd5b3480156100b657600080fd5b5060408051602060046024803582810135601f81018590048502860185019096528585526101a8958335600160a060020a031695369560449491939091019190819084018382808284375050604080516020601f60608a01358b0180359182018390048302840183018552818452989b600160a060020a038b35169b838c01359b958601359a91995097506080909401955091935091820191819084018382808284375050604080516020601f89358b018035918201839004830284018301909452808352979a99988101979196509182019450925082915084018382808284375094975061046c9650505050505050565b005b3480156101b657600080fd5b506040805160206004803580820135601f81018490048402850184019095528484526101a894369492936024939284019190819084018382808284375050604080516020601f60608a01358b0180359182018390048302840183018552818452989b600160a060020a038b35169b838c01359b958601359a91995097506080909401955091935091820191819084018382808284375050604080516020601f89358b018035918201839004830284018301909452808352979a9998810197919650918201945092508291508401838280828437509497506109619650505050505050565b3480156102a657600080fd5b506101a8600160a060020a0360043516610de7565b6040805160206004803580820135601f81018490048402850184019095528484526101a8943694929360249392840191908190840183828082843750949750610e0a9650505050505050565b34801561031357600080fd5b5061032b600160a060020a0360043516602435610fd2565b6040805192835260208301919091528051918290030190f35b34801561035057600080fd5b5060408051602060046024803582810135601f81018590048502860185019096528585526101a8958335600160a060020a03169536956044949193909101919081908401838280828437509497505093359450610ff39350505050565b3480156103b957600080fd5b506103c2611270565b60408051918252519081900360200190f35b3480156103e057600080fd5b5060408051602060046024803582810135601f810185900485028601850190965285855261032b958335600160a060020a03169536956044949193909101919081908401838280828437509497506112759650505050505050565b34801561044757600080fd5b5061045061136d565b60408051600160a060020a039092168252519081900360200190f35b60006060600080896040516020018082805190602001908083835b602083106104a65780518252601f199092019160209182019101610487565b6001836020036101000a0380198251168184511680821785525050505050509050019150506040516020818303038152906040526040518082805190602001908083835b602083106105095780518252601f1990920191602091820191016104ea565b6001836020036101000a0380198251168184511680821785525050505050509050019150506040518091039020935087600160008d600160a060020a0316600160a060020a0316815260200190815260200160002060008660001916600019168152602001908152602001600020600001541015151561058857600080fd5b61059987600163ffffffff61137c16565b600160a060020a038c1660009081526001602081815260408084208985529091529091200154146105c957600080fd5b60408051600680825260e08201909252906020820160c0803883390190505092506105f4838b611391565b600091508a8a8a308b8b6040516020018087600160a060020a0316600160a060020a03166c0100000000000000000000000002815260140186805190602001908083835b602083106106575780518252601f199092019160209182019101610638565b6001836020036101000a03801982511681845116808217855250505050505090500185600160a060020a0316600160a060020a03166c0100000000000000000000000002815260140184600160a060020a0316600160a060020a03166c0100000000000000000000000002815260140183815260200182815260200196505050505050506040516020818303038152906040526040518082805190602001908083835b602083106107195780518252601f1990920191602091820191016106fa565b6001836020036101000a0380198251168184511680821785525050505050509050019150506040518091039020905061075483828888611437565b9150600260ff8316101561076757600080fd5b6107728b858a611482565b8a600160a060020a031663a9059cbb8a8a6040518363ffffffff167c01000000000000000000000000000000000000000000000000000000000281526004018083600160a060020a0316600160a060020a0316815260200182815260200192505050602060405180830381600087803b1580156107ee57600080fd5b505af1158015610802573d6000803e3d6000fd5b505050506040513d602081101561081857600080fd5b5051151561082557600080fd5b7f3c787786801bcab2749cb2c8202e63081bfdb0ef3bc5c9cea89cacd3e7ef4cf38b338c8c8c876040518087600160a060020a0316600160a060020a0316815260200186600160a060020a0316600160a060020a031681526020018060200185600160a060020a0316600160a060020a031681526020018481526020018360ff16815260200180602001838103835287818151815260200191508051906020019080838360005b838110156108e45781810151838201526020016108cc565b50505050905090810190601f1680156109115780820380516001836020036101000a031916815260200191505b50928303905250600d81527f7769746864726177746f6b656e0000000000000000000000000000000000000060208201526040805191829003019650945050505050a15050505050505050505050565b60006060600080896040516020018082805190602001908083835b6020831061099b5780518252601f19909201916020918201910161097c565b6001836020036101000a0380198251168184511680821785525050505050509050019150506040516020818303038152906040526040518082805190602001908083835b602083106109fe5780518252601f1990920191602091820191016109df565b51815160209384036101000a6000190180199092169116179052604080519290940182900390912060008181526000805160206117cb83398151915290925292902054919750508a11159150610a55905057600080fd5b610a6687600163ffffffff61137c16565b60008581526000805160206117cb833981519152602052604090206001015414610a8f57600080fd5b60408051600680825260e08201909252906020820160c080388339019050509250610aba838b611391565b600091508989308a8a6040516020018086805190602001908083835b60208310610af55780518252601f199092019160209182019101610ad6565b6001836020036101000a03801982511681845116808217855250505050505090500185600160a060020a0316600160a060020a03166c0100000000000000000000000002815260140184600160a060020a0316600160a060020a03166c01000000000000000000000000028152601401838152602001828152602001955050505050506040516020818303038152906040526040518082805190602001908083835b60208310610bb65780518252601f199092019160209182019101610b97565b6001836020036101000a03801982511681845116808217855250505050505090500191505060405180910390209050610bf183828888611437565b9150600260ff83161015610c0457600080fd5b60008481526000805160206117cb8339815191526020526040902054610c30908963ffffffff61137c16565b60008581526000805160206117cb83398151915260205260409020908155600190810154610c5d91611519565b60008581526000805160206117cb8339815191526020526040808220600101929092559051600160a060020a038b16918a156108fc02918b91818181858888f19350505050158015610cb3573d6000803e3d6000fd5b507f3c787786801bcab2749cb2c8202e63081bfdb0ef3bc5c9cea89cacd3e7ef4cf36000338c8c8c876040518087600160a060020a0316815260200186600160a060020a0316600160a060020a031681526020018060200185600160a060020a0316600160a060020a031681526020018481526020018360ff16815260200180602001838103835287818151815260200191508051906020019080838360005b83811015610d6b578181015183820152602001610d53565b50505050905090810190601f168015610d985780820380516001836020036101000a031916815260200191505b50928303905250600881527f776974686472617700000000000000000000000000000000000000000000000060208201526040805191829003019650945050505050a150505050505050505050565b600054600160a060020a03163314610dfe57600080fd5b80600160a060020a0316ff5b6000816040516020018082805190602001908083835b60208310610e3f5780518252601f199092019160209182019101610e20565b6001836020036101000a0380198251168184511680821785525050505050509050019150506040516020818303038152906040526040518082805190602001908083835b60208310610ea25780518252601f199092019160209182019101610e83565b51815160209384036101000a6000190180199092169116179052604080519290940182900390912060008181526000805160206117cb83398151915290925292902054919450610ef89350909150349050611519565b60008281526000805160206117cb83398151915260209081526040808320939093558251828152338183018190523494820185905260806060830181815288519184019190915287517fd5d9ab68ad56311de2cda7e56730c5a58bcd4c9d071b9fe5f8efcdb1ccc9251d9692949293899390929160a08401918501908083838b5b83811015610f91578181015183820152602001610f79565b50505050905090810190601f168015610fbe5780820380516001836020036101000a031916815260200191505b509550505050505060405180910390a15050565b60016020818152600093845260408085209091529183529120805491015482565b6000826040516020018082805190602001908083835b602083106110285780518252601f199092019160209182019101611009565b6001836020036101000a0380198251168184511680821785525050505050509050019150506040516020818303038152906040526040518082805190602001908083835b6020831061108b5780518252601f19909201916020918201910161106c565b51815160209384036101000a6000190180199092169116179052604080519290940182900382207f23b872dd000000000000000000000000000000000000000000000000000000008352336004840152306024840152604483018990529351939650600160a060020a038a1695506323b872dd945060648083019491935090918290030181600087803b15801561112157600080fd5b505af1158015611135573d6000803e3d6000fd5b505050506040513d602081101561114b57600080fd5b5051151561115857600080fd5b600160a060020a038416600090815260016020908152604080832084845290915290205461118c908363ffffffff61151916565b600160a060020a038516600081815260016020908152604080832086845282528083209490945583519283523383820181905293830186905260806060840181815288519185019190915287517fd5d9ab68ad56311de2cda7e56730c5a58bcd4c9d071b9fe5f8efcdb1ccc9251d958a95909489948b94929360a0850192918601918190849084905b8381101561122d578181015183820152602001611215565b50505050905090810190601f16801561125a5780820380516001836020036101000a031916815260200191505b509550505050505060405180910390a150505050565b303190565b6000806000836040516020018082805190602001908083835b602083106112ad5780518252601f19909201916020918201910161128e565b6001836020036101000a0380198251168184511680821785525050505050509050019150506040516020818303038152906040526040518082805190602001908083835b602083106113105780518252601f1990920191602091820191016112f1565b51815160209384036101000a60001901801990921691161790526040805192909401829003909120600160a060020a039a909a16600090815260018083528482209b82529a90915291909120805498015497989650505050505050565b600054600160a060020a031681565b60008282111561138b57600080fd5b50900390565b6000806000603c845110156113a557611430565b50505060148101516028820151603c83015184518390869060009081106113c857fe5b600160a060020a0390921660209283029091019091015284518290869060019081106113f057fe5b600160a060020a03909216602092830290910190910152845181908690600290811061141857fe5b600160a060020a039092166020928302909101909101525b5050505050565b60408051600380825260808201909252600091606091839160208201848038833901905050915061146b8287898888611532565b50600061147782611579565b979650505050505050565b600160a060020a03831660009081526001602090815260408083208584529091529020546114b6908263ffffffff61137c16565b600160a060020a0384166000908152600160208181526040808420878552909152909120918255908101546114ea91611519565b600160a060020a0390931660009081526001602081815260408084209584529490529290209091019190915550565b60008282018381101561152b57600080fd5b9392505050565b815160009015611553576115468584611669565b9050611553868286611675565b815115611571576115648583611669565b9050611571868286611675565b505050505050565b60408051600380825260808201909252600091606091839182919060208201858038833901905050925060018360008151811015156115b457fe5b60ff90921660209283029091019091015282516001908490829081106115d657fe5b60ff9092166020928302909101909101528251600190849060029081106115f957fe5b60ff9092166020928302909101909101525060009050805b60038160ff16101561166157828160ff1681518110151561162e57fe5b90602001906020020151858260ff1681518110151561164957fe5b60209081029091010151029190910190600101611611565b509392505050565b600061152b83836116f5565b60005b60038160ff1610156116ef57818160ff1681518110151561169557fe5b90602001906020020151600160a060020a031683600160a060020a03161415156116be576116e7565b6001848260ff168151811015156116d157fe5b60ff9092166020928302909101909101526116ef565b600101611678565b50505050565b6000806000808451604114151561170f57600093506117c1565b50505060208201516040830151606084015160001a601b60ff8216101561173457601b015b8060ff16601b1415801561174c57508060ff16601c14155b1561175a57600093506117c1565b60408051600080825260208083018085528a905260ff8516838501526060830187905260808301869052925160019360a0808501949193601f19840193928390039091019190865af11580156117b4573d6000803e3d6000fd5b5050506020604051035193505b505050929150505600a6eef7e35abe7026729641147f7915573c7e97b47efa546f5f6e3230263bcb49a165627a7a72305820f3f4353f4402d0c48bd24413a5b47206e81cced88aed0169a85f26969f3940c00029`

// DeployPalletOne deploys a new Ethereum contract, binding an instance of PalletOne to it.
func DeployPalletOne(auth *bind.TransactOpts, backend bind.ContractBackend, admin_ common.Address) (common.Address, *types.Transaction, *PalletOne, error) {
	parsed, err := abi.JSON(strings.NewReader(PalletOneABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(PalletOneBin), backend, admin_)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &PalletOne{PalletOneCaller: PalletOneCaller{contract: contract}, PalletOneTransactor: PalletOneTransactor{contract: contract}, PalletOneFilterer: PalletOneFilterer{contract: contract}}, nil
}

// PalletOne is an auto generated Go binding around an Ethereum contract.
type PalletOne struct {
	PalletOneCaller     // Read-only binding to the contract
	PalletOneTransactor // Write-only binding to the contract
	PalletOneFilterer   // Log filterer for contract events
}

// PalletOneCaller is an auto generated read-only Go binding around an Ethereum contract.
type PalletOneCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// PalletOneTransactor is an auto generated write-only Go binding around an Ethereum contract.
type PalletOneTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// PalletOneFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type PalletOneFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// PalletOneSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type PalletOneSession struct {
	Contract     *PalletOne        // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// PalletOneCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type PalletOneCallerSession struct {
	Contract *PalletOneCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts    // Call options to use throughout this session
}

// PalletOneTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type PalletOneTransactorSession struct {
	Contract     *PalletOneTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts    // Transaction auth options to use throughout this session
}

// PalletOneRaw is an auto generated low-level Go binding around an Ethereum contract.
type PalletOneRaw struct {
	Contract *PalletOne // Generic contract binding to access the raw methods on
}

// PalletOneCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type PalletOneCallerRaw struct {
	Contract *PalletOneCaller // Generic read-only contract binding to access the raw methods on
}

// PalletOneTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type PalletOneTransactorRaw struct {
	Contract *PalletOneTransactor // Generic write-only contract binding to access the raw methods on
}

// NewPalletOne creates a new instance of PalletOne, bound to a specific deployed contract.
func NewPalletOne(address common.Address, backend bind.ContractBackend) (*PalletOne, error) {
	contract, err := bindPalletOne(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &PalletOne{PalletOneCaller: PalletOneCaller{contract: contract}, PalletOneTransactor: PalletOneTransactor{contract: contract}, PalletOneFilterer: PalletOneFilterer{contract: contract}}, nil
}

// NewPalletOneCaller creates a new read-only instance of PalletOne, bound to a specific deployed contract.
func NewPalletOneCaller(address common.Address, caller bind.ContractCaller) (*PalletOneCaller, error) {
	contract, err := bindPalletOne(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &PalletOneCaller{contract: contract}, nil
}

// NewPalletOneTransactor creates a new write-only instance of PalletOne, bound to a specific deployed contract.
func NewPalletOneTransactor(address common.Address, transactor bind.ContractTransactor) (*PalletOneTransactor, error) {
	contract, err := bindPalletOne(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &PalletOneTransactor{contract: contract}, nil
}

// NewPalletOneFilterer creates a new log filterer instance of PalletOne, bound to a specific deployed contract.
func NewPalletOneFilterer(address common.Address, filterer bind.ContractFilterer) (*PalletOneFilterer, error) {
	contract, err := bindPalletOne(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &PalletOneFilterer{contract: contract}, nil
}

// bindPalletOne binds a generic wrapper to an already deployed contract.
func bindPalletOne(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(PalletOneABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_PalletOne *PalletOneRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _PalletOne.Contract.PalletOneCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_PalletOne *PalletOneRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _PalletOne.Contract.PalletOneTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_PalletOne *PalletOneRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _PalletOne.Contract.PalletOneTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_PalletOne *PalletOneCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _PalletOne.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_PalletOne *PalletOneTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _PalletOne.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_PalletOne *PalletOneTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _PalletOne.Contract.contract.Transact(opts, method, params...)
}

// Admin is a free data retrieval call binding the contract method 0xf851a440.
//
// Solidity: function admin() constant returns(address)
func (_PalletOne *PalletOneCaller) Admin(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _PalletOne.contract.Call(opts, out, "admin")
	return *ret0, err
}

// Admin is a free data retrieval call binding the contract method 0xf851a440.
//
// Solidity: function admin() constant returns(address)
func (_PalletOne *PalletOneSession) Admin() (common.Address, error) {
	return _PalletOne.Contract.Admin(&_PalletOne.CallOpts)
}

// Admin is a free data retrieval call binding the contract method 0xf851a440.
//
// Solidity: function admin() constant returns(address)
func (_PalletOne *PalletOneCallerSession) Admin() (common.Address, error) {
	return _PalletOne.Contract.Admin(&_PalletOne.CallOpts)
}

// Getmultisig is a free data retrieval call binding the contract method 0xe7a64ff2.
//
// Solidity: function getmultisig(addr address, redeem bytes) constant returns(uint256, uint256)
func (_PalletOne *PalletOneCaller) Getmultisig(opts *bind.CallOpts, addr common.Address, redeem []byte) (*big.Int, *big.Int, error) {
	var (
		ret0 = new(*big.Int)
		ret1 = new(*big.Int)
	)
	out := &[]interface{}{
		ret0,
		ret1,
	}
	err := _PalletOne.contract.Call(opts, out, "getmultisig", addr, redeem)
	return *ret0, *ret1, err
}

// Getmultisig is a free data retrieval call binding the contract method 0xe7a64ff2.
//
// Solidity: function getmultisig(addr address, redeem bytes) constant returns(uint256, uint256)
func (_PalletOne *PalletOneSession) Getmultisig(addr common.Address, redeem []byte) (*big.Int, *big.Int, error) {
	return _PalletOne.Contract.Getmultisig(&_PalletOne.CallOpts, addr, redeem)
}

// Getmultisig is a free data retrieval call binding the contract method 0xe7a64ff2.
//
// Solidity: function getmultisig(addr address, redeem bytes) constant returns(uint256, uint256)
func (_PalletOne *PalletOneCallerSession) Getmultisig(addr common.Address, redeem []byte) (*big.Int, *big.Int, error) {
	return _PalletOne.Contract.Getmultisig(&_PalletOne.CallOpts, addr, redeem)
}

// MyEthBal is a free data retrieval call binding the contract method 0xc8fc638a.
//
// Solidity: function my_eth_bal() constant returns(uint256)
func (_PalletOne *PalletOneCaller) MyEthBal(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _PalletOne.contract.Call(opts, out, "my_eth_bal")
	return *ret0, err
}

// MyEthBal is a free data retrieval call binding the contract method 0xc8fc638a.
//
// Solidity: function my_eth_bal() constant returns(uint256)
func (_PalletOne *PalletOneSession) MyEthBal() (*big.Int, error) {
	return _PalletOne.Contract.MyEthBal(&_PalletOne.CallOpts)
}

// MyEthBal is a free data retrieval call binding the contract method 0xc8fc638a.
//
// Solidity: function my_eth_bal() constant returns(uint256)
func (_PalletOne *PalletOneCallerSession) MyEthBal() (*big.Int, error) {
	return _PalletOne.Contract.MyEthBal(&_PalletOne.CallOpts)
}

// Tokens is a free data retrieval call binding the contract method 0xa9964d1c.
//
// Solidity: function tokens( address,  bytes32) constant returns(balance uint256, nonece uint256)
func (_PalletOne *PalletOneCaller) Tokens(opts *bind.CallOpts, arg0 common.Address, arg1 [32]byte) (struct {
	Balance *big.Int
	Nonece  *big.Int
}, error) {
	ret := new(struct {
		Balance *big.Int
		Nonece  *big.Int
	})
	out := ret
	err := _PalletOne.contract.Call(opts, out, "tokens", arg0, arg1)
	return *ret, err
}

// Tokens is a free data retrieval call binding the contract method 0xa9964d1c.
//
// Solidity: function tokens( address,  bytes32) constant returns(balance uint256, nonece uint256)
func (_PalletOne *PalletOneSession) Tokens(arg0 common.Address, arg1 [32]byte) (struct {
	Balance *big.Int
	Nonece  *big.Int
}, error) {
	return _PalletOne.Contract.Tokens(&_PalletOne.CallOpts, arg0, arg1)
}

// Tokens is a free data retrieval call binding the contract method 0xa9964d1c.
//
// Solidity: function tokens( address,  bytes32) constant returns(balance uint256, nonece uint256)
func (_PalletOne *PalletOneCallerSession) Tokens(arg0 common.Address, arg1 [32]byte) (struct {
	Balance *big.Int
	Nonece  *big.Int
}, error) {
	return _PalletOne.Contract.Tokens(&_PalletOne.CallOpts, arg0, arg1)
}

// Deposit is a paid mutator transaction binding the contract method 0x98b1e06a.
//
// Solidity: function deposit(redeem bytes) returns()
func (_PalletOne *PalletOneTransactor) Deposit(opts *bind.TransactOpts, redeem []byte) (*types.Transaction, error) {
	return _PalletOne.contract.Transact(opts, "deposit", redeem)
}

// Deposit is a paid mutator transaction binding the contract method 0x98b1e06a.
//
// Solidity: function deposit(redeem bytes) returns()
func (_PalletOne *PalletOneSession) Deposit(redeem []byte) (*types.Transaction, error) {
	return _PalletOne.Contract.Deposit(&_PalletOne.TransactOpts, redeem)
}

// Deposit is a paid mutator transaction binding the contract method 0x98b1e06a.
//
// Solidity: function deposit(redeem bytes) returns()
func (_PalletOne *PalletOneTransactorSession) Deposit(redeem []byte) (*types.Transaction, error) {
	return _PalletOne.Contract.Deposit(&_PalletOne.TransactOpts, redeem)
}

// Deposittoken is a paid mutator transaction binding the contract method 0xc0c6cf4e.
//
// Solidity: function deposittoken(token address, redeem bytes, amount uint256) returns()
func (_PalletOne *PalletOneTransactor) Deposittoken(opts *bind.TransactOpts, token common.Address, redeem []byte, amount *big.Int) (*types.Transaction, error) {
	return _PalletOne.contract.Transact(opts, "deposittoken", token, redeem, amount)
}

// Deposittoken is a paid mutator transaction binding the contract method 0xc0c6cf4e.
//
// Solidity: function deposittoken(token address, redeem bytes, amount uint256) returns()
func (_PalletOne *PalletOneSession) Deposittoken(token common.Address, redeem []byte, amount *big.Int) (*types.Transaction, error) {
	return _PalletOne.Contract.Deposittoken(&_PalletOne.TransactOpts, token, redeem, amount)
}

// Deposittoken is a paid mutator transaction binding the contract method 0xc0c6cf4e.
//
// Solidity: function deposittoken(token address, redeem bytes, amount uint256) returns()
func (_PalletOne *PalletOneTransactorSession) Deposittoken(token common.Address, redeem []byte, amount *big.Int) (*types.Transaction, error) {
	return _PalletOne.Contract.Deposittoken(&_PalletOne.TransactOpts, token, redeem, amount)
}

// Suicideto is a paid mutator transaction binding the contract method 0x8e644ec3.
//
// Solidity: function suicideto(addr address) returns()
func (_PalletOne *PalletOneTransactor) Suicideto(opts *bind.TransactOpts, addr common.Address) (*types.Transaction, error) {
	return _PalletOne.contract.Transact(opts, "suicideto", addr)
}

// Suicideto is a paid mutator transaction binding the contract method 0x8e644ec3.
//
// Solidity: function suicideto(addr address) returns()
func (_PalletOne *PalletOneSession) Suicideto(addr common.Address) (*types.Transaction, error) {
	return _PalletOne.Contract.Suicideto(&_PalletOne.TransactOpts, addr)
}

// Suicideto is a paid mutator transaction binding the contract method 0x8e644ec3.
//
// Solidity: function suicideto(addr address) returns()
func (_PalletOne *PalletOneTransactorSession) Suicideto(addr common.Address) (*types.Transaction, error) {
	return _PalletOne.Contract.Suicideto(&_PalletOne.TransactOpts, addr)
}

// Withdraw is a paid mutator transaction binding the contract method 0x8c2e0321.
//
// Solidity: function withdraw(redeem bytes, recver address, amount uint256, nonece uint256, sigstr1 bytes, sigstr2 bytes) returns()
func (_PalletOne *PalletOneTransactor) Withdraw(opts *bind.TransactOpts, redeem []byte, recver common.Address, amount *big.Int, nonece *big.Int, sigstr1 []byte, sigstr2 []byte) (*types.Transaction, error) {
	return _PalletOne.contract.Transact(opts, "withdraw", redeem, recver, amount, nonece, sigstr1, sigstr2)
}

// Withdraw is a paid mutator transaction binding the contract method 0x8c2e0321.
//
// Solidity: function withdraw(redeem bytes, recver address, amount uint256, nonece uint256, sigstr1 bytes, sigstr2 bytes) returns()
func (_PalletOne *PalletOneSession) Withdraw(redeem []byte, recver common.Address, amount *big.Int, nonece *big.Int, sigstr1 []byte, sigstr2 []byte) (*types.Transaction, error) {
	return _PalletOne.Contract.Withdraw(&_PalletOne.TransactOpts, redeem, recver, amount, nonece, sigstr1, sigstr2)
}

// Withdraw is a paid mutator transaction binding the contract method 0x8c2e0321.
//
// Solidity: function withdraw(redeem bytes, recver address, amount uint256, nonece uint256, sigstr1 bytes, sigstr2 bytes) returns()
func (_PalletOne *PalletOneTransactorSession) Withdraw(redeem []byte, recver common.Address, amount *big.Int, nonece *big.Int, sigstr1 []byte, sigstr2 []byte) (*types.Transaction, error) {
	return _PalletOne.Contract.Withdraw(&_PalletOne.TransactOpts, redeem, recver, amount, nonece, sigstr1, sigstr2)
}

// Withdrawtoken is a paid mutator transaction binding the contract method 0x33890eca.
//
// Solidity: function withdrawtoken(token address, redeem bytes, recver address, amount uint256, nonece uint256, sigstr1 bytes, sigstr2 bytes) returns()
func (_PalletOne *PalletOneTransactor) Withdrawtoken(opts *bind.TransactOpts, token common.Address, redeem []byte, recver common.Address, amount *big.Int, nonece *big.Int, sigstr1 []byte, sigstr2 []byte) (*types.Transaction, error) {
	return _PalletOne.contract.Transact(opts, "withdrawtoken", token, redeem, recver, amount, nonece, sigstr1, sigstr2)
}

// Withdrawtoken is a paid mutator transaction binding the contract method 0x33890eca.
//
// Solidity: function withdrawtoken(token address, redeem bytes, recver address, amount uint256, nonece uint256, sigstr1 bytes, sigstr2 bytes) returns()
func (_PalletOne *PalletOneSession) Withdrawtoken(token common.Address, redeem []byte, recver common.Address, amount *big.Int, nonece *big.Int, sigstr1 []byte, sigstr2 []byte) (*types.Transaction, error) {
	return _PalletOne.Contract.Withdrawtoken(&_PalletOne.TransactOpts, token, redeem, recver, amount, nonece, sigstr1, sigstr2)
}

// Withdrawtoken is a paid mutator transaction binding the contract method 0x33890eca.
//
// Solidity: function withdrawtoken(token address, redeem bytes, recver address, amount uint256, nonece uint256, sigstr1 bytes, sigstr2 bytes) returns()
func (_PalletOne *PalletOneTransactorSession) Withdrawtoken(token common.Address, redeem []byte, recver common.Address, amount *big.Int, nonece *big.Int, sigstr1 []byte, sigstr2 []byte) (*types.Transaction, error) {
	return _PalletOne.Contract.Withdrawtoken(&_PalletOne.TransactOpts, token, redeem, recver, amount, nonece, sigstr1, sigstr2)
}

// PalletOneDepositIterator is returned from FilterDeposit and is used to iterate over the raw logs and unpacked data for Deposit events raised by the PalletOne contract.
type PalletOneDepositIterator struct {
	Event *PalletOneDeposit // Event containing the contract specifics and raw log

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
func (it *PalletOneDepositIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PalletOneDeposit)
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
		it.Event = new(PalletOneDeposit)
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
func (it *PalletOneDepositIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PalletOneDepositIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PalletOneDeposit represents a Deposit event raised by the PalletOne contract.
type PalletOneDeposit struct {
	Token  common.Address
	User   common.Address
	Amount *big.Int
	Redeem []byte
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterDeposit is a free log retrieval operation binding the contract event 0xd5d9ab68ad56311de2cda7e56730c5a58bcd4c9d071b9fe5f8efcdb1ccc9251d.
//
// Solidity: e Deposit(token address, user address, amount uint256, redeem bytes)
func (_PalletOne *PalletOneFilterer) FilterDeposit(opts *bind.FilterOpts) (*PalletOneDepositIterator, error) {

	logs, sub, err := _PalletOne.contract.FilterLogs(opts, "Deposit")
	if err != nil {
		return nil, err
	}
	return &PalletOneDepositIterator{contract: _PalletOne.contract, event: "Deposit", logs: logs, sub: sub}, nil
}

// WatchDeposit is a free log subscription operation binding the contract event 0xd5d9ab68ad56311de2cda7e56730c5a58bcd4c9d071b9fe5f8efcdb1ccc9251d.
//
// Solidity: e Deposit(token address, user address, amount uint256, redeem bytes)
func (_PalletOne *PalletOneFilterer) WatchDeposit(opts *bind.WatchOpts, sink chan<- *PalletOneDeposit) (event.Subscription, error) {

	logs, sub, err := _PalletOne.contract.WatchLogs(opts, "Deposit")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PalletOneDeposit)
				if err := _PalletOne.contract.UnpackLog(event, "Deposit", log); err != nil {
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

// PalletOneWithdrawIterator is returned from FilterWithdraw and is used to iterate over the raw logs and unpacked data for Withdraw events raised by the PalletOne contract.
type PalletOneWithdrawIterator struct {
	Event *PalletOneWithdraw // Event containing the contract specifics and raw log

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
func (it *PalletOneWithdrawIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PalletOneWithdraw)
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
		it.Event = new(PalletOneWithdraw)
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
func (it *PalletOneWithdrawIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PalletOneWithdrawIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PalletOneWithdraw represents a Withdraw event raised by the PalletOne contract.
type PalletOneWithdraw struct {
	Token        common.Address
	User         common.Address
	Redeem       []byte
	Recver       common.Address
	Amount       *big.Int
	Confirmvalue *big.Int
	State        string
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterWithdraw is a free log retrieval operation binding the contract event 0x3c787786801bcab2749cb2c8202e63081bfdb0ef3bc5c9cea89cacd3e7ef4cf3.
//
// Solidity: e Withdraw(token address, user address, redeem bytes, recver address, amount uint256, confirmvalue uint256, state string)
func (_PalletOne *PalletOneFilterer) FilterWithdraw(opts *bind.FilterOpts) (*PalletOneWithdrawIterator, error) {

	logs, sub, err := _PalletOne.contract.FilterLogs(opts, "Withdraw")
	if err != nil {
		return nil, err
	}
	return &PalletOneWithdrawIterator{contract: _PalletOne.contract, event: "Withdraw", logs: logs, sub: sub}, nil
}

// WatchWithdraw is a free log subscription operation binding the contract event 0x3c787786801bcab2749cb2c8202e63081bfdb0ef3bc5c9cea89cacd3e7ef4cf3.
//
// Solidity: e Withdraw(token address, user address, redeem bytes, recver address, amount uint256, confirmvalue uint256, state string)
func (_PalletOne *PalletOneFilterer) WatchWithdraw(opts *bind.WatchOpts, sink chan<- *PalletOneWithdraw) (event.Subscription, error) {

	logs, sub, err := _PalletOne.contract.WatchLogs(opts, "Withdraw")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PalletOneWithdraw)
				if err := _PalletOne.contract.UnpackLog(event, "Withdraw", log); err != nil {
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
