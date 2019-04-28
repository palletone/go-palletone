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
	"bytes"
	"errors"
	"sort"

	"fmt"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/crypto"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/tokenengine/internal/txscript"
)

const (
	SigHashOld          uint32 = 0x0
	SigHashAll          uint32 = 0x1
	SigHashNone         uint32 = 0x2
	SigHashSingle       uint32 = 0x3
	SigHashAnyOneCanPay uint32 = 0x80
	// sigHashMask defines the number of bits of the hash type which is used
	// to identify which outputs are signed.
	sigHashMask = 0x1f
)

type AddressGetSign func(common.Address, []byte) ([]byte, error)
type AddressGetPubKey func(common.Address) ([]byte, error)

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
	return addrs[0].Address, nil
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
func GenerateP2CHUnlockScript(signs [][]byte, redeemScript []byte) []byte {
	builder := txscript.NewScriptBuilder()
	for _, sign := range signs {
		builder = builder.AddData(sign)
	}
	unlock, _ := builder.AddData(redeemScript).Script()
	return unlock
}

//validate this transaction and input index script can unlock the utxo.
func ScriptValidate(utxoLockScript []byte, pickupJuryRedeemScript txscript.PickupJuryRedeemScript, tx *modules.Transaction, msgIdx, inputIndex int) error {
	acc := &account{}
	txCopy := tx
	if tx.IsContractTx() {
		isRequestMsg := false
		for idx, msg := range tx.TxMessages {
			if msg.App.IsRequest() {
				isRequestMsg = true
			}
			if idx == msgIdx && !isRequestMsg {
				txCopy = tx.GetRequestTx()
			}
		}
	}
	vm, err := txscript.NewEngine(utxoLockScript, pickupJuryRedeemScript, txCopy, msgIdx, inputIndex, txscript.StandardVerifyFlags, nil, acc)
	if err != nil {
		log.Error("Failed to create script: ", err)
		return err
	}
	return vm.Execute()
}

//对交易中的Payment类型中的某个Input生成解锁脚本
//func SignOnePaymentInput(tx *modules.Transaction, msgIdx, id int, utxoLockScript []byte, privKey *ecdsa.PrivateKey, juryVersion int) ([]byte, error) {
//	lookupKey := func(a common.Address) (*ecdsa.PrivateKey, bool, error) {
//		return privKey, true, nil
//	}
//	sigScript, err := txscript.SignTxOutput(tx, msgIdx, id, utxoLockScript, txscript.SigHashAll,
//		txscript.KeyClosure(lookupKey), nil, nil, juryVersion)
//	if err != nil {
//		return []byte{}, err
//	}
//	return sigScript, nil
//}
func MultiSignOnePaymentInput(tx *modules.Transaction, hashType uint32, msgIdx, id int, utxoLockScript []byte, redeemScript []byte,
	pubKeyFn AddressGetPubKey, hashFn AddressGetSign, previousScript []byte) ([]byte, error) {

	lookupRedeemScript := func(a common.Address) ([]byte, error) {
		return redeemScript, nil
	}
	tmpAcc := &account{pubKeyFn: pubKeyFn, signFn: hashFn}
	sigScript, err := txscript.SignTxOutput(tx, msgIdx, id, utxoLockScript, txscript.SigHashType(hashType),
		tmpAcc, txscript.ScriptClosure(lookupRedeemScript), previousScript)
	if err != nil {
		return []byte{}, err
	}
	return sigScript, nil
}

type account struct {
	pubKeyFn AddressGetPubKey
	signFn   AddressGetSign
}

func (a *account) Hash(msg []byte) ([]byte, error) {
	return crypto.Keccak256(msg), nil
}
func (a *account) Sign(address common.Address, digest []byte) ([]byte, error) {
	return a.signFn(address, digest)
}
func (a *account) Verify(pubKey, signature, digest []byte) (bool, error) {
	log.Debugf("Pubkey:%x,Signature:%x,hash:%x", pubKey, signature, digest)
	return crypto.VerifySignature(pubKey, digest, signature), nil
}
func (a *account) GetPubKey(address common.Address) ([]byte, error) {
	return a.pubKeyFn(address)
}

//func (a *account) GetSignFunction(addr common.Address) txscript.SignHash {
//	signFn := func(hash []byte) ([]byte, error) {
//		return a.hashFn(addr, hash)
//	}
//	return signFn
//}
//func (a *account) GetPubKey(addr common.Address) ([]byte, error) {
//	return a.pubKeyFn(addr)
//}

//为钱包计算要签名某个Input对应的Hash
func CalcSignatureHash(tx *modules.Transaction, hashType uint32, msgIdx, inputIdx int, lockOrRedeemScript []byte) ([]byte, error) {
	acc := &account{}
	return txscript.CalcSignatureHash(lockOrRedeemScript, txscript.SigHashType(hashType), tx, msgIdx, inputIdx, acc)
}

//Sign a full transaction
func SignTxAllPaymentInput(tx *modules.Transaction, hashType uint32, utxoLockScripts map[modules.OutPoint][]byte,
	redeemScript []byte, pubKeyFn AddressGetPubKey, hashFn AddressGetSign) ([]common.SignatureError, error) {

	lookupRedeemScript := func(a common.Address) ([]byte, error) {

		return redeemScript, nil
	}
	tmpAcc := &account{pubKeyFn: pubKeyFn, signFn: hashFn}
	var signErrors []common.SignatureError
	for i, msg := range tx.TxMessages {
		if msg.App == modules.APP_PAYMENT {
			pay, ok := msg.Payload.(*modules.PaymentPayload)
			if !ok {
				return nil, errors.New("Invalid payment message")
			}
			for j, input := range pay.Inputs {
				utxoLockScript, find := utxoLockScripts[*input.PreviousOutPoint]
				if !find {
					errMsg := fmt.Sprintf("Don't find utxo for outpoint[%s]", input.PreviousOutPoint.String())
					log.Error(errMsg)
					return nil, errors.New(errMsg)
				}
				checkscript := make([]byte, len(utxoLockScript))
				copy(checkscript, utxoLockScript)
				if (hashType&uint32(txscript.SigHashSingle)) != uint32(txscript.SigHashSingle) || j < len(pay.Outputs) {
					sigScript, err := txscript.SignTxOutput(tx, i, j, utxoLockScript, txscript.SigHashType(hashType),
						tmpAcc, txscript.ScriptClosure(lookupRedeemScript),
						input.SignatureScript)
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
