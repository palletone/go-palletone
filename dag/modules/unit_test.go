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
	"crypto/ecdsa"
	"log"
	"testing"
	"time"
	"unsafe"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/crypto"
)

func TestNewUnit(t *testing.T) {
	txs := make(Transactions, 0)
	unit := NewUnit(&Header{Extra: []byte("hello"), Creationdate: time.Now().Unix()}, txs)
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
		Address: addr,
		R:       []byte("12345678901234567890"),
		S:       []byte("09876543210987654321"),
		V:       []byte("1"),
	}
	//author := Author{
	//	Address:        addr,
	//	Pubkey:         []byte("1234567890123456789"),
	//	TxAuthentifier: auth,
	//}
	w := []*Authentifier{}
	w = append(w, &auth)
	assetID := IDType16{}
	assetID.SetBytes([]byte("0000000011111111"))
	h := Header{
		ParentsHash: []common.Hash{u1, u2},
		AssetIDs:    []IDType16{assetID},
		Authors:     &auth,
		Witness:     w,
		TxRoot:      common.Hash{},
		Number:      ChainIndex{AssetID: assetID, IsMain: true, Index: 0},
	}

	newH := CopyHeader(&h)
	newH.Authors = nil
	newH.Witness = []*Authentifier{}
	hh := Header{}
	log.Printf("newh=%v \n oldH=%v \n hh=%v", *newH, h, hh)
}

// test unit's size of header
func TestUnitSize(t *testing.T) {
	key := new(ecdsa.PrivateKey)
	key, _ = crypto.GenerateKey()
	h := new(Header)
	h.AssetIDs = append(h.AssetIDs, PTNCOIN)
	au := new(Authentifier)
	address := crypto.PubkeyToAddress(&key.PublicKey)
	log.Println("address:", address)

	//author := &Author{
	//	Address:        address,
	//	Pubkey:         []byte("1234567890123456789"),
	//	TxAuthentifier: *au,
	//}

	h.Witness = append(h.Witness, au)
	h.Number.AssetID = PTNCOIN
	h.Number.Index = uint64(333333)
	h.Extra = make([]byte, 20)
	h.ParentsHash = append(h.ParentsHash, h.TxRoot)

	h.TxRoot = h.Hash()
	sig, _ := crypto.Sign(h.TxRoot[:], key)
	au.R = sig[:32]
	au.S = sig[32:64]
	au.V = sig[64:]
	h.Authors = au

	log.Println("size: ", unsafe.Sizeof(h))

}
