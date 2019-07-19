package ptnapi

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

	"time"

	"math/big"
	"math/rand"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/rlp"
	"github.com/palletone/go-palletone/common"
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
	LockTime = 0

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

	amounts = append(amounts, ptnjson.AddressAmt{to, amount})
	if len(amounts) == 0 || !amount.IsPositive() {
		return "", fmt.Errorf("amounts is invalid")
	}
	dbUtxos, err := s.b.GetAddrRawUtxos(from)
	if err != nil {
		return "", err
	}

	ptn := dagconfig.DagConfig.GasToken

	poolTxs, err := s.b.GetPoolTxsByAddr(from)
	allutxos, err := SelectUtxoFromDagAndPool(dbUtxos, poolTxs, from, ptn)
	if err != nil {
		return "", fmt.Errorf("Select utxo err")
	}
	limitdao, _ := decimal.NewFromString("0.0001")
	if !fee.GreaterThanOrEqual(limitdao) {
		return "", fmt.Errorf("fee cannot less than 1 PTN ")
	}
	daoAmount := ptnjson.Ptn2Dao(amount.Add(fee))
	if daoAmount <= 100000000 {
		return "", fmt.Errorf("amount cannot less than 1 dao ")
	}
	utxos, _ := convertUtxoMap2Utxos(allutxos)
	taken_utxo, change, err := core.Select_utxo_Greedy(utxos, daoAmount)
	if err != nil {
		return "", fmt.Errorf("Select utxo err")
	}

	var inputs []ptnjson.TransactionInput
	var input ptnjson.TransactionInput
	for _, u := range taken_utxo {
		utxo := u.(*modules.UtxoWithOutPoint)
		input.Txid = utxo.TxHash.String()
		input.MessageIndex = utxo.MessageIndex
		input.Vout = utxo.OutIndex
		inputs = append(inputs, input)
	}

	if change > 0 {
		amounts = append(amounts, ptnjson.AddressAmt{from, ptnjson.Dao2Ptn(change)})
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
		return nil, nil, err
	}
	poolTxs, err := s.b.GetPoolTxsByAddr(from)

	utxosPTN, err := SelectUtxoFromDagAndPool(dbUtxos, poolTxs, from, ptn)
	if err != nil {
		return nil, nil, fmt.Errorf("Select utxo err")
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
		return nil, nil, fmt.Errorf("Select utxo err")
	}
	tokenAmount := ptnjson.JsonAmt2AssetAmt(tokenAsset, amount)
	pay2, usedUtxo2, err := createPayment(fromAddr, toAddr, tokenAmount, 0, utxosToken)
	if err != nil {
		return nil, nil, err
	}
	tx.TxMessages = append(tx.TxMessages, modules.NewMessage(modules.APP_PAYMENT, pay2))
	for _, u := range usedUtxo2 {
		usedUtxo1 = append(usedUtxo1, u)
	}
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
		return nil, nil, fmt.Errorf("Select utxo err")
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
		payPTN.AddTxOut(modules.NewTxOut(amountToken, tokenengine.GenerateLockScript(toAddr), asset))
	}
	if change > 0 {
		payPTN.AddTxOut(modules.NewTxOut(change, tokenengine.GenerateLockScript(fromAddr), asset))
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
	var inputjson []walletjson.InputJson
	for _, input := range c.Inputs {
		txHash := common.HexToHash(input.Txid)

		inputjson = append(inputjson, walletjson.InputJson{TxHash: input.Txid, MessageIndex: input.MessageIndex, OutIndex: input.Vout, HashForSign: "", Signature: ""})
		prevOut := modules.NewOutPoint(txHash, input.MessageIndex, input.Vout)
		txInput := modules.NewTxIn(prevOut, []byte{})
		pload.AddTxIn(txInput)
	}
	var OutputJson []walletjson.OutputJson
	// Add all transaction outputs to the transaction after performing
	//	// some validity checks.
	//	//only support mainnet
	//	var params *chaincfg.Params
	var ppscript []byte
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
		pkScript := tokenengine.GenerateLockScript(addr)
		ppscript = pkScript
		// Convert the amount to satoshi.
		dao := ptnjson.Ptn2Dao(ptnAmt)
		if err != nil {
			context := "Failed to convert amount"
			return "", internalRPCError(err.Error(), context)
		}
		assetId := dagconfig.DagConfig.GetGasToken()
		txOut := modules.NewTxOut(uint64(dao), pkScript, assetId.ToAsset())
		pload.AddTxOut(txOut)
		OutputJson = append(OutputJson, walletjson.OutputJson{Amount: uint64(dao), Asset: assetId.String(), ToAddress: addr.String()})
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
		if ok == false {
			continue
		}
		for inputindex, _ := range payload.Inputs {
			hashforsign, err := tokenengine.CalcSignatureHash(mtxtmp, tokenengine.SigHashAll, msgindex, inputindex, ppscript)
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
	if upper_type != "ALL" && upper_type != "NONE" && upper_type != "SINGLE" {
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
		if ok == false {
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
			input := ptnjson.RawTxInput{TxHash, uvu.OutIndex, uvu.MessageIndex, PkScriptHex, ""}
			srawinputs = append(srawinputs, input)
			addr, err = tokenengine.GetAddressFromScript(hexutil.MustDecode(uvu.PkScriptHex))
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
	var outAmount uint64
	var outpoint_txhash common.Hash
	for _, msg := range tx.TxMessages {
		payload, ok := msg.Payload.(*modules.PaymentPayload)
		if ok == false {
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
		if ok == false {
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
		amounts = append(amounts, ptnjson.AddressAmt{outOne.Address, outOne.Amount})
		amount = amount.Add(outOne.Amount)
	}
	if len(amounts) == 0 || !amount.IsPositive() {
		return common.Hash{}, err
	}

	dbUtxos, err := s.b.GetAddrRawUtxos(proofTransactionGenParams.From)
	if err != nil {
		return common.Hash{}, err
	}
	poolTxs, err := s.b.GetPoolTxsByAddr(proofTransactionGenParams.From)
	if err == nil {
		if err != nil {
			return common.Hash{}, fmt.Errorf("Select utxo err")
		}
	} // end of pooltx is not nil
	utxos, err := SelectUtxoFromDagAndPool(dbUtxos, poolTxs, proofTransactionGenParams.From, dagconfig.DagConfig.GasToken)

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
		return common.Hash{}, fmt.Errorf("Select utxo err")
	}

	var inputs []ptnjson.TransactionInput
	var input ptnjson.TransactionInput
	for _, u := range taken_utxo {
		utxo := u.(*modules.UtxoWithOutPoint)
		input.Txid = utxo.TxHash.String()
		input.MessageIndex = utxo.MessageIndex
		input.Vout = utxo.OutIndex
		inputs = append(inputs, input)
	}

	if change > 0 {
		amounts = append(amounts, ptnjson.AddressAmt{proofTransactionGenParams.From, ptnjson.Dao2Ptn(change)})
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
	PkScript := tokenengine.GenerateLockScript(from)
	PkScriptHex := hexutil.Encode(PkScript)
	for _, msg := range tx.TxMessages {
		payload, ok := msg.Payload.(*modules.PaymentPayload)
		if ok == false {
			continue
		}
		for _, txin := range payload.Inputs {
			TxHash := txin.PreviousOutPoint.TxHash.String()
			OutIndex := txin.PreviousOutPoint.OutIndex
			MessageIndex := txin.PreviousOutPoint.MessageIndex
			input := ptnjson.RawTxInput{TxHash, OutIndex, MessageIndex, PkScriptHex, ""}
			srawinputs = append(srawinputs, input)
			addr, err = tokenengine.GetAddressFromScript(hexutil.MustDecode(PkScriptHex))
			if err != nil {
				return common.Hash{}, err
			}
		}
	}
	//const max = uint64(time.Duration(math.MaxInt64) / time.Second)
	var duration *uint64
	const max = uint64(time.Duration(math.MaxInt64) / time.Second)
	var d time.Duration
	if duration == nil {
		d = 300 * time.Second
	} else if *duration > max {
		return common.Hash{}, err
	} else {
		d = time.Duration(*duration) * time.Second
	}

	ks := s.b.GetKeyStore()
	err = ks.TimedUnlock(accounts.Account{Address: addr}, password, d)
	if err != nil {
		errors.New("get addr by outpoint is err")
		return common.Hash{}, err
	}

	newsign := ptnjson.NewSignRawTransactionCmd(result, &srawinputs, &keys, ptnjson.String("ALL"))
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
		if ok == false {
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
	var inputjson []walletjson.InputJson
	for _, input := range c.Inputs {
		txHash := common.HexToHash(input.Txid)

		inputjson = append(inputjson, walletjson.InputJson{TxHash: input.Txid, MessageIndex: input.MessageIndex, OutIndex: input.Vout, HashForSign: "", Signature: ""})
		prevOut := modules.NewOutPoint(txHash, input.MessageIndex, input.Vout)
		txInput := modules.NewTxIn(prevOut, []byte{})
		pload.AddTxIn(txInput)
	}
	var OutputJson []walletjson.OutputJson
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
		pkScript := tokenengine.GenerateLockScript(addr)
		// Convert the amount to satoshi.
		dao := ptnjson.Ptn2Dao(ptnAmt)
		if err != nil {
			context := "Failed to convert amount"
			return "", internalRPCError(err.Error(), context)
		}
		assetId := dagconfig.DagConfig.GetGasToken()
		txOut := modules.NewTxOut(uint64(dao), pkScript, assetId.ToAsset())
		pload.AddTxOut(txOut)
		OutputJson = append(OutputJson, walletjson.OutputJson{Amount: uint64(dao), Asset: assetId.String(), ToAddress: addr.String()})
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

//sign rawtranscation
//create raw transction
func (s *PublicWalletAPI) GetPtnTestCoin(ctx context.Context, from string, to string, amount, password string, duration *uint64) (common.Hash, error) {
	var LockTime int64
	LockTime = 0

	amounts := []ptnjson.AddressAmt{}
	if to == "" {
		return common.Hash{}, nil
	}
	a, err := RandFromString(amount)
	if err != nil {
		return common.Hash{}, err
	}
	amounts = append(amounts, ptnjson.AddressAmt{to, a})

	utxoJsons, err := s.b.GetAddrUtxos(from)
	if err != nil {
		return common.Hash{}, err
	}
	utxos := core.Utxos{}
	ptn := dagconfig.DagConfig.GetGasToken()
	for _, json := range utxoJsons {
		//utxos = append(utxos, &json)
		if json.Asset == ptn.String() {
			utxos = append(utxos, &ptnjson.UtxoJson{TxHash: json.TxHash, MessageIndex: json.MessageIndex, OutIndex: json.OutIndex, Amount: json.Amount, Asset: json.Asset, PkScriptHex: json.PkScriptHex, PkScriptString: json.PkScriptString, LockTime: json.LockTime})
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

	var inputs []ptnjson.TransactionInput
	var input ptnjson.TransactionInput
	for _, u := range taken_utxo {
		utxo := u.(*ptnjson.UtxoJson)
		input.Txid = utxo.TxHash
		input.MessageIndex = utxo.MessageIndex
		input.Vout = utxo.OutIndex
		inputs = append(inputs, input)
	}

	if change > 0 {
		amounts = append(amounts, ptnjson.AddressAmt{from, ptnjson.Dao2Ptn(change)})
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
		if ok == false {
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
			input := ptnjson.RawTxInput{TxHash, uvu.OutIndex, uvu.MessageIndex, PkScriptHex, ""}
			srawinputs = append(srawinputs, input)
			addr, err = tokenengine.GetAddressFromScript(hexutil.MustDecode(uvu.PkScriptHex))
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
		errors.New("get addr by outpoint is err")
		//return nil, err
		return common.Hash{}, err
	}

	newsign := ptnjson.NewSignRawTransactionCmd(result, &srawinputs, &keys, ptnjson.String("ALL"))
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
		if ok == false {
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
	rd := big.NewInt(int64(r))
	for {
		r = rand.Int()
		rd = big.NewInt(int64(r))

		rand_number = decimal.NewFromBigInt(rd, int32(exp))
		result = rand_number.Mod(input_number)
		if result.IsZero() == false {
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
	err = s.unlockKS(fromAddr, password, duration)
	if err != nil {
		return common.Hash{}, err
	}
	//3.
	_, err = tokenengine.SignTxAllPaymentInput(rawTx, 1, utxoLockScripts, nil, getPubKeyFn, getSignFn)
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
	err = s.unlockKS(fromAddr, password, nil)
	if err != nil {
		return common.Hash{}, err
	}
	//3.
	_, err = tokenengine.SignTxAllPaymentInput(rawTx, 1, utxoLockScripts, nil, getPubKeyFn, getSignFn)
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
		get.FileHash = string(file.MainData)
		get.ExtraData = string(file.ExtraData)
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

//好像某个UTXO是被那个交易花费的
func (s *PublicWalletAPI) GetStxo(ctx context.Context, txid string, msgIdx int, outIdx int) (*ptnjson.StxoJson, error) {
	outpoint := modules.NewOutPoint(common.HexToHash(txid), uint32(msgIdx), uint32(outIdx))
	return s.b.GetStxoEntry(outpoint)
}
