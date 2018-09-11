package tokenengine

import (
	"testing"

	//"github.com/palletone/go-palletone/tokenengine/btcd/chaincfg/chainhash"
	"github.com/palletone/go-palletone/common"

	"crypto/ecdsa"
	"encoding/hex"
	"github.com/palletone/go-palletone/common/crypto"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/stretchr/testify/assert"
	"github.com/palletone/go-palletone/tokenengine/internal/txscript"
	"github.com/palletone/go-palletone/dag/asset"
	"fmt"
)

func TestGetAddressFromScript(t *testing.T) {
	addrStr := "P1JEStL6tb7TB8e6ZJSpJhQoqin2A6pabdA"
	addr, _ := common.StringToAddress(addrStr)
	p2pkhLock := GenerateP2PKHLockScript(addr.Bytes())
	getAddr, _ := GetAddressFromScript(p2pkhLock)
	t.Logf("Get Address:%s", getAddr.Str())
	assert.True(t, getAddr == addr, "Address parse error")

	addr2, _ := common.StringToAddress("P35SbSqXuXcHrtZuJKzbStpcqzwCg88jXfn")
	p2shLock := GenerateP2SHLockScript(addr2.Bytes())
	getAddr2, _ := GetAddressFromScript(p2shLock)
	t.Logf("Get Script Address:%s", getAddr2.Str())
	assert.True(t, getAddr2 == addr2, "Address parse error")

}

func TestSignAndVerifyATx(t *testing.T) {

	privKeyBytes, _ := hex.DecodeString("2BE3B4B671FF5B8009E6876CCCC8808676C1C279EE824D0AB530294838DC1644")
	privKey, _ := crypto.ToECDSA(privKeyBytes)
	pubKey := privKey.PublicKey
	pubKeyBytes := crypto.CompressPubkey(&pubKey)
	pubKeyHash := crypto.Hash160(pubKeyBytes)
	t.Logf("Public Key:%x", pubKeyBytes)
	addr := crypto.PubkeyToAddress(&privKey.PublicKey)
	t.Logf("Addr:%s", addr.String())
	lockScript := GenerateP2PKHLockScript(pubKeyHash)
	t.Logf("UTXO lock script:%x", lockScript)

	tx := &modules.Transaction{
		TxMessages: make([]*modules.Message, 0),
	}
	payment := &modules.PaymentPayload{}
	utxoTxId, _ := common.NewHashFromStr("5651870aa8c894376dbd960a22171d0ad7be057a730e14d7103ed4a6dbb34873")
	outPoint := modules.NewOutPoint(utxoTxId, 0, 0)
	txIn := modules.NewTxIn(outPoint, []byte{})
	payment.AddTxIn(*txIn)
	asset0:=&modules.Asset{}
	payment.AddTxOut(*modules.NewTxOut(1, lockScript, asset0))
	payment2 := &modules.PaymentPayload{}
	utxoTxId2, _ := common.NewHashFromStr("1651870aa8c894376dbd960a22171d0ad7be057a730e14d7103ed4a6dbb34873")
	outPoint2 := modules.NewOutPoint(utxoTxId2, 1, 1)
	txIn2 := modules.NewTxIn(outPoint2, []byte{})
	payment2.AddTxIn(*txIn2)
	asset1:=&modules.Asset{AssetId:asset.NewAsset()}
	payment2.AddTxOut(*modules.NewTxOut(1, lockScript, asset1))
	tx.TxMessages = append(tx.TxMessages, modules.NewMessage(modules.APP_PAYMENT, payment))
	tx.TxMessages = append(tx.TxMessages, modules.NewMessage(modules.APP_PAYMENT, payment2))

	tx.TxMessages = append(tx.TxMessages, modules.NewMessage(modules.APP_TEXT, &modules.TextPayload{Text: []byte("Hello PalletOne")}))

	//signResult, err := SignOnePaymentInput(tx, 0, 0, lockScript, privKey)
	//if err != nil {
	//	t.Errorf("Sign error:%s", err)
	//}
	//t.Logf("Sign Result:%x", signResult)
	//t.Logf("msg len:%d", len(tx.TxMessages))
	//tx.TxMessages[0].Payload.(*modules.PaymentPayload).Input[0].SignatureScript = signResult
	//
	//signResult2, err := SignOnePaymentInput(tx, 1, 0, lockScript, privKey)
	//tx.TxMessages[1].Payload.(*modules.PaymentPayload).Input[0].SignatureScript = signResult2
	lockScripts := map[modules.OutPoint][]byte{
		*outPoint:  lockScript[:],
		*outPoint2: GenerateP2PKHLockScript(pubKeyHash),
	}
	privKeys := map[common.Address]*ecdsa.PrivateKey{
		addr: privKey,
	}
	err := SignTxAllPaymentInput(tx, lockScripts, privKeys)
	if err != nil {
		t.Logf("Sign error:%s", err)
	}
	err = ScriptValidate(lockScript,  tx, 0,0)
	if err != nil {
		t.Logf("validate error:%s", err)
	}
	// textPayload :=tx.TxMessages[2].Payload.(*modules.TextPayload)
	//textPayload.Text=[]byte("Bad")
	//fmt.Printf("%s", tx.TxMessages[2].Payload.(*modules.TextPayload))

	err = ScriptValidate(lockScript, tx, 1,0)
	assert.Nil(t,err,fmt.Sprintf( "validate error:%s",err))

}


var (
	prvKey1,_ =crypto.FromWIF("KwN8TdhAMeU8b9UrEYTNTVEvDsy9CSyepwRVNEy2Fc9nbGqDZw4J") //"0454b0699a590b6fc8e66e81db1ca36e99d7c767cdfe44a217b6e105c5db97d5" //P1QJNzZhqGoxNL2igkdthNBQLNWdNGTWzQU
	pubKey1B,_=hex.DecodeString("02f9286c44fe7ebff9788425d5025ad764cdf5aec3daef862d143c5f09134d75b0")
	address1,_=common.StringToAddress("P1QJNzZhqGoxNL2igkdthNBQLNWdNGTWzQU")

	prvKey2,_ =crypto.FromWIF("Ky7gQF2rxXLjGSymCtCMa67N2YMt98fRgxyy5WfH92FpbWDxWVRM") //"3892859c02b1be2ce494e61c60181051d79ff21dca22fae1dc349887335b6676" //P1N4nEffoUskPrbnoEqBR69JQDX2vv9vYa8\
	pubKey2B,_=hex.DecodeString("02a2ba6f2a6e1467334d032ec54ac862c655d7e8bd6bbbce36c771fcdc0ddfb01f")
	address2,_=common.StringToAddress("P1N4nEffoUskPrbnoEqBR69JQDX2vv9vYa8")

	prvKey3,_ =crypto.FromWIF("KzRHTanikQgR5oqUts69JTrCXRuy9Zod5qXdnAbYwvUnuUDJ3Rro")//"5f7754e5407fc2a81f453645cbd92878a6341d30dbfe2e680fc81628d47e8023" //P1MzuBUT7ubGpkAFqUB6chqTSXmBThQv2HT
	pubKey3B,_=hex.DecodeString("020945d0c9ed05cf5ca9fe38dde799d7b73f4a5bfb71fc1a3c1dca79e2d86462a7")
	address3,_=common.StringToAddress("P1MzuBUT7ubGpkAFqUB6chqTSXmBThQv2HT")
	)
func build23Address() ([]byte,[]byte,string) {

	redeemScript:= GenerateRedeemScript(2,[][]byte{pubKey1B,pubKey2B,pubKey3B})
	lockScript := GenerateP2SHLockScript(crypto.Hash160(redeemScript))
	addressMulti,_:=GetAddressFromScript(lockScript)

	return  lockScript,redeemScript, addressMulti.Str()
	}

//构造一个2/3签名的地址和UTXO，然后用其中的2个私钥对其进行签名
func TestMultiSign1Step(t *testing.T)  {
	lockScript,redeemScript,addressMulti:= build23Address()
	t.Logf("MultiSign Address:%s\n",addressMulti)
	t.Logf("RedeemScript: %x\n",redeemScript)
	t.Logf("RedeemScript: %d\n",redeemScript)
	tx := &modules.Transaction{
		TxMessages: make([]*modules.Message, 0),
	}
	asset0:=&modules.Asset{}
	payment := &modules.PaymentPayload{}
	utxoTxId, _ := common.NewHashFromStr("1111870aa8c894376dbd960a22171d0ad7be057a730e14d7103ed4a6dbb34873")
	outPoint := modules.NewOutPoint(utxoTxId, 0, 0)
	txIn := modules.NewTxIn(outPoint, []byte{})
	payment.AddTxIn(*txIn)
	p1lockScript := GenerateP2PKHLockScript(crypto.Hash160(pubKey1B))
	payment.AddTxOut(*modules.NewTxOut(1, p1lockScript, asset0))
	tx.TxMessages = append(tx.TxMessages, modules.NewMessage(modules.APP_PAYMENT, payment))
	//scriptCp:=make([]byte,len(lockScript))
	//copy(scriptCp,lockScript)
	privKeys := map[common.Address]*ecdsa.PrivateKey{
		address1: prvKey1,
		address2: prvKey2,
	}

	sign12,err := MultiSignOnePaymentInput(tx,0,0, lockScript,redeemScript, privKeys,nil)
	if err != nil {
		t.Logf("Sign error:%s", err)
	}
	t.Logf("PrvKey1&2 sign result:%x\n",sign12)
	pay1:=tx.TxMessages[0].Payload.(*modules.PaymentPayload)
	pay1.Input[0].SignatureScript=sign12
	str,_:= txscript.DisasmString(sign12)
	t.Logf("Signed script:{%s}",str)

	err=ScriptValidate(lockScript,tx,0,0)
	assert.Nil(t,err,fmt.Sprintf( "validate error:%s",err))
}

//构造一个2/3签名的地址和UTXO，然后用其中的2个私钥分两步对其进行签名
func TestMultiSign2Step(t *testing.T)  {
	lockScript,redeemScript,addressMulti:= build23Address()
	t.Logf("MultiSign Address:%s\n",addressMulti)
	t.Logf("RedeemScript: %x\n",redeemScript)
	t.Logf("RedeemScript: %d\n",redeemScript)
	tx := &modules.Transaction{
		TxMessages: make([]*modules.Message, 0),
	}
	payment := &modules.PaymentPayload{}
	utxoTxId, _ := common.NewHashFromStr("1111870aa8c894376dbd960a22171d0ad7be057a730e14d7103ed4a6dbb34873")
	outPoint := modules.NewOutPoint(utxoTxId, 0, 0)
	txIn := modules.NewTxIn(outPoint, []byte{})
	payment.AddTxIn(*txIn)
	p1lockScript := GenerateP2PKHLockScript(crypto.Hash160(pubKey1B))
	asset0:=&modules.Asset{}
	payment.AddTxOut(*modules.NewTxOut(1, p1lockScript, asset0))
	tx.TxMessages = append(tx.TxMessages, modules.NewMessage(modules.APP_PAYMENT, payment))
	//scriptCp:=make([]byte,len(lockScript))
	//copy(scriptCp,lockScript)
	privKeys := map[common.Address]*ecdsa.PrivateKey{
		address1: prvKey1,
	}
	sign1,err := MultiSignOnePaymentInput(tx,0,0, lockScript,redeemScript, privKeys,nil)
	if err != nil {
		t.Logf("Sign error:%s", err)
	}
	t.Logf("PrvKey1 sign result:%x\n",sign1)
	pay1:=tx.TxMessages[0].Payload.(*modules.PaymentPayload)
	pay1.Input[0].SignatureScript=sign1

	privKeys2 := map[common.Address]*ecdsa.PrivateKey{
		address2: prvKey2,
	}
	//scriptCp2:=make([]byte,len(lockScript))
	//copy(scriptCp2,lockScript)
	sign2,err := MultiSignOnePaymentInput(tx,0,0, lockScript,redeemScript, privKeys2,sign1)
	if err != nil {
		t.Logf("Sign error:%s", err)
	}
	t.Logf("PrvKey2 sign result:%x\n",sign2)

	pay1:=tx.TxMessages[0].Payload.(*modules.PaymentPayload)
	pay1.Input[0].SignatureScript=sign2
	str,_:= txscript.DisasmString(sign2)
	t.Logf("Signed script:{%s}",str)

	err=ScriptValidate(lockScript,tx,0,0)
	assert.Nil(t,err,fmt.Sprintf( "validate error:%s",err))
}
