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
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/ptndb"
	"github.com/palletone/go-palletone/common/util"
	"github.com/palletone/go-palletone/contracts/syscontract"
	"github.com/palletone/go-palletone/dag/constants"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
)

func TestGetContractState(t *testing.T) {
	db, _ := ptndb.NewMemDatabase()
	statedb := NewStateDb(db)
	id := []byte("TestContract")
	contract := &modules.Contract{ContractId: id, Name: "TestContract1", TemplateId: []byte("Temp")}
	err := statedb.SaveContract(contract)
	assert.Nil(t, err, "save contract to statedb fail")
	version := &modules.StateVersion{Height: &modules.ChainIndex{Index: 123}, TxIndex: 1}
	ws := modules.NewWriteSet("name", []byte("TestName1"))
	err = statedb.SaveContractState(id, ws, version)
	assert.Nil(t, err, "Save contract state fail")
	value, version2, _ := statedb.GetContractState(id, "name")
	log.Debug("test debug: ", "version", version.String(), "value", string(value))
	assert.Equal(t, version, version2, "version not same.")
	log.Debug(fmt.Sprintf("get value from db:%s", value))
	assert.Equal(t, value, []byte("TestName1"), "value not same.")
	data, err := statedb.GetContractStatesById(id)
	assert.Nil(t, err)
	assert.True(t, len(data) > 0, "GetContractAllState don't return any data.")
	for key, v := range data {
		log.Debug(fmt.Sprintf("Key:%s,V:%s,version:%s", key, v.Value, v.Version))
	}
}

func TestStateDb_GetApprovedMediatorList(t *testing.T) {
	db, _ := ptndb.NewMemDatabase()
	statedb := NewStateDb(db)
	depositeContractAddress := common.HexToAddress("0x00000000000000000000000000000000000000011C")
	contractId := depositeContractAddress.Bytes()
	addr1 := "P1G988UGLytFgPwxy1bzY3FkzPT46ThDhTJ"

	addr2 := "P1FbTqEaSLNfhp1hCwNmRkj5BkMjTNU8jRp"

	list1 := make(map[string]bool)
	list1[addr1] = true
	list1[addr2] = true
	mediatorListBytes, err := json.Marshal(list1)
	assert.Nil(t, err, "json marshal error: ")
	version := &modules.StateVersion{Height: &modules.ChainIndex{Index: 123}, TxIndex: 1}
	ws := modules.NewWriteSet("MediatorList", mediatorListBytes)

	err = statedb.SaveContractState(contractId, ws, version)
	assert.Nil(t, err, "save mediatorlist error: ")
}

func TestGetContract(t *testing.T) {
	var contract modules.Contract

	contract.ContractId = []byte("123456")
	contract.TemplateId = []byte("Temp")
	contract.Name = "test"

	db, _ := ptndb.NewMemDatabase()
	statedb := NewStateDb(db)
	err := statedb.SaveContract(&contract)
	assert.Nil(t, err)
	dbContract, err := statedb.GetContract(contract.ContractId)
	assert.Nil(t, err)
	t.Logf("%#v", dbContract)
	assertRlpHashEqual(t, contract, dbContract)
}

func assertRlpHashEqual(t assert.TestingT, a, b interface{}) {
	hash1 := util.RlpHash(a)
	hash2 := util.RlpHash(b)
	assert.Equal(t, hash1, hash2)
}

func TestStateDb_GetSysParamWithoutVote(t *testing.T) {
	version := &modules.StateVersion{Height: &modules.ChainIndex{Index: 123}, TxIndex: 1}
	db, _ := ptndb.NewMemDatabase()
	statedb := NewStateDb(db)
	modifies := make(map[string]string)
	modifies["key1"] = "val1"
	modifies["key2"] = "val2"

	modifiesByte, _ := json.Marshal(modifies)
	err := statedb.SaveSysConfigContract(modules.DesiredSysParamsWithoutVote, modifiesByte, version)
	if err != nil {
		t.Error(err.Error())
	}
	modifies, err = statedb.GetSysParamWithoutVote()
	if err != nil {
		t.Error(err.Error())
	}
	t.Logf("----%#v\n", modifies)
}

func TestStateDb_GetSysParamsWithVotes(t *testing.T) {
	sysTokenIDInfo := &modules.SysTokenIDInfo{}
	sysSupportResult := &modules.SysSupportResult{}
	sysVoteResult := &modules.SysVoteResult{}
	sysTokenIDInfo.CreateTime = time.Now().UTC()
	sysTokenIDInfo.AssetID = modules.DesiredSysParamsWithVote
	sysTokenIDInfo.CreateAddr = "P1--------xxxxxxxxxxxxxxxxx"
	sysTokenIDInfo.IsVoteEnd = false
	sysTokenIDInfo.TotalSupply = 10
	sysSupportResult.TopicIndex = 1
	sysSupportResult.TopicTitle = "lalala"
	sysVoteResult.Num = 1
	sysVoteResult.SelectOption = "2"
	sysSupportResult.VoteResults = []*modules.SysVoteResult{sysVoteResult}
	sysTokenIDInfo.SupportResults = []*modules.SysSupportResult{sysSupportResult}
	db, _ := ptndb.NewMemDatabase()
	statedb := NewStateDb(db)
	infoByte, _ := json.Marshal(sysTokenIDInfo)
	version := &modules.StateVersion{Height: &modules.ChainIndex{Index: 123}, TxIndex: 1}
	err := statedb.SaveSysConfigContract(modules.DesiredSysParamsWithVote, infoByte, version)
	if err != nil {
		t.Error(err.Error())
	}
	sysTokenIDInfo, err = statedb.GetSysParamsWithVotes()
	if err != nil {
		t.Error(err.Error())
	}
	t.Logf("---%#v\n", sysTokenIDInfo)
	t.Logf("---%#v\n", sysTokenIDInfo.SupportResults[0])
	t.Logf("---%#v\n", sysTokenIDInfo.SupportResults[0].VoteResults[0])
}
func TestStateDb_DeleteContractState(t *testing.T) {
	// db, remove := newTestLDB()
	// defer remove()
	db, _ := ptndb.NewMemDatabase()
	statedb := NewStateDb(db)
	depositeContractAddress := common.HexToAddress("0x00000000000000000000000000000000000000011C")
	contractId := depositeContractAddress.Bytes()
	version := &modules.StateVersion{Height: &modules.ChainIndex{Index: 123}, TxIndex: 1}
	ws1 := modules.NewWriteSet("AA", []byte("100"))
	ws2 := modules.NewWriteSet("AB", []byte("1000"))
	ws := []modules.ContractWriteSet{}
	ws = append(ws, *ws1)
	ws = append(ws, *ws2)
	err := statedb.SaveContractStates(contractId, ws, version)
	assert.Nil(t, err)
	result, err := statedb.GetContractStatesByPrefix(contractId, "A")
	assert.Nil(t, err)
	assert.Equal(t, 2, len(result))
	// key := getContractStateKey(contractId, "AA")
	// statedb.DeleteState(key)
	ws3 := modules.ContractWriteSet{true, "AA", nil, nil}
	ws = []modules.ContractWriteSet{}
	ws = append(ws, ws3)
	err = statedb.SaveContractStates(contractId, ws, version)
	assert.Nil(t, err)
	result, err = statedb.GetContractStatesByPrefix(contractId, "A")
	assert.Nil(t, err)
	assert.Equal(t, 1, len(result))
}
func newTestLDB() (*ptndb.LDBDatabase, func()) {
	dirname, err := ioutil.TempDir(os.TempDir(), "ptndb_test_")
	if err != nil {
		panic("failed to create test file: " + err.Error())
	}
	db, err := ptndb.NewLDBDatabase(dirname, 0, 0)
	if err != nil {
		panic("failed to create test database: " + err.Error())
	}

	return db, func() {
		db.Close()
		os.RemoveAll(dirname)
	}
}

func TestJurors(t *testing.T) {
	db, _ := ptndb.NewMemDatabase()
	statedb := NewStateDb(db)
	depositeContractAddress := syscontract.DepositContractAddress
	contractId := depositeContractAddress.Bytes()
	version := &modules.StateVersion{Height: &modules.ChainIndex{Index: 123}, TxIndex: 1}
	j1 := &modules.Juror{}
	j1.Address = "p1"
	b1, _ := json.Marshal(j1)
	j2 := &modules.Juror{}
	j2.Address = "p2"
	b2, _ := json.Marshal(j2)
	ws1 := modules.NewWriteSet(string(constants.DEPOSIT_JURY_BALANCE_PREFIX)+"p1", b1)
	ws2 := modules.NewWriteSet(string(constants.DEPOSIT_JURY_BALANCE_PREFIX)+"p2", b2)
	ws := []modules.ContractWriteSet{}
	ws = append(ws, *ws1)
	ws = append(ws, *ws2)
	err := statedb.SaveContractStates(contractId, ws, version)
	assert.Nil(t, err)
	juror, err := statedb.GetAllJuror()
	assert.Nil(t, err)
	assert.Equal(t, 2, len(juror))
	j, err := statedb.GetJurorByAddr("p1")
	assert.Nil(t, err)
	assert.NotNil(t, j)
}
