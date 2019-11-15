/*
   This file is part of go-palletone.
   go-palletone is free software: you can redistribute it and/or modify
   it under the terms of the GNU General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.
   go-palletone is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU General Public License for more details.
   You should have received a copy of the GNU General Public License
   along with go-palletone.  If not, see <http://www.gnu.org/licenses/>.
*/

/*
 * @author PalletOne core developers <dev@pallet.one>
 * @date 2018
 */

package modules

import (
	"encoding/hex"
	"fmt"
	"log"
	"reflect"
	"testing"
	"time"
	"unsafe"

	"encoding/json"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/crypto"
	"github.com/stretchr/testify/assert"
)

func TestNewUnit(t *testing.T) {
	txs := make(Transactions, 0)
	h := NewHeader([]common.Hash{}, []byte("hello"))
	h.SetNumber(NewPTNIdType(), 1)
	h.SetTime(time.Now().Unix())
	unit := NewUnit(h, txs)
	hash := unit.Hash()
	unit.UnitHash = common.Hash{}
	if unit.UnitHash != (common.Hash{}) {
		t.Fatal("unit hash initialized failed.")
	}
	unit.UnitHash.Set(unit.UnitHeader.Hash())

	if unit.UnitHash != hash {
		t.Fatal("wrong unit hash.")
	}
}

// test interface
type USB interface {
	Name() string
	Connect()
}
type PhoncConnecter struct {
	name string
}

func (pc PhoncConnecter) Name() string {
	return pc.name
}
func (pc PhoncConnecter) Connect() {
	log.Println(pc.name)
}
func TestInteface(t *testing.T) {
	// 第一种直接在声明结构时赋值
	var a USB
	a = PhoncConnecter{"PhoneC"}
	a.Connect()
	// 第二种，先给结构赋值后在将值给接口去调用
	var b = PhoncConnecter{}
	b.name = "b"
	var c USB
	c = b
	c.Connect()
}

func TestCopyHeader(t *testing.T) {
	u1 := common.Hash{}
	u1.SetString("00000000000000000000000000000000")
	u2 := common.Hash{}
	u2.SetString("111111111111111111111111111111111")

	auth := Authentifier{
		Signature: []byte("1234567890123456789"),
		PubKey:    []byte("1234567890123456789"),
	}

	h := NewHeader([]common.Hash{u1, u2}, []byte{})
	h.SetAuthor(auth)
	h.SetGroupSign([]byte("sign"))
	h.SetGroupPubkey([]byte("sign"))
	h.SetTxRoot(common.Hash{})
	h.SetNumber(PTNCOIN, 11)
	newH := new(Header)
	//newH := CopyHeader(h)
	newH.CopyHeader(h)

	assert.Equal(t, h.Hash().String(), newH.Hash().String())
	newH.hash = common.Hash{}
	h.hash = common.Hash{}
	newH.SetExtra([]byte("add extra"))
	newH.SetNumber(PTNCOIN, 22)
	newH.SetGroupSign([]byte("sign123"))
	newH.SetGroupPubkey([]byte("sign123"))
	assert.NotEqual(t, h.Hash().String(), newH.Hash().String())
	log.Printf("\n newh=%v,hash:%s \n oldH=%v ,hash:%s \n ", *newH.Header(), newH.Hash().String(),
		h.Header(), h.Hash().String())
}

// test unit's size of header
func TestUnitSize(t *testing.T) {

	key, _ := crypto.MyCryptoLib.KeyGen()
	pubKey, _ := crypto.MyCryptoLib.PrivateKeyToPubKey(key)
	h := new(Header)
	au := Authentifier{}

	address := crypto.PubkeyBytesToAddress(pubKey)
	log.Println("address:", address)

	h.SetGroupSign([]byte("group_sign"))
	h.SetGroupPubkey([]byte("group_pubKey"))
	h.SetNumber(PTNCOIN, 333333)
	h.SetExtra(make([]byte, 20))
	h.SetParentHash([]common.Hash{h.TxRoot()})

	h.SetTxRoot(h.hash)
	sig, _ := crypto.MyCryptoLib.Sign(key, h.TxRoot().Bytes())
	au.Signature = sig
	au.PubKey = pubKey
	h.SetAuthor(au)

	log.Println("size: ", unsafe.Sizeof(h))
}

func TestOutPointToKey(t *testing.T) {

	testPoint := OutPoint{TxHash: common.HexToHash("123567890acbdefg"), MessageIndex: 2147483647, OutIndex: 2147483647}
	key := testPoint.ToKey()

	result := KeyToOutpoint(key)
	if !reflect.DeepEqual(testPoint, *result) {
		t.Fatal("test failed.", result.TxHash.String(), result.MessageIndex, result.OutIndex)
	}
}

func TestHeaderPointer(t *testing.T) {
	h := new(Header)
	//h.AssetIDs = []AssetId{PTNCOIN}
	h.SetTime(time.Now().Unix())
	h.SetExtra([]byte("jay"))

	h.SetNumber(PTNCOIN, 1)

	h1 := CopyHeader(h)
	h1.SetTxRoot(h.hash)
	h2 := new(Header)
	h2.SetNumber(h1.GetNumber().AssetID, h1.GetNumber().Index)
	fmt.Println("h:=1", h.GetNumber().Index, "h1:=1", h1.GetNumber().Index, "h2:=1", h2.GetNumber().Index)

	h1.SetNumber(PTNCOIN, 100)

	if h.GetNumber().Index == h1.GetNumber().Index {
		fmt.Println("failed copy:", h.GetNumber().Index, h1.GetNumber().Index)
	} else {
		fmt.Println("success copy!")
	}
	fmt.Println("h:1", h.GetNumber().Index, "h1:=100", h1.GetNumber().Index, "h2:=1", h2.GetNumber().Index)

	h.SetNumber(PTNCOIN, 666)

	h1.SetNumber(PTNCOIN, 888)
	fmt.Println("h:=666", h.GetNumber().Index, "h1:=888", h1.GetNumber().Index, "h2:=1", h2.GetNumber().Index)
}

func TestHeaderRLP(t *testing.T) {
	key, _ := crypto.MyCryptoLib.KeyGen()
	pubKey, _ := crypto.MyCryptoLib.PrivateKeyToPubKey(key)
	h := new(headerTemp)
	au := Authentifier{}
	address := crypto.PubkeyBytesToAddress(pubKey)
	log.Println("address:", address)

	h.GroupSign = []byte("group_sign")
	h.GroupPubKey = []byte("group_pubKey")
	h.Number = &ChainIndex{}
	h.Number.AssetID, _, _ = String2AssetId("DEVIN")
	h.Number.Index = uint64(0)
	h.Extra = make([]byte, 20)
	h.CryptoLib = []byte{0x1, 0x2}
	h.ParentsHash = append(h.ParentsHash, h.TxRoot)
	h.TxRoot = common.HexToHash("c35639062e40f8891cef2526b387f42e353b8f403b930106bb5aa3519e59e35f")
	sig, _ := crypto.MyCryptoLib.Sign(key, h.TxRoot[:])
	au.Signature = sig
	au.PubKey = pubKey
	h.Authors = au
	h.Time = 123

	t.Log("data", h)
	bytes, err := rlp.EncodeToBytes(h)
	assert.Nil(t, err)
	t.Logf("Rlp data:%x", bytes)
	h2 := &headerTemp{}
	err = rlp.DecodeBytes(bytes, h2)
	t.Log("data", h2)
	assertEqualRlp(t, h, h2)
}

func assertEqualRlp(t *testing.T, a, b interface{}) {
	aa, err := rlp.EncodeToBytes(a)
	if err != nil {
		t.Error(err)
	}
	bb, err := rlp.EncodeToBytes(b)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, aa, bb)
}

func TestHeader_Copy(t *testing.T) {
	h := mockHeader()
	data, _ := json.Marshal(h.Header())
	t.Log("Header1", string(data))
	headerHash := "0x4dcf5cffcc5eb4f103d9222d4551e337c73f7f5d0c4f50de170920cc42db302b"
	t.Logf("Header Hash:%s, sign:%s", h.Hash().String(), string(h.group_sign))
	assert.Equal(t, headerHash, h.Hash().String())
	h2 := new(Header)
	h2.CopyHeader(h)
	//h2 := CopyHeader(h)
	data, _ = json.Marshal(h2.Header())
	t.Log("Header2", string(data), "h2_hash", h2.Hash().String(), string(h2.group_sign))
	assert.Equal(t, headerHash, h2.Hash().String())
	h.hash = common.Hash{}
	h2.hash = common.Hash{}
	h2.SetParentHash([]common.Hash{common.HexToHash(headerHash)})
	h2.SetAuthor(Authentifier{PubKey: []byte("Test")})

	h2.SetNumber(h2.GetNumber().AssetID, 999)
	h2.SetExtra([]byte("dddd"))
	h2.SetTime(321)
	data, _ = json.Marshal(h.Header())
	t.Log("Header1", string(data), "h_hash", h.Hash().String(), string(h.group_sign))
	assert.Equal(t, headerHash, h.Hash().String())
}
func mockHeader() *Header {
	key, _ := hex.DecodeString("ebe665c202f9393b85fe9bddbc31f39f7ad9a1eb14149a60f4ff23e806c111a6")
	pubKey, _ := crypto.MyCryptoLib.PrivateKeyToPubKey(key)
	h := new(Header)
	h.SetGroupSign([]byte("group_sign"))
	h.SetGroupPubkey([]byte("group_pubKey"))

	asset_id, _, _ := String2AssetId("DEVIN")
	h.SetNumber(asset_id, 123)
	h.SetExtra([]byte("Extra"))
	h.SetCryptoLib([]byte{0x1, 0x2})
	h.SetParentHash([]common.Hash{
		common.HexToHash("57c56162990aac482ae2b66196cd1f5129e6f026578470ab105042bf42d6a2dc")})
	h.SetTxRoot(common.HexToHash("c35639062e40f8891cef2526b387f42e353b8f403b930106bb5aa3519e59e35f"))
	sig, _ := crypto.MyCryptoLib.Sign(key, h.TxRoot().Bytes())
	au := Authentifier{}
	au.Signature = sig
	au.PubKey = pubKey
	h.SetAuthor(au)
	h.SetTime(123)
	h.SetTxsIllegal([]uint16{666})
	return h
}
