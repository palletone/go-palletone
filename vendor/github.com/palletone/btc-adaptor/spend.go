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
package btcadaptor

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"

	"github.com/palletone/adaptor"
	"github.com/palletone/btc-adaptor/txscript"
)

func HashMessage(input *adaptor.HashMessageInput) (*adaptor.HashMessageOutput, error) {
	var output adaptor.HashMessageOutput
	output.Hash = chainhash.DoubleHashB(input.Message)

	return &output, nil
}

func SignMessage(input *adaptor.SignMessageInput) (*adaptor.SignMessageOutput, error) {
	priKey, _ := btcec.PrivKeyFromBytes(btcec.S256(), input.PrivateKey)

	var buf bytes.Buffer
	wire.WriteVarString(&buf, 0, "Bitcoin Signed Message:\n")
	wire.WriteVarString(&buf, 0, string(input.Message))
	messageHash := chainhash.DoubleHashB(buf.Bytes())
	sigbytes, err := btcec.SignCompact(btcec.S256(), priKey, messageHash, true)
	if err != nil {
		return nil, err
	}

	//result for return
	var output adaptor.SignMessageOutput
	output.Signature = []byte(base64.StdEncoding.EncodeToString(sigbytes))

	return &output, nil
}

func VerifySignature(input *adaptor.VerifySignatureInput) (*adaptor.VerifySignatureOutput, error) {
	// Decode base64 signature.
	sig, err := base64.StdEncoding.DecodeString(string(input.Signature))
	if err != nil {
		return nil, errors.New("Malformed base64 encoding: " + err.Error())
	}

	// Validate the signature - this just shows that it was valid at all.
	// we will compare it with the key next.
	var buf bytes.Buffer
	wire.WriteVarString(&buf, 0, "Bitcoin Signed Message:\n")
	wire.WriteVarString(&buf, 0, string(input.Message))
	expectedMessageHash := chainhash.DoubleHashB(buf.Bytes())
	pk, _, err := btcec.RecoverCompact(btcec.S256(), sig, expectedMessageHash)
	if err != nil {
		// Mirror Bitcoin Core behavior, which treats error in
		// RecoverCompact as invalid signature.
		return nil, errors.New("RecoverCompact failed: " + err.Error())
	}

	// Reconstruct the pubkey hash.
	serializedPK := pk.SerializeCompressed()

	//result for return
	var output adaptor.VerifySignatureOutput
	output.Pass = bytes.Equal(serializedPK, input.PublicKey) // Return boolean if addresses match.

	return &output, nil
}

// signatureError records the underlying error when validating a transaction input signature.
type signatureError struct {
	InputIndex uint32
	Error      error
}

func signTransactionReal(tx *wire.MsgTx, hashType txscript.SigHashType,
	additionalPrevScripts map[wire.OutPoint][]byte,
	additionalKeysByAddress map[string]*btcutil.WIF,
	p2shRedeemScriptsByAddress map[string][]byte, chainParams *chaincfg.Params) []signatureError {

	signErrors := []signatureError{}
	//var signErrors []SignatureErroerr := walletdb.View(w.db, func(dbtx walletdb.ReadTx) error {
	//addrmgrNs := dbtx.ReadBucket(waddrmgrNamespaceKey)
	//txmgrNs := dbtx.ReadBucket(wtxmgrNamespaceKey)
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
				signErrors = append(signErrors, signatureError{
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
			signErrors = append(signErrors, signatureError{
				InputIndex: uint32(i),
				Error:      err,
			})
		}
	}
	//return nil
	//})
	return signErrors
}

func SignTransaction(input *adaptor.SignTransactionInput, netID int) (*adaptor.SignTransactionOutput, error) {
	//check empty
	if 0 == len(input.Transaction) {
		return nil, errors.New("the Transaction is empty")
	}
	if 0 == len(input.PrivateKey) {
		return nil, errors.New("the PrivateKey is empty")
	}
	extraLen := len(input.Extra)
	if 0 == extraLen {
		return nil, errors.New("the Extra is empty, must be oneSigAddr or multiSigRedeem")
	}

	//chainnet
	realNet := GetNet(netID)

	keys := map[string]*btcutil.WIF{}
	priKey, _ := btcec.PrivKeyFromBytes(btcec.S256(), input.PrivateKey)
	wif, err := btcutil.NewWIF(priKey, realNet, true)
	if err != nil {
		return nil, fmt.Errorf("NewWIF failed : %s", err.Error())
	}
	addr, err := btcutil.NewAddressPubKey(wif.SerializePubKey(), realNet)
	if err != nil {
		return nil, err
	}
	addrStr := addr.EncodeAddress()
	keys[addrStr] = wif

	//deserialize to MsgTx
	var tx wire.MsgTx
	err = tx.Deserialize(bytes.NewReader(input.Transaction))
	if err != nil {
		return nil, fmt.Errorf("Deserialize tx failed : %s", err.Error())
	}

	//sign the UTXO hash, must know RedeemHex which contains in RawTxInput
	isRedeem := false
	if extraLen > 35 {
		isRedeem = true
	}
	scripts := make(map[string][]byte)
	var scriptPkScript []byte
	if isRedeem {
		redeem, err := hex.DecodeString(string(input.Extra))
		if err != nil {
			return nil, fmt.Errorf("hex.DecodeString redeem in the Extra failed : %s", err.Error())
		}
		//get multisig payScript
		scriptAddr, err := btcutil.NewAddressScriptHash(redeem, realNet)
		if err != nil {
			return nil, fmt.Errorf("NewAddressScriptHash redeem failed : %s", err.Error())
		}
		scripts[scriptAddr.String()] = redeem
		scriptPkScript, err = txscript.PayToAddrScript(scriptAddr)
		if err != nil {
			return nil, fmt.Errorf("PayToAddrScript redeem failed : %s", err.Error())
		}
	} else {
		if addrStr != string(input.Extra) {
			return nil, fmt.Errorf("address in the Extra is not match with the PrivateKey")
		}
		address, err := btcutil.DecodeAddress(string(input.Extra), realNet)
		if err != nil {
			return nil, fmt.Errorf("DecodeAddress oneAddr failed : %s", err.Error())
		}
		// Create a public key script that pays to the address.
		scriptPkScript, err = txscript.PayToAddrScript(address)
		if err != nil {
			return nil, fmt.Errorf("PayToAddrScript oneAddr failed : %s", err.Error())
		}
	}

	inputs := make(map[wire.OutPoint][]byte)
	for _, txinOne := range tx.TxIn {
		//fmt.Println(txinOne.PreviousOutPoint.Hash.String(), txinOne.PreviousOutPoint.Index) //Debug
		inputs[wire.OutPoint{Hash: txinOne.PreviousOutPoint.Hash, Index: txinOne.PreviousOutPoint.Index}] = scriptPkScript
	}

	signErrs := signTransactionReal(&tx, txscript.SigHashAll, inputs, keys, scripts, realNet)
	if !isRedeem && len(signErrs) != 0 {
		return nil, fmt.Errorf("signTransactionReal failed : not Complete")
	}

	var buf bytes.Buffer
	buf.Grow(tx.SerializeSize())
	if err = tx.Serialize(&buf); err != nil {
		return nil, err
	}

	var signatures []byte
	for _, txinOne := range tx.TxIn {
		//fmt.Printf("%x\n", txinOne.SignatureScript) //Debug
		signatures = append(signatures, txinOne.SignatureScript...) //todo [][]byte ?
	}

	var output adaptor.SignTransactionOutput
	output.Signature = signatures
	output.SignedTx = buf.Bytes()
	//output.Extra

	return &output, nil
}

//type SendTransactionHttppResponse struct {
//	//Status string `json:"status"`
//	Data struct {
//		Network string `json:"network"`
//		Txid    string `json:"txid"`
//	} `json:"data"`
//}
//
//func SendTransactionHttp(input *adaptor.SendTransactionInput, netID int) (*adaptor.SendTransactionOutput, error) {
//	//check empty string
//	if 0 == len(input.Transaction) {
//		return nil, errors.New("the Transaction is empty")
//	}
//
//	var request string
//	if netID == NETID_MAIN {
//		request = base + "send_tx/BTC/"
//	} else {
//		request = base + "send_tx/BTCTEST/"
//	}
//
//	//
//	params := map[string]string{"tx_hex": hex.EncodeToString(input.Transaction)}
//	paramsJson, err := json.Marshal(params)
//	if err != nil {
//		return nil, err
//	}
//
//	strRespose, err, _ := httpPost(request, string(paramsJson))
//	if err != nil {
//		return nil, err
//	}
//
//	var txResult SendTransactionHttppResponse
//	err = json.Unmarshal([]byte(strRespose), &txResult)
//	if err != nil {
//		return nil, err
//	}
//
//	//result for return
//	var output adaptor.SendTransactionOutput
//	txID, _ := hex.DecodeString(txResult.Data.Txid)
//	output.TxID = txID
//
//	return &output, nil
//}

func SendTransaction(input *adaptor.SendTransactionInput, rpcParams *RPCParams) (*adaptor.SendTransactionOutput, error) {
	//check empty string
	if 0 == len(input.Transaction) {
		return nil, errors.New("the Transaction is empty")
	}

	//deserialize to MsgTx
	var tx wire.MsgTx
	err := tx.Deserialize(bytes.NewReader(input.Transaction))
	if err != nil {
		return nil, fmt.Errorf("Deserialize failed : %s", err.Error())
	}

	//get rpc client
	client, err := GetClient(rpcParams)
	if err != nil {
		return nil, err
	}
	defer client.Shutdown()

	//send to network
	hashTX, err := client.SendRawTransaction(&tx, false) //BTC API
	if err != nil {
		return nil, fmt.Errorf("SendRawTransaction failed : %s", err.Error())
	}

	//result for return
	var ouput adaptor.SendTransactionOutput
	txID, _ := hex.DecodeString(hashTX.String())
	ouput.TxID = txID

	return &ouput, nil
}

func BindTxAndSignature(input *adaptor.BindTxAndSignatureInput, netID int) (*adaptor.BindTxAndSignatureOutput, error) {
	//check empty string
	if 0 == len(input.SignedTxs) {
		return nil, errors.New("Params error : NO Merge TransactionHexs.")
	}
	if 0 == len(input.Extra) {
		return nil, errors.New("the Extra is empty, must be multiSigRedeem")
	}

	//chainnet
	realNet := GetNet(netID)

	//decode redeem's hexString to bytes
	redeem, err := hex.DecodeString(string(input.Extra))
	if err != nil {
		return nil, fmt.Errorf("hex.DecodeString redeem in the Extra failed : %s", err.Error())
	}
	//get addresses an n of multisig redeem
	_, addresses, nrequired, err := txscript.ExtractPkScriptAddrs(redeem, realNet)
	if err != nil {
		return nil, fmt.Errorf("ExtractPkScriptAddrs redeem failed : %s", err.Error())
	}
	//get multisig payScript
	scriptAddr, err := btcutil.NewAddressScriptHash(redeem, realNet)
	if err != nil {
		return nil, fmt.Errorf("NewAddressScriptHash redeem failed : %s", err.Error())
	}
	scriptPkScript, err := txscript.PayToAddrScript(scriptAddr)
	if err != nil {
		return nil, fmt.Errorf("PayToAddrScript redeem failed : %s", err.Error())
	}

	//deserialize to MsgTx
	var tx wire.MsgTx
	err = tx.Deserialize(bytes.NewReader(input.Transaction))
	if err != nil {
		return nil, fmt.Errorf("Deserialize tx failed : %s", err.Error())
	}

	//deal merge txs
	var txs []wire.MsgTx
	for i := range input.SignedTxs {
		var tx wire.MsgTx
		//deserialize to MsgTx
		err = tx.Deserialize(bytes.NewReader(input.SignedTxs[i]))
		if err != nil {
			continue
		}
		txs = append(txs, tx)
	}
	if len(txs) == 0 {
		return nil, errors.New("Params error : All Merge TransactionHexs is invalid.")
	}

	//merge txs
	for i := range tx.TxIn {
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
			return nil, fmt.Errorf("signTransactionReal failed : not Complete")
		}
	}

	//SerializeSize transaction to bytes
	buf := bytes.NewBuffer(make([]byte, 0, tx.SerializeSize()))
	if err := tx.Serialize(buf); err != nil {
		return nil, fmt.Errorf("Serialize tx failed : %s", err.Error())
	}
	//result for return
	var output adaptor.BindTxAndSignatureOutput
	output.SignedTx = buf.Bytes()
	//output.Extra

	return &output, nil
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
func checkScripts(tx *wire.MsgTx, idx int, inputAmt int64, scriptPkScript []byte) error {
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
	err = checkScripts(tx, 0, amount, scriptPkScript)
	if err != nil {
		complete = false
	} else {
		complete = true
	}

	return hex.EncodeToString(buf.Bytes()), hex.EncodeToString(sigScript), complete
}
