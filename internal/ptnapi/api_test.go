package ethapi

import (
	"fmt"
        "bytes"
	"testing"
	"encoding/json"
        "encoding/hex"
	"strings"
        "github.com/palletone/go-palletone/tokenengine/btcd/txscript"
        "github.com/palletone/go-palletone/tokenengine/btcd/chaincfg"
        "github.com/palletone/go-palletone/tokenengine/btcutil"
        "github.com/palletone/go-palletone/tokenengine/btcd/btcjson"
        "github.com/palletone/go-palletone/tokenengine/btcd/wire"
)
type RawTransactionGenParams struct {
	Inputs []struct {
		Txid string `json:"txid"`
		Vout uint32 `json:"vout"`
	} `json:"inputs"`
	Outputs []struct {
		Address string  `json:"address"`
		Amount  float64 `json:"amount"`
	} `json:"outputs"`
	Locktime int64 `json:"locktime"`
}

func TestRawTransactionGen(t *testing.T) {

    params := `{
    "inputs": [
		{
           "txid": "0b987a442dd830f4e40639058030d250f526c2330fb31b64c24be880339bdfd1",
           "vout": 0
		}
    ],
    "outputs": [
		{
           "address": "P1CJfWNRCx4AAfjfqHimurLgdzX7rJZ3Qce",
           "amount": 0.79
		}
    ],
    "locktime": 0
	}`
        params= params
     
	testResult := "0100000001d1df9b3380e84bc2641bb30f33c226f550d23080053906e4f430d82d447a980b0000000000feffffff01c271b504000000001976a9147c0099353492e6d45dd440940605d092506e773988ac0a000000"
        var rawTransactionGenParams RawTransactionGenParams
	err := json.Unmarshal([]byte(params), &rawTransactionGenParams)
	if err != nil {
                fmt.Println("-------test ---- 45455454545   return ")
		return
	}
        //transaction inputs
       
	var inputs []btcjson.TransactionInput
	for _, inputOne := range rawTransactionGenParams.Inputs {
		input := btcjson.TransactionInput{inputOne.Txid, inputOne.Vout}
		inputs = append(inputs, input)
	}
	if len(inputs) == 0 {
                fmt.Println("-----------5858588585858")
		return
	}
//realNet := &chaincfg.MainNetParams
	amounts := map[string]float64{}
	for _, outOne := range rawTransactionGenParams.Outputs {
		if len(outOne.Address) == 0 || outOne.Amount <= 0 {
			continue
		}
		amounts[outOne.Address] = float64(outOne.Amount * 1e8)
	}
     
	if len(amounts) == 0 {
		return
	}
       
        arg := btcjson.NewCreateRawTransactionCmd(inputs, amounts, &rawTransactionGenParams.Locktime)
       
	result ,_ := CreateRawTransaction(arg)
       
	if !strings.Contains(result, testResult) {
		t.Errorf("unexpected result - got: %v, "+"want: %v", result, testResult)
	}
	fmt.Println(result)
	return 
}
/*
func TestDecodeRawTransaction(t *testing.T) {

	rpcParams := RPCParams{
		Host:      "localhost:18332",
		RPCUser:   "zxl",
		RPCPasswd: "123456",
		CertPath:  "C:/Users/zxl/AppData/Local/Btcwallet/rpc.cert",
	}

	testResult := `{"hex":"","txid":"0bf2bbdabd7561fe035eb383d14e376f04690c62301cc78d89dd189f7e6c3a72","version":1,"locktime":0,"vin":[{"txid":"132154398e312b69b62973f8f6a91797bba9996bc60dc1d7b1f8697df196088d","vout":0,"scriptSig":{"asm":"","hex":""},"sequence":4294967295}],"vout":[{"value":0.98811339,"n":0,"scriptPubKey":{"asm":"OP_DUP OP_HASH160 bddc9a62e9b7c3cfdbe1c817520e24e32c339f32 OP_EQUALVERIFY OP_CHECKSIG","hex":"76a914bddc9a62e9b7c3cfdbe1c817520e24e32c339f3288ac","reqSigs":1,"type":"pubkeyhash","addresses":["mxprH5bkXtn9tTTAxdQGPXrvruCUvsBNKt"]}}]}`

	parms := ` {
		    "rawtx": "01000000018d0896f17d69f8b1d7c10dc66b99a9bb9717a9f6f87329b6692b318e395421130000000000ffffffff01cbbde305000000001976a914bddc9a62e9b7c3cfdbe1c817520e24e32c339f3288ac00000000"
		  	}`
	result := DecodeRawTransaction(parms, &rpcParams)
	fmt.Println(result)
	if !strings.Contains(result, testResult) {
		t.Errorf("unexpected result - got: %v, "+"want: %v", result, testResult)
	}
}*/

type SignTransactionParams struct {
	TransactionHex string   `json:"transactionhex"`
	RedeemHex      string   `json:"redeemhex"`
	Privkeys       []string `json:"privkeys"`
}
func TestSignTransaction(t *testing.T) {
	//from TestRawTransactionGen A --> B C
	/*params := `{
        "transactionhex": "0100000001d1df9b3380e84bc2641bb30f33c226f550d23080053906e4f430d82d447a980b0000000000ffffffff01c071b504000000001976a9148c7d9b27d1ec710891223283424adae60ba4f4ba88ac00000000",
        "redeemhex": "",
	"privkeys": ["cPXW9UVJdjLvCmAxPHdQ1gHkpD5paWpf2PmH5MwsXN5MxnRjbAgE"]
  	}`*/
        /*params := `{
    "transactionhex": "010000000236045404e65bd741109db92227ca0dc9274ef717a6612c96cd77b24a17d1bcd70000000000ffffffff7c1f7d5407b41abf29d41cf6f122ef2d40f76d956900d2c89314970951ef5b940000000000ffffffff014431d309000000001976a914bddc9a62e9b7c3cfdbe1c817520e24e32c339f3288ac00000000",
    "redeemhex": "522103940ab29fbf214da2d8ec99c47db63879957311bd90d2f1c635828604d541051421020106ca23b4f28dbc83838ee4745accf90e5621fe70df5b1ee8f7e1b3b41b64cb21029d80ff37838e4989a6aa26af41149d4f671976329e9ddb9b78fdea9814ae6ef553ae",
        "privkeys": ["cUakDAWEeNeXTo3B93WBs9HRMfaFDegXcbEGooLz8BSxRBfmpYcX"]
        }`*/
         params := `{
    "transactionhex": "010000000236045404e65bd741109db92227ca0dc9274ef717a6612c96cd77b24a17d1bcd700000000b400473044022024e6a6ca006f25ccd3ebf5dadf21397a6d7266536cd336061cd17cff189d95e402205af143f6726d75ac77bc8c80edcb6c56579053d2aa31601b23bc8da41385dd86014c69522103940ab29fbf214da2d8ec99c47db63879957311bd90d2f1c635828604d541051421020106ca23b4f28dbc83838ee4745accf90e5621fe70df5b1ee8f7e1b3b41b64cb21029d80ff37838e4989a6aa26af41149d4f671976329e9ddb9b78fdea9814ae6ef553aeffffffff7c1f7d5407b41abf29d41cf6f122ef2d40f76d956900d2c89314970951ef5b9400000000b40047304402206a1d7a2ae07840957bee708b6d3e1fbe7858760ac378b1e21209b348c1e2a5c402204255cd4cd4e5b5805d44bbebe7464aa021377dca5fc6bf4a5632eb2d8bc9f9e4014c69522103940ab29fbf214da2d8ec99c47db63879957311bd90d2f1c635828604d541051421020106ca23b4f28dbc83838ee4745accf90e5621fe70df5b1ee8f7e1b3b41b64cb21029d80ff37838e4989a6aa26af41149d4f671976329e9ddb9b78fdea9814ae6ef553aeffffffff014431d309000000001976a914bddc9a62e9b7c3cfdbe1c817520e24e32c339f3288ac00000000",
    "redeemhex": "522103940ab29fbf214da2d8ec99c47db63879957311bd90d2f1c635828604d541051421020106ca23b4f28dbc83838ee4745accf90e5621fe70df5b1ee8f7e1b3b41b64cb21029d80ff37838e4989a6aa26af41149d4f671976329e9ddb9b78fdea9814ae6ef553ae",
        "privkeys": ["cQJB6w8SxVNoprVwp2xyxUFxvExMbpR2qj3banXYYXmhtTc1WxC8"]
        }`
      
        var signTransactionParams SignTransactionParams
	err := json.Unmarshal([]byte(params), &signTransactionParams)
	if err != nil {
            return
	}
      
	//check empty string
	if "" == signTransactionParams.TransactionHex {
		return
	}
	//decode Transaction hexString to bytes
	rawTXBytes, err := hex.DecodeString(signTransactionParams.TransactionHex)
	if err != nil {
		return
	}
	//deserialize to MsgTx
       
	var tx wire.MsgTx
	err = tx.Deserialize(bytes.NewReader(rawTXBytes))
	if err != nil {
		return
	}
        //get private keys for sign
	var keys []string
	for _, key := range signTransactionParams.Privkeys {
		key = strings.TrimSpace(key) //Trim whitespace
		if len(key) == 0 {
			continue
		}
		keys = append(keys, key)
	}
	if len(keys) == 0 {
		return
	}
       
        realNet := &chaincfg.MainNetParams
        //sign the UTXO hash, must know RedeemHex which contains in RawTxInput
        
	var rawInputs []btcjson.RawTxInput
	for {
        
		//if "" == signTransactionParams.RedeemHex {
                //        fmt.Println("------------169-----------")
		//	break
		//}
        
		//decode redeem's hexString to bytes
		redeem, err := hex.DecodeString(signTransactionParams.RedeemHex)
		if err != nil {
			break
		}
        
		//get multisig payScript
		scriptAddr, err := btcutil.NewAddressScriptHash(redeem, realNet)
		scriptPkScript, err := txscript.PayToAddrScript(scriptAddr)
		//multisig transaction need redeem for sign
		for _, txinOne := range tx.TxIn {
			rawInput := btcjson.RawTxInput{
				txinOne.PreviousOutPoint.Hash.String(), //txid
				txinOne.PreviousOutPoint.Index,         //outindex
				hex.EncodeToString(scriptPkScript),     //multisig pay script
				signTransactionParams.RedeemHex}        //redeem
			rawInputs = append(rawInputs, rawInput)
		}
		break
	}
        txHex := ""
	if &tx != nil {
		// Serialize the transaction and convert to hex string.
              
		buf := bytes.NewBuffer(make([]byte, 0, tx.SerializeSize()))
		if err := tx.Serialize(buf); err != nil {
			return
		}
		txHex = hex.EncodeToString(buf.Bytes())
	}
       
        send_args := btcjson.NewSignRawTransactionCmd(txHex, &rawInputs, &keys, btcjson.String("ALL"))
	//the return 'transactionhex' is used in next step
        
	resultTransToMultsigAddr,err := SignRawTransaction(send_args)
	//	if !strings.Contains(resultTransToMultsigAddr, theComplete) {
	//		t.Errorf("complete - got: false, want: true")
	//	}
       
	fmt.Println(resultTransToMultsigAddr)
	return
}


