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
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/ptndb"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/stretchr/testify/assert"
)

func TestGetContractState(t *testing.T) {
	db, _ := ptndb.NewMemDatabase()
	l := log.NewTestLog()
	statedb := NewStateDb(db, l)
	id := []byte("TestContract")
	contract := &modules.Contract{Id: id, Name: "TestContract1", Code: []byte("code")}
	err := statedb.SaveContract(contract)
	assert.Nil(t, err, "save contract to statedb fail")
	version := &modules.StateVersion{Height: modules.ChainIndex{Index: 123, IsMain: true}, TxIndex: 1}
	err = statedb.SaveContractState(id, "name", "TestName", version)
	assert.Nil(t, err, "Save contract state fail")
	version, value := statedb.GetContractState(id, "name")
	log.Debug("test debug: ", "version", version.String())
	log.Debug(fmt.Sprintf("value:%#x", value))
	assert.Equal(t, value, []byte("TestName"), "value not same.")
	data := statedb.GetContractAllState()
	for _, v := range data {
		log.Debug(fmt.Sprintf("K:%s,V:%x,version:%s", v.Key, v.Value, v.Version))
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
