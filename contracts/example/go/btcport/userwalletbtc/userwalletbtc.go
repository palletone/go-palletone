package main

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/shopspring/decimal"
	"math/big"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"github.com/naoina/toml"

	"github.com/palletone/adaptor"
	"github.com/palletone/btc-adaptor"
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
	NameKey     map[string][]byte
	NamePubkey  map[string][]byte
	NameAddress map[string]string
	AddressKey  map[string][]byte
}

var gWallet = NewWallet()
var gWalletFile = "./btcwallet.toml"

var gTomlConfig = toml.DefaultConfig

func NewWallet() *MyWallet {
	return &MyWallet{
		BtcConfig: BTCConfig{
			NetID:        1,
			Host:         "localhost:18334",
			RPCUser:      "test",
			RPCPasswd:    "123456",
			CertPath:     "C:\\Users\\zxl\\AppData\\Local\\Btcd\\rpc.cert",
			WalletPasswd: "1",
		},
		NameKey:     map[string][]byte{},
		NamePubkey:  map[string][]byte{},
		NameAddress: map[string]string{},
		AddressKey:  map[string][]byte{}}
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
	btc := btcadaptor.NewAdaptorBTC(gWallet.BtcConfig.NetID, btcadaptor.RPCParams{})
	//
	key, _ := btc.NewPrivateKey(&adaptor.NewPrivateKeyInput{})
	gWallet.NameKey[name] = key.PrivateKey

	//
	pubkey, _ := btc.GetPublicKey(&adaptor.GetPublicKeyInput{PrivateKey: key.PrivateKey})
	gWallet.NamePubkey[name] = pubkey.PublicKey

	//
	address, _ := btc.GetAddress(&adaptor.GetAddressInput{Key: pubkey.PublicKey})
	gWallet.NameAddress[name] = address.Address
	gWallet.AddressKey[address.Address] = key.PrivateKey

	return saveConfig(gWalletFile, gWallet)

	return saveConfig(gWalletFile, gWallet)
}

type outputIndexValue struct {
	OutputIndex string
	Value       uint64
}

// A slice of outputIndexValue that implements sort.Interface to sort by Value.
type outputIndexValueList []outputIndexValue

func (p outputIndexValueList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p outputIndexValueList) Len() int           { return len(p) }
func (p outputIndexValueList) Less(i, j int) bool { return p[i].Value > p[j].Value }

// A function to turn a map into a outputIndexValueList, then sort and return it.
func sortByValue(tpl outputIndexValueList) outputIndexValueList {
	sort.Stable(tpl) //sort.Sort(tpl)
	return tpl
}

func selUnspends(outputIndexMap map[string]float64, btcAmout uint64) []outputIndexValue {
	var smlUnspends []outputIndexValue
	var bigUnspends []outputIndexValue
	var selUnspends []outputIndexValue
	for outputIndex, value := range outputIndexMap {
		amount := uint64(value * 1e8)
		if amount == btcAmout {
			selUnspends = append(selUnspends, outputIndexValue{outputIndex, amount})
			break
		} else if amount > btcAmout {
			bigUnspends = append(bigUnspends, outputIndexValue{outputIndex, amount})
		} else {
			smlUnspends = append(smlUnspends, outputIndexValue{outputIndex, amount})
		}
	}
	//
	if len(selUnspends) != 0 {
		return selUnspends
	}
	//
	selAmount := uint64(0)
	if len(smlUnspends) > 0 {
		smlUnspendsSort := sortByValue(smlUnspends)
		for i := range smlUnspendsSort {
			selAmount += smlUnspends[i].Value
			selUnspends = append(selUnspends, smlUnspends[i])
			if selAmount >= btcAmout {
				break
			}
		}
	}
	if selAmount >= btcAmout {
		return selUnspends
	}
	//
	if len(bigUnspends) == 0 {
		return bigUnspends
	}
	selUnspends = []outputIndexValue{}
	minIndex := int64(0)
	minValue := bigUnspends[0].Value
	for i := range bigUnspends {
		if bigUnspends[i].Value < minValue {
			minIndex = int64(i)
			minValue = bigUnspends[i].Value
		}
	}
	selUnspends = append(selUnspends, bigUnspends[minIndex])
	return selUnspends
}

type CreateTransferTokenTxInput struct {
	FromAddress string               `json:"from_address"`
	ToAddress   string               `json:"to_address"`
	Amount      *adaptor.AmountAsset `json:"amount"`
	Fee         *adaptor.AmountAsset `json:"fee"`
	Data        []byte               `json:"data"`
	Extra       []byte               `json:"extra"`
}

func CreateTransferTokenTx(input *CreateTransferTokenTxInput, rpcParams *btcadaptor.RPCParams, netID int) (*adaptor.CreateTransferTokenTxOutput, error) {
	//chainnet
	realNet := btcadaptor.GetNet(netID)

	//convert address from string
	addr, err := btcutil.DecodeAddress(input.FromAddress, realNet)
	if err != nil {
		return nil, fmt.Errorf("DecodeAddress FromAddress failed %s", err.Error())
	}
	if len(input.Extra)%33 != 0 {
		return nil, fmt.Errorf("input.Extra len invalid, txid:22+index:1")
	}

	//get rpc client
	client, err := btcadaptor.GetClient(rpcParams)
	if err != nil {
		return nil, err
	}
	defer client.Shutdown()

	//check amount
	fee := +input.Fee.Amount.Uint64()
	if 0 == fee {
		return nil, fmt.Errorf("input.Fee invalid, must not be zero")
	}
	amount := input.Amount.Amount.Uint64()
	btcAmount := amount + fee

	//1.get all unspend
	outputIndexMap, err := getAllUnspend(client, addr)
	if err != nil {
		return nil, err
	}
	if len(outputIndexMap) == 0 {
		return nil, fmt.Errorf("getAllUnspend failed : no utxos")
	}
	//for outputIndex, value := range outputIndexMap {
	//	fmt.Println(outputIndex, value)
	//}

	//2.remove extra utxo
	for i := 0; i < len(input.Extra); i += 33 {
		idIndexHex := hex.EncodeToString(input.Extra[i:33])
		if _, exist := outputIndexMap[idIndexHex]; exist {
			delete(outputIndexMap, idIndexHex)
		}
	}

	//3.select greet
	outputIndexSel := selUnspends(outputIndexMap, btcAmount)
	if len(outputIndexSel) == 0 {
		return nil, fmt.Errorf("selUnspends failed : balance is not enough")
	}
	//for _, out := range outputIndexSel { //Debug
	//	fmt.Println(out.OutputIndex, out.Value)
	//}

	msgTx := wire.NewMsgTx(1)
	//transaction inputs
	allInputAmount := uint64(0)
	extra := []byte{}
	for _, outputIndexV := range outputIndexSel {
		//fmt.Println(outputIndexV.OutputIndex, outputIndexV.Value)
		voutByte, _ := hex.DecodeString(outputIndexV.OutputIndex[64:66])
		vout := uint64(voutByte[0])
		hash, err := chainhash.NewHashFromStr(outputIndexV.OutputIndex[0:64])
		if err != nil {
			return nil, fmt.Errorf("NewHashFromStr outputIndexSel failed")
		}
		input := &wire.TxIn{PreviousOutPoint: wire.OutPoint{*hash, uint32(vout)}}
		msgTx.AddTxIn(input)
		allInputAmount += outputIndexV.Value
		outputIndexByte, _ := hex.DecodeString(outputIndexV.OutputIndex)
		extra = append(extra, outputIndexByte...)
	}
	if len(msgTx.TxIn) == 0 {
		return nil, fmt.Errorf("Process TxIn error : NO Input.")
	}

	//transaction outputs
	addrTo, err := btcutil.DecodeAddress(input.ToAddress, realNet)
	if err != nil {
		return nil, fmt.Errorf("DecodeAddress ToAddress failed %s", err.Error())
	} else {
		pkScript, _ := txscript.PayToAddrScript(addrTo)
		txOut := wire.NewTxOut(int64(amount), pkScript)
		msgTx.AddTxOut(txOut)
	}
	if 0 != len(input.Data) {
		pkScript, _ := txscript.NullDataScript(input.Data)
		txOut := wire.NewTxOut(int64(0), pkScript)
		msgTx.AddTxOut(txOut)
	}

	//change
	change := allInputAmount - amount - fee
	//fmt.Println(change, allInputAmount, btcAmount, fee) //Debug
	if change > 0 {
		pkScript, _ := txscript.PayToAddrScript(addr)
		txOut := wire.NewTxOut(int64(change), pkScript)
		msgTx.AddTxOut(txOut)
	}
	if len(msgTx.TxOut) == 0 {
		return nil, fmt.Errorf("Process TxOut error : NO Output.")
	}
	for _, out := range msgTx.TxOut { //Debug
		fmt.Println(out.Value)
	}

	//SerializeSize transaction to bytes
	buf := bytes.NewBuffer(make([]byte, 0, msgTx.SerializeSize()))
	if err := msgTx.Serialize(buf); err != nil {
		return nil, err
	}
	//result for return
	var output adaptor.CreateTransferTokenTxOutput
	output.Transaction = buf.Bytes()
	output.Extra = extra

	return &output, nil
}

func helper() {
	fmt.Println("functions : init, give, getbalance, sendtomulti, spendmulti, spendmultidouble")
	fmt.Println("Params : init")
	fmt.Println("Params : sendToMulti, multiSigAddr, amount, fee, ptnAddr, btcPrivateKey")
}

func getAllUnspend(client *rpcclient.Client, addr btcutil.Address) (map[string]float64, error) {
	//get all raw transaction
	count := 999999
	msgTxs, err := client.SearchRawTransactionsVerbose(addr, 0, count, true, false, []string{}) //BTCD API
	if err != nil {
		return map[string]float64{}, fmt.Errorf("SearchRawTransactionsVerbose failed %s", err.Error())
	}

	addrStr := addr.String()
	//save utxo to map, check next one transanction is spend or not
	outputIndex := map[string]float64{}
	//the result for return
	for _, msgTx := range msgTxs {
		if int(msgTx.Confirmations) < btcadaptor.MinConfirm {
			continue
		}
		//transaction inputs
		for _, in := range msgTx.Vin {
			//check is spend or not
			idIndex := in.Txid + fmt.Sprintf("%02x", in.Vout)
			_, exist := outputIndex[idIndex]
			if exist { //spend
				delete(outputIndex, idIndex)
			}
		}

		//transaction outputs
		for _, out := range msgTx.Vout {
			if 0 == len(out.ScriptPubKey.Addresses) {
				continue
			}
			if out.ScriptPubKey.Addresses[0] == addrStr {
				outputIndex[msgTx.Txid+fmt.Sprintf("%02x", out.N)] = out.Value
			}
		}
	}

	return outputIndex, nil
}

func getAddrByPrikey(prikey []byte) string {
	btc := btcadaptor.NewAdaptorBTC(gWallet.BtcConfig.NetID, btcadaptor.RPCParams{})
	pubkey, _ := btc.GetPublicKey(&adaptor.GetPublicKeyInput{PrivateKey: prikey})
	address, _ := btc.GetAddress(&adaptor.GetAddressInput{Key: pubkey.PublicKey})
	return address.Address
}

func signAndSend(btc *btcadaptor.AdaptorBTC, rawTransaction, prikey []byte, redeemHex string) error {
	//2.sign tx
	var input adaptor.SignTransactionInput
	input.PrivateKey = prikey
	//input.Transaction = Hex2Bytes("f9024981848203e883200b20946817cfb2c442693d850332c3b755b2342ec4afb280b902248c2e032100000000000000000000000000000000000000000000000000000000000000c0000000000000000000000000aaa919a7c465be9b053673c567d73be8603179630000000000000000000000000000000000000000000000000de0b6b3a76400000000000000000000000000000000000000000000000000000000000000000001000000000000000000000000000000000000000000000000000000000000012000000000000000000000000000000000000000000000000000000000000001a000000000000000000000000000000000000000000000000000000000000000407d7116a8706ae08baa7f4909e26728fa7a5f0365aaa919a7c465be9b053673c567d73be8603179636c7110482920e0af149a82189251f292a84148a85b7cd70d00000000000000000000000000000000000000000000000000000000000000417197961c5ae032ed6f33650f1f3a3ba111e8548a3dad14b3afa1cb6bc8f4601a6cb2b21aedcd575784e923942f3130f3290d56522ab2b28afca478e489426a4601000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000041ae94b0e599ef0508ba7bec41db5b46d5a065b30d3d5c4b0a4c85ea2d4899d6607e80e3314ee0741049963d30fb3aceaa5506e13835a41ef54a8f44a04ef0f1e40100000000000000000000000000000000000000000000000000000000000000808080")
	input.Transaction = rawTransaction
	//set input address
	if "" != redeemHex {
		input.Extra = []byte(redeemHex)
	} else {
		addr := getAddrByPrikey(prikey)
		input.Extra = []byte(addr)
	}
	resultSign, err := btc.SignTransaction(&input)
	if err != nil {
		fmt.Println("SignTransaction failed : ", err.Error())
		return err
	} else {
		fmt.Printf("tx: %x\n", resultSign.SignedTx)
	}

	//3.send tx
	var sendTransactionParams adaptor.SendTransactionInput
	sendTransactionParams.Transaction = resultSign.SignedTx
	resultSend, err := btc.SendTransaction(&sendTransactionParams)
	if err != nil {
		fmt.Println("SendTransaction failed : ", err.Error())
		return err
	} else {
		fmt.Printf("txid : %x\n", resultSend.TxID)
	}
	return nil
}

func aliceSendBTCToMultiSigAddr(multiAddr, amount, fee, ptnAddr, prikeyHex string) error {
	amountValue, err := strconv.ParseFloat(amount, 64)
	if err != nil {
		return fmt.Errorf("Amount is Invalid.")
	}
	feeValue, err := strconv.ParseFloat(fee, 64)
	if err != nil {
		return fmt.Errorf("Fee is Invalid.")
	}

	if "0x" == prikeyHex[0:2] || "0X" == prikeyHex[0:2] {
		prikeyHex = prikeyHex[2:]
	}
	prikey, err := hex.DecodeString(prikeyHex)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	rpc := btcadaptor.RPCParams{Host: gWallet.BtcConfig.Host, RPCUser: gWallet.BtcConfig.RPCUser,
		RPCPasswd: gWallet.BtcConfig.RPCPasswd, CertPath: gWallet.BtcConfig.CertPath}

	callerAddr := getAddrByPrikey(prikey)

	var txInput CreateTransferTokenTxInput
	txInput.FromAddress = callerAddr
	txInput.ToAddress = multiAddr

	amt := new(big.Int)
	amt.SetUint64(uint64(decimal.NewFromFloat(amountValue).Mul(decimal.New(1, 8)).IntPart()))
	txInput.Amount = adaptor.NewAmountAsset(amt, "BTC")
	amtFee := new(big.Int)
	amtFee.SetUint64(uint64(decimal.NewFromFloat(feeValue).Mul(decimal.New(1, 8)).IntPart())) //dao

	//fmt.Println(amountValue, feeValue, amt.String(), amtFee.String())
	//amtInt := uint64(decimal.NewFromFloat(amountValue).Mul(decimal.New(1, 8)).IntPart())
	//feeInt := uint64(decimal.NewFromFloat(feeValue).Mul(decimal.New(1, 8)).IntPart())
	//fmt.Println(amtInt, feeInt)
	//return nil

	txInput.Fee = adaptor.NewAmountAsset(amtFee, "BTC")
	txInput.Data = []byte(ptnAddr)

	resultTx, err := CreateTransferTokenTx(&txInput, &rpc, gWallet.BtcConfig.NetID)
	if err != nil {
		fmt.Println(err.Error())
		return err
	} else {
		fmt.Println(resultTx)
	}

	//return nil
	btc := btcadaptor.NewAdaptorBTC(gWallet.BtcConfig.NetID, rpc)
	return signAndSend(btc, resultTx.Transaction, prikey, "")
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
	case "sendtomulti": //alice send btc to multisigAddr
		if len(args) < 7 {
			fmt.Println("Params : sendToMulti, multiSigAddr, amount, fee, ptnAddr, btcPrivateKey")
			return
		}
		err := aliceSendBTCToMultiSigAddr(args[2], args[3], args[4], args[5], args[6])
		if err != nil {
			fmt.Println(err.Error())
		}
	default:
		fmt.Println("Invalid cmd.")
		helper()
	}
}
