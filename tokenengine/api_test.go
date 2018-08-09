package tokenengine

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/palletone/go-palletone/tokenengine/btcd/chaincfg/chainhash"
	"github.com/palletone/go-palletone/tokenengine/btcd/txscript"
	"github.com/palletone/go-palletone/tokenengine/btcd/wire"
)

func TestValidateTokenPayment(t *testing.T) {
	//https://www.blockchain.com/zh-cn/btc/tx/e34fc0ecc23bd60392e5e80c1fd2d8e87a3e4fd50e8fbd0af65bfc7d5bf3e1e4
	pubKeyHash, _ := hex.DecodeString("8c7130fa0c30e3f777310a3aa571e529c1d3e15f")
	// lock, _ := txscript.NewScriptBuilder().AddOp(txscript.OP_DUP).AddOp(txscript.OP_HASH160).
	// 	AddData(pubKeyHash).AddOp(txscript.OP_EQUALVERIFY).AddOp(txscript.OP_CHECKSIG).
	// 	Script() //https://www.blockchain.com/zh-cn/btc/tx/5651870aa8c894376dbd960a22171d0ad7be057a730e14d7103ed4a6dbb34873
	// fmt.Println(lock)

	lock := GenerateP2PKHLockScript(pubKeyHash)
	tx := buildTx()
	//hashCache := txscript.NewTxSigHashes(tx)
	amount := int64(31980) //0.0003198 BTC
	// vm, err := txscript.NewEngine(lock, tx, 0, txscript.StandardVerifyFlags, nil, nil, amount)
	// if err != nil {
	// 	fmt.Errorf("Failed to create script: %v", err)
	// }
	// e := vm.Execute()
	e := ScriptValidate(lock, amount, tx, 0)
	if e != nil {
		t.Fatalf("vm execute error:%s\n", e)
	} else {
		t.Log("Scuess execute script!")
	}
}

func buildTx() *wire.MsgTx {
	sign, _ := hex.DecodeString("3045022100cac2be20a3e6667057583d97a49a89225aa242567bcfcc1a619060e292dad3d0022065a04f0720e42b7d18ff840226d25b262a68a50dd18ab87e4d111b173e85cec401")
	pubKey, _ := hex.DecodeString("0206e6e881f5d183cfa4b868dbe7b7eac26569c16df5329400561f272e21f1c739")
	unlock, _ := txscript.NewScriptBuilder().AddData(sign).AddData(pubKey).Script()

	tx := wire.NewMsgTx(1)
	utxoHash, _ := chainhash.NewHashFromStr("5651870aa8c894376dbd960a22171d0ad7be057a730e14d7103ed4a6dbb34873")
	point := wire.OutPoint{Hash: *utxoHash, Index: 0}
	tx.AddTxIn(wire.NewTxIn(&point, unlock, nil))

	pubKeyHash, _ := hex.DecodeString("540ad1b52601df08b6b43cc52808c97c4901351d")
	lock, _ := txscript.NewScriptBuilder().AddOp(txscript.OP_DUP).AddOp(txscript.OP_HASH160).
		AddData(pubKeyHash).AddOp(txscript.OP_EQUALVERIFY).AddOp(txscript.OP_CHECKSIG).
		Script()
	tx.AddTxOut(wire.NewTxOut(21970, lock))

	data, _ := hex.DecodeString("e69bbee6b885e5928c2d3ee69bbee980b8e5a4ab28e69bbee78e89e5bfa0292d3ee69bbee4b8bee59bbd2d3ee69bbee6af85")
	opreturn, _ := txscript.NewScriptBuilder().AddOp(txscript.OP_RETURN).AddData(data).Script()
	tx.AddTxOut(wire.NewTxOut(1, opreturn)) //0.00000001 BTC
	buf := bytes.NewBuffer(make([]byte, 0, tx.SerializeSize()))
	_ = tx.Serialize(buf)
	fmt.Printf("RawTX:%x\n", buf.Bytes())

	//fmt.Printf("Hash:%x\n", chainhash.HashB(buf.Bytes()))
	fmt.Printf("DoubleHash:%s\n", chainhash.DoubleHashH(buf.Bytes()).String())
	fmt.Printf("TxHash:%s\n", tx.TxHash().String())
	return tx
}
