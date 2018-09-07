package main

import (
	"bufio"
	//	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/naoina/toml"
	//	"github.com/palletone/eth-adaptor"
)

type ETHConfig struct {
	NetID  int
	Rawurl string
}
type MyWallet struct {
	EthConfig   ETHConfig
	NameKey     map[string]string
	NamePubkey  map[string]string
	NameAddress map[string]string
	AddressKey  map[string]string
}

var (
	gWallet     = NewWallet()
	gWalletFile = "./ethwallet.toml"

	gTomlConfig = toml.DefaultConfig

	//multisig contract 2/3 withdraw
	contractABI  = "[{\"constant\":false,\"inputs\":[{\"name\":\"token\",\"type\":\"address\"},{\"name\":\"redeem\",\"type\":\"bytes\"},{\"name\":\"recver\",\"type\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\"},{\"name\":\"nonece\",\"type\":\"uint256\"},{\"name\":\"sigstr1\",\"type\":\"bytes\"},{\"name\":\"sigstr2\",\"type\":\"bytes\"}],\"name\":\"withdrawtoken\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"redeem\",\"type\":\"bytes\"},{\"name\":\"recver\",\"type\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\"},{\"name\":\"nonece\",\"type\":\"uint256\"},{\"name\":\"sigstr1\",\"type\":\"bytes\"},{\"name\":\"sigstr2\",\"type\":\"bytes\"}],\"name\":\"withdraw\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"suicideto\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"redeem\",\"type\":\"bytes\"}],\"name\":\"deposit\",\"outputs\":[],\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"address\"},{\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"tokens\",\"outputs\":[{\"name\":\"balance\",\"type\":\"uint256\"},{\"name\":\"nonece\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"token\",\"type\":\"address\"},{\"name\":\"redeem\",\"type\":\"bytes\"},{\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"deposittoken\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"my_eth_bal\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"addr\",\"type\":\"address\"},{\"name\":\"redeem\",\"type\":\"bytes\"}],\"name\":\"getmultisig\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"},{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"admin\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"name\":\"admin_\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"fallback\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"token\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"user\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"amount\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"redeem\",\"type\":\"bytes\"}],\"name\":\"Deposit\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"token\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"user\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"redeem\",\"type\":\"bytes\"},{\"indexed\":false,\"name\":\"recver\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"amount\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"confirmvalue\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"state\",\"type\":\"string\"}],\"name\":\"Withdraw\",\"type\":\"event\"}]"
	contractAddr = "0x6817Cfb2c442693d850332c3B755B2342Ec4aFB2"
)

func NewWallet() *MyWallet {
	return &MyWallet{
		EthConfig: ETHConfig{
			NetID:  1,
			Rawurl: "\\\\.\\pipe\\geth.ipc",
		},
		NameKey:     map[string]string{},
		NamePubkey:  map[string]string{},
		NameAddress: map[string]string{},
		AddressKey:  map[string]string{},
	}
}

func loadConfig(file string, w *MyWallet) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()

	err = gTomlConfig.NewDecoder(bufio.NewReader(f)).Decode(w)
	return err
}

func saveConfig(file string, w *MyWallet) error {
	configFile, err := os.Create(file)
	defer configFile.Close()
	if err != nil {
		return err
	}

	configToml, err := gTomlConfig.Marshal(w)
	if err != nil {
		return err
	}

	_, err = configFile.Write(configToml)
	if err != nil {
		return err
	}

	return nil
}

func createKey(name string) error {
	//	var ethadaptor adaptoreth.AdaptorETH
	//	//
	//	key := ethadaptor.NewPrivateKey()
	//	gWallet.NameKey[name] = key

	//	//
	//	pubkey := ethadaptor.GetPublicKey(key)
	//	gWallet.NamePubkey[name] = pubkey

	//	//
	//	address := ethadaptor.GetAddress(key)
	//	gWallet.NameAddress[name] = address
	//	gWallet.AddressKey[address] = key

	return saveConfig(gWalletFile, gWallet)
}

func bobSendETHToMultiSigAddr(value string, gasPrice string, gasLimit string, redeem string) error {
	//	//
	//	sender := "bob"

	//	//
	//	callerAddr := gWallet.NameAddress[sender]
	//	//	value := "1000000000000000000"
	//	//	gasPrice := "1000"
	//	//	gasLimit := "2100000"
	//	//
	//	method := "deposit"
	//	paramsArray := []string{redeem}
	//	paramsJson, err := json.Marshal(paramsArray)
	//	if err != nil {
	//		fmt.Println(err.Error())
	//		return err
	//	}

	//	//
	//	var ethadaptor adaptoreth.AdaptorETH
	//	ethadaptor.Rawurl = gWallet.EthConfig.Rawurl

	//	//
	//	var invokeContractParams adaptoreth.GenInvokeContractTXParams
	//	invokeContractParams.ContractABI = contractABI
	//	invokeContractParams.ContractAddr = contractAddr
	//	invokeContractParams.CallerAddr = callerAddr //user
	//	invokeContractParams.Value = value
	//	invokeContractParams.GasPrice = gasPrice
	//	invokeContractParams.GasLimit = gasLimit
	//	invokeContractParams.Method = method //params
	//	invokeContractParams.Params = string(paramsJson)

	//	//1.gen tx
	//	resultTx, err := ethadaptor.GenInvokeContractTX(&invokeContractParams)
	//	if err != nil {
	//		fmt.Println(err.Error())
	//		return err
	//	} else {
	//		fmt.Println(resultTx)
	//	}
	//	//parse result
	//	var genInvokeContractTXResult adaptoreth.GenInvokeContractTXResult
	//	err = json.Unmarshal([]byte(resultTx), &genInvokeContractTXResult)
	//	if err != nil {
	//		fmt.Println(err.Error())
	//	}

	//	//2.sign tx
	//	var signTransactionParams adaptoreth.SignTransactionParams
	//	signTransactionParams.PrivateKeyHex = gWallet.NameKey[sender]
	//	fmt.Println("gWallet.NameKey[sender] : ", gWallet.NameKey[sender])
	//	signTransactionParams.TransactionHex = genInvokeContractTXResult.TransactionHex
	//	resultSign, err := ethadaptor.SignTransaction(&signTransactionParams)
	//	if err != nil {
	//		fmt.Println(err.Error())
	//		return err
	//	} else {
	//		fmt.Println(resultSign)
	//	}

	//	//parse result
	//	var signTransactionResult adaptoreth.SignTransactionResult
	//	err = json.Unmarshal([]byte(resultSign), &signTransactionResult)
	//	if err != nil {
	//		fmt.Println(err.Error())
	//	}

	//	//3.send tx
	//	var sendTransactionParams adaptoreth.SendTransactionParams
	//	sendTransactionParams.TransactionHex = signTransactionResult.TransactionHex
	//	resultSend, err := ethadaptor.SendTransaction(&sendTransactionParams)
	//	if err != nil {
	//		fmt.Println(err.Error())
	//		return err
	//	} else {
	//		fmt.Println(resultSend)
	//	}

	return nil
}

func aliceSpendEtHFromMultiAddr(gasPrice string, gasLimit string, redeem string, amount string, sigJury string) error {
	//	//
	//	spender := "alice"

	//	//keccak256(abi.encodePacked(redeem, recver, address(this), amount, nonece));
	//	paramTypesArray := []string{"Bytes", "Address", "Address", "Uint", "Uint"} //eth
	//	paramTypesJson, err := json.Marshal(paramTypesArray)
	//	if err != nil {
	//		fmt.Println(err.Error())
	//	}
	//	calSigParamsArray := []string{
	//		redeem,
	//		gWallet.NameAddress[spender],
	//		contractAddr,
	//		amount,
	//		"1"}
	//	calSigParamsJson, err := json.Marshal(calSigParamsArray)
	//	if err != nil {
	//		fmt.Println(err.Error())
	//		return err
	//	}

	//	//
	//	var ethadaptor adaptoreth.AdaptorETH
	//	ethadaptor.Rawurl = gWallet.EthConfig.Rawurl

	//	//
	//	var sigParams adaptoreth.Keccak256HashPackedSigParams
	//	sigParams.ParamTypes = string(paramTypesJson)
	//	sigParams.Params = string(calSigParamsJson)
	//	sigParams.PrivateKeyHex = gWallet.NameKey[spender]

	//	//0.calculate Alice's signature
	//	resultSig, err := ethadaptor.Keccak256HashPackedSig(&sigParams)
	//	if err != nil {
	//		fmt.Println(err.Error())
	//		return err
	//	} else {
	//		fmt.Println(resultSig)
	//	}

	//	//parse result
	//	var calSigResult adaptoreth.Keccak256HashPackedSigResult
	//	err = json.Unmarshal([]byte(resultSig), &calSigResult)
	//	if err != nil {
	//		fmt.Println(err.Error())
	//	}
	//	sigAlice := calSigResult.Signature

	//	//
	//	callerAddr := gWallet.NameAddress[spender]
	//	value := "0"
	//	//	gasPrice := "1000"
	//	//	gasLimit := "2100000"
	//	//
	//	method := "withdraw"
	//	paramsArray := []string{
	//		redeem,
	//		callerAddr,
	//		amount, //"1000000000000000000"
	//		"1",    //nonece,
	//		sigJury,
	//		sigAlice}
	//	paramsJson, err := json.Marshal(paramsArray)
	//	if err != nil {
	//		fmt.Println(err.Error())
	//		return err
	//	}
	//	//
	//	var invokeContractParams adaptoreth.GenInvokeContractTXParams
	//	invokeContractParams.ContractABI = contractABI
	//	invokeContractParams.ContractAddr = contractAddr
	//	invokeContractParams.CallerAddr = callerAddr //user
	//	invokeContractParams.Value = value
	//	invokeContractParams.GasPrice = gasPrice
	//	invokeContractParams.GasLimit = gasLimit
	//	invokeContractParams.Method = method //params
	//	invokeContractParams.Params = string(paramsJson)

	//	//1.gen tx
	//	resultTx, err := ethadaptor.GenInvokeContractTX(&invokeContractParams)
	//	if err != nil {
	//		fmt.Println(err.Error())
	//		return err
	//	} else {
	//		fmt.Println(resultTx)
	//	}
	//	//parse result
	//	var genInvokeContractTXResult adaptoreth.GenInvokeContractTXResult
	//	err = json.Unmarshal([]byte(resultTx), &genInvokeContractTXResult)
	//	if err != nil {
	//		fmt.Println(err.Error())
	//	}

	//	//2.sign tx
	//	var signTransactionParams adaptoreth.SignTransactionParams
	//	signTransactionParams.PrivateKeyHex = gWallet.NameKey[spender]
	//	signTransactionParams.TransactionHex = genInvokeContractTXResult.TransactionHex
	//	resultSign, err := ethadaptor.SignTransaction(&signTransactionParams)
	//	if err != nil {
	//		fmt.Println(err.Error())
	//		return err
	//	} else {
	//		fmt.Println(resultSign)
	//	}

	//	//parse result
	//	var signTransactionResult adaptoreth.SignTransactionResult
	//	err = json.Unmarshal([]byte(resultSign), &signTransactionResult)
	//	if err != nil {
	//		fmt.Println(err.Error())
	//	}

	//	//3.send tx
	//	var sendTransactionParams adaptoreth.SendTransactionParams
	//	sendTransactionParams.TransactionHex = signTransactionResult.TransactionHex
	//	resultSend, err := ethadaptor.SendTransaction(&sendTransactionParams)
	//	if err != nil {
	//		fmt.Println(err.Error())
	//		return err
	//	} else {
	//		fmt.Println(resultSend)
	//	}

	return nil
}

func main() {
	err := loadConfig(gWalletFile, gWallet)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	args := os.Args
	cmd := strings.ToLower(args[1])

	switch cmd {
	case "init":
		createKey("alice")
		createKey("bob")
	case "bob":
		if len(args) < 6 {
			fmt.Println("Params : bob, value, gasPrice, gasLimit, redeem")
			return
		}
		err := bobSendETHToMultiSigAddr(args[2], args[3], args[4], args[5])
		if err != nil {
			fmt.Println(err.Error())
		}
	case "alice":
		if len(args) < 7 {
			fmt.Println("Params : alice, gasPrice, gasLimit, redeem, amount, sigJury")
			return
		}
		err := aliceSpendEtHFromMultiAddr(args[2], args[3], args[4], args[5], args[6])
		if err != nil {
			fmt.Println(err.Error())
		}
	default:
		fmt.Println("Invalid cmd.")
	}
}
