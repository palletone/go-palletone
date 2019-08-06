/*
   This file is part of go-palletone.
   go-palletone is free software: you can redistribute it and/or modify
   it under the terms of the GNU General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.
   go-palletone is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU General Public License for more details.
   You should have received a copy of the GNU General Public License
   along with go-palletone.  If not, see <http://www.gnu.org/licenses/>.
*/
/*
 * @author PalletOne core developers <dev@pallet.one>
 * @date 2018
 */
package adaptorbtc

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	//"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"

	"github.com/palletone/btc-adaptor/txscript"

	"github.com/palletone/adaptor"
)

// decodeHexStr decodes the hex encoding of a string, possibly prepending a
// leading '0' character if there is an odd number of bytes in the hex string.
// This is to prevent an error for an invalid hex string when using an odd
// number of bytes when calling hex.Decode.
func decodeHexStr(hexStr string) ([]byte, error) {
	if len(hexStr)%2 != 0 {
		hexStr = "0" + hexStr
	}
	decoded, err := hex.DecodeString(hexStr)
	if err != nil {
		return nil, &btcjson.RPCError{
			Code:    btcjson.ErrRPCDecodeHexString,
			Message: "Hex string decode failed: " + err.Error(),
		}
	}
	return decoded, nil
}

type RawTxInput struct {
	Txid         string `json:"txid"`
	Vout         uint32 `json:"vout"`
	ScriptPubKey string `json:"scriptPubKey"`
}
type SignRawTransactionCmd struct {
	RawTx     string
	Inputs    *[]RawTxInput
	RedeemHex []string
	PrivKeys  *[]string
	Flags     *string `jsonrpcdefault:"\"ALL\""`
}

// signRawTransaction handles the signrawtransaction command.
func signRawTransactionCmd(cmd *SignRawTransactionCmd, chainParams *chaincfg.Params) (interface{}, error) {
	serializedTx, err := decodeHexStr(cmd.RawTx)
	if err != nil {
		return nil, err
	}
	var tx wire.MsgTx
	err = tx.Deserialize(bytes.NewBuffer(serializedTx))
	if err != nil {
		return nil, errors.New("TX decode failed")
	}

	var hashType txscript.SigHashType
	switch *cmd.Flags {
	case "ALL":
		hashType = txscript.SigHashAll
	case "NONE":
		hashType = txscript.SigHashNone
	case "SINGLE":
		hashType = txscript.SigHashSingle
	case "ALL|ANYONECANPAY":
		hashType = txscript.SigHashAll | txscript.SigHashAnyOneCanPay
	case "NONE|ANYONECANPAY":
		hashType = txscript.SigHashNone | txscript.SigHashAnyOneCanPay
	case "SINGLE|ANYONECANPAY":
		hashType = txscript.SigHashSingle | txscript.SigHashAnyOneCanPay
	default:
		return nil, errors.New("Invalid sighash parameter")
	}

	// TODO: really we probably should look these up with btcd anyway to
	// make sure that they match the blockchain if present.
	inputs := make(map[wire.OutPoint][]byte)
	scripts := make(map[string][]byte)
	var cmdInputs []RawTxInput
	if cmd.Inputs != nil {
		cmdInputs = *cmd.Inputs
	}
	for _, inputOne := range cmdInputs {
		inputHash, err := chainhash.NewHashFromStr(inputOne.Txid)
		if err != nil {
			return nil, err
		}

		if "" == inputOne.ScriptPubKey {
			return nil, errors.New("ScriptPubKey is empty")
		}
		script, err := decodeHexStr(inputOne.ScriptPubKey)
		if err != nil {
			return nil, err
		}
		inputs[wire.OutPoint{Hash: *inputHash, Index: inputOne.Vout}] = script

	}
	//
	for i := range cmd.RedeemHex {
		if "" == cmd.RedeemHex[i] {
			continue
		}
		redeemScript, err := decodeHexStr(cmd.RedeemHex[i])
		if err != nil {
			return nil, err
		}

		addr, err := btcutil.NewAddressScriptHash(redeemScript,
			chainParams)
		if err != nil {
			return nil, err
		}
		scripts[addr.String()] = redeemScript
	}

	// Parse list of private keys, if present. If there are any keys here
	// they are the keys that we may use for signing. If empty we will
	// use any keys known to us already.
	var keys map[string]*btcutil.WIF
	if cmd.PrivKeys != nil {
		keys = make(map[string]*btcutil.WIF)

		for _, key := range *cmd.PrivKeys {
			wif, err := btcutil.DecodeWIF(key)
			if err != nil {
				return nil, err
			}

			if !wif.IsForNet(chainParams) {
				s := "key network doesn't match wallet's"
				return nil, errors.New(s)
			}

			addr, err := btcutil.NewAddressPubKey(wif.SerializePubKey(),
				chainParams)
			if err != nil {
				return nil, err
			}
			keys[addr.EncodeAddress()] = wif
		}
	}

	// All args collected. Now we can sign all the inputs that we can.
	// `complete' denotes that we successfully signed all outputs and that
	// all scripts will run to completion. This is returned as part of the
	// reply.
	signErrs, err := SignTransactionReal(&tx, hashType, inputs, keys, scripts, chainParams)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	buf.Grow(tx.SerializeSize())

	// All returned errors (not OOM, which panics) encounted during
	// bytes.Buffer writes are unexpected.
	if err = tx.Serialize(&buf); err != nil {
		panic(err)
	}

	signErrors := make([]btcjson.SignRawTransactionError, 0, len(signErrs))
	for _, e := range signErrs {
		input := tx.TxIn[e.InputIndex]
		signErrors = append(signErrors, btcjson.SignRawTransactionError{
			TxID:      input.PreviousOutPoint.Hash.String(),
			Vout:      input.PreviousOutPoint.Index,
			ScriptSig: hex.EncodeToString(input.SignatureScript),
			Sequence:  input.Sequence,
			Error:     e.Error.Error(),
		})
	}

	return btcjson.SignRawTransactionResult{
		Hex:      hex.EncodeToString(buf.Bytes()),
		Complete: len(signErrors) == 0,
		Errors:   signErrors,
	}, nil
}

// SignatureError records the underlying error when validating a transaction
// input signature.
type SignatureError struct {
	InputIndex uint32
	Error      error
}

func SignTransactionReal(tx *wire.MsgTx, hashType txscript.SigHashType,
	additionalPrevScripts map[wire.OutPoint][]byte,
	additionalKeysByAddress map[string]*btcutil.WIF,
	p2shRedeemScriptsByAddress map[string][]byte, chainParams *chaincfg.Params) ([]SignatureError, error) {

	signErrors := []SignatureError{}
	//var signErrors []SignatureErroerr := walletdb.View(w.db, func(dbtx walletdb.ReadTx) error {
	//addrmgrNs := dbtx.ReadBucket(waddrmgrNamespaceKey)
	//txmgrNs := dbtx.ReadBucket(wtxmgrNamespaceKey)
	var err error
	for i, txIn := range tx.TxIn {
		prevOutScript, ok := additionalPrevScripts[txIn.PreviousOutPoint]
		if !ok {
			/*prevHash := &txIn.PreviousOutPoint.Hash
			prevIndex := txIn.PreviousOutPoint.Index
			txDetails, err := w.TxStore.TxDetails(txmgrNs, prevHash)
			if err != nil {
				return fmt.Errorf("cannot query previous transaction "+
					"details for %v: %v", txIn.PreviousOutPoint, err)
			}
			if txDetails == nil {
				return fmt.Errorf("%v not found",
					txIn.PreviousOutPoint)
			}
			prevOutScript = txDetails.MsgTx.TxOut[prevIndex].PkScript*/
		}

		// Set up our callbacks that we pass to txscript so it can
		// look up the appropriate keys and scripts by address.
		getKey := txscript.KeyClosure(func(addr btcutil.Address) (*btcec.PrivateKey, bool, error) {
			if len(additionalKeysByAddress) != 0 {
				addrStr := addr.EncodeAddress()
				wif, ok := additionalKeysByAddress[addrStr]
				if !ok {
					return nil, false, errors.New("no key for address")
				}
				return wif.PrivKey, wif.CompressPubKey, nil
			}
			return nil, false, errors.New("no key for address")
		})
		getScript := txscript.ScriptClosure(func(addr btcutil.Address) ([]byte, error) {
			// If keys were provided then we can only use the
			// redeem scripts provided with our inputs, too.
			if len(additionalKeysByAddress) != 0 {
				addrStr := addr.EncodeAddress()
				script, ok := p2shRedeemScriptsByAddress[addrStr]
				if !ok {
					return nil, errors.New("no script for address")
				}
				return script, nil
			}
			return nil, errors.New("no script for address")
		})

		// SigHashSingle inputs can only be signed if there's a
		// corresponding output. However this could be already signed,
		// so we always verify the output.
		if (hashType&txscript.SigHashSingle) !=
			txscript.SigHashSingle || i < len(tx.TxOut) {

			script, err := txscript.SignTxOutput(chainParams,
				tx, i, prevOutScript, hashType, getKey,
				getScript, txIn.SignatureScript)
			// Failure to sign isn't an error, it just means that
			// the tx isn't complete.
			if err != nil {
				signErrors = append(signErrors, SignatureError{
					InputIndex: uint32(i),
					Error:      err,
				})
				continue
			}
			txIn.SignatureScript = script
		}

		// Either it was already signed or we just signed it.
		// Find out if it is completely satisfied or still needs more.
		vm, err := txscript.NewEngine(prevOutScript, tx, i,
			txscript.StandardVerifyFlags, nil, nil, 0)
		if err == nil {
			err = vm.Execute()
		}
		if err != nil {
			signErrors = append(signErrors, SignatureError{
				InputIndex: uint32(i),
				Error:      err,
			})
		}
	}
	//return nil
	//})
	return signErrors, err
}

func SignTransaction(signTransactionParams *adaptor.SignTransactionParams, netID int) (string, error) {
	//check empty string
	if "" == signTransactionParams.TransactionHex {
		return "", errors.New("Params error : NO TransactionHex.")
	}

	//chainnet
	realNet := GetNet(netID)

	var err error
	//sign the UTXO hash, must know RedeemHex which contains in RawTxInput
	var rawInputs []RawTxInput
	for {
		//decode Transaction hexString to bytes
		rawTXBytes, err1 := hex.DecodeString(signTransactionParams.TransactionHex)
		if err1 != nil {
			err = err1
			break
		}
		//deserialize to MsgTx
		var tx wire.MsgTx
		err = tx.Deserialize(bytes.NewReader(rawTXBytes))
		if err != nil {
			break
		}

		payScript := ""
		if "" != signTransactionParams.FromAddr {
			address, err := btcutil.DecodeAddress(signTransactionParams.FromAddr, realNet)
			if err != nil {
				break
			}
			// Create a public key script that pays to the address.
			script, err := txscript.PayToAddrScript(address)
			if err != nil {
				break
			}
			payScript = hex.EncodeToString(script)
		} //todo redeem
		//multisig transaction need redeem for sign
		for i, txinOne := range tx.TxIn {
			if "" == signTransactionParams.FromAddr {
				if i >= len(signTransactionParams.InputRedeemIndex) {
					err = errors.New("RedeemIndex not enough")
					break
				}
				//decode redeem's hexString to bytes
				if signTransactionParams.InputRedeemIndex[i] >= len(signTransactionParams.RedeemHex) {
					err = errors.New("RedeemIndex invalid")
					break
				}
				redeem, err := hex.DecodeString(signTransactionParams.RedeemHex[signTransactionParams.InputRedeemIndex[i]])
				if err != nil {
					break
				}
				//get multisig payScript
				scriptAddr, err := btcutil.NewAddressScriptHash(redeem, realNet)
				if err != nil {
					break
				}
				scriptPkScript, err := txscript.PayToAddrScript(scriptAddr)
				if err != nil {
					break
				}
				payScript = hex.EncodeToString(scriptPkScript)
			}
			rawInput := RawTxInput{
				txinOne.PreviousOutPoint.Hash.String(), //txid
				txinOne.PreviousOutPoint.Index,         //outindex
				payScript}                              //multisig pay script
			rawInputs = append(rawInputs, rawInput)
		}

		break
	}
	if err != nil {
		return "", err
	}

	//
	var cmd SignRawTransactionCmd
	cmd.RawTx = signTransactionParams.TransactionHex
	cmd.Inputs = &rawInputs
	cmd.RedeemHex = append(cmd.RedeemHex, signTransactionParams.RedeemHex...)
	cmd.PrivKeys = &signTransactionParams.Privkeys
	flags := "ALL"
	cmd.Flags = &flags

	//if complete ruturn true
	result, err := signRawTransactionCmd(&cmd, realNet)
	if err != nil {
		return "", err
	}

	//result for return
	signRawResult := result.(btcjson.SignRawTransactionResult)
	var signTransactionResult adaptor.SignTransactionResult
	signTransactionResult.TransactionHex = signRawResult.Hex
	signTransactionResult.Complete = signRawResult.Complete

	jsonResult, err := json.Marshal(signTransactionResult)
	if err != nil {
		return "", err
	}

	return string(jsonResult), nil
}

type SendTransactionHttppResponse struct {
	//Status string `json:"status"`
	Data struct {
		Network string `json:"network"`
		Txid    string `json:"txid"`
	} `json:"data"`
}

func SendTransactionHttp(sendTransactionParams *adaptor.SendTransactionHttpParams, netID int) (string, error) {
	//check empty string
	if "" == sendTransactionParams.TransactionHex {
		return "", errors.New("Params error : NO TransactionHex.")
	}

	var request string
	if netID == NETID_MAIN {
		request = base + "send_tx/BTC/"
	} else {
		request = base + "send_tx/BTCTEST/"
	}

	//
	params := map[string]string{"tx_hex": sendTransactionParams.TransactionHex}
	paramsJson, err := json.Marshal(params)
	if err != nil {
		return "", err
	}

	strRespose, err, _ := httpPost(request, string(paramsJson))
	if err != nil {
		return "", err
	}

	var txResult SendTransactionHttppResponse
	err = json.Unmarshal([]byte(strRespose), &txResult)
	if err != nil {
		return "", err
	}

	//result for return
	var sendTransactionResult adaptor.SendTransactionHttpResult
	sendTransactionResult.TransactionHah = txResult.Data.Txid

	jsonResult, err := json.Marshal(sendTransactionResult)
	if err != nil {
		return "", err
	}

	return string(jsonResult), nil
}

func SendTransaction(params *adaptor.SendTransactionParams, rpcParams *RPCParams) string {
	//check empty string
	if "" == params.TransactionHex {
		return "Params error : NO TransactionHex."
	}

	//decode Transaction hexString to bytes
	rawTXBytes, err := hex.DecodeString(params.TransactionHex)
	if err != nil {
		return err.Error()
	}
	//deserialize to MsgTx
	var tx wire.MsgTx
	err = tx.Deserialize(bytes.NewReader(rawTXBytes))
	if err != nil {
		return err.Error()
	}

	//get rpc client
	client, err := GetClient(rpcParams)
	if err != nil {
		return err.Error()
	}
	defer client.Shutdown()

	//send to network
	hashTX, err := client.SendRawTransaction(&tx, false)
	if err != nil {
		return err.Error()
	}

	//result for return
	var sendTransactionResult adaptor.SendTransactionResult
	sendTransactionResult.TransactionHah = hashTX.String()

	jsonResult, err := json.Marshal(sendTransactionResult)
	if err != nil {
		return err.Error()
	}

	return string(jsonResult)
}

func SignTxSend(signTxSendParams *adaptor.SignTxSendParams, rpcParams *RPCParams, netID int) (string, error) {
	//check empty string
	if "" == signTxSendParams.TransactionHex {
		return "", errors.New("Params error : NO TransactionHex.")
	}

	//chainnet
	realNet := GetNet(netID)

	var err error
	//sign the UTXO hash, must know RedeemHex which contains in RawTxInput
	var rawInputs []RawTxInput
	for {
		//decode Transaction hexString to bytes
		rawTXBytes, err := hex.DecodeString(signTxSendParams.TransactionHex)
		if err != nil {
			break
		}
		//deserialize to MsgTx
		var tx wire.MsgTx
		err = tx.Deserialize(bytes.NewReader(rawTXBytes))
		if err != nil {
			break
		}

		payScript := ""
		if "" != signTxSendParams.FromAddr {
			address, err := btcutil.DecodeAddress(signTxSendParams.FromAddr, realNet)
			if err != nil {
				break
			}
			// Create a public key script that pays to the address.
			script, err := txscript.PayToAddrScript(address)
			if err != nil {
				break
			}
			payScript = hex.EncodeToString(script)
		} //todo redeem
		//multisig transaction need redeem for sign
		for i, txinOne := range tx.TxIn {
			if "" == signTxSendParams.FromAddr {
				if i >= len(signTxSendParams.InputRedeemIndex) {
					err = errors.New("RedeemIndex not enough")
					break
				}
				//decode redeem's hexString to bytes
				if signTxSendParams.InputRedeemIndex[i] >= len(signTxSendParams.RedeemHex) {
					err = errors.New("RedeemIndex invalid")
					break
				}
				redeem, err := hex.DecodeString(signTxSendParams.RedeemHex[signTxSendParams.InputRedeemIndex[i]])
				if err != nil {
					break
				}
				//get multisig payScript
				scriptAddr, err := btcutil.NewAddressScriptHash(redeem, realNet)
				if err != nil {
					break
				}
				scriptPkScript, err := txscript.PayToAddrScript(scriptAddr)
				if err != nil {
					break
				}
				payScript = hex.EncodeToString(scriptPkScript)
			}
			rawInput := RawTxInput{
				txinOne.PreviousOutPoint.Hash.String(), //txid
				txinOne.PreviousOutPoint.Index,         //outindex
				payScript}                              //multisig pay script
			rawInputs = append(rawInputs, rawInput)
		}

		break
	}
	if err != nil {
		return "", err
	}

	//
	var cmd SignRawTransactionCmd
	cmd.RawTx = signTxSendParams.TransactionHex
	cmd.Inputs = &rawInputs
	cmd.RedeemHex = append(cmd.RedeemHex, signTxSendParams.RedeemHex...)
	cmd.PrivKeys = &signTxSendParams.Privkeys
	flags := "ALL"
	cmd.Flags = &flags

	//if complete ruturn true
	result, err := signRawTransactionCmd(&cmd, realNet)
	if err != nil {
		return "", err
	}

	//result for return
	signRawResult := result.(btcjson.SignRawTransactionResult)
	var signTxSendResult adaptor.SignTxSendResult
	if signRawResult.Complete {
		//get rpc client
		client, err := GetClient(rpcParams)
		if err != nil {
			return "", err
		}
		defer client.Shutdown()

		//decode Transaction hexString to bytes
		rawTXBytes, err := hex.DecodeString(signRawResult.Hex)
		if err != nil {
			return "", err
		}
		//deserialize to MsgTx
		var resultTX wire.MsgTx
		err = resultTX.Deserialize(bytes.NewReader(rawTXBytes))
		if err != nil {
			return "", err
		}

		//send to network
		hashTX, err := client.SendRawTransaction(&resultTX, false)
		if err != nil {
			return "", err
		}
		signTxSendResult.TransactionHah = hashTX.String()

		//SerializeSize transaction to bytes
		bufTX := bytes.NewBuffer(make([]byte, 0, resultTX.SerializeSize()))
		if err := resultTX.Serialize(bufTX); err != nil {
			return "", err
		}

		signTxSendResult.TransactionHex = hex.EncodeToString(bufTX.Bytes())
		signTxSendResult.Complete = true
	}

	jsonResult, err := json.Marshal(signTxSendResult)
	if err != nil {
		return "", err
	}

	return string(jsonResult), nil
}

func MergeTransaction(mergeTransactionParams *adaptor.MergeTransactionParams, netID int) (string, error) {
	//check empty string
	if 0 == len(mergeTransactionParams.MergeTransactionHexs) {
		return "", errors.New("Params error : NO Merge TransactionHexs.")
	}

	//chainnet
	realNet := GetNet(netID)

	//deal user tx
	var tx wire.MsgTx
	var err error
	var redeem []byte
	var addresses []btcutil.Address
	var nrequired int
	var scriptPkScript []byte
	for {
		if 0 == len(mergeTransactionParams.RedeemHex) {
			err = errors.New("RedeemHex's length is 0")
			break
		}

		//decode Transaction hexString to bytes
		rawTXBytes, err := hex.DecodeString(mergeTransactionParams.UserTransactionHex)
		if err != nil {
			break
		}
		//deserialize to MsgTx
		err = tx.Deserialize(bytes.NewReader(rawTXBytes))
		if err != nil {
			break
		}

		break
	}
	if err != nil {
		return "", err
	}

	//deal merge txs
	var txs []wire.MsgTx
	for i := range mergeTransactionParams.MergeTransactionHexs {
		var tx wire.MsgTx
		//decode Transaction hexString to bytes
		rawTXBytes, err := hex.DecodeString(mergeTransactionParams.MergeTransactionHexs[i])
		if err != nil {
			continue
		}
		//deserialize to MsgTx
		err = tx.Deserialize(bytes.NewReader(rawTXBytes))
		if err != nil {
			continue
		}
		txs = append(txs, tx)
	}
	if len(txs) == 0 {
		return "", errors.New("Params error : All Merge TransactionHexs is invalid.")
	}

	//merge tx
	complete := true
	for i := range tx.TxIn {
		if i >= len(mergeTransactionParams.InputRedeemIndex) {
			err = errors.New("RedeemIndex not enough")
			break
		}
		//decode redeem's hexString to bytes
		if mergeTransactionParams.InputRedeemIndex[i] >= len(mergeTransactionParams.RedeemHex) {
			err = errors.New("RedeemIndex invalid")
			break
		}
		//decode redeem's hexString to bytes
		redeem, err = hex.DecodeString(mergeTransactionParams.RedeemHex[mergeTransactionParams.InputRedeemIndex[i]])
		if err != nil {
			break
		}

		//get addresses an n of multisig redeem
		_, addresses, nrequired, err = txscript.ExtractPkScriptAddrs(redeem,
			realNet)
		if err != nil {
			break
		}

		//get multisig payScript
		scriptAddr, err := btcutil.NewAddressScriptHash(redeem, realNet)
		if err != nil {
			break
		}
		scriptPkScript, err = txscript.PayToAddrScript(scriptAddr)
		if err != nil {
			break
		}

		//
		sigScripts := make([][]byte, 0)
		for j := range txs {
			if i < len(txs[j].TxIn) {
				sigScripts = append(sigScripts, txs[j].TxIn[i].SignatureScript)
			}
		}

		//
		script, doneSigs := txscript.MergeMultiSigScript(&tx, i, addresses, nrequired, redeem, sigScripts)
		if doneSigs > 0 {
			tx.TxIn[i].SignatureScript = script
		}

		// Either it was already signed or we just signed it.
		// Find out if it is completely satisfied or still needs more.
		vm, err := txscript.NewEngine(scriptPkScript, &tx, i,
			txscript.StandardVerifyFlags, nil, nil, 0)
		if err == nil {
			err = vm.Execute()
		}
		if err != nil {
			complete = false
		}
	}

	//SerializeSize transaction to bytes
	buf := bytes.NewBuffer(make([]byte, 0, tx.SerializeSize()))
	if err := tx.Serialize(buf); err != nil {
		return "", err
	}
	//result for return
	var mergeTransactionResult adaptor.MergeTransactionResult
	mergeTransactionResult.TransactionHex = hex.EncodeToString(buf.Bytes())
	mergeTransactionResult.TransactionHash = tx.TxHash().String()
	mergeTransactionResult.Complete = complete

	jsonResult, err := json.Marshal(mergeTransactionResult)
	if err != nil {
		return "", err
	}

	return string(jsonResult), nil
}

func SignMessage(signMessageParams *adaptor.SignMessageParams) (string, error) {
	wif, err := btcutil.DecodeWIF(signMessageParams.Privkey)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	wire.WriteVarString(&buf, 0, "Bitcoin Signed Message:\n")
	wire.WriteVarString(&buf, 0, signMessageParams.Message)
	messageHash := chainhash.DoubleHashB(buf.Bytes())
	sigbytes, err := btcec.SignCompact(btcec.S256(), wif.PrivKey,
		messageHash, true)
	if err != nil {
		return "", err
	}

	//result for return
	var mergeTransactionResult adaptor.SignMessageResult
	mergeTransactionResult.Signature = base64.StdEncoding.EncodeToString(sigbytes)

	jsonResult, err := json.Marshal(mergeTransactionResult)
	if err != nil {
		return "", err
	}

	return string(jsonResult), nil
}

func VerifyMessage(verifyMessageParams *adaptor.VerifyMessageParams, netID int) (string, error) {
	realNet := GetNet(netID)

	// Decode the provided address.
	addr, err := btcutil.DecodeAddress(verifyMessageParams.Address, realNet)
	if err != nil {
		return "", errors.New("Invalid address or key: " + err.Error())
	}

	// Only P2PKH addresses are valid for signing.
	if _, ok := addr.(*btcutil.AddressPubKeyHash); !ok {
		return "", errors.New("Address is not a pay-to-pubkey-hash address")
	}

	// Decode base64 signature.
	sig, err := base64.StdEncoding.DecodeString(verifyMessageParams.Signature)
	if err != nil {
		return "", errors.New("Malformed base64 encoding: " + err.Error())
	}

	// Validate the signature - this just shows that it was valid at all.
	// we will compare it with the key next.
	var buf bytes.Buffer
	wire.WriteVarString(&buf, 0, "Bitcoin Signed Message:\n")
	wire.WriteVarString(&buf, 0, verifyMessageParams.Message)
	expectedMessageHash := chainhash.DoubleHashB(buf.Bytes())
	pk, wasCompressed, err := btcec.RecoverCompact(btcec.S256(), sig,
		expectedMessageHash)
	if err != nil {
		// Mirror Bitcoin Core behavior, which treats error in
		// RecoverCompact as invalid signature.
		return "", errors.New("RecoverCompact failed: " + err.Error())
	}

	// Reconstruct the pubkey hash.
	var serializedPK []byte
	if wasCompressed {
		serializedPK = pk.SerializeCompressed()
	} else {
		serializedPK = pk.SerializeUncompressed()
	}
	address, err := btcutil.NewAddressPubKey(serializedPK, realNet)
	if err != nil {
		// Again mirror Bitcoin Core behavior, which treats error in public key
		// reconstruction as invalid signature.
		return "", errors.New("AddressPubKey failed: " + err.Error())
	}

	//result for return
	var verifyMessageResult adaptor.VerifyMessageResult
	verifyMessageResult.Valid = (address.EncodeAddress() == verifyMessageParams.Address) // Return boolean if addresses match.

	jsonResult, err := json.Marshal(verifyMessageResult)
	if err != nil {
		return "", err
	}

	return string(jsonResult), nil
}

//==== ==== ==== ==== ==== ==== ==== ==== ==== ==== ==== ==== ==== ==== ==== ===

type addressToKey struct {
	key        *btcec.PrivateKey
	compressed bool
}

//find the privatekey by address
func mkGetKey(keys map[string]addressToKey) txscript.KeyDB {
	if keys == nil {
		return txscript.KeyClosure(func(addr btcutil.Address) (*btcec.PrivateKey, bool, error) {
			return nil, false, errors.New("nope")
		})
	}
	return txscript.KeyClosure(func(addr btcutil.Address) (*btcec.PrivateKey, bool, error) {
		a2k, ok := keys[addr.EncodeAddress()]
		if !ok {
			return nil, false, errors.New("nope")
		}
		return a2k.key, a2k.compressed, nil
	})
}

//find the redeemhex by address
func mkGetScript(scripts map[string][]byte) txscript.ScriptDB {
	if scripts == nil {
		return txscript.ScriptClosure(func(addr btcutil.Address) ([]byte, error) {
			return nil, errors.New("nope")
		})
	}
	return txscript.ScriptClosure(func(addr btcutil.Address) ([]byte, error) {
		script, ok := scripts[addr.EncodeAddress()]
		if !ok {
			return nil, errors.New("nope")
		}
		return script, nil
	})
}

//if complete, ruturn nil
func checkScripts(tx *wire.MsgTx, idx int, inputAmt int64,
	sigScript, scriptPkScript []byte) error {
	vm, err := txscript.NewEngine(scriptPkScript, tx, idx,
		txscript.ScriptBip16|txscript.ScriptVerifyDERSignatures,
		nil, nil, inputAmt)
	if err != nil {
		return err
	}

	err = vm.Execute()
	if err != nil {
		return err
	}

	return nil
}

// one input, one output, signed one by one.
// if not complete, return signedTransaction partSigedScript and false,
// if complete, ruturn signedTransaction lastSigedScript and true.
func MultisignOneByOne(prevTxHash string, index uint,
	amount int64, fee int64, recvAddress string,
	redeem string, partSigedScript string,
	wifKey string, netID int) (signedTransaction, newSigedScript string, complete bool) {
	//chainnet
	realNet := GetNet(netID)

	//
	hash, _ := chainhash.NewHashFromStr(prevTxHash)
	outPoint := wire.NewOutPoint(hash, uint32(index))
	//
	txIn := wire.NewTxIn(outPoint, nil, nil)
	inputs := []*wire.TxIn{txIn}

	//
	var recvAmount = amount - fee
	addr, err := btcutil.DecodeAddress(recvAddress, realNet)
	if err != nil {
		return "", "", false
	}
	pubkeyScript, _ := txscript.PayToAddrScript(addr)
	//
	outputs := []*wire.TxOut{}
	outputs = append(outputs, wire.NewTxOut(recvAmount, pubkeyScript))

	//
	tx := &wire.MsgTx{
		Version:  1,
		TxIn:     inputs,
		TxOut:    outputs,
		LockTime: 0,
	}

	//
	var sigOldBytes []byte
	if partSigedScript == "" {
		sigOldBytes = nil
	} else {
		sigOldBytes, err = hex.DecodeString(partSigedScript)
		if err != nil {
			return "", "", false
		}
	}

	//
	key, err := btcutil.DecodeWIF(wifKey)
	//
	pub, err := btcutil.NewAddressPubKey(key.SerializePubKey(), realNet)
	if err != nil {
		return "", "", false
	}

	//
	pkScript, err := hex.DecodeString(redeem)
	//
	scriptAddr, err := btcutil.NewAddressScriptHash(pkScript, realNet)
	if err != nil {
		return "", "", false
	}

	//
	scriptPkScript, err := txscript.PayToAddrScript(scriptAddr)
	if err != nil {
		return "", "", false
	}

	// Two part multisig, sign with one key then the other.
	// Sign with the other key and merge
	sigScript, err := txscript.SignTxOutput(realNet,
		tx, 0, scriptPkScript, txscript.SigHashAll,
		mkGetKey(map[string]addressToKey{
			pub.EncodeAddress(): {key.PrivKey, true},
		}), mkGetScript(map[string][]byte{
			scriptAddr.EncodeAddress(): pkScript,
		}), sigOldBytes)
	if err != nil {
		return "", "", false
	}
	tx.TxIn[0].SignatureScript = sigScript

	//
	buf := bytes.NewBuffer(make([]byte, 0, tx.SerializeSize()))
	if err := tx.Serialize(buf); err != nil {
		return "", "", false
	}

	//
	err = checkScripts(tx, 0, amount, sigScript, scriptPkScript)
	if err != nil {
		complete = false
	} else {
		complete = true
	}

	return hex.EncodeToString(buf.Bytes()), hex.EncodeToString(sigScript), complete
}
