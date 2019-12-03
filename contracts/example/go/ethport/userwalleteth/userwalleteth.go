package main

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"math/big"
	"os"
	"strconv"
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

const EthmultisigABI = "[{\"constant\":true,\"inputs\":[],\"name\":\"name\",\"outputs\":[{\"name\":\"\",\"type\":\"string\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"spender\",\"type\":\"address\"},{\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"approve\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"totalSupply\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"from\",\"type\":\"address\"},{\"name\":\"to\",\"type\":\"address\"},{\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"transferFrom\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_addr\",\"type\":\"address\"},{\"name\":\"_ptnhex\",\"type\":\"address\"}],\"name\":\"resetMapAddr\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"decimals\",\"outputs\":[{\"name\":\"\",\"type\":\"uint8\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"getMapPtnAddr\",\"outputs\":[{\"name\":\"\",\"type\":\"string\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"address\"}],\"name\":\"addrmapPTN\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"ptnAddr\",\"type\":\"address\"}],\"name\":\"getMapEthAddr\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_owner\",\"type\":\"address\"}],\"name\":\"balanceOf\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"recver\",\"type\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\"},{\"name\":\"reqid\",\"type\":\"bytes32\"},{\"name\":\"sigstr1\",\"type\":\"bytes\"},{\"name\":\"sigstr2\",\"type\":\"bytes\"},{\"name\":\"sigstr3\",\"type\":\"bytes\"}],\"name\":\"withdraw\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"address\"}],\"name\":\"addrmap\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"suicideto\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"addrHex\",\"type\":\"address\"}],\"name\":\"encodeBase58\",\"outputs\":[{\"name\":\"\",\"type\":\"string\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"symbol\",\"outputs\":[{\"name\":\"\",\"type\":\"string\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_ptnhex\",\"type\":\"address\"},{\"name\":\"_amt\",\"type\":\"uint256\"}],\"name\":\"transfer\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"my_eth_bal\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"owner\",\"type\":\"address\"},{\"name\":\"spender\",\"type\":\"address\"}],\"name\":\"allowance\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"addra\",\"type\":\"address\"},{\"name\":\"addrb\",\"type\":\"address\"},{\"name\":\"addrc\",\"type\":\"address\"},{\"name\":\"addrd\",\"type\":\"address\"}],\"name\":\"setaddrs\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"reqid\",\"type\":\"bytes32\"}],\"name\":\"getmultisig\",\"outputs\":[{\"name\":\"\",\"type\":\"uint8\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"name\":\"addra\",\"type\":\"address\"},{\"name\":\"addrb\",\"type\":\"address\"},{\"name\":\"addrc\",\"type\":\"address\"},{\"name\":\"addrd\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"fallback\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"token\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"user\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"Deposit\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"token\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"user\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"recver\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"amount\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"reqid\",\"type\":\"bytes32\"}],\"name\":\"Withdraw\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"to\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Transfer\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"spender\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Approval\",\"type\":\"event\"}]"

const EthmultisigBin = `0x6080604052633b9aca0060025534801561001857600080fd5b506040516080806116bb83398101604090815281516020830151918301516060909301516003805433600160a060020a031991821617909155600480548216600160a060020a039485161790556005805482169484169490941790935560068054841694831694909417909355600780549092169216919091179055611618806100a36000396000f3006080604052600436106101115763ffffffff7c010000000000000000000000000000000000000000000000000000000060003504166306fdde038114610152578063095ea7b3146101dc57806318160ddd1461021457806323b872dd1461023b5780632da2ff3414610265578063313ce5671461028e57806348cedf90146102b95780634e11092f146102da5780636e932a1c1461031757806370a082311461033857806373432d0a146103595780638c5cecaa146104415780638e644ec314610462578063927f526f1461048357806395d89b41146104a4578063a9059cbb146104b9578063c8fc638a146104dd578063dd62ed3e146104f2578063df5aa58d14610519578063e76480fc1461054c575b6040805160008152336020820152348183015290517f5548c837ab068cf56a2c2479df0882a4922fd203edb7517321831d95078c5f629181900360600190a1005b34801561015e57600080fd5b50610167610564565b6040805160208082528351818301528351919283929083019185019080838360005b838110156101a1578181015183820152602001610189565b50505050905090810190601f1680156101ce5780820380516001836020036101000a031916815260200191505b509250505060405180910390f35b3480156101e857600080fd5b50610200600160a060020a036004351660243561059b565b604080519115158252519081900360200190f35b34801561022057600080fd5b506102296105a4565b60408051918252519081900360200190f35b34801561024757600080fd5b50610200600160a060020a03600435811690602435166044356105aa565b34801561027157600080fd5b5061028c600160a060020a03600435811690602435166105b3565b005b34801561029a57600080fd5b506102a3610670565b6040805160ff9092168252519081900360200190f35b3480156102c557600080fd5b50610167600160a060020a0360043516610675565b3480156102e657600080fd5b506102fb600160a060020a03600435166106da565b60408051600160a060020a039092168252519081900360200190f35b34801561032357600080fd5b506102fb600160a060020a03600435166106f5565b34801561034457600080fd5b50610229600160a060020a0360043516610713565b34801561036557600080fd5b50604080516020601f60643560048181013592830184900484028501840190955281845261028c94600160a060020a03813516946024803595604435953695608494930191819084018382808284375050604080516020601f89358b018035918201839004830284018301909452808352979a99988101979196509182019450925082915084018382808284375050604080516020601f89358b018035918201839004830284018301909452808352979a9998810197919650918201945092508291508401838280828437509497506107679650505050505050565b34801561044d57600080fd5b506102fb600160a060020a03600435166109d1565b34801561046e57600080fd5b5061028c600160a060020a03600435166109ec565b34801561048f57600080fd5b50610167600160a060020a0360043516610a0f565b3480156104b057600080fd5b50610167610a34565b3480156104c557600080fd5b50610200600160a060020a0360043516602435610a6b565b3480156104e957600080fd5b50610229610b3a565b3480156104fe57600080fd5b50610229600160a060020a036004358116906024351661059b565b34801561052557600080fd5b5061028c600160a060020a0360043581169060243581169060443581169060643516610b3f565b34801561055857600080fd5b506102a3600435610bb3565b60408051808201909152600881527f45544820506f7274000000000000000000000000000000000000000000000000602082015281565b60005b92915050565b60025490565b60009392505050565b600354600160a060020a031633146105ca57600080fd5b600160a060020a0382811660009081526020819052604090205481169082161480156106125750600160a060020a038181166000908152600160205260409020548116908316145b1561066757600160a060020a03808316600090815260208181526040808320805473ffffffffffffffffffffffffffffffffffffffff199081169091559385168352600190915290208054909116905561066c565b600080fd5b5050565b600081565b600160a060020a038181166000908152602081905260409020546060911615156106ae57506040805160208101909152600081526106d5565b600160a060020a038083166000908152602081905260409020546106d29116610a0f565b90505b919050565b600160205260009081526040902054600160a060020a031681565b600160a060020a039081166000908152600160205260409020541690565b600160a060020a03818116600090815260208190526040812054909116151561075f57600160a060020a0382811660009081526001602052604090205416151561075f575060016106d5565b5060006106d5565b60008481526008602052604081205460609190819060ff161561078957600080fd5b60408051600480825260a08201909252906020820160808038833950506004548251929550600160a060020a03169185915060009081106107c657fe5b600160a060020a03928316602091820290920101526005548451911690849060019081106107f057fe5b600160a060020a039283166020918202909201015260065484519116908490600290811061081a57fe5b600160a060020a039283166020918202909201015260075484519116908490600390811061084457fe5b600160a060020a039283166020918202909201810191909152604080516c0100000000000000000000000030810282850152938d169093026034840152604883018b905260688084018b90528151808503909101815260889093019081905282516000955090918291908401908083835b602083106108d45780518252601f1990920191602091820191016108b5565b6001836020036101000a038019825116818451168082178552505050505050905001915050604051809103902090506109108382888888610bc8565b9150600360ff8316101561092357600080fd5b600087815260086020526040808220805460ff1916600117905551600160a060020a038b16918a156108fc02918b91818181858888f1935050505015801561096f573d6000803e3d6000fd5b506040805160008152336020820152600160a060020a038b1681830152606081018a90526080810189905290517f2e0d455354ec8eaf3f01b1e7570bcfc70a2bc7126f7f641541df1e6efbc8f4109181900360a00190a1505050505050505050565b600060208190529081526040902054600160a060020a031681565b600354600160a060020a03163314610a0357600080fd5b80600160a060020a0316ff5b60608181610a2c610a27610a2284610c16565b610dee565b610f87565b949350505050565b60408051808201909152600781527f455448506f727400000000000000000000000000000000000000000000000000602082015281565b33600090815260208190526040812054600160a060020a0316158015610aa95750600160a060020a0383811660009081526001602052604090205416155b1561066757336000818152602081815260408083208054600160a060020a03891673ffffffffffffffffffffffffffffffffffffffff199182168117909255818552600184529382902080549094168517909355805186815290519293927fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef929181900390910190a350600161059e565b303190565b600354600160a060020a03163314610b5657600080fd5b60048054600160a060020a0395861673ffffffffffffffffffffffffffffffffffffffff19918216179091556005805494861694821694909417909355600680549285169284169290921790915560078054919093169116179055565b60009081526008602052604090205460ff1690565b60408051600480825260a08201909252600091606091839160208201608080388339019050509150610bfe82888a8989896110d4565b506000610c0a8261113a565b98975050505050505050565b6040805160198082528183019092526060916c010000000000000000000000008402918391600091829182919060208201610320803883395050815191955060009186915082908110610c6557fe5b906020010190600160f860020a031916908160001a905350600092505b60148360ff161015610cdb578460ff841660148110610c9d57fe5b1a60f860020a02848460010160ff16815181101515610cb857fe5b906020010190600160f860020a031916908160001a905350600190920191610c82565b6040805160008082526bffffffffffffffffffffffff1988166001830152915160029283926015808201936020939092839003909101908290865af1158015610d28573d6000803e3d6000fd5b5050506040513d6020811015610d3d57600080fd5b505160408051918252516020828101929091908190038201816000865af1158015610d6c573d6000803e3d6000fd5b5050506040513d6020811015610d8157600080fd5b50519150600090505b60048160ff161015610de0578160ff821660208110610da557fe5b1a60f860020a02848260150160ff16815181101515610dc057fe5b906020010190600160f860020a031916908160001a905350600101610d8a565b8395505b5050505050919050565b606080600080600080865160001415610e17576040805160208101909152600081529550610de4565b60408051602880825261052082019092529060208201610500803883390190505094506000856000815181101515610e4b57fe5b60ff90921660209283029091019091015260019350600092505b86518360ff161015610f6257868360ff16815181101515610e8257fe5b90602001015160f860020a900460f860020a0260f860020a900460ff169150600090505b8360ff168160ff161015610f1757848160ff16815181101515610ec557fe5b9060200190602002015160ff166101000282019150603a82811515610ee657fe5b06858260ff16815181101515610ef857fe5b60ff909216602092830290910190910152603a82049150600101610ea6565b6000821115610f5757603a8206858560ff16815181101515610f3557fe5b60ff909216602092830290910190910152600190930192603a82049150610f17565b826001019250610e65565b610f7c610f77610f72878761124e565b6112db565b611371565b979650505050505050565b606080606060008085516002016040519080825280601f01601f191660200182016040528015610fc1578160200160208202803883390190505b508051909450849350600192507f50000000000000000000000000000000000000000000000000000000000000009084906000908110610ffd57fe5b906020010190600160f860020a031916908160001a905350825160018301927f310000000000000000000000000000000000000000000000000000000000000091859190811061104957fe5b906020010190600160f860020a031916908160001a905350600090505b85518160ff1610156110ca57858160ff1681518110151561108357fe5b90602001015160f860020a900460f860020a0283838060010194508151811015156110aa57fe5b906020010190600160f860020a031916908160001a905350600101611066565b5091949350505050565b8251600090156110f5576110e88685611484565b90506110f5878287611497565b825115611113576111068684611484565b9050611113878287611497565b815115611131576111248683611484565b9050611131878287611497565b50505050505050565b60408051600480825260a08201909252600091606091839182919060208201608080388339019050509250600183600081518110151561117657fe5b60ff909216602092830290910190910152825160019084908290811061119857fe5b60ff9092166020928302909101909101528251600190849060029081106111bb57fe5b60ff9092166020928302909101909101528251600190849060039081106111de57fe5b60ff9092166020928302909101909101525060009050805b60048160ff16101561124657828160ff1681518110151561121357fe5b90602001906020020151858260ff1681518110151561122e57fe5b602090810290910101510291909101906001016111f6565b509392505050565b60608060008360ff16604051908082528060200260200182016040528015611280578160200160208202803883390190505b509150600090505b8360ff168160ff16101561124657848160ff168151811015156112a757fe5b90602001906020020151828260ff168151811015156112c257fe5b60ff909216602092830290910190910152600101611288565b6060806000835160405190808252806020026020018201604052801561130b578160200160208202803883390190505b509150600090505b83518160ff16101561136a578351849060ff831681036000190190811061133657fe5b90602001906020020151828260ff1681518110151561135157fe5b60ff909216602092830290910190910152600101611313565b5092915050565b606080600083516040519080825280601f01601f1916602001820160405280156113a5578160200160208202803883390190505b509150600090505b83518160ff16101561136a57606060405190810160405280603a81526020017f31323334353637383941424344454647484a4b4c4d4e5051525354555657585981526020017f5a6162636465666768696a6b6d6e6f707172737475767778797a000000000000815250848260ff1681518110151561142757fe5b9060200190602002015160ff1681518110151561144057fe5b90602001015160f860020a900460f860020a02828260ff1681518110151561146457fe5b906020010190600160f860020a031916908160001a9053506001016113ad565b60006114908383611517565b9392505050565b60005b60048160ff16101561151157818160ff168151811015156114b757fe5b90602001906020020151600160a060020a031683600160a060020a03161415156114e057611509565b6001848260ff168151811015156114f357fe5b60ff909216602092830290910190910152611511565b60010161149a565b50505050565b6000806000808451604114151561153157600093506115e3565b50505060208201516040830151606084015160001a601b60ff8216101561155657601b015b8060ff16601b1415801561156e57508060ff16601c14155b1561157c57600093506115e3565b60408051600080825260208083018085528a905260ff8516838501526060830187905260808301869052925160019360a0808501949193601f19840193928390039091019190865af11580156115d6573d6000803e3d6000fd5b5050506020604051035193505b505050929150505600a165627a7a723058204ffd1c311b43165d4af93e4f77779e3406671fe0e831a08ee40c576f8a867d800029`

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

//func sendETH(rawURL, priKey, toAddr, amountWei string) (string, error) {
//	client, err := ethclient.Dial(gWallet.EthConfig.Rawurl)
//	if err != nil {
//		return "", err
//	}
//
//	privateKey, err := crypto.HexToECDSA(priKey)
//	if err != nil {
//		return "", err
//	}
//
//	publicKey := privateKey.Public()
//	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
//	if !ok {
//		return "", fmt.Errorf("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
//	}
//
//	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
//	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
//	if err != nil {
//		return "", err
//	}
//	//nonce = 509
//
//	//value := big.NewInt(1000000000000000000) // in wei (1 eth)
//	value := new(big.Int)
//	value.SetString(amountWei, 10)
//	gasLimit := uint64(21000) // in units
//	gasPrice, err := client.SuggestGasPrice(context.Background())
//	if err != nil {
//		return "", err
//	}
//	//fmt.Println(gasPrice.String())
//	//gasPrice.SetString("1", 10)
//	//fmt.Println(gasPrice.String())
//
//	toAddress := common.HexToAddress(toAddr)
//	var data []byte
//	tx := types.NewTransaction(nonce, toAddress, value, gasLimit, gasPrice, data)
//
//	chainID, err := client.NetworkID(context.Background())
//	if err != nil {
//		return "", err
//	}
//
//	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
//	if err != nil {
//		return "", err
//	}
//
//	err = client.SendTransaction(context.Background(), signedTx)
//	if err != nil {
//		return "", err
//	}
//
//	fmt.Printf("tx sent: %s\n", signedTx.Hash().Hex())
//	return signedTx.Hash().String(), nil
//}

func sendETHToMultiSigAddr(contractAddr, value, gasPrice, gasLimit, prikeyHex string) error {

	if "0x" == prikeyHex[0:2] || "0X" == prikeyHex[0:2] {
		prikeyHex = prikeyHex[2:]
	}
	prikey, err := hex.DecodeString(prikeyHex)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	eth := ethadaptor.NewAdaptorETHTestnet()

	callerAddr := getAddrByPrikey(prikey)

	var txInput adaptor.CreateTransferTokenTxInput
	txInput.FromAddress = callerAddr
	txInput.ToAddress = contractAddr
	amt := new(big.Int)
	amt.SetString(value, 10) //10000000000 10gwei*2100000
	txInput.Amount = adaptor.NewAmountAsset(amt, "ETH")
	amtFee := new(big.Int)
	amtFee.SetString("21000000000000000", 10) //10000000000 10gwei*2100000
	txInput.Fee = adaptor.NewAmountAsset(amtFee, "ETH")

	resultTx, err := eth.CreateTransferTokenTx(&txInput)
	if err != nil {
		fmt.Println(err.Error())
		return err
	} else {
		fmt.Println(resultTx)
	}

	return signAndSend(eth, resultTx.Transaction, prikey)
}

func getAddrByPrikey(prikey []byte) string {
	eth := ethadaptor.NewAdaptorETHTestnet()
	pubkey, _ := eth.GetPublicKey(&adaptor.GetPublicKeyInput{PrivateKey: prikey})
	address, _ := eth.GetAddress(&adaptor.GetAddressInput{Key: pubkey.PublicKey})
	return address.Address
}

func spendEtHFromMultiAddr(contractAddr, gasPrice, gasLimit, reqid, ethRecvddr, amount, fee, sig1, sig2, sig3, prikeyHex string) error {

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

	amountU, _ := strconv.ParseUint(amount, 10, 64)
	feeU, _ := strconv.ParseUint(fee, 10, 64)
	amountBigInt := new(big.Int)
	amountBigInt.SetUint64(amountU - feeU)
	amountBigInt.Mul(amountBigInt, big.NewInt(1e10)) //eth's decimal is 18, ethToken in PTN is decimal is 8

	invokeContractParams.Args = append(invokeContractParams.Args, []byte(amountBigInt.String()))
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
		fmt.Printf("RawTransaction: %x\n", resultTx.RawTransaction)
	}

	return signAndSend(eth, resultTx.RawTransaction, prikey)
}

func signAndSend(eth *ethadaptor.AdaptorETH, rawTransaction, prikey []byte) error {
	//2.sign tx
	var input adaptor.SignTransactionInput
	input.PrivateKey = prikey
	//input.Transaction = Hex2Bytes("f9024981848203e883200b20946817cfb2c442693d850332c3b755b2342ec4afb280b902248c2e032100000000000000000000000000000000000000000000000000000000000000c0000000000000000000000000aaa919a7c465be9b053673c567d73be8603179630000000000000000000000000000000000000000000000000de0b6b3a76400000000000000000000000000000000000000000000000000000000000000000001000000000000000000000000000000000000000000000000000000000000012000000000000000000000000000000000000000000000000000000000000001a000000000000000000000000000000000000000000000000000000000000000407d7116a8706ae08baa7f4909e26728fa7a5f0365aaa919a7c465be9b053673c567d73be8603179636c7110482920e0af149a82189251f292a84148a85b7cd70d00000000000000000000000000000000000000000000000000000000000000417197961c5ae032ed6f33650f1f3a3ba111e8548a3dad14b3afa1cb6bc8f4601a6cb2b21aedcd575784e923942f3130f3290d56522ab2b28afca478e489426a4601000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000041ae94b0e599ef0508ba7bec41db5b46d5a065b30d3d5c4b0a4c85ea2d4899d6607e80e3314ee0741049963d30fb3aceaa5506e13835a41ef54a8f44a04ef0f1e40100000000000000000000000000000000000000000000000000000000000000808080")
	input.Transaction = rawTransaction
	resultSign, err := eth.SignTransaction(&input)
	if err != nil {
		fmt.Println("SignTransaction failed : ", err.Error())
		return err
	} else {
		fmt.Printf("tx: %x\n", resultSign.Extra)
	}

	//3.send tx
	var sendTransactionParams adaptor.SendTransactionInput
	sendTransactionParams.Transaction = resultSign.Extra
	resultSend, err := eth.SendTransaction(&sendTransactionParams)
	if err != nil {
		fmt.Println("SendTransaction failed : ", err.Error())
		return err
	} else {
		fmt.Printf("%x\n", resultSend.TxID)
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
	invokeContractParams.Function = "setaddrs"
	invokeContractParams.Extra = []byte(EthmultisigABI)
	invokeContractParams.Args = append(invokeContractParams.Args, []byte(addr1))
	invokeContractParams.Args = append(invokeContractParams.Args, []byte(addr2))
	invokeContractParams.Args = append(invokeContractParams.Args, []byte(addr3))
	invokeContractParams.Args = append(invokeContractParams.Args, []byte(addr4))

	//1.gen tx
	resultTx, err := eth.CreateContractInvokeTx(&invokeContractParams)
	if err != nil {
		fmt.Println("CreateContractInvokeTx failed : ", err.Error())
		return err
	} else {
		fmt.Printf("RawTransaction: %x\n", resultTx.RawTransaction)
	}

	return signAndSend(eth, resultTx.RawTransaction, prikey)
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
	fmt.Println("Params : withdraw, contractAddr, gasPrice, gasLimit, reqid, ethAddr, amount, fee, sig1, sig2, sig3, ethPrivateKey")
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
		if len(args) < 13 {
			fmt.Println("Params : withdraw, contractAddr, gasPrice, gasLimit, reqid, ethAddr, amount, fee, sig1, sig2, sig3, ethPrivateKey")
			return
		}
		err := spendEtHFromMultiAddr(args[2], args[3], args[4], args[5], args[6], args[7], args[8], args[9], args[10], args[11], args[12])
		if err != nil {
			fmt.Println(err.Error())
		}
	default:
		fmt.Println("Invalid cmd.")
		helper()
	}
}
