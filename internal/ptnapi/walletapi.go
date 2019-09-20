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
	"time"

	"github.com/ethereum/go-ethereum/rlp"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/hexutil"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/math"
	"github.com/palletone/go-palletone/core"
	"github.com/palletone/go-palletone/core/accounts"
	"github.com/palletone/go-palletone/core/certficate"
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
	b Backend
}

func NewPublicWalletAPI(b Backend) *PublicWalletAPI {
	return &PublicWalletAPI{b}
}
func NewPrivateWalletAPI(b Backend) *PrivateWalletAPI {
	return &PrivateWalletAPI{b}
}
func (s *PublicWalletAPI) CreateRawTransaction(ctx context.Context, from string, to string, amount, fee decimal.Decimal) (string, error) {

	//realNet := &chaincfg.MainNetParams
	var LockTime int64
	//LockTime = 0

	amounts := []ptnjson.AddressAmt{}
	if from == "" {
		return "", fmt.Errorf("sender address is empty")
	}
	if to == "" {
		return "", fmt.Errorf("receiver address is empty")
	}
	_, ferr := common.StringToAddress(from)
	if ferr != nil {
		return "", fmt.Errorf("sender address is invalid")
	}
	_, terr := common.StringToAddress(to)
	if terr != nil {
		return "", fmt.Errorf("receiver address is invalid")
	}

	amounts = append(amounts, ptnjson.AddressAmt{Address: to, Amount: amount})
	if len(amounts) == 0 || !amount.IsPositive() {
		return "", fmt.Errorf("amounts is invalid")
	}
	dbUtxos, err := s.b.GetAddrRawUtxos(from)
	if err != nil {
		return "", err
	}

	ptn := dagconfig.DagConfig.GasToken

	poolTxs, _ := s.b.GetPoolTxsByAddr(from)
	//if len(poolTxs) == 0 {
	//      return "", fmt.Errorf("GetPoolTxsByAddr Err")
	//}
	allutxos, err := SelectUtxoFromDagAndPool(dbUtxos, poolTxs, from, ptn)
	if err != nil {
		return "", fmt.Errorf("SelectUtxoFromDagAndPool utxo err")
	}
	limitdao, _ := decimal.NewFromString("0.0001")
	if !fee.GreaterThanOrEqual(limitdao) {
		return "", fmt.Errorf("fee cannot less than 100000 Dao ")
	}
	daoAmount := ptnjson.Ptn2Dao(amount.Add(fee))
	if daoAmount <= 100000000 {
		return "", fmt.Errorf("amount cannot less than 1 dao ")
	}
	utxos, _ := convertUtxoMap2Utxos(allutxos)
	taken_utxo, change, err := core.Select_utxo_Greedy(utxos, daoAmount)
	if err != nil {
		return "", fmt.Errorf("Select_utxo_Greedy utxo err")
	}

	inputs := []ptnjson.TransactionInput{}
	var input ptnjson.TransactionInput
	for _, u := range taken_utxo {
		utxo := u.(*modules.UtxoWithOutPoint)
		input.Txid = utxo.TxHash.String()
		input.MessageIndex = utxo.MessageIndex
		input.Vout = utxo.OutIndex
		inputs = append(inputs, input)
	}

	if change > 0 {
		amounts = append(amounts, ptnjson.AddressAmt{Address: from, Amount: ptnjson.Dao2Ptn(change)})
	}

	arg := ptnjson.NewCreateRawTransactionCmd(inputs, amounts, &LockTime)
	result, _ := CreateRawTransaction(arg)

	return result, nil
}
func (s *PrivateWalletAPI) buildRawTransferTx(tokenId, from, to string, amount, gasFee decimal.Decimal) (*modules.Transaction, []*modules.UtxoWithOutPoint, error) {
	//参数检查
	tokenAsset, err := modules.StringToAsset(tokenId)
	if err != nil {
		return nil, nil, err
	}
	if !gasFee.IsPositive() {
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
	ptnAmount := uint64(0)
	ptn := dagconfig.DagConfig.GasToken
	if tokenId == ptn {
		ptnAmount = ptnjson.Ptn2Dao(amount)
	}
	tx := &modules.Transaction{}

	//构造转移PTN的Message0
	dbUtxos, err := s.b.GetAddrRawUtxos(from)
	if err != nil {
		return nil, nil, fmt.Errorf("GetAddrRawUtxos utxo err")
	}
	poolTxs, _ := s.b.GetPoolTxsByAddr(from)
	//if len(poolTxs) == 0 {
	//       return nil, nil, fmt.Errorf("GetPoolTxsByAddr utxo err")
	//}

	utxosPTN, err := SelectUtxoFromDagAndPool(dbUtxos, poolTxs, from, ptn)
	if err != nil {
		return nil, nil, fmt.Errorf("SelectUtxoFromDagAndPool utxo err")
	}
	feeAmount := ptnjson.Ptn2Dao(gasFee)
	pay1, usedUtxo1, err := createPayment(fromAddr, toAddr, ptnAmount, feeAmount, utxosPTN)
	if err != nil {
		return nil, nil, err
	}
	tx.TxMessages = append(tx.TxMessages, modules.NewMessage(modules.APP_PAYMENT, pay1))
	if tokenId == ptn {
		return tx, usedUtxo1, nil
	}
	//构造转移Token的Message1
	utxosToken, err := SelectUtxoFromDagAndPool(dbUtxos, poolTxs, from, tokenId)
	if err != nil {
		return nil, nil, fmt.Errorf("SelectUtxoFromDagAndPool token utxo err")
	}
	tokenAmount := ptnjson.JsonAmt2AssetAmt(tokenAsset, amount)
	pay2, usedUtxo2, err := createPayment(fromAddr, toAddr, tokenAmount, 0, utxosToken)
	if err != nil {
		return nil, nil, err
	}
	tx.TxMessages = append(tx.TxMessages, modules.NewMessage(modules.APP_PAYMENT, pay2))
	//for _, u := range usedUtxo2 {
	usedUtxo1 = append(usedUtxo1, usedUtxo2...)
	//}
	return tx, usedUtxo1, nil
}
func createPayment(fromAddr, toAddr common.Address, amountToken uint64, feePTN uint64,
	utxosPTN map[modules.OutPoint]*modules.Utxo) (*modules.PaymentPayload, []*modules.UtxoWithOutPoint, error) {
	if len(utxosPTN) == 0 {
		return nil, nil, fmt.Errorf("No PTN Utxo or No Token Utxo")
	}

	//PTN
	utxoPTNView, asset := convertUtxoMap2Utxos(utxosPTN)

	utxosPTNTaken, change, err := core.Select_utxo_Greedy(utxoPTNView, amountToken+feePTN)
	if err != nil {
		return nil, nil, fmt.Errorf("createPayment Select_utxo_Greedy utxo err")
	}
	usedUtxo := []*modules.UtxoWithOutPoint{}
	//ptn payment
	payPTN := &modules.PaymentPayload{}
	//ptn inputs
	for _, u := range utxosPTNTaken {
		utxo := u.(*modules.UtxoWithOutPoint)
		usedUtxo = append(usedUtxo, utxo)
		prevOut := &utxo.OutPoint // modules.NewOutPoint(txHash, utxo.MessageIndex, utxo.OutIndex)
		txInput := modules.NewTxIn(prevOut, []byte{})
		payPTN.AddTxIn(txInput)
	}
	//ptn outputs
	if amountToken > 0 {
		payPTN.AddTxOut(modules.NewTxOut(amountToken, tokenengine.Instance.GenerateLockScript(toAddr), asset))
	}
	if change > 0 {
		payPTN.AddTxOut(modules.NewTxOut(change, tokenengine.Instance.GenerateLockScript(fromAddr), asset))
	}
	//
	////Token
	//utxoTokenView := ConvertUtxoMap2Utxos(utxosToken)
	//utxosTkTaken, change, err := core.Select_utxo_Greedy(utxoTokenView, amountToken)
	//if err != nil {
	//	return nil, nil, fmt.Errorf("Select token utxo err")
	//}
	////token payment
	//payToken := &modules.PaymentPayload{}
	////ptn inputs
	//for _, u := range utxosTkTaken {
	//	utxo := u.(*modules.UtxoWithOutPoint)
	//	usedUtxo = append(usedUtxo, utxo)
	//	prevOut := &utxo.OutPoint //  modules.NewOutPoint(txHash, utxo.MessageIndex, utxo.OutIndex)
	//	txInput := modules.NewTxIn(prevOut, []byte{})
	//	payToken.AddTxIn(txInput)
	//}
	////token outputs
	//payToken.AddTxOut(modules.NewTxOut(amountToken, tokenengine.GenerateLockScript(toAddr), asset))
	//if change > 0 {
	//	payToken.AddTxOut(modules.NewTxOut(change, tokenengine.GenerateLockScript(fromAddr), asset))
	//}
	//
	////tx
	//	//tx.TxMessages = append(tx.TxMessages, modules.NewMessage(modules.APP_PAYMENT, payToken))
	return payPTN, usedUtxo, nil
}

func WalletCreateTransaction(c *ptnjson.CreateRawTransactionCmd) (string, error) {

	// Validate the locktime, if given.
	if c.LockTime != nil &&
		(*c.LockTime < 0 || *c.LockTime > int64(MaxTxInSequenceNum)) {
		return "", &ptnjson.RPCError{
			Code:    ptnjson.ErrRPCInvalidParameter,
			Message: "Locktime out of range",
		}
	}
	// Add all transaction inputs to a new transaction after performing
	// some validity checks.
	//先构造PaymentPayload结构，再组装成Transaction结构
	pload := new(modules.PaymentPayload)
	//var inputjson []walletjson.InputJson
	for _, input := range c.Inputs {
		txHash := common.HexToHash(input.Txid)

		//inputjson = append(inputjson, walletjson.InputJson{TxHash: input.Txid, MessageIndex: input.MessageIndex, OutIndex: input.Vout, HashForSign: "", Signature: ""})
		prevOut := modules.NewOutPoint(txHash, input.MessageIndex, input.Vout)
		txInput := modules.NewTxIn(prevOut, []byte{})
		pload.AddTxIn(txInput)
	}
	//var OutputJson []walletjson.OutputJson
	// Add all transaction outputs to the transaction after performing
	//	// some validity checks.
	//	//only support mainnet
	//	var params *chaincfg.Params
	var ppscript []byte
	for _, addramt := range c.Amounts {
		encodedAddr := addramt.Address
		ptnAmt := addramt.Amount
		// amount := ptnjson.Ptn2Dao(ptnAmt)
		// Ensure amount is in the valid range for monetary amounts.
		// if amount <= 0 /*|| amount > ptnjson.MaxDao*/ {
		// 	return "", &ptnjson.RPCError{
		// 		Code:    ptnjson.ErrRPCType,
		// 		Message: "Invalid amount",
		// 	}
		// }
		addr, err := common.StringToAddress(encodedAddr)
		if err != nil {
			return "", &ptnjson.RPCError{
				Code:    ptnjson.ErrRPCInvalidAddressOrKey,
				Message: "Invalid address or key",
			}
		}
		switch addr.GetType() {
		case common.PublicKeyHash:
		case common.ScriptHash:
		case common.ContractHash:
			//case *ptnjson.AddressPubKeyHash:
			//case *ptnjson.AddressScriptHash:
		default:
			return "", &ptnjson.RPCError{
				Code:    ptnjson.ErrRPCInvalidAddressOrKey,
				Message: "Invalid address or key",
			}
		}
		// Create a new script which pays to the provided address.
		pkScript := tokenengine.Instance.GenerateLockScript(addr)
		ppscript = pkScript
		// Convert the amount to satoshi.
		dao := ptnjson.Ptn2Dao(ptnAmt)
		//if err != nil {
		//	context := "Failed to convert amount"
		//	return "", internalRPCError(err.Error(), context)
		//}
		assetId := dagconfig.DagConfig.GetGasToken()
		txOut := modules.NewTxOut(dao, pkScript, assetId.ToAsset())
		pload.AddTxOut(txOut)
		//OutputJson = append(OutputJson, walletjson.OutputJson{Amount: uint64(dao), Asset: assetId.String(), ToAddress: addr.String()})
	}
	//	// Set the Locktime, if given.
	if c.LockTime != nil {
		pload.LockTime = uint32(*c.LockTime)
	}
	//	// Return the serialized and hex-encoded transaction.  Note that this
	//	// is intentionally not directly returning because the first return
	//	// value is a string and it would result in returning an empty string to
	//	// the client instead of nothing (nil) in the case of an error.

	mtx := &modules.Transaction{
		TxMessages: make([]*modules.Message, 0),
	}
	mtx.TxMessages = append(mtx.TxMessages, modules.NewMessage(modules.APP_PAYMENT, pload))
	//mtx.TxHash = mtx.Hash()
	//sign mtx
	mtxtmp := mtx
	for msgindex, msg := range mtxtmp.TxMessages {
		payload, ok := msg.Payload.(*modules.PaymentPayload)
		if !ok {
			continue
		}
		for inputindex := range payload.Inputs {
			hashforsign, err := tokenengine.Instance.CalcSignatureHash(mtxtmp, tokenengine.SigHashAll, msgindex, inputindex, ppscript)
			if err != nil {
				return "", err
			}
			payloadtmp := mtx.TxMessages[msgindex].Payload.(*modules.PaymentPayload)
			payloadtmp.Inputs[inputindex].SignatureScript = hashforsign
		}
	}

	mtxbt, err := rlp.EncodeToBytes(mtx)
	if err != nil {
		return "", err
	}
	//log.Debugf("payload input outpoint:%s", pload.Input[0].PreviousOutPoint.TxHash.String())
	mtxHex := hex.EncodeToString(mtxbt)
	return mtxHex, nil
	//return string(bytetxjson), nil
}
func (s *PrivateWalletAPI) SignRawTransaction(ctx context.Context, params string, hashtype string, password string, duration *uint64) (ptnjson.SignRawTransactionResult, error) {

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

	tx := &modules.Transaction{
		TxMessages: make([]*modules.Message, 0),
	}
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
	for _, msg := range tx.TxMessages {
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
	const max = uint64(time.Duration(math.MaxInt64) / time.Second)
	var d time.Duration
	if duration == nil {
		d = 300 * time.Second
	} else if *duration > max {
		return ptnjson.SignRawTransactionResult{}, errors.New("unlock duration too large")
	} else {
		d = time.Duration(*duration) * time.Second
	}
	ks := s.b.GetKeyStore()
	err = ks.TimedUnlock(accounts.Account{Address: addr}, password, d)
	if err != nil {
		newErr := errors.New("get addr by outpoint get err:" + err.Error())
		log.Error(newErr.Error())
		return ptnjson.SignRawTransactionResult{}, newErr
	}

	newsign := ptnjson.NewSignRawTransactionCmd(params, &srawinputs, &keys, ptnjson.String(hashtype))
	result, err := SignRawTransaction(newsign, getPubKeyFn, getSignFn, addr)
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

	tx := &modules.Transaction{
		TxMessages: make([]*modules.Message, 0),
	}
	if err := rlp.DecodeBytes(serializedTx, tx); err != nil {
		return common.Hash{}, errors.New("encodedTx decode is invalid")
	}

	if 0 == len(tx.TxMessages) {
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
	return submitTransaction(ctx, s.b, tx)
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
	tx := &modules.Transaction{
		TxMessages: make([]*modules.Message, 0),
	}
	err = json.Unmarshal(btxjson, tx)
	if err != nil {
		return common.Hash{}, errors.New("Json Unmarshal To Tx is invalid")
	}

	if 0 == len(tx.TxMessages) {
		return common.Hash{}, errors.New("Invalid Tx, message length is 0")
	}
	var outAmount uint64
	var outpoint_txhash common.Hash
	for _, msg := range tx.TxMessages {
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
	return submitTransaction(ctx, s.b, tx)
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
	if 0 == len(tx.TxMessages) {
		return common.Hash{}, errors.New("Invalid Tx, message length is 0")
	}
	var outAmount uint64
	for _, msg := range tx.TxMessages {
		payload, ok := msg.Payload.(*modules.PaymentPayload)
		if !ok {
			continue
		}

		for _, txout := range payload.Outputs {
			outAmount += txout.Value
		}
	}
	return submitTransaction(ctx, s.b, tx)
}

func (s *PublicWalletAPI) CreateProofTransaction(ctx context.Context, params string, password string) (common.Hash, error) {

	var proofTransactionGenParams ptnjson.ProofTransactionGenParams
	err := json.Unmarshal([]byte(params), &proofTransactionGenParams)
	if err != nil {
		return common.Hash{}, err
	}
	var amount decimal.Decimal
	//realNet := &chaincfg.MainNetParams
	amounts := []ptnjson.AddressAmt{}
	for _, outOne := range proofTransactionGenParams.Outputs {
		if len(outOne.Address) == 0 || outOne.Amount.LessThanOrEqual(decimal.New(0, 0)) {
			continue
		}
		amounts = append(amounts, ptnjson.AddressAmt{Address: outOne.Address, Amount: outOne.Amount})
		amount = amount.Add(outOne.Amount)
	}
	if len(amounts) == 0 || !amount.IsPositive() {
		return common.Hash{}, err
	}

	dbUtxos, err := s.b.GetAddrRawUtxos(proofTransactionGenParams.From)
	if err != nil {
		return common.Hash{}, err
	}
	poolTxs, _ := s.b.GetPoolTxsByAddr(proofTransactionGenParams.From)
	//if len(poolTxs) == 0 {
	//return common.Hash{}, fmt.Errorf("Select utxo err")
	//} // end of pooltx is not nil
	utxos, err := SelectUtxoFromDagAndPool(dbUtxos, poolTxs, proofTransactionGenParams.From, dagconfig.DagConfig.GasToken)
	if err != nil {
		return common.Hash{}, fmt.Errorf("SelectUtxoFromDagAndPool err")
	}
	//dagOutpoint := []modules.OutPoint{}
	//ptn := dagconfig.DagConfig.GasToken
	//for _, json := range utxoJsons {
	//	//utxos = append(utxos, &json)
	//	if json.Asset == ptn {
	//		utxos = append(utxos, &ptnjson.UtxoJson{TxHash: json.TxHash, MessageIndex: json.MessageIndex, OutIndex: json.OutIndex, Amount: json.Amount, Asset: json.Asset, PkScriptHex: json.PkScriptHex, PkScriptString: json.PkScriptString, LockTime: json.LockTime})
	//		dagOutpoint = append(dagOutpoint, modules.OutPoint{TxHash: common.HexToHash(json.TxHash), MessageIndex: json.MessageIndex, OutIndex: json.OutIndex})
	//	}
	//}

	fee := proofTransactionGenParams.Fee
	if !fee.IsPositive() {
		return common.Hash{}, fmt.Errorf("fee is ZERO ")
	}
	daoAmount := ptnjson.Ptn2Dao(amount.Add(fee))
	utxoList, _ := convertUtxoMap2Utxos(utxos)
	taken_utxo, change, err := core.Select_utxo_Greedy(utxoList, daoAmount)
	if err != nil {
		return common.Hash{}, fmt.Errorf("CreateProofTransaction Select_utxo_Greedy utxo err")
	}

	inputs := []ptnjson.TransactionInput{}
	var input ptnjson.TransactionInput
	for _, u := range taken_utxo {
		utxo := u.(*modules.UtxoWithOutPoint)
		input.Txid = utxo.TxHash.String()
		input.MessageIndex = utxo.MessageIndex
		input.Vout = utxo.OutIndex
		inputs = append(inputs, input)
	}

	if change > 0 {
		amounts = append(amounts, ptnjson.AddressAmt{Address: proofTransactionGenParams.From, Amount: ptnjson.Dao2Ptn(change)})
	}

	if len(inputs) == 0 {
		return common.Hash{}, nil
	}
	arg := ptnjson.NewCreateProofTransactionCmd(inputs, amounts, &proofTransactionGenParams.Locktime, proofTransactionGenParams.Proof, proofTransactionGenParams.Extra)
	result, _ := WalletCreateProofTransaction(arg)
	//transaction inputs
	serializedTx, err := decodeHexStr(result)
	if err != nil {
		return common.Hash{}, err
	}

	tx := &modules.Transaction{
		TxMessages: make([]*modules.Message, 0),
	}
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
		//account, _ := MakeAddress(ks, addr.String())
		//privKey, _ := ks.DumpPrivateKey(account, "1")
		return ks.SignMessage(addr, msg)
		//return crypto.Sign(hash, privKey)
	}
	var srawinputs []ptnjson.RawTxInput
	var addr common.Address
	var keys []string
	from, _ := common.StringToAddress(proofTransactionGenParams.From)
	PkScript := tokenengine.Instance.GenerateLockScript(from)
	PkScriptHex := hexutil.Encode(PkScript)
	for _, msg := range tx.TxMessages {
		payload, ok := msg.Payload.(*modules.PaymentPayload)
		if !ok {
			continue
		}
		for _, txin := range payload.Inputs {
			TxHash := txin.PreviousOutPoint.TxHash.String()
			OutIndex := txin.PreviousOutPoint.OutIndex
			MessageIndex := txin.PreviousOutPoint.MessageIndex
			input := ptnjson.RawTxInput{Txid: TxHash, Vout: OutIndex, MessageIndex: MessageIndex, ScriptPubKey: PkScriptHex, RedeemScript: ""}
			srawinputs = append(srawinputs, input)
			addr, err = tokenengine.Instance.GetAddressFromScript(hexutil.MustDecode(PkScriptHex))
			if err != nil {
				return common.Hash{}, err
			}
		}
	}
	//const max = uint64(time.Duration(math.MaxInt64) / time.Second)
	d := 300 * time.Second

	ks := s.b.GetKeyStore()
	err = ks.TimedUnlock(accounts.Account{Address: addr}, password, d)
	if err != nil {
		return common.Hash{}, errors.New("get addr by outpoint is err")
	}

	newsign := ptnjson.NewSignRawTransactionCmd(result, &srawinputs, &keys, ptnjson.String(ALL))
	signresult, _ := SignRawTransaction(newsign, getPubKeyFn, getSignFn, addr)

	stx := new(modules.Transaction)

	sserializedTx, err := decodeHexStr(signresult.Hex)
	if err != nil {
		return common.Hash{}, err
	}

	if err := rlp.DecodeBytes(sserializedTx, stx); err != nil {
		return common.Hash{}, err
	}
	if 0 == len(stx.TxMessages) {
		log.Info("+++++++++++++++++++++++++++++++++++++++++invalid Tx++++++")
		return common.Hash{}, errors.New("Invalid Tx, message length is 0")
	}
	var outAmount uint64
	for _, msg := range stx.TxMessages {
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

	log.Debugf("Tx outpoint tx hash:%s", stx.TxMessages[0].Payload.(*modules.PaymentPayload).Inputs[0].PreviousOutPoint.TxHash.String())
	return submitTransaction(ctx, s.b, stx)
}
func WalletCreateProofTransaction( /*s *rpcServer*/ c *ptnjson.CreateProofTransactionCmd) (string, error) {

	// Validate the locktime, if given.
	if c.LockTime != nil &&
		(*c.LockTime < 0 || *c.LockTime > int64(MaxTxInSequenceNum)) {
		return "", &ptnjson.RPCError{
			Code:    ptnjson.ErrRPCInvalidParameter,
			Message: "Locktime out of range",
		}
	}
	textPayload := new(modules.DataPayload)
	textPayload.MainData = []byte(c.Proof)
	textPayload.ExtraData = []byte(c.Extra)
	// Add all transaction inputs to a new transaction after performing
	// some validity checks.
	//先构造PaymentPayload结构，再组装成Transaction结构
	pload := new(modules.PaymentPayload)
	//var inputjson []walletjson.InputJson
	for _, input := range c.Inputs {
		txHash := common.HexToHash(input.Txid)

		//inputjson = append(inputjson, walletjson.InputJson{TxHash: input.Txid, MessageIndex: input.MessageIndex, OutIndex: input.Vout, HashForSign: "", Signature: ""})
		prevOut := modules.NewOutPoint(txHash, input.MessageIndex, input.Vout)
		txInput := modules.NewTxIn(prevOut, []byte{})
		pload.AddTxIn(txInput)
	}
	//var OutputJson []walletjson.OutputJson
	// Add all transaction outputs to the transaction after performing
	// some validity checks.
	// only support mainnet
	// var params *chaincfg.Params
	for _, addramt := range c.Amounts {
		encodedAddr := addramt.Address
		ptnAmt := addramt.Amount
		amount := ptnjson.Ptn2Dao(ptnAmt)
		// Ensure amount is in the valid range for monetary amounts.
		if amount <= 0 /*|| amount > ptnjson.MaxDao*/ {
			return "", &ptnjson.RPCError{
				Code:    ptnjson.ErrRPCType,
				Message: "Invalid amount",
			}
		}
		addr, err := common.StringToAddress(encodedAddr)
		if err != nil {
			return "", &ptnjson.RPCError{
				Code:    ptnjson.ErrRPCInvalidAddressOrKey,
				Message: "Invalid address or key",
			}
		}
		switch addr.GetType() {
		case common.PublicKeyHash:
		case common.ScriptHash:
		case common.ContractHash:
			//case *ptnjson.AddressPubKeyHash:
			//case *ptnjson.AddressScriptHash:
		default:
			return "", &ptnjson.RPCError{
				Code:    ptnjson.ErrRPCInvalidAddressOrKey,
				Message: "Invalid address or key",
			}
		}
		// Create a new script which pays to the provided address.
		pkScript := tokenengine.Instance.GenerateLockScript(addr)
		// Convert the amount to satoshi.
		dao := ptnjson.Ptn2Dao(ptnAmt)
		//if err != nil {
		//	context := "Failed to convert amount"
		//	return "", internalRPCError(err.Error(), context)
		//}
		assetId := dagconfig.DagConfig.GetGasToken()
		txOut := modules.NewTxOut(dao, pkScript, assetId.ToAsset())
		pload.AddTxOut(txOut)
		//OutputJson = append(OutputJson, walletjson.OutputJson{Amount: dao, Asset: assetId.String(), ToAddress: addr.String()})
	}
	// Set the Locktime, if given.
	if c.LockTime != nil {
		pload.LockTime = uint32(*c.LockTime)
	}
	//	// Return the serialized and hex-encoded transaction.  Note that this
	//	// is intentionally not directly returning because the first return
	//	// value is a string and it would result in returning an empty string to
	//	// the client instead of nothing (nil) in the case of an error.

	mtx := &modules.Transaction{
		TxMessages: make([]*modules.Message, 0),
	}
	mtx.TxMessages = append(mtx.TxMessages, modules.NewMessage(modules.APP_PAYMENT, pload))

	mtx.TxMessages = append(mtx.TxMessages, modules.NewMessage(modules.APP_DATA, textPayload))
	mtxbt, err := rlp.EncodeToBytes(mtx)
	if err != nil {
		return "", err
	}
	//log.Debugf("payload input outpoint:%s", pload.Input[0].PreviousOutPoint.TxHash.String())
	mtxHex := hex.EncodeToString(mtxbt)
	return mtxHex, nil
	//return string(bytetxproofjson), nil
}

func (s *PublicWalletAPI) GetAddrUtxos(ctx context.Context, addr string) (string, error) {

	items, err := s.b.GetAddrUtxos(addr)
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
func (s *PublicWalletAPI) GetBalance(ctx context.Context, address string) (map[string]decimal.Decimal, error) {
	utxos, err := s.b.GetAddrUtxos(address)
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
func (s *PublicWalletAPI) GetBalance2(ctx context.Context, address string) (*walletjson.StableUnstable, error) {
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

//获得某地址的通证流水
func (s *PublicWalletAPI) GetAddrTokenFlow(ctx context.Context, addr string, token string) ([]*ptnjson.TokenFlowJson, error) {
	result, err := s.b.GetAddrTokenFlow(addr, token)
	return result, err
}

//sign rawtranscation
//create raw transction
func (s *PublicWalletAPI) GetPtnTestCoin(ctx context.Context, from string, to string, amount, password string, duration *uint64) (common.Hash, error) {
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

	utxoJsons, err := s.b.GetAddrUtxos(from)
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

	tx := &modules.Transaction{
		TxMessages: make([]*modules.Message, 0),
	}
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
	for _, msg := range tx.TxMessages {
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
	const max = uint64(time.Duration(math.MaxInt64) / time.Second)
	var d time.Duration
	if duration == nil {
		d = 300 * time.Second
	} else if *duration > max {
		//return false, errors.New("unlock duration too large")
		return common.Hash{}, err
	} else {
		d = time.Duration(*duration) * time.Second
	}
	ks := s.b.GetKeyStore()
	err = ks.TimedUnlock(accounts.Account{Address: addr}, password, d)
	if err != nil {
		//return nil, err
		return common.Hash{}, errors.New("get addr by outpoint is err")
	}

	newsign := ptnjson.NewSignRawTransactionCmd(result, &srawinputs, &keys, ptnjson.String(ALL))
	signresult, _ := SignRawTransaction(newsign, getPubKeyFn, getSignFn, addr)

	fmt.Println(signresult)
	stx := new(modules.Transaction)

	sserializedTx, err := decodeHexStr(signresult.Hex)
	if err != nil {
		return common.Hash{}, err
	}

	if err := rlp.DecodeBytes(sserializedTx, stx); err != nil {
		return common.Hash{}, err
	}
	if 0 == len(stx.TxMessages) {
		log.Info("+++++++++++++++++++++++++++++++++++++++++invalid Tx++++++")
		return common.Hash{}, errors.New("Invalid Tx, message length is 0")
	}
	var outAmount uint64
	for _, msg := range stx.TxMessages {
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

	log.Debugf("Tx outpoint tx hash:%s", stx.TxMessages[0].Payload.(*modules.PaymentPayload).Inputs[0].PreviousOutPoint.TxHash.String())
	return submitTransaction(ctx, s.b, stx)
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

func (s *PrivateWalletAPI) unlockKS(addr common.Address, password string, duration *uint64) error {
	const max = uint64(time.Duration(math.MaxInt64) / time.Second)
	var d time.Duration
	if duration == nil {
		d = 300 * time.Second
	} else if *duration > max {
		return errors.New("unlock duration too large")
	} else {
		d = time.Duration(*duration) * time.Second
	}
	ks := s.b.GetKeyStore()
	err := ks.TimedUnlock(accounts.Account{Address: addr}, password, d)
	if err != nil {
		return errors.New("get addr by outpoint is err")
	}
	return nil
}

func (s *PrivateWalletAPI) TransferPtn(ctx context.Context, from string, to string,
	amount decimal.Decimal, fee decimal.Decimal, Extra string, password string, duration *uint64) (common.Hash, error) {
	gasToken := dagconfig.DagConfig.GasToken
	return s.TransferToken(ctx, gasToken, from, to, amount, fee, Extra, password, duration)
}

func (s *PrivateWalletAPI) TransferToken(ctx context.Context, asset string, from string, to string,
	amount decimal.Decimal, fee decimal.Decimal, Extra string, password string, duration *uint64) (common.Hash, error) {
	rawTx, usedUtxo, err := s.buildRawTransferTx(asset, from, to, amount, fee)
	if err != nil {
		return common.Hash{}, err
	}
	if Extra != "" {
		textPayload := new(modules.DataPayload)
		textPayload.Reference = []byte(asset)
		textPayload.MainData = []byte(Extra)
		rawTx.TxMessages = append(rawTx.TxMessages, modules.NewMessage(modules.APP_DATA, textPayload))
	}
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
	fromAddr, err := common.StringToAddress(from)
	if err != nil {
		return common.Hash{}, err
	}
	err = s.unlockKS(fromAddr, password, duration)
	if err != nil {
		return common.Hash{}, err
	}
	//3.
	_, err = tokenengine.Instance.SignTxAllPaymentInput(rawTx, 1, utxoLockScripts, nil, getPubKeyFn, getSignFn)
	if err != nil {
		return common.Hash{}, err
	}
	txJson, _ := json.Marshal(rawTx)
	log.DebugDynamic(func() string { return "SignedTx:" + string(txJson) })
	//4.
	return submitTransaction(ctx, s.b, rawTx)
}

func (s *PrivateWalletAPI) CreateProofOfExistenceTx(ctx context.Context, addr string,
	mainData, extraData, reference string, password string) (common.Hash, error) {
	gasToken := dagconfig.DagConfig.GasToken
	ptn1 := decimal.New(1, 0)
	rawTx, usedUtxo, err := s.buildRawTransferTx(gasToken, addr, addr, decimal.New(0, 0), ptn1)
	if err != nil {
		return common.Hash{}, err
	}

	textPayload := new(modules.DataPayload)
	textPayload.MainData = []byte(mainData)
	textPayload.ExtraData = []byte(extraData)
	textPayload.Reference = []byte(reference)
	rawTx.TxMessages = append(rawTx.TxMessages, modules.NewMessage(modules.APP_DATA, textPayload))

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
	err = s.unlockKS(fromAddr, password, nil)
	if err != nil {
		return common.Hash{}, err
	}
	//3.
	_, err = tokenengine.Instance.SignTxAllPaymentInput(rawTx, 1, utxoLockScripts, nil, getPubKeyFn, getSignFn)
	if err != nil {
		return common.Hash{}, err
	}
	txJson, _ := json.Marshal(rawTx)
	log.DebugDynamic(func() string { return "SignedTx:" + string(txJson) })
	//4.
	return submitTransaction(ctx, s.b, rawTx)
}

//创建一笔溯源交易，调用721合约
func (s *PrivateWalletAPI) CreateTraceability(ctx context.Context, addr, uid, symbol, mainData, extraData, reference string) (common.Hash, error) {
	password := "1"
	caddr := "PCGTta3M4t3yXu8uRgkKvaWd2d8DRijspoq"
	contractAddr, _ := common.StringToAddress(caddr)
	str := "[{\"TokenID\":\"" + uid + "\",\"MetaData\":\"\"}]"
	gasToken := dagconfig.DagConfig.GasToken
	ptn1 := decimal.New(1, 0)
	rawTx, usedUtxo, err := s.buildRawTransferTx(gasToken, addr, addr, decimal.New(0, 0), ptn1)
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

	rawTx.TxMessages = append(rawTx.TxMessages, modules.NewMessage(modules.APP_DATA, textPayload))
	rawTx.TxMessages = append(rawTx.TxMessages, modules.NewMessage(modules.APP_CONTRACT_INVOKE_REQUEST, ccinvokePayload))
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
	err = s.unlockKS(fromAddr, password, nil)
	if err != nil {
		return common.Hash{}, err
	}
	//3.
	_, err = tokenengine.Instance.SignTxAllPaymentInput(rawTx, 1, utxoLockScripts, nil, getPubKeyFn, getSignFn)
	if err != nil {
		return common.Hash{}, err
	}
	txJson, _ := json.Marshal(rawTx)
	log.DebugDynamic(func() string { return "SignedTx:" + string(txJson) })
	//4.
	return submitTransaction(ctx, s.b, rawTx)
}

func (s *PublicWalletAPI) getFileInfo(filehash string) (string, error) {
	//get fileinfos
	files, err := s.b.GetFileInfo(filehash)
	if err != nil {
		return "null", err
	}
	var timestamp int64
	gets := []walletjson.GetFileInfos{}
	for _, file := range files {
		get := walletjson.GetFileInfos{}
		get.ParentsHash = file.ParentsHash.String()
		get.FileHash = file.MainData
		get.ExtraData = file.ExtraData
		get.Reference = file.Reference
		timestamp = int64(file.Timestamp)
		tm := time.Unix(timestamp, 0)
		get.Timestamp = tm.String()
		get.TransactionHash = file.Txid.String()
		get.UintHeight = file.UintHeight
		get.UnitHash = file.UnitHash.String()
		gets = append(gets, get)
	}

	result := walletjson.ConvertGetFileInfos2Json(gets)

	return result, nil
}

func (s *PublicWalletAPI) GetFileInfoByTxid(ctx context.Context, txid string) (string, error) {
	if len(txid) == 66 {
		result, err := s.getFileInfo(txid)
		return result, err
	}
	err := errors.New("Parameter input error")
	return "", err
}

func (s *PublicWalletAPI) GetFileInfoByFileHash(ctx context.Context, filehash string) (string, error) {
	result, err := s.getFileInfo(filehash)
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

//affiliation  gptn.mediator1
func (s *PrivateWalletAPI) GenCert(ctx context.Context, caAddress, userAddress, passwd, name, data, roleType, affiliation string) (*ContractDeployRsp, error) {
	contractAddr := "PCGTta3M4t3yXu8uRgkKvaWd2d8DRv2vsEk"
	// 参数检查
	_, err := common.StringToAddress(userAddress)
	if err != nil {
		return nil, fmt.Errorf("invalid account address: %v", userAddress)
	}
	caAddr, err := common.StringToAddress(caAddress)
	if err != nil {
		return nil, fmt.Errorf("invalid account address: %v", userAddress)
	}
	cAddr, err := common.StringToAddress(contractAddr)
	if err != nil {
		return nil, fmt.Errorf("invalid account address: %v", userAddress)
	}

	ks := s.b.GetKeyStore()
	account, err := MakeAddress(ks, userAddress)
	if err != nil {
		return nil, err
	}

	//导出私钥 用于证书的生成
	privKey, err := ks.DumpPrivateKey(account, passwd)
	if err != nil {
		return nil, err
	}
	err = s.unlockKS(caAddr, "1", nil)
	if err != nil {
		return nil, err
	}

	ca := certficate.CertINfo{}
	ca.Address = userAddress
	ca.Name = name
	ca.Data = data
	ca.Type = roleType
	ca.Affiliation = affiliation
	ca.Key = privKey
	//生成证书 获取证书byte
	certBytes, err := certficate.GenCert(ca)
	log.Infof("GenCert Success! CertBytes[%s]", certBytes)
	if err != nil {
		return nil, err
	}
	//调用系统合约 将证书byte存入到数字身份系统合约中
	args := make([][]byte, 3)
	args[0] = []byte("addMemberCert")
	args[1] = []byte(userAddress)
	args[2] = certBytes

	reqId, err := s.b.ContractInvokeReqTx(caAddr, caAddr, 10000, 10000, nil, cAddr, args, 0)
	if err != nil {
		return nil, err
	}
	log.Infof("GenCert reqId[%s]", hex.EncodeToString(reqId[:]))
	rsp := &ContractDeployRsp{
		ReqId:      hex.EncodeToString(reqId[:]),
		ContractId: contractAddr,
	}

	return rsp, nil
}

//吊销证书  将crl存入到数字身份系统合约中
func (s *PrivateWalletAPI) RevokeCert(ctx context.Context, caAddress, passwd, userAddress string) (*ContractDeployRsp, error) {
	contractAddr := "PCGTta3M4t3yXu8uRgkKvaWd2d8DRv2vsEk"
	// 参数检查
	_, err := common.StringToAddress(userAddress)
	if err != nil {
		return nil, fmt.Errorf("invalid account address: %v", userAddress)
	}
	caAddr, err := common.StringToAddress(caAddress)
	if err != nil {
		return nil, fmt.Errorf("invalid account address: %v", caAddress)
	}
	cAddr, err := common.StringToAddress(contractAddr)
	if err != nil {
		return nil, fmt.Errorf("invalid account address: %v", contractAddr)
	}
	//导出ca私钥用于吊销用户证书
	ks := s.b.GetKeyStore()
	account, err := MakeAddress(ks, caAddress)
	if err != nil {
		return nil, err
	}

	privKey, err := ks.DumpPrivateKey(account, passwd)
	if err != nil {
		return nil, err
	}
	err = s.unlockKS(caAddr, passwd, nil)
	if err != nil {
		return nil, err
	}
	reason := "PalletOne system administrator revokes certificate!"

	ca := certficate.CertINfo{}
	ca.Address = userAddress
	ca.Key = privKey
	crlByte, err := certficate.RevokeCert(ca, reason)
	if err != nil {
		return nil, err
	}
	log.Infof("RevokeCert Success!  CrlByte[%s]", crlByte)

	//调用系统合约 将CrlByte存入到数字身份系统合约中
	args := make([][]byte, 2)
	args[0] = []byte("addCRL")
	args[1] = crlByte

	reqId, err := s.b.ContractInvokeReqTx(caAddr, caAddr, 10000, 10000, nil, cAddr, args, 0)
	if err != nil {
		return nil, err
	}
	log.Infof("RevokeCert reqId[%s]", hex.EncodeToString(reqId[:]))
	rsp := &ContractDeployRsp{
		ReqId:      hex.EncodeToString(reqId[:]),
		ContractId: contractAddr,
	}
	return rsp, nil
}

//好像某个UTXO是被那个交易花费的
func (s *PublicWalletAPI) GetStxo(ctx context.Context, txid string, msgIdx int, outIdx int) (*ptnjson.StxoJson, error) {
	outpoint := modules.NewOutPoint(common.HexToHash(txid), uint32(msgIdx), uint32(outIdx))
	return s.b.GetStxoEntry(outpoint)
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

		tx := &modules.Transaction{
			TxMessages: make([]*modules.Message, 0),
		}
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
			err = s.b.SendTx(ctx, tx)
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
		err = s.b.SendTx(ctx, tx)
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
	tx := &modules.Transaction{}
	tx.AddMessage(modules.NewMessage(modules.APP_PAYMENT, payment))
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
