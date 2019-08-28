package main

import (
	"bufio"
	//"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/naoina/toml"
	//"github.com/palletone/adaptor"
	//"github.com/palletone/eth-adaptor"
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

const contractABI = "[{\"constant\":true,\"inputs\":[{\"name\":\"reqid\",\"type\":\"string\"}],\"name\":\"getmultisig\",\"outputs\":[{\"name\":\"\",\"type\":\"uint8\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"suicideto\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"ptnaddr\",\"type\":\"string\"}],\"name\":\"deposit\",\"outputs\":[],\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"my_eth_bal\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"addra\",\"type\":\"address\"},{\"name\":\"addrb\",\"type\":\"address\"},{\"name\":\"addrc\",\"type\":\"address\"},{\"name\":\"addrd\",\"type\":\"address\"}],\"name\":\"setaddrs\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"recver\",\"type\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\"},{\"name\":\"reqid\",\"type\":\"string\"},{\"name\":\"sigstr1\",\"type\":\"bytes\"},{\"name\":\"sigstr2\",\"type\":\"bytes\"},{\"name\":\"sigstr3\",\"type\":\"bytes\"}],\"name\":\"withdraw\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"name\":\"addra\",\"type\":\"address\"},{\"name\":\"addrb\",\"type\":\"address\"},{\"name\":\"addrc\",\"type\":\"address\"},{\"name\":\"addrd\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"fallback\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"token\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"user\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"amount\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"ptnaddr\",\"type\":\"string\"}],\"name\":\"Deposit\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"token\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"user\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"recver\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"amount\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"reqid\",\"type\":\"string\"},{\"indexed\":false,\"name\":\"confirmvalue\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"state\",\"type\":\"string\"}],\"name\":\"Withdraw\",\"type\":\"event\"}]"
const contractBin = `0x608060405234801561001057600080fd5b50604051608080610d2683398101604090815281516020830151918301516060909301516000805433600160a060020a0319918216178255600180548216600160a060020a039586161790556002805482169585169590951790945560038054851695841695909517909455600480549093169116179055610c8e90819061009890396000f3006080604052600436106100775763ffffffff7c01000000000000000000000000000000000000000000000000000000006000350416634a72d18481146100895780638e644ec3146100f8578063a26e11861461011b578063c8fc638a14610167578063df5aa58d1461018e578063f77f0f54146101c1575b34801561008357600080fd5b50600080fd5b34801561009557600080fd5b506040805160206004803580820135601f81018490048402850184019095528484526100e29436949293602493928401919081908401838280828437509497506102e49650505050505050565b6040805160ff9092168252519081900360200190f35b34801561010457600080fd5b50610119600160a060020a036004351661034f565b005b6040805160206004803580820135601f81018490048402850184019095528484526101199436949293602493928401919081908401838280828437509497506103729650505050505050565b34801561017357600080fd5b5061017c610442565b60408051918252519081900360200190f35b34801561019a57600080fd5b50610119600160a060020a0360043581169060243581169060443581169060643516610447565b3480156101cd57600080fd5b50604080516020600460443581810135601f8101849004840285018401909552848452610119948235600160a060020a031694602480359536959460649492019190819084018382808284375050604080516020601f89358b018035918201839004830284018301909452808352979a99988101979196509182019450925082915084018382808284375050604080516020601f89358b018035918201839004830284018301909452808352979a99988101979196509182019450925082915084018382808284375050604080516020601f89358b018035918201839004830284018301909452808352979a9998810197919650918201945092508291508401838280828437509497506104c59650505050505050565b60006005826040518082805190602001908083835b602083106103185780518252601f1990920191602091820191016102f9565b51815160209384036101000a600019018019909216911617905292019485525060405193849003019092205460ff16949350505050565b600054600160a060020a0316331461036657600080fd5b80600160a060020a0316ff5b7fef519b7eb82aaf6ac376a6df2d793843ebfd593de5f1a0601d3cc6ab49ebb39560003334846040518085600160a060020a0316815260200184600160a060020a0316600160a060020a0316815260200183815260200180602001828103825283818151815260200191508051906020019080838360005b838110156104025781810151838201526020016103ea565b50505050905090810190601f16801561042f5780820380516001836020036101000a031916815260200191505b509550505050505060405180910390a150565b303190565b600054600160a060020a0316331461045e57600080fd5b600080543373ffffffffffffffffffffffffffffffffffffffff1991821617909155600180548216600160a060020a039687161790556002805482169486169490941790935560038054841692851692909217909155600480549092169216919091179055565b60606000806005876040518082805190602001908083835b602083106104fc5780518252601f1990920191602091820191016104dd565b51815160209384036101000a600019018019909216911617905292019485525060405193849003019092205460ff16159150610539905057600080fd5b60408051600480825260a08201909252906020820160808038833950506001548251929550600160a060020a031691859150600090811061057657fe5b600160a060020a03928316602091820290920101526002548451911690849060019081106105a057fe5b600160a060020a03928316602091820290920101526003548451911690849060029081106105ca57fe5b600160a060020a03928316602091820290920101526004548451911690849060039081106105f457fe5b90602001906020020190600160a060020a03169081600160a060020a03168152505060009150308989896040516020018085600160a060020a0316600160a060020a03166c0100000000000000000000000002815260140184600160a060020a0316600160a060020a03166c0100000000000000000000000002815260140183815260200182805190602001908083835b602083106106a45780518252601f199092019160209182019101610685565b6001836020036101000a0380198251168184511680821785525050505050509050019450505050506040516020818303038152906040526040518082805190602001908083835b6020831061070a5780518252601f1990920191602091820191016106eb565b6001836020036101000a038019825116818451168082178552505050505050905001915050604051809103902090506107468382888888610932565b9150600360ff8316101561075957600080fd5b60016005886040518082805190602001908083835b6020831061078d5780518252601f19909201916020918201910161076e565b51815160209384036101000a600019018019909216911617905292019485525060405193849003018320805460ff191660ff959095169490941790935550600160a060020a038b1691506108fc8a1502908a906000818181858888f193505050501580156107ff573d6000803e3d6000fd5b507ffa582145410f16bc37d3c04740e9718ecddce920ef2491fec7fdf3f238557dd96000338b8b8b876040518087600160a060020a0316815260200186600160a060020a0316600160a060020a0316815260200185600160a060020a0316600160a060020a03168152602001848152602001806020018360ff16815260200180602001838103835285818151815260200191508051906020019080838360005b838110156108b757818101518382015260200161089f565b50505050905090810190601f1680156108e45780820380516001836020036101000a031916815260200191505b50928303905250600881527f776974686472617700000000000000000000000000000000000000000000000060208201526040805191829003019650945050505050a1505050505050505050565b60408051600480825260a0820190925260009160609183916020820160808038833901905050915061096882888a898989610980565b506000610974826109e6565b98975050505050505050565b8251600090156109a1576109948685610afa565b90506109a1878287610b0d565b8251156109bf576109b28684610afa565b90506109bf878287610b0d565b8151156109dd576109d08683610afa565b90506109dd878287610b0d565b50505050505050565b60408051600480825260a082019092526000916060918391829190602082016080803883390190505092506001836000815181101515610a2257fe5b60ff9092166020928302909101909101528251600190849082908110610a4457fe5b60ff909216602092830290910190910152825160019084906002908110610a6757fe5b60ff909216602092830290910190910152825160019084906003908110610a8a57fe5b60ff9092166020928302909101909101525060009050805b60048160ff161015610af257828160ff16815181101515610abf57fe5b90602001906020020151858260ff16815181101515610ada57fe5b60209081029091010151029190910190600101610aa2565b509392505050565b6000610b068383610b8d565b9392505050565b60005b60048160ff161015610b8757818160ff16815181101515610b2d57fe5b90602001906020020151600160a060020a031683600160a060020a0316141515610b5657610b7f565b6001848260ff16815181101515610b6957fe5b60ff909216602092830290910190910152610b87565b600101610b10565b50505050565b60008060008084516041141515610ba75760009350610c59565b50505060208201516040830151606084015160001a601b60ff82161015610bcc57601b015b8060ff16601b14158015610be457508060ff16601c14155b15610bf25760009350610c59565b60408051600080825260208083018085528a905260ff8516838501526060830187905260808301869052925160019360a0808501949193601f19840193928390039091019190865af1158015610c4c573d6000803e3d6000fd5b5050506020604051035193505b505050929150505600a165627a7a723058208ff3883a74b7c3539c5490e16395c3f7e7599fbc21f2a41cbcc9913954d410fe0029`

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
	if err != nil {
		return err
	}
	defer configFile.Close()

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
	//var ethadaptor adaptoreth.AdaptorETH
	////
	//key := ethadaptor.NewPrivateKey()
	//gWallet.NameKey[name] = key
	//
	////
	//pubkey := ethadaptor.GetPublicKey(key)
	//gWallet.NamePubkey[name] = pubkey
	//
	////
	//address := ethadaptor.GetAddress(key)
	//gWallet.NameAddress[name] = address
	//gWallet.AddressKey[address] = key

	return saveConfig(gWalletFile, gWallet)
}

func sendETHToMultiSigAddr(contractAddr, value, gasPrice, gasLimit, ptnAddr, privateKey string) error {
	//
	//	value := "1000000000000000000"
	//	gasPrice := "1000"
	//	gasLimit := "2100000"

	////
	//method := "deposit"
	//paramsArray := []string{ptnAddr}
	//paramsJson, err := json.Marshal(paramsArray)
	//if err != nil {
	//	fmt.Println(err.Error())
	//	return err
	//}
	//
	////
	//var ethadaptor adaptoreth.AdaptorETH
	//ethadaptor.Rawurl = gWallet.EthConfig.Rawurl
	//
	//callerAddr := ethadaptor.GetAddress(privateKey)
	////
	//var invokeContractParams adaptor.GenInvokeContractTXParams
	//invokeContractParams.ContractABI = contractABI
	//invokeContractParams.ContractAddr = contractAddr
	//invokeContractParams.CallerAddr = callerAddr //user
	//invokeContractParams.Value = value
	//invokeContractParams.GasPrice = gasPrice
	//invokeContractParams.GasLimit = gasLimit
	//invokeContractParams.Method = method //params
	//invokeContractParams.Params = string(paramsJson)
	//
	////1.gen tx
	//resultTx, err := ethadaptor.GenInvokeContractTX(&invokeContractParams)
	//if err != nil {
	//	fmt.Println(err.Error())
	//	return err
	//} else {
	//	fmt.Println(resultTx)
	//}
	////parse result
	//var genInvokeContractTXResult adaptor.GenInvokeContractTXResult
	//err = json.Unmarshal([]byte(resultTx), &genInvokeContractTXResult)
	//if err != nil {
	//	fmt.Println(err.Error())
	//}
	//
	////2.sign tx
	//var signTransactionParams adaptor.ETHSignTransactionParams
	//signTransactionParams.PrivateKeyHex = privateKey
	//signTransactionParams.TransactionHex = genInvokeContractTXResult.TransactionHex
	//resultSign, err := ethadaptor.SignTransaction(&signTransactionParams)
	//if err != nil {
	//	fmt.Println(err.Error())
	//	return err
	//} else {
	//	fmt.Println(resultSign)
	//}
	//
	////parse result
	//var signTransactionResult adaptor.SignTransactionResult
	//err = json.Unmarshal([]byte(resultSign), &signTransactionResult)
	//if err != nil {
	//	fmt.Println(err.Error())
	//}
	//
	////3.send tx
	//var sendTransactionParams adaptor.SendTransactionParams
	//sendTransactionParams.TransactionHex = signTransactionResult.TransactionHex
	//resultSend, err := ethadaptor.SendTransaction(&sendTransactionParams)
	//if err != nil {
	//	fmt.Println(err.Error())
	//	return err
	//} else {
	//	fmt.Println(resultSend)
	//}

	return nil
}

func spendEtHFromMultiAddr(contractAddr, gasPrice, gasLimit, ethRecvddr, amount, reqid, sig1, sig2, sig3, privateKey string) error {

	//var ethadaptor adaptoreth.AdaptorETH
	//ethadaptor.Rawurl = gWallet.EthConfig.Rawurl
	//
	//callerAddr := ethadaptor.GetAddress(privateKey)
	//
	////
	//value := "0"
	////	gasPrice := "1000"
	////	gasLimit := "2100000"
	//
	////
	//method := "withdraw"
	//paramsArray := []string{
	//	ethRecvddr,
	//	amount, //"1000000000000000000"
	//	reqid,
	//	sig1,
	//	sig2,
	//	sig3}
	//paramsJson, err := json.Marshal(paramsArray)
	//if err != nil {
	//	fmt.Println(err.Error())
	//	return err
	//}
	//
	////
	//var invokeContractParams adaptor.GenInvokeContractTXParams
	//invokeContractParams.ContractABI = contractABI
	//invokeContractParams.ContractAddr = contractAddr
	//invokeContractParams.CallerAddr = callerAddr //user
	//invokeContractParams.Value = value
	//invokeContractParams.GasPrice = gasPrice
	//invokeContractParams.GasLimit = gasLimit
	//invokeContractParams.Method = method //params
	//invokeContractParams.Params = string(paramsJson)
	//
	////1.gen tx
	//resultTx, err := ethadaptor.GenInvokeContractTX(&invokeContractParams)
	//if err != nil {
	//	fmt.Println(err.Error())
	//	return err
	//} else {
	//	fmt.Println(resultTx)
	//}
	////parse result
	//var genInvokeContractTXResult adaptor.GenInvokeContractTXResult
	//err = json.Unmarshal([]byte(resultTx), &genInvokeContractTXResult)
	//if err != nil {
	//	fmt.Println(err.Error())
	//}
	//
	////2.sign tx
	//var signTransactionParams adaptor.ETHSignTransactionParams
	//signTransactionParams.PrivateKeyHex = privateKey
	//signTransactionParams.TransactionHex = genInvokeContractTXResult.TransactionHex
	//resultSign, err := ethadaptor.SignTransaction(&signTransactionParams)
	//if err != nil {
	//	fmt.Println(err.Error())
	//	return err
	//} else {
	//	fmt.Println(resultSign)
	//}
	//
	////parse result
	//var signTransactionResult adaptor.ETHSignTransactionResult
	//err = json.Unmarshal([]byte(resultSign), &signTransactionResult)
	//if err != nil {
	//	fmt.Println(err.Error())
	//}
	//
	////3.send tx
	//var sendTransactionParams adaptor.SendTransactionParams
	//sendTransactionParams.TransactionHex = signTransactionResult.TransactionHex
	//resultSend, err := ethadaptor.SendTransaction(&sendTransactionParams)
	//if err != nil {
	//	fmt.Println(err.Error())
	//	return err
	//} else {
	//	fmt.Println(resultSend)
	//}

	return nil
}

func setJuryAddrs(contractAddr, gasPrice, gasLimit, addr1, addr2, addr3, addr4, privateKey string) error {
	//var ethadaptor adaptoreth.AdaptorETH
	//ethadaptor.Rawurl = gWallet.EthConfig.Rawurl
	//
	//callerAddr := ethadaptor.GetAddress(privateKey)
	//
	////
	//value := "0"
	////	gasPrice := "1000"
	////	gasLimit := "2100000"
	//
	////
	//method := "setaddrs"
	//paramsArray := []string{
	//	addr1,
	//	addr2,
	//	addr3,
	//	addr4}
	//paramsJson, err := json.Marshal(paramsArray)
	//if err != nil {
	//	fmt.Println(err.Error())
	//	return err
	//}
	//
	////
	//var invokeContractParams adaptor.GenInvokeContractTXParams
	//invokeContractParams.ContractABI = contractABI
	//invokeContractParams.ContractAddr = contractAddr
	//invokeContractParams.CallerAddr = callerAddr //user
	//invokeContractParams.Value = value
	//invokeContractParams.GasPrice = gasPrice
	//invokeContractParams.GasLimit = gasLimit
	//invokeContractParams.Method = method //params
	//invokeContractParams.Params = string(paramsJson)
	//
	////1.gen tx
	//resultTx, err := ethadaptor.GenInvokeContractTX(&invokeContractParams)
	//if err != nil {
	//	fmt.Println(err.Error())
	//	return err
	//} else {
	//	fmt.Println(resultTx)
	//}
	////parse result
	//var genInvokeContractTXResult adaptor.GenInvokeContractTXResult
	//err = json.Unmarshal([]byte(resultTx), &genInvokeContractTXResult)
	//if err != nil {
	//	fmt.Println(err.Error())
	//}
	//
	////2.sign tx
	//var signTransactionParams adaptor.ETHSignTransactionParams
	//signTransactionParams.PrivateKeyHex = privateKey
	//signTransactionParams.TransactionHex = genInvokeContractTXResult.TransactionHex
	//resultSign, err := ethadaptor.SignTransaction(&signTransactionParams)
	//if err != nil {
	//	fmt.Println(err.Error())
	//	return err
	//} else {
	//	fmt.Println(resultSign)
	//}
	//
	////parse result
	//var signTransactionResult adaptor.ETHSignTransactionResult
	//err = json.Unmarshal([]byte(resultSign), &signTransactionResult)
	//if err != nil {
	//	fmt.Println(err.Error())
	//}
	//
	////3.send tx
	//var sendTransactionParams adaptor.SendTransactionParams
	//sendTransactionParams.TransactionHex = signTransactionResult.TransactionHex
	//resultSend, err := ethadaptor.SendTransaction(&sendTransactionParams)
	//if err != nil {
	//	fmt.Println(err.Error())
	//	return err
	//} else {
	//	fmt.Println(resultSend)
	//}

	return nil
}

func deploy(gasPrice, gasLimit, addr1, addr2, addr3, addr4, privateKey string) error {
	//var ethadaptor adaptoreth.AdaptorETH
	//ethadaptor.Rawurl = gWallet.EthConfig.Rawurl
	//
	//callerAddr := ethadaptor.GetAddress(privateKey)
	//
	////
	//value := "0"
	////	gasPrice := "1000"
	////	gasLimit := "2100000"
	//
	////
	//paramsArray := []string{
	//	addr1,
	//	addr2,
	//	addr3,
	//	addr4}
	//paramsJson, err := json.Marshal(paramsArray)
	//if err != nil {
	//	fmt.Println(err.Error())
	//	return err
	//}
	//
	////
	//var invokeContractParams adaptor.GenDeployContractTXParams
	//invokeContractParams.ContractABI = contractABI
	//invokeContractParams.ContractBin = contractBin
	//invokeContractParams.DeployerAddr = callerAddr
	//invokeContractParams.Value = value
	//invokeContractParams.GasPrice = gasPrice
	//invokeContractParams.GasLimit = gasLimit
	//invokeContractParams.Params = string(paramsJson)
	//
	////1.gen tx
	//resultTx, err := ethadaptor.GenDeployContractTX(&invokeContractParams)
	//if err != nil {
	//	fmt.Println(err.Error())
	//	return err
	//} else {
	//	fmt.Println(resultTx)
	//}
	////parse result
	//var genDeployContractTXResult adaptor.GenDeployContractTXResult
	//err = json.Unmarshal([]byte(resultTx), &genDeployContractTXResult)
	//if err != nil {
	//	fmt.Println(err.Error())
	//}
	//fmt.Println("ContractAddr:", genDeployContractTXResult.ContractAddr)
	//
	////2.sign tx
	//var signTransactionParams adaptor.ETHSignTransactionParams
	//signTransactionParams.PrivateKeyHex = privateKey
	//signTransactionParams.TransactionHex = genDeployContractTXResult.TransactionHex
	//resultSign, err := ethadaptor.SignTransaction(&signTransactionParams)
	//if err != nil {
	//	fmt.Println(err.Error())
	//	return err
	//} else {
	//	fmt.Println(resultSign)
	//}
	//
	////parse result
	//var signTransactionResult adaptor.ETHSignTransactionResult
	//err = json.Unmarshal([]byte(resultSign), &signTransactionResult)
	//if err != nil {
	//	fmt.Println(err.Error())
	//}
	//
	////3.send tx
	//var sendTransactionParams adaptor.SendTransactionParams
	//sendTransactionParams.TransactionHex = signTransactionResult.TransactionHex
	//resultSend, err := ethadaptor.SendTransaction(&sendTransactionParams)
	//if err != nil {
	//	fmt.Println(err.Error())
	//	return err
	//} else {
	//	fmt.Println(resultSend)
	//}

	return nil
}

func helper() {
	fmt.Println("functions : send, withdraw")
	fmt.Println("Params : deploy, gasPrice, gasLimit, ethAddr1, ethAddr2, ethAddr3, ethAddr4, ethPrivateKey")
	fmt.Println("Params : setaddrs, contractAddr, gasPrice, gasLimit, ethAddr1, ethAddr2, ethAddr3, ethAddr4, ethPrivateKey")
	fmt.Println("Params : deposit, contractAddr, value, gasPrice, gasLimit, ptnAddr, ethPrivateKey")
	fmt.Println("Params : withdraw, contractAddr, gasPrice, gasLimit, ethAddr, amount, reqid, sig1, sig2, sig3, ethPrivateKey")
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
		if len(args) < 9 {
			fmt.Println("Params : deploy, gasPrice, gasLimit, ethAddr1, ethAddr2, ethAddr3, ethAddr4, ethPrivateKey")
			return
		}
		err := deploy(args[2], args[3], args[4], args[5], args[6], args[7], args[8])
		if err != nil {
			fmt.Println(err.Error())
		}
	case "setaddrs": //set jury addrs to multisigContract
		if len(args) < 10 {
			fmt.Println("Params : setaddrs, contractAddr, gasPrice, gasLimit, ethAddr1, ethAddr2, ethAddr3, ethAddr4, ethPrivateKey")
			return
		}
		err := setJuryAddrs(args[2], args[3], args[4], args[5], args[6], args[7], args[8], args[9])
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
		if len(args) < 12 {
			fmt.Println("Params : withdraw, contractAddr, gasPrice, gasLimit, ethAddr, amount, reqid, sig1, sig2, sig3, ethPrivateKey")
			return
		}
		err := spendEtHFromMultiAddr(args[2], args[3], args[4], args[5], args[6], args[7], args[8], args[9], args[10], args[11])
		if err != nil {
			fmt.Println(err.Error())
		}
	default:
		fmt.Println("Invalid cmd.")
		helper()
	}
}
