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
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/crypto"
	"github.com/stretchr/testify/assert"
	"log"
	"reflect"
	"testing"
)

var tt = int64(1574390000)

func TestNewUnit(t *testing.T) {
	txs := make(Transactions, 0)
	b := []byte{}

	h := NewHeader([]common.Hash{}, common.Hash{}, b, b, b, []byte("hello"), []uint16{}, NewPTNIdType(),
		1, tt)
	fmt.Println(tt)
	unit := NewUnit(h, txs)
	hash := unit.Hash()
	if hash == (common.Hash{}) {
		t.Fatal("unit hash initialized failed.")
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
	b := []byte{}
	h := NewHeader([]common.Hash{u1, u2}, common.Hash{}, auth.PubKey, auth.Signature, b, b, []uint16{}, PTNCOIN,
		11, tt)
	h.SetGroupSign([]byte("sign"))
	h.SetGroupPubkey([]byte("sign"))
	newH := new(Header)
	newH.CopyHeader(h)

	assert.Equal(t, h.Hash().String(), newH.Hash().String())
	newH.hash = common.Hash{}
	h.hash = common.Hash{}

	newH.SetAuthor(Authentifier{PubKey: []byte("test_pub"), Signature: []byte("test_sig")})
	newH.SetGroupSign([]byte("sign123"))
	newH.SetGroupPubkey([]byte("sign123"))
	assert.NotEqual(t, h.Hash().String(), newH.Hash().String())
	//log.Printf("\n newh=%v,hash:%s \n oldH=%v ,hash:%s \n ", *newH.Header(), newH.Hash().String(),
	//	h.Header(), h.Hash().String())
}

// test unit's size of header
func TestUnitSize(t *testing.T) {
	key, _ := crypto.MyCryptoLib.KeyGen()
	pubKey, _ := crypto.MyCryptoLib.PrivateKeyToPubKey(key)
	hash := common.HexToHash("c35639062e40f8891cef2526b387f42e353b8f403b930106bb5aa3519e59e35f")
	address := crypto.PubkeyBytesToAddress(pubKey)
	log.Println("address:", address)
	b := []byte{}
	h := NewHeader([]common.Hash{hash}, hash, b, b, b, b, []uint16{}, PTNCOIN, 0, int64(1598766666))

	sig, _ := crypto.MyCryptoLib.Sign(key, h.TxRoot().Bytes())
	au := Authentifier{}
	au.Signature = sig
	au.PubKey = pubKey
	h.SetAuthor(au)
	h.SetGroupSign([]byte("group_sign"))
	h.SetGroupPubkey([]byte("group_pubKey"))

	log.Println("size2:", h.Size())
}

func TestOutPointToKey(t *testing.T) {

	testPoint := OutPoint{TxHash: common.HexToHash("123567890acbdefg"), MessageIndex: 2147483647, OutIndex: 2147483647}
	key := testPoint.ToKey()

	result := KeyToOutpoint(key)
	if !reflect.DeepEqual(testPoint, *result) {
		t.Fatal("test failed.", result.TxHash.String(), result.MessageIndex, result.OutIndex)
	}
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
	data, _ := json.Marshal(h)
	t.Log("Header1", string(data))
	headerHash := "0x4dcf5cffcc5eb4f103d9222d4551e337c73f7f5d0c4f50de170920cc42db302b"
	t.Logf("Header Hash:%s, sign:%s", h.Hash().String(), string(h.group_sign))
	assert.Equal(t, headerHash, h.Hash().String())
	//h2 := new(Header)
	//h2.CopyHeader(h)
	h2 := CopyHeader(h)
	data, _ = json.Marshal(h2)
	t.Log("Header2", string(data), "h2_hash", h2.Hash().String(), string(h2.group_sign))
	assert.Equal(t, headerHash, h2.Hash().String())
	h.hash = common.Hash{}
	h2.hash = common.Hash{}

	h2.SetAuthor(Authentifier{PubKey: []byte("Test")})
	data, _ = json.Marshal(h)
	t.Log("Header1", string(data), "h_hash", h.Hash().String(), string(h.group_sign))
	assert.Equal(t, headerHash, h.Hash().String())
}
func mockHeader() *Header {
	key, _ := hex.DecodeString("ebe665c202f9393b85fe9bddbc31f39f7ad9a1eb14149a60f4ff23e806c111a6")
	pubKey, _ := crypto.MyCryptoLib.PrivateKeyToPubKey(key)
	asset_id, _, _ := String2AssetId("DEVIN")
	p := common.HexToHash("57c56162990aac482ae2b66196cd1f5129e6f026578470ab105042bf42d6a2dc")
	root := common.HexToHash("c35639062e40f8891cef2526b387f42e353b8f403b930106bb5aa3519e59e35f")

	b := []byte{}
	h := NewHeader([]common.Hash{p}, root, b, b, []byte("Extra"), []byte{0x1, 0x2}, []uint16{666},
		asset_id, 123, int64(123))
	sig, _ := crypto.MyCryptoLib.Sign(key, h.TxRoot().Bytes())
	au := Authentifier{}
	au.Signature = sig
	au.PubKey = pubKey
	h.SetAuthor(au)

	//h.SetGroupSign([]byte("group_sign"))
	//h.SetGroupPubkey([]byte("group_pubKey"))
	return h
}
