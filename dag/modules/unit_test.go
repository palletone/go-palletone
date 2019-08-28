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
	unit := NewUnit(&Header{Extra: []byte("hello"), Time: time.Now().Unix()}, txs)
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
	addr := common.Address{}
	addr.SetString("0000000011111111")

	auth := Authentifier{
		Signature: []byte("1234567890123456789"),
		PubKey:    []byte("1234567890123456789"),
	}
	w := make([]byte, 0)
	w = append(w, []byte("sign")...)
	assetID := AssetId{}
	assetID.SetBytes([]byte("0000000011111111"))
	h := Header{
		ParentsHash: []common.Hash{u1, u2},
		Authors:     auth,
		GroupSign:   w,
		GroupPubKey: w,
		TxRoot:      common.Hash{},
		Number:      &ChainIndex{AssetID: assetID, Index: 0},
	}

	newH := CopyHeader(&h)
	//newH.GroupSign = make([]byte, 0)
	//newH.GroupPubKey = make([]byte, 0)
	hh := Header{}
	log.Printf("\n newh=%v \n oldH=%v \n hh=%v", *newH, h, hh)
}

// test unit's size of header
func TestUnitSize(t *testing.T) {

	key, _ := crypto.MyCryptoLib.KeyGen()
	pubKey, _ := crypto.MyCryptoLib.PrivateKeyToPubKey(key)
	h := new(Header)
	au := Authentifier{}

	address := crypto.PubkeyBytesToAddress(pubKey)
	log.Println("address:", address)

	h.GroupSign = []byte("group_sign")
	h.GroupPubKey = []byte("group_pubKey")
	h.Number = &ChainIndex{}
	h.Number.AssetID = PTNCOIN
	h.Number.Index = uint64(333333)
	h.Extra = make([]byte, 20)
	h.ParentsHash = append(h.ParentsHash, h.TxRoot)

	h.TxRoot = h.Hash()
	sig, _ := crypto.MyCryptoLib.Sign(key, h.TxRoot[:])
	au.Signature = sig
	au.PubKey = pubKey
	h.Authors = au

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
	h.Time = time.Now().Unix()
	h.Extra = []byte("jay")
	index := new(ChainIndex)
	index.AssetID = PTNCOIN
	index.Index = 1
	//index.IsMain = true
	h.Number = index

	h1 := CopyHeader(h)
	h1.TxRoot = h.Hash()
	h2 := new(Header)
	h2.Number = h1.Number
	fmt.Println("h:=1", h.Number.Index, "h1:=1", h1.Number.Index, "h2:=1", h2.Number.Index)
	h1.Number.Index = 100

	if h.Number.Index == h1.Number.Index {
		fmt.Println("failed copy:", h.Number.Index)
	} else {
		fmt.Println("success copy!")
	}
	fmt.Println("h:1", h.Number.Index, "h1:=100", h1.Number.Index, "h2:=100", h2.Number.Index)
	h.Number.Index = 666
	h1.Number.Index = 888
	fmt.Println("h:=666", h.Number.Index, "h1:=888", h1.Number.Index, "h2:=888", h2.Number.Index)
}

func TestHeaderRLP(t *testing.T) {
	key, _ := crypto.MyCryptoLib.KeyGen()
	pubKey, _ := crypto.MyCryptoLib.PrivateKeyToPubKey(key)
	h := new(headerTemp)
	//h.AssetIDs = append(h.AssetIDs, PTNCOIN)
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
	h:=mockHeader()
	data,_:= json.Marshal(h)
	t.Log("Header1",string(data))
	headerHash:="0x4dcf5cffcc5eb4f103d9222d4551e337c73f7f5d0c4f50de170920cc42db302b"
	t.Logf("Header Hash:%s",h.Hash().String())
	assert.Equal(t,headerHash,h.Hash().String())
	h2:=&Header{}
	h2.CopyHeader(h)
	data,_= json.Marshal(h2)
	t.Log("Header2", string(data))
	assert.Equal(t,headerHash,h2.Hash().String())
	h2.ParentsHash=append(h2.ParentsHash,common.HexToHash(headerHash))
	h2.Authors.PubKey=[]byte("Test")
	h2.Number.Index=999
	h2.Extra=[]byte("dddd")
	data,_= json.Marshal(h)
	t.Log("Header1", string(data))
	assert.Equal(t,headerHash,h.Hash().String())
}
func mockHeader() *Header{
	key, _ := hex.DecodeString("ebe665c202f9393b85fe9bddbc31f39f7ad9a1eb14149a60f4ff23e806c111a6")
	pubKey, _ := crypto.MyCryptoLib.PrivateKeyToPubKey(key)
	h:=&Header{}
	h.GroupSign = []byte("group_sign")
	h.GroupPubKey = []byte("group_pubKey")
	h.Number = &ChainIndex{}
	h.Number.AssetID, _, _ = String2AssetId("DEVIN")
	h.Number.Index = uint64(123)
	h.Extra = []byte("Extra")
	h.CryptoLib = []byte{0x1, 0x2}
	h.ParentsHash =[]common.Hash{
		common.HexToHash("57c56162990aac482ae2b66196cd1f5129e6f026578470ab105042bf42d6a2dc")}
	h.TxRoot = common.HexToHash("c35639062e40f8891cef2526b387f42e353b8f403b930106bb5aa3519e59e35f")
	sig, _ := crypto.MyCryptoLib.Sign(key, h.TxRoot[:])
	au := Authentifier{}
	au.Signature = sig
	au.PubKey = pubKey
	h.Authors = au
	h.Time = 123
	h.TxsIllegal=[]uint16{666}
	return h
}