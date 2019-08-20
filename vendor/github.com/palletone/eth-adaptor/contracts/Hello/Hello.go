// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package Hello

import (
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// HelloABI is the input ABI used to generate the binding from.
const HelloABI = "[{\"constant\":true,\"inputs\":[],\"name\":\"info\",\"outputs\":[{\"name\":\"\",\"type\":\"string\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"a\",\"type\":\"int256\"},{\"name\":\"a256\",\"type\":\"int256\"},{\"name\":\"au\",\"type\":\"uint256\"},{\"name\":\"a8\",\"type\":\"int8\"}],\"name\":\"testpraram3\",\"outputs\":[{\"name\":\"\",\"type\":\"int256\"},{\"name\":\"\",\"type\":\"int256\"},{\"name\":\"\",\"type\":\"uint256\"},{\"name\":\"\",\"type\":\"int8\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"a\",\"type\":\"int256\"},{\"name\":\"b\",\"type\":\"bool\"},{\"name\":\"str\",\"type\":\"string\"},{\"name\":\"addr\",\"type\":\"address\"},{\"name\":\"bs\",\"type\":\"bytes\"},{\"name\":\"bs32\",\"type\":\"bytes32\"}],\"name\":\"testpraram\",\"outputs\":[{\"name\":\"\",\"type\":\"int256\"},{\"name\":\"\",\"type\":\"bool\"},{\"name\":\"\",\"type\":\"string\"},{\"name\":\"\",\"type\":\"address\"},{\"name\":\"\",\"type\":\"bytes\"},{\"name\":\"\",\"type\":\"bytes32\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"a\",\"type\":\"uint256\"},{\"name\":\"b\",\"type\":\"bool\"},{\"name\":\"str\",\"type\":\"string\"},{\"name\":\"addr\",\"type\":\"address\"},{\"name\":\"bs\",\"type\":\"bytes\"},{\"name\":\"bs32\",\"type\":\"bytes28\"}],\"name\":\"testpraram2\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"},{\"name\":\"\",\"type\":\"bool\"},{\"name\":\"\",\"type\":\"string\"},{\"name\":\"\",\"type\":\"address\"},{\"name\":\"\",\"type\":\"bytes\"},{\"name\":\"\",\"type\":\"bytes28\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_str\",\"type\":\"string\"}],\"name\":\"saySomething\",\"outputs\":[{\"name\":\"\",\"type\":\"string\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"name\":\"_owner\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"}]"

// HelloBin is the compiled bytecode used for deploying new contracts.
const HelloBin = `0x608060405234801561001057600080fd5b50604051602080610743833981016040525160008054600160a060020a03909216600160a060020a03199092169190911790556106f1806100526000396000f3006080604052600436106100775763ffffffff7c0100000000000000000000000000000000000000000000000000000000600035041663370158ea811461007c57806363c5f5fa1461010657806364ef0ecf146101575780638da5cb5b14610314578063ddc33af114610345578063fe6b378314610479575b600080fd5b34801561008857600080fd5b506100916104d2565b6040805160208082528351818301528351919283929083019185019080838360005b838110156100cb5781810151838201526020016100b3565b50505050905090810190601f1680156100f85780820380516001836020036101000a031916815260200191505b509250505060405180910390f35b34801561011257600080fd5b5061012a60043560243560443560643560000b61055f565b60408051948552602085019390935283830191909152600090810b900b6060830152519081900360800190f35b34801561016357600080fd5b50604080516020600460443581810135601f8101849004840285018401909552848452610208948235946024803515159536959460649492019190819084018382808284375050604080516020601f818a01358b0180359182018390048302840183018552818452989b600160a060020a038b35169b909a90999401975091955091820193509150819084018382808284375094975050933594506105679350505050565b60408051878152861515602080830191909152600160a060020a038616606083015260a0820184905260c09282018381528751938301939093528651919291608084019160e08501919089019080838360005b8381101561027357818101518382015260200161025b565b50505050905090810190601f1680156102a05780820380516001836020036101000a031916815260200191505b50838103825285518152855160209182019187019080838360005b838110156102d35781810151838201526020016102bb565b50505050905090810190601f1680156103005780820380516001836020036101000a031916815260200191505b509850505050505050505060405180910390f35b34801561032057600080fd5b50610329610573565b60408051600160a060020a039092168252519081900360200190f35b34801561035157600080fd5b50604080516020600460443581810135601f81018490048402850184019095528484526103fd948235946024803515159536959460649492019190819084018382808284375050604080516020601f818a01358b0180359182018390048302840183018552818452989b600160a060020a038b35169b909a9099940197509195509182019350915081908401838280828437509497505050923563ffffffff1916935061056792505050565b60405180878152602001861515151581526020018060200185600160a060020a0316600160a060020a03168152602001806020018463ffffffff191663ffffffff19168152602001838103835287818151815260200191508051906020019080838360008381101561027357818101518382015260200161025b565b34801561048557600080fd5b506040805160206004803580820135601f81018490048402850184019095528484526100919436949293602493928401919081908401838280828437509497506105829650505050505050565b60018054604080516020600284861615610100026000190190941693909304601f810184900484028201840190925281815292918301828280156105575780601f1061052c57610100808354040283529160200191610557565b820191906000526020600020905b81548152906001019060200180831161053a57829003601f168201915b505050505081565b929391929091565b94959394929391929091565b600054600160a060020a031681565b805160609061059890600190602085019061062a565b5060018054604080516020600284861615610100026000190190941693909304601f8101849004840282018401909252818152929183018282801561061e5780601f106105f35761010080835404028352916020019161061e565b820191906000526020600020905b81548152906001019060200180831161060157829003601f168201915b50505050509050919050565b828054600181600116156101000203166002900490600052602060002090601f016020900481019282601f1061066b57805160ff1916838001178555610698565b82800160010185558215610698579182015b8281111561069857825182559160200191906001019061067d565b506106a49291506106a8565b5090565b6106c291905b808211156106a457600081556001016106ae565b905600a165627a7a723058202f46be86a7d80bf4d739247923467b412f8f61fd0441b54d5d1845babf033de30029`

// DeployHello deploys a new Ethereum contract, binding an instance of Hello to it.
func DeployHello(auth *bind.TransactOpts, backend bind.ContractBackend, _owner common.Address) (common.Address, *types.Transaction, *Hello, error) {
	parsed, err := abi.JSON(strings.NewReader(HelloABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(HelloBin), backend, _owner)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &Hello{HelloCaller: HelloCaller{contract: contract}, HelloTransactor: HelloTransactor{contract: contract}, HelloFilterer: HelloFilterer{contract: contract}}, nil
}

// Hello is an auto generated Go binding around an Ethereum contract.
type Hello struct {
	HelloCaller     // Read-only binding to the contract
	HelloTransactor // Write-only binding to the contract
	HelloFilterer   // Log filterer for contract events
}

// HelloCaller is an auto generated read-only Go binding around an Ethereum contract.
type HelloCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// HelloTransactor is an auto generated write-only Go binding around an Ethereum contract.
type HelloTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// HelloFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type HelloFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// HelloSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type HelloSession struct {
	Contract     *Hello            // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// HelloCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type HelloCallerSession struct {
	Contract *HelloCaller  // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts // Call options to use throughout this session
}

// HelloTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type HelloTransactorSession struct {
	Contract     *HelloTransactor  // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// HelloRaw is an auto generated low-level Go binding around an Ethereum contract.
type HelloRaw struct {
	Contract *Hello // Generic contract binding to access the raw methods on
}

// HelloCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type HelloCallerRaw struct {
	Contract *HelloCaller // Generic read-only contract binding to access the raw methods on
}

// HelloTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type HelloTransactorRaw struct {
	Contract *HelloTransactor // Generic write-only contract binding to access the raw methods on
}

// NewHello creates a new instance of Hello, bound to a specific deployed contract.
func NewHello(address common.Address, backend bind.ContractBackend) (*Hello, error) {
	contract, err := bindHello(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Hello{HelloCaller: HelloCaller{contract: contract}, HelloTransactor: HelloTransactor{contract: contract}, HelloFilterer: HelloFilterer{contract: contract}}, nil
}

// NewHelloCaller creates a new read-only instance of Hello, bound to a specific deployed contract.
func NewHelloCaller(address common.Address, caller bind.ContractCaller) (*HelloCaller, error) {
	contract, err := bindHello(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &HelloCaller{contract: contract}, nil
}

// NewHelloTransactor creates a new write-only instance of Hello, bound to a specific deployed contract.
func NewHelloTransactor(address common.Address, transactor bind.ContractTransactor) (*HelloTransactor, error) {
	contract, err := bindHello(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &HelloTransactor{contract: contract}, nil
}

// NewHelloFilterer creates a new log filterer instance of Hello, bound to a specific deployed contract.
func NewHelloFilterer(address common.Address, filterer bind.ContractFilterer) (*HelloFilterer, error) {
	contract, err := bindHello(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &HelloFilterer{contract: contract}, nil
}

// bindHello binds a generic wrapper to an already deployed contract.
func bindHello(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(HelloABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Hello *HelloRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Hello.Contract.HelloCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Hello *HelloRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Hello.Contract.HelloTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Hello *HelloRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Hello.Contract.HelloTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Hello *HelloCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Hello.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Hello *HelloTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Hello.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Hello *HelloTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Hello.Contract.contract.Transact(opts, method, params...)
}

// Info is a free data retrieval call binding the contract method 0x370158ea.
//
// Solidity: function info() constant returns(string)
func (_Hello *HelloCaller) Info(opts *bind.CallOpts) (string, error) {
	var (
		ret0 = new(string)
	)
	out := ret0
	err := _Hello.contract.Call(opts, out, "info")
	return *ret0, err
}

// Info is a free data retrieval call binding the contract method 0x370158ea.
//
// Solidity: function info() constant returns(string)
func (_Hello *HelloSession) Info() (string, error) {
	return _Hello.Contract.Info(&_Hello.CallOpts)
}

// Info is a free data retrieval call binding the contract method 0x370158ea.
//
// Solidity: function info() constant returns(string)
func (_Hello *HelloCallerSession) Info() (string, error) {
	return _Hello.Contract.Info(&_Hello.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() constant returns(address)
func (_Hello *HelloCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _Hello.contract.Call(opts, out, "owner")
	return *ret0, err
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() constant returns(address)
func (_Hello *HelloSession) Owner() (common.Address, error) {
	return _Hello.Contract.Owner(&_Hello.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() constant returns(address)
func (_Hello *HelloCallerSession) Owner() (common.Address, error) {
	return _Hello.Contract.Owner(&_Hello.CallOpts)
}

// Testpraram is a free data retrieval call binding the contract method 0x64ef0ecf.
//
// Solidity: function testpraram(a int256, b bool, str string, addr address, bs bytes, bs32 bytes32) constant returns(int256, bool, string, address, bytes, bytes32)
func (_Hello *HelloCaller) Testpraram(opts *bind.CallOpts, a *big.Int, b bool, str string, addr common.Address, bs []byte, bs32 [32]byte) (*big.Int, bool, string, common.Address, []byte, [32]byte, error) {
	var (
		ret0 = new(*big.Int)
		ret1 = new(bool)
		ret2 = new(string)
		ret3 = new(common.Address)
		ret4 = new([]byte)
		ret5 = new([32]byte)
	)
	out := &[]interface{}{
		ret0,
		ret1,
		ret2,
		ret3,
		ret4,
		ret5,
	}
	err := _Hello.contract.Call(opts, out, "testpraram", a, b, str, addr, bs, bs32)
	return *ret0, *ret1, *ret2, *ret3, *ret4, *ret5, err
}

// Testpraram is a free data retrieval call binding the contract method 0x64ef0ecf.
//
// Solidity: function testpraram(a int256, b bool, str string, addr address, bs bytes, bs32 bytes32) constant returns(int256, bool, string, address, bytes, bytes32)
func (_Hello *HelloSession) Testpraram(a *big.Int, b bool, str string, addr common.Address, bs []byte, bs32 [32]byte) (*big.Int, bool, string, common.Address, []byte, [32]byte, error) {
	return _Hello.Contract.Testpraram(&_Hello.CallOpts, a, b, str, addr, bs, bs32)
}

// Testpraram is a free data retrieval call binding the contract method 0x64ef0ecf.
//
// Solidity: function testpraram(a int256, b bool, str string, addr address, bs bytes, bs32 bytes32) constant returns(int256, bool, string, address, bytes, bytes32)
func (_Hello *HelloCallerSession) Testpraram(a *big.Int, b bool, str string, addr common.Address, bs []byte, bs32 [32]byte) (*big.Int, bool, string, common.Address, []byte, [32]byte, error) {
	return _Hello.Contract.Testpraram(&_Hello.CallOpts, a, b, str, addr, bs, bs32)
}

// Testpraram2 is a free data retrieval call binding the contract method 0xddc33af1.
//
// Solidity: function testpraram2(a uint256, b bool, str string, addr address, bs bytes, bs32 bytes28) constant returns(uint256, bool, string, address, bytes, bytes28)
func (_Hello *HelloCaller) Testpraram2(opts *bind.CallOpts, a *big.Int, b bool, str string, addr common.Address, bs []byte, bs32 [28]byte) (*big.Int, bool, string, common.Address, []byte, [28]byte, error) {
	var (
		ret0 = new(*big.Int)
		ret1 = new(bool)
		ret2 = new(string)
		ret3 = new(common.Address)
		ret4 = new([]byte)
		ret5 = new([28]byte)
	)
	out := &[]interface{}{
		ret0,
		ret1,
		ret2,
		ret3,
		ret4,
		ret5,
	}
	err := _Hello.contract.Call(opts, out, "testpraram2", a, b, str, addr, bs, bs32)
	return *ret0, *ret1, *ret2, *ret3, *ret4, *ret5, err
}

// Testpraram2 is a free data retrieval call binding the contract method 0xddc33af1.
//
// Solidity: function testpraram2(a uint256, b bool, str string, addr address, bs bytes, bs32 bytes28) constant returns(uint256, bool, string, address, bytes, bytes28)
func (_Hello *HelloSession) Testpraram2(a *big.Int, b bool, str string, addr common.Address, bs []byte, bs32 [28]byte) (*big.Int, bool, string, common.Address, []byte, [28]byte, error) {
	return _Hello.Contract.Testpraram2(&_Hello.CallOpts, a, b, str, addr, bs, bs32)
}

// Testpraram2 is a free data retrieval call binding the contract method 0xddc33af1.
//
// Solidity: function testpraram2(a uint256, b bool, str string, addr address, bs bytes, bs32 bytes28) constant returns(uint256, bool, string, address, bytes, bytes28)
func (_Hello *HelloCallerSession) Testpraram2(a *big.Int, b bool, str string, addr common.Address, bs []byte, bs32 [28]byte) (*big.Int, bool, string, common.Address, []byte, [28]byte, error) {
	return _Hello.Contract.Testpraram2(&_Hello.CallOpts, a, b, str, addr, bs, bs32)
}

// Testpraram3 is a free data retrieval call binding the contract method 0x63c5f5fa.
//
// Solidity: function testpraram3(a int256, a256 int256, au uint256, a8 int8) constant returns(int256, int256, uint256, int8)
func (_Hello *HelloCaller) Testpraram3(opts *bind.CallOpts, a *big.Int, a256 *big.Int, au *big.Int, a8 int8) (*big.Int, *big.Int, *big.Int, int8, error) {
	var (
		ret0 = new(*big.Int)
		ret1 = new(*big.Int)
		ret2 = new(*big.Int)
		ret3 = new(int8)
	)
	out := &[]interface{}{
		ret0,
		ret1,
		ret2,
		ret3,
	}
	err := _Hello.contract.Call(opts, out, "testpraram3", a, a256, au, a8)
	return *ret0, *ret1, *ret2, *ret3, err
}

// Testpraram3 is a free data retrieval call binding the contract method 0x63c5f5fa.
//
// Solidity: function testpraram3(a int256, a256 int256, au uint256, a8 int8) constant returns(int256, int256, uint256, int8)
func (_Hello *HelloSession) Testpraram3(a *big.Int, a256 *big.Int, au *big.Int, a8 int8) (*big.Int, *big.Int, *big.Int, int8, error) {
	return _Hello.Contract.Testpraram3(&_Hello.CallOpts, a, a256, au, a8)
}

// Testpraram3 is a free data retrieval call binding the contract method 0x63c5f5fa.
//
// Solidity: function testpraram3(a int256, a256 int256, au uint256, a8 int8) constant returns(int256, int256, uint256, int8)
func (_Hello *HelloCallerSession) Testpraram3(a *big.Int, a256 *big.Int, au *big.Int, a8 int8) (*big.Int, *big.Int, *big.Int, int8, error) {
	return _Hello.Contract.Testpraram3(&_Hello.CallOpts, a, a256, au, a8)
}

// SaySomething is a paid mutator transaction binding the contract method 0xfe6b3783.
//
// Solidity: function saySomething(_str string) returns(string)
func (_Hello *HelloTransactor) SaySomething(opts *bind.TransactOpts, _str string) (*types.Transaction, error) {
	return _Hello.contract.Transact(opts, "saySomething", _str)
}

// SaySomething is a paid mutator transaction binding the contract method 0xfe6b3783.
//
// Solidity: function saySomething(_str string) returns(string)
func (_Hello *HelloSession) SaySomething(_str string) (*types.Transaction, error) {
	return _Hello.Contract.SaySomething(&_Hello.TransactOpts, _str)
}

// SaySomething is a paid mutator transaction binding the contract method 0xfe6b3783.
//
// Solidity: function saySomething(_str string) returns(string)
func (_Hello *HelloTransactorSession) SaySomething(_str string) (*types.Transaction, error) {
	return _Hello.Contract.SaySomething(&_Hello.TransactOpts, _str)
}
