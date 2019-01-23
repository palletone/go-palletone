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

package storage

//func TestStateDb_GetSortedMediatorVote(t *testing.T) {
//	db, _ := ptndb.NewMemDatabase()
//	log := log.NewTestLog()
//	statedb := NewStateDb(db, log)
//	addr1 := common.NewAddress([]byte{11}, common.PublicKeyHash)
//	addr2 := common.NewAddress([]byte{22}, common.PublicKeyHash)
//	addr3 := common.NewAddress([]byte{33}, common.PublicKeyHash)
//	statedb.SaveAccountInfo(addr1, &modules.AccountInfo{PtnBalance: 100, MediatorVote: common.StringToAddress("P13tZsVm4pMgssJc4Zh9h46Wb4ZhqqXfLgZ")})
//	statedb.SaveAccountInfo(addr2, &modules.AccountInfo{PtnBalance: 200, MediatorVote: []byte("[\"P1NqAPUGdYnpnf51tgNbhanXMgg5uF125ex\"]")})
//	statedb.SaveAccountInfo(addr3, &modules.AccountInfo{PtnBalance: 300, MediatorVote: []byte("[\"P13tZsVm4pMgssJc4Zh9h46Wb4ZhqqXfLgZ\"]")})
//	result, err := statedb.GetSortedMediatorVote(0)
//	assert.Nil(t, err)
//	assert.True(t, len(result) > 0)
//	for addr, voteCount := range result {
//		t.Logf("Addr:%s,Count:%d", addr, voteCount)
//	}
//}
