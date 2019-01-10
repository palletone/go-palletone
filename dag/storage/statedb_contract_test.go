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

import (
	"fmt"
	"testing"

	//"github.com/palletone/go-palletone/common/crypto"
	"encoding/json"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/ptndb"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/stretchr/testify/assert"
)

func TestGetContractState(t *testing.T) {
	db, _ := ptndb.NewMemDatabase()
	//l := log.NewTestLog()
	statedb := NewStateDb(db)
	id := []byte("TestContract")
	contract := &modules.Contract{Id: id, Name: "TestContract1", Code: []byte("code")}
	err := statedb.SaveContract(contract)
	assert.Nil(t, err, "save contract to statedb fail")
	version := &modules.StateVersion{Height: modules.ChainIndex{Index: 123, IsMain: true}, TxIndex: 1}
	err = statedb.SaveContractState(id, "name", "TestName1", version)
	assert.Nil(t, err, "Save contract state fail")
	version2, value := statedb.GetContractState(id, "name")
	log.Debug("test debug: ", "version", version.String())
	assert.Equal(t, version, version2, "version not same.")
	log.Debug(fmt.Sprintf("get value from db:%s", value))
	assert.Equal(t, value, []byte("TestName1"), "value not same.")
	data, _ := statedb.GetContractStatesById(id)
	assert.True(t, len(data) > 0, "GetContractAllState don't return any data.")
	for key, v := range data {
		log.Debug(fmt.Sprintf("Key:%s,V:%s,version:%s", key, v.Value, v.Version))
	}
}

func TestStateDb_GetApprovedMediatorList(t *testing.T) {
	db, _ := ptndb.NewMemDatabase()
	//l := log.NewTestLog()
	statedb := NewStateDb(db)
	depositeContractAddress := common.HexToAddress("0x00000000000000000000000000000000000000011C")
	contractId := depositeContractAddress.Bytes()
	//fmt.Println(contractId)
	addr1 := "P1G988UGLytFgPwxy1bzY3FkzPT46ThDhTJ"
	mediator1 := &modules.MediatorRegisterInfo{
		Address: addr1,
	}
	//assert.Nil(t, err, "string 2 address fail: ")
	addr2 := "P1FbTqEaSLNfhp1hCwNmRkj5BkMjTNU8jRp"
	mediator2 := &modules.MediatorRegisterInfo{
		Address: addr2,
	}
	//assert.Nil(t, err, "string 2 address fail: ")
	mediatorList := []*modules.MediatorRegisterInfo{mediator1, mediator2}
	mediatorListBytes, err := json.Marshal(mediatorList)
	assert.Nil(t, err, "json marshal error: ")
	version := &modules.StateVersion{Height: modules.ChainIndex{Index: 123, IsMain: true}, TxIndex: 1}
	err = statedb.SaveContractState(contractId, "MediatorList", mediatorListBytes, version)
	assert.Nil(t, err, "save mediatorlist error: ")
	list, err := statedb.GetApprovedMediatorList()
	assert.Nil(t, err, "get mediator candidate list error: ")
	assert.True(t, len(list) == 2, "len is erroe")
	for _, mediatorAddr := range list {
		fmt.Println(mediatorAddr)
	}
}

//func TestGetContract(t *testing.T) {
//	var keys []string
//	var results []interface{}
//	var origin modules.Contract
//
//	origin.Id = common.HexToHash("123456")
//
//	origin.Name = "test"
//	origin.Code = []byte(`logger.PrintLn("hello world")`)
//	origin.Input = []byte("input")
//
//	db, _ := ptndb.NewMemDatabase()
//
//	log.Debug("store error: ", StoreBytes(db, append(CONTRACT_PREFIX, origin.Id[:]...), origin))
//	keys = append(keys, "Id", "id", "Name", "Code", "code", "codes", "inputs")
//	results = append(results, common.HexToHash("123456"), nil, "test", []byte(`logger.PrintLn("hello world")`), nil, nil, nil)
//	log.Debug("test data: ", keys)
//
//	for i, k := range keys {
//		data, err := GetContractKeyValue(db, origin.Id, k)
//		if !reflect.DeepEqual(data, results[i]) {
//			t.Error("test error:", err, "the expect key is:", k, " value is :", results[i], ",but the return value is: ", data)
//		}
//	}
//}
