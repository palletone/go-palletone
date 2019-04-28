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
	"testing"

	//"github.com/palletone/go-palletone/tokenengine/btcd/chaincfg/chainhash"
	"encoding/hex"
	"fmt"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/crypto"

	"github.com/ethereum/go-ethereum/rlp"
	"github.com/palletone/go-palletone/contracts/syscontract"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/tokenengine/internal/txscript"
	"github.com/stretchr/testify/assert"
)

func TestGetAddressFromScript(t *testing.T) {
	addrStr := "P1JEStL6tb7TB8e6ZJSpJhQoqin2A6pabdA"
	addr, _ := common.StringToAddress(addrStr)
	p2pkhLock := GenerateP2PKHLockScript(addr.Bytes())
	t.Logf("P2PKH script:%x", p2pkhLock)
	getAddr, _ := GetAddressFromScript(p2pkhLock)
	t.Logf("Get Address:%s", getAddr.Str())
	assert.True(t, getAddr == addr, "Address parse error")

	addr2, _ := common.StringToAddress("P35SbSqXuXcHrtZuJKzbStpcqzwCg88jXfn")
	p2shLock := GenerateP2SHLockScript(addr2.Bytes())
	getAddr2, _ := GetAddressFromScript(p2shLock)
	t.Logf("Get Script Address:%s", getAddr2.Str())
	assert.True(t, getAddr2 == addr2, "Address parse error")

}

func TestGenerateP2CHLockScript(t *testing.T) {
	addrStr := "PCGTta3M4t3yXu8uRgkKvaWd2d8DR32W9vM"
	addr, err := common.StringToAddress(addrStr)
	assert.Nil(t, err)
	p2ch1 := GenerateLockScript(addr)
	p2ch1Str, _ := DisasmString(p2ch1)
	t.Logf("Pay to contract hash lock script:%x, string:%s", p2ch1, p2ch1Str)
	p2ch2 := GenerateP2CHLockScript(addr)
	assert.Equal(t, p2ch1, p2ch2)
	addr2, err := GetAddressFromScript(p2ch1)
	assert.Nil(t, err, "Err must nil")
	assert.Equal(t, addr2.String(), addrStr)
	t.Logf("get address:%s", addr2.String())

	deposit := syscontract.DepositContractAddress
	script := GenerateP2CHLockScript(deposit)
	t.Logf("%x", script)
}
func TestDecodeScriptBytes(t *testing.T) {
	data, _ := hex.DecodeString("21021b11a1d070a173edc33a25ad8382602b2ad9b58670f031059b8866fc8e1b16aeac")
	str, _ := txscript.DisasmString(data)
	t.Log(str)
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
	utxoTxId := common.HexToHash("5651870aa8c894376dbd960a22171d0ad7be057a730e14d7103ed4a6dbb34873")
	outPoint := modules.NewOutPoint(utxoTxId, 0, 0)
	txIn := modules.NewTxIn(outPoint, []byte{})
	payment.AddTxIn(txIn)
	asset0 := &modules.Asset{}
	payment.AddTxOut(modules.NewTxOut(1, lockScript, asset0))
	payment2 := &modules.PaymentPayload{}
	utxoTxId2 := common.HexToHash("1651870aa8c894376dbd960a22171d0ad7be057a730e14d7103ed4a6dbb34873")
	outPoint2 := modules.NewOutPoint(utxoTxId2, 1, 1)
	txIn2 := modules.NewTxIn(outPoint2, []byte{})
	payment2.AddTxIn(txIn2)
	asset1 := &modules.Asset{AssetId: modules.PTNCOIN}
	payment2.AddTxOut(modules.NewTxOut(1, lockScript, asset1))
	tx.TxMessages = append(tx.TxMessages, modules.NewMessage(modules.APP_PAYMENT, payment))
	tx.TxMessages = append(tx.TxMessages, modules.NewMessage(modules.APP_PAYMENT, payment2))

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
		*outPoint2: GenerateP2PKHLockScript(pubKeyHash),
	}
	//privKeys := map[common.Address]*ecdsa.PrivateKey{
	//	addr: privKey,
	//}
	getPubKeyFn := func(common.Address) ([]byte, error) {
		return crypto.CompressPubkey(&privKey.PublicKey), nil
	}
	getSignFn := func(addr common.Address, hash []byte) ([]byte, error) {
		s, e := crypto.Sign(hash, privKey)
		return s[0:64], e
	}
	var hashtype uint32
	hashtype = 1
	_, err := SignTxAllPaymentInput(tx, hashtype, lockScripts, nil, getPubKeyFn, getSignFn)
	if err != nil {
		t.Logf("Sign error:%s", err)
	}
	unlockScript := tx.TxMessages[0].Payload.(*modules.PaymentPayload).Inputs[0].SignatureScript
	t.Logf("UnlockScript:%x", unlockScript)
	s, _ := DisasmString(unlockScript)
	t.Logf("UnlockScript string:%s", s)
	err = ScriptValidate(lockScript, nil, tx, 0, 0)
	if err != nil {
		t.Logf("validate error:%s", err)
	}
	// textPayload :=tx.TxMessages[2].Payload.(*modules.DataPayload)
	//textPayload.Text=[]byte("Bad")
	//fmt.Printf("%s", tx.TxMessages[2].Payload.(*modules.DataPayload))

	err = ScriptValidate(lockScript, nil, tx, 1, 0)
	assert.Nil(t, err, fmt.Sprintf("validate error:%s", err))

}
func TestHashNone1Payment(t *testing.T) {
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
		return crypto.CompressPubkey(&privKey.PublicKey), nil
	}
	getSignFn := func(addr common.Address, hash []byte) ([]byte, error) {
		s, e := crypto.Sign(hash, privKey)
		return s[0:64], e
	}
	var hashtype uint32
	hashtype = SigHashNone
	_, err := SignTxAllPaymentInput(tx, hashtype, lockScripts, nil, getPubKeyFn, getSignFn)
	if err != nil {
		t.Logf("Sign error:%s", err)
	}
	unlockScript := tx.TxMessages[0].Payload.(*modules.PaymentPayload).Inputs[0].SignatureScript
	t.Logf("UnlockScript:%x", unlockScript)
	s, _ := DisasmString(unlockScript)
	t.Logf("UnlockScript string:%s", s)
	//Add any output
	tx.TxMessages[0].Payload.(*modules.PaymentPayload).AddTxOut(modules.NewTxOut(1, lockScript, asset0))
	err = ScriptValidate(lockScript, nil, tx, 0, 0)
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
	pubKey1B, _ = hex.DecodeString("02f9286c44fe7ebff9788425d5025ad764cdf5aec3daef862d143c5f09134d75b0")
	address1, _ = common.StringToAddress("P1QJNzZhqGoxNL2igkdthNBQLNWdNGTWzQU")

	prvKey2, _  = crypto.FromWIF("Ky7gQF2rxXLjGSymCtCMa67N2YMt98fRgxyy5WfH92FpbWDxWVRM") //"3892859c02b1be2ce494e61c60181051d79ff21dca22fae1dc349887335b6676" //P1N4nEffoUskPrbnoEqBR69JQDX2vv9vYa8\
	pubKey2B, _ = hex.DecodeString("02a2ba6f2a6e1467334d032ec54ac862c655d7e8bd6bbbce36c771fcdc0ddfb01f")
	address2, _ = common.StringToAddress("P1N4nEffoUskPrbnoEqBR69JQDX2vv9vYa8")

	prvKey3, _  = crypto.FromWIF("KzRHTanikQgR5oqUts69JTrCXRuy9Zod5qXdnAbYwvUnuUDJ3Rro") //"5f7754e5407fc2a81f453645cbd92878a6341d30dbfe2e680fc81628d47e8023" //P1MzuBUT7ubGpkAFqUB6chqTSXmBThQv2HT
	pubKey3B, _ = hex.DecodeString("020945d0c9ed05cf5ca9fe38dde799d7b73f4a5bfb71fc1a3c1dca79e2d86462a7")
	address3, _ = common.StringToAddress("P1MzuBUT7ubGpkAFqUB6chqTSXmBThQv2HT")

	prvKey4, _  = crypto.FromWIF("L3nZf9ds5JG5Sq2WMCxP6QSfHK6WuSpnsU8Qk2ygfGD92h553xhx") //"c3ecda5c797ef8d7ded2d332eb1cb83198ef88ede1bf9de7b60910644b45f83f" //P1MzuBUT7ubGpkAFqUB6chqTSXmBThQv2HT
	pubKey4B, _ = hex.DecodeString("0342ccc3459303c6a24fd3382249af438763c7fab9ca57e919aec658f7d05eab68")
	address4, _ = common.StringToAddress("P1Lcf8CTxgUwmFamn2qM7SrAukNyezakAbK")
)

func build23Address() ([]byte, []byte, string) {

	redeemScript := GenerateRedeemScript(2, [][]byte{pubKey1B, pubKey2B, pubKey3B})
	lockScript := GenerateP2SHLockScript(crypto.Hash160(redeemScript))
	addressMulti, _ := GetAddressFromScript(lockScript)

	return lockScript, redeemScript, addressMulti.Str()
}

//构造一个2/3签名的地址和UTXO，然后用其中的2个私钥对其进行签名
func TestMultiSign1Step(t *testing.T) {
	lockScript, redeemScript, addressMulti := build23Address()
	t.Logf("MultiSign Address:%s\n", addressMulti)
	t.Logf("RedeemScript: %x\n", redeemScript)
	r, _ := DisasmString(redeemScript)
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
	p1lockScript := GenerateP2PKHLockScript(crypto.Hash160(pubKey1B))
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
	getSignFn := func(addr common.Address, hash []byte) ([]byte, error) {

		if addr == address1 {
			return crypto.Sign(hash, prvKey1)
		}
		if addr == address2 {
			return crypto.Sign(hash, prvKey2)
		}
		return nil, nil
	}
	sign12, err := MultiSignOnePaymentInput(tx, SigHashAll, 0, 0, lockScript, redeemScript, getPubKeyFn, getSignFn, nil)
	if err != nil {
		t.Logf("Sign error:%s", err)
	}
	t.Logf("PrvKey1&2 sign result:%x\n", sign12)
	pay1 := tx.TxMessages[0].Payload.(*modules.PaymentPayload)
	pay1.Inputs[0].SignatureScript = sign12
	str, _ := txscript.DisasmString(sign12)
	t.Logf("Signed script:{%s}", str)

	err = ScriptValidate(lockScript, nil, tx, 0, 0)
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
	p1lockScript := GenerateP2PKHLockScript(crypto.Hash160(pubKey1B))
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
	getSignFn := func(addr common.Address, hash []byte) ([]byte, error) {

		if addr == address1 {
			return crypto.Sign(hash, prvKey1)
		}
		if addr == address2 {
			return crypto.Sign(hash, prvKey2)
		}
		return nil, nil
	}
	sign1, err := MultiSignOnePaymentInput(tx, SigHashAll, 0, 0, lockScript, redeemScript, getPubKeyFn, getSignFn, nil)
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
	sign2, err := MultiSignOnePaymentInput(tx, SigHashAll, 0, 0, lockScript, redeemScript, getPubKeyFn, getSignFn, sign1)
	if err != nil {
		t.Logf("Sign error:%s", err)
	}
	t.Logf("PrvKey2 sign result:%x\n", sign2)

	pay1 = tx.TxMessages[0].Payload.(*modules.PaymentPayload)
	pay1.Inputs[0].SignatureScript = sign2
	str, _ := txscript.DisasmString(sign2)
	t.Logf("Signed script:{%s}", str)

	err = ScriptValidate(lockScript, nil, tx, 0, 0)
	assert.Nil(t, err, fmt.Sprintf("validate error:%s", err))
}
func mockPickupJuryRedeemScript(addr common.Address) ([]byte, error) {
	return GenerateRedeemScript(2, [][]byte{pubKey1B, pubKey2B, pubKey3B, pubKey4B}), nil
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
	p1lockScript := GenerateP2PKHLockScript(crypto.Hash160(pubKey1B))
	payment.AddTxOut(modules.NewTxOut(1, p1lockScript, asset0))
	tx.TxMessages = append(tx.TxMessages, modules.NewMessage(modules.APP_PAYMENT, payment))
	//scriptCp:=make([]byte,len(lockScript))
	//copy(scriptCp,lockScript)
	contractAddr, _ := common.StringToAddress("PCGTta3M4t3yXu8uRgkKvaWd2d8DR32W9vM")
	lockScript := GenerateP2CHLockScript(contractAddr) //Token 锁定到保证金合约中
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
	getSignFn := func(addr common.Address, hash []byte) ([]byte, error) {

		if addr == address1 {
			return crypto.Sign(hash, prvKey1)
		}
		if addr == address2 {
			return crypto.Sign(hash, prvKey2)
		}
		if addr == address3 {
			return crypto.Sign(hash, prvKey3)
		}
		if addr == address4 {
			return crypto.Sign(hash, prvKey4)
		}
		return nil, nil
	}
	redeemScript, _ := mockPickupJuryRedeemScript(contractAddr)
	r, _ := txscript.DisasmString(redeemScript)
	t.Logf("RedeemScript:%s", r)
	sign12, err := MultiSignOnePaymentInput(tx, SigHashAll, 0, 0, lockScript, redeemScript, getPubKeyFn, getSignFn, nil)
	if err != nil {
		t.Logf("Sign error:%s", err)
	}
	t.Logf("PrvKey1&2 sign result:%x\n", sign12)
	pay1 := tx.TxMessages[0].Payload.(*modules.PaymentPayload)
	pay1.Inputs[0].SignatureScript = sign12
	str, _ := txscript.DisasmString(sign12)
	t.Logf("Signed script:{%s}", str)

	err = ScriptValidate(lockScript, mockPickupJuryRedeemScript, tx, 0, 0)
	assert.Nil(t, err, fmt.Sprintf("validate error:%s", err))
}
func Test1(t *testing.T) {
	rlpData, _ := hex.DecodeString("f921bff921bcf8de80b8dbf8d9f88cf88ab86441f2e765b573b0bd83ff7f601822cd61398f5c67641cc73dfeb5217ec2a8d6ab8948db63834b05f70696a4bd0419cb6da90bb6259aab33c125a91562c600d5f7c20121029a128b478a874a272309f93e8d2ba1ae8f8235d3f7c1d10132f74f3fe8a56eee80a0a4c10e9527ef8c4cdbf6a9b231f4459d60ba53d3a7a4d23de662f8d2e784f57e8080f848f84688016345785d89ff9a9976a914381db54b7e5d5d4f4fb1cd63cb5a94d5e839810488ace290400082bb080000000000000000000000900000000000000000000000000000000080dc649ad9866a75727930368a627463706f7274322d338570746e3132c0f9204c01b92048f92045a099faf9f35c99a3e5e17fcc97302f959dd137f1277bc5109bb673ad04748f1561866a75727930368a627463706f7274322d338570746e313280b920041f8b08000000000000ffec7d6b6f1b4992607f65fd8abc02c64d5a6cbef434db3446925f6adbb24692ddd3230b9e242b8b2cbb58c5aecc12a91604dc00f7c2018399c3dd1ceec3e1eec32eb0c002bb9805160bcc00bbbf663cd3f32f16f9ceaaca2229c976cf00e60789accc8c8c888c8c888c8cccc2c9a0d92783499c90ce17abf26b63187ff6fe3ead76abb5b1b6f659abd56a6d6e64ffd3928df6baf8ce3e9fb5daeb9df5ce67a0f51e7128fda498c0e4b356ebbb5938b7dea27285bcf8ff57f269de762ac7a300033f0811083098c08480d807c3f88b090c4344e208359c8af99356f31384008e7d328509ea82f338050318810479012649d04f0902010130f29a7102c6b117f8e74e2520208d3c9400324280a0648c6957f4c7a3fd17e0118a5002437090f6c360009e0603146104200613fa048f9007fae74e85d67f48fb3f12fd8387711a79900471540728202394803394e0208ec0aaec4180ab8338712a554828ce098827b4550dc0e81c8490e886169a35691e08220674144f1020234828b1d3200c411f8114233f0deb4ea59f12f0f5def1e3e72f8ec1f6fe37e0ebedc3c3edfde36fbe04d3808ce29400748638a0603c0903e481294c1218917310fb4ee5d983c3ddc7dbfbc7db3b7b4ff78ebf0171021eee1def3f383a020f9f1f826d70b07d78bcb7fbe2e9f62138787178f0fce8410380234451424e650e4f7d362809021e22300871c3a97c13a7008fe234f4c0089e2190a0010ace90072018c493f3c503e55460184743465b567c00d8f34114933ac00881bb234226dd66733a9d368651da88936133e42070f35ec3b9dd749ab71d701becc693f324188e08d8db790676e364d200db61080ee9330c0e1146c919f21ab4ee8f614a4671020e58a7cf2304069cbc3314c613946070d743673fe63835e208dd63ad3c4810e8b4da5b0eb8dd749c091cbc854304c630881c2718535d08aa4ec545d120f68268d8ec07114cce5df3d11b1c47f4813f26f4df189211fd8fe384fdc62419c4d199f81a4443ec3a4ec51d066494f61b8378dc548c6a9a5c6b0ee2f198435eb66e338c87cbd68f4802070437f12818bb4e65d2074b354b50f36cbc2b5a1fa4fde62489498c9b138492e5baf6e0b08992244eb0eb54bcf132ddd226e3d84b43845da7e638e47c82c0cef1ee011d1e4c927440c08573e9387e1a0d4075026e8bc21ad88b0252c524ed034a68637704033a6ee888a4fdbd88a0c487035403937ee310e1494ce7c68553491049938837394a07038471350ac25a591f67f15b74f55efc3a80c910836e0fd0b68d47883c4ca3015548db9177001338460425b85a739c0a9e066430023e6d38801801378802721f4d621c906dcf4bdcae539188bfce9555693f75d64b4d36c788ec1cef1ec76f51b48d312299f6f9421b80212aed3e5b54d25876906f299fdb9a51dde22570ba73bc9b69663ccf3493ed266986be495a4691596b98a7db433e4c4342abd0494fc7920e9e7bf1ca7d4025fa95db7de5be88de46f13402be184ae08215e08315e0be722f5d0d9d490a6b5695c06a4ee5928a1893efafd2e4fc191e52066a19afd09f0863c0558953d98ef01425009c9cf6cf09a28d9bcd67888c62affb08d109fa169dd781229bfe946de52c12205fabfa46771c9468017e4e11edba63f6d0fd39ef2d413e35e831b30ea2e56e822041cfd2900447c1904932a626d33ed7a10727244e9adbfcffcef16e6318bb75a7d9849e073ee7dd7d0ec668dc4749a380f720d3591179fa1104800209156ec39ea0730c4e4eb3744e5411adb80fd4278808ff222a46b4fc5979f958f30aa721e114647974c84a0ce40f3a472339d805f4271d3c7a0d7929edfb1079088d8f0649302185ca092b7c8d5929ad2dc0224c51cc130d6521c799693bc661866935c5283930c5a80edea4c9397fa459c8a7cc7c6d58e5325b07cc1ad428d1cda653814c1fd245c8110375447d83aad1c949ab7b5a732abc4a9c902aacf196039bd4757be5b27221660a70b383216abb9776988d67a0073aa0496d52498d7dd003ab736b1872d70370324191575d50b10e34f76b4e85fa7001a52f81d1106586e1c2a9bcc75e4dde07a74c4951e8a2cd21fa76e79c20ccc691a243e5a8f10c267804432bf09a53097c56fb3ff44014840c5da1a2a2206480642750f542e787ea83d9cbe72961b22586abeaf6c9c0ad83226acb741827b8b18fa655370715584503f8300811951cb0425b0b455ee3cc09e361e33eeaa743bfeaeab9030ae4802ef81176eb6226550be5356ac0049e165e444148a7e8208e3001f87cdc8f434c659d1a6cd003941daf2133edb94a5f19a2d203ae31be6eb6a2413bad68e0f03a57933d64b6aa07980ec957e02a8a960a7d9447ea78460bc92cdff04584a998d2c2947fcdd7389aa088114c4b892e4513fa9076c4b498d529ba92c6caf96f54e7607886e478bc363db9230209aa1619ce653144515537ad51b96c995299f13cb99eacba06ee600431e82314014a94cbe58e63349466dcaa009591d7ba4f3da2fa4eb59e3fad739dcc9b60869f539828c9f2b3ba8098b54f73e6e527b2f66f5092889997c5682ef7cbc1715d9019048c22cfa6b28e50e45161a8b6eb40025694b935f53011f31f349b24f662f0eeb7bff8fe6ffef39fffeb2ffff8fffed322ba651f9a65d723d802274369a6534db0e888af481b6ca83a8fd1acaa6bd0e6091a9cd918748806677319d45938ee12c44de95770caa9d6546435b97e4e1b9dc184d9d06778084e4e0dbfdea95032c4d47a118dc5e43279734b345c38c3f2d82b7080cb52a9c5e254359b8cfd07a9505c06692e5d3f287a94fa928851843ae0423a3bcbb194e3f439a69040141380a2381d8e327388b26d92f32bad5e0fe52b65c7449934e1d988076a5844e593e0b4c1d74d8278bbcb392977371dd9d75798851c2d1a7262aafb859ad195b88be673fd0bca1ef075121024d65c907ee30bb0107943a4042b3ba645635407061dcba23a653d53c42ceec40a70e7fb4636f1c82021171df618c4cdecf5849bc4c1080dde2a41a6606be02e6897111c21e481368fd554192a4724a9b94276e8ef3a78ada4c01b0b29eab0aa7b1eebe0a4757ae5292cbb02417406c34069a039632bddbf3a607871f5cb587f85a1cdbb92571e4f6541f86f8ab71cd5617e50178fe26d6fdce0885c688fa1dcd792686b4d258c29e8e59c2cea3fab21a4f02234adcace6ae279e30811cec5442a40d19c156b71b5c4bc7e7061dd39de15cb4521adc6f2bddb03423085be0d581893229f35c98f10d953457478add2541e107bc8c4872aa8212280f702449861b9905885d6e6aab67cdcf50a64056852c45c945380b1500033bcee8c279303b7f2232c8cbb05aa322d0ae822afe29b38e5fb2b54e63c34c1b1889fda9caa222a808d2f5dd359f11191908c75bae2b2c454c67aca9868495b156080c61372ee2ea2b9d880d36ab7f0569f2843d1ad79c6d586a97686a42fb724c6c5861cf32caf0a3e900587cf6983211951265087a7b32c0e968635d911f505ca8324d6b09d32f92aaa3d9f8546fca2dc0ecca5630908256113932ade686ea4848a132f11e1155b88b544be2c5cbc65c2ba0a9f84c39d979d9b30ae04a4958f37740f0da6ad58355e91f98b7823bbb085953821cbadd06c5e4a115689a3328764ad60ad048bf130c5a16184e7edded53cfab5a9ba06f58b80b90bc9e5a1b895327aec049b5b0c57a75844ffde03b91948ee1236d332a432445136a0dc88daa6a645f64b37c15ea224f0cf9f218ce1107da03db0e3044618b27dc6d799feacdb6025db7815d9a658833da7558e826104499aa05c152c9f1b1b4b05387aa74a338bcce396d0df57e616638b0d92e6c64bba9202fd380e25767c6da577bbce58fba36058ed93c1f1ec31c4a33a505f8f8221fb25f76597ddeca23de6b7bace3223c603b5a523aa03b699c7ee650e4e43c2eb69a4f355f468f63294e5ebc901ed498a1d3935e605873330e62d3d7d1862a4b77b783b3e623b7267b0101736182463c3c912fb3cd9be74844958568bd094780a452c6f25e566b0a477f198376c309954fb394c08f722b2b1761cf395670002fab3261433054c51efa73ee881317c8bd456ea56cda9f014a5c64e307c1079018ca8fe7fc10054fba95f0729ff1ed4f452b69ffaaa67d6e571bc271b884e6b1c07233b4680ccf7a6bbaacd518e8f107971fcd3e78f09997c04d568f466518c656a319b6b51546695e35960d1ab64167873559d81cd8d145d118ea68d16d04515fd2f114bc92ce64a98356745bac5f16cef7e2657224b4ee5659cf2ac062a3fab1d557e16a73cab601ca72cefc10f63482545f28b3da73576e3c80f92311612a8200cc47343070f115b145314abd750b6b70bacc9abde21af6051ba46d3cc1e997c2876c9e8cf6bea47d17ae96df00faa17cd9e0a5ab1c0c6f7a913b33d8b87b792ece6b68aad65b2c26eba6bfb171ceaa2e88da5f3ca79f91e625e198846e0ab9c946fe294c79a06300c4136b8c9d28729852c85dae30573e8cc2d6c15c19c62851b2793a19ec5d7c9ccd95c94c35415aad99ce04639c98f10010ac2e7982b48bed0b83251541f7085d8ed498bdb2a26ed64486a70a54d910c7c5bd14970da504af42ed860552b833822419422a7423bae90d9de7da6ab396f6c3098aa5f6109122ca43288a3b3c61e8961358848b5a41185c9166974282a22156371eac3f10cac0089538d9346253203c090c83c3973d6b1c7b3150958ad596db340f49a9303fb7a95cc46108f162e54e7ae5433305898ef52b26d0e3522cf654549e00a4693228159c790c955c98071f9bbdd465bb56bf1408cd07be1451696e2899e252b6a96cca506dc06941e6e32021f6800f9cd9dd2e03b536a426765f69be94248ed4b89d953dcabaac97e8dbabdb99b6a5a8b167634e5b60c975a56808d8da83de3712d5b4bf2a3a75990af20d2c032a866e342f0fc794ab855b506be4c7875d0bab631e4c4e7ba5cce1ecedf60a43608673619b9672ba4cd70f5958dc939fbf439355ca63b5fe2d06b0f58bac0390ff8250c532413800dfff68c3e972b02b00d70180c10887d85273f40349e84688c2282792a82f2602807e913d03f07ac8b4686caa70126d4d1e7bff4f904a3b8068ea670520deae00d45adc689b8009393e0b40e26276f4e418ffdabb347e0d20ee4298aaa359ddc2c8088f161db21b5f2a6181bfdb3f08b6a3a61e68531ef1e4343fc90fc5299f47419c5765ec1184e28a41840b39f3a5d65459c5930f224fc80343856b460e79c01afa6611643939b1732dd83c07e48abd640b3a91340d250af9cd3d05cb30820b82a2625e192b08c97aa4690f6fee425b6d9d49df38304f9c12c6732448a111e87b27f431e78a42018961561646d45fd94d7751047e8c94bedae3c79c9dd13b637a6bc35dc389a8401a9b2ca8d27e8bc4e5d0bc3d6f33db2bb60ad6874446a03788bce857551506a166f40da916e4f0ed8057d1c441e9a19ae3cf768b6491cb0be4f564f97b07f062a8001bc024242627bd9e8096fc98aa80fab2a9bc9b20cbff6a951cad49028e898054cfff484f2a932140559392c9451209c4a738c557294f1b02e01300afb09826f0579a64ce996636b4bb17891d9ad1abc25bd5517aa761885da63d60eb3043736c0dd13d08c877462ca2c2e39c5cd364e25ef7ae71b4b4e6987c4a8a134d472ecccb6645c0d7ca081dfcb8f90623ae3faa5dc632e6f3187918263c69c2fa6be1885ac5d9624a507e8e41a07d11e15b5ecb08c83880b5bb767023b69293ee5d86daa20b1c231db290b701728d08c33aa7bd97b50138fe594b381514c5c3c566673d9d9a956f026874ba39987706a845f1ea1e823c434937c9fcadd01cde623448c421d65357683ca8f44ed459394607636887d5555025640ab3c4f09af7372cabfca2a312fa0759ec683b7241823e90d2930a128d0a148de8d198b0cf287ce324e98ddffb2f877c2b513dd0854cb4ed3014b885746354b829aa0d9ec93c19c586f41326e14f12d83a6293a845332cb3121a1cfcce35c717486124e5915cabd0549a10ef08b275558034d308664d43888a7ed160b126c31775bf83d11c56ba69c1eb63ff61021e9fce403b869c1df582aa82b9b6703b9741acca8f41723b98519a2e3b97946ba974e85c4045a4d504e93a5a61a93dd37c4a4513a265750e7327e4125bbab40f0604ce0d50115ebae34ed66310b9b5c529d67e24797cb459d77a9d9d19033b4889028a98be9702186a79e950b656dbe10a3c990087c60a2712f6794aede796a188e673a7a9741c5ecf10bd523c3e772a980bbec7d5e34da7575c03df930a176a38b42a4bd6466979e10b852a43dd37176eb91e90b156db75bb7a3606892fc116c1bcef6685a362b3299b3be60c1895fd330a299a52ad1154668a62c22cf3ce1cec8c9a958069b96911f6c637eb93e94cbbb00b6a3b6bc3eefa2dc82e408be91fdb0c3d2d663371e4f42449040982dd38d6d3a5e68e161ce54e7f9270d0f1bd6199f8eca1a0456ded21925d977a543c5762bc17bb6d8889ca4690b91631535101c4823473b350bd3e3992a2ec84aaf40a2aaab45a4a7e95d4aa571007f890a8d622604cb2a6ecba8340de3aa6a4db7cc8dd3021df70c2543f47195dc38d7a5fde6826534da0b8c92855a2dc528296ab63cdd8fd1ac781102c334db165bd522d772d7508dd7578f79026ea41f4b807d3405c945e2261ad27acef0068a536064d19c79f1d5aa33cf45aa3b059c8645549502957596d2a0b2f23c154ad5116407fd963c41282b2bbf553cb8e60942d1baf4c20ac3ffb34b9e53814902d9599d93d393553aec1717ed3ae85c5e3a95281dd382b64151eb4b1080bb204ac75f82606585d12419659de93dc51716475d505730655e2d7d2046d0ce0838094e4f5aa7a76cbbf043f6d1167db0bdd6f9065474501217ce45796f663f97e8e00a5efe12d0025fbaf74a5f5de422bb97a0d9f4e384efa2da2a679709365b6a9cfe139658dda1b11f13a0b4e19c2cc0231479d638d5875d5d147b2db7bd65c9815756f025f68bfbe5458c6ee6e797c233626c46391ccd2100e291e9c353d0b32a995d251fefcaa11bcb1019ae79b1d035431f05fb4296f3a9ff0ac204a503fbc103055971c9e6e76178865e9059ccf63d79b2018fd6c8dd4f1e43cf85fc08cb3abf821831e9a1e82e1ba933fabd2806d074a9b96d600ee47d14da527458ea94d9d64cc3134955997216ce2ecdb7cae1c1d2ac96d8ae94c85d375dc7d4d7f361a9749d39d94bec0e9e9512b6b04ca6224bf8e365d851a778db12a02c6378ad5c277673d0b558370f92649c48562a444ecda9729193bdb2f43083e2569ec57c4615d9daaed9d3c70c84bec8a2f35112c6e6b1d12e81865acae407e7af7ffcc10fdeab98fa090f9e57efb55bad56ab762a33bee49e88710abf22764d6cdbdfbcf77ba2775eb10e5e039dec7000138cf6227e4bc649fbb40edaad3ad85853e79505f4bb3dc050691990807ca6f6907fc05cb8a6ba9c965f91f39aed06a80766528cbc8776ef7e95e708b321038402cd66c395a479f3c2ab6679e79adf2045d854002579c2624917aa1c3d910ea81778819f49d4636b53792ea197e5dd85c84428d467375d88fb4a44fe881e1a55205215b278aff48ae0443ea291cdc00e25bffbd5fffee3fffdfbeffffb3f7dffcfbf535bf3d43a8b5465993d254e1065359299d26d249864d31bca07cd807fdd016b36c11f7eff777ffa5fffff0fbfffdb3ffe9fff293c4e23115dec8396a02f67add20d9a027d5d2d3f9d908b7fb0e53f5dfb33a753c63bbad9c5fb184ec4d9d8ae3814368693135e4e1bb3045201adb5c8550a7c76a3cd2cc0848313b04f4a2cfce997a23297b06200472cebf325f52540dbd2ae165f03237851eaaa19e96985fb616c9aeb3ebfaa4380cd5c7dc111d26128b9e7291f19d782c92bc12a8bc966512f16eaba123f755a18fdb2b22225f74ffff06fef7efdcb8ce41e05437d7995b937530c391ac146350f9bcd77bffcd5bbdffc170973e19d771d7dbd1deda596fd49b1b9e185773c6c7ab3cbee388c8f78d15de7e617dddd80ee1c8c4f57ddcdbdeaeeee87b8e90ebcfbf57f7bf7bb7f91d388688b92d90e583029058e730e3ecd435786f917710ebcfbd5fff8f37ffc85c6959f8e97822d6343d7c482b75f80c46b16d061a7da1946dfffe36fdffdeb6fa4532196a82a2051e64a18a65712f1e1509e038077ae93a9e4cdf43759c1e4d60cadf2b5a5587a680da4966193941f8e597a0d384909f811369490827d95e342140ae5a1686c5c27a82c6d69eff976f1db85572c2d6cafef5229f0ce4da4ca2865de9277a4d0ceaf75250a6d587a114aa616e7c5d2ace0a41b071b3e8c4ccef1e2f49d8956aca5cbd66c128449e62e610b303954252c48f44d2e7c171606519545a9e5c0b3da0426a41aa16955bef8c33ac2fe98340e9220227e95af3e00a60d8368088ed8111f30901cd3032e5efcf003bcff29f7fe2fca4df6e77dbe016cfefbbf5aabadf54efefd5febedcd4feffffa189fc56f3ccabdde88bfd648be84e7afe8251654aa398255f322154c12192c3011623183cefeea4fbff51e46db939f6cfef44d7af6b3fd3bdf0ca2f6fafefe775f9f4de3370f1eaebd72ebafb2c8b196eb9d4ebbb5ea6d7a9b681379ed750fded9ea6f6dc2ce1d7fb38de046ebcee6daba37581b6cb6d0daa08de0d6aa3718b456573b1b77eea0cebabf7ec76fc34ebbd581edf5cdc146bfb3bab5daeeafa1f6c0bfd3dabad3f65bfd8dd53ef4a0b7dad9f4fb7e6bab05573d7fadd55fdb6ac13ef2daab5b2db8e66f323ceeb4fb687513a13ef2fa6bde26f2377c7fb383d65aab5e7f75a3dfefaf6f6eae6f6eb53ba8d5bfb3b189fc96bfb9d5196cf6d7fa03afdf595f858811aaf8facaed9ebc72a34eba31391f751e78c1237c703c7bb1b3b5fff268f6ecfe838df3fe7d6fe7b0cd5a8d9347c993af3a1b7788ffedcefdaff7dfaeee3f2507dfaeee7c7707adafc3adfb3eaf16ec1f3e78b21afdeccd8b87c9ec8ebf799c04876f7e72fe267af478eb0979f22c7de59e5eba4a870bcb8549a2bc273dc69dcc208b7541b7979182dc169709e42b1c0b1390a014e79b76685ba5e8c3a8aad642b4ae5c072d791561d7b228a26096b980d0c42177e11d4021b7bc994aeef327ae30368a569692c26895e92f6664c91da1308cdd3a70a72348bd1d7718c79e7b9d9c972ccf44c51c27612d33613964ccd193a70e97c06dfedddca57787e6f859ce4731e0eaa254b198a3c34e9feda3e9a20b4d8ddb4c458b9b61a481a8c1d57e4cb3991102ce3616c2d102a07e49963b153d137e18e7e4d3e7837f72fe5f8a51326539267d32c8febabe4738dfff6b6fb45bab59ffafd35eed7c7affeb47f994fa7ffdd40f62eb7b2ed5cb1ba54b18e3e5de7319c138886093c4e3b0fc159822afa9f45d927d32f842d531de05c9eec531dffeb68f88b8d08c5fc8e0541ec798e4f2ba9ccae1c1ee0b8c92c2c30388f1d4331feea2841c4032cad4fc9a21262a8b87e2b4e1b3735e68e0b44306024fa071762afb708c9ea07306d8d8b4917dd06271697b59b174b22dc5a2e889bdf5a5e350a33514a8f6c03e9af2efd59a59f2300811e801b7410780eb8486184856eb381e8782b41ea0058dfbfc658982466e910ce0e0b6e28f3e71784b3ea3d64d31abab79c5e24f6c64bb7220db2c39848ead7a04dc301ec070449fb5b756573b2eab23865a5473bf9b85ea391f4056e2b63bab6beb1bbc488e396fe3f287e6987781db664f2fe91f31905dfb485ea83a7c34bb73eb8871eb96d5d1e36a8773a91ca930861e675f95bde5590671a67a0c8ce42a5ff94c316e3c9fa088359a171895f9701ecbf0f41bbb618c11db91e79e8f211c8d7d34bd8f06b187922ad330f4c12182f4b75fab35785975aa7d730adccc335b9e1076ade0900aae491177c4254d1c655d53e1be905a099fd2a6b7800d4aa5b369f7ebf2c0c4fb437a262eec0eefaaee66294096f418be8c7882ceab111c4bae198c6237849081d0aa40fca7265f2797b2740551d2e0aab527754343cd545ec277b1c53b2632cda60749702670a93915d95e4c9b138ade29bffac2315f649083c3dfbcc55f4f58e5ef423421f1c9a5807110d955062dab03f7732c5f89da05aebc185feec243916c92ed585e0c58ec5694a87e21dbf59235f46c3d81627797532903925ab40d8d5b972c3682b1c119da662b2da20fc98b6d5ff58b1f4f573f7da4e7ca2409f4cb3b0d216089a6ecf0b6ed1615b9ab3c4ffe542e17dfa30e30d8e337aa345cb13ce268b12cbd42372c7fe961184352e5d544f2d2523dca74d37c97151fa185ddf9085da52f76b0bdd091cad32db909428a50c94d117c47de52603d575e5eafae3a328f9b93e211f3337982ba0c9ef508f79c8aba6771b2db36375c4825d73d9512caaff0f802c861ba340ec6bc5f8d64d4656ea0ad2a2dc8d494bea1adb228cbd7179e60490b5e9a69a3bc4a5b1359a8f2e573990466dff9e1a9de2a19b12b4692b216c61e7c50982d980c2218568af3dc5c769d2c22faaa832289d70b96654faf1ecf8e9078ffba9eb847b9027986573fb11f0cb660c70fdd5b0050eb987d835d6915a9c86b32ef86870c8b72a1f1aedeca43fb0092a091a995fa234c09502c5802a8f966dd6bdab4b101e29365fb64d93eb06533c5ede35b32367bb6e73aa779bf5a9addda273bf271ed88cd8614ef2e31eb5ed38888c60f93782cee15546292ab5266630af6658e10f14b16e65a1ce3e8db2db383f724348ac585b3fa2558488928d972937b898ab09ab885a200a22427aab84f93c59da52e650ca50db83a822b7a995f495e5918794ad2ba3dcd80b2b3b51776a859c1bb94a0cb3585a583ea2d131b79ef619e57566665b9c5d865712ce408949f051aa1708212beed95d9f492b7c862d0652f4dafb3a5749d252d9298e9f4fa10913e0c6134406e6e5fd2950c162f5c2f2be530b95912b7924aa3cfac31f7dacadb33fe719f681e18d304b9b5e2865f31886744150a73f0d62d5a6d0fefc7e4c12cc0840d0a4b74e5076a78bc5784a6801f84c8a95416462c324a5e07d68452368292d6e6a0d9a455b27d2f35f72cc9865a62581e1b67cb7632c485e35d1d06520a91a3015c3a95c1d8332fbd3d8e9fc65394c8035735c7a9e069400623402b5e389501c408b84c5ebaa0d9a45fb85afe1cb3e04fa5a2237342bbd6642b2a47ac15fdc25b011c8f119d80ecf50b3cef2d8ffe669123d791cd2cdfa4ced4a1274672e7542479aeca2f6bf2cbbafcb25176f1ee9c31bb943c3066266385e0028a3cc60512f369800391617e25665c6fa2957066ce8ae6e3b0ca505cdd7ccac73e429ebe5d9cda07c27cd121223bac093b40cf83e1300d490180581050b1668b02637afc40598b9f3e9f3e9f3e9f3e9f3e37fdfc7b000000ffff67afae7d009a0000c0c28080f86e05b86bf869f867f865a10399b6b409e24d7805dc4ebcc35a4ef3f6d374e72b1d8b5b371cea9bb67989693bb841ea88d0b224ef2eb5e6fac8f4acee3e5c6204fb77d333521b4807eca00e3fdc9f22dfd162345b11d6e949c69fc9e736659f2970424e89eec7aae67967451a331401")
	hash := crypto.Keccak256(rlpData)
	t.Logf("Hash:%x", hash)
	tx := &modules.Transaction{}
	err := rlp.DecodeBytes(rlpData, tx)
	assert.Nil(t, err)
	lockScript, _ := hex.DecodeString("76a914381db54b7e5d5d4f4fb1cd63cb5a94d5e839810488ac")
	err = ScriptValidate(lockScript, nil, tx, 0, 0)
	assert.Nil(t, err)
}
