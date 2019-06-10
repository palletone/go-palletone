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

const contractABI = "[{\"constant\":false,\"inputs\":[{\"name\":\"addra\",\"type\":\"address\"},{\"name\":\"addrb\",\"type\":\"address\"},{\"name\":\"addrc\",\"type\":\"address\"}],\"name\":\"setaddrs\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"reqid\",\"type\":\"string\"}],\"name\":\"getmultisig\",\"outputs\":[{\"name\":\"\",\"type\":\"uint8\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"suicideto\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"ptnaddr\",\"type\":\"string\"}],\"name\":\"deposit\",\"outputs\":[],\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"my_eth_bal\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"recver\",\"type\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\"},{\"name\":\"reqid\",\"type\":\"string\"},{\"name\":\"sigstr1\",\"type\":\"bytes\"},{\"name\":\"sigstr2\",\"type\":\"bytes\"}],\"name\":\"withdraw\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"name\":\"addra\",\"type\":\"address\"},{\"name\":\"addrb\",\"type\":\"address\"},{\"name\":\"addrc\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"fallback\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"token\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"user\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"amount\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"ptnaddr\",\"type\":\"string\"}],\"name\":\"Deposit\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"token\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"user\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"recver\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"amount\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"reqid\",\"type\":\"string\"},{\"indexed\":false,\"name\":\"confirmvalue\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"state\",\"type\":\"string\"}],\"name\":\"Withdraw\",\"type\":\"event\"}]"
const contractBin = `0x608060405234801561001057600080fd5b50604051606080610c508339810160409081528151602083015191909201516000805433600160a060020a0319918216178255600180548216600160a060020a03968716179055600280548216948616949094179093556003805490931693909116929092179055610bc890819061008890396000f3006080604052600436106100775763ffffffff7c01000000000000000000000000000000000000000000000000000000006000350416630f8f1e7f81146100895780634a72d184146100b85780638e644ec314610127578063a26e118614610148578063c8fc638a14610194578063e3b98fb8146101bb575b34801561008357600080fd5b50600080fd5b34801561009557600080fd5b506100b6600160a060020a03600435811690602435811690604435166102a0565b005b3480156100c457600080fd5b506040805160206004803580820135601f810184900484028501840190955284845261011194369492936024939284019190819084018382808284375094975061030e9650505050505050565b6040805160ff9092168252519081900360200190f35b34801561013357600080fd5b506100b6600160a060020a0360043516610379565b6040805160206004803580820135601f81018490048402850184019095528484526100b694369492936024939284019190819084018382808284375094975061039c9650505050505050565b3480156101a057600080fd5b506101a961046c565b60408051918252519081900360200190f35b3480156101c757600080fd5b50604080516020600460443581810135601f81018490048402850184019095528484526100b6948235600160a060020a031694602480359536959460649492019190819084018382808284375050604080516020601f89358b018035918201839004830284018301909452808352979a99988101979196509182019450925082915084018382808284375050604080516020601f89358b018035918201839004830284018301909452808352979a9998810197919650918201945092508291508401838280828437509497506104719650505050505050565b600054600160a060020a031633146102b757600080fd5b6000805473ffffffffffffffffffffffffffffffffffffffff19908116331790915560018054600160a060020a03958616908316179055600280549385169382169390931790925560038054919093169116179055565b60006004826040518082805190602001908083835b602083106103425780518252601f199092019160209182019101610323565b51815160209384036101000a600019018019909216911617905292019485525060405193849003019092205460ff16949350505050565b600054600160a060020a0316331461039057600080fd5b80600160a060020a0316ff5b7fef519b7eb82aaf6ac376a6df2d793843ebfd593de5f1a0601d3cc6ab49ebb39560003334846040518085600160a060020a0316815260200184600160a060020a0316600160a060020a0316815260200183815260200180602001828103825283818151815260200191508051906020019080838360005b8381101561042c578181015183820152602001610414565b50505050905090810190601f1680156104595780820380516001836020036101000a031916815260200191505b509550505050505060405180910390a150565b303190565b60606000806004866040518082805190602001908083835b602083106104a85780518252601f199092019160209182019101610489565b51815160209384036101000a600019018019909216911617905292019485525060405193849003019092205460ff161591506104e5905057600080fd5b60408051600380825260808201909252906020820160608038833950506001548251929550600160a060020a031691859150600090811061052257fe5b600160a060020a039283166020918202909201015260025484519116908490600190811061054c57fe5b600160a060020a039283166020918202909201015260035484519116908490600290811061057657fe5b90602001906020020190600160a060020a03169081600160a060020a03168152505060009150308888886040516020018085600160a060020a0316600160a060020a03166c0100000000000000000000000002815260140184600160a060020a0316600160a060020a03166c0100000000000000000000000002815260140183815260200182805190602001908083835b602083106106265780518252601f199092019160209182019101610607565b6001836020036101000a0380198251168184511680821785525050505050509050019450505050506040516020818303038152906040526040518082805190602001908083835b6020831061068c5780518252601f19909201916020918201910161066d565b6001836020036101000a038019825116818451168082178552505050505050905001915050604051809103902090506106c7838287876108b2565b9150600260ff831610156106da57600080fd5b60016004876040518082805190602001908083835b6020831061070e5780518252601f1990920191602091820191016106ef565b51815160209384036101000a600019018019909216911617905292019485525060405193849003018320805460ff191660ff959095169490941790935550600160a060020a038a1691506108fc8915029089906000818181858888f19350505050158015610780573d6000803e3d6000fd5b507ffa582145410f16bc37d3c04740e9718ecddce920ef2491fec7fdf3f238557dd96000338a8a8a876040518087600160a060020a0316815260200186600160a060020a0316600160a060020a0316815260200185600160a060020a0316600160a060020a03168152602001848152602001806020018360ff16815260200180602001838103835285818151815260200191508051906020019080838360005b83811015610838578181015183820152602001610820565b50505050905090810190601f1680156108655780820380516001836020036101000a031916815260200191505b50928303905250600881527f776974686472617700000000000000000000000000000000000000000000000060208201526040805191829003019650945050505050a15050505050505050565b6040805160038082526080820190925260009160609183916020820184803883390190505091506108e682878988886108fd565b5060006108f282610944565b979650505050505050565b81516000901561091e576109118584610a34565b905061091e868286610a47565b81511561093c5761092f8583610a34565b905061093c868286610a47565b505050505050565b604080516003808252608082019092526000916060918391829190602082018580388339019050509250600183600081518110151561097f57fe5b60ff90921660209283029091019091015282516001908490829081106109a157fe5b60ff9092166020928302909101909101528251600190849060029081106109c457fe5b60ff9092166020928302909101909101525060009050805b60038160ff161015610a2c57828160ff168151811015156109f957fe5b90602001906020020151858260ff16815181101515610a1457fe5b602090810290910101510291909101906001016109dc565b509392505050565b6000610a408383610ac7565b9392505050565b60005b60038160ff161015610ac157818160ff16815181101515610a6757fe5b90602001906020020151600160a060020a031683600160a060020a0316141515610a9057610ab9565b6001848260ff16815181101515610aa357fe5b60ff909216602092830290910190910152610ac1565b600101610a4a565b50505050565b60008060008084516041141515610ae15760009350610b93565b50505060208201516040830151606084015160001a601b60ff82161015610b0657601b015b8060ff16601b14158015610b1e57508060ff16601c14155b15610b2c5760009350610b93565b60408051600080825260208083018085528a905260ff8516838501526060830187905260808301869052925160019360a0808501949193601f19840193928390039091019190865af1158015610b86573d6000803e3d6000fd5b5050506020604051035193505b505050929150505600a165627a7a723058207ab9eef361575bf2ac1918f0d0068d92b6bb45ea0a29db6a77504bb65f6bd0870029`

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

func setJuryAddrs(contractAddr, gasPrice, gasLimit, addr1, addr2, addr3, privateKey string) error {
	var ethadaptor adaptoreth.AdaptorETH
	ethadaptor.Rawurl = gWallet.EthConfig.Rawurl

	callerAddr := ethadaptor.GetAddress(privateKey)

	//
	value := "0"
	//	gasPrice := "1000"
	//	gasLimit := "2100000"

	//
	method := "setaddrs"
	paramsArray := []string{
		addr1,
		addr2,
		addr3}
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

func deploy(gasPrice, gasLimit, addr1, addr2, addr3, privateKey string) error {
	var ethadaptor adaptoreth.AdaptorETH
	ethadaptor.Rawurl = gWallet.EthConfig.Rawurl

	callerAddr := ethadaptor.GetAddress(privateKey)

	//
	value := "0"
	//	gasPrice := "1000"
	//	gasLimit := "2100000"

	//
	paramsArray := []string{
		addr1,
		addr2,
		addr3}
	paramsJson, err := json.Marshal(paramsArray)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	//
	var invokeContractParams adaptor.GenDeployContractTXParams
	invokeContractParams.ContractABI = contractABI
	invokeContractParams.ContractBin = contractBin
	invokeContractParams.DeployerAddr = callerAddr
	invokeContractParams.Value = value
	invokeContractParams.GasPrice = gasPrice
	invokeContractParams.GasLimit = gasLimit
	invokeContractParams.Params = string(paramsJson)

	//1.gen tx
	resultTx, err := ethadaptor.GenDeployContractTX(&invokeContractParams)
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
	fmt.Println("Params : deploy, gasPrice, gasLimit, ethAddr1, ethAddr2, ethAddr3, ethPrivateKey")
	fmt.Println("Params : setaddrs, contractAddr, value, gasPrice, gasLimit, ethAddr1, ethAddr2, ethAddr3, ethPrivateKey")
	fmt.Println("Params : deposit, contractAddr, value, gasPrice, gasLimit, ptnAddr, ethPrivateKey")
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
	case "deploy": //deploy multisigContract
		if len(args) < 8 {
			fmt.Println("Params : deploy, gasPrice, gasLimit, ethAddr1, ethAddr2, ethAddr3, ethPrivateKey")
			return
		}
		err := deploy(args[2], args[3], args[4], args[5], args[6], args[7])
		if err != nil {
			fmt.Println(err.Error())
		}
	case "setaddrs": //set jury addrs to multisigContract
		if len(args) < 9 {
			fmt.Println("Params : setaddrs, contractAddr, value, gasPrice, gasLimit, ethAddr1, ethAddr2, ethAddr3, ethPrivateKey")
			return
		}
		err := setJuryAddrs(args[2], args[3], args[4], args[5], args[6], args[7], args[8])
		if err != nil {
			fmt.Println(err.Error())
		}
	case "deposit": //deposit eth to multisigContract
		if len(args) < 8 {
			fmt.Println("Params : deposit, contractAddr, value, gasPrice, gasLimit, ptnAddr, ethPrivateKey")
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
