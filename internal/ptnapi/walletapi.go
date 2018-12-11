package ptnapi
import (
	"context"
	//"errors"
	"fmt"
	"encoding/json"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/core"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/ptnjson"
	"github.com/palletone/go-palletone/ptnjson/walletjson"
	"github.com/palletone/go-palletone/tokenengine"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/shopspring/decimal"
)

func (s *PublicTransactionPoolAPI) WalletCreateTransaction(ctx context.Context, from string, to string, amount, fee decimal.Decimal) (string, error) {

	//realNet := &chaincfg.MainNetParams
	var LockTime int64
	LockTime = 0

	amounts := map[string]decimal.Decimal{}
	if to == "" {
		return "", fmt.Errorf("amounts is empty")
	}

	amounts[to] = amount

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
		amounts[from] = ptnjson.Dao2Ptn(change)
	}

	arg := ptnjson.NewCreateRawTransactionCmd(inputs, amounts, &LockTime)
	result, _ := WalletCreateTransaction(arg)
	fmt.Println(result)
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
	for encodedAddr, ptnAmt := range c.Amounts {
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
	PaymentJson := walletjson.PaymentJson{}
	PaymentJson.Inputs = inputjson
	PaymentJson.Outputs = OutputJson
	
	mtx := &modules.Transaction{
		TxMessages: make([]*modules.Message, 0),
	}
	mtx.TxMessages = append(mtx.TxMessages, modules.NewMessage(modules.APP_PAYMENT, pload))
	//mtx.TxHash = mtx.Hash()
	// sign mtx 
	for _,input := range PaymentJson.Inputs{
        hashforsign,err := tokenengine.CalcSignatureHash(mtx,int(input.MessageIndex),int(input.OutIndex),nil)
        if err != nil {
		    return "", err
	    }
        input.HashForSign = string(hashforsign)
	}
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
func (s *PublicTransactionPoolAPI) WalletSendTransaction(ctx context.Context, params string) (common.Hash, error) {
    var RawTxjsonGenParams walletjson.RawTxjsonGenParams
	err := json.Unmarshal([]byte(params), &RawTxjsonGenParams)
	if err != nil {
		return common.Hash{}, err
	}

	pload := new(modules.PaymentPayload)
	for _, input := range RawTxjsonGenParams.Inputs {
		txHash, err := common.NewHashFromStr(input.TxHash)
		if err != nil {
			return common.Hash{}, rpcDecodeHexError(input.TxHash)
		}
		prevOut := modules.NewOutPoint(txHash, input.MessageIndex, input.OutIndex)
		txInput := modules.NewTxIn(prevOut, []byte(input.Signature))
		pload.AddTxIn(txInput)
	}
	for _, output := range RawTxjsonGenParams.Outputs {
		Addr,err :=common.StringToAddress(output.Address)
		if err != nil {
		    return common.Hash{}, err
	    }
		pkScript := tokenengine.GenerateLockScript(Addr)
		asset,err :=modules.StringToAsset(output.Asset)
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