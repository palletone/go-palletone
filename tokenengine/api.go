/*
 *
 *    This file is part of go-palletone.
 *    go-palletone is free software: you can redistribute it and/or modify
 *    it under the terms of the GNU General Public License as published by
 *    the Free Software Foundation, either version 3 of the License, or
 *    (at your option) any later version.
 *    go-palletone is distributed in the hope that it will be useful,
 *    but WITHOUT ANY WARRANTY; without even the implied warranty of
 *    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *    GNU General Public License for more details.
 *    You should have received a copy of the GNU General Public License
 *    along with go-palletone.  If not, see <http://www.gnu.org/licenses/>.
 * /
 *
 *  * @author PalletOne core developer <dev@pallet.one>
 *  * @date 2018
 *
 */

package tokenengine

import (
	"crypto/ecdsa"
	"errors"
	"fmt"

	"github.com/palletone/go-palletone/common"
	//"github.com/btcsuite/btcd/btcec"
	"bytes"

	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/tokenengine/internal/txscript"
	"sort"
)

//Generate a P2PKH lock script, just only need input 20bytes public key hash.
//You can use Address.Bytes() to get address hash.
func GenerateP2PKHLockScript(pubKeyHash []byte) []byte {

	lock, _ := txscript.NewScriptBuilder().AddOp(txscript.OP_DUP).AddOp(txscript.OP_HASH160).
		AddData(pubKeyHash).AddOp(txscript.OP_EQUALVERIFY).AddOp(txscript.OP_CHECKSIG).
		Script()
	return lock
}

//Give redeem script hash 160 result, generate a P2SH lock script.
//If you have built your redeem script, please use crypto.Hash160() to gnerate hash
func GenerateP2SHLockScript(redeemScriptHash []byte) []byte {

	lock, _ := txscript.NewScriptBuilder().AddOp(txscript.OP_HASH160).
		AddData(redeemScriptHash).AddOp(txscript.OP_EQUAL).
		Script()
	return lock
}

//
func GenerateP2CHLockScript(contractId common.Address) []byte {
	lock, _ := txscript.NewScriptBuilder().AddData(contractId.Bytes()).AddOp(txscript.OP_JURY_REDEEM_EQUAL).
		Script()
	return lock
}

//根据锁定脚本获得对应的地址
func GetAddressFromScript(lockScript []byte) (common.Address, error) {

	scriptCp := make([]byte, len(lockScript))
	copy(scriptCp, lockScript)
	scriptClass, addrs, _, err := txscript.ExtractPkScriptAddrs(scriptCp)
	if err != nil {
		return common.Address{}, err
	}
	if scriptClass == txscript.NonStandardTy {
		return common.Address{}, err
	}
	if len(addrs) != 1 {
		return common.Address{}, err
	}
	return addrs[0], nil
}

type PubKey4Sort [][]byte

func (c PubKey4Sort) Len() int {
	return len(c)
}
func (c PubKey4Sort) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}
func (c PubKey4Sort) Less(i, j int) bool {
	return bytes.Compare(c[i], c[j]) > 0
}

//生成多签用的赎回脚本
//Generate redeem script
func GenerateRedeemScript(needed byte, pubKeys [][]byte) []byte {
	if needed == 0x0 {
		return []byte{}
	}
	if len(pubKeys) == 1 { //Mediator单签
		redeemScript, _ := txscript.NewScriptBuilder().AddData(pubKeys[0]).AddOp(txscript.OP_CHECKSIG).Script()
		return redeemScript
	}
	// pubkeys 排序
	sort.Sort(PubKey4Sort(pubKeys))
	builder := txscript.NewScriptBuilder().AddOp(needed + 80) //OP_Number
	for _, pubKey := range pubKeys {
		builder = builder.AddData(pubKey)
	}

	redeemScript, _ := builder.AddOp(byte(len(pubKeys) + 80)).
		AddOp(txscript.OP_CHECKMULTISIG).Script()
	return redeemScript
}

//根据地址产生对应的锁定脚本
func GenerateLockScript(address common.Address) []byte {

	//t, _ := address.Validate()
	//if t == common.PublicKeyHash {
	//	return GenerateP2PKHLockScript(address.Bytes())
	//} else {
	//	return GenerateP2SHLockScript(address.Bytes())
	//}
	script, _ := txscript.PayToAddrScript(address)
	return script
}

/*
//Give a lock script, and parse it then pick the address string out.
func PickAddress(lockscript []byte) (common.Address, error) {
	log.Debug(string(lockscript))
	//Mock
	if lockscript[0] == txscript.OP_DUP { //P2PKH
		pubKeyHash := lockscript[2:22]
		return common.PubKeyHashToAddress(pubKeyHash), nil
	} else { //P2SH
		redeemScriptHash := lockscript[1:21]
		return common.ScriptHashToAddress(redeemScriptHash), nil
	}
	//return "12gpXQVcCL2qhTNQgyLVdCFG2Qs2px98nV", nil
}*/

//根据签名和公钥信息生成解锁脚本
//Use signature and public key to generate a P2PKH unlock script
func GenerateP2PKHUnlockScript(sign []byte, pubKey []byte) []byte {
	unlock, _ := txscript.NewScriptBuilder().AddData(sign).AddData(pubKey).Script()
	return unlock
}

//根据收集到的签名和脚本生成解锁脚本
//Use collection signatures and redeem script to unlock
func GenerateP2SHUnlockScript(signs [][]byte, redeemScript []byte) []byte {
	builder := txscript.NewScriptBuilder()
	for _, sign := range signs {
		builder = builder.AddData(sign)
	}
	unlock, _ := builder.AddData(redeemScript).Script()
	return unlock
}

//根据收集到的签名和脚本生成解锁合约上的Token的脚本
func GenerateP2CHUnlockScript(signs [][]byte, redeemScript []byte, version int) []byte {
	builder := txscript.NewScriptBuilder()
	for _, sign := range signs {
		builder = builder.AddData(sign)
	}
	unlock, _ := builder.AddData(redeemScript).AddInt64(int64(version)).Script()
	return unlock
}

//validate this transaction and input index script can unlock the utxo.
func ScriptValidate(utxoLockScript []byte, pickupJuryRedeemScript txscript.PickupJuryRedeemScript, tx *modules.Transaction, msgIdx, inputIndex int) error {
	vm, err := txscript.NewEngine(utxoLockScript, pickupJuryRedeemScript, tx, msgIdx, inputIndex, txscript.StandardVerifyFlags, nil, nil, 0)
	if err != nil {
		log.Error("Failed to create script: ", err)
		return err
	}
	return vm.Execute()
}

//对交易中的Payment类型中的某个Input生成解锁脚本
func SignOnePaymentInput(tx *modules.Transaction, msgIdx, id int, utxoLockScript []byte, privKey *ecdsa.PrivateKey) ([]byte, error) {
	lookupKey := func(a common.Address) (*ecdsa.PrivateKey, bool, error) {
		return privKey, true, nil
	}
	sigScript, err := txscript.SignTxOutput(tx, msgIdx, id, utxoLockScript, txscript.SigHashAll,
		txscript.KeyClosure(lookupKey), nil, nil)
	if err != nil {
		return []byte{}, err
	}
	return sigScript, nil
}
func MultiSignOnePaymentInput(tx *modules.Transaction, msgIdx, id int, utxoLockScript []byte, redeemScript []byte, privKeys map[common.Address]*ecdsa.PrivateKey, previousScript []byte) ([]byte, error) {
	lookupKey := func(a common.Address) (*ecdsa.PrivateKey, bool, error) {
		if privKey, ok := privKeys[a]; ok {
			return privKey, true, nil
		}
		return nil, false, errors.New("PrivateKey not exist")
	}
	lookupRedeemScript := func(a common.Address) ([]byte, error) {
		return redeemScript, nil
	}
	sigScript, err := txscript.SignTxOutput(tx, msgIdx, id, utxoLockScript, txscript.SigHashAll,
		txscript.KeyClosure(lookupKey), txscript.ScriptClosure(lookupRedeemScript), previousScript)
	if err != nil {
		return []byte{}, err
	}
	return sigScript, nil
}

//Sign a full transaction
func SignTxAllPaymentInput(tx *modules.Transaction, hashType uint32, utxoLockScripts map[modules.OutPoint][]byte, redeemScript []byte, privKeys map[common.Address]*ecdsa.PrivateKey) ([]common.SignatureError, error) {
	lookupKey := func(a common.Address) (*ecdsa.PrivateKey, bool, error) {
		if privKey, ok := privKeys[a]; ok {
			return privKey, true, nil
		}
		return nil, false, nil
	}
	lookupRedeemScript := func(a common.Address) ([]byte, error) {
		//addrStr := a.String()
		//redeemScript, ok := scripts[addrStr]
		//if !ok {
		//	return nil, errors.New("no script for address")
		//}
		return redeemScript, nil
	}
	var signErrors []common.SignatureError
	for i, msg := range tx.TxMessages {
		if msg.App == modules.APP_PAYMENT {
			pay, ok := msg.Payload.(*modules.PaymentPayload)
			if !ok {
				fmt.Println("Get Payment payload error:")
			} else {
				fmt.Println("Payment payload:", pay)
			}
			for j, input := range pay.Inputs {
				utxoLockScript, _ := utxoLockScripts[*input.PreviousOutPoint]
				checkscript := make([]byte, len(utxoLockScript))
				copy(checkscript, utxoLockScript)
				if (hashType&txscript.SigHashSingle) !=
					txscript.SigHashSingle || j < len(pay.Outputs) {
					sigScript, err := txscript.SignTxOutput(tx, i, j, utxoLockScript, hashType,
						txscript.KeyClosure(lookupKey), txscript.ScriptClosure(lookupRedeemScript), input.SignatureScript)
					if err != nil {
						signErrors = append(signErrors, common.SignatureError{
							InputIndex: uint32(j),
							MsgIndex:   uint32(i),
							Error:      err,
						})
						return signErrors, err
					}
					input.SignatureScript = sigScript
					checkscript = nil
				}
			}
		}
	}
	return signErrors, nil
}

//传入一个脚本二进制，解析为可读的文本形式
func DisasmString(script []byte) (string, error) {
	return txscript.DisasmString(script)
}
func IsUnspendable(script []byte) bool {
	return txscript.IsUnspendable(script)
}
