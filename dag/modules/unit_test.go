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
	"log"
	"testing"
	"github.com/palletone/go-palletone/common"
)

func TestNewUnit(t *testing.T) {
	txs := make(Transactions, 0)
	unit := NewUnit(&Header{}, txs)
	log.Println("unit", unit)
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
	author := Author{
		Address: addr,
		Pubkey: []byte("12345678901234567890"),
		TxAuthentifier: Authentifier{R:"jsjjsjlsllls"},
	}
	w := []Author{}
	w = append(w, author)
	assetID := IDType16{}
	assetID.SetBytes([]byte("0000000011111111"))
	h := Header{
		ParentUnits: []common.Hash{u1, u2},
		AssetIDs: []IDType16{assetID},
		Authors: &author,
		Witness: w,
		GasLimit: 1,
		GasUsed: 1,
		Root: common.Hash{},
		Number: ChainIndex{AssetID:assetID, IsMain:true, Index:0},
	}

	newH := CopyHeader(&h)
	newH.Authors = nil
	newH.Witness = []Author{}
	hh := Header{}
	log.Printf("newh=%v \n oldH=%v \n hh=%v", *newH, h, hh)
}