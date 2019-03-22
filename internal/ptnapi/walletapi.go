package ptnapi

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/rlp"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/hexutil"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/math"
	"github.com/palletone/go-palletone/core"
	"github.com/palletone/go-palletone/core/accounts"
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

type PublicWalletAPI struct {
	b Backend
}

func NewPublicWalletAPI(b Backend) *PublicWalletAPI {
	return &PublicWalletAPI{b}
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
	utxoJsons, err := s.b.GetAddrUtxos(from)
	if err != nil {
		return "", err
	}
	utxos := core.Utxos{}
	ptn := modules.CoreAsset.String()
	for _, json := range utxoJsons {
		//utxos = append(utxos, &json)
		if json.Asset == ptn {
			utxos = append(utxos, &ptnjson.UtxoJson{TxHash: json.TxHash, MessageIndex: json.MessageIndex, OutIndex: json.OutIndex, Amount: json.Amount, Asset: json.Asset, PkScriptHex: json.PkScriptHex, PkScriptString: json.PkScriptString, LockTime: json.LockTime})
		}
	}
	if !fee.IsPositive() {
		return "", fmt.Errorf("fee is ZERO ")
	}
	daoAmount := ptnjson.Ptn2Dao(amount.Add(fee))

	taken_utxo, change, err := core.Select_utxo_Greedy(utxos, daoAmount)
	if err != nil {
		return "", fmt.Errorf("Select utxo err")
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
	result, _ := WalletCreateTransaction(arg)
	//fmt.Println(result)
	return result, nil
}
func WalletCreateTransaction( /*s *rpcServer*/ c *ptnjson.CreateRawTransactionCmd) (string, error) {

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
		asset := modules.NewPTNAsset()
		txOut := modules.NewTxOut(uint64(dao), pkScript, asset)
		pload.AddTxOut(txOut)
		OutputJson = append(OutputJson, walletjson.OutputJson{Amount: uint64(dao), Asset: asset.String(), ToAddress: addr.String()})
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
	// sign mtx
	for index, input := range inputjson {
		hashforsign, err := tokenengine.CalcSignatureHash(mtx, int(input.MessageIndex), int(input.OutIndex), nil)
		if err != nil {
			return "", err
		}
		sh := common.BytesToHash(hashforsign)
		inputjson[index].HashForSign = sh.String()
	}
	//PaymentJson := walletjson.PaymentJson{}
	//PaymentJson.Inputs = inputjson
	//PaymentJson.Outputs = OutputJson
	//txjson := walletjson.TxJson{}
	//txjson.Payload = append(txjson.Payload, PaymentJson)
	/*bytetxjson, err := json.Marshal(txjson)
	if err != nil {
		return "", err
	}*/
	bytetxjson, err := json.Marshal(mtx)
	if err != nil {
		return "", err
	}
	mtxbt, err := rlp.EncodeToBytes(bytetxjson)
	if err != nil {
		return "", err
	}
	//log.Debugf("payload input outpoint:%s", pload.Input[0].PreviousOutPoint.TxHash.String())
	mtxHex := hex.EncodeToString(mtxbt)
	return mtxHex, nil
	//return string(bytetxjson), nil
}

// walletSendTransaction will add the signed transaction to the transaction pool.
// The sender is responsible for signing the transaction and using the correct nonce.
func (s *PublicWalletAPI) SendRawTransaction(ctx context.Context, params string) (common.Hash, error) {
	var RawTxjsonGenParams walletjson.TxJson
	err := json.Unmarshal([]byte(params), &RawTxjsonGenParams)
	if err != nil {
		return common.Hash{}, err
	}
	//fmt.Printf("---------------------------RawTxjsonGenParams----------%+v\n",RawTxjsonGenParams)
	pload := new(modules.PaymentPayload)
	for _, input := range RawTxjsonGenParams.Payload[0].Inputs {
		txHash := common.HexToHash(input.TxHash)
		if err != nil {
			return common.Hash{}, rpcDecodeHexError(input.TxHash)
		}
		prevOut := modules.NewOutPoint(txHash, input.MessageIndex, input.OutIndex)
		hh, err := hexutil.Decode(input.Signature)
		if err != nil {
			return common.Hash{}, rpcDecodeHexError(input.TxHash)
		}
		txInput := modules.NewTxIn(prevOut, hh)
		pload.AddTxIn(txInput)
	}
	for _, output := range RawTxjsonGenParams.Payload[0].Outputs {
		Addr, err := common.StringToAddress(output.ToAddress)
		if err != nil {
			return common.Hash{}, err
		}
		pkScript := tokenengine.GenerateLockScript(Addr)
		asset, err := modules.StringToAsset(output.Asset)
		txOut := modules.NewTxOut(uint64(output.Amount), pkScript, asset)
		pload.AddTxOut(txOut)
	}
	mtx := &modules.Transaction{
		TxMessages: make([]*modules.Message, 0),
	}
	mtx.TxMessages = append(mtx.TxMessages, modules.NewMessage(modules.APP_PAYMENT, pload))

	log.Debugf("Tx outpoint tx hash:%s", mtx.TxMessages[0].Payload.(*modules.PaymentPayload).Inputs[0].PreviousOutPoint.TxHash.String())
	return submitTransaction(ctx, s.b, mtx)
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

	utxoJsons, err := s.b.GetAddrUtxos(proofTransactionGenParams.From)
	if err != nil {
		return common.Hash{}, err
	}
	utxos := core.Utxos{}
	dagOutpoint := []modules.OutPoint{}
	ptn := modules.CoreAsset.String()
	for _, json := range utxoJsons {
		//utxos = append(utxos, &json)
		if json.Asset == ptn {
			utxos = append(utxos, &ptnjson.UtxoJson{TxHash: json.TxHash, MessageIndex: json.MessageIndex, OutIndex: json.OutIndex, Amount: json.Amount, Asset: json.Asset, PkScriptHex: json.PkScriptHex, PkScriptString: json.PkScriptString, LockTime: json.LockTime})
			dagOutpoint = append(dagOutpoint, modules.OutPoint{TxHash: common.HexToHash(json.TxHash), MessageIndex: json.MessageIndex, OutIndex: json.OutIndex})
		}
	}
	poolTxs, err := s.b.GetPoolTxsByAddr(proofTransactionGenParams.From)
	if err == nil {
		utxos, err = SelectUtxoFromDagAndPool(s.b, poolTxs, dagOutpoint, proofTransactionGenParams.From, "PTN")
		if err != nil {
			return common.Hash{}, fmt.Errorf("Select utxo err")
		}
	} // end of pooltx is not nil
	fee := proofTransactionGenParams.Fee
	if !fee.IsPositive() {
		return common.Hash{}, fmt.Errorf("fee is ZERO ")
	}
	daoAmount := ptnjson.Ptn2Dao(amount.Add(fee))
	taken_utxo, change, err := core.Select_utxo_Greedy(utxos, daoAmount)
	if err != nil {
		return common.Hash{}, fmt.Errorf("Select utxo err")
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
	getSignFn := func(addr common.Address, hash []byte) ([]byte, error) {
		ks := s.b.GetKeyStore()
		//account, _ := MakeAddress(ks, addr.String())
		//privKey, _ := ks.DumpPrivateKey(account, "1")
		return ks.SignHash(addr, hash)
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
			/*inpoint := modules.OutPoint{
				TxHash:       txin.PreviousOutPoint.TxHash,
				OutIndex:     txin.PreviousOutPoint.OutIndex,
				MessageIndex: txin.PreviousOutPoint.MessageIndex,
			}
			uvu, eerr := s.b.GetUtxoEntry(&inpoint)
			if eerr != nil {
				return common.Hash{}, err
			}*/
			TxHash := txin.PreviousOutPoint.TxHash.String()
			OutIndex := txin.PreviousOutPoint.OutIndex
			MessageIndex := txin.PreviousOutPoint.MessageIndex
			input := ptnjson.RawTxInput{TxHash, OutIndex, MessageIndex, PkScriptHex, ""}
			srawinputs = append(srawinputs, input)
			addr, err = tokenengine.GetAddressFromScript(hexutil.MustDecode(PkScriptHex))
			if err != nil {
				return common.Hash{}, err
				//fmt.Println("get addr by outpoint is err")
			}
		}
	}
	const max = uint64(time.Duration(math.MaxInt64) / time.Second)
	var d time.Duration
	//var duration
	//if duration == nil {
	d = 300 * time.Second
	//} else if *duration > max {

	//return common.Hash{}, err
	//} else {
	//	d = time.Duration(*duration) * time.Second
	//}
	ks := s.b.GetKeyStore()
	err = ks.TimedUnlock(accounts.Account{Address: addr}, password, d)
	if err != nil {
		errors.New("get addr by outpoint is err")
		//return nil, err
		return common.Hash{}, err
	}

	newsign := ptnjson.NewSignRawTransactionCmd(result, &srawinputs, &keys, ptnjson.String("ALL"))
	signresult, _ := SignRawTransaction(newsign, getPubKeyFn, getSignFn,addr)

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
	//	// some validity checks.
	//	//only support mainnet
	//	var params *chaincfg.Params
	for _, addramt := range c.Amounts {
		encodedAddr := addramt.Address
		ptnAmt := addramt.Amount
		amount := ptnjson.Ptn2Dao(ptnAmt)
		//		// Ensure amount is in the valid range for monetary amounts.
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
		asset := modules.NewPTNAsset()
		txOut := modules.NewTxOut(uint64(dao), pkScript, asset)
		pload.AddTxOut(txOut)
		OutputJson = append(OutputJson, walletjson.OutputJson{Amount: uint64(dao), Asset: asset.String(), ToAddress: addr.String()})
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

	mtx.TxMessages = append(mtx.TxMessages, modules.NewMessage(modules.APP_DATA, textPayload))
	//mtx.TxHash = mtx.Hash()
	// sign mtx
	/*for index, input := range inputjson {
		hashforsign, err := tokenengine.CalcSignatureHash(mtx, int(input.MessageIndex), int(input.OutIndex), nil)
		if err != nil {
			return "", err
		}
		sh := common.BytesToHash(hashforsign)
		inputjson[index].HashForSign = sh.String()
	}
	ProofJson := walletjson.ProofJson{}
	ProofJson.Inputs = inputjson
	ProofJson.Outputs = OutputJson
	ProofJson.Proof = string(textPayload.MainData)
	ProofJson.Extra = string(textPayload.ExtraData)
	txproofjson := walletjson.TxProofJson{}
	txproofjson.Payload = append(txproofjson.Payload,ProofJson)
	bytetxproofjson, err := json.Marshal(txproofjson)
	if err != nil {
		return "", err
	}*/
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

func (s *PublicWalletAPI) GetTranscations(ctx context.Context, address string) (string, error) {
	txs, err := s.b.GetAddrTransactions(address)
	if err != nil {
		return "null", err
	}

	gets := []ptnjson.GetTransactions{}
	for _, tx := range txs {

		get := ptnjson.GetTransactions{}
		get.Txid = tx.Hash().String()

		for _, msg := range tx.TxMessages {
			payload, ok := msg.Payload.(*modules.PaymentPayload)

			if ok == false {
				continue
			}

			for _, txin := range payload.Inputs {

				if txin.PreviousOutPoint != nil {
					addr, err := s.b.GetAddrByOutPoint(txin.PreviousOutPoint)
					if err != nil {

						return "null", err
					}

					get.Inputs = append(get.Inputs, addr.String())
				} else {
					get.Inputs = append(get.Inputs, "coinbase")
				}

			}

			for _, txout := range payload.Outputs {
				var gout ptnjson.GetTranscationOut
				addr, err := tokenengine.GetAddressFromScript(txout.PkScript)
				if err != nil {
					return "null", err
				}
				gout.Addr = addr.String()
				gout.Value = txout.Value
				gout.Asset = txout.Asset.String()
				get.Outputs = append(get.Outputs, gout)
			}

			gets = append(gets, get)
		}

	}
	result := ptnjson.ConvertGetTransactions2Json(gets)

	return result, nil
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
	a, err := decimal.RandFromString(amount)
	if err != nil {
		return common.Hash{}, err
	}
	amounts = append(amounts, ptnjson.AddressAmt{to, a})

	utxoJsons, err := s.b.GetAddrUtxos(from)
	if err != nil {
		return common.Hash{}, err
	}
	utxos := core.Utxos{}
	ptn := modules.CoreAsset.String()
	for _, json := range utxoJsons {
		//utxos = append(utxos, &json)
		if json.Asset == ptn {
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
	getSignFn := func(addr common.Address, hash []byte) ([]byte, error) {
		ks := s.b.GetKeyStore()
		//account, _ := MakeAddress(ks, addr.String())
		//privKey, _ := ks.DumpPrivateKey(account, "1")
		return ks.SignHash(addr, hash)
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
				//fmt.Println("get addr by outpoint is err")
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
	signresult, _ := SignRawTransaction(newsign, getPubKeyFn, getSignFn,addr)

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

func (s *PublicWalletAPI) Ccinvoketx(ctx context.Context, from, to, daoAmount, daoFee, deployId string, param []string) (string, error) {
	contractAddr, _ := common.StringToAddress(deployId)

	fromAddr, _ := common.StringToAddress(from)
	toAddr, _ := common.StringToAddress(to)
	amount, _ := strconv.ParseUint(daoAmount, 10, 64)
	fee, _ := strconv.ParseUint(daoFee, 10, 64)

	log.Info("-----Ccinvoketx:", "contractId", contractAddr.String())
	log.Info("-----Ccinvoketx:", "fromAddr", fromAddr.String())
	log.Info("-----Ccinvoketx:", "toAddr", toAddr.String())
	log.Info("-----Ccinvoketx:", "amount", amount)
	log.Info("-----Ccinvoketx:", "fee", fee)

	if fee <= 0 {
		return "", fmt.Errorf("fee is ZERO ")
	}
	args := make([][]byte, len(param))
	for i, arg := range param {
		args[i] = []byte(arg)
		fmt.Printf("index[%d], value[%s]\n", i, arg)
	}
	reqId, err := s.b.ContractInvokeReqTx(fromAddr, toAddr, amount, fee, contractAddr, args, 0)
	log.Info("-----ContractInvokeTxReq:" + hex.EncodeToString(reqId[:]))

	return hex.EncodeToString(reqId[:]), err
}

func (s *PublicWalletAPI) unlockKS(addr common.Address, password string, duration *uint64) error {
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
		errors.New("get addr by outpoint is err")
		return err
	}
	return nil
}

func (s *PublicWalletAPI) TransferToken(ctx context.Context, asset string, from string, to string,
	amount decimal.Decimal, fee decimal.Decimal, Extra string, password string, duration *uint64) (common.Hash, error) {
	//
	tokenAsset, err := modules.StringToAsset(asset)
	if err != nil {
		fmt.Println(err.Error())
		return common.Hash{}, err
	}
	if !fee.IsPositive() {
		return common.Hash{}, fmt.Errorf("fee is ZERO ")
	}
	//
	fromAddr, err := common.StringToAddress(from)
	if err != nil {
		fmt.Println(err.Error())
		return common.Hash{}, err
	}
	toAddr, err := common.StringToAddress(to)
	if err != nil {
		fmt.Println(err.Error())
		return common.Hash{}, err
	}
	//all utxos
	utxoJsons, err := s.b.GetAddrUtxos(from)
	if err != nil {
		return common.Hash{}, err
	}
	//ptn utxos and token utxos
	utxosPTN := core.Utxos{}
	utxosToken := core.Utxos{}
	ptn := modules.CoreAsset.String()
	dagOutpoint_token := []modules.OutPoint{}
	dagOutpoint_ptn := []modules.OutPoint{}
	for _, json := range utxoJsons {
		//utxos = append(utxos, &json)
		if json.Asset == asset {
			utxosToken = append(utxosToken, &ptnjson.UtxoJson{TxHash: json.TxHash, MessageIndex: json.MessageIndex, OutIndex: json.OutIndex, Amount: json.Amount, Asset: json.Asset, PkScriptHex: json.PkScriptHex, PkScriptString: json.PkScriptString, LockTime: json.LockTime})
			dagOutpoint_token = append(dagOutpoint_token, modules.OutPoint{TxHash: common.HexToHash(json.TxHash), MessageIndex: json.MessageIndex, OutIndex: json.OutIndex})
		}
		if json.Asset == ptn {
			utxosPTN = append(utxosPTN, &ptnjson.UtxoJson{TxHash: json.TxHash, MessageIndex: json.MessageIndex, OutIndex: json.OutIndex, Amount: json.Amount, Asset: json.Asset, PkScriptHex: json.PkScriptHex, PkScriptString: json.PkScriptString, LockTime: json.LockTime})
			dagOutpoint_ptn = append(dagOutpoint_ptn, modules.OutPoint{TxHash: common.HexToHash(json.TxHash), MessageIndex: json.MessageIndex, OutIndex: json.OutIndex})
		}
	}
	poolTxs, err := s.b.GetPoolTxsByAddr(from)
	if len(poolTxs) > 0 {
		utxosToken, err = SelectUtxoFromDagAndPool(s.b, poolTxs, dagOutpoint_token, from, asset)
		if err != nil {
			return common.Hash{}, fmt.Errorf("Select utxo err")
		}
		utxosPTN, err = SelectUtxoFromDagAndPool(s.b, poolTxs, dagOutpoint_ptn, from, "PTN")
		if err != nil {
			return common.Hash{}, fmt.Errorf("Select utxo err")
		}
	}
	//else{
	//ptn utxos and token utxos
	/*for _, json := range utxoJsons {
		if json.Asset == ptn {
			utxosPTN = append(utxosPTN, &ptnjson.UtxoJson{TxHash: json.TxHash,
				MessageIndex:   json.MessageIndex,
				OutIndex:       json.OutIndex,
				Amount:         json.Amount,
				Asset:          json.Asset,
				PkScriptHex:    json.PkScriptHex,
				PkScriptString: json.PkScriptString,
				LockTime:       json.LockTime})
		} else {
			if json.Asset == asset {
				utxosToken = append(utxosToken, &ptnjson.UtxoJson{TxHash: json.TxHash,
					MessageIndex:   json.MessageIndex,
					OutIndex:       json.OutIndex,
					Amount:         json.Amount,
					Asset:          json.Asset,
					PkScriptHex:    json.PkScriptHex,
					PkScriptString: json.PkScriptString,
					LockTime:       json.LockTime})
			}
		}
	}*/
	// }
	//1.
	tokenAmount := ptnjson.JsonAmt2AssetAmt(tokenAsset, amount)
	feeAmount := ptnjson.Ptn2Dao(fee)
	tx, err := createTokenTx(fromAddr, toAddr, tokenAmount, feeAmount, utxosPTN, utxosToken, tokenAsset)
	if err != nil {
		return common.Hash{}, err
	}
	if Extra != "" {
		textPayload := new(modules.DataPayload)
		textPayload.MainData = []byte(asset)
		textPayload.ExtraData = []byte(Extra)
		tx.TxMessages = append(tx.TxMessages, modules.NewMessage(modules.APP_DATA, textPayload))
	}

	//lockscript
	getPubKeyFn := func(addr common.Address) ([]byte, error) {
		//TODO use keystore
		ks := s.b.GetKeyStore()
		return ks.GetPublicKey(addr)
	}
	//sign tx
	getSignFn := func(addr common.Address, hash []byte) ([]byte, error) {
		ks := s.b.GetKeyStore()
		return ks.SignHash(addr, hash)
	}
	//raw inputs
	var rawInputs []ptnjson.RawTxInput
	PkScript := tokenengine.GenerateLockScript(fromAddr)
	PkScriptHex := trimx(hexutil.Encode(PkScript))
	for _, msg := range tx.TxMessages {
		payload, ok := msg.Payload.(*modules.PaymentPayload)
		if ok == false {
			continue
		}
		for _, txin := range payload.Inputs {
			/*inpoint := modules.OutPoint{
				TxHash:       txin.PreviousOutPoint.TxHash,
				OutIndex:     txin.PreviousOutPoint.OutIndex,
				MessageIndex: txin.PreviousOutPoint.MessageIndex,
			}
			uvu, eerr := s.b.GetUtxoEntry(&inpoint)
			if eerr != nil {
				return common.Hash{}, err
			}*/
			TxHash := trimx(txin.PreviousOutPoint.TxHash.String())
			OutIndex := txin.PreviousOutPoint.OutIndex
			MessageIndex := txin.PreviousOutPoint.MessageIndex
			input := ptnjson.RawTxInput{TxHash, OutIndex, MessageIndex, PkScriptHex, ""}
			rawInputs = append(rawInputs, input)
			/*addr, err := tokenengine.GetAddressFromScript(hexutil.MustDecode(PkScriptHex))
			if err != nil {
				return common.Hash{}, err
				//fmt.Println("get addr by outpoint is err")
			}*/
			/*TxHash := trimx(uvu.TxHash)
			PkScriptHex := trimx(uvu.PkScriptHex)
			input := ptnjson.RawTxInput{TxHash, uvu.OutIndex, uvu.MessageIndex, PkScriptHex, ""}
			rawInputs = append(rawInputs, input)*/
		}
	}
	//2.
	err = s.unlockKS(fromAddr, password, duration)
	if err != nil {
		return common.Hash{}, err
	}
	//3.
	err = signTokenTx(tx, rawInputs, "ALL", getPubKeyFn, getSignFn)
	if err != nil {
		return common.Hash{}, err
	}
	//4.
	return submitTransaction(ctx, s.b, tx)
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

//contract command
//install
func (s *PublicWalletAPI) Ccinstall(ctx context.Context, ccname string, ccpath string, ccversion string) (hexutil.Bytes, error) {
	log.Info("CcInstall:" + ccname + ":" + ccpath + "_" + ccversion)

	templateId, err := s.b.ContractInstall(ccname, ccpath, ccversion)
	return hexutil.Bytes(templateId), err
}
func (s *PublicWalletAPI) Ccquery(ctx context.Context, deployId string, param []string) (string, error) {
	contractId, _ := common.StringToAddress(deployId)
	log.Info("-----Ccquery:", "contractId", contractId.String())
	args := make([][]byte, len(param))
	for i, arg := range param {
		args[i] = []byte(arg)
		fmt.Printf("index[%d],value[%s]\n", i, arg)
	}
	//参数前面加入msg0,这里为空
	var fullArgs [][]byte
	msgArg := []byte("query has no msg0")
	fullArgs = append(fullArgs, msgArg)
	fullArgs = append(fullArgs, args...)

	txid := fmt.Sprintf("%08v", rand.New(rand.NewSource(time.Now().UnixNano())).Int31n(100000000))

	rsp, err := s.b.ContractQuery(contractId[:], txid[:], fullArgs, 0)
	if err != nil {
		return "", err
	}
	return string(rsp), nil
}

func (s *PublicWalletAPI) Ccstop(ctx context.Context, deployId string, txid string) error {
	depId, _ := hex.DecodeString(deployId)
	log.Info("Ccstop:" + deployId + ":" + txid + "_")
	//TODO deleteImage 为 true 时，目前是会删除基础镜像的
	err := s.b.ContractStop(depId, txid, false)
	return err
}
