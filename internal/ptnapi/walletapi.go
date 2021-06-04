package ptnapi

import (
	"bufio"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/palletone/go-palletone/core/accounts/keystore"

	"github.com/ethereum/go-ethereum/rlp"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/crypto"
	"github.com/palletone/go-palletone/common/hexutil"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/math"
	"github.com/palletone/go-palletone/core"
	"github.com/palletone/go-palletone/core/accounts"
	"github.com/palletone/go-palletone/dag/dagconfig"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/ptnjson"
	"github.com/palletone/go-palletone/ptnjson/walletjson"
	"github.com/palletone/go-palletone/tokenengine"
	"github.com/shopspring/decimal"
)

// Start forking command.
func (s *PublicWalletAPI) Forking(ctx context.Context, rate uint64) uint64 {
	return forking(ctx, s.b)
}
func (s *PrivateWalletAPI) Forking(ctx context.Context, rate uint64) uint64 {
	return forking(ctx, s.b)
}

type PublicWalletAPI struct {
	b Backend
}

type PrivateWalletAPI struct {
	b    Backend
	lock sync.Mutex
}

func NewPublicWalletAPI(b Backend) *PublicWalletAPI {
	return &PublicWalletAPI{b}
}
func NewPrivateWalletAPI(b Backend) *PrivateWalletAPI {
	return &PrivateWalletAPI{b: b}
}
func (s *PublicWalletAPI) CreateRawTransaction(ctx context.Context, from string, to string, amount, fee decimal.Decimal, lockTime uint32) (string, error) {
	s.b.Lock()
	defer s.b.Unlock()
	tx, _, err := buildRawTransferTx(s.b, dagconfig.DagConfig.GasToken, from, to, amount, fee, "")
	if err != nil {
		return "", err
	}
	if lockTime != 0 {
		pay := tx.Messages()[0].Payload.(*modules.PaymentPayload)
		pay.LockTime = lockTime
		tx.ModifiedMsg(0, modules.NewMessage(modules.APP_PAYMENT, pay))
	}

	mtxbt, err := rlp.EncodeToBytes(tx)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(mtxbt), nil
}

func (s *PrivateWalletAPI) SignRawTransaction(ctx context.Context, params string, hashtype string, password string, duration *Int) (ptnjson.SignRawTransactionResult, error) {

	//transaction inputs
	if params == "" {
		return ptnjson.SignRawTransactionResult{}, errors.New("Params is empty")
	}
	upper_type := strings.ToUpper(hashtype)
	if upper_type != ALL && upper_type != NONE && upper_type != SINGLE {
		return ptnjson.SignRawTransactionResult{}, errors.New("Hashtype is error,error type:" + hashtype)
	}
	serializedTx, err := decodeHexStr(params)
	if err != nil {
		return ptnjson.SignRawTransactionResult{}, errors.New("Params is invalid")
	}
	tx := modules.NewTransaction(make([]*modules.Message, 0))

	if err := rlp.DecodeBytes(serializedTx, &tx); err != nil {
		return ptnjson.SignRawTransactionResult{}, errors.New("Params decode is invalid")
	}

	getPubKeyFn := func(addr common.Address) ([]byte, error) {
		//TODO use keystore
		ks := s.b.GetKeyStore()

		return ks.GetPublicKey(addr)
		//privKey, _ := ks.DumpPrivateKey(account, "1")
		//return crypto.CompressPubkey(&privKey.PublicKey), nil
	}
	getSignFn := func(addr common.Address, msg []byte) ([]byte, error) {
		ks := s.b.GetKeyStore()
		//account, _ := MakeAddress(ks, addr.String())
		//privKey, _ := ks.DumpPrivateKey(account, "1")
		return ks.SignMessage(addr, msg)
		//return crypto.Sign(hash, privKey)
	}
	var srawinputs []ptnjson.RawTxInput

	var addr common.Address
	var keys []string
	for _, msg := range tx.TxMessages() {
		payload, ok := msg.Payload.(*modules.PaymentPayload)
		if !ok {
			continue
		}
		for _, txin := range payload.Inputs {
			inpoint := modules.OutPoint{
				TxHash:       txin.PreviousOutPoint.TxHash,
				OutIndex:     txin.PreviousOutPoint.OutIndex,
				MessageIndex: txin.PreviousOutPoint.MessageIndex,
			}
			uvu, eerr := s.b.GetUtxoEntry(&inpoint)
			if eerr != nil {
				log.Error(eerr.Error())
				return ptnjson.SignRawTransactionResult{}, err
			}
			TxHash := trimx(uvu.TxHash)
			PkScriptHex := trimx(uvu.PkScriptHex)
			input := ptnjson.RawTxInput{Txid: TxHash, Vout: uvu.OutIndex, MessageIndex: uvu.MessageIndex, ScriptPubKey: PkScriptHex, RedeemScript: ""}
			srawinputs = append(srawinputs, input)
			addr, err = tokenengine.Instance.GetAddressFromScript(hexutil.MustDecode(uvu.PkScriptHex))
			if err != nil {
				log.Error(err.Error())
				return ptnjson.SignRawTransactionResult{}, errors.New("get addr FromScript is err")
			}
		}
		/*for _, txout := range payload.Outputs {
			err = tokenengine.ScriptValidate(txout.PkScript, tx, 0, 0)
			if err != nil {
			}
		}*/
	}
	//const max = uint64(time.Duration(math.MaxInt64) / time.Second)
	//var d time.Duration
	//if duration == nil {
	//	d = 300 * time.Second
	//} else if *duration > max {
	//	return ptnjson.SignRawTransactionResult{}, errors.New("unlock duration too large")
	//} else {
	//	d = time.Duration(*duration) * time.Second
	//}
	//ks := s.b.GetKeyStore()
	//err = ks.TimedUnlock(accounts.Account{Address: addr}, password, d)
	err = unlockKS(s.b, addr, password, duration)
	if err != nil {
		newErr := errors.New("get addr by outpoint get err:" + err.Error())
		log.Error(newErr.Error())
		return ptnjson.SignRawTransactionResult{}, newErr
	}

	newsign := ptnjson.NewSignRawTransactionCmd(params, &srawinputs, &keys, ptnjson.String(hashtype))
	result, err := SignRawTransactionOld(newsign, getPubKeyFn, getSignFn, addr)
	if !result.Complete {
		log.Error("Not complete!!!")
		for _, e := range result.Errors {
			log.Error("SignError:" + e.Error)
		}
	}
	return result, err
}

func (s *PrivateWalletAPI) MultiSignRawTransaction(ctx context.Context, params, redeemScript string, addr common.Address, hashtype string, password string, duration *Int) (ptnjson.SignRawTransactionResult, error) {

	//transaction inputs
	if params == "" {
		return ptnjson.SignRawTransactionResult{}, errors.New("Params is empty")
	}
	upper_type := strings.ToUpper(hashtype)
	if upper_type != ALL && upper_type != NONE && upper_type != SINGLE {
		return ptnjson.SignRawTransactionResult{}, errors.New("Hashtype is error,error type:" + hashtype)
	}
	serializedTx, err := decodeHexStr(params)
	if err != nil {
		return ptnjson.SignRawTransactionResult{}, errors.New("Params is invalid")
	}

	tx := new(modules.Transaction)
	if err := rlp.DecodeBytes(serializedTx, tx); err != nil {
		return ptnjson.SignRawTransactionResult{}, errors.New("Params decode is invalid")
	}
	rs, err := hex.DecodeString(redeemScript)
	if err != nil {
		return ptnjson.SignRawTransactionResult{}, errors.New("redeemScript is invalid")
	}
	bls := tokenengine.Instance.GenerateP2SHLockScript(crypto.Hash160(rs))
	lockScript := hex.EncodeToString(bls)
	getPubKeyFn := func(addr common.Address) ([]byte, error) {
		//TODO use keystore
		ks := s.b.GetKeyStore()

		return ks.GetPublicKey(addr)
	}
	getSignFn := func(addr common.Address, msg []byte) ([]byte, error) {
		ks := s.b.GetKeyStore()
		return ks.SignMessage(addr, msg)
	}
	var srawinputs []ptnjson.RawTxInput

	//var addr common.Address
	var keys []string
	for _, msg := range tx.TxMessages() {
		payload, ok := msg.Payload.(*modules.PaymentPayload)
		if !ok {
			continue
		}
		for _, txin := range payload.Inputs {
			inpoint := modules.OutPoint{
				TxHash:       txin.PreviousOutPoint.TxHash,
				OutIndex:     txin.PreviousOutPoint.OutIndex,
				MessageIndex: txin.PreviousOutPoint.MessageIndex,
			}
			uvu, eerr := s.b.GetUtxoEntry(&inpoint)
			if eerr != nil {
				log.Error(eerr.Error())
				return ptnjson.SignRawTransactionResult{}, err
			}
			TxHash := trimx(uvu.TxHash)
			//PkScriptHex := trimx(uvu.PkScriptHex)
			PkScriptHex := trimx(lockScript)
			input := ptnjson.RawTxInput{Txid: TxHash, Vout: uvu.OutIndex, MessageIndex: uvu.MessageIndex,
				ScriptPubKey: PkScriptHex, RedeemScript: redeemScript}
			srawinputs = append(srawinputs, input)
		}
	}
	err = unlockKS(s.b, addr, password, duration)
	if err != nil {
		newErr := errors.New("get addr by outpoint get err:" + err.Error())
		log.Error(newErr.Error())
		return ptnjson.SignRawTransactionResult{}, newErr
	}

	newsign := ptnjson.NewMultiSignRawTransactionCmd(params, &srawinputs, &keys, ptnjson.String(hashtype))
	result, err := MultiSignRawTransaction(newsign, getPubKeyFn, getSignFn, addr)
	if !result.Complete {
		log.Error("Not complete!!!")
		for _, e := range result.Errors {
			log.Error("SignError:" + e.Error)
		}
	}
	return result, err
}

// walletSendTransaction will add the signed transaction to the transaction pool.
// The sender is responsible for signing the transaction and using the correct nonce.
func (s *PublicWalletAPI) SendRawTransaction(ctx context.Context, signedTxHex string) (common.Hash, error) {

	serializedTx, err := hex.DecodeString(signedTxHex)
	if err != nil {
		return common.Hash{}, errors.New("Decode Signedtx is invalid")
	}

	tx := modules.NewTransaction(make([]*modules.Message, 0))
	if err := rlp.DecodeBytes(serializedTx, tx); err != nil {
		return common.Hash{}, errors.New("encodedTx decode is invalid")
	}

	if 0 == len(tx.Messages()) {
		return common.Hash{}, errors.New("Invalid Tx, message length is 0")
	}
	//var outAmount uint64
	//var outpoint_txhash common.Hash
	//for _, msg := range tx.TxMessages {
	//	payload, ok := msg.Payload.(*modules.PaymentPayload)
	//	if ok == false {
	//		continue
	//	}
	//
	//	for _, txout := range payload.Outputs {
	//		outAmount += txout.Value
	//	}
	//	//log.Info("payment info", "info", payload)
	//	outpoint_txhash = payload.Inputs[0].PreviousOutPoint.TxHash
	//}
	//log.Debugf("Tx outpoint tx hash:%s", outpoint_txhash.String())
	return submitTransaction(s.b, tx)
}

func (s *PublicWalletAPI) SendJsonTransaction(ctx context.Context, params string) (common.Hash, error) {

	decoded, err := hex.DecodeString(params)
	if err != nil {
		return common.Hash{}, errors.New("Decode Signedtx is invalid")
	}
	var btxjson []byte
	if err := rlp.DecodeBytes(decoded, &btxjson); err != nil {
		return common.Hash{}, errors.New("RLP Decode To Byte is invalid")
	}
	tx := modules.NewTransaction(make([]*modules.Message, 0))
	err = json.Unmarshal(btxjson, tx)
	if err != nil {
		return common.Hash{}, errors.New("Json Unmarshal To Tx is invalid")
	}
	msgs := tx.TxMessages()
	if 0 == len(msgs) {
		return common.Hash{}, errors.New("Invalid Tx, message length is 0")
	}
	var outAmount uint64
	var outpoint_txhash common.Hash
	for _, msg := range msgs {
		payload, ok := msg.Payload.(*modules.PaymentPayload)
		if !ok {
			continue
		}

		for _, txout := range payload.Outputs {
			outAmount += txout.Value
		}
		log.Info("payment info", "info", payload)
		outpoint_txhash = payload.Inputs[0].PreviousOutPoint.TxHash
	}
	log.Infof("Tx outpoint tx hash:%s", outpoint_txhash.String())
	return submitTransaction(s.b, tx)
}
func (s *PublicWalletAPI) SendRlpTransaction(ctx context.Context, encodedTx string) (common.Hash, error) {
	//transaction inputs
	if encodedTx == "" {
		return common.Hash{}, errors.New("Params is Empty")
	}
	tx := new(modules.Transaction)
	serializedTx, err := decodeHexStr(encodedTx)
	if err != nil {
		return common.Hash{}, errors.New("encodedTx is invalid")
	}

	if err := rlp.DecodeBytes(serializedTx, tx); err != nil {
		return common.Hash{}, errors.New("encodedTx decode is invalid")
	}
	if 0 == len(tx.Messages()) {
		return common.Hash{}, errors.New("Invalid Tx, message length is 0")
	}
	var outAmount uint64
	for _, msg := range tx.TxMessages() {
		payload, ok := msg.Payload.(*modules.PaymentPayload)
		if !ok {
			continue
		}

		for _, txout := range payload.Outputs {
			outAmount += txout.Value
		}
	}
	return submitTransaction(s.b, tx)
}

func (s *PublicWalletAPI) GetAddrUtxos(ctx context.Context, addr string) (string, error) {

	items, err := s.b.GetDagAddrUtxos(addr)
	if err != nil {
		return "", err
	}

	info := NewPublicReturnInfo("address_utxos", items)
	result_json, _ := json.Marshal(info)
	return string(result_json), nil

}
func (s *PublicWalletAPI) GetAddrUtxos2(ctx context.Context, addr string) (string, error) {
	utxos, err := s.b.GetAddrUtxos2(addr)
	if err != nil {
		return "", err
	}

	info := NewPublicReturnInfo("address_utxos", utxos)
	result_json, _ := json.Marshal(info)
	return string(result_json), nil

}
func (s *PublicWalletAPI) GetBalance(ctx context.Context, addr string) (map[string]decimal.Decimal, error) {
	realAddr, err := getRealAddress(s.b.GetKeyStore(), addr)
	if err != nil {
		return nil, err
	}
	address := realAddr.String()
	utxos, err := s.b.GetDagAddrUtxos(address)
	if err != nil {
		return nil, err
	}
	result := make(map[string]decimal.Decimal)
	for _, utxo := range utxos {
		asset, _ := modules.StringToAsset(utxo.Asset)
		if bal, ok := result[utxo.Asset]; ok {
			result[utxo.Asset] = bal.Add(ptnjson.AssetAmt2JsonAmt(asset, utxo.Amount))
		} else {
			result[utxo.Asset] = ptnjson.AssetAmt2JsonAmt(asset, utxo.Amount)
		}
	}
	return result, nil
}
func (s *PublicWalletAPI) GetBalance2(ctx context.Context, addr string) (*walletjson.StableUnstable, error) {
	realAddr, err := getRealAddress(s.b.GetKeyStore(), addr)
	if err != nil {
		return nil, err
	}
	address := realAddr.String()
	utxos, err := s.b.GetAddrUtxos2(address)
	if err != nil {
		return nil, err
	}
	balance1 := utxos2Balance(utxos, true)
	balance2 := utxos2Balance(utxos, false)
	return &walletjson.StableUnstable{Stable: balance1, Unstable: balance2}, nil
}
func utxos2Balance(utxos []*ptnjson.UtxoJson, stable bool) map[string]decimal.Decimal {
	result := make(map[string]decimal.Decimal)
	for _, utxo := range utxos {
		if stable && utxo.FlagStatus == "Unstable" {
			continue
		}
		if !stable && utxo.FlagStatus != "Unstable" {
			continue
		}
		asset, _ := modules.StringToAsset(utxo.Asset)
		if bal, ok := result[utxo.Asset]; ok {
			result[utxo.Asset] = bal.Add(ptnjson.AssetAmt2JsonAmt(asset, utxo.Amount))
		} else {
			result[utxo.Asset] = ptnjson.AssetAmt2JsonAmt(asset, utxo.Amount)
		}
	}
	return result
}

//func (s *PublicWalletAPI) GetTranscations(ctx context.Context, address string) (string, error) {
//	txs, err := s.b.GetAddrTransactions(address)
//	if err != nil {
//		return "null", err
//	}
//
//	gets := []ptnjson.GetTransactions{}
//	for _, tx := range txs {
//
//		get := ptnjson.GetTransactions{}
//		get.Txid = tx.Hash().String()
//
//		for _, msg := range tx.TxMessages {
//			payload, ok := msg.Payload.(*modules.PaymentPayload)
//
//			if ok == false {
//				continue
//			}
//
//			for _, txin := range payload.Inputs {
//
//				if txin.PreviousOutPoint != nil {
//					addr, err := s.b.GetAddrByOutPoint(txin.PreviousOutPoint)
//					if err != nil {
//
//						return "null", err
//					}
//
//					get.Inputs = append(get.Inputs, addr.String())
//				} else {
//					get.Inputs = append(get.Inputs, "coinbase")
//				}
//
//			}
//
//			for _, txout := range payload.Outputs {
//				var gout ptnjson.GetTranscationOut
//				addr, err := tokenengine.GetAddressFromScript(txout.PkScript)
//				if err != nil {
//					return "null", err
//				}
//				gout.Addr = addr.String()
//				gout.Value = txout.Value
//				gout.Asset = txout.Asset.String()
//				get.Outputs = append(get.Outputs, gout)
//			}
//
//			gets = append(gets, get)
//		}
//
//	}
//	result := ptnjson.ConvertGetTransactions2Json(gets)
//
//	return result, nil
//}

func (s *PublicWalletAPI) GetAddrTxHistory(ctx context.Context, addr string) ([]*ptnjson.TxHistoryJson, error) {
	result, err := s.b.GetAddrTxHistory(addr)

	return result, err
}
func (s *PublicWalletAPI) GetContractInvokeHistory(ctx context.Context, contractAddr string) ([]*ptnjson.ContractInvokeHistoryJson, error) {
	result, err := s.b.GetContractInvokeHistory(contractAddr)
	return result, err
}

//获得某地址的通证流水
func (s *PublicWalletAPI) GetAddrTokenFlow(ctx context.Context, addr string, token string) ([]*ptnjson.TokenFlowJson, error) {
	result, err := s.b.GetAddrTokenFlow(addr, token)
	return result, err
}

//sign rawtranscation
//create raw transction
func (s *PrivateWalletAPI) GetPtnTestCoin(ctx context.Context, from string, to string, amount,
	password string, duration *Int) (common.Hash, error) {
	//var LockTime int64
	LockTime := int64(0)

	amounts := []ptnjson.AddressAmt{}
	if to == "" {
		return common.Hash{}, nil
	}
	a, err := RandFromString(amount)
	if err != nil {
		return common.Hash{}, err
	}
	amounts = append(amounts, ptnjson.AddressAmt{Address: to, Amount: a})

	utxoJsons, err := s.b.GetDagAddrUtxos(from)
	if err != nil {
		return common.Hash{}, err
	}
	utxos := core.Utxos{}
	ptn := dagconfig.DagConfig.GetGasToken()
	for _, jsonu := range utxoJsons {
		//utxos = append(utxos, &json)
		if jsonu.Asset == ptn.String() {
			utxos = append(utxos, &ptnjson.UtxoJson{
				TxHash:         jsonu.TxHash,
				MessageIndex:   jsonu.MessageIndex,
				OutIndex:       jsonu.OutIndex,
				Amount:         jsonu.Amount,
				Asset:          jsonu.Asset,
				PkScriptHex:    jsonu.PkScriptHex,
				PkScriptString: jsonu.PkScriptString,
				LockTime:       jsonu.LockTime})
		}
	}
	fee, err := decimal.NewFromString("1")
	if err != nil {
		return common.Hash{}, err
	}
	daoAmount := ptnjson.Ptn2Dao(a.Add(fee))
	taken_utxo, change, err := core.Select_utxo_Greedy(utxos, daoAmount)
	if err != nil {
		return common.Hash{}, err
	}

	inputs := []ptnjson.TransactionInput{}
	var input ptnjson.TransactionInput
	for _, u := range taken_utxo {
		utxo := u.(*ptnjson.UtxoJson)
		input.Txid = utxo.TxHash
		input.MessageIndex = utxo.MessageIndex
		input.Vout = utxo.OutIndex
		inputs = append(inputs, input)
	}

	if change > 0 {
		amounts = append(amounts, ptnjson.AddressAmt{Address: from, Amount: ptnjson.Dao2Ptn(change)})
	}

	arg := ptnjson.NewCreateRawTransactionCmd(inputs, amounts, &LockTime)
	result, _ := CreateRawTransaction(arg)
	fmt.Println(result)
	//transaction inputs
	serializedTx, err := decodeHexStr(result)
	if err != nil {
		return common.Hash{}, err
	}

	tx := modules.NewTransaction(make([]*modules.Message, 0))
	if err := rlp.DecodeBytes(serializedTx, &tx); err != nil {
		return common.Hash{}, err
	}

	getPubKeyFn := func(addr common.Address) ([]byte, error) {
		//TODO use keystore
		ks := s.b.GetKeyStore()

		return ks.GetPublicKey(addr)
		//privKey, _ := ks.DumpPrivateKey(account, "1")
		//return crypto.CompressPubkey(&privKey.PublicKey), nil
	}
	getSignFn := func(addr common.Address, msg []byte) ([]byte, error) {
		ks := s.b.GetKeyStore()
		return ks.SignMessage(addr, msg)
		//return crypto.Sign(hash, privKey)
	}
	var srawinputs []ptnjson.RawTxInput

	var addr common.Address
	var keys []string
	for _, msg := range tx.TxMessages() {
		payload, ok := msg.Payload.(*modules.PaymentPayload)
		if !ok {
			continue
		}
		for _, txin := range payload.Inputs {
			inpoint := modules.OutPoint{
				TxHash:       txin.PreviousOutPoint.TxHash,
				OutIndex:     txin.PreviousOutPoint.OutIndex,
				MessageIndex: txin.PreviousOutPoint.MessageIndex,
			}
			uvu, eerr := s.b.GetUtxoEntry(&inpoint)
			if eerr != nil {
				return common.Hash{}, err
			}
			TxHash := trimx(uvu.TxHash)
			PkScriptHex := trimx(uvu.PkScriptHex)
			input := ptnjson.RawTxInput{Txid: TxHash, Vout: uvu.OutIndex, MessageIndex: uvu.MessageIndex, ScriptPubKey: PkScriptHex, RedeemScript: ""}
			srawinputs = append(srawinputs, input)
			addr, err = tokenengine.Instance.GetAddressFromScript(hexutil.MustDecode(uvu.PkScriptHex))
			if err != nil {
				return common.Hash{}, err
			}
		}
	}
	//const max = uint64(time.Duration(math.MaxInt64) / time.Second)
	//var d time.Duration
	//if duration == nil {
	//	d = 300 * time.Second
	//} else if *duration > max {
	//	//return false, errors.New("unlock duration too large")
	//	return common.Hash{}, err
	//} else {
	//	d = time.Duration(*duration) * time.Second
	//}
	//ks := s.b.GetKeyStore()
	//err = ks.TimedUnlock(accounts.Account{Address: addr}, password, d)
	err = unlockKS(s.b, addr, password, duration)
	if err != nil {
		//return nil, err
		return common.Hash{}, errors.New("TimedUnlock Account err")
	}

	newsign := ptnjson.NewSignRawTransactionCmd(result, &srawinputs, &keys, ptnjson.String(ALL))
	signresult, _ := SignRawTransactionOld(newsign, getPubKeyFn, getSignFn, addr)

	fmt.Println(signresult)
	stx := new(modules.Transaction)

	sserializedTx, err := decodeHexStr(signresult.Hex)
	if err != nil {
		return common.Hash{}, err
	}

	if err := rlp.DecodeBytes(sserializedTx, stx); err != nil {
		return common.Hash{}, err
	}
	s_msgs := stx.TxMessages()
	if 0 == len(s_msgs) {
		log.Info("+++++++++++++++++++++++++++++++++++++++++invalid Tx++++++")
		return common.Hash{}, errors.New("Invalid Tx, message length is 0")
	}
	var outAmount uint64
	for _, msg := range s_msgs {
		payload, ok := msg.Payload.(*modules.PaymentPayload)
		if !ok {
			continue
		}

		for _, txout := range payload.Outputs {
			log.Info("+++++++++++++++++++++++++++++++++++++++++", "tx_outAmount", txout.Value, "outInfo", txout)
			outAmount += txout.Value
		}
	}
	log.Info("--------------------------send tx ----------------------------", "txOutAmount", outAmount)

	log.Debugf("Tx outpoint tx hash:%s", s_msgs[0].Payload.(*modules.PaymentPayload).Inputs[0].PreviousOutPoint.TxHash.String())
	return submitTransaction(s.b, stx)
}

func RandFromString(value string) (decimal.Decimal, error) {
	originalInput := value
	var intString string
	var exp int64

	// Check if number is using scientific notation
	eIndex := strings.IndexAny(value, "Ee")
	if eIndex != -1 {
		expInt, err := strconv.ParseInt(value[eIndex+1:], 10, 32)
		if err != nil {
			if e, ok := err.(*strconv.NumError); ok && e.Err == strconv.ErrRange {
				return decimal.Decimal{}, fmt.Errorf("can't convert %s to decimal: fractional part too long", value)
			}
			return decimal.Decimal{}, fmt.Errorf("can't convert %s to decimal: exponent is not numeric", value)
		}
		value = value[:eIndex]
		exp = expInt
	}

	parts := strings.Split(value, ".")
	if len(parts) == 1 {
		// There is no decimal point, we can just parse the original string as
		// an int
		intString = value
	} else if len(parts) == 2 {
		// strip the insignificant digits for more accurate comparisons.
		decimalPart := strings.TrimRight(parts[1], "0")
		intString = parts[0] + decimalPart
		expInt := -len(decimalPart)
		exp += int64(expInt)
	} else {
		return decimal.Decimal{}, fmt.Errorf("can't convert %s to decimal: too many .s", value)
	}

	dValue := new(big.Int)
	_, ok := dValue.SetString(intString, 10)
	if !ok {
		return decimal.Decimal{}, fmt.Errorf("can't convert %s to decimal", value)
	}

	if exp < math.MinInt32 || exp > math.MaxInt32 {
		// NOTE(vadim): I doubt a string could realistically be this long
		return decimal.Decimal{}, fmt.Errorf("can't convert %s to decimal: fractional part too long", originalInput)
	}
	rand.Seed(time.Now().UnixNano())

	input_number := decimal.NewFromBigInt(dValue, int32(exp))
	result := decimal.Decimal{}
	rand_number := decimal.Decimal{}
	r := rand.Int()
	rr := int64(r)
	rd := big.NewInt(rr)
	for {
		//r = rand.Int()
		//rd = big.NewInt(int64(r))

		rand_number = decimal.NewFromBigInt(rd, int32(exp))
		result = rand_number.Mod(input_number)
		if !result.IsZero() {
			break
		}
	}
	return result, nil
}

func (s *PrivateWalletAPI) TransferPtn(ctx context.Context, from string, to string,
	amount decimal.Decimal, fee decimal.Decimal, data *string, password *string, duration *Int) (common.Hash, error) {
	gasToken := dagconfig.DagConfig.GasToken
	return s.TransferToken(ctx, gasToken, from, to, amount, fee, data, password, duration)
}
func getRealAddress(ks *keystore.KeyStore, addr string) (common.Address, error) {
	toArray := strings.Split(addr, ":")
	to := toArray[0]
	if len(toArray) == 2 { //HD wallet address format
		toAccountIndex, err := strconv.Atoi(toArray[1])
		if err != nil {
			return common.Address{}, errors.New("invalid to address format")
		}
		toAddr, err := common.StringToAddress(toArray[0])
		if err != nil {
			return common.Address{}, errors.New("invalid to address format")
		}
		var toAccount accounts.Account
		if ks.IsUnlock(toAddr) {
			toAccount, err = ks.GetHdAccount(accounts.Account{Address: toAddr}, uint32(toAccountIndex))
		} else {
			return common.Address{}, errors.New("First, please unlock address :" + addr)
		}
		if err != nil {
			return common.Address{}, errors.New("GetHdAccountWithPassphrase error:" + err.Error())
		}
		return toAccount.Address, nil
	}
	return common.StringToAddress(to)
}

func (s *PrivateWalletAPI) TransferToken(ctx context.Context, asset string, from string, to string,
	amount decimal.Decimal, fee decimal.Decimal, data *string, pwd *string, duration *Int) (common.Hash, error) {
	password := ""
	if pwd != nil {
		password = *pwd
	}
	s.b.Lock()
	defer s.b.Unlock()
	//1. build payment tx
	start := time.Now()
	rawTx, usedUtxo, err := buildRawTransferTx(s.b, asset, from, to, amount, fee, password)
	if err != nil {
		log.Error("buildRawTransferTx error:" + err.Error())
		return common.Hash{}, err
	}
	log.Debugf("build raw tx spend:%v", time.Since(start))
	//2. append data payload
	if data != nil && len(*data) > 0 {
		textPayload := new(modules.DataPayload)
		textPayload.Reference = []byte(asset)
		textPayload.MainData = []byte(*data)
		rawTx.AddMessage(modules.NewMessage(modules.APP_DATA, textPayload))
	}
	//3. sign
	start = time.Now()
	//ks := s.b.GetKeyStore()
	err = signRawTransaction(s.b, rawTx, from, password, duration, 1, usedUtxo)
	if err != nil {
		data, _ := json.Marshal(rawTx)
		log.Debug(string(data))
		data, _ = json.Marshal(usedUtxo)
		log.Debugf("used utxo:%s", string(data))
		log.Error("signRawTransaction error:" + err.Error())
		return common.Hash{}, err
	}
	//save tx to memory dag
	//err = saveTransaction2mDag(rawTx)
	//if err != nil {
	//	log.Errorf("CcinvokeToken err:%s", err.Error())
	//	return common.Hash{}, err
	//}

	log.Debugf("sign raw tx spend:%v", time.Since(start))
	//4. send
	txHash, err := submitTransaction(s.b, rawTx)
	if err != nil {
		log.Error("submitTransaction error:" + err.Error())
		return common.Hash{}, err
	}
	return txHash, nil
}

//转移Token，并确认打包后返回
func (s *PrivateWalletAPI) TransferTokenSync(ctx context.Context, asset string, fromStr string, toStr string,
	amount decimal.Decimal, fee decimal.Decimal, data *string, pwd *string, timeout *Int) (*ptnjson.TxHashWithUnitInfoJson, error) {
	password := ""
	if pwd != nil {
		password = *pwd
	}
	start := time.Now()
	log.Infof("Received transfer token request from:%s, to:%s,amount:%s", fromStr, toStr, amount.String())
	s.lock.Lock()
	log.Infof("Wait for lock time spend:%s", time.Since(start).String())
	tx, usedUtxo, err := buildRawTransferTx(s.b, asset, fromStr, toStr, amount, fee, password)
	if err != nil {
		s.lock.Unlock()
		return nil, err
	}
	//2. append data payload
	if data != nil && len(*data) > 0 {
		textPayload := new(modules.DataPayload)
		textPayload.Reference = []byte(asset)
		textPayload.MainData = []byte(*data)
		tx.AddMessage(modules.NewMessage(modules.APP_DATA, textPayload))
	}
	//3. sign
	//ks := s.b.GetKeyStore()
	err = signRawTransaction(s.b, tx, fromStr, password, timeout, 1, usedUtxo)
	if err != nil {
		log.Errorf("TransferTokenSync error:%s", err.Error())
		s.lock.Unlock()
		return nil, err
	}
	log.Infof("TransferTokenSync generate tx[%s]", tx.Hash().String())

	_, err = submitTransaction(s.b, tx)
	s.lock.Unlock()
	log.Infof("Generate and send tx[%s] spend time:%s", tx.Hash().String(), time.Since(start).String())
	if err != nil {
		return nil, err
	}
	start2 := time.Now()
	headCh := make(chan modules.SaveUnitEvent, 10)
	defer close(headCh)
	headSub := s.b.Dag().SubscribeSaveUnitEvent(headCh)
	defer headSub.Unsubscribe()
	timer := time.NewTimer(20 * time.Second)
	for {
		select {
		case u := <-headCh:
			log.Infof("SubscribeSaveUnitEvent received unit:%s", u.Unit.DisplayId())
			for i, utx := range u.Unit.Transactions() {
				if utx.Hash() == tx.Hash() {
					txInfo := &ptnjson.TxHashWithUnitInfoJson{
						Timestamp:   time.Unix(u.Unit.Timestamp(), 0),
						UnitHash:    u.Unit.Hash().String(),
						UnitHeight:  u.Unit.NumberU64(),
						TxIndex:     uint64(i),
						TxHash:      tx.Hash().String(),
						RequestHash: tx.RequestHash().String(),
					}
					log.Infof("receive tx[%s] packed event, spend time:%s, total spend:%s",
						tx.Hash().String(), time.Since(start2).String(),
						time.Since(start).String())
					return txInfo, nil
				}
			}
		case <-timer.C:
			return nil, errors.New(fmt.Sprintf("get tx[%s] package status timeout", tx.Hash().String()))
			// Err() channel will be closed when unsubscribing.
		case err := <-headSub.Err():
			return nil, err
		}
	}
}

//构造手续费代付的交易并广播
func (s *PrivateWalletAPI) TransferToken2(ctx context.Context, asset string, fromStr string, toStr string,
	gasFromStr string, amount decimal.Decimal, fee decimal.Decimal, data *string,
	pwd *string, timeout *Int) (common.Hash, error) {
	password := ""
	if pwd != nil {
		password = *pwd
	}
	toArray := strings.Split(toStr, ":")
	to := toArray[0]
	ks := s.b.GetKeyStore()
	if len(toArray) == 2 { //HD wallet address format
		toAccountIndex, err := strconv.Atoi(toArray[1])
		if err != nil {
			return common.Hash{}, errors.New("invalid to address format")
		}
		toAddr, err := common.StringToAddress(toArray[0])
		if err != nil {
			return common.Hash{}, errors.New("invalid to address format")
		}
		var toAccount accounts.Account
		if ks.IsUnlock(toAddr) {
			toAccount, err = ks.GetHdAccount(accounts.Account{Address: toAddr}, uint32(toAccountIndex))

		} else {
			toAccount, err = ks.GetHdAccountWithPassphrase(accounts.Account{Address: toAddr}, password, uint32(toAccountIndex))

		}
		if err != nil {
			return common.Hash{}, errors.New("GetHdAccountWithPassphrase error:" + err.Error())
		}
		to = toAccount.Address.String()
	}
	fromArray := strings.Split(fromStr, ":")
	from := fromArray[0]
	if len(fromArray) == 2 {
		fromAccountIndex, err := strconv.Atoi(fromArray[1])
		if err != nil {
			return common.Hash{}, errors.New("invalid to address format")
		}
		fromAddr, err := common.StringToAddress(fromArray[0])
		if err != nil {
			return common.Hash{}, errors.New("invalid to address format")
		}
		var fromAccount accounts.Account
		if ks.IsUnlock(fromAddr) {
			fromAccount, err = ks.GetHdAccount(accounts.Account{Address: fromAddr}, uint32(fromAccountIndex))

		} else {
			fromAccount, err = ks.GetHdAccountWithPassphrase(accounts.Account{Address: fromAddr}, password, uint32(fromAccountIndex))

		}
		if err != nil {
			return common.Hash{}, errors.New("GetHdAccountWithPassphrase error:" + err.Error())
		}
		from = fromAccount.Address.String()
	}
	gasFromArray := strings.Split(gasFromStr, ":")
	gasFrom := gasFromArray[0]
	if len(gasFromArray) == 2 {
		gasFromAccountIndex, err := strconv.Atoi(gasFromArray[1])
		if err != nil {
			return common.Hash{}, errors.New("invalid to address format")
		}
		gasFromAddr, err := common.StringToAddress(gasFromArray[0])
		if err != nil {
			return common.Hash{}, errors.New("invalid to address format")
		}
		var fromAccount accounts.Account
		if ks.IsUnlock(gasFromAddr) {
			fromAccount, err = ks.GetHdAccount(accounts.Account{Address: gasFromAddr}, uint32(gasFromAccountIndex))

		} else {
			fromAccount, err = ks.GetHdAccountWithPassphrase(accounts.Account{Address: gasFromAddr}, password, uint32(gasFromAccountIndex))

		}
		if err != nil {
			return common.Hash{}, errors.New("GetHdAccountWithPassphrase error:" + err.Error())
		}
		gasFrom = fromAccount.Address.String()
	}
	rawTx, usedUtxo, err := s.buildRawTransferTx2(asset, from, to, gasFrom, amount, fee)
	if err != nil {
		return common.Hash{}, err
	}
	if data != nil && len(*data) > 0 {
		textPayload := new(modules.DataPayload)
		textPayload.Reference = []byte(asset)
		textPayload.MainData = []byte(*data)
		rawTx.AddMessage(modules.NewMessage(modules.APP_DATA, textPayload))
	}
	//lockscript
	getPubKeyFn := func(addr common.Address) ([]byte, error) {
		return ks.GetPublicKey(addr)
	}
	//sign tx
	getSignFn := func(addr common.Address, msg []byte) ([]byte, error) {
		return ks.SignMessage(addr, msg)
	}
	utxoLockScripts := make(map[modules.OutPoint][]byte)
	for _, utxo := range usedUtxo {
		utxoLockScripts[utxo.OutPoint] = utxo.PkScript
	}
	if len(fromArray) == 1 {
		fromAddr, err := common.StringToAddress(from)
		if err != nil {
			return common.Hash{}, err
		}
		err = unlockKS(s.b, fromAddr, password, timeout)
		if err != nil {
			return common.Hash{}, err
		}
	}
	//3.
	_, err = tokenengine.Instance.SignTxAllPaymentInput(rawTx, 1, utxoLockScripts, nil, getPubKeyFn, getSignFn)
	if err != nil {
		return common.Hash{}, err
	}
	txJson, _ := json.Marshal(rawTx)
	log.DebugDynamic(func() string { return "SignedTx:" + string(txJson) })
	//4.
	return submitTransaction(s.b, rawTx)
}

//构建转移给多个地址Token的交易
//addrAndAmountJson 为 Address:Amount的Map
func (s *PrivateWalletAPI) TransferToken2MultiAddr(asset string, from string, addrAndAmountJson string,
	fee decimal.Decimal, password string) (common.Hash, error) {
	addrAmount := make(map[string]decimal.Decimal)
	err := json.Unmarshal([]byte(addrAndAmountJson), &addrAmount)
	if err != nil {
		return common.Hash{}, err
	}
	s.b.Lock()
	defer s.b.Unlock()
	//1. build payment tx
	start := time.Now()
	rawTx, usedUtxo, err := buildRawTransferTx2(s.b, asset, from, addrAmount, fee, password)
	if err != nil {
		log.Error("buildRawTransferTx error:" + err.Error())
		return common.Hash{}, err
	}
	log.Debugf("build raw tx spend:%v", time.Since(start))

	//3. sign
	start = time.Now()
	err = signRawTransaction(s.b, rawTx, from, password, nil, 1, usedUtxo)
	if err != nil {
		data, _ := json.Marshal(rawTx)
		log.Debug(string(data))
		data, _ = json.Marshal(usedUtxo)
		log.Debugf("used utxo:%s", string(data))
		log.Error("signRawTransaction error:" + err.Error())
		return common.Hash{}, err
	}

	log.Debugf("sign raw tx spend:%v", time.Since(start))
	//4. send
	txHash, err := submitTransaction(s.b, rawTx)
	if err != nil {
		log.Error("submitTransaction error:" + err.Error())
		return common.Hash{}, err
	}
	return txHash, nil
}

func (s *PrivateWalletAPI) CreateTxWithOutFee(ctx context.Context, asset, fromStr, toStr string, amount decimal.Decimal, pwd *string, duration *Int) (ptnjson.SignRawTransactionResult, error) {
	password := ""
	if pwd != nil {
		password = *pwd
	}
	toArray := strings.Split(toStr, ":")
	to := toArray[0]
	ks := s.b.GetKeyStore()
	if len(toArray) == 2 { //HD wallet address format
		toAccountIndex, err := strconv.Atoi(toArray[1])
		if err != nil {
			return ptnjson.SignRawTransactionResult{}, errors.New("invalid to address format")
		}
		toAddr, err := common.StringToAddress(toArray[0])
		if err != nil {
			return ptnjson.SignRawTransactionResult{}, errors.New("invalid to address format")
		}
		var toAccount accounts.Account
		if ks.IsUnlock(toAddr) {
			toAccount, err = ks.GetHdAccount(accounts.Account{Address: toAddr}, uint32(toAccountIndex))

		} else {
			toAccount, err = ks.GetHdAccountWithPassphrase(accounts.Account{Address: toAddr}, password, uint32(toAccountIndex))

		}
		if err != nil {
			return ptnjson.SignRawTransactionResult{}, errors.New("GetHdAccountWithPassphrase error:" + err.Error())
		}
		to = toAccount.Address.String()
	}

	getPubKeyFn := func(addr common.Address) ([]byte, error) {
		//TODO use keystore
		ks := s.b.GetKeyStore()
		return ks.GetPublicKey(addr)
	}
	getSignFn := func(addr common.Address, msg []byte) ([]byte, error) {
		ks := s.b.GetKeyStore()
		return ks.SignMessage(addr, msg)
	}

	fromArray := strings.Split(fromStr, ":")
	from := fromArray[0]
	if len(fromArray) == 2 {
		fromAccountIndex, err := strconv.Atoi(fromArray[1])
		if err != nil {
			return ptnjson.SignRawTransactionResult{}, errors.New("invalid to address format")
		}
		fromAddr, err := common.StringToAddress(fromArray[0])
		if err != nil {
			return ptnjson.SignRawTransactionResult{}, errors.New("invalid to address format")
		}
		var fromAccount accounts.Account
		if ks.IsUnlock(fromAddr) {
			fromAccount, err = ks.GetHdAccount(accounts.Account{Address: fromAddr}, uint32(fromAccountIndex))
		} else {
			fromAccount, err = ks.GetHdAccountWithPassphrase(accounts.Account{Address: fromAddr}, password, uint32(fromAccountIndex))

		}
		if err != nil {
			return ptnjson.SignRawTransactionResult{}, errors.New("GetHdAccountWithPassphrase error:" + err.Error())
		}
		from = fromAccount.Address.String()
	}

	utxoLockScripts := make(map[modules.OutPoint][]byte)
	payload, err := s.buildRawTxWithoutFee(asset, from, to, amount)
	if err != nil {
		return ptnjson.SignRawTransactionResult{}, err
	}
	inpoint := modules.OutPoint{
		TxHash:       payload.Inputs[0].PreviousOutPoint.TxHash,
		OutIndex:     payload.Inputs[0].PreviousOutPoint.OutIndex,
		MessageIndex: payload.Inputs[0].PreviousOutPoint.MessageIndex,
	}
	uvu, eerr := s.b.GetUtxoEntry(&inpoint)
	if eerr != nil {
		log.Error(eerr.Error())
		return ptnjson.SignRawTransactionResult{}, err
	}
	PkScriptHex := trimx(uvu.PkScriptHex)
	utxoLockScripts[inpoint] = hexutil.MustDecode("0x" + PkScriptHex)

	newtx := modules.NewTransaction(make([]*modules.Message, 0))
	newtx.AddMessage(modules.NewMessage(modules.APP_PAYMENT, payload))
	//3.
	from_Addr, err := common.StringToAddress(from)
	if err != nil {
		return ptnjson.SignRawTransactionResult{}, err
	}
	err = unlockKS(s.b, from_Addr, password, duration)
	if err != nil {
		return ptnjson.SignRawTransactionResult{}, err
	}
	signErrs, err := tokenengine.Instance.SignTxAllPaymentInput(newtx, 0x41, utxoLockScripts, nil, getPubKeyFn, getSignFn)
	if err != nil {
		return ptnjson.SignRawTransactionResult{}, err
	}

	//log.Debugf("payload input outpoint:%s", pload.Input[0].PreviousOutPoint.TxHash.String())
	mtxbt, err := rlp.EncodeToBytes(newtx)
	if err != nil {
		return ptnjson.SignRawTransactionResult{}, err
	}
	signedHex := hex.EncodeToString(mtxbt)
	signErrors := make([]ptnjson.SignRawTransactionError, 0, len(signErrs))
	return ptnjson.SignRawTransactionResult{
		Hex:      signedHex,
		Txid:     newtx.Hash().String(),
		Complete: len(signErrors) == 0,
		Errors:   signErrors,
	}, err
}
func (s *PrivateWalletAPI) buildNoGasPoETx(addr string,
	mainData, extraData, reference string, password string) (*modules.Transaction, error) {
	textPayload := new(modules.DataPayload)
	textPayload.MainData = []byte(mainData)
	textPayload.ExtraData = []byte(extraData)
	textPayload.Reference = []byte(reference)
	tx := modules.NewTransaction([]*modules.Message{modules.NewMessage(modules.APP_DATA, textPayload)})
	address, err := common.StringToAddress(addr)
	if err != nil {
		return nil, err
	}
	return signRawNoGasTx(s.b, tx, address, password)
}
func (s *PrivateWalletAPI) CreateProofOfExistenceTxSync(addr string,
	mainData, extraData, reference string, password *string) (common.Hash, error) {
	txHash, err := s.CreateProofOfExistenceTx(addr, mainData, extraData, reference, password)
	if err != nil {
		return txHash, err
	}
	start := time.Now()
	headCh := make(chan modules.SaveUnitEvent, 10)
	defer close(headCh)
	headSub := s.b.Dag().SubscribeSaveUnitEvent(headCh)
	defer headSub.Unsubscribe()
	timer := time.NewTimer(20 * time.Second)
	for {
		select {
		case u := <-headCh:
			log.Debugf("SubscribeSaveUnitEvent received unit:%s", u.Unit.DisplayId())
			for _, utx := range u.Unit.Transactions() {
				if utx.Hash() == txHash {
					log.Debugf("receive tx[%s] packed event, spend time:%s",
						txHash.String(), time.Since(start).String())
					return txHash, nil
				}
			}
		case <-timer.C:
			return common.Hash{}, errors.New(fmt.Sprintf("get tx[%s] package status timeout", txHash.String()))
			// Err() channel will be closed when unsubscribing.
		case err := <-headSub.Err():
			return common.Hash{}, err
		}
	}

}
func (s *PrivateWalletAPI) CreateProofOfExistenceTx(addr string,
	mainData, extraData, reference string, pwd *string) (common.Hash, error) {
	password := ""
	if pwd != nil {
		password = *pwd
	}
	if !s.b.EnableGasFee() {
		tx, err := s.buildNoGasPoETx(addr, mainData, extraData, reference, password)
		if err != nil {
			return common.Hash{}, err
		}
		return submitTransaction(s.b, tx)
	}
	//无GasFee时，不需要加锁，可以并发创建
	s.b.Lock()
	defer s.b.Unlock()
	gasToken := dagconfig.DagConfig.GasToken
	ptn1 := decimal.New(1, -2)
	rawTx, usedUtxo, err := buildRawTransferTx(s.b, gasToken, addr, addr, decimal.New(0, 0), ptn1, password)
	if err != nil {
		return common.Hash{}, err
	}

	textPayload := new(modules.DataPayload)
	textPayload.MainData = []byte(mainData)
	textPayload.ExtraData = []byte(extraData)
	textPayload.Reference = []byte(reference)
	rawTx.AddMessage(modules.NewMessage(modules.APP_DATA, textPayload))

	//lockscript
	getPubKeyFn := func(addr common.Address) ([]byte, error) {
		//TODO use keystore
		ks := s.b.GetKeyStore()
		return ks.GetPublicKey(addr)
	}
	//sign tx
	getSignFn := func(addr common.Address, msg []byte) ([]byte, error) {
		ks := s.b.GetKeyStore()
		return ks.SignMessage(addr, msg)
	}
	utxoLockScripts := make(map[modules.OutPoint][]byte)
	for _, utxo := range usedUtxo {
		utxoLockScripts[utxo.OutPoint] = utxo.PkScript
	}
	fromAddr, err := common.StringToAddress(addr)
	if err != nil {
		return common.Hash{}, err
	}
	err = unlockKS(s.b, fromAddr, password, nil)
	if err != nil {
		return common.Hash{}, err
	}
	//3.
	_, err = tokenengine.Instance.SignTxAllPaymentInput(rawTx, 1, utxoLockScripts, nil, getPubKeyFn, getSignFn)
	if err != nil {
		return common.Hash{}, err
	}
	log.Infof("Create ProofOfExistence Tx[%s] for main data:%s", rawTx.Hash().String(), mainData)
	log.DebugDynamic(func() string { return "SignedTx:" + rawTx.String() })
	//4.
	return submitTransaction(s.b, rawTx)
}

//创建一笔溯源交易，调用721合约
func (s *PrivateWalletAPI) CreateTraceability(ctx context.Context, addr, uid, symbol, mainData, extraData, reference string) (common.Hash, error) {
	contractAddr, _ := common.StringToAddress("PCGTta3M4t3yXu8uRgkKvaWd2d8DRijspoq")
	str := "[{\"TokenID\":\"" + uid + "\",\"MetaData\":\"\"}]"
	gasToken := dagconfig.DagConfig.GasToken
	ptn1 := decimal.New(1, -1)
	rawTx, usedUtxo, err := buildRawTransferTx(s.b, gasToken, addr, addr, decimal.New(0, 0), ptn1, "")
	if err != nil {
		return common.Hash{}, err
	}

	textPayload := new(modules.DataPayload)
	textPayload.MainData = []byte(mainData)
	textPayload.ExtraData = []byte(extraData)
	textPayload.Reference = []byte(reference)

	args := make([][]byte, 4)
	args[0] = []byte("supplyToken")
	args[1] = []byte(symbol)
	args[2] = []byte("1")
	args[3] = []byte(str)
	ccinvokePayload := new(modules.ContractInvokeRequestPayload)
	ccinvokePayload.Args = args
	ccinvokePayload.ContractId = contractAddr.Bytes()
	ccinvokePayload.Timeout = 0

	rawTx.AddMessage(modules.NewMessage(modules.APP_DATA, textPayload))
	rawTx.AddMessage(modules.NewMessage(modules.APP_CONTRACT_INVOKE_REQUEST, ccinvokePayload))
	//lockscript
	getPubKeyFn := func(addr common.Address) ([]byte, error) {
		//TODO use keystore
		ks := s.b.GetKeyStore()
		return ks.GetPublicKey(addr)
	}
	//sign tx
	getSignFn := func(addr common.Address, msg []byte) ([]byte, error) {
		ks := s.b.GetKeyStore()
		return ks.SignMessage(addr, msg)
	}
	utxoLockScripts := make(map[modules.OutPoint][]byte)
	for _, utxo := range usedUtxo {
		utxoLockScripts[utxo.OutPoint] = utxo.PkScript
	}
	fromAddr, err := common.StringToAddress(addr)
	if err != nil {
		return common.Hash{}, err
	}
	err = unlockKS(s.b, fromAddr, "1", nil)
	if err != nil {
		return common.Hash{}, err
	}
	//3.
	_, err = tokenengine.Instance.SignTxAllPaymentInput(rawTx, 1, utxoLockScripts, nil, getPubKeyFn, getSignFn)
	if err != nil {
		return common.Hash{}, err
	}

	log.DebugDynamic(func() string { return "SignedTx:" + rawTx.String() })
	//4.
	return submitTransaction(s.b, rawTx)
}

//
//根据maindata信息 查询存证结果  filehash  --> maindata
//func (s *PublicWalletAPI) getFileInfo(filehash string) ([]*ptnjson.ProofOfExistenceJson, error) {
//	files, err := s.b.GetFileInfo(filehash)
//	if err != nil {
//		return nil, err
//	}
//	result := []*ptnjson.ProofOfExistenceJson{}
//	for _, file := range files {
//		tx, err := s.b.GetTxByHash(file.Txid)
//		if err != nil {
//			return nil, err
//		}
//		poe := ptnjson.ConvertTx2ProofOfExistence(tx)
//		result = append(result, poe)
//	}
//	return result, nil
//}

func (s *PublicWalletAPI) getProofOfExistencesByMaindata(maindata string) ([]*ptnjson.ProofOfExistenceJson, error) {
	files, err := s.b.GetFileInfo(maindata)
	if err != nil {
		return nil, err
	}
	result := []*ptnjson.ProofOfExistenceJson{}
	for _, file := range files {
		tx, err := s.b.GetTxByHash(file.Txid)
		if err != nil {
			return nil, err
		}
		poe := ptnjson.ConvertTx2ProofOfExistence(tx)
		result = append(result, poe)
	}
	return result, nil
}

//根据交易哈希 查询存证结果
func (s *PublicWalletAPI) GetFileInfoByTxid(ctx context.Context, txid common.Hash) (*ptnjson.ProofOfExistenceJson, error) {
	tx, err := s.b.GetTxByHash(txid)
	if err != nil {
		return nil, err
	}
	return ptnjson.ConvertTx2ProofOfExistence(tx), err
}

//GetProofOfExistencesByMaindata替代GetFileInfoByFileHash
func (s *PublicWalletAPI) GetFileInfoByFileHash(ctx context.Context, maindata string) ([]*ptnjson.ProofOfExistenceJson, error) {
	result, err := s.getProofOfExistencesByMaindata(maindata)
	return result, err
}

func (s *PublicWalletAPI) GetProofOfExistencesByMaindata(ctx context.Context, maindata string) ([]*ptnjson.ProofOfExistenceJson, error) {
	result, err := s.getProofOfExistencesByMaindata(maindata)
	return result, err
}

func (s *PublicWalletAPI) GetOneTokenInfo(ctx context.Context, symbol string) (string, error) {
	GlobalStateContractId := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	fmt.Println(modules.GlobalPrefix + strings.ToUpper(symbol))
	result, _, err := s.b.GetContractState(GlobalStateContractId, modules.GlobalPrefix+strings.ToUpper(symbol))
	return string(result), err
}

func (s *PublicWalletAPI) GetAllTokenInfo(ctx context.Context) (string, error) {
	GlobalStateContractId := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	result, err := s.b.GetContractStatesByPrefix(GlobalStateContractId, modules.GlobalPrefix)
	if nil == result || nil != err {
		return "There is no PRC20 and PRC721 Token Yet", nil
	}
	var all []modules.GlobalTokenInfo
	for key, val := range result {
		fmt.Println(key, val.Value)
		var oneToken modules.GlobalTokenInfo
		err := json.Unmarshal(val.Value, &oneToken)
		if nil == err {
			all = append(all, oneToken)
		}
	}
	allToken, err := json.Marshal(all)
	if nil != err {
		return "There is no PRC20 and PRC721 Token Yet", nil
	}

	return string(allToken), err
}

func (s *PublicWalletAPI) GetProofOfExistencesByRef(ctx context.Context, reference string) ([]*ptnjson.ProofOfExistenceJson, error) {
	return s.b.QueryProofOfExistenceByReference(reference)
}

func (s *PublicWalletAPI) GetProofOfExistencesByAsset(ctx context.Context, asset string) ([]*ptnjson.ProofOfExistenceJson, error) {
	return s.b.GetAssetExistence(asset)
}

//好像某个UTXO是被那个交易花费的
func (s *PublicWalletAPI) GetStxo(txid string, msgIdx int, outIdx int) (*ptnjson.StxoJson, error) {
	outpoint := modules.NewOutPoint(common.HexToHash(txid), uint32(msgIdx), uint32(outIdx))
	return s.b.GetStxoEntry(outpoint)
}
func (s *PublicWalletAPI) GetUtxo(txid string, msgIdx int, outIdx int) (*ptnjson.UtxoJson, error) {
	outpoint := modules.NewOutPoint(common.HexToHash(txid), uint32(msgIdx), uint32(outIdx))
	return s.b.GetUtxoEntry(outpoint)
}

// 压测交易池，批量添加交易
func (s *PublicWalletAPI) AddBatchTxs(ctx context.Context, path string) (int, error) {
	tt := time.Now()
	sign_txs, err := readTxs(path)
	if err != nil {
		return 0, err
	}
	log.Infof("add_batch_txs ,path: %s, read txs spent time: %s", path, time.Since(tt))
	ttt := time.Now()
	txs := make([]*modules.Transaction, 0)
	for _, str := range sign_txs {
		serializedTx, err := hex.DecodeString(str)
		if err != nil {
			return 0, errors.New("Decode Signedtx is invalid")
		}

		tx := modules.NewTransaction(make([]*modules.Message, 0))
		if err := rlp.DecodeBytes(serializedTx, tx); err != nil {
			return 0, errors.New("encodedTx decode is invalid")
		}
		txs = append(txs, tx)
	}
	go submitTxs(ctx, s.b, txs)
	log.Infof("add_batch_txs ,send txs spent time: %s", time.Since(ttt))
	return len(txs), nil
}
func readTxs(path string) ([]string, error) {
	if !common.IsExisted(path) {
		return nil, fmt.Errorf("file:%s, is not exist.", path)
	}
	txs := make([]string, 0)
	f, err := os.Open(path)
	if err != nil {
		log.Infof("open file failed, err:%s", err.Error())
		return nil, err
	}
	defer f.Close()

	rd := bufio.NewReader(f)
	for {
		line, err1 := rd.ReadString('\n')
		if err1 != nil || io.EOF == err1 {
			if err1 != io.EOF {
				err = err1
			}
			break
		}
		line = strings.Replace(line, "\r\n", "", -1)
		txs = append(txs, line)
	}
	if len(txs) <= 5000 {
		return txs, err
	} else {
		return txs[:5000], err
	}

}

//将UTXO碎片聚集成整的UTXO
func (s *PrivateWalletAPI) AggregateUtxo(ctx context.Context,
	address string, fee decimal.Decimal) ([]common.Hash, error) {
	ptn := dagconfig.DagConfig.GasToken
	addr, err := common.StringToAddress(address)
	if err != nil {
		return nil, err
	}

	//err = s.unlockKS(addr, password, nil)
	//if err != nil {
	//	return nil, err
	//}
	utxos, err := s.b.GetAddrRawUtxos(address)
	if err != nil {
		return nil, err
	}
	var batchInputLen = 500 //以1000个Input为一个Tx，构造转账交易给自己
	var payment *modules.PaymentPayload = nil
	var inputCount = 0
	var inputAmtSum = uint64(0)
	var asset *modules.Asset = nil
	result := []common.Hash{}
	for outpoint, utxo := range utxos {
		if utxo.Asset.String() != ptn { //按Asset对UTXO进行过滤
			continue
		}
		asset = utxo.Asset
		if payment == nil { //一个新的Tx
			payment = &modules.PaymentPayload{}
		}
		//附加一个Input
		out := outpoint
		payment.AddTxIn(modules.NewTxIn(&out, nil))
		inputAmtSum += utxo.Amount
		inputCount++
		if inputCount%batchInputLen == 0 { //满足构造Tx的数量了
			feeAmount := ptnjson.JsonAmt2AssetAmt(asset, fee)
			if inputAmtSum <= feeAmount {
				return result, nil
			}
			tx, err := s.generateTx(payment, addr, inputAmtSum, feeAmount, asset)
			if err != nil {
				return nil, err
			}
			log.Infof("Try to send aggregate UTXO tx[%s]", tx.Hash().String())
			err = s.b.SendTx(tx)
			inputAmtSum = 0
			payment = nil
			if err != nil {
				return nil, err
			}
			result = append(result, tx.Hash())
		}
	}
	if inputAmtSum != 0 {
		feeAmount := ptnjson.JsonAmt2AssetAmt(asset, fee)
		if inputAmtSum <= feeAmount {
			return result, nil
		}
		tx, err := s.generateTx(payment, addr, inputAmtSum, feeAmount, asset)
		if err != nil {
			return nil, err
		}
		log.Infof("Try to send aggregate UTXO tx[%s]", tx.Hash().String())
		err = s.b.SendTx(tx)
		if err != nil {
			return nil, err
		}
		result = append(result, tx.Hash())
	}
	return result, nil
}
func (s *PrivateWalletAPI) generateTx(payment *modules.PaymentPayload, address common.Address,
	inputAmtSum, fee uint64, asset *modules.Asset) (*modules.Transaction, error) {
	out := modules.NewTxOut(inputAmtSum-fee, tokenengine.Instance.GenerateLockScript(address), asset)
	payment.AddTxOut(out)

	tx := modules.NewTransaction([]*modules.Message{modules.NewMessage(modules.APP_PAYMENT, payment)})
	//Sign
	utxoLockScripts := make(map[modules.OutPoint][]byte)
	lockScript := tokenengine.Instance.GenerateLockScript(address)
	for _, input := range payment.Inputs {
		utxoLockScripts[*input.PreviousOutPoint] = lockScript
	}
	getPubKeyFn := func(addr common.Address) ([]byte, error) {
		//TODO use keystore
		ks := s.b.GetKeyStore()
		return ks.GetPublicKey(addr)
	}
	//sign tx
	getSignFn := func(addr common.Address, msg []byte) ([]byte, error) {
		ks := s.b.GetKeyStore()
		return ks.SignMessage(addr, msg)
	}
	signErrs, err := tokenengine.Instance.SignTxAllPaymentInput(tx, 1, utxoLockScripts, nil, getPubKeyFn, getSignFn)
	if err != nil {
		log.Errorf("%v", signErrs)
		//TODO
		return nil, err
	}
	return tx, nil
}

//构造2笔Message，Msg0的FromAddress是商家，Msg1的FromAddress是用户
func (s *PrivateWalletAPI) buildRawTransferTx2(tokenId, from, to, gasFrom string, amount, gasFee decimal.Decimal) (*modules.Transaction, []*modules.UtxoWithOutPoint, error) {
	//参数检查
	tokenAsset, err := modules.StringToAsset(tokenId)
	if err != nil {
		return nil, nil, err
	}
	if !gasFee.IsPositive() && s.b.EnableGasFee() {
		return nil, nil, fmt.Errorf("fee is ZERO ")
	}
	//
	fromAddr, err := common.StringToAddress(from)
	if err != nil {
		fmt.Println(err.Error())
		return nil, nil, err
	}
	toAddr, err := common.StringToAddress(to)
	if err != nil {
		fmt.Println(err.Error())
		return nil, nil, err
	}
	gasAddr, err := common.StringToAddress(gasFrom)
	if err != nil {
		fmt.Println(err.Error())
		return nil, nil, err
	}
	ptnAmount := uint64(0)
	ptn := dagconfig.DagConfig.GasToken
	if tokenId == ptn {
		ptnAmount = ptnjson.Ptn2Dao(amount)
	}

	//构造转移PTN的Message0
	//FromAddress是商家
	dbUtxos, err := s.b.GetAddrRawUtxos(gasFrom)
	if err != nil {
		return nil, nil, fmt.Errorf("GetAddrRawUtxos utxo err")
	}
	poolTxs, _ := s.b.GetUnpackedTxsByAddr(gasFrom)

	utxosPTN, err := SelectUtxoFromDagAndPool(dbUtxos, make(map[common.Hash]common.Hash), poolTxs, gasFrom, ptn)
	if err != nil {
		return nil, nil, fmt.Errorf("SelectUtxoFromDagAndPool utxo err")
	}
	feeAmount := ptnjson.Ptn2Dao(gasFee)
	pay1, usedUtxo1, err := createPayment(gasAddr, toAddr, ptnAmount, feeAmount, utxosPTN)
	if err != nil {
		return nil, nil, err
	}

	tx := modules.NewTransaction([]*modules.Message{modules.NewMessage(modules.APP_PAYMENT, pay1)})
	if tokenId == ptn {
		return tx, usedUtxo1, nil
	}
	//构造转移Token的Message1
	//FromAddress是用户
	//构造转移Token的Message1
	dbUtxos2, err := s.b.GetAddrRawUtxos(from)
	if err != nil {
		return nil, nil, fmt.Errorf("GetAddrRawUtxos utxo err")
	}
	poolTxs2, _ := s.b.GetUnpackedTxsByAddr(from)
	utxosToken, err := SelectUtxoFromDagAndPool(dbUtxos2, make(map[common.Hash]common.Hash), poolTxs2, from, tokenId)
	if err != nil {
		return nil, nil, fmt.Errorf("SelectUtxoFromDagAndPool token utxo err")
	}
	tokenAmount := ptnjson.JsonAmt2AssetAmt(tokenAsset, amount)
	pay2, usedUtxo2, err := createPayment(fromAddr, toAddr, tokenAmount, 0, utxosToken)
	if err != nil {
		return nil, nil, err
	}
	tx.AddMessage(modules.NewMessage(modules.APP_PAYMENT, pay2))
	//for _, u := range usedUtxo2 {
	usedUtxo1 = append(usedUtxo1, usedUtxo2...)
	//}
	return tx, usedUtxo1, nil
}

//构造1笔Message，Msg0的FromAddress是用户
func (s *PrivateWalletAPI) buildRawTxWithoutFee(tokenId, from, to string, amount decimal.Decimal) (*modules.PaymentPayload, error) {
	//参数检查
	tokenAsset, err := modules.StringToAsset(tokenId)
	if err != nil {
		return nil, err
	}
	//
	fromAddr, err := common.StringToAddress(from)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	toAddr, err := common.StringToAddress(to)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	//构造转移Token的Message1
	//FromAddress是用户
	//构造转移Token的Message1
	dbUtxos2, err := s.b.GetAddrRawUtxos(from)
	if err != nil {
		return nil, fmt.Errorf("GetAddrRawUtxos utxo err")
	}
	poolTxs2, _ := s.b.GetUnpackedTxsByAddr(from)
	utxosToken, err := SelectUtxoFromDagAndPool(dbUtxos2, make(map[common.Hash]common.Hash), poolTxs2, from, tokenId)
	if err != nil {
		return nil, fmt.Errorf("SelectUtxoFromDagAndPool token utxo err")
	}
	tokenAmount := ptnjson.JsonAmt2AssetAmt(tokenAsset, amount)
	pay2, _, err := createPayment(fromAddr, toAddr, tokenAmount, 0, utxosToken)
	if err != nil {
		return nil, err
	}

	//mtx := modules.NewTransaction([]*modules.Message{modules.NewMessage(modules.APP_PAYMENT, pay2)})

	//mtxbt, err := rlp.EncodeToBytes(mtx)
	//if err != nil {
	//	return "", err
	//}
	//log.Debugf("payload input outpoint:%s", pload.Input[0].PreviousOutPoint.TxHash.String())
	//mtxHex := hex.EncodeToString(mtxbt)
	return pay2, nil
}

func (s *PrivateWalletAPI) SignAndFeeTransaction(ctx context.Context, params string, gasFrom string, gasFee decimal.Decimal, Extra string, pwd *string, duration *Int) (ptnjson.SignRawTransactionResult, error) {
	//transaction inputs
	password := ""
	if pwd != nil {
		password = *pwd
	}
	if params == "" {
		return ptnjson.SignRawTransactionResult{}, errors.New("Params is empty")
	}
	upper_type := strings.ToUpper("all")
	if upper_type != ALL && upper_type != NONE && upper_type != SINGLE {
		return ptnjson.SignRawTransactionResult{}, errors.New("Hashtype is error,error type")
	}
	serializedTx, err := decodeHexStr(params)
	if err != nil {
		return ptnjson.SignRawTransactionResult{}, errors.New("Params is invalid")
	}

	tokentx := modules.NewTransaction(make([]*modules.Message, 0))

	if err := rlp.DecodeBytes(serializedTx, &tokentx); err != nil {
		return ptnjson.SignRawTransactionResult{}, errors.New("Params decode is invalid")
	}

	//构造转移PTN的Message0
	//FromAddress是商家
	dbUtxos, err := s.b.GetAddrRawUtxos(gasFrom)
	if err != nil {
		return ptnjson.SignRawTransactionResult{}, errors.New("GetAddrRawUtxos utxo err")
	}
	poolTxs, _ := s.b.GetUnpackedTxsByAddr(gasFrom)

	ptn := dagconfig.DagConfig.GasToken

	utxosPTN, err := SelectUtxoFromDagAndPool(dbUtxos, make(map[common.Hash]common.Hash), poolTxs, gasFrom, ptn)
	if err != nil {
		return ptnjson.SignRawTransactionResult{}, errors.New("SelectUtxoFromDagAndPool utxo err")
	}
	gasAddr, err := common.StringToAddress(gasFrom)
	if err != nil {
		fmt.Println(err.Error())
		return ptnjson.SignRawTransactionResult{}, errors.New("gasFrom addr err")
	}

	ptnAmount := uint64(0)
	feeAmount := ptnjson.Ptn2Dao(gasFee)
	gaspayload, _, err := createPayment(gasAddr, gasAddr, ptnAmount, feeAmount, utxosPTN)
	if err != nil {
		return ptnjson.SignRawTransactionResult{}, errors.New("fee createPayment  err")
	}

	//tx.AddMessage(modules.NewMessage(modules.APP_PAYMENT, pay1))

	gasfeetx := modules.NewTransaction(make([]*modules.Message, 0))
	gasfeetx.AddMessage(modules.NewMessage(modules.APP_PAYMENT, gaspayload))
	//for _, msg := range tx.TxMessages() {
	//	newtx.AddMessage(msg)
	//}
	getPubKeyFn := func(addr common.Address) ([]byte, error) {
		//TODO use keystore
		ks := s.b.GetKeyStore()
		return ks.GetPublicKey(addr)
	}
	getSignFn := func(addr common.Address, msg []byte) ([]byte, error) {
		ks := s.b.GetKeyStore()
		return ks.SignMessage(addr, msg)
	}

	utxoLockScripts := make(map[modules.OutPoint][]byte)

	inpoint := modules.OutPoint{
		TxHash:       gaspayload.Inputs[0].PreviousOutPoint.TxHash,
		OutIndex:     gaspayload.Inputs[0].PreviousOutPoint.OutIndex,
		MessageIndex: gaspayload.Inputs[0].PreviousOutPoint.MessageIndex,
	}
	uvu, eerr := s.b.GetUtxoEntry(&inpoint)
	if eerr != nil {
		log.Error(eerr.Error())
		return ptnjson.SignRawTransactionResult{}, err
	}
	PkScriptHex := trimx(uvu.PkScriptHex)
	utxoLockScripts[inpoint] = hexutil.MustDecode("0x" + PkScriptHex)

	err = unlockKS(s.b, gasAddr, password, duration)
	if err != nil {
		newErr := errors.New("get addr by outpoint get err:" + err.Error())
		log.Error(newErr.Error())
		return ptnjson.SignRawTransactionResult{}, err
	}

	if Extra != "" {
		textPayload := new(modules.DataPayload)
		textPayload.Reference = []byte(Extra)
		textPayload.MainData = []byte(Extra)
		gasfeetx.AddMessage(modules.NewMessage(modules.APP_DATA, textPayload))
	}

	for _, msg := range tokentx.TxMessages() {
		gasfeetx.AddMessage(msg)
	}
	//3.
	signErrs, err := tokenengine.Instance.SignTxAllPaymentInput(gasfeetx, 1, utxoLockScripts, nil, getPubKeyFn, getSignFn)
	if err != nil {
		return ptnjson.SignRawTransactionResult{}, err
	}

	//resulttx := modules.NewTransaction(make([]*modules.Message, 0))
	//resulttx.AddMessage(modules.NewMessage(modules.APP_PAYMENT, pay1))

	mtxbt, err := rlp.EncodeToBytes(gasfeetx)
	if err != nil {
		return ptnjson.SignRawTransactionResult{}, err
	}
	signedHex := hex.EncodeToString(mtxbt)
	signErrors := make([]ptnjson.SignRawTransactionError, 0, len(signErrs))
	return ptnjson.SignRawTransactionResult{
		Hex:      signedHex,
		Txid:     gasfeetx.Hash().String(),
		Complete: len(signErrors) == 0,
		Errors:   signErrors,
	}, err
}
