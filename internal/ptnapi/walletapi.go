package ptnapi

import (
	"context"
	//"errors"
	"time"
	"errors"
	"encoding/json"
	"fmt"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/hexutil"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/rlp"
	"github.com/palletone/go-palletone/common/math"
	"github.com/palletone/go-palletone/core/accounts"
	"github.com/palletone/go-palletone/core"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/ptnjson"
	"github.com/palletone/go-palletone/ptnjson/walletjson"
	"github.com/palletone/go-palletone/tokenengine"
	"github.com/palletone/go-palletone/core/accounts/keystore"
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
	if to == "" {
		return "", fmt.Errorf("amounts is empty")
	}

	amounts = append(amounts, ptnjson.AddressAmt{to, amount})

	utxoJsons, err := s.b.GetAddrUtxos(from)
	if err != nil {
		return "", err
	}
	utxos := core.Utxos{}
	for _, json := range utxoJsons {
		//utxos = append(utxos, &json)
		utxos = append(utxos, &ptnjson.UtxoJson{TxHash: json.TxHash, MessageIndex: json.MessageIndex, OutIndex: json.OutIndex, Amount: json.Amount, Asset: json.Asset, PkScriptHex: json.PkScriptHex, PkScriptString: json.PkScriptString, LockTime: json.LockTime})
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
		txHash, err := common.NewHashFromStr(input.Txid)
		if err != nil {
			return "", rpcDecodeHexError(input.Txid)
		}
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
		if amount <= 0 || amount > ptnjson.MaxDao {
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
	PaymentJson := walletjson.PaymentJson{}
	PaymentJson.Inputs = inputjson
	PaymentJson.Outputs = OutputJson
	txjson := walletjson.TxJson{}
	txjson.Payload = append(txjson.Payload, PaymentJson)
	bytetxjson, err := json.Marshal(txjson)
	if err != nil {
		return "", err
	}

	return string(bytetxjson), nil
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
		txHash, err := common.NewHashFromStr(input.TxHash)
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

//sign rawtranscation
//create raw transction
func (s *PublicWalletAPI) GetPtnTestCoin(ctx context.Context, from string, to string, amount, password string, duration *uint64) (common.Hash, error) {
	var LockTime int64
	LockTime = 0

	amounts := []ptnjson.AddressAmt{}
	if to == "" {
		return common.Hash{}, nil
	}
	a, err := decimal.NewFromString(amount)
	if err != nil {
		return common.Hash{}, nil
	}
	amounts = append(amounts, ptnjson.AddressAmt{to, a})

	utxoJsons, err := s.b.GetAddrUtxos(from)
	if err != nil {
		return common.Hash{}, err
	}
	utxos := core.Utxos{}
	for _, json := range utxoJsons {
		//utxos = append(utxos, &json)
		utxos = append(utxos, &ptnjson.UtxoJson{TxHash: json.TxHash, MessageIndex: json.MessageIndex, OutIndex: json.OutIndex, Amount: json.Amount, Asset: json.Asset, PkScriptHex: json.PkScriptHex, PkScriptString: json.PkScriptString, LockTime: json.LockTime})
	}
	fee, err := decimal.NewFromString("1")
	if err != nil {
		return common.Hash{}, nil
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
		ks := s.b.AccountManager().Backends(keystore.KeyStoreType)[0].(*keystore.KeyStore)

		return ks.GetPublicKey(addr)
		//privKey, _ := ks.DumpPrivateKey(account, "1")
		//return crypto.CompressPubkey(&privKey.PublicKey), nil
	}
	getSignFn := func(addr common.Address, hash []byte) ([]byte, error) {
		ks := s.b.AccountManager().Backends(keystore.KeyStoreType)[0].(*keystore.KeyStore)
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
				fmt.Println("get addr by outpoint is err")
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
	ks := s.b.AccountManager().Backends(keystore.KeyStoreType)[0].(*keystore.KeyStore)
	err = ks.TimedUnlock(accounts.Account{Address: addr}, password, d)
	if err != nil {
		errors.New("get addr by outpoint is err")
		//return nil, err
		return common.Hash{}, err
	}

	newsign := ptnjson.NewSignRawTransactionCmd(result, &srawinputs, &keys, ptnjson.String("ALL"))
	signresult, _ := SignRawTransaction(newsign, getPubKeyFn, getSignFn)

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
