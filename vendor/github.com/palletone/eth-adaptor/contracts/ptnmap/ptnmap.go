// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package ptnmap

import (
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
	"github.com/palletone/eth-adaptor/bind"
)

// IERC20ABI is the input ABI used to generate the binding from.
const IERC20ABI = "[{\"constant\":false,\"inputs\":[{\"name\":\"spender\",\"type\":\"address\"},{\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"approve\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"totalSupply\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"from\",\"type\":\"address\"},{\"name\":\"to\",\"type\":\"address\"},{\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"transferFrom\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"who\",\"type\":\"address\"}],\"name\":\"balanceOf\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"to\",\"type\":\"address\"},{\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"transfer\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"owner\",\"type\":\"address\"},{\"name\":\"spender\",\"type\":\"address\"}],\"name\":\"allowance\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"to\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Transfer\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"spender\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Approval\",\"type\":\"event\"}]"

// IERC20Bin is the compiled bytecode used for deploying new contracts.
const IERC20Bin = `0x`

// DeployIERC20 deploys a new Ethereum contract, binding an instance of IERC20 to it.
func DeployIERC20(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *IERC20, error) {
	parsed, err := abi.JSON(strings.NewReader(IERC20ABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(IERC20Bin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &IERC20{IERC20Caller: IERC20Caller{contract: contract}, IERC20Transactor: IERC20Transactor{contract: contract}, IERC20Filterer: IERC20Filterer{contract: contract}}, nil
}

// IERC20 is an auto generated Go binding around an Ethereum contract.
type IERC20 struct {
	IERC20Caller     // Read-only binding to the contract
	IERC20Transactor // Write-only binding to the contract
	IERC20Filterer   // Log filterer for contract events
}

// IERC20Caller is an auto generated read-only Go binding around an Ethereum contract.
type IERC20Caller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// IERC20Transactor is an auto generated write-only Go binding around an Ethereum contract.
type IERC20Transactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// IERC20Filterer is an auto generated log filtering Go binding around an Ethereum contract events.
type IERC20Filterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// IERC20Session is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type IERC20Session struct {
	Contract     *IERC20           // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// IERC20CallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type IERC20CallerSession struct {
	Contract *IERC20Caller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts // Call options to use throughout this session
}

// IERC20TransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type IERC20TransactorSession struct {
	Contract     *IERC20Transactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// IERC20Raw is an auto generated low-level Go binding around an Ethereum contract.
type IERC20Raw struct {
	Contract *IERC20 // Generic contract binding to access the raw methods on
}

// IERC20CallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type IERC20CallerRaw struct {
	Contract *IERC20Caller // Generic read-only contract binding to access the raw methods on
}

// IERC20TransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type IERC20TransactorRaw struct {
	Contract *IERC20Transactor // Generic write-only contract binding to access the raw methods on
}

// NewIERC20 creates a new instance of IERC20, bound to a specific deployed contract.
func NewIERC20(address common.Address, backend bind.ContractBackend) (*IERC20, error) {
	contract, err := bindIERC20(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &IERC20{IERC20Caller: IERC20Caller{contract: contract}, IERC20Transactor: IERC20Transactor{contract: contract}, IERC20Filterer: IERC20Filterer{contract: contract}}, nil
}

// NewIERC20Caller creates a new read-only instance of IERC20, bound to a specific deployed contract.
func NewIERC20Caller(address common.Address, caller bind.ContractCaller) (*IERC20Caller, error) {
	contract, err := bindIERC20(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &IERC20Caller{contract: contract}, nil
}

// NewIERC20Transactor creates a new write-only instance of IERC20, bound to a specific deployed contract.
func NewIERC20Transactor(address common.Address, transactor bind.ContractTransactor) (*IERC20Transactor, error) {
	contract, err := bindIERC20(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &IERC20Transactor{contract: contract}, nil
}

// NewIERC20Filterer creates a new log filterer instance of IERC20, bound to a specific deployed contract.
func NewIERC20Filterer(address common.Address, filterer bind.ContractFilterer) (*IERC20Filterer, error) {
	contract, err := bindIERC20(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &IERC20Filterer{contract: contract}, nil
}

// bindIERC20 binds a generic wrapper to an already deployed contract.
func bindIERC20(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(IERC20ABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_IERC20 *IERC20Raw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _IERC20.Contract.IERC20Caller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_IERC20 *IERC20Raw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _IERC20.Contract.IERC20Transactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_IERC20 *IERC20Raw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _IERC20.Contract.IERC20Transactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_IERC20 *IERC20CallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _IERC20.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_IERC20 *IERC20TransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _IERC20.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_IERC20 *IERC20TransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _IERC20.Contract.contract.Transact(opts, method, params...)
}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(owner address, spender address) constant returns(uint256)
func (_IERC20 *IERC20Caller) Allowance(opts *bind.CallOpts, owner common.Address, spender common.Address) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _IERC20.contract.Call(opts, out, "allowance", owner, spender)
	return *ret0, err
}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(owner address, spender address) constant returns(uint256)
func (_IERC20 *IERC20Session) Allowance(owner common.Address, spender common.Address) (*big.Int, error) {
	return _IERC20.Contract.Allowance(&_IERC20.CallOpts, owner, spender)
}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(owner address, spender address) constant returns(uint256)
func (_IERC20 *IERC20CallerSession) Allowance(owner common.Address, spender common.Address) (*big.Int, error) {
	return _IERC20.Contract.Allowance(&_IERC20.CallOpts, owner, spender)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(who address) constant returns(uint256)
func (_IERC20 *IERC20Caller) BalanceOf(opts *bind.CallOpts, who common.Address) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _IERC20.contract.Call(opts, out, "balanceOf", who)
	return *ret0, err
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(who address) constant returns(uint256)
func (_IERC20 *IERC20Session) BalanceOf(who common.Address) (*big.Int, error) {
	return _IERC20.Contract.BalanceOf(&_IERC20.CallOpts, who)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(who address) constant returns(uint256)
func (_IERC20 *IERC20CallerSession) BalanceOf(who common.Address) (*big.Int, error) {
	return _IERC20.Contract.BalanceOf(&_IERC20.CallOpts, who)
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() constant returns(uint256)
func (_IERC20 *IERC20Caller) TotalSupply(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _IERC20.contract.Call(opts, out, "totalSupply")
	return *ret0, err
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() constant returns(uint256)
func (_IERC20 *IERC20Session) TotalSupply() (*big.Int, error) {
	return _IERC20.Contract.TotalSupply(&_IERC20.CallOpts)
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() constant returns(uint256)
func (_IERC20 *IERC20CallerSession) TotalSupply() (*big.Int, error) {
	return _IERC20.Contract.TotalSupply(&_IERC20.CallOpts)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(spender address, value uint256) returns(bool)
func (_IERC20 *IERC20Transactor) Approve(opts *bind.TransactOpts, spender common.Address, value *big.Int) (*types.Transaction, error) {
	return _IERC20.contract.Transact(opts, "approve", spender, value)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(spender address, value uint256) returns(bool)
func (_IERC20 *IERC20Session) Approve(spender common.Address, value *big.Int) (*types.Transaction, error) {
	return _IERC20.Contract.Approve(&_IERC20.TransactOpts, spender, value)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(spender address, value uint256) returns(bool)
func (_IERC20 *IERC20TransactorSession) Approve(spender common.Address, value *big.Int) (*types.Transaction, error) {
	return _IERC20.Contract.Approve(&_IERC20.TransactOpts, spender, value)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(to address, value uint256) returns(bool)
func (_IERC20 *IERC20Transactor) Transfer(opts *bind.TransactOpts, to common.Address, value *big.Int) (*types.Transaction, error) {
	return _IERC20.contract.Transact(opts, "transfer", to, value)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(to address, value uint256) returns(bool)
func (_IERC20 *IERC20Session) Transfer(to common.Address, value *big.Int) (*types.Transaction, error) {
	return _IERC20.Contract.Transfer(&_IERC20.TransactOpts, to, value)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(to address, value uint256) returns(bool)
func (_IERC20 *IERC20TransactorSession) Transfer(to common.Address, value *big.Int) (*types.Transaction, error) {
	return _IERC20.Contract.Transfer(&_IERC20.TransactOpts, to, value)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(from address, to address, value uint256) returns(bool)
func (_IERC20 *IERC20Transactor) TransferFrom(opts *bind.TransactOpts, from common.Address, to common.Address, value *big.Int) (*types.Transaction, error) {
	return _IERC20.contract.Transact(opts, "transferFrom", from, to, value)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(from address, to address, value uint256) returns(bool)
func (_IERC20 *IERC20Session) TransferFrom(from common.Address, to common.Address, value *big.Int) (*types.Transaction, error) {
	return _IERC20.Contract.TransferFrom(&_IERC20.TransactOpts, from, to, value)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(from address, to address, value uint256) returns(bool)
func (_IERC20 *IERC20TransactorSession) TransferFrom(from common.Address, to common.Address, value *big.Int) (*types.Transaction, error) {
	return _IERC20.Contract.TransferFrom(&_IERC20.TransactOpts, from, to, value)
}

// IERC20ApprovalIterator is returned from FilterApproval and is used to iterate over the raw logs and unpacked data for Approval events raised by the IERC20 contract.
type IERC20ApprovalIterator struct {
	Event *IERC20Approval // Event containing the contract specifics and raw log

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
func (it *IERC20ApprovalIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(IERC20Approval)
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
		it.Event = new(IERC20Approval)
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
func (it *IERC20ApprovalIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *IERC20ApprovalIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// IERC20Approval represents a Approval event raised by the IERC20 contract.
type IERC20Approval struct {
	Owner   common.Address
	Spender common.Address
	Value   *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterApproval is a free log retrieval operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: e Approval(owner indexed address, spender indexed address, value uint256)
func (_IERC20 *IERC20Filterer) FilterApproval(opts *bind.FilterOpts, owner []common.Address, spender []common.Address) (*IERC20ApprovalIterator, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var spenderRule []interface{}
	for _, spenderItem := range spender {
		spenderRule = append(spenderRule, spenderItem)
	}

	logs, sub, err := _IERC20.contract.FilterLogs(opts, "Approval", ownerRule, spenderRule)
	if err != nil {
		return nil, err
	}
	return &IERC20ApprovalIterator{contract: _IERC20.contract, event: "Approval", logs: logs, sub: sub}, nil
}

// WatchApproval is a free log subscription operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: e Approval(owner indexed address, spender indexed address, value uint256)
func (_IERC20 *IERC20Filterer) WatchApproval(opts *bind.WatchOpts, sink chan<- *IERC20Approval, owner []common.Address, spender []common.Address) (event.Subscription, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var spenderRule []interface{}
	for _, spenderItem := range spender {
		spenderRule = append(spenderRule, spenderItem)
	}

	logs, sub, err := _IERC20.contract.WatchLogs(opts, "Approval", ownerRule, spenderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(IERC20Approval)
				if err := _IERC20.contract.UnpackLog(event, "Approval", log); err != nil {
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

// IERC20TransferIterator is returned from FilterTransfer and is used to iterate over the raw logs and unpacked data for Transfer events raised by the IERC20 contract.
type IERC20TransferIterator struct {
	Event *IERC20Transfer // Event containing the contract specifics and raw log

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
func (it *IERC20TransferIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(IERC20Transfer)
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
		it.Event = new(IERC20Transfer)
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
func (it *IERC20TransferIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *IERC20TransferIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// IERC20Transfer represents a Transfer event raised by the IERC20 contract.
type IERC20Transfer struct {
	From  common.Address
	To    common.Address
	Value *big.Int
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterTransfer is a free log retrieval operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: e Transfer(from indexed address, to indexed address, value uint256)
func (_IERC20 *IERC20Filterer) FilterTransfer(opts *bind.FilterOpts, from []common.Address, to []common.Address) (*IERC20TransferIterator, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _IERC20.contract.FilterLogs(opts, "Transfer", fromRule, toRule)
	if err != nil {
		return nil, err
	}
	return &IERC20TransferIterator{contract: _IERC20.contract, event: "Transfer", logs: logs, sub: sub}, nil
}

// WatchTransfer is a free log subscription operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: e Transfer(from indexed address, to indexed address, value uint256)
func (_IERC20 *IERC20Filterer) WatchTransfer(opts *bind.WatchOpts, sink chan<- *IERC20Transfer, from []common.Address, to []common.Address) (event.Subscription, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _IERC20.contract.WatchLogs(opts, "Transfer", fromRule, toRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(IERC20Transfer)
				if err := _IERC20.contract.UnpackLog(event, "Transfer", log); err != nil {
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
const LSafeMathBin = `0x604c602c600b82828239805160001a60731460008114601c57601e565bfe5b5030600052607381538281f30073000000000000000000000000000000000000000030146080604052600080fd00a165627a7a72305820b5e1f6f26f308df02c72f9331c8599ec7a9dc9382fb75ae07597a5fc236588ad0029`

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

// PTNMapABI is the input ABI used to generate the binding from.
const PTNMapABI = "[{\"constant\":true,\"inputs\":[],\"name\":\"name\",\"outputs\":[{\"name\":\"\",\"type\":\"string\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"spender\",\"type\":\"address\"},{\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"approve\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"ptnToken\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"totalSupply\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"from\",\"type\":\"address\"},{\"name\":\"to\",\"type\":\"address\"},{\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"transferFrom\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_addr\",\"type\":\"address\"},{\"name\":\"_ptnhex\",\"type\":\"address\"}],\"name\":\"resetMapAddr\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"decimals\",\"outputs\":[{\"name\":\"\",\"type\":\"uint8\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"getMapPtnAddr\",\"outputs\":[{\"name\":\"\",\"type\":\"string\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"address\"}],\"name\":\"addrmapPTN\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"ptnAddr\",\"type\":\"address\"}],\"name\":\"getMapEthAddr\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_owner\",\"type\":\"address\"}],\"name\":\"balanceOf\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"address\"}],\"name\":\"addrmap\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"addrHex\",\"type\":\"address\"}],\"name\":\"encodeBase58\",\"outputs\":[{\"name\":\"\",\"type\":\"string\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"symbol\",\"outputs\":[{\"name\":\"\",\"type\":\"string\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_ptnhex\",\"type\":\"address\"},{\"name\":\"_amt\",\"type\":\"uint256\"}],\"name\":\"transfer\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"owner\",\"type\":\"address\"},{\"name\":\"spender\",\"type\":\"address\"}],\"name\":\"allowance\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"name\":\"_erc20Addr\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"fallback\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"to\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Transfer\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"spender\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Approval\",\"type\":\"event\"}]"

// PTNMapBin is the compiled bytecode used for deploying new contracts.
const PTNMapBin = `0x6080604052633b9aca0060035534801561001857600080fd5b50604051602080610ef0833981016040525160048054600160a060020a0319908116331790915560008054600160a060020a0390931692909116919091179055610e89806100676000396000f3006080604052600436106100e55763ffffffff7c010000000000000000000000000000000000000000000000000000000060003504166306fdde0381146100e7578063095ea7b3146101715780630ac74e72146101a957806318160ddd146101da57806323b872dd146102015780632da2ff341461022b578063313ce5671461025257806348cedf901461027d5780634e11092f1461029e5780636e932a1c146102bf57806370a08231146102e05780638c5cecaa14610301578063927f526f1461032257806395d89b4114610343578063a9059cbb14610358578063dd62ed3e1461037c575b005b3480156100f357600080fd5b506100fc6103a3565b6040805160208082528351818301528351919283929083019185019080838360005b8381101561013657818101518382015260200161011e565b50505050905090810190601f1680156101635780820380516001836020036101000a031916815260200191505b509250505060405180910390f35b34801561017d57600080fd5b50610195600160a060020a03600435166024356103da565b604080519115158252519081900360200190f35b3480156101b557600080fd5b506101be6103e3565b60408051600160a060020a039092168252519081900360200190f35b3480156101e657600080fd5b506101ef6103f2565b60408051918252519081900360200190f35b34801561020d57600080fd5b50610195600160a060020a0360043581169060243516604435610490565b34801561023757600080fd5b506100e5600160a060020a0360043581169060243516610499565b34801561025e57600080fd5b50610267610558565b6040805160ff9092168252519081900360200190f35b34801561028957600080fd5b506100fc600160a060020a036004351661055d565b3480156102aa57600080fd5b506101be600160a060020a036004351661058c565b3480156102cb57600080fd5b506101be600160a060020a03600435166105a7565b3480156102ec57600080fd5b506101ef600160a060020a03600435166105c5565b34801561030d57600080fd5b506101be600160a060020a0360043516610619565b34801561032e57600080fd5b506100fc600160a060020a0360043516610634565b34801561034f57600080fd5b506100fc610659565b34801561036457600080fd5b50610195600160a060020a0360043516602435610690565b34801561038857600080fd5b506101ef600160a060020a03600435811690602435166103da565b60408051808201909152600b81527f50544e204d617070696e67000000000000000000000000000000000000000000602082015281565b60005b92915050565b600054600160a060020a031681565b60008060009054906101000a9004600160a060020a0316600160a060020a03166318160ddd6040518163ffffffff167c0100000000000000000000000000000000000000000000000000000000028152600401602060405180830381600087803b15801561045f57600080fd5b505af1158015610473573d6000803e3d6000fd5b505050506040513d602081101561048957600080fd5b5051905090565b60009392505050565b600454600160a060020a031633146104b057600080fd5b600160a060020a0382811660009081526001602052604090205481169082161480156104f85750600160a060020a038181166000908152600260205260409020548116908316145b1561054f57600160a060020a038083166000908152600160209081526040808320805473ffffffffffffffffffffffffffffffffffffffff1990811690915593851683526002909152902080549091169055610554565b600080fd5b5050565b600081565b600160a060020a038082166000908152600160205260409020546060916105849116610634565b90505b919050565b600260205260009081526040902054600160a060020a031681565b600160a060020a039081166000908152600260205260409020541690565b600160a060020a03818116600090815260016020526040812054909116151561061157600160a060020a0382811660009081526002602052604090205416151561061157506001610587565b506000610587565b600160205260009081526040902054600160a060020a031681565b6060818161065161064c61064784610761565b610939565b610ad2565b949350505050565b60408051808201909152600681527f50544e4d61700000000000000000000000000000000000000000000000000000602082015281565b33600090815260016020526040812054600160a060020a03161580156106ce5750600160a060020a0383811660009081526002602052604090205416155b1561054f573360008181526001602090815260408083208054600160a060020a03891673ffffffffffffffffffffffffffffffffffffffff199182168117909255818552600284529382902080549094168517909355805186815290519293927fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef929181900390910190a35060016103dd565b6040805160198082528183019092526060916c0100000000000000000000000084029183916000918291829190602082016103208038833950508151919550600091869150829081106107b057fe5b906020010190600160f860020a031916908160001a905350600092505b60148360ff161015610826578460ff8416601481106107e857fe5b1a60f860020a02848460010160ff1681518110151561080357fe5b906020010190600160f860020a031916908160001a9053506001909201916107cd565b6040805160008082526bffffffffffffffffffffffff1988166001830152915160029283926015808201936020939092839003909101908290865af1158015610873573d6000803e3d6000fd5b5050506040513d602081101561088857600080fd5b505160408051918252516020828101929091908190038201816000865af11580156108b7573d6000803e3d6000fd5b5050506040513d60208110156108cc57600080fd5b50519150600090505b60048160ff16101561092b578160ff8216602081106108f057fe5b1a60f860020a02848260150160ff1681518110151561090b57fe5b906020010190600160f860020a031916908160001a9053506001016108d5565b8395505b5050505050919050565b60608060008060008086516000141561096257604080516020810190915260008152955061092f565b6040805160288082526105208201909252906020820161050080388339019050509450600085600081518110151561099657fe5b60ff90921660209283029091019091015260019350600092505b86518360ff161015610aad57868360ff168151811015156109cd57fe5b90602001015160f860020a900460f860020a0260f860020a900460ff169150600090505b8360ff168160ff161015610a6257848160ff16815181101515610a1057fe5b9060200190602002015160ff166101000282019150603a82811515610a3157fe5b06858260ff16815181101515610a4357fe5b60ff909216602092830290910190910152603a820491506001016109f1565b6000821115610aa257603a8206858560ff16815181101515610a8057fe5b60ff909216602092830290910190910152600190930192603a82049150610a62565b8260010192506109b0565b610ac7610ac2610abd8787610c1f565b610cb4565b610d4a565b979650505050505050565b606080606060008085516002016040519080825280601f01601f191660200182016040528015610b0c578160200160208202803883390190505b508051909450849350600192507f50000000000000000000000000000000000000000000000000000000000000009084906000908110610b4857fe5b906020010190600160f860020a031916908160001a905350825160018301927f3100000000000000000000000000000000000000000000000000000000000000918591908110610b9457fe5b906020010190600160f860020a031916908160001a905350600090505b85518160ff161015610c1557858160ff16815181101515610bce57fe5b90602001015160f860020a900460f860020a028383806001019450815181101515610bf557fe5b906020010190600160f860020a031916908160001a905350600101610bb1565b5091949350505050565b60608060008360ff16604051908082528060200260200182016040528015610c51578160200160208202803883390190505b509150600090505b8360ff168160ff161015610cac57848160ff16815181101515610c7857fe5b90602001906020020151828260ff16815181101515610c9357fe5b60ff909216602092830290910190910152600101610c59565b509392505050565b60608060008351604051908082528060200260200182016040528015610ce4578160200160208202803883390190505b509150600090505b83518160ff161015610d43578351849060ff8316810360001901908110610d0f57fe5b90602001906020020151828260ff16815181101515610d2a57fe5b60ff909216602092830290910190910152600101610cec565b5092915050565b606080600083516040519080825280601f01601f191660200182016040528015610d7e578160200160208202803883390190505b509150600090505b83518160ff161015610d4357606060405190810160405280603a81526020017f31323334353637383941424344454647484a4b4c4d4e5051525354555657585981526020017f5a6162636465666768696a6b6d6e6f707172737475767778797a000000000000815250848260ff16815181101515610e0057fe5b9060200190602002015160ff16815181101515610e1957fe5b90602001015160f860020a900460f860020a02828260ff16815181101515610e3d57fe5b906020010190600160f860020a031916908160001a905350600101610d865600a165627a7a72305820ef4385fc314b289231e641735c357210d96b0873965f0def341c13e4e9533ace0029`

// DeployPTNMap deploys a new Ethereum contract, binding an instance of PTNMap to it.
func DeployPTNMap(auth *bind.TransactOpts, backend bind.ContractBackend, _erc20Addr common.Address) (common.Address, *types.Transaction, *PTNMap, error) {
	parsed, err := abi.JSON(strings.NewReader(PTNMapABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(PTNMapBin), backend, _erc20Addr)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &PTNMap{PTNMapCaller: PTNMapCaller{contract: contract}, PTNMapTransactor: PTNMapTransactor{contract: contract}, PTNMapFilterer: PTNMapFilterer{contract: contract}}, nil
}

// PTNMap is an auto generated Go binding around an Ethereum contract.
type PTNMap struct {
	PTNMapCaller     // Read-only binding to the contract
	PTNMapTransactor // Write-only binding to the contract
	PTNMapFilterer   // Log filterer for contract events
}

// PTNMapCaller is an auto generated read-only Go binding around an Ethereum contract.
type PTNMapCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// PTNMapTransactor is an auto generated write-only Go binding around an Ethereum contract.
type PTNMapTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// PTNMapFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type PTNMapFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// PTNMapSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type PTNMapSession struct {
	Contract     *PTNMap           // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// PTNMapCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type PTNMapCallerSession struct {
	Contract *PTNMapCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts // Call options to use throughout this session
}

// PTNMapTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type PTNMapTransactorSession struct {
	Contract     *PTNMapTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// PTNMapRaw is an auto generated low-level Go binding around an Ethereum contract.
type PTNMapRaw struct {
	Contract *PTNMap // Generic contract binding to access the raw methods on
}

// PTNMapCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type PTNMapCallerRaw struct {
	Contract *PTNMapCaller // Generic read-only contract binding to access the raw methods on
}

// PTNMapTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type PTNMapTransactorRaw struct {
	Contract *PTNMapTransactor // Generic write-only contract binding to access the raw methods on
}

// NewPTNMap creates a new instance of PTNMap, bound to a specific deployed contract.
func NewPTNMap(address common.Address, backend bind.ContractBackend) (*PTNMap, error) {
	contract, err := bindPTNMap(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &PTNMap{PTNMapCaller: PTNMapCaller{contract: contract}, PTNMapTransactor: PTNMapTransactor{contract: contract}, PTNMapFilterer: PTNMapFilterer{contract: contract}}, nil
}

// NewPTNMapCaller creates a new read-only instance of PTNMap, bound to a specific deployed contract.
func NewPTNMapCaller(address common.Address, caller bind.ContractCaller) (*PTNMapCaller, error) {
	contract, err := bindPTNMap(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &PTNMapCaller{contract: contract}, nil
}

// NewPTNMapTransactor creates a new write-only instance of PTNMap, bound to a specific deployed contract.
func NewPTNMapTransactor(address common.Address, transactor bind.ContractTransactor) (*PTNMapTransactor, error) {
	contract, err := bindPTNMap(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &PTNMapTransactor{contract: contract}, nil
}

// NewPTNMapFilterer creates a new log filterer instance of PTNMap, bound to a specific deployed contract.
func NewPTNMapFilterer(address common.Address, filterer bind.ContractFilterer) (*PTNMapFilterer, error) {
	contract, err := bindPTNMap(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &PTNMapFilterer{contract: contract}, nil
}

// bindPTNMap binds a generic wrapper to an already deployed contract.
func bindPTNMap(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(PTNMapABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_PTNMap *PTNMapRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _PTNMap.Contract.PTNMapCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_PTNMap *PTNMapRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _PTNMap.Contract.PTNMapTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_PTNMap *PTNMapRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _PTNMap.Contract.PTNMapTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_PTNMap *PTNMapCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _PTNMap.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_PTNMap *PTNMapTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _PTNMap.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_PTNMap *PTNMapTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _PTNMap.Contract.contract.Transact(opts, method, params...)
}

// Addrmap is a free data retrieval call binding the contract method 0x8c5cecaa.
//
// Solidity: function addrmap( address) constant returns(address)
func (_PTNMap *PTNMapCaller) Addrmap(opts *bind.CallOpts, arg0 common.Address) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _PTNMap.contract.Call(opts, out, "addrmap", arg0)
	return *ret0, err
}

// Addrmap is a free data retrieval call binding the contract method 0x8c5cecaa.
//
// Solidity: function addrmap( address) constant returns(address)
func (_PTNMap *PTNMapSession) Addrmap(arg0 common.Address) (common.Address, error) {
	return _PTNMap.Contract.Addrmap(&_PTNMap.CallOpts, arg0)
}

// Addrmap is a free data retrieval call binding the contract method 0x8c5cecaa.
//
// Solidity: function addrmap( address) constant returns(address)
func (_PTNMap *PTNMapCallerSession) Addrmap(arg0 common.Address) (common.Address, error) {
	return _PTNMap.Contract.Addrmap(&_PTNMap.CallOpts, arg0)
}

// AddrmapPTN is a free data retrieval call binding the contract method 0x4e11092f.
//
// Solidity: function addrmapPTN( address) constant returns(address)
func (_PTNMap *PTNMapCaller) AddrmapPTN(opts *bind.CallOpts, arg0 common.Address) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _PTNMap.contract.Call(opts, out, "addrmapPTN", arg0)
	return *ret0, err
}

// AddrmapPTN is a free data retrieval call binding the contract method 0x4e11092f.
//
// Solidity: function addrmapPTN( address) constant returns(address)
func (_PTNMap *PTNMapSession) AddrmapPTN(arg0 common.Address) (common.Address, error) {
	return _PTNMap.Contract.AddrmapPTN(&_PTNMap.CallOpts, arg0)
}

// AddrmapPTN is a free data retrieval call binding the contract method 0x4e11092f.
//
// Solidity: function addrmapPTN( address) constant returns(address)
func (_PTNMap *PTNMapCallerSession) AddrmapPTN(arg0 common.Address) (common.Address, error) {
	return _PTNMap.Contract.AddrmapPTN(&_PTNMap.CallOpts, arg0)
}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(owner address, spender address) constant returns(uint256)
func (_PTNMap *PTNMapCaller) Allowance(opts *bind.CallOpts, owner common.Address, spender common.Address) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _PTNMap.contract.Call(opts, out, "allowance", owner, spender)
	return *ret0, err
}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(owner address, spender address) constant returns(uint256)
func (_PTNMap *PTNMapSession) Allowance(owner common.Address, spender common.Address) (*big.Int, error) {
	return _PTNMap.Contract.Allowance(&_PTNMap.CallOpts, owner, spender)
}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(owner address, spender address) constant returns(uint256)
func (_PTNMap *PTNMapCallerSession) Allowance(owner common.Address, spender common.Address) (*big.Int, error) {
	return _PTNMap.Contract.Allowance(&_PTNMap.CallOpts, owner, spender)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(_owner address) constant returns(uint256)
func (_PTNMap *PTNMapCaller) BalanceOf(opts *bind.CallOpts, _owner common.Address) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _PTNMap.contract.Call(opts, out, "balanceOf", _owner)
	return *ret0, err
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(_owner address) constant returns(uint256)
func (_PTNMap *PTNMapSession) BalanceOf(_owner common.Address) (*big.Int, error) {
	return _PTNMap.Contract.BalanceOf(&_PTNMap.CallOpts, _owner)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(_owner address) constant returns(uint256)
func (_PTNMap *PTNMapCallerSession) BalanceOf(_owner common.Address) (*big.Int, error) {
	return _PTNMap.Contract.BalanceOf(&_PTNMap.CallOpts, _owner)
}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() constant returns(uint8)
func (_PTNMap *PTNMapCaller) Decimals(opts *bind.CallOpts) (uint8, error) {
	var (
		ret0 = new(uint8)
	)
	out := ret0
	err := _PTNMap.contract.Call(opts, out, "decimals")
	return *ret0, err
}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() constant returns(uint8)
func (_PTNMap *PTNMapSession) Decimals() (uint8, error) {
	return _PTNMap.Contract.Decimals(&_PTNMap.CallOpts)
}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() constant returns(uint8)
func (_PTNMap *PTNMapCallerSession) Decimals() (uint8, error) {
	return _PTNMap.Contract.Decimals(&_PTNMap.CallOpts)
}

// EncodeBase58 is a free data retrieval call binding the contract method 0x927f526f.
//
// Solidity: function encodeBase58(addrHex address) constant returns(string)
func (_PTNMap *PTNMapCaller) EncodeBase58(opts *bind.CallOpts, addrHex common.Address) (string, error) {
	var (
		ret0 = new(string)
	)
	out := ret0
	err := _PTNMap.contract.Call(opts, out, "encodeBase58", addrHex)
	return *ret0, err
}

// EncodeBase58 is a free data retrieval call binding the contract method 0x927f526f.
//
// Solidity: function encodeBase58(addrHex address) constant returns(string)
func (_PTNMap *PTNMapSession) EncodeBase58(addrHex common.Address) (string, error) {
	return _PTNMap.Contract.EncodeBase58(&_PTNMap.CallOpts, addrHex)
}

// EncodeBase58 is a free data retrieval call binding the contract method 0x927f526f.
//
// Solidity: function encodeBase58(addrHex address) constant returns(string)
func (_PTNMap *PTNMapCallerSession) EncodeBase58(addrHex common.Address) (string, error) {
	return _PTNMap.Contract.EncodeBase58(&_PTNMap.CallOpts, addrHex)
}

// GetMapEthAddr is a free data retrieval call binding the contract method 0x6e932a1c.
//
// Solidity: function getMapEthAddr(ptnAddr address) constant returns(address)
func (_PTNMap *PTNMapCaller) GetMapEthAddr(opts *bind.CallOpts, ptnAddr common.Address) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _PTNMap.contract.Call(opts, out, "getMapEthAddr", ptnAddr)
	return *ret0, err
}

// GetMapEthAddr is a free data retrieval call binding the contract method 0x6e932a1c.
//
// Solidity: function getMapEthAddr(ptnAddr address) constant returns(address)
func (_PTNMap *PTNMapSession) GetMapEthAddr(ptnAddr common.Address) (common.Address, error) {
	return _PTNMap.Contract.GetMapEthAddr(&_PTNMap.CallOpts, ptnAddr)
}

// GetMapEthAddr is a free data retrieval call binding the contract method 0x6e932a1c.
//
// Solidity: function getMapEthAddr(ptnAddr address) constant returns(address)
func (_PTNMap *PTNMapCallerSession) GetMapEthAddr(ptnAddr common.Address) (common.Address, error) {
	return _PTNMap.Contract.GetMapEthAddr(&_PTNMap.CallOpts, ptnAddr)
}

// GetMapPtnAddr is a free data retrieval call binding the contract method 0x48cedf90.
//
// Solidity: function getMapPtnAddr(addr address) constant returns(string)
func (_PTNMap *PTNMapCaller) GetMapPtnAddr(opts *bind.CallOpts, addr common.Address) (string, error) {
	var (
		ret0 = new(string)
	)
	out := ret0
	err := _PTNMap.contract.Call(opts, out, "getMapPtnAddr", addr)
	return *ret0, err
}

// GetMapPtnAddr is a free data retrieval call binding the contract method 0x48cedf90.
//
// Solidity: function getMapPtnAddr(addr address) constant returns(string)
func (_PTNMap *PTNMapSession) GetMapPtnAddr(addr common.Address) (string, error) {
	return _PTNMap.Contract.GetMapPtnAddr(&_PTNMap.CallOpts, addr)
}

// GetMapPtnAddr is a free data retrieval call binding the contract method 0x48cedf90.
//
// Solidity: function getMapPtnAddr(addr address) constant returns(string)
func (_PTNMap *PTNMapCallerSession) GetMapPtnAddr(addr common.Address) (string, error) {
	return _PTNMap.Contract.GetMapPtnAddr(&_PTNMap.CallOpts, addr)
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() constant returns(string)
func (_PTNMap *PTNMapCaller) Name(opts *bind.CallOpts) (string, error) {
	var (
		ret0 = new(string)
	)
	out := ret0
	err := _PTNMap.contract.Call(opts, out, "name")
	return *ret0, err
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() constant returns(string)
func (_PTNMap *PTNMapSession) Name() (string, error) {
	return _PTNMap.Contract.Name(&_PTNMap.CallOpts)
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() constant returns(string)
func (_PTNMap *PTNMapCallerSession) Name() (string, error) {
	return _PTNMap.Contract.Name(&_PTNMap.CallOpts)
}

// PtnToken is a free data retrieval call binding the contract method 0x0ac74e72.
//
// Solidity: function ptnToken() constant returns(address)
func (_PTNMap *PTNMapCaller) PtnToken(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _PTNMap.contract.Call(opts, out, "ptnToken")
	return *ret0, err
}

// PtnToken is a free data retrieval call binding the contract method 0x0ac74e72.
//
// Solidity: function ptnToken() constant returns(address)
func (_PTNMap *PTNMapSession) PtnToken() (common.Address, error) {
	return _PTNMap.Contract.PtnToken(&_PTNMap.CallOpts)
}

// PtnToken is a free data retrieval call binding the contract method 0x0ac74e72.
//
// Solidity: function ptnToken() constant returns(address)
func (_PTNMap *PTNMapCallerSession) PtnToken() (common.Address, error) {
	return _PTNMap.Contract.PtnToken(&_PTNMap.CallOpts)
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() constant returns(string)
func (_PTNMap *PTNMapCaller) Symbol(opts *bind.CallOpts) (string, error) {
	var (
		ret0 = new(string)
	)
	out := ret0
	err := _PTNMap.contract.Call(opts, out, "symbol")
	return *ret0, err
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() constant returns(string)
func (_PTNMap *PTNMapSession) Symbol() (string, error) {
	return _PTNMap.Contract.Symbol(&_PTNMap.CallOpts)
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() constant returns(string)
func (_PTNMap *PTNMapCallerSession) Symbol() (string, error) {
	return _PTNMap.Contract.Symbol(&_PTNMap.CallOpts)
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() constant returns(uint256)
func (_PTNMap *PTNMapCaller) TotalSupply(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _PTNMap.contract.Call(opts, out, "totalSupply")
	return *ret0, err
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() constant returns(uint256)
func (_PTNMap *PTNMapSession) TotalSupply() (*big.Int, error) {
	return _PTNMap.Contract.TotalSupply(&_PTNMap.CallOpts)
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() constant returns(uint256)
func (_PTNMap *PTNMapCallerSession) TotalSupply() (*big.Int, error) {
	return _PTNMap.Contract.TotalSupply(&_PTNMap.CallOpts)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(spender address, value uint256) returns(bool)
func (_PTNMap *PTNMapTransactor) Approve(opts *bind.TransactOpts, spender common.Address, value *big.Int) (*types.Transaction, error) {
	return _PTNMap.contract.Transact(opts, "approve", spender, value)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(spender address, value uint256) returns(bool)
func (_PTNMap *PTNMapSession) Approve(spender common.Address, value *big.Int) (*types.Transaction, error) {
	return _PTNMap.Contract.Approve(&_PTNMap.TransactOpts, spender, value)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(spender address, value uint256) returns(bool)
func (_PTNMap *PTNMapTransactorSession) Approve(spender common.Address, value *big.Int) (*types.Transaction, error) {
	return _PTNMap.Contract.Approve(&_PTNMap.TransactOpts, spender, value)
}

// ResetMapAddr is a paid mutator transaction binding the contract method 0x2da2ff34.
//
// Solidity: function resetMapAddr(_addr address, _ptnhex address) returns()
func (_PTNMap *PTNMapTransactor) ResetMapAddr(opts *bind.TransactOpts, _addr common.Address, _ptnhex common.Address) (*types.Transaction, error) {
	return _PTNMap.contract.Transact(opts, "resetMapAddr", _addr, _ptnhex)
}

// ResetMapAddr is a paid mutator transaction binding the contract method 0x2da2ff34.
//
// Solidity: function resetMapAddr(_addr address, _ptnhex address) returns()
func (_PTNMap *PTNMapSession) ResetMapAddr(_addr common.Address, _ptnhex common.Address) (*types.Transaction, error) {
	return _PTNMap.Contract.ResetMapAddr(&_PTNMap.TransactOpts, _addr, _ptnhex)
}

// ResetMapAddr is a paid mutator transaction binding the contract method 0x2da2ff34.
//
// Solidity: function resetMapAddr(_addr address, _ptnhex address) returns()
func (_PTNMap *PTNMapTransactorSession) ResetMapAddr(_addr common.Address, _ptnhex common.Address) (*types.Transaction, error) {
	return _PTNMap.Contract.ResetMapAddr(&_PTNMap.TransactOpts, _addr, _ptnhex)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(_ptnhex address, _amt uint256) returns(bool)
func (_PTNMap *PTNMapTransactor) Transfer(opts *bind.TransactOpts, _ptnhex common.Address, _amt *big.Int) (*types.Transaction, error) {
	return _PTNMap.contract.Transact(opts, "transfer", _ptnhex, _amt)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(_ptnhex address, _amt uint256) returns(bool)
func (_PTNMap *PTNMapSession) Transfer(_ptnhex common.Address, _amt *big.Int) (*types.Transaction, error) {
	return _PTNMap.Contract.Transfer(&_PTNMap.TransactOpts, _ptnhex, _amt)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(_ptnhex address, _amt uint256) returns(bool)
func (_PTNMap *PTNMapTransactorSession) Transfer(_ptnhex common.Address, _amt *big.Int) (*types.Transaction, error) {
	return _PTNMap.Contract.Transfer(&_PTNMap.TransactOpts, _ptnhex, _amt)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(from address, to address, value uint256) returns(bool)
func (_PTNMap *PTNMapTransactor) TransferFrom(opts *bind.TransactOpts, from common.Address, to common.Address, value *big.Int) (*types.Transaction, error) {
	return _PTNMap.contract.Transact(opts, "transferFrom", from, to, value)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(from address, to address, value uint256) returns(bool)
func (_PTNMap *PTNMapSession) TransferFrom(from common.Address, to common.Address, value *big.Int) (*types.Transaction, error) {
	return _PTNMap.Contract.TransferFrom(&_PTNMap.TransactOpts, from, to, value)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(from address, to address, value uint256) returns(bool)
func (_PTNMap *PTNMapTransactorSession) TransferFrom(from common.Address, to common.Address, value *big.Int) (*types.Transaction, error) {
	return _PTNMap.Contract.TransferFrom(&_PTNMap.TransactOpts, from, to, value)
}

// PTNMapApprovalIterator is returned from FilterApproval and is used to iterate over the raw logs and unpacked data for Approval events raised by the PTNMap contract.
type PTNMapApprovalIterator struct {
	Event *PTNMapApproval // Event containing the contract specifics and raw log

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
func (it *PTNMapApprovalIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PTNMapApproval)
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
		it.Event = new(PTNMapApproval)
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
func (it *PTNMapApprovalIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PTNMapApprovalIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PTNMapApproval represents a Approval event raised by the PTNMap contract.
type PTNMapApproval struct {
	Owner   common.Address
	Spender common.Address
	Value   *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterApproval is a free log retrieval operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: e Approval(owner indexed address, spender indexed address, value uint256)
func (_PTNMap *PTNMapFilterer) FilterApproval(opts *bind.FilterOpts, owner []common.Address, spender []common.Address) (*PTNMapApprovalIterator, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var spenderRule []interface{}
	for _, spenderItem := range spender {
		spenderRule = append(spenderRule, spenderItem)
	}

	logs, sub, err := _PTNMap.contract.FilterLogs(opts, "Approval", ownerRule, spenderRule)
	if err != nil {
		return nil, err
	}
	return &PTNMapApprovalIterator{contract: _PTNMap.contract, event: "Approval", logs: logs, sub: sub}, nil
}

// WatchApproval is a free log subscription operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: e Approval(owner indexed address, spender indexed address, value uint256)
func (_PTNMap *PTNMapFilterer) WatchApproval(opts *bind.WatchOpts, sink chan<- *PTNMapApproval, owner []common.Address, spender []common.Address) (event.Subscription, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var spenderRule []interface{}
	for _, spenderItem := range spender {
		spenderRule = append(spenderRule, spenderItem)
	}

	logs, sub, err := _PTNMap.contract.WatchLogs(opts, "Approval", ownerRule, spenderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PTNMapApproval)
				if err := _PTNMap.contract.UnpackLog(event, "Approval", log); err != nil {
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

// PTNMapTransferIterator is returned from FilterTransfer and is used to iterate over the raw logs and unpacked data for Transfer events raised by the PTNMap contract.
type PTNMapTransferIterator struct {
	Event *PTNMapTransfer // Event containing the contract specifics and raw log

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
func (it *PTNMapTransferIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PTNMapTransfer)
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
		it.Event = new(PTNMapTransfer)
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
func (it *PTNMapTransferIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PTNMapTransferIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PTNMapTransfer represents a Transfer event raised by the PTNMap contract.
type PTNMapTransfer struct {
	From  common.Address
	To    common.Address
	Value *big.Int
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterTransfer is a free log retrieval operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: e Transfer(from indexed address, to indexed address, value uint256)
func (_PTNMap *PTNMapFilterer) FilterTransfer(opts *bind.FilterOpts, from []common.Address, to []common.Address) (*PTNMapTransferIterator, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _PTNMap.contract.FilterLogs(opts, "Transfer", fromRule, toRule)
	if err != nil {
		return nil, err
	}
	return &PTNMapTransferIterator{contract: _PTNMap.contract, event: "Transfer", logs: logs, sub: sub}, nil
}

// WatchTransfer is a free log subscription operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: e Transfer(from indexed address, to indexed address, value uint256)
func (_PTNMap *PTNMapFilterer) WatchTransfer(opts *bind.WatchOpts, sink chan<- *PTNMapTransfer, from []common.Address, to []common.Address) (event.Subscription, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _PTNMap.contract.WatchLogs(opts, "Transfer", fromRule, toRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PTNMapTransfer)
				if err := _PTNMap.contract.UnpackLog(event, "Transfer", log); err != nil {
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
