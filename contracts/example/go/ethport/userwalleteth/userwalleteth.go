package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/naoina/toml"

	"github.com/palletone/adaptor"
	"github.com/palletone/eth-adaptor"
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
)

const contractABI = "[{\"constant\":true,\"inputs\":[{\"name\":\"reqid\",\"type\":\"string\"}],\"name\":\"getmultisig\",\"outputs\":[{\"name\":\"\",\"type\":\"uint8\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"suicideto\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"ptnaddr\",\"type\":\"string\"}],\"name\":\"deposit\",\"outputs\":[],\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"my_eth_bal\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"recver\",\"type\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\"},{\"name\":\"reqid\",\"type\":\"string\"},{\"name\":\"sigstr1\",\"type\":\"bytes\"},{\"name\":\"sigstr2\",\"type\":\"bytes\"}],\"name\":\"withdraw\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"name\":\"addra\",\"type\":\"address\"},{\"name\":\"addrb\",\"type\":\"address\"},{\"name\":\"addrc\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"fallback\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"token\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"user\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"amount\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"ptnaddr\",\"type\":\"string\"}],\"name\":\"Deposit\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"token\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"user\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"recver\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"amount\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"reqid\",\"type\":\"string\"},{\"indexed\":false,\"name\":\"confirmvalue\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"state\",\"type\":\"string\"}],\"name\":\"Withdraw\",\"type\":\"event\"}]"

func NewWallet() *MyWallet {
	return &MyWallet{
		EthConfig: ETHConfig{
			NetID:  1,
			Rawurl: "https://ropsten.infura.io/",
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
	var ethadaptor adaptoreth.AdaptorETH
	//
	key := ethadaptor.NewPrivateKey()
	gWallet.NameKey[name] = key

	//
	pubkey := ethadaptor.GetPublicKey(key)
	gWallet.NamePubkey[name] = pubkey

	//
	address := ethadaptor.GetAddress(key)
	gWallet.NameAddress[name] = address
	gWallet.AddressKey[address] = key

	return saveConfig(gWalletFile, gWallet)
}

func sendETHToMultiSigAddr(contractAddr, value, gasPrice, gasLimit, ptnAddr, privateKey string) error {
	//
	//	value := "1000000000000000000"
	//	gasPrice := "1000"
	//	gasLimit := "2100000"

	//
	method := "deposit"
	paramsArray := []string{ptnAddr}
	paramsJson, err := json.Marshal(paramsArray)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	//
	var ethadaptor adaptoreth.AdaptorETH
	ethadaptor.Rawurl = gWallet.EthConfig.Rawurl

	callerAddr := ethadaptor.GetAddress(privateKey)
	//
	var invokeContractParams adaptor.GenInvokeContractTXParams
	invokeContractParams.ContractABI = contractABI
	invokeContractParams.ContractAddr = contractAddr
	invokeContractParams.CallerAddr = callerAddr //user
	invokeContractParams.Value = value
	invokeContractParams.GasPrice = gasPrice
	invokeContractParams.GasLimit = gasLimit
	invokeContractParams.Method = method //params
	invokeContractParams.Params = string(paramsJson)

	//1.gen tx
	resultTx, err := ethadaptor.GenInvokeContractTX(&invokeContractParams)
	if err != nil {
		fmt.Println(err.Error())
		return err
	} else {
		fmt.Println(resultTx)
	}
	//parse result
	var genInvokeContractTXResult adaptor.GenInvokeContractTXResult
	err = json.Unmarshal([]byte(resultTx), &genInvokeContractTXResult)
	if err != nil {
		fmt.Println(err.Error())
	}

	//2.sign tx
	var signTransactionParams adaptor.ETHSignTransactionParams
	signTransactionParams.PrivateKeyHex = privateKey
	signTransactionParams.TransactionHex = genInvokeContractTXResult.TransactionHex
	resultSign, err := ethadaptor.SignTransaction(&signTransactionParams)
	if err != nil {
		fmt.Println(err.Error())
		return err
	} else {
		fmt.Println(resultSign)
	}

	//parse result
	var signTransactionResult adaptor.SignTransactionResult
	err = json.Unmarshal([]byte(resultSign), &signTransactionResult)
	if err != nil {
		fmt.Println(err.Error())
	}

	//3.send tx
	var sendTransactionParams adaptor.SendTransactionParams
	sendTransactionParams.TransactionHex = signTransactionResult.TransactionHex
	resultSend, err := ethadaptor.SendTransaction(&sendTransactionParams)
	if err != nil {
		fmt.Println(err.Error())
		return err
	} else {
		fmt.Println(resultSend)
	}

	return nil
}

func spendEtHFromMultiAddr(contractAddr, gasPrice, gasLimit, ethRecvddr, amount, reqid, sig1, sig2, privateKey string) error {

	var ethadaptor adaptoreth.AdaptorETH
	ethadaptor.Rawurl = gWallet.EthConfig.Rawurl

	callerAddr := ethadaptor.GetAddress(privateKey)

	//
	value := "0"
	//	gasPrice := "1000"
	//	gasLimit := "2100000"

	//
	method := "withdraw"
	paramsArray := []string{
		ethRecvddr,
		amount, //"1000000000000000000"
		reqid,
		sig1,
		sig2}
	paramsJson, err := json.Marshal(paramsArray)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	//
	var invokeContractParams adaptor.GenInvokeContractTXParams
	invokeContractParams.ContractABI = contractABI
	invokeContractParams.ContractAddr = contractAddr
	invokeContractParams.CallerAddr = callerAddr //user
	invokeContractParams.Value = value
	invokeContractParams.GasPrice = gasPrice
	invokeContractParams.GasLimit = gasLimit
	invokeContractParams.Method = method //params
	invokeContractParams.Params = string(paramsJson)

	//1.gen tx
	resultTx, err := ethadaptor.GenInvokeContractTX(&invokeContractParams)
	if err != nil {
		fmt.Println(err.Error())
		return err
	} else {
		fmt.Println(resultTx)
	}
	//parse result
	var genInvokeContractTXResult adaptor.GenInvokeContractTXResult
	err = json.Unmarshal([]byte(resultTx), &genInvokeContractTXResult)
	if err != nil {
		fmt.Println(err.Error())
	}

	//2.sign tx
	var signTransactionParams adaptor.ETHSignTransactionParams
	signTransactionParams.PrivateKeyHex = privateKey
	signTransactionParams.TransactionHex = genInvokeContractTXResult.TransactionHex
	resultSign, err := ethadaptor.SignTransaction(&signTransactionParams)
	if err != nil {
		fmt.Println(err.Error())
		return err
	} else {
		fmt.Println(resultSign)
	}

	//parse result
	var signTransactionResult adaptor.ETHSignTransactionResult
	err = json.Unmarshal([]byte(resultSign), &signTransactionResult)
	if err != nil {
		fmt.Println(err.Error())
	}

	//3.send tx
	var sendTransactionParams adaptor.SendTransactionParams
	sendTransactionParams.TransactionHex = signTransactionResult.TransactionHex
	resultSend, err := ethadaptor.SendTransaction(&sendTransactionParams)
	if err != nil {
		fmt.Println(err.Error())
		return err
	} else {
		fmt.Println(resultSend)
	}

	return nil
}
func helper() {
	fmt.Println("functions : send, withdraw")
	fmt.Println("Params : send, contractAddr, value, gasPrice, gasLimit, ptnAddr, ethPrivateKey")
	fmt.Println("Params : withdraw, contractAddr, gasPrice, gasLimit, ethAddr, amount, reqid, sig1, sig2, ethPrivateKey")
}
func main() {
	f, err := os.Open(gWalletFile)
	if err != nil && os.IsNotExist(err) { //check wallet config file
		saveConfig(gWalletFile, gWallet)
	} else {
		f.Close()
		err = loadConfig(gWalletFile, gWallet) //load wallet config
		if err != nil {
			fmt.Println(err.Error())
			return
		}
	}

	args := os.Args
	if len(args) < 2 {
		helper()
		return
	}
	cmd := strings.ToLower(args[1])

	switch cmd {
	case "init": //init alice's key and bob's key
		createKey("alice")
		createKey("bob")
	case "send": //send eth to multisigContract
		if len(args) < 8 {
			fmt.Println("Params : send, contractAddr, value, gasPrice, gasLimit, ptnAddr, ethPrivateKey")
			return
		}
		err := sendETHToMultiSigAddr(args[2], args[3], args[4], args[5], args[6], args[7])
		if err != nil {
			fmt.Println(err.Error())
		}
	case "withdraw": //withdraw eth of multisigContract
		if len(args) < 11 {
			fmt.Println("Params : withdraw, contractAddr, gasPrice, gasLimit, ethAddr, amount, reqid, sig1, sig2, ethPrivateKey")
			return
		}
		err := spendEtHFromMultiAddr(args[2], args[3], args[4], args[5], args[6], args[7], args[8], args[9], args[10])
		if err != nil {
			fmt.Println(err.Error())
		}
	default:
		fmt.Println("Invalid cmd.")
		helper()
	}
}
