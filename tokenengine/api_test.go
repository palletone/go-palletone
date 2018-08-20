package tokenengine

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/btcsuite/btcd/btcec"
	"github.com/palletone/go-palletone/tokenengine/btcd/chaincfg"
	"github.com/palletone/go-palletone/tokenengine/btcd/chaincfg/chainhash"
	"github.com/palletone/go-palletone/tokenengine/btcd/txscript"
	"github.com/palletone/go-palletone/tokenengine/btcd/wire"
	"github.com/palletone/go-palletone/tokenengine/btcutil"
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

func TestSign2InputTx(t *testing.T) {
	// privKeyB, _ := hex.DecodeString("BF6BC3F19A3BCED41DCAFAF952B5779DAE42722B8F9A660592832090CC40CC28")
	// privKey, pubKey := btcec.PrivKeyFromBytes(btcec.S256(), privKeyB)
	// t.Logf("PrivateKey:%s", privKey)
	// t.Logf("PUbKey:%s", pubKey)
	// tx := build2InputTx()
	// t.Logf("Tx:%x", tx)

	// Ordinarily the private key would come from whatever storage mechanism
	// is being used, but for this example just hard code it.
	privKeyBytes, err := hex.DecodeString("22a47fa09a223f2aa079edf85a7c2" +
		"d4f8720ee63e502ee2869afab7de234b80c")
	if err != nil {
		fmt.Println(err)
		return
	}
	privKey, pubKey := btcec.PrivKeyFromBytes(btcec.S256(), privKeyBytes)
	pubKeyHash := btcutil.Hash160(pubKey.SerializeCompressed())
	addr, err := btcutil.NewAddressPubKeyHash(pubKeyHash,
		&chaincfg.MainNetParams)
	if err != nil {
		fmt.Println(err)
		return
	}

	// For this example, create a fake transaction that represents what
	// would ordinarily be the real transaction that is being spent.  It
	// contains a single output that pays to address in the amount of 1 BTC.
	originTx := wire.NewMsgTx(wire.TxVersion)
	prevOut := wire.NewOutPoint(&chainhash.Hash{}, ^uint32(0))
	txIn := wire.NewTxIn(prevOut, []byte{txscript.OP_0, txscript.OP_0}, nil)
	originTx.AddTxIn(txIn)
	pkScript, err := txscript.PayToAddrScript(addr)
	if err != nil {
		fmt.Println(err)
		return
	}
	txOut := wire.NewTxOut(100000000, pkScript)
	originTx.AddTxOut(txOut)
	originTxHash := originTx.TxHash()

	// Create the transaction to redeem the fake transaction.
	redeemTx := wire.NewMsgTx(wire.TxVersion)

	// Add the input(s) the redeeming transaction will spend.  There is no
	// signature script at this point since it hasn't been created or signed
	// yet, hence nil is provided for it.
	prevOut = wire.NewOutPoint(&originTxHash, 0)
	txIn = wire.NewTxIn(prevOut, nil, nil)
	redeemTx.AddTxIn(txIn)

	// Ordinarily this would contain that actual destination of the funds,
	// but for this example don't bother.
	txOut = wire.NewTxOut(0, nil)
	redeemTx.AddTxOut(txOut)

	// Sign the redeeming transaction.
	lookupKey := func(a btcutil.Address) (*btcec.PrivateKey, bool, error) {
		// Ordinarily this function would involve looking up the private
		// key for the provided address, but since the only thing being
		// signed in this example uses the address associated with the
		// private key from above, simply return it with the compressed
		// flag set since the address is using the associated compressed
		// public key.
		//
		// NOTE: If you want to prove the code is actually signing the
		// transaction properly, uncomment the following line which
		// intentionally returns an invalid key to sign with, which in
		// turn will result in a failure during the script execution
		// when verifying the signature.
		//
		// privKey.D.SetInt64(12345)
		//
		return privKey, true, nil
	}
	// Notice that the script database parameter is nil here since it isn't
	// used.  It must be specified when pay-to-script-hash transactions are
	// being signed.
	sigScript, err := txscript.SignTxOutput(&chaincfg.MainNetParams,
		redeemTx, 0, originTx.TxOut[0].PkScript, txscript.SigHashAll,
		txscript.KeyClosure(lookupKey), nil, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	redeemTx.TxIn[0].SignatureScript = sigScript

	// Prove that the transaction has been validly signed by executing the
	// script pair.
	flags := txscript.ScriptBip16 | txscript.ScriptVerifyDERSignatures |
		txscript.ScriptStrictMultiSig |
		txscript.ScriptDiscourageUpgradableNops
	vm, err := txscript.NewEngine(originTx.TxOut[0].PkScript, redeemTx, 0,
		flags, nil, nil, -1)
	if err != nil {
		fmt.Println(err)
		return
	}
	if err := vm.Execute(); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Transaction successfully signed")

}

func TestBuild2InputTx(t *testing.T) {
	//https://testnet.blockchain.info/tx/f0d9d482eb122535e32a3ae92809dd87839e63410d5fd52816fc9fc6215018cc?show_adv=true

	tx := wire.NewMsgTx(wire.TxVersion)
	//https://testnet.blockchain.info/tx-index/239152566/1  0.4BTC
	utxoHash, _ := chainhash.NewHashFromStr("1dda832890f85288fec616ef1f4113c0c86b7bf36b560ea244fd8a6ed12ada52")
	point := wire.OutPoint{Hash: *utxoHash, Index: 1}
	tx.AddTxIn(wire.NewTxIn(&point, nil, nil))
	//24f284aed2b9dbc19f0d435b1fe1ee3b3ddc763f28ca28bad798d22b6bea0c66  1.1BTC
	utxoHash2, _ := chainhash.NewHashFromStr("24f284aed2b9dbc19f0d435b1fe1ee3b3ddc763f28ca28bad798d22b6bea0c66")
	point2 := wire.OutPoint{Hash: *utxoHash2, Index: 1}
	tx.AddTxIn(wire.NewTxIn(&point2, nil, nil))

	//找零0.2991024 BTC
	pubKeyHash, _ := hex.DecodeString("b5407cec767317d41442aab35bad2712626e17ca")
	lock, _ := txscript.NewScriptBuilder().AddOp(txscript.OP_DUP).AddOp(txscript.OP_HASH160).
		AddData(pubKeyHash).AddOp(txscript.OP_EQUALVERIFY).AddOp(txscript.OP_CHECKSIG).
		Script()
	tx.AddTxOut(wire.NewTxOut(29910240, lock))

	pubKeyHash2, _ := hex.DecodeString("be09abcbfda1f2c26899f062979ab0708731235a")
	lock2, _ := txscript.NewScriptBuilder().AddOp(txscript.OP_DUP).AddOp(txscript.OP_HASH160).
		AddData(pubKeyHash2).AddOp(txscript.OP_EQUALVERIFY).AddOp(txscript.OP_CHECKSIG).
		Script()
	tx.AddTxOut(wire.NewTxOut(120000000, lock2))

	//txscript.SignTxOutput(&chaincfg.TestNet3Params,tx,0,pkScript,txscript.SigHashAll,)

	// Sign the redeeming transaction.
	lookupKey := func(a btcutil.Address) (*btcec.PrivateKey, bool, error) {

		privKeyBytes, err := hex.DecodeString("2BE3B4B671FF5B8009E6876CCCC8808676C1C279EE824D0AB530294838DC1644")
		if err != nil {
			fmt.Println(err)
		}
		privKey, _ := btcec.PrivKeyFromBytes(btcec.S256(), privKeyBytes)
		return privKey, true, nil
	}
	// Notice that the script database parameter is nil here since it isn't
	// used.  It must be specified when pay-to-script-hash transactions are
	// being signed.
	pubKeyHash3, _ := hex.DecodeString("b5407cec767317d41442aab35bad2712626e17ca")
	pkScript, _ := txscript.NewScriptBuilder().AddOp(txscript.OP_DUP).AddOp(txscript.OP_HASH160).
		AddData(pubKeyHash3).AddOp(txscript.OP_EQUALVERIFY).AddOp(txscript.OP_CHECKSIG).
		Script()

	sigScript, err := txscript.SignTxOutput(&chaincfg.TestNet3Params,
		tx, 0, pkScript, txscript.SigHashAll,
		txscript.KeyClosure(lookupKey), nil, nil)
	if err != nil {
		fmt.Println(err)

	}

	sigScript2, err := txscript.SignTxOutput(&chaincfg.TestNet3Params,
		tx, 1, pkScript, txscript.SigHashAll,
		txscript.KeyClosure(lookupKey), nil, nil)
	if err != nil {
		fmt.Println(err)

	}
	t.Logf("Raw Txhash is:%s\n", tx.TxHash().String())
	tx.TxIn[0].SignatureScript = sigScript
	t.Logf("Sign:%x\n", sigScript)

	tx.TxIn[1].SignatureScript = sigScript2

	t.Logf("Sign2:%x\n", sigScript2)
	t.Logf("Txhash is:%s\n", tx.TxHash().String())
	buf := bytes.NewBuffer(make([]byte, 0, tx.SerializeSize()))
	_ = tx.Serialize(buf)
	t.Logf("RawTX:%x\n", buf.Bytes())

}
