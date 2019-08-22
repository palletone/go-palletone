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
	"encoding/hex"
	"fmt"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/crypto"
	"testing"

	"encoding/json"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/palletone/go-palletone/contracts/syscontract"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/tokenengine/internal/txscript"
	"github.com/stretchr/testify/assert"
)

func TestGetAddressFromScript(t *testing.T) {
	addrStr := "P1JEStL6tb7TB8e6ZJSpJhQoqin2A6pabdA"
	addr, _ := common.StringToAddress(addrStr)
	p2pkhLock := Instance.GenerateP2PKHLockScript(addr.Bytes())
	t.Logf("P2PKH script:%x", p2pkhLock)
	getAddr, _ := Instance.GetAddressFromScript(p2pkhLock)
	t.Logf("Get Address:%s", getAddr.Str())
	assert.True(t, getAddr == addr, "Address parse error")

	addr2, _ := common.StringToAddress("P35SbSqXuXcHrtZuJKzbStpcqzwCg88jXfn")
	p2shLock := Instance.GenerateP2SHLockScript(addr2.Bytes())
	getAddr2, _ := Instance.GetAddressFromScript(p2shLock)
	t.Logf("Get Script Address:%s", getAddr2.Str())
	assert.True(t, getAddr2 == addr2, "Address parse error")

}

func TestGenerateP2CHLockScript(t *testing.T) {
	addrStr := "PCGTta3M4t3yXu8uRgkKvaWd2d8DR32W9vM"
	addr, err := common.StringToAddress(addrStr)
	assert.Nil(t, err)
	p2ch1 := Instance.GenerateLockScript(addr)
	p2ch1Str, _ := Instance.DisasmString(p2ch1)
	t.Logf("Pay to contract hash lock script:%x, string:%s", p2ch1, p2ch1Str)
	p2ch2 := Instance.GenerateP2CHLockScript(addr)
	assert.Equal(t, p2ch1, p2ch2)
	addr2, err := Instance.GetAddressFromScript(p2ch1)
	assert.Nil(t, err, "Err must nil")
	assert.Equal(t, addr2.String(), addrStr)
	t.Logf("get address:%s", addr2.String())

	deposit := syscontract.DepositContractAddress
	script := Instance.GenerateP2CHLockScript(deposit)
	t.Logf("%x", script)
}
func TestDecodeScriptBytes(t *testing.T) {
	data, _ := hex.DecodeString("21021b11a1d070a173edc33a25ad8382602b2ad9b58670f031059b8866fc8e1b16aeac")
	str, _ := txscript.DisasmString(data)
	t.Log(str)
}
func TestSignAndVerify2PaymentTx(t *testing.T) {
	privKeyBytes, _ := hex.DecodeString("2BE3B4B671FF5B8009E6876CCCC8808676C1C279EE824D0AB530294838DC1644")

	pubKeyBytes, _ := crypto.MyCryptoLib.PrivateKeyToPubKey(privKeyBytes)

	pubKeyHash := crypto.Hash160(pubKeyBytes)
	t.Logf("Public Key:%x", pubKeyBytes)
	addr := crypto.PubkeyBytesToAddress(pubKeyBytes)
	t.Logf("Addr:%s", addr.String())
	lockScript := Instance.GenerateP2PKHLockScript(pubKeyHash)
	t.Logf("UTXO lock script:%x", lockScript)

	tx := &modules.Transaction{
		TxMessages: make([]*modules.Message, 0),
	}
	pay1 := &modules.PaymentPayload{}
	utxoTxId := common.HexToHash("5651870aa8c894376dbd960a22171d0ad7be057a730e14d7103ed4a6dbb34873")
	outPoint := modules.NewOutPoint(utxoTxId, 0, 0)
	txIn := modules.NewTxIn(outPoint, []byte{})
	pay1.AddTxIn(txIn)
	asset0 := &modules.Asset{}
	pay1.AddTxOut(modules.NewTxOut(1, lockScript, asset0))
	pay2 := &modules.PaymentPayload{}
	utxoTxId2 := common.HexToHash("1651870aa8c894376dbd960a22171d0ad7be057a730e14d7103ed4a6dbb34873")
	outPoint2 := modules.NewOutPoint(utxoTxId2, 1, 1)
	txIn2 := modules.NewTxIn(outPoint2, []byte{})
	pay2.AddTxIn(txIn2)
	asset1 := &modules.Asset{AssetId: modules.PTNCOIN}
	pay2.AddTxOut(modules.NewTxOut(1, lockScript, asset1))
	tx.TxMessages = append(tx.TxMessages, modules.NewMessage(modules.APP_PAYMENT, pay1))
	tx.TxMessages = append(tx.TxMessages, modules.NewMessage(modules.APP_PAYMENT, pay2))

	tx.TxMessages = append(tx.TxMessages, modules.NewMessage(modules.APP_DATA, &modules.DataPayload{MainData: []byte("Hello PalletOne")}))

	//signResult, err := SignOnePaymentInput(tx, 0, 0, lockScript, privKey)
	//if err != nil {
	//	t.Errorf("Sign error:%s", err)
	//}
	//t.Logf("Sign Result:%x", signResult)
	//t.Logf("msg len:%d", len(tx.TxMessages))
	//tx.TxMessages[0].Payload.(*modules.PaymentPayload).Inputs[0].SignatureScript = signResult
	//
	//signResult2, err := SignOnePaymentInput(tx, 1, 0, lockScript, privKey)
	//tx.TxMessages[1].Payload.(*modules.PaymentPayload).Input[0].SignatureScript = signResult2
	lockScripts := map[modules.OutPoint][]byte{
		*outPoint:  lockScript[:],
		*outPoint2: Instance.GenerateP2PKHLockScript(pubKeyHash),
	}
	//privKeys := map[common.Address]*ecdsa.PrivateKey{
	//	addr: privKey,
	//}
	getPubKeyFn := func(common.Address) ([]byte, error) {
		return pubKeyBytes, nil
	}
	getSignFn := func(addr common.Address, hash []byte) ([]byte, error) {
		s, e := crypto.MyCryptoLib.Sign(privKeyBytes, hash)
		return s, e
	}
	var hashtype uint32
	hashtype = 1
	_, err := Instance.SignTxAllPaymentInput(tx, hashtype, lockScripts, nil, getPubKeyFn, getSignFn)
	if err != nil {
		t.Logf("Sign error:%s", err)
	}
	unlockScript := tx.TxMessages[0].Payload.(*modules.PaymentPayload).Inputs[0].SignatureScript
	t.Logf("UnlockScript:%x", unlockScript)
	s, _ := Instance.DisasmString(unlockScript)
	t.Logf("UnlockScript string:%s", s)
	err = Instance.ScriptValidate(lockScript, nil, tx, 0, 0)
	if err != nil {
		t.Logf("validate error:%s", err)
	}
	// textPayload :=tx.TxMessages[2].Payload.(*modules.DataPayload)
	//textPayload.Text=[]byte("Bad")
	//fmt.Printf("%s", tx.TxMessages[2].Payload.(*modules.DataPayload))

	err = Instance.ScriptValidate(lockScript, nil, tx, 1, 0)
	assert.Nil(t, err, fmt.Sprintf("validate error:%s", err))

}

func TestHashNone1Payment(t *testing.T) {
	privKeyBytes, _ := hex.DecodeString("2BE3B4B671FF5B8009E6876CCCC8808676C1C279EE824D0AB530294838DC1644")
	pubKeyBytes, _ := crypto.MyCryptoLib.PrivateKeyToPubKey(privKeyBytes)

	pubKeyHash := crypto.Hash160(pubKeyBytes)
	t.Logf("Public Key:%x", pubKeyBytes)
	addr := crypto.PubkeyBytesToAddress(pubKeyBytes)
	t.Logf("Addr:%s", addr.String())
	lockScript := Instance.GenerateP2PKHLockScript(pubKeyHash)
	t.Logf("UTXO lock script:%x", lockScript)

	tx := &modules.Transaction{
		TxMessages: make([]*modules.Message, 0),
	}
	payment := &modules.PaymentPayload{}
	utxoTxId := common.HexToHash("5651870aa8c894376dbd960a22171d0ad7be057a730e14d7103ed4a6dbb34873")
	outPoint := modules.NewOutPoint(utxoTxId, 0, 0)
	txIn := modules.NewTxIn(outPoint, []byte{})
	payment.AddTxIn(txIn)
	asset0 := modules.NewPTNAsset()

	tx.TxMessages = append(tx.TxMessages, modules.NewMessage(modules.APP_PAYMENT, payment))

	//tx.TxMessages = append(tx.TxMessages, modules.NewMessage(modules.APP_DATA, &modules.DataPayload{MainData: []byte("Hello PalletOne")}))

	lockScripts := map[modules.OutPoint][]byte{
		*outPoint: lockScript[:],
	}

	getPubKeyFn := func(common.Address) ([]byte, error) {
		return pubKeyBytes, nil
	}
	getSignFn := func(addr common.Address, hash []byte) ([]byte, error) {
		s, e := crypto.MyCryptoLib.Sign(privKeyBytes, hash)
		return s, e
	}
	var hashtype uint32
	hashtype = SigHashNone
	_, err := Instance.SignTxAllPaymentInput(tx, hashtype, lockScripts, nil, getPubKeyFn, getSignFn)
	if err != nil {
		t.Logf("Sign error:%s", err)
	}
	unlockScript := tx.TxMessages[0].Payload.(*modules.PaymentPayload).Inputs[0].SignatureScript
	t.Logf("UnlockScript:%x", unlockScript)
	s, _ := Instance.DisasmString(unlockScript)
	t.Logf("UnlockScript string:%s", s)
	//Add any output
	tx.TxMessages[0].Payload.(*modules.PaymentPayload).AddTxOut(modules.NewTxOut(1, lockScript, asset0))
	err = Instance.ScriptValidate(lockScript, nil, tx, 0, 0)
	if err != nil {
		t.Logf("validate error:%s", err)
	}
	// textPayload :=tx.TxMessages[2].Payload.(*modules.DataPayload)
	//textPayload.Text=[]byte("Bad")
	//fmt.Printf("%s", tx.TxMessages[2].Payload.(*modules.DataPayload))

	//err = ScriptValidate(lockScript, nil, tx, 1, 0)
	assert.Nil(t, err, fmt.Sprintf("validate error:%s", err))
}

var (
	prvKey1, _  = crypto.FromWIF("KwN8TdhAMeU8b9UrEYTNTVEvDsy9CSyepwRVNEy2Fc9nbGqDZw4J") //"0454b0699a590b6fc8e66e81db1ca36e99d7c767cdfe44a217b6e105c5db97d5" //P1QJNzZhqGoxNL2igkdthNBQLNWdNGTWzQU
	prvKey1B    = crypto.FromECDSA(prvKey1)
	pubKey1B, _ = hex.DecodeString("02f9286c44fe7ebff9788425d5025ad764cdf5aec3daef862d143c5f09134d75b0")
	address1, _ = common.StringToAddress("P1QJNzZhqGoxNL2igkdthNBQLNWdNGTWzQU")

	prvKey2, _  = crypto.FromWIF("Ky7gQF2rxXLjGSymCtCMa67N2YMt98fRgxyy5WfH92FpbWDxWVRM") //"3892859c02b1be2ce494e61c60181051d79ff21dca22fae1dc349887335b6676" //P1N4nEffoUskPrbnoEqBR69JQDX2vv9vYa8\
	prvKey2B    = crypto.FromECDSA(prvKey2)
	pubKey2B, _ = hex.DecodeString("02a2ba6f2a6e1467334d032ec54ac862c655d7e8bd6bbbce36c771fcdc0ddfb01f")
	address2, _ = common.StringToAddress("P1N4nEffoUskPrbnoEqBR69JQDX2vv9vYa8")

	prvKey3, _  = crypto.FromWIF("KzRHTanikQgR5oqUts69JTrCXRuy9Zod5qXdnAbYwvUnuUDJ3Rro") //"5f7754e5407fc2a81f453645cbd92878a6341d30dbfe2e680fc81628d47e8023" //P1MzuBUT7ubGpkAFqUB6chqTSXmBThQv2HT
	prvKey3B    = crypto.FromECDSA(prvKey3)
	pubKey3B, _ = hex.DecodeString("020945d0c9ed05cf5ca9fe38dde799d7b73f4a5bfb71fc1a3c1dca79e2d86462a7")
	address3, _ = common.StringToAddress("P1MzuBUT7ubGpkAFqUB6chqTSXmBThQv2HT")

	prvKey4, _  = crypto.FromWIF("L3nZf9ds5JG5Sq2WMCxP6QSfHK6WuSpnsU8Qk2ygfGD92h553xhx") //"c3ecda5c797ef8d7ded2d332eb1cb83198ef88ede1bf9de7b60910644b45f83f" //P1MzuBUT7ubGpkAFqUB6chqTSXmBThQv2HT
	prvKey4B    = crypto.FromECDSA(prvKey4)
	pubKey4B, _ = hex.DecodeString("0342ccc3459303c6a24fd3382249af438763c7fab9ca57e919aec658f7d05eab68")
	address4, _ = common.StringToAddress("P1Lcf8CTxgUwmFamn2qM7SrAukNyezakAbK")
)

func build23Address() ([]byte, []byte, string) {

	redeemScript := Instance.GenerateRedeemScript(2, [][]byte{pubKey1B, pubKey2B, pubKey3B})
	lockScript := Instance.GenerateP2SHLockScript(crypto.Hash160(redeemScript))
	addressMulti, _ := Instance.GetAddressFromScript(lockScript)

	return lockScript, redeemScript, addressMulti.Str()
}
func build22Address() ([]byte, []byte, string) {

	redeemScript := Instance.GenerateRedeemScript(2, [][]byte{pubKey1B, pubKey2B})
	lockScript := Instance.GenerateP2SHLockScript(crypto.Hash160(redeemScript))
	addressMulti, _ := Instance.GetAddressFromScript(lockScript)

	return lockScript, redeemScript, addressMulti.Str()
}

//构造一个2/3签名的地址和UTXO，然后用其中的2个私钥对其进行签名
func TestMultiSign1Step(t *testing.T) {
	lockScript, redeemScript, addressMulti := build23Address()
	t.Logf("MultiSign Address:%s\n", addressMulti)
	t.Logf("RedeemScript: %x\n", redeemScript)
	r, _ := Instance.DisasmString(redeemScript)
	t.Logf("RedeemScript: %s\n", r)
	tx := &modules.Transaction{
		TxMessages: make([]*modules.Message, 0),
	}
	asset0 := &modules.Asset{}
	payment := &modules.PaymentPayload{}
	utxoTxId := common.HexToHash("1111870aa8c894376dbd960a22171d0ad7be057a730e14d7103ed4a6dbb34873")
	outPoint := modules.NewOutPoint(utxoTxId, 0, 0)
	txIn := modules.NewTxIn(outPoint, []byte{})
	payment.AddTxIn(txIn)
	p1lockScript := Instance.GenerateP2PKHLockScript(crypto.Hash160(pubKey1B))
	payment.AddTxOut(modules.NewTxOut(1, p1lockScript, asset0))
	tx.TxMessages = append(tx.TxMessages, modules.NewMessage(modules.APP_PAYMENT, payment))
	//scriptCp:=make([]byte,len(lockScript))
	//copy(scriptCp,lockScript)
	//privKeys := map[common.Address]*ecdsa.PrivateKey{
	//	address1: prvKey1,
	//	address2: prvKey2,
	//}
	getPubKeyFn := func(addr common.Address) ([]byte, error) {
		if addr == address1 {
			return crypto.CompressPubkey(&prvKey1.PublicKey), nil
		}
		if addr == address2 {
			return crypto.CompressPubkey(&prvKey2.PublicKey), nil
		}
		return nil, nil
	}
	getSignFn := func(addr common.Address, msg []byte) ([]byte, error) {

		if addr == address1 {
			return crypto.MyCryptoLib.Sign(prvKey1B, msg)
		}
		if addr == address2 {
			return crypto.MyCryptoLib.Sign(prvKey2B, msg)
		}
		return nil, nil
	}
	sign12, err := Instance.MultiSignOnePaymentInput(tx, SigHashAll, 0, 0, lockScript, redeemScript, getPubKeyFn, getSignFn, nil)
	if err != nil {
		t.Logf("Sign error:%s", err)
	}
	t.Logf("PrvKey1&2 sign result:%x\n", sign12)
	pay1 := tx.TxMessages[0].Payload.(*modules.PaymentPayload)
	pay1.Inputs[0].SignatureScript = sign12
	str, _ := txscript.DisasmString(sign12)
	t.Logf("Signed script:{%s}", str)

	err = Instance.ScriptValidate(lockScript, nil, tx, 0, 0)
	assert.Nil(t, err, fmt.Sprintf("validate error:%s", err))
}

//构造一个2/3签名的地址和UTXO，然后用其中的2个私钥分两步对其进行签名
func TestMultiSign2Step(t *testing.T) {
	lockScript, redeemScript, addressMulti := build23Address()
	t.Logf("MultiSign Address:%s\n", addressMulti)
	t.Logf("RedeemScript: %x\n", redeemScript)
	t.Logf("RedeemScript: %d\n", redeemScript)
	tx := &modules.Transaction{
		TxMessages: make([]*modules.Message, 0),
	}
	payment := &modules.PaymentPayload{}
	utxoTxId := common.HexToHash("1111870aa8c894376dbd960a22171d0ad7be057a730e14d7103ed4a6dbb34873")
	outPoint := modules.NewOutPoint(utxoTxId, 0, 0)
	txIn := modules.NewTxIn(outPoint, []byte{})
	payment.AddTxIn(txIn)
	p1lockScript := Instance.GenerateP2PKHLockScript(crypto.Hash160(pubKey1B))
	asset0 := &modules.Asset{}
	payment.AddTxOut(modules.NewTxOut(1, p1lockScript, asset0))
	tx.TxMessages = append(tx.TxMessages, modules.NewMessage(modules.APP_PAYMENT, payment))
	//scriptCp:=make([]byte,len(lockScript))
	//copy(scriptCp,lockScript)
	//privKeys := map[common.Address]*ecdsa.PrivateKey{
	//	address1: prvKey1,
	//}
	getPubKeyFn := func(addr common.Address) ([]byte, error) {
		if addr == address1 {
			return crypto.CompressPubkey(&prvKey1.PublicKey), nil
		}
		if addr == address2 {
			return crypto.CompressPubkey(&prvKey2.PublicKey), nil
		}
		return nil, nil
	}
	getSignFn := func(addr common.Address, msg []byte) ([]byte, error) {

		if addr == address1 {
			return crypto.MyCryptoLib.Sign(prvKey1B, msg)
		}
		if addr == address2 {
			return crypto.MyCryptoLib.Sign(prvKey2B, msg)
		}
		return nil, nil
	}
	sign1, err := Instance.MultiSignOnePaymentInput(tx, SigHashAll, 0, 0, lockScript, redeemScript, getPubKeyFn, getSignFn, nil)
	if err != nil {
		t.Logf("Sign error:%s", err)
	}
	t.Logf("PrvKey1 sign result:%x\n", sign1)
	pay1 := tx.TxMessages[0].Payload.(*modules.PaymentPayload)
	pay1.Inputs[0].SignatureScript = sign1

	//privKeys2 := map[common.Address]*ecdsa.PrivateKey{
	//	address2: prvKey2,
	//}
	//scriptCp2:=make([]byte,len(lockScript))
	//copy(scriptCp2,lockScript)
	sign2, err := Instance.MultiSignOnePaymentInput(tx, SigHashAll, 0, 0, lockScript, redeemScript, getPubKeyFn, getSignFn, sign1)
	if err != nil {
		t.Logf("Sign error:%s", err)
	}
	t.Logf("PrvKey2 sign result:%x\n", sign2)

	pay1 = tx.TxMessages[0].Payload.(*modules.PaymentPayload)
	pay1.Inputs[0].SignatureScript = sign2
	str, _ := txscript.DisasmString(sign2)
	t.Logf("Signed script:{%s}", str)

	err = Instance.ScriptValidate(lockScript, nil, tx, 0, 0)
	assert.Nil(t, err, fmt.Sprintf("validate error:%s", err))
}
func mockPickupJuryRedeemScript(addr common.Address) ([]byte, error) {
	return Instance.GenerateRedeemScript(2, [][]byte{pubKey1B, pubKey2B, pubKey3B, pubKey4B}), nil
}
func TestContractPayout(t *testing.T) {
	tx := &modules.Transaction{
		TxMessages: make([]*modules.Message, 0),
	}
	asset0 := &modules.Asset{}
	payment := &modules.PaymentPayload{}
	utxoTxId := common.HexToHash("1111870aa8c894376dbd960a22171d0ad7be057a730e14d7103ed4a6dbb34873")
	outPoint := modules.NewOutPoint(utxoTxId, 0, 0)
	txIn := modules.NewTxIn(outPoint, []byte{})
	payment.AddTxIn(txIn)
	outPoint1 := modules.NewOutPoint(utxoTxId, 0, 1)
	txIn1 := modules.NewTxIn(outPoint1, []byte{})
	payment.AddTxIn(txIn1)
	p1lockScript := Instance.GenerateP2PKHLockScript(crypto.Hash160(pubKey1B))
	payment.AddTxOut(modules.NewTxOut(1, p1lockScript, asset0))
	tx.TxMessages = append(tx.TxMessages, modules.NewMessage(modules.APP_PAYMENT, payment))
	//scriptCp:=make([]byte,len(lockScript))
	//copy(scriptCp,lockScript)
	contractAddr, _ := common.StringToAddress("PCGTta3M4t3yXu8uRgkKvaWd2d8DR32W9vM")
	lockScript := Instance.GenerateP2CHLockScript(contractAddr) //Token 锁定到保证金合约中
	l, _ := txscript.DisasmString(lockScript)
	t.Logf("Lock Script:%s", l)
	//privKeys := map[common.Address]*ecdsa.PrivateKey{
	//	address1: prvKey1,
	//	address2: prvKey2,
	//}
	getPubKeyFn := func(addr common.Address) ([]byte, error) {
		if addr == address1 {
			return crypto.CompressPubkey(&prvKey1.PublicKey), nil
		}
		if addr == address2 {
			return crypto.CompressPubkey(&prvKey2.PublicKey), nil
		}
		if addr == address3 {
			return crypto.CompressPubkey(&prvKey3.PublicKey), nil
		}
		if addr == address4 {
			return crypto.CompressPubkey(&prvKey4.PublicKey), nil
		}
		return nil, nil
	}
	getSignFn := func(addr common.Address, msg []byte) ([]byte, error) {

		if addr == address1 {
			return crypto.MyCryptoLib.Sign(prvKey1B, msg)
		}
		if addr == address2 {
			return crypto.MyCryptoLib.Sign(prvKey2B, msg)
		}
		if addr == address3 {
			return crypto.MyCryptoLib.Sign(prvKey3B, msg)
		}
		if addr == address4 {
			return crypto.MyCryptoLib.Sign(prvKey4B, msg)
		}
		return nil, nil
	}
	redeemScript, _ := mockPickupJuryRedeemScript(contractAddr)
	r, _ := txscript.DisasmString(redeemScript)
	t.Logf("RedeemScript:%s", r)
	//Sign input0
	sign0, err := Instance.MultiSignOnePaymentInput(tx, SigHashRaw, 0, 0, lockScript, redeemScript, getPubKeyFn, getSignFn, nil)
	if err != nil {
		t.Logf("Sign error:%s", err)
	}
	t.Logf("PrvKey1&2 sign result:%x\n", sign0)
	pay1 := tx.TxMessages[0].Payload.(*modules.PaymentPayload)
	pay1.Inputs[0].SignatureScript = sign0
	str, _ := txscript.DisasmString(sign0)
	t.Logf("Signed script:{%s}", str)
	//Sign input1
	sign1, err := Instance.MultiSignOnePaymentInput(tx, SigHashRaw, 0, 1, lockScript, redeemScript, getPubKeyFn, getSignFn, nil)
	if err != nil {
		t.Logf("Sign error:%s", err)
	}
	t.Logf("PrvKey1&2 sign result:%x\n", sign1)
	pay1.Inputs[1].SignatureScript = sign1
	str1, _ := txscript.DisasmString(sign1)
	t.Logf("Signed script:{%s}", str1)

	err = Instance.ScriptValidate(lockScript, mockPickupJuryRedeemScript, tx, 0, 0)
	assert.Nil(t, err, fmt.Sprintf("validate error:%s", err))

	err = Instance.ScriptValidate(lockScript, mockPickupJuryRedeemScript, tx, 0, 1)
	assert.Nil(t, err, fmt.Sprintf("validate error:%s", err))

}

func TestGenerateRedeemScript(t *testing.T) {
	jury := []string{"03f28978eeb02f2c97db338f76b093ada5a36022964d85ff3ad5c04b800602b071", "0312857d1e1c5c151f37cede37d694f6b8661f37690f5b3b87946489c479e6dc80"}

	pubKeys := [][]byte{}
	for _, jurior := range jury {
		pubKey1, _ := hex.DecodeString(jurior)
		pubKeys = append(pubKeys, pubKey1)
	}
	redeem := Instance.GenerateRedeemScript(2, pubKeys)

	fmt.Printf("%x", redeem)
}

func TestUserContractPayout(t *testing.T) {
	tx := &modules.Transaction{
		TxMessages: make([]*modules.Message, 0),
	}
	asset0 := modules.NewPTNAsset()
	payment := &modules.PaymentPayload{}
	utxoTxId := common.HexToHash("1111870aa8c894376dbd960a22171d0ad7be057a730e14d7103ed4a6dbb34873")
	outPoint := modules.NewOutPoint(utxoTxId, 0, 0)
	txIn := modules.NewTxIn(outPoint, []byte{})
	payment.AddTxIn(txIn)
	outPoint1 := modules.NewOutPoint(utxoTxId, 0, 1)
	txIn1 := modules.NewTxIn(outPoint1, []byte{})
	payment.AddTxIn(txIn1)
	p1lockScript := Instance.GenerateP2PKHLockScript(crypto.Hash160(pubKey1B))
	payment.AddTxOut(modules.NewTxOut(1, p1lockScript, asset0))
	tx.TxMessages = append(tx.TxMessages, modules.NewMessage(modules.APP_PAYMENT, payment))
	//scriptCp:=make([]byte,len(lockScript))
	//copy(scriptCp,lockScript)
	contractAddr, _ := common.StringToAddress("PCGTta3M4t3yXu8uRgkKvaWd2d8DR32W9vM")
	lockScript := Instance.GenerateP2CHLockScript(contractAddr) //Token 锁定到合约中
	l, _ := txscript.DisasmString(lockScript)
	t.Logf("Lock Script:%s", l)
	getPubKeyFn := func(addr common.Address) ([]byte, error) {
		if addr == address1 {
			return crypto.CompressPubkey(&prvKey1.PublicKey), nil
		}
		if addr == address2 {
			return crypto.CompressPubkey(&prvKey2.PublicKey), nil
		}
		if addr == address3 {
			return crypto.CompressPubkey(&prvKey3.PublicKey), nil
		}
		if addr == address4 {
			return crypto.CompressPubkey(&prvKey4.PublicKey), nil
		}
		return nil, nil
	}
	getSignFn := func(addr common.Address, msg []byte) ([]byte, error) {

		if addr == address1 {
			return crypto.MyCryptoLib.Sign(prvKey1B, msg)
		}
		if addr == address2 {
			return crypto.MyCryptoLib.Sign(prvKey2B, msg)
		}
		if addr == address3 {
			return crypto.MyCryptoLib.Sign(prvKey3B, msg)
		}
		if addr == address4 {
			return crypto.MyCryptoLib.Sign(prvKey4B, msg)
		}
		return nil, nil
	}
	//构造用户合约的赎回脚本
	redeemScript, _ := mockPickupJuryRedeemScript(contractAddr)
	r, _ := txscript.DisasmString(redeemScript)
	t.Logf("RedeemScript:%s", r)
	lockScripts := make(map[modules.OutPoint][]byte)
	lockScripts[*outPoint] = lockScript
	lockScripts[*outPoint1] = lockScript
	//签名所有input中锁定脚本为空的情况。
	errMsg, err := Instance.SignTxAllPaymentInput(tx, SigHashRaw, lockScripts, redeemScript, getPubKeyFn, getSignFn)
	if err != nil {
		t.Logf("Sign error:%s", err)
	}
	if len(errMsg) > 0 {
		t.Logf("err:%s", errMsg[0].Error.Error())
	}
	err = Instance.ScriptValidate(lockScript, mockPickupJuryRedeemScript, tx, 0, 0)
	assert.Nil(t, err, fmt.Sprintf("validate error:%s", err))

	err = Instance.ScriptValidate(lockScript, mockPickupJuryRedeemScript, tx, 0, 1)
	assert.Nil(t, err, fmt.Sprintf("validate error:%s", err))

	addrs, err := Instance.GetScriptSigners(tx, 0, 0)
	assert.Nil(t, err)
	for _, addr := range addrs {
		t.Logf("Signed address:%s", addr.String())
	}
}
func TestMergeContractUnlockScript(t *testing.T) {
	sign1 := []byte("111111111111")
	sign2 := []byte("222222222222")
	sign3 := []byte("333333333333")
	redeemScript, _ := mockPickupJuryRedeemScript(common.Address{})
	result := Instance.MergeContractUnlockScript([][]byte{sign1, sign2, sign3}, redeemScript)
	t.Logf("%x", result)
	txt, _ := Instance.DisasmString(result)
	t.Log(txt)
}
func Test22MutiSign(t *testing.T) {
	lockScript, redeemScript, addressMulti := build22Address()
	t.Logf("MultiSign Address:%s\n", addressMulti)
	t.Logf("RedeemScript: %x\n", redeemScript)
	r, _ := Instance.DisasmString(redeemScript)
	t.Logf("RedeemScript: %s\n", r)
	tx := &modules.Transaction{
		TxMessages: make([]*modules.Message, 0),
	}
	asset0 := &modules.Asset{}
	payment := &modules.PaymentPayload{}
	utxoTxId := common.HexToHash("1111870aa8c894376dbd960a22171d0ad7be057a730e14d7103ed4a6dbb34873")
	outPoint := modules.NewOutPoint(utxoTxId, 0, 0)
	txIn := modules.NewTxIn(outPoint, []byte{})
	payment.AddTxIn(txIn)
	p1lockScript := Instance.GenerateP2PKHLockScript(crypto.Hash160(pubKey1B))
	payment.AddTxOut(modules.NewTxOut(1, p1lockScript, asset0))
	tx.TxMessages = append(tx.TxMessages, modules.NewMessage(modules.APP_PAYMENT, payment))
	//scriptCp:=make([]byte,len(lockScript))
	//copy(scriptCp,lockScript)
	//privKeys := map[common.Address]*ecdsa.PrivateKey{
	//	address1: prvKey1,
	//	address2: prvKey2,
	//}
	getPubKeyFn := func(addr common.Address) ([]byte, error) {
		if addr == address1 {
			return crypto.CompressPubkey(&prvKey1.PublicKey), nil
		}
		if addr == address2 {
			return crypto.CompressPubkey(&prvKey2.PublicKey), nil
		}
		return nil, nil
	}
	getSignFn := func(addr common.Address, msg []byte) ([]byte, error) {

		if addr == address1 {
			return crypto.MyCryptoLib.Sign(prvKey1B, msg)
		}
		if addr == address2 {
			return crypto.MyCryptoLib.Sign(prvKey2B, msg)
		}
		if addr == address3 {
			return crypto.MyCryptoLib.Sign(prvKey3B, msg)
		}
		if addr == address4 {
			return crypto.MyCryptoLib.Sign(prvKey4B, msg)
		}
		return nil, nil
	}
	sign12, err := Instance.MultiSignOnePaymentInput(tx, SigHashAll, 0, 0, lockScript, redeemScript, getPubKeyFn, getSignFn, nil)
	if err != nil {
		t.Logf("Sign error:%s", err)
	}
	t.Logf("PrvKey1&2 sign result:%x\n", sign12)
	pay1 := tx.TxMessages[0].Payload.(*modules.PaymentPayload)
	pay1.Inputs[0].SignatureScript = sign12
	str, _ := txscript.DisasmString(sign12)
	t.Logf("Unlock script:{%s}", str)

	err = Instance.ScriptValidate(lockScript, nil, tx, 0, 0)
	assert.Nil(t, err, fmt.Sprintf("validate error:%s", err))
}
func TestSampleTx(t *testing.T) {
	privKeyBytes, _ := hex.DecodeString("2BE3B4B671FF5B8009E6876CCCC8808676C1C279EE824D0AB530294838DC1644")
	pubKeyBytes, _ := crypto.MyCryptoLib.PrivateKeyToPubKey(privKeyBytes)
	pubKeyHash := crypto.Hash160(pubKeyBytes)
	t.Logf("Public Key:%x", pubKeyBytes)
	addr := crypto.PubkeyBytesToAddress(pubKeyBytes)
	t.Logf("Addr:%s", addr.String())
	lockScript := Instance.GenerateP2PKHLockScript(pubKeyHash)
	t.Logf("UTXO lock script:%x", lockScript)

	tx := &modules.Transaction{
		TxMessages: make([]*modules.Message, 0),
	}
	pay1 := &modules.PaymentPayload{}
	utxoTxId := common.HexToHash("5651870aa8c894376dbd960a22171d0ad7be057a730e14d7103ed4a6dbb34873")
	outPoint := modules.NewOutPoint(utxoTxId, 0, 0)
	txIn := modules.NewTxIn(outPoint, []byte{})
	pay1.AddTxIn(txIn)
	asset0 := &modules.Asset{}
	pay1.AddTxOut(modules.NewTxOut(1, lockScript, asset0))

	tx.TxMessages = append(tx.TxMessages, modules.NewMessage(modules.APP_PAYMENT, pay1))

	lockScripts := map[modules.OutPoint][]byte{
		*outPoint: lockScript[:],
	}
	getPubKeyFn := func(common.Address) ([]byte, error) {
		return pubKeyBytes, nil
	}
	getSignFn := func(addr common.Address, hash []byte) ([]byte, error) {
		s, e := crypto.MyCryptoLib.Sign(privKeyBytes, hash)
		return s, e
	}
	var hashtype uint32
	hashtype = 1
	_, err := Instance.SignTxAllPaymentInput(tx, hashtype, lockScripts, nil, getPubKeyFn, getSignFn)
	if err != nil {
		t.Logf("Sign error:%s", err)
	}
	txjson, _ := json.Marshal(tx)
	txrlp, _ := rlp.EncodeToBytes(tx)
	t.Logf("Tx len:%d, txjson:%s", len(txrlp), string(txjson))
	unlockScript := tx.TxMessages[0].Payload.(*modules.PaymentPayload).Inputs[0].SignatureScript
	t.Logf("UnlockScript:%x", unlockScript)
	s, _ := Instance.DisasmString(unlockScript)
	t.Logf("UnlockScript string:%s", s)
	err = Instance.ScriptValidate(lockScript, nil, tx, 0, 0)
	if err != nil {
		t.Logf("validate error:%s", err)
	}

}
func generateSignedTx() (*modules.Transaction, []byte) {
	privKeyBytes, _ := hex.DecodeString("2BE3B4B671FF5B8009E6876CCCC8808676C1C279EE824D0AB530294838DC1644")
	privKey, _ := crypto.ToECDSA(privKeyBytes)
	pubKey := privKey.PublicKey
	pubKeyBytes := crypto.CompressPubkey(&pubKey)
	pubKeyHash := crypto.Hash160(pubKeyBytes)

	lockScript := Instance.GenerateP2PKHLockScript(pubKeyHash)

	tx := &modules.Transaction{
		TxMessages: make([]*modules.Message, 0),
	}
	pay1 := &modules.PaymentPayload{}
	utxoTxId := common.HexToHash("5651870aa8c894376dbd960a22171d0ad7be057a730e14d7103ed4a6dbb34873")
	outPoint := modules.NewOutPoint(utxoTxId, 0, 0)
	txIn := modules.NewTxIn(outPoint, []byte{})
	pay1.AddTxIn(txIn)
	asset0 := &modules.Asset{}
	pay1.AddTxOut(modules.NewTxOut(1, lockScript, asset0))

	tx.TxMessages = append(tx.TxMessages, modules.NewMessage(modules.APP_PAYMENT, pay1))

	lockScripts := map[modules.OutPoint][]byte{
		*outPoint: lockScript[:],
	}
	getPubKeyFn := func(common.Address) ([]byte, error) {
		return crypto.CompressPubkey(&privKey.PublicKey), nil
	}
	getSignFn := func(addr common.Address, hash []byte) ([]byte, error) {
		s, e := crypto.Sign(hash, privKey)
		return s, e
	}
	var hashtype uint32
	hashtype = 1
	Instance.SignTxAllPaymentInput(tx, hashtype, lockScripts, nil, getPubKeyFn, getSignFn)
	return tx, lockScript
}

func BenchmarkScriptValidate(b *testing.B) {
	tx, lockScript := generateSignedTx()
	for i := 0; i < b.N; i++ {

		err := Instance.ScriptValidate(lockScript, nil, tx, 0, 0)
		if err != nil {
			b.Logf("validate error:%s", err)
		}
	}
}

//func Test2WScriptValidate(t *testing.T) {
//	tx, lockScript := generateSignedTx()
//	start := time.Now()
//	for i := 0; i < 20000; i++ {
//		err := ScriptValidate(lockScript, nil, tx, 0, 0)
//		if err != nil {
//			t.Logf("validate error:%s", err)
//		}
//	}
//	t.Logf("Cost time:%v", time.Since(start))
//}
