package tokenengine

import (
	"testing"

	//"github.com/palletone/go-palletone/tokenengine/btcd/chaincfg/chainhash"
	"github.com/palletone/go-palletone/common"

	"github.com/palletone/go-palletone/dag/modules"
	"encoding/hex"
	"github.com/palletone/go-palletone/common/crypto"
)

func TestSignATx(t *testing.T) {
	tx:=buildSampleTx()
	privKeyBytes, _ := hex.DecodeString("2BE3B4B671FF5B8009E6876CCCC8808676C1C279EE824D0AB530294838DC1644")
	privKey,_:=crypto.ToECDSA(privKeyBytes)
	addr:=crypto.PubkeyToAddress( &privKey.PublicKey)

	utxoLock:=GenerateLockScript(addr)
	signResult,err:= SignOnePaymentInput(tx,0,0,utxoLock,privKey)
	if err!=nil{
		t.Errorf("Sign error:%s",err)
	}
	t.Logf("Sign Result:%x",signResult)

}
func buildSampleTx() *modules.Transaction {
	tx := modules.Transaction{
		TxMessages: make([]*modules.Message,1),
	}
	payment := modules.PaymentPayload{}
	utxoTxId, _ := common.NewHashFromStr("5651870aa8c894376dbd960a22171d0ad7be057a730e14d7103ed4a6dbb34873")
	outPoint := modules.NewOutPoint(utxoTxId, 0, 0)
	txIn:=modules.NewTxIn(outPoint, []byte{})
	payment.AddTxIn(*txIn)
	pubKeyHash, _ := hex.DecodeString("540ad1b52601df08b6b43cc52808c97c4901351d")
	lockScript := GenerateP2PKHLockScript(pubKeyHash)
	payment.AddTxOut(*modules.NewTxOut(1, lockScript, modules.Asset{}))
	tx.TxMessages = append(tx.TxMessages, modules.NewMessage(modules.APP_PAYMENT, payment))
	return &tx
}
