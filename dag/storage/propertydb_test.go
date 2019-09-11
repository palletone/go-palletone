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
 *  * @date 2018-2019
 *
 */

package storage

import (
	"fmt"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/ptndb"
	"github.com/palletone/go-palletone/contracts/list"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPropertyDb_StoreMediatorSchl(t *testing.T) {
	addr1, _ := common.StringToAddress("P1FeeyyaQzetqLfMb2jk3YrmJujwa3FHwke")
	addr2, _ := common.StringToAddress("P151GBRxoZoqqcGeFoaf66R1hfs8WKc3Wdo")
	addr3, _ := common.StringToAddress("P1NnBhh78xhShyrQcD8tKZFq5mkQV3U6uWr")
	ms := &modules.MediatorSchedule{CurrentShuffledMediators: []common.Address{addr1, addr2, addr3}}
	db, _ := ptndb.NewMemDatabase()
	pdb := NewPropertyDb(db)
	err := pdb.StoreMediatorSchl(ms)
	assert.Nil(t, err)
	dbms, err := pdb.RetrieveMediatorSchl()
	assert.Nil(t, err)
	assert.Equal(t, 3, len(dbms.CurrentShuffledMediators))
	t.Log(dbms.String())
}

func TestPropertyDb_RetrieveChaincodes(t *testing.T) {
	db, _ := ptndb.NewMemDatabase()
	pdb := NewPropertyDb(db)
	addr1, _ := common.StringToAddress("P1FeeyyaQzetqLfMb2jk3YrmJujwa3FHwke")
	addr2, _ := common.StringToAddress("P151GBRxoZoqqcGeFoaf66R1hfs8WKc3Wdo")
	cc1 := &list.CCInfo{Id: addr1.Bytes(), TempleId: []byte("addr1"), Name: "addr1"}
	cc2 := &list.CCInfo{Id: addr2.Bytes(), TempleId: []byte("addr2"), Name: "addr2"}
	pdb.SaveChaincode(addr1, cc1)
	pdb.SaveChaincode(addr2, cc2)
	cc11, _ := pdb.GetChaincode(addr1)
	fmt.Printf("cc1 %#v\n", cc11)
	cc22, _ := pdb.GetChaincode(addr2)
	fmt.Printf("cc2 %#v\n", cc22)
	cc1122, _ := pdb.RetrieveChaincodes()
	for _, c := range cc1122 {
		fmt.Printf("========%#v\n", c)
	}
	assert.Equal(t, 2, len(cc1122))
}
