package main

import (
	"bufio"
	//"encoding/json"
	"errors"
	"fmt"
	"os"
	//"strconv"
	"strings"
	"time"

	"github.com/naoina/toml"
	//"github.com/palletone/adaptor"
	//"github.com/palletone/btc-adaptor"
)

type BTCConfig struct {
	NetID        int
	Host         string
	RPCUser      string
	RPCPasswd    string
	CertPath     string
	WalletPasswd string
}
type MyWallet struct {
	BtcConfig   BTCConfig
	NameKey     map[string]string
	NamePubkey  map[string]string
	NameAddress map[string]string
	AddressKey  map[string]string
}

var gWallet = NewWallet()
var gWalletFile = "./btcwallet.toml"

var gTomlConfig = toml.DefaultConfig

func NewWallet() *MyWallet {
	return &MyWallet{
		BtcConfig: BTCConfig{
			NetID:        1,
			Host:         "localhost:18332",
			RPCUser:      "zxl",
			RPCPasswd:    "123456",
			CertPath:     "",
			WalletPasswd: "1",
		},
		NameKey:     map[string]string{},
		NamePubkey:  map[string]string{},
		NameAddress: map[string]string{},
		AddressKey:  map[string]string{}}
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
	//var btcadaptor adaptorbtc.AdaptorBTC
	//btcadaptor.NetID = gWallet.BtcConfig.NetID
	////
	//key := btcadaptor.NewPrivateKey()
	//gWallet.NameKey[name] = key
	//
	////
	//pubkey := btcadaptor.GetPublicKey(key)
	//gWallet.NamePubkey[name] = pubkey
	//fmt.Println(name, "'s pubkey : ", pubkey)
	//
	////
	//addr := btcadaptor.GetAddress(key)
	//gWallet.NameAddress[name] = addr
	//gWallet.AddressKey[addr] = key

	return saveConfig(gWalletFile, gWallet)
}

func giveAlice(txid string, index string, amount string, fee string, prikey string) error {
	//
	//vout, err := strconv.Atoi(index)
	//if err != nil {
	//	return errors.New("Index is Invalid.")
	//}
	//
	//amountValue, err := strconv.ParseFloat(amount, 64)
	//if err != nil {
	//	return errors.New("Amount is Invalid.")
	//}
	//feeValue, err := strconv.ParseFloat(fee, 64)
	//if err != nil {
	//	return errors.New("Fee is Invalid.")
	//}
	////
	//var rawTransactionGenParams adaptor.RawTransactionGenParams
	//rawTransactionGenParams.Inputs = append(rawTransactionGenParams.Inputs, adaptor.Input{Txid: txid, Vout: uint32(vout)})
	//rawTransactionGenParams.Outputs = append(rawTransactionGenParams.Outputs, adaptor.Output{Address: gWallet.NameAddress["alice"],
	//	Amount: amountValue - feeValue})
	//
	//var btcadaptor adaptorbtc.AdaptorBTC
	//btcadaptor.NetID = gWallet.BtcConfig.NetID
	//btcadaptor.Host = gWallet.BtcConfig.Host
	//btcadaptor.RPCUser = gWallet.BtcConfig.RPCUser
	//btcadaptor.RPCPasswd = gWallet.BtcConfig.RPCPasswd
	//btcadaptor.CertPath = gWallet.BtcConfig.CertPath
	////
	//rawResult, err := btcadaptor.RawTransactionGen(&rawTransactionGenParams)
	//if err != nil {
	//	fmt.Println(err.Error())
	//	return err
	//} else {
	//	fmt.Println(rawResult)
	//}
	////
	//var rawTransactionGenResult adaptor.RawTransactionGenResult
	//err = json.Unmarshal([]byte(rawResult), &rawTransactionGenResult)
	//if err != nil {
	//	fmt.Println(err.Error())
	//}
	////
	//var signTxSendParams adaptor.SignTxSendParams
	//signTxSendParams.TransactionHex = rawTransactionGenResult.Rawtx
	//signTxSendParams.Privkeys = append(signTxSendParams.Privkeys, prikey)
	//sendReuslt, err := btcadaptor.SignTxSend(&signTxSendParams)
	//if err != nil {
	//	fmt.Println(err.Error())
	//	return err
	//} else {
	//	fmt.Println(sendReuslt)
	//}

	return nil
}

func getBalance(addr string) error {
	////
	//var btcadaptor adaptorbtc.AdaptorBTC
	//btcadaptor.NetID = gWallet.BtcConfig.NetID
	//
	//getBalanceParams := &adaptor.GetBalanceHttpParams{Address: addr, Minconf: 6}
	//result, err := btcadaptor.GetBalanceHttp(getBalanceParams)
	//if err != nil {
	//	return err
	//}
	//
	//fmt.Println(result)
	return nil
}
func checkTxAmount(txid string, index int, txAmount float64) error {
	//var btcadaptor adaptorbtc.AdaptorBTC
	//btcadaptor.NetID = gWallet.BtcConfig.NetID
	//prams := adaptor.GetTransactionHttpParams{TxHash: txid}
	//result, err := btcadaptor.GetTransactionHttp(&prams)
	//if err != nil {
	//	return err
	//}
	//fmt.Println("==== ==== The input tx start ==== ====")
	//fmt.Println(result)
	//fmt.Println("==== ==== The input tx end ==== ====")
	//
	//var txResult adaptor.GetTransactionHttpResult
	//err = json.Unmarshal([]byte(result), &txResult)
	//if err != nil {
	//	return err
	//}
	//
	//if index < len(txResult.Outputs) {
	//	if txResult.Outputs[index].Value != txAmount {
	//		return errors.New("The txAmount is invalid")
	//	}
	//	return nil
	//}

	return errors.New("The index is invalid")
}

func aliceSendBTCToMultiSigAddr(txid string, index string, txAmount string, fee string, multiSigAddr string, prikey string) error {
	////
	//vout, err := strconv.Atoi(index)
	//if err != nil {
	//	return errors.New("Index is Invalid.")
	//}
	//
	//amountValue, err := strconv.ParseFloat(txAmount, 64)
	//if err != nil {
	//	return errors.New("Amount is Invalid.")
	//}
	//feeValue, err := strconv.ParseFloat(fee, 64)
	//if err != nil {
	//	return errors.New("Fee is Invalid.")
	//}
	//
	//err = checkTxAmount(txid, vout, amountValue)
	//if err != nil {
	//	return err
	//}
	//
	////
	//var rawTransactionGenParams adaptor.RawTransactionGenParams
	//rawTransactionGenParams.Inputs = append(rawTransactionGenParams.Inputs, adaptor.Input{Txid: txid, Vout: uint32(vout)})
	//rawTransactionGenParams.Outputs = append(rawTransactionGenParams.Outputs, adaptor.Output{Address: multiSigAddr, Amount: amountValue - feeValue})
	////
	//var btcadaptor adaptorbtc.AdaptorBTC
	//btcadaptor.NetID = gWallet.BtcConfig.NetID
	//aliceAddr := btcadaptor.GetAddress(gWallet.NameKey["alice"])
	////
	//rawResult, err := btcadaptor.RawTransactionGen(&rawTransactionGenParams)
	//if err != nil {
	//	fmt.Println(err.Error())
	//	return err
	//} else {
	//	fmt.Println("==== ==== Raw tx start ==== ====")
	//	fmt.Println(rawResult)
	//	fmt.Println("==== ==== Raw tx end ==== ====")
	//}
	////
	//var rawTransactionGenResult adaptor.RawTransactionGenResult
	//err = json.Unmarshal([]byte(rawResult), &rawTransactionGenResult)
	//if err != nil {
	//	return err
	//}
	////
	//var signTxParams adaptor.SignTransactionParams
	//signTxParams.TransactionHex = rawTransactionGenResult.Rawtx
	//if "" == prikey {
	//	signTxParams.FromAddr = aliceAddr
	//	signTxParams.Privkeys = append(signTxParams.Privkeys, gWallet.NameKey["alice"])
	//} else {
	//	signTxParams.FromAddr = btcadaptor.GetAddress(prikey)
	//	signTxParams.Privkeys = append(signTxParams.Privkeys, prikey)
	//}
	//signReuslt, err := btcadaptor.SignTransaction(&signTxParams)
	//if err != nil {
	//	return err
	//} else {
	//	var signTransactionResult adaptor.SignTransactionResult
	//	err := json.Unmarshal([]byte(signReuslt), &signTransactionResult)
	//	if err != nil {
	//		return err
	//	}
	//	fmt.Println(signTransactionResult.Complete)
	//	if signTransactionResult.Complete {
	//		sendTxParams := adaptor.SendTransactionHttpParams{TransactionHex: signTransactionResult.TransactionHex}
	//		sendResult, err := btcadaptor.SendTransactionHttp(&sendTxParams)
	//		if err != nil {
	//			return err
	//		}
	//		fmt.Println(sendResult)
	//	}
	//}

	return nil
}

func spendBTCFromMultiAddr(txid string, index string, txAmount string, fee string, toAddr string, redeem string, prikey string, prikey2 string, prikey3 string) error {
	////
	//vout, err := strconv.Atoi(index)
	//if err != nil {
	//	return errors.New("Index is Invalid.")
	//}
	//
	//amountValue, err := strconv.ParseFloat(txAmount, 64)
	//if err != nil {
	//	return errors.New("Amount is Invalid.")
	//}
	//feeValue, err := strconv.ParseFloat(fee, 64)
	//if err != nil {
	//	return errors.New("Fee is Invalid.")
	//}
	//
	//err = checkTxAmount(txid, vout, amountValue)
	//if err != nil {
	//	return err
	//}
	//
	////
	//var btcadaptor adaptorbtc.AdaptorBTC
	//btcadaptor.NetID = gWallet.BtcConfig.NetID
	////
	//var rawTransactionGenParams adaptor.RawTransactionGenParams
	//rawTransactionGenParams.Inputs = append(rawTransactionGenParams.Inputs, adaptor.Input{Txid: txid, Vout: uint32(vout)})
	//rawTransactionGenParams.Outputs = append(rawTransactionGenParams.Outputs, adaptor.Output{Address: toAddr, Amount: amountValue - feeValue})
	////
	//btcadaptor.Host = gWallet.BtcConfig.Host
	//btcadaptor.RPCUser = gWallet.BtcConfig.RPCUser
	//btcadaptor.RPCPasswd = gWallet.BtcConfig.RPCPasswd
	//btcadaptor.CertPath = gWallet.BtcConfig.CertPath
	////
	//rawResult, err := btcadaptor.RawTransactionGen(&rawTransactionGenParams)
	//if err != nil {
	//	return err
	//} else {
	//	fmt.Println("==== ==== Raw tx start ==== ====")
	//	fmt.Println(rawResult)
	//	fmt.Println("==== ==== Raw tx end ==== ====")
	//}
	////
	//var rawTransactionGenResult adaptor.RawTransactionGenResult
	//err = json.Unmarshal([]byte(rawResult), &rawTransactionGenResult)
	//if err != nil {
	//	return err
	//}
	////
	//var signTxParams adaptor.SignTransactionParams
	//signTxParams.TransactionHex = rawTransactionGenResult.Rawtx
	//signTxParams.InputRedeemIndex = []int{0}
	//signTxParams.RedeemHex = append(signTxParams.RedeemHex, redeem)
	//signTxParams.Privkeys = append(signTxParams.Privkeys, prikey)
	//signTxParams.Privkeys = append(signTxParams.Privkeys, prikey2)
	//signTxParams.Privkeys = append(signTxParams.Privkeys, prikey3)
	//signReuslt, err := btcadaptor.SignTransaction(&signTxParams)
	//if err != nil {
	//	return err
	//} else {
	//	fmt.Println("==== ==== Signed tx start ==== ====")
	//	fmt.Println(signReuslt)
	//	fmt.Println("==== ==== Signed tx end ==== ====")
	//}
	//var signTxResult adaptor.SignTransactionResult
	//err = json.Unmarshal([]byte(signReuslt), &signTxResult)
	//if err != nil {
	//	return err
	//}
	//if signTxResult.Complete {
	//	sendTxParams := adaptor.SendTransactionHttpParams{TransactionHex: signTxResult.TransactionHex}
	//	sendResult, err := btcadaptor.SendTransactionHttp(&sendTxParams)
	//	if err != nil {
	//		return err
	//	}
	//	fmt.Println(sendResult)
	//} else {
	//	fmt.Println("Not complete")
	//}
	return nil
}

func spendBTCFromMultiAddr2(txid string, index string, txAmount string, txid2 string, index2 string, txAmount2 string,
	fee string, toAddr string, redeem string, redeem2 string, prikey string, prikey2 string) error {
	////
	//vout, err := strconv.Atoi(index)
	//if err != nil {
	//	return errors.New("Index is Invalid.")
	//}
	//vout2, err := strconv.Atoi(index2)
	//if err != nil {
	//	return errors.New("Index2 is Invalid.")
	//}
	//
	//amountValue, err := strconv.ParseFloat(txAmount, 64)
	//if err != nil {
	//	return errors.New("Amount is Invalid.")
	//}
	//amountValue2, err := strconv.ParseFloat(txAmount2, 64)
	//if err != nil {
	//	return errors.New("Amount2 is Invalid.")
	//}
	//feeValue, err := strconv.ParseFloat(fee, 64)
	//if err != nil {
	//	return errors.New("Fee is Invalid.")
	//}
	//
	//err = checkTxAmount(txid, vout, amountValue)
	//if err != nil {
	//	return err
	//}
	//err = checkTxAmount(txid2, vout2, amountValue2)
	//if err != nil {
	//	return err
	//}
	////
	//var btcadaptor adaptorbtc.AdaptorBTC
	//btcadaptor.NetID = gWallet.BtcConfig.NetID
	////
	//var rawTransactionGenParams adaptor.RawTransactionGenParams
	//rawTransactionGenParams.Inputs = append(rawTransactionGenParams.Inputs, adaptor.Input{Txid: txid, Vout: uint32(vout)})
	//rawTransactionGenParams.Inputs = append(rawTransactionGenParams.Inputs, adaptor.Input{Txid: txid2, Vout: uint32(vout2)})
	//rawTransactionGenParams.Outputs = append(rawTransactionGenParams.Outputs, adaptor.Output{Address: toAddr, Amount: amountValue + amountValue2 - feeValue})
	////
	//btcadaptor.Host = gWallet.BtcConfig.Host
	//btcadaptor.RPCUser = gWallet.BtcConfig.RPCUser
	//btcadaptor.RPCPasswd = gWallet.BtcConfig.RPCPasswd
	//btcadaptor.CertPath = gWallet.BtcConfig.CertPath
	////
	//rawResult, err := btcadaptor.RawTransactionGen(&rawTransactionGenParams)
	//if err != nil {
	//	return err
	//} else {
	//	fmt.Println("==== ==== Raw tx start ==== ====")
	//	fmt.Println(rawResult)
	//	fmt.Println("==== ==== Raw tx end ==== ====")
	//}
	////
	//var rawTransactionGenResult adaptor.RawTransactionGenResult
	//err = json.Unmarshal([]byte(rawResult), &rawTransactionGenResult)
	//if err != nil {
	//	return err
	//}
	////
	//var signTxParams adaptor.SignTransactionParams
	//signTxParams.TransactionHex = rawTransactionGenResult.Rawtx
	//signTxParams.InputRedeemIndex = []int{0, 1}
	//signTxParams.RedeemHex = append(signTxParams.RedeemHex, redeem)
	//signTxParams.RedeemHex = append(signTxParams.RedeemHex, redeem2)
	//signTxParams.Privkeys = append(signTxParams.Privkeys, prikey)
	//signTxParams.Privkeys = append(signTxParams.Privkeys, prikey2)
	//signReuslt, err := btcadaptor.SignTransaction(&signTxParams)
	//if err != nil {
	//	return err
	//} else {
	//	fmt.Println("==== ==== Signed tx start ==== ====")
	//	fmt.Println(signReuslt)
	//	fmt.Println("==== ==== Signed tx end ==== ====")
	//}
	//var signTxResult adaptor.SignTransactionResult
	//err = json.Unmarshal([]byte(signReuslt), &signTxResult)
	//if err != nil {
	//	return err
	//}
	//if signTxResult.Complete {
	//	sendTxParams := adaptor.SendTransactionHttpParams{TransactionHex: signTxResult.TransactionHex}
	//	sendResult, err := btcadaptor.SendTransactionHttp(&sendTxParams)
	//	if err != nil {
	//		return err
	//	}
	//	fmt.Println(sendResult)
	//} else {
	//	fmt.Println("Not complete")
	//}
	return nil
}

func signTx(rawtx, redeem1, redeem2, prikey string) (string, error) {
	//var signTxParams adaptor.SignTransactionParams
	//signTxParams.TransactionHex = rawtx
	//signTxParams.InputRedeemIndex = []int{0, 1}
	//signTxParams.RedeemHex = append(signTxParams.RedeemHex, redeem1)
	//signTxParams.RedeemHex = append(signTxParams.RedeemHex, redeem2)
	//signTxParams.Privkeys = append(signTxParams.Privkeys, prikey)
	//
	////
	//var btcadaptor adaptorbtc.AdaptorBTC
	//btcadaptor.NetID = gWallet.BtcConfig.NetID
	//signReuslt, err := btcadaptor.SignTransaction(&signTxParams)
	//if err != nil {
	//	return "", err
	//} else {
	//	fmt.Println("==== ==== Signed tx start ==== ====")
	//	fmt.Println(signReuslt)
	//	fmt.Println("==== ==== Signed tx end ==== ====")
	//}
	//var signTxResult adaptor.SignTransactionResult
	//err = json.Unmarshal([]byte(signReuslt), &signTxResult)
	//if err != nil {
	//	return "", err
	//}
	//return signTxResult.TransactionHex, nil
	return "", nil
}

func mergeMultiAddr2(txid string, index string, txAmount string, txid2 string, index2 string, txAmount2 string,
	fee string, toAddr string, redeem string, redeem2 string, prikey string, prikey2 string) error {
	////
	//vout, err := strconv.Atoi(index)
	//if err != nil {
	//	return errors.New("Index is Invalid.")
	//}
	//vout2, err := strconv.Atoi(index2)
	//if err != nil {
	//	return errors.New("Index2 is Invalid.")
	//}
	//
	//amountValue, err := strconv.ParseFloat(txAmount, 64)
	//if err != nil {
	//	return errors.New("Amount is Invalid.")
	//}
	//amountValue2, err := strconv.ParseFloat(txAmount2, 64)
	//if err != nil {
	//	return errors.New("Amount2 is Invalid.")
	//}
	//feeValue, err := strconv.ParseFloat(fee, 64)
	//if err != nil {
	//	return errors.New("Fee is Invalid.")
	//}
	//
	//err = checkTxAmount(txid, vout, amountValue)
	//if err != nil {
	//	return err
	//}
	//err = checkTxAmount(txid2, vout2, amountValue2)
	//if err != nil {
	//	return err
	//}
	////
	//var btcadaptor adaptorbtc.AdaptorBTC
	//btcadaptor.NetID = gWallet.BtcConfig.NetID
	////
	//var rawTransactionGenParams adaptor.RawTransactionGenParams
	//rawTransactionGenParams.Inputs = append(rawTransactionGenParams.Inputs, adaptor.Input{Txid: txid, Vout: uint32(vout)})
	//rawTransactionGenParams.Inputs = append(rawTransactionGenParams.Inputs, adaptor.Input{Txid: txid2, Vout: uint32(vout2)})
	//rawTransactionGenParams.Outputs = append(rawTransactionGenParams.Outputs, adaptor.Output{Address: toAddr, Amount: amountValue + amountValue2 - feeValue})
	////
	//btcadaptor.Host = gWallet.BtcConfig.Host
	//btcadaptor.RPCUser = gWallet.BtcConfig.RPCUser
	//btcadaptor.RPCPasswd = gWallet.BtcConfig.RPCPasswd
	//btcadaptor.CertPath = gWallet.BtcConfig.CertPath
	////
	//rawResult, err := btcadaptor.RawTransactionGen(&rawTransactionGenParams)
	//if err != nil {
	//	return err
	//} else {
	//	fmt.Println("==== ==== Raw tx start ==== ====")
	//	fmt.Println(rawResult)
	//	fmt.Println("==== ==== Raw tx end ==== ====")
	//}
	////
	//var rawTransactionGenResult adaptor.RawTransactionGenResult
	//err = json.Unmarshal([]byte(rawResult), &rawTransactionGenResult)
	//if err != nil {
	//	return err
	//}
	//
	//sign1, err := signTx(rawTransactionGenResult.Rawtx, redeem, redeem2, prikey)
	//if err != nil {
	//	return err
	//}
	//sign2, err := signTx(rawTransactionGenResult.Rawtx, redeem, redeem2, prikey2)
	//if err != nil {
	//	return err
	//}
	////
	//var mergeTxParams adaptor.MergeTransactionParams
	//mergeTxParams.UserTransactionHex = rawTransactionGenResult.Rawtx
	//mergeTxParams.InputRedeemIndex = []int{0, 1}
	//mergeTxParams.RedeemHex = append(mergeTxParams.RedeemHex, redeem)
	//mergeTxParams.RedeemHex = append(mergeTxParams.RedeemHex, redeem2)
	//mergeTxParams.MergeTransactionHexs = append(mergeTxParams.MergeTransactionHexs, sign1)
	//mergeTxParams.MergeTransactionHexs = append(mergeTxParams.MergeTransactionHexs, sign2)
	//mergeReuslt, err := btcadaptor.MergeTransaction(&mergeTxParams)
	//if err != nil {
	//	return err
	//} else {
	//	fmt.Println("==== ==== merge tx start ==== ====")
	//	fmt.Println(mergeReuslt)
	//	fmt.Println("==== ==== merge tx end ==== ====")
	//}
	//var mergeTxResult adaptor.MergeTransactionResult
	//err = json.Unmarshal([]byte(mergeReuslt), &mergeTxResult)
	//if err != nil {
	//	fmt.Println("Not complete1")
	//	return err
	//}
	//
	//if mergeTxResult.Complete {
	//	sendTxParams := adaptor.SendTransactionHttpParams{TransactionHex: mergeTxResult.TransactionHex}
	//	sendResult, err := btcadaptor.SendTransactionHttp(&sendTxParams)
	//	if err != nil {
	//		return err
	//	}
	//	fmt.Println(sendResult)
	//} else {
	//	fmt.Println("Not complete")
	//}
	return nil
}

func helper() {
	fmt.Println("functions : init, give, getbalance, sendtomulti, spendmulti, spendmultidouble")
	fmt.Println("Params : init")
	fmt.Println("Params : give, txid, index, txAmount, fee, prikey")
	fmt.Println("Params : getbalance, addr")
	fmt.Println("Params : sendToMulti, txid, index, txAmount, fee, multiSigAddr")
	fmt.Println("Params : spendmulti, txid, index, amount, fee, toAddr, redeem, key1, key2, key3")
	fmt.Println("Params : spendmulti2, txid, index, amount, txid2, index2, amount2, fee, toAddr, redeem, redeem2, key1, key2")
	fmt.Println("Params : mergemulti2, txid, index, amount, txid2, index2, amount2, fee, toAddr, redeem, redeem2, key1, key2")
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
	case "init": //init alice's key
		createKey("alice")
	case "give": //give alice some btc for test
		if len(args) < 7 {
			fmt.Println("Params : give, txid, index, amount, fee, prikey")
			return
		}
		err := giveAlice(args[2], args[3], args[4], args[5], args[6])
		if err != nil {
			fmt.Println(err.Error())
		}
	case "getbalance":
		if len(args) < 2 {
			fmt.Println("Params : getbalance, addr")
			return
		}
		err := getBalance(args[2])
		if err != nil {
			fmt.Println(err.Error())
		}
	case "sendtomulti": //alice send btc to multisigAddr
		if len(args) < 7 {
			fmt.Println("Params : sendToMulti, txid, index, amount, fee, multiSigAddr, [prikey]")
			return
		}
		prikey := ""
		if len(args) > 7 {
			prikey = args[7]
		}
		err := aliceSendBTCToMultiSigAddr(args[2], args[3], args[4], args[5], args[6], prikey)
		if err != nil {
			fmt.Println(err.Error())
		}
	case "spendmulti":
		if len(args) < 11 {
			fmt.Println("Params : spendmulti, txid, index, amount, fee, toAddr, redeem, key1, key2, key3")
			return
		}
		err := spendBTCFromMultiAddr(args[2], args[3], args[4], args[5], args[6], args[7], args[8], args[9], args[10])
		if err != nil {
			fmt.Println(err.Error())
		}
	case "mergemulti2":
		if len(args) < 14 {
			fmt.Println("Params : mergemulti2, txid, index, amount, txid2, index2, amount2, fee, toAddr, redeem, redeem2, key1, key2")
			return
		}
		start := time.Now()
		err := mergeMultiAddr2(args[2], args[3], args[4], args[5], args[6], args[7], args[8], args[9], args[10], args[11], args[12], args[13])
		if err != nil {
			fmt.Println(err.Error())
		}
		end := time.Now()
		fmt.Println(end.Sub(start))
	case "spendmulti2":
		if len(args) < 14 {
			fmt.Println("Params : spendmulti2, txid, index, amount, txid2, index2, amount2, fee, toAddr, redeem, redeem2, key1, key2")
			return
		}
		start := time.Now()
		err := spendBTCFromMultiAddr2(args[2], args[3], args[4], args[5], args[6], args[7], args[8], args[9], args[10], args[11], args[12], args[13])
		if err != nil {
			fmt.Println(err.Error())
		}
		end := time.Now()
		fmt.Println(end.Sub(start))
	default:
		fmt.Println("Invalid cmd.")
		helper()
	}
}
