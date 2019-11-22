package main

import (
	"bufio"
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"math/big"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
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
	NameKey     map[string][]byte
	NamePubkey  map[string][]byte
	NameAddress map[string]string
	AddressKey  map[string][]byte
}

var (
	gWallet     = NewWallet()
	gWalletFile = "./ethwallet.toml"

	gTomlConfig = toml.DefaultConfig
)

const EthmultisigABI = "[{\"constant\":true,\"inputs\":[{\"name\":\"reqid\",\"type\":\"string\"}],\"name\":\"getmultisig\",\"outputs\":[{\"name\":\"\",\"type\":\"uint8\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"suicideto\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"ptnaddr\",\"type\":\"string\"}],\"name\":\"deposit\",\"outputs\":[],\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"my_eth_bal\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"addra\",\"type\":\"address\"},{\"name\":\"addrb\",\"type\":\"address\"},{\"name\":\"addrc\",\"type\":\"address\"},{\"name\":\"addrd\",\"type\":\"address\"}],\"name\":\"setaddrs\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"recver\",\"type\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\"},{\"name\":\"reqid\",\"type\":\"string\"},{\"name\":\"sigstr1\",\"type\":\"bytes\"},{\"name\":\"sigstr2\",\"type\":\"bytes\"},{\"name\":\"sigstr3\",\"type\":\"bytes\"}],\"name\":\"withdraw\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"name\":\"addra\",\"type\":\"address\"},{\"name\":\"addrb\",\"type\":\"address\"},{\"name\":\"addrc\",\"type\":\"address\"},{\"name\":\"addrd\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"fallback\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"token\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"user\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"amount\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"ptnaddr\",\"type\":\"string\"}],\"name\":\"Deposit\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"token\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"user\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"recver\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"amount\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"reqid\",\"type\":\"string\"},{\"indexed\":false,\"name\":\"confirmvalue\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"state\",\"type\":\"string\"}],\"name\":\"Withdraw\",\"type\":\"event\"}]"

const EthmultisigBin = `0x608060405234801561001057600080fd5b50604051608080610d0f83398101604090815281516020830151918301516060909301516000805433600160a060020a0319918216178255600180548216600160a060020a039586161790556002805482169585169590951790945560038054851695841695909517909455600480549093169116179055610c7790819061009890396000f3006080604052600436106100775763ffffffff7c01000000000000000000000000000000000000000000000000000000006000350416634a72d184811461007c5780638e644ec3146100eb578063a26e11861461010e578063c8fc638a1461015a578063df5aa58d14610181578063f77f0f54146101b4575b600080fd5b34801561008857600080fd5b506040805160206004803580820135601f81018490048402850184019095528484526100d59436949293602493928401919081908401838280828437509497506102d79650505050505050565b6040805160ff9092168252519081900360200190f35b3480156100f757600080fd5b5061010c600160a060020a0360043516610342565b005b6040805160206004803580820135601f810184900484028501840190955284845261010c9436949293602493928401919081908401838280828437509497506103659650505050505050565b34801561016657600080fd5b5061016f610435565b60408051918252519081900360200190f35b34801561018d57600080fd5b5061010c600160a060020a036004358116906024358116906044358116906064351661043a565b3480156101c057600080fd5b50604080516020600460443581810135601f810184900484028501840190955284845261010c948235600160a060020a031694602480359536959460649492019190819084018382808284375050604080516020601f89358b018035918201839004830284018301909452808352979a99988101979196509182019450925082915084018382808284375050604080516020601f89358b018035918201839004830284018301909452808352979a99988101979196509182019450925082915084018382808284375050604080516020601f89358b018035918201839004830284018301909452808352979a9998810197919650918201945092508291508401838280828437509497506104ae9650505050505050565b60006005826040518082805190602001908083835b6020831061030b5780518252601f1990920191602091820191016102ec565b51815160209384036101000a600019018019909216911617905292019485525060405193849003019092205460ff16949350505050565b600054600160a060020a0316331461035957600080fd5b80600160a060020a0316ff5b7fef519b7eb82aaf6ac376a6df2d793843ebfd593de5f1a0601d3cc6ab49ebb39560003334846040518085600160a060020a0316815260200184600160a060020a0316600160a060020a0316815260200183815260200180602001828103825283818151815260200191508051906020019080838360005b838110156103f55781810151838201526020016103dd565b50505050905090810190601f1680156104225780820380516001836020036101000a031916815260200191505b509550505050505060405180910390a150565b303190565b600054600160a060020a0316331461045157600080fd5b60018054600160a060020a0395861673ffffffffffffffffffffffffffffffffffffffff19918216179091556002805494861694821694909417909355600380549285169284169290921790915560048054919093169116179055565b60606000806005876040518082805190602001908083835b602083106104e55780518252601f1990920191602091820191016104c6565b51815160209384036101000a600019018019909216911617905292019485525060405193849003019092205460ff16159150610522905057600080fd5b60408051600480825260a08201909252906020820160808038833950506001548251929550600160a060020a031691859150600090811061055f57fe5b600160a060020a039283166020918202909201015260025484519116908490600190811061058957fe5b600160a060020a03928316602091820290920101526003548451911690849060029081106105b357fe5b600160a060020a03928316602091820290920101526004548451911690849060039081106105dd57fe5b90602001906020020190600160a060020a03169081600160a060020a03168152505060009150308989896040516020018085600160a060020a0316600160a060020a03166c0100000000000000000000000002815260140184600160a060020a0316600160a060020a03166c0100000000000000000000000002815260140183815260200182805190602001908083835b6020831061068d5780518252601f19909201916020918201910161066e565b6001836020036101000a0380198251168184511680821785525050505050509050019450505050506040516020818303038152906040526040518082805190602001908083835b602083106106f35780518252601f1990920191602091820191016106d4565b6001836020036101000a0380198251168184511680821785525050505050509050019150506040518091039020905061072f838288888861091b565b9150600360ff8316101561074257600080fd5b60016005886040518082805190602001908083835b602083106107765780518252601f199092019160209182019101610757565b51815160209384036101000a600019018019909216911617905292019485525060405193849003018320805460ff191660ff959095169490941790935550600160a060020a038b1691506108fc8a1502908a906000818181858888f193505050501580156107e8573d6000803e3d6000fd5b507ffa582145410f16bc37d3c04740e9718ecddce920ef2491fec7fdf3f238557dd96000338b8b8b876040518087600160a060020a0316815260200186600160a060020a0316600160a060020a0316815260200185600160a060020a0316600160a060020a03168152602001848152602001806020018360ff16815260200180602001838103835285818151815260200191508051906020019080838360005b838110156108a0578181015183820152602001610888565b50505050905090810190601f1680156108cd5780820380516001836020036101000a031916815260200191505b50928303905250600881527f776974686472617700000000000000000000000000000000000000000000000060208201526040805191829003019650945050505050a1505050505050505050565b60408051600480825260a0820190925260009160609183916020820160808038833901905050915061095182888a898989610969565b50600061095d826109cf565b98975050505050505050565b82516000901561098a5761097d8685610ae3565b905061098a878287610af6565b8251156109a85761099b8684610ae3565b90506109a8878287610af6565b8151156109c6576109b98683610ae3565b90506109c6878287610af6565b50505050505050565b60408051600480825260a082019092526000916060918391829190602082016080803883390190505092506001836000815181101515610a0b57fe5b60ff9092166020928302909101909101528251600190849082908110610a2d57fe5b60ff909216602092830290910190910152825160019084906002908110610a5057fe5b60ff909216602092830290910190910152825160019084906003908110610a7357fe5b60ff9092166020928302909101909101525060009050805b60048160ff161015610adb57828160ff16815181101515610aa857fe5b90602001906020020151858260ff16815181101515610ac357fe5b60209081029091010151029190910190600101610a8b565b509392505050565b6000610aef8383610b76565b9392505050565b60005b60048160ff161015610b7057818160ff16815181101515610b1657fe5b90602001906020020151600160a060020a031683600160a060020a0316141515610b3f57610b68565b6001848260ff16815181101515610b5257fe5b60ff909216602092830290910190910152610b70565b600101610af9565b50505050565b60008060008084516041141515610b905760009350610c42565b50505060208201516040830151606084015160001a601b60ff82161015610bb557601b015b8060ff16601b14158015610bcd57508060ff16601c14155b15610bdb5760009350610c42565b60408051600080825260208083018085528a905260ff8516838501526060830187905260808301869052925160019360a0808501949193601f19840193928390039091019190865af1158015610c35573d6000803e3d6000fd5b5050506020604051035193505b505050929150505600a165627a7a723058200525f2f74689b5b1e1ae7f42f90da9c80254dabf84e8ba0927838745ce41070d0029`

func NewWallet() *MyWallet {
	return &MyWallet{
		EthConfig: ETHConfig{
			NetID:  1,
			Rawurl: "https://ropsten.infura.io/",
		},
		NameKey:     map[string][]byte{},
		NamePubkey:  map[string][]byte{},
		NameAddress: map[string]string{},
		AddressKey:  map[string][]byte{},
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
	eth := ethadaptor.NewAdaptorETHTestnet()
	//
	key, _ := eth.NewPrivateKey(&adaptor.NewPrivateKeyInput{})
	gWallet.NameKey[name] = key.PrivateKey

	//
	pubkey, _ := eth.GetPublicKey(&adaptor.GetPublicKeyInput{PrivateKey: key.PrivateKey})
	gWallet.NamePubkey[name] = pubkey.PublicKey

	//
	address, _ := eth.GetAddress(&adaptor.GetAddressInput{Key: pubkey.PublicKey})
	gWallet.NameAddress[name] = address.Address
	gWallet.AddressKey[address.Address] = key.PrivateKey

	return saveConfig(gWalletFile, gWallet)
}

func sendETH(rawURL, priKey, toAddr, amountWei string) (string, error) {
	client, err := ethclient.Dial(gWallet.EthConfig.Rawurl)
	if err != nil {
		return "", err
	}

	privateKey, err := crypto.HexToECDSA(priKey)
	if err != nil {
		return "", err
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return "", fmt.Errorf("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		return "", err
	}
	//nonce = 509

	//value := big.NewInt(1000000000000000000) // in wei (1 eth)
	value := new(big.Int)
	value.SetString(amountWei, 10)
	gasLimit := uint64(21000) // in units
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return "", err
	}
	//fmt.Println(gasPrice.String())
	//gasPrice.SetString("1", 10)
	//fmt.Println(gasPrice.String())

	toAddress := common.HexToAddress(toAddr)
	var data []byte
	tx := types.NewTransaction(nonce, toAddress, value, gasLimit, gasPrice, data)

	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		return "", err
	}

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		return "", err
	}

	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return "", err
	}

	fmt.Printf("tx sent: %s\n", signedTx.Hash().Hex())
	return signedTx.Hash().String(), nil
}

func sendETHToMultiSigAddr(contractAddr, value, gasPrice, gasLimit, privateKey string) error {
	txid, err := sendETH(gWallet.EthConfig.Rawurl, privateKey, contractAddr, value)
	if txid != "" {
		fmt.Println("sendETHToMultiSigAddr: ", txid)
	}
	return err
}

func getAddrByPrikey(prikey []byte) string {
	eth := ethadaptor.NewAdaptorETHTestnet()
	pubkey, _ := eth.GetPublicKey(&adaptor.GetPublicKeyInput{PrivateKey: prikey})
	address, _ := eth.GetAddress(&adaptor.GetAddressInput{Key: pubkey.PublicKey})
	return address.Address
}

func spendEtHFromMultiAddr(contractAddr, gasPrice, gasLimit, ethRecvddr, amount, reqid, sig1, sig2, sig3, prikeyHex string) error {

	eth := ethadaptor.NewAdaptorETHTestnet()

	if "0x" == prikeyHex[0:2] || "0X" == prikeyHex[0:2] {
		prikeyHex = prikeyHex[2:]
	}
	prikey, err := hex.DecodeString(prikeyHex)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	callerAddr := getAddrByPrikey(prikey)

	//
	var invokeContractParams adaptor.CreateContractInvokeTxInput
	invokeContractParams.Address = callerAddr
	invokeContractParams.ContractAddress = contractAddr
	amt := new(big.Int)
	amt.SetString("21000000000000000", 10) //10000000000 10gwei*2100000
	invokeContractParams.Fee = adaptor.NewAmountAsset(amt, "ETH")
	invokeContractParams.Function = "withdraw"
	invokeContractParams.Extra = []byte(EthmultisigABI)
	invokeContractParams.Args = append(invokeContractParams.Args, []byte(ethRecvddr))
	invokeContractParams.Args = append(invokeContractParams.Args, []byte(amount))
	invokeContractParams.Args = append(invokeContractParams.Args, []byte(reqid))
	invokeContractParams.Args = append(invokeContractParams.Args, []byte(sig1))
	invokeContractParams.Args = append(invokeContractParams.Args, []byte(sig2))
	invokeContractParams.Args = append(invokeContractParams.Args, []byte(sig3))

	//1.gen tx
	resultTx, err := eth.CreateContractInvokeTx(&invokeContractParams)
	if err != nil {
		fmt.Println(err.Error())
		return err
	} else {
		fmt.Println(resultTx)
	}

	//2.sign tx
	var input adaptor.SignTransactionInput
	input.PrivateKey = prikey
	//input.Transaction = Hex2Bytes("f9024981848203e883200b20946817cfb2c442693d850332c3b755b2342ec4afb280b902248c2e032100000000000000000000000000000000000000000000000000000000000000c0000000000000000000000000aaa919a7c465be9b053673c567d73be8603179630000000000000000000000000000000000000000000000000de0b6b3a76400000000000000000000000000000000000000000000000000000000000000000001000000000000000000000000000000000000000000000000000000000000012000000000000000000000000000000000000000000000000000000000000001a000000000000000000000000000000000000000000000000000000000000000407d7116a8706ae08baa7f4909e26728fa7a5f0365aaa919a7c465be9b053673c567d73be8603179636c7110482920e0af149a82189251f292a84148a85b7cd70d00000000000000000000000000000000000000000000000000000000000000417197961c5ae032ed6f33650f1f3a3ba111e8548a3dad14b3afa1cb6bc8f4601a6cb2b21aedcd575784e923942f3130f3290d56522ab2b28afca478e489426a4601000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000041ae94b0e599ef0508ba7bec41db5b46d5a065b30d3d5c4b0a4c85ea2d4899d6607e80e3314ee0741049963d30fb3aceaa5506e13835a41ef54a8f44a04ef0f1e40100000000000000000000000000000000000000000000000000000000000000808080")
	input.Transaction = resultTx.RawTransaction
	resultSign, err := eth.SignTransaction(&input)
	if err != nil {
		fmt.Println(err.Error())
		return err
	} else {
		fmt.Println(resultSign)
	}

	//3.send tx
	var sendTransactionParams adaptor.SendTransactionInput
	input.Transaction = resultSign.Extra
	resultSend, err := eth.SendTransaction(&sendTransactionParams)
	if err != nil {
		fmt.Println(err.Error())
		return err
	} else {
		fmt.Println(resultSend)
	}

	return nil
}

func setJuryAddrs(contractAddr, gasPrice, gasLimit, addr1, addr2, addr3, addr4, prikeyHex string) error {
	eth := ethadaptor.NewAdaptorETHTestnet()

	if "0x" == prikeyHex[0:2] || "0X" == prikeyHex[0:2] {
		prikeyHex = prikeyHex[2:]
	}
	prikey, err := hex.DecodeString(prikeyHex)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	callerAddr := getAddrByPrikey(prikey)

	//
	var invokeContractParams adaptor.CreateContractInvokeTxInput
	invokeContractParams.Address = callerAddr
	invokeContractParams.ContractAddress = contractAddr
	amt := new(big.Int)
	amt.SetString("21000000000000000", 10) //10000000000 10gwei*2100000
	invokeContractParams.Fee = adaptor.NewAmountAsset(amt, "ETH")
	invokeContractParams.Function = "withdraw"
	invokeContractParams.Extra = []byte(EthmultisigABI)
	invokeContractParams.Args = append(invokeContractParams.Args, []byte(addr1))
	invokeContractParams.Args = append(invokeContractParams.Args, []byte(addr2))
	invokeContractParams.Args = append(invokeContractParams.Args, []byte(addr3))
	invokeContractParams.Args = append(invokeContractParams.Args, []byte(addr4))

	//1.gen tx
	resultTx, err := eth.CreateContractInvokeTx(&invokeContractParams)
	if err != nil {
		fmt.Println(err.Error())
		return err
	} else {
		fmt.Println(resultTx)
	}

	//2.sign tx
	var input adaptor.SignTransactionInput
	input.PrivateKey = prikey
	//input.Transaction = Hex2Bytes("f9024981848203e883200b20946817cfb2c442693d850332c3b755b2342ec4afb280b902248c2e032100000000000000000000000000000000000000000000000000000000000000c0000000000000000000000000aaa919a7c465be9b053673c567d73be8603179630000000000000000000000000000000000000000000000000de0b6b3a76400000000000000000000000000000000000000000000000000000000000000000001000000000000000000000000000000000000000000000000000000000000012000000000000000000000000000000000000000000000000000000000000001a000000000000000000000000000000000000000000000000000000000000000407d7116a8706ae08baa7f4909e26728fa7a5f0365aaa919a7c465be9b053673c567d73be8603179636c7110482920e0af149a82189251f292a84148a85b7cd70d00000000000000000000000000000000000000000000000000000000000000417197961c5ae032ed6f33650f1f3a3ba111e8548a3dad14b3afa1cb6bc8f4601a6cb2b21aedcd575784e923942f3130f3290d56522ab2b28afca478e489426a4601000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000041ae94b0e599ef0508ba7bec41db5b46d5a065b30d3d5c4b0a4c85ea2d4899d6607e80e3314ee0741049963d30fb3aceaa5506e13835a41ef54a8f44a04ef0f1e40100000000000000000000000000000000000000000000000000000000000000808080")
	input.Transaction = resultTx.RawTransaction
	resultSign, err := eth.SignTransaction(&input)
	if err != nil {
		fmt.Println(err.Error())
		return err
	} else {
		fmt.Println(resultSign)
	}

	//3.send tx
	var sendTransactionParams adaptor.SendTransactionInput
	input.Transaction = resultSign.Extra
	resultSend, err := eth.SendTransaction(&sendTransactionParams)
	if err != nil {
		fmt.Println(err.Error())
		return err
	} else {
		fmt.Println(resultSend)
	}

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
	fmt.Println("Params : deposit, contractAddr, value, gasPrice, gasLimit, ethPrivateKey")
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
		if len(args) < 7 {
			fmt.Println("Params : deposit, contractAddr, value, gasPrice, gasLimit, ethPrivateKey")
			return
		}
		err := sendETHToMultiSigAddr(args[2], args[3], args[4], args[5], args[6])
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
