package tokenengine

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/btcsuite/btcd/btcec"
        "github.com/palletone/go-palletone/common/rlp"
	"github.com/palletone/go-palletone/tokenengine/btcd/chaincfg"
	//"github.com/palletone/go-palletone/tokenengine/btcd/chaincfg/chainhash"
	"github.com/palletone/go-palletone/tokenengine/btcd/txscript"
	"github.com/palletone/go-palletone/tokenengine/btcutil"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/dag/modules"
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

func buildTx() *modules.PaymentPayload {
	aid := modules.IDType16{}
	aid.SetBytes([]byte("1111111111111111222222222222222222"))
	ast := modules.Asset{
		AssertId: aid,
		UniqueId: aid,
		ChainId:  1,
	}
	sign, _ := hex.DecodeString("3045022100cac2be20a3e6667057583d97a49a89225aa242567bcfcc1a619060e292dad3d0022065a04f0720e42b7d18ff840226d25b262a68a50dd18ab87e4d111b173e85cec401")
	pubKey, _ := hex.DecodeString("0206e6e881f5d183cfa4b868dbe7b7eac26569c16df5329400561f272e21f1c739")
	unlock, _ := txscript.NewScriptBuilder().AddData(sign).AddData(pubKey).Script()

	tx := modules.NewPaymentPayload()
	utxoHash, _ := common.NewHashFromStr("5651870aa8c894376dbd960a22171d0ad7be057a730e14d7103ed4a6dbb34873")
	point := modules.OutPoint{TxHash: *utxoHash, OutIndex: 0,MessageIndex:0}
	tx.AddTxIn(*modules.NewTxIn(&point, unlock))

	pubKeyHash, _ := hex.DecodeString("540ad1b52601df08b6b43cc52808c97c4901351d")
	lock, _ := txscript.NewScriptBuilder().AddOp(txscript.OP_DUP).AddOp(txscript.OP_HASH160).
		AddData(pubKeyHash).AddOp(txscript.OP_EQUALVERIFY).AddOp(txscript.OP_CHECKSIG).
		Script()
	tx.AddTxOut(*modules.NewTxOut(21970, lock,ast))

	data, _ := hex.DecodeString("e69bbee6b885e5928c2d3ee69bbee980b8e5a4ab28e69bbee78e89e5bfa0292d3ee69bbee4b8bee59bbd2d3ee69bbee6af85")
	opreturn, _ := txscript.NewScriptBuilder().AddOp(txscript.OP_RETURN).AddData(data).Script()
	tx.AddTxOut(*modules.NewTxOut(1, opreturn,ast)) //0.00000001 BTC
        fmt.Printf("-----72  tx is %+v\n",tx)
	buf := bytes.NewBuffer(make([]byte, 0, tx.SerializeSize()))
	buf.Grow(tx.SerializeSize())
	//_ = tx.Serialize(buf)
	mtxbt ,err := rlp.EncodeToBytes(buf)
	if err != nil {
		return nil
	}
	txHex := hex.EncodeToString(mtxbt)
	fmt.Println(txHex)
	fmt.Printf("RawTX:%x\n", buf.Bytes())

	//fmt.Printf("Hash:%x\n", chainhash.HashB(buf.Bytes()))
	fmt.Printf("DoubleHash:%s\n", common.DoubleHashH(buf.Bytes()).String())
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
	aid := modules.IDType16{}
	aid.SetBytes([]byte("1111111111111111222222222222222222"))
	ast := modules.Asset{
		AssertId: aid,
		UniqueId: aid,
		ChainId:  1,
	}
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
	originTx := modules.NewPaymentPayload()
	prevOut := modules.NewOutPoint(&common.Hash{}, ^uint32(0),^uint32(0))
	txIn := modules.NewTxIn(prevOut, []byte{txscript.OP_0, txscript.OP_0})
	originTx.AddTxIn(*txIn)
	pkScript, err := txscript.PayToAddrScript(addr)
	if err != nil {
		fmt.Println(err)
		return
	}
	txOut := modules.NewTxOut(100000000, pkScript,ast)
	originTx.AddTxOut(*txOut)
	originTxHash := originTx.TxHash()

	// Create the transaction to redeem the fake transaction.
	redeemTx := modules.NewPaymentPayload()

	// Add the input(s) the redeeming transaction will spend.  There is no
	// signature script at this point since it hasn't been created or signed
	// yet, hence nil is provided for it.
	prevOut = modules.NewOutPoint(&originTxHash, 0,0)
	txIn = modules.NewTxIn(prevOut, nil)
	redeemTx.AddTxIn(*txIn)

	// Ordinarily this would contain that actual destination of the funds,
	// but for this example don't bother.
	txOut = modules.NewTxOut(0, nil,ast)
	redeemTx.AddTxOut(*txOut)

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
		redeemTx, 0,0, originTx.Output[0].PkScript, txscript.SigHashAll,
		txscript.KeyClosure(lookupKey), nil, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	redeemTx.Input[0].SignatureScript = sigScript

	// Prove that the transaction has been validly signed by executing the
	// script pair.
	flags := txscript.ScriptBip16 | txscript.ScriptVerifyDERSignatures |
		txscript.ScriptStrictMultiSig |
		txscript.ScriptDiscourageUpgradableNops
	vm, err := txscript.NewEngine(originTx.Output[0].PkScript, redeemTx, 0,
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

func buildRawTx() *modules.PaymentPayload {
	//https://testnet.blockchain.info/tx/f0d9d482eb122535e32a3ae92809dd87839e63410d5fd52816fc9fc6215018cc?show_adv=true
       aid := modules.IDType16{}
	aid.SetBytes([]byte("1111111111111111222222222222222222"))
	ast := modules.Asset{
		AssertId: aid,
		UniqueId: aid,
		ChainId:  1,
	}
	tx := modules.NewPaymentPayload()
	//https://testnet.blockchain.info/tx-index/239152566/1  0.4BTC
	utxoHash, _ := common.NewHashFromStr("1dda832890f85288fec616ef1f4113c0c86b7bf36b560ea244fd8a6ed12ada52")
	point := modules.OutPoint{TxHash: *utxoHash, OutIndex: 1,MessageIndex:0}
	//构建第一个Input，指向一个0.4BTC的UTXO，第二个参数是解锁脚本，现在是nil
	tx.AddTxIn(*modules.NewTxIn(&point, nil))
	//https://testnet.blockchain.info/tx-index/239157459/1  1.1BTC
	utxoHash2, _ := common.NewHashFromStr("24f284aed2b9dbc19f0d435b1fe1ee3b3ddc763f28ca28bad798d22b6bea0c66")
	point2 := modules.OutPoint{TxHash: *utxoHash2, OutIndex: 1,MessageIndex:0}
	//构建第二个Input，指向一个1.1BTC的UTXO，第二个参数是解锁脚本，现在是nil
	tx.AddTxIn(*modules.NewTxIn(&point2, nil))

	//找零的地址（这里是16进制形式，变成Base58格式就是mx3KrUjRzzqYTcsyyvWBiHBncLrrTPXnkV）
	pubKeyHash, _ := hex.DecodeString("b5407cec767317d41442aab35bad2712626e17ca")
	lock, _ := txscript.NewScriptBuilder().AddOp(txscript.OP_DUP).AddOp(txscript.OP_HASH160).
		AddData(pubKeyHash).AddOp(txscript.OP_EQUALVERIFY).AddOp(txscript.OP_CHECKSIG).
		Script()
	//构建第一个Output，是找零0.2991024 BTC
	tx.AddTxOut(*modules.NewTxOut(29910240, lock,ast))
	//支付给了某个地址，仍然是16进制形式，Base58形式是：mxqnGTekzKqnMqNFHKYi8FhV99WcvQGhfH。
	pubKeyHash2, _ := hex.DecodeString("be09abcbfda1f2c26899f062979ab0708731235a")
	lock2, _ := txscript.NewScriptBuilder().AddOp(txscript.OP_DUP).AddOp(txscript.OP_HASH160).
		AddData(pubKeyHash2).AddOp(txscript.OP_EQUALVERIFY).AddOp(txscript.OP_CHECKSIG).
		Script()
	//构建第二个Output，支付1.2 BTC出去
	tx.AddTxOut(*modules.NewTxOut(120000000, lock2,ast))
	return tx
}

func TestBuild2InputTx(t *testing.T) {
	tx := buildRawTx()

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
	tx.Input[0].SignatureScript = sigScript
	t.Logf("Sign:%x\n", sigScript)

	tx.Input[1].SignatureScript = sigScript2

	t.Logf("Sign2:%x\n", sigScript2)
	t.Logf("Txhash is:%s\n", tx.TxHash().String())
	buf := bytes.NewBuffer(make([]byte, 0, tx.SerializeSize()))
	//                  _ = tx.Serialize(buf)
        mtxbt ,err := rlp.EncodeToBytes(buf)
	if err != nil {
		fmt.Println(err)
	}
	txHex := hex.EncodeToString(mtxbt)
	fmt.Println(txHex)
	t.Logf("RawTX:%x\n", buf.Bytes())

}
