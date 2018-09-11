package tokenengine

import (
	"github.com/palletone/go-palletone/common"
	//"github.com/btcsuite/btcd/btcec"
        "fmt"
	"crypto/ecdsa"
	"errors"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/tokenengine/internal/txscript"
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

//生成多签用的赎回脚本
//Generate redeem script
func GenerateRedeemScript(needed byte, pubKeys [][]byte) []byte {
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

//validate this transaction and input index script can unlock the utxo.
func ScriptValidate(utxoLockScript []byte, tx *modules.Transaction, msgIdx, inputIndex int) error {
	vm, err := txscript.NewEngine(utxoLockScript, tx, msgIdx, inputIndex, txscript.StandardVerifyFlags, nil, nil, 0)
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
func SignTxAllPaymentInput(tx *modules.Transaction, utxoLockScripts map[modules.OutPoint][]byte, redeemScript []byte,privKeys map[common.Address]*ecdsa.PrivateKey) error {
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
	for i, msg := range tx.TxMessages {
		if msg.App == modules.APP_PAYMENT {
			pay , ok:= msg.Payload.(*modules.PaymentPayload)
			if !ok {
				fmt.Println("Get Payment payload error:")
			} else {
	 			fmt.Println("Payment payload:", pay)
	 		}
			for j, input := range pay.Input {
				utxoLockScript , _:= utxoLockScripts[*input.PreviousOutPoint]
				checkscript := make([]byte, len(utxoLockScript))
	 			copy(checkscript, utxoLockScript)

				sigScript, err := txscript.SignTxOutput(tx, i, j, utxoLockScript, txscript.SigHashAll,
					txscript.KeyClosure(lookupKey),  txscript.ScriptClosure(lookupRedeemScript), input.SignatureScript)
				if err != nil {
					return err
				}
				input.SignatureScript = sigScript
				checkscript = nil
			}
		}
	}
	return nil
}

//传入一个脚本二进制，解析为可读的文本形式
func DisasmString(script []byte) (string, error) {
	return txscript.DisasmString(script)
}
func IsUnspendable(script []byte) bool {
	return txscript.IsUnspendable(script)
}
