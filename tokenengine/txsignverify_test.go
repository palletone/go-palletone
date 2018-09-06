package tokenengine

import (
	"testing"

	//"github.com/palletone/go-palletone/tokenengine/btcd/chaincfg/chainhash"
	"github.com/palletone/go-palletone/common"

	"github.com/palletone/go-palletone/dag/modules"
	"encoding/hex"
	"github.com/palletone/go-palletone/common/crypto"

)

func TestSignAndVerifyATx(t *testing.T) {

	privKeyBytes, _ := hex.DecodeString("2BE3B4B671FF5B8009E6876CCCC8808676C1C279EE824D0AB530294838DC1644")
	privKey, _ := crypto.ToECDSA(privKeyBytes)
	pubKey := privKey.PublicKey
	pubKeyBytes := crypto.CompressPubkey(&pubKey)
	pubKeyHash := crypto.Hash160(pubKeyBytes)
	t.Logf("Public Key:%x",pubKeyBytes)
	addr := crypto.PubkeyToAddress(&privKey.PublicKey)
	t.Logf("Addr:%s", addr.String())
	lockScript := GenerateP2PKHLockScript(pubKeyHash)
	t.Logf("UTXO lock script:%x", lockScript)

	tx := &modules.Transaction{
		TxMessages: make([]*modules.Message, 0),
	}
	payment := modules.PaymentPayload{}
	utxoTxId, _ := common.NewHashFromStr("5651870aa8c894376dbd960a22171d0ad7be057a730e14d7103ed4a6dbb34873")
	outPoint := modules.NewOutPoint(utxoTxId, 0, 0)
	txIn := modules.NewTxIn(outPoint, []byte{})
	payment.AddTxIn(*txIn)

	payment.AddTxOut(*modules.NewTxOut(1, lockScript, modules.Asset{}))
	tx.TxMessages = append(tx.TxMessages, modules.NewMessage(modules.APP_PAYMENT, payment))
	tx.TxMessages = append(tx.TxMessages, modules.NewMessage(modules.APP_TEXT, modules.TextPayload{Text: []byte("Hello PalletOne"),}))

	signResult, err := SignOnePaymentInput(tx, 0, 0, lockScript, privKey)
	if err != nil {
		t.Errorf("Sign error:%s", err)
	}
	t.Logf("Sign Result:%x", signResult)
	t.Logf("msg len:%d", len(tx.TxMessages))
	tx.TxMessages[0].Payload.(modules.PaymentPayload).Input[0].SignatureScript = signResult


	err = ScriptValidate(GenerateP2PKHLockScript(pubKeyHash), 100, tx, 0, 0)
	if err != nil {
		t.Logf("validate error:%s", err)
	}
	t.Logf("Good! all validated")
}
