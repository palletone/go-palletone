// Copyright 2014 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package crypto

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"testing"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/hexutil"
)

var testAddr = "P136gdm7CfJcAeG2RFZNXvwwteg3uGzVqr5"
var testPrivHex = "289c2857d4598e37fb9647507e47a309d6133539bf21a8b9cb6df88fd5232032"

// These tests are sanity checks.
// They should ensure that we don't e.g. use Sha3-224 instead of Sha3-256
// and that the sha3 library uses keccak-f permutation.
func TestKeccak256Hash(t *testing.T) {
	msg := []byte("abc")
	exp, _ := hex.DecodeString("3a985da74fe225b2045c172d6bd390bd855f086e3e9d525b46bfe24511431532")
	checkhash(t, "Sha3-256-array", func(in []byte) []byte { h := Keccak256Hash(in); return h[:] }, msg, exp)
}

func TestToECDSAErrors(t *testing.T) {
	if _, err := HexToECDSA("0000000000000000000000000000000000000000000000000000000000000000"); err == nil {
		t.Fatal("HexToECDSA should've returned error")
	}
	if _, err := HexToECDSA("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"); err == nil {
		t.Fatal("HexToECDSA should've returned error")
	}
}

func BenchmarkSha3(b *testing.B) {
	a := []byte("hello world")
	for i := 0; i < b.N; i++ {
		Keccak256(a)
	}
}

/*
func TestSign(t *testing.T) {
	key, _ := HexToECDSA(testPrivHex)

	// t.Logf("Address is :%s", PubkeyToAddress(&key.PublicKey).String())
	addr, _ := common.StringToAddress(testAddr)

	msg := Keccak256([]byte("foo"))
	sig, err := Sign(msg, key)
	if err != nil {
		t.Errorf("Sign error: %s", err)
	}
	recoveredPub, err := Ecrecover(msg, sig)
	if err != nil {
		t.Errorf("ECRecover error: %s", err)
	}
	pubKey := ToECDSAPub(recoveredPub)
	recoveredAddr := PubkeyToAddress(pubKey)
	if addr != recoveredAddr {
		t.Errorf("Address mismatch: want: %s have: %s", addr.Str(), recoveredAddr.Str())
	}

	// should be equal to SigToPub
	recoveredPub2, err := SigToPub(msg, sig)
	if err != nil {
		t.Errorf("ECRecover error: %s", err)
	}
	recoveredAddr2 := PubkeyToAddress(recoveredPub2)
	if addr != recoveredAddr2 {
		t.Errorf("Address mismatch: want: %x have: %x", addr, recoveredAddr2)
	}
}
*/
func TestInvalidSign(t *testing.T) {
	if _, err := MyCryptoLib. Sign(make([]byte, 1), nil); err == nil {
		t.Errorf("expected sign with hash 1 byte to error")
	}
	if _, err := MyCryptoLib.Sign(make([]byte, 33), nil); err == nil {
		t.Errorf("expected sign with hash 33 byte to error")
	}
}

// func TestNewContractAddress(t *testing.T) {
// 	key, _ := HexToECDSA(testPrivHex)
// 	addr := common.HexToAddress(testAddrHex)
// 	genAddr := PubkeyToAddress(key.PublicKey)
// 	// sanity check before using addr to create contract address
// 	checkAddr(t, genAddr, addr)

// 	caddr0 := CreateAddress(addr, 0)
// 	caddr1 := CreateAddress(addr, 1)
// 	caddr2 := CreateAddress(addr, 2)
// 	checkAddr(t, common.HexToAddress("333c3310824b7c685133f2bedb2ca4b8b4df633d"), caddr0)
// 	checkAddr(t, common.HexToAddress("8bda78331c916a08481428e4b07c96d3e916d165"), caddr1)
// 	checkAddr(t, common.HexToAddress("c9ddedf451bc62ce88bf9292afb13df35b670699"), caddr2)
// }

func TestLoadECDSAFile(t *testing.T) {
	keyBytes := common.FromHex(testPrivHex)
	fileName0 := "test_key0"
	fileName1 := "test_key1"
	checkKey := func(k *ecdsa.PrivateKey) {
		addr, _ := common.StringToAddress(testAddr)
		checkAddr(t, PubkeyToAddress(&k.PublicKey), addr)
		loadedKeyBytes := FromECDSA(k)
		if !bytes.Equal(loadedKeyBytes, keyBytes) {
			t.Fatalf("private key mismatch: want: %x have: %x", keyBytes, loadedKeyBytes)
		}
	}

	ioutil.WriteFile(fileName0, []byte(testPrivHex), 0600)
	defer os.Remove(fileName0)

	key0, err := LoadECDSA(fileName0)
	if err != nil {
		t.Fatal(err)
	}
	checkKey(key0)

	// again, this time with SaveECDSA instead of manual save:
	err = SaveECDSA(fileName1, key0)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(fileName1)

	key1, err := LoadECDSA(fileName1)
	if err != nil {
		t.Fatal(err)
	}
	checkKey(key1)
}

func TestValidateSignatureValues(t *testing.T) {
	check := func(expected bool, v byte, r, s *big.Int) {
		if ValidateSignatureValues(v, r, s, false) != expected {
			t.Errorf("mismatch for v: %d r: %d s: %d want: %v", v, r, s, expected)
		}
	}
	minusOne := big.NewInt(-1)
	one := common.Big1
	zero := common.Big0
	secp256k1nMinus1 := new(big.Int).Sub(secp256k1_N, common.Big1)

	// correct v,r,s
	check(true, 0, one, one)
	check(true, 1, one, one)
	// incorrect v, correct r,s,
	check(false, 2, one, one)
	check(false, 3, one, one)

	// incorrect v, combinations of incorrect/correct r,s at lower limit
	check(false, 2, zero, zero)
	check(false, 2, zero, one)
	check(false, 2, one, zero)
	check(false, 2, one, one)

	// correct v for any combination of incorrect r,s
	check(false, 0, zero, zero)
	check(false, 0, zero, one)
	check(false, 0, one, zero)

	check(false, 1, zero, zero)
	check(false, 1, zero, one)
	check(false, 1, one, zero)

	// correct sig with max r,s
	check(true, 0, secp256k1nMinus1, secp256k1nMinus1)
	// correct v, combinations of incorrect r,s at upper limit
	check(false, 0, secp256k1_N, secp256k1nMinus1)
	check(false, 0, secp256k1nMinus1, secp256k1_N)
	check(false, 0, secp256k1_N, secp256k1_N)

	// current callers ensures r,s cannot be negative, but let's test for that too
	// as crypto package could be used stand-alone
	check(false, 0, minusOne, one)
	check(false, 0, one, minusOne)
}

func checkhash(t *testing.T, name string, f func([]byte) []byte, msg, exp []byte) {
	sum := f(msg)
	if !bytes.Equal(exp, sum) {
		t.Fatalf("hash %s mismatch: want: %x have: %x", name, exp, sum)
	}
}

func checkAddr(t *testing.T, addr0, addr1 common.Address) {
	if addr0 != addr1 {
		t.Fatalf("address mismatch: want: %s have: %s", addr0.String(), addr1.String())
	}
}

// test to help Python team with integration of libsecp256k1
// skip but keep it after they are done
func TestPythonIntegration(t *testing.T) {
	kh := "289c2857d4598e37fb9647507e47a309d6133539bf21a8b9cb6df88fd5232032"
	k0, _ := hex.DecodeString(kh)

	msg0 := Keccak256([]byte("foo"))
	sig0, _ := MyCryptoLib.Sign(k0,msg0)

	msg1 := common.FromHex("00000000000000000000000000000000")
	sig1, _ := MyCryptoLib.Sign(k0,msg0)

	t.Logf("msg: %x, privkey: %s sig: %x\n", msg0, kh, sig0)
	t.Logf("msg: %x, privkey: %s sig: %x\n", msg1, kh, sig1)
}

func TestPubkeyToAddress(t *testing.T) {
	prvKey, _ := MyCryptoLib.KeyGen()

	t.Logf("Private Key: %s", hex.EncodeToString(prvKey))
	pubKey,_ := MyCryptoLib.PrivateKeyToPubKey(prvKey)


	t.Logf("Public Key: %s", hex.EncodeToString(pubKey))
	address := PubkeyBytesToAddress(pubKey)
	addStr := address.Str()
	t.Logf("Address: %s", addStr)
}
func PubkeyToAddress(p *ecdsa.PublicKey) common.Address {
	pubBytes := compressPubkey(p)
	return PubkeyBytesToAddress(pubBytes)
}
func TestImportPrivateKeyAndGenerateAddress(t *testing.T) {
	prvKeyHex := "0x734e7c08b3651305c45422b9dc1e3fc0d67bc2bf8f3b50bff28a6760fb3e1057"
	prvKeyB, _ := hexutil.Decode(prvKeyHex)
	prvKey, _ := ToECDSA(prvKeyB)
	pubKey := prvKey.PublicKey
	addr := PubkeyToAddress(&pubKey)
	t.Logf("Address:[%s]", addr)
}

func ExamplePubkeyToAddress() {
	prvKeyString := testPrivHex
	prvKeyBytes, _ := hex.DecodeString(prvKeyString)
	prvKey, _ := ToECDSA(prvKeyBytes)
	pubKey := prvKey.PublicKey
	address := PubkeyToAddress(&pubKey)
	addStr := address.Str()
	fmt.Println("Encoded Address Data:", addStr)

	// Output:
	// Encoded Address Data: P136gdm7CfJcAeG2RFZNXvwwteg3uGzVqr5
}

func TestScriptToAddress(t *testing.T) {
	redeemScript := "2 04C16B8698A9ABF84250A7C3EA7EEDEF9897D1C8C6ADF47F06CF73370D74DCCA01CDCA79DCC5C395D7EEC6984D83F1F50C900A24DD47F569FD4193AF5DE762C58704A2192968D8655D6A935BEAF2CA23E3FB87A3495E7AF308EDF08DAC3C1FCBFC2C75B4B0F4D0B1B70CD2423657738C0C2B1D5CE65C97D78D0E34224858008E8B49047E63248B75DB7379BE9CDA8CE5751D16485F431E46117B9D0C1837C9D5737812F393DA7D4420D7E1A9162F0279CFC10F1E8E8F3020DECDBC3C0DD389D99779650421D65CBD7149B255382ED7F78E946580657EE6FDA162A187543A9D85BAAA93A4AB3A8F044DADA618D087227440645ABE8A35DA8C5B73997AD343BE5C2AFD94A5043752580AFA1ECED3C68D446BCAB69AC0BA7DF50D56231BE0AABF1FDEEC78A6A45E394BA29A1EDF518C022DD618DA774D207D137AAB59E0B000EB7ED238F4D800 5 CHECKMULTISIG"
	address := ScriptToAddress([]byte(redeemScript))
	addrStr := address.Str()
	t.Log(addrStr)
	if addrStr[1] != []byte("3")[0] {
		t.Error("Invalid script address")
	}
}

//func TestNewContractIdAddress(t *testing.T) {
//	contractId := "0000000000000"
//	address := ContractIdToAddress([]byte(contractId))
//	t.Log(address.Str())
//}

//func TestToWIFAndFromWIF(t *testing.T) {
//	prvKey, _ := GenerateKey()
//	wif := ToWIF(FromECDSA(prvKey))
//	assert.True(t, wif[0] == 'K' || wif[0] == 'L', "Invalid WIF format")
//
//	pk, err := FromWIF(wif)
//	if err != nil {
//		t.Errorf("FromWIF error:%s", err)
//	}
//
//	assert.True(t, bytes.Equal(FromECDSA(pk), FromECDSA(prvKey)), "Export private key not equal import key")
//}
