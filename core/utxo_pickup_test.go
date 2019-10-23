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

package core

import (
        "fmt"
	"github.com/palletone/go-palletone/common"
	"github.com/stretchr/testify/assert"
	"testing"
)

type Utxo4Test struct {
	Amount uint64
	TxId   common.Hash
	MsgIdx uint32
	OutIdx uint32
}

func (u *Utxo4Test) GetAmount() uint64 {
	return u.Amount
}
func TestSelect_utxo_Greedy(t *testing.T) {
	//log.NewTestLog()

	utxos := []*Utxo4Test{}
	utxos = append(utxos, &Utxo4Test{Amount: 3, TxId: common.Hash{}, MsgIdx: 0, OutIdx: 2})
	utxos = append(utxos, &Utxo4Test{Amount: 1, TxId: common.Hash{}, MsgIdx: 0, OutIdx: 1})
	utxos = append(utxos, &Utxo4Test{Amount: 2, TxId: common.Hash{}, MsgIdx: 0, OutIdx: 3})
	utxos = append(utxos, &Utxo4Test{Amount: 5, TxId: common.Hash{}, MsgIdx: 0, OutIdx: 4})
	ut := Utxos{}
	for _, u := range utxos {
		ut = append(ut, u)
	}
	result, change, err := Select_utxo_Greedy(ut, 4)
	assert.Nil(t, err)
	assert.Equal(t, len(result), 3)
	assert.Equal(t, change, uint64(2))
	result, change, err = Select_utxo_Greedy(ut, 6)
	assert.Nil(t, err)
	for _, u := range result {
		t.Logf("Selected: %+v\n", u)
	}
	assert.Equal(t, len(result), 3)
	assert.Equal(t, change, uint64(0))
	result, change, err = Select_utxo_Greedy(ut, 12)
	assert.NotNil(t, err)
	t.Logf("get error:%s", err)

}
func TestNew_Selectutxo_Greedy(t *testing.T) {
	//log.NewTestLog()

	utxos := []*Utxo4Test{}
	for i := 0;i < 600; i++ {
             utxos = append(utxos, &Utxo4Test{Amount: uint64(i+1), TxId: common.Hash{}, MsgIdx: 0, OutIdx: uint32(i)})
	}
	
	ut := Utxos{}
	for _, u := range utxos {
		ut = append(ut, u)
	}
	result, change, err := New_SelectUtxo_Greedy(ut, 400)
        //fmt.Println("result is ",result)
        fmt.Println("change is ",change)
	assert.Nil(t, err)
	assert.Equal(t, len(result), 500)
	assert.Equal(t, change, uint64(174850))

        result, change, err = New_SelectUtxo_Greedy(ut, 200000)
        //fmt.Println("result is ",result)
        fmt.Println("change is ",change)
        assert.NotNil(t, err)
        assert.Equal(t, len(result), 0)
        assert.Equal(t, change, uint64(0))

        result, change, err = New_SelectUtxo_Greedy(ut, 174850)
        //fmt.Println("result is ",result)
        fmt.Println("change is ",change)
        assert.Nil(t, err)
        assert.Equal(t, len(result), 500)
        assert.Equal(t, change, uint64(400))
     
        result, change, err = New_SelectUtxo_Greedy(ut, 600)
        //fmt.Println("result is ",result)
        fmt.Println("change is ",change)
        assert.Nil(t, err)
        assert.Equal(t, len(result), 500)
        assert.Equal(t, change, uint64(174650))
}
