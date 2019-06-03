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
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/stretchr/testify/assert"
)

func TestGetContractState(t *testing.T) {
	db, _ := ptndb.NewMemDatabase()
	//l := log.NewTestLog()
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
	//l := log.NewTestLog()
	statedb := NewStateDb(db)
	depositeContractAddress := common.HexToAddress("0x00000000000000000000000000000000000000011C")
	contractId := depositeContractAddress.Bytes()
	//fmt.Println(contractId)
	addr1 := "P1G988UGLytFgPwxy1bzY3FkzPT46ThDhTJ"
	//mediator1 := &core.MediatorApplyInfo{
	//	Address: addr1,
	//}
	//assert.Nil(t, err, "string 2 address fail: ")
	addr2 := "P1FbTqEaSLNfhp1hCwNmRkj5BkMjTNU8jRp"
	//mediator2 := &core.MediatorApplyInfo{
	//	Address: addr2,
	//}
	//assert.Nil(t, err, "string 2 address fail: ")
	//mediatorList := []*core.MediatorApplyInfo{mediator1, mediator2}
	list1 := make(map[string]bool)
	list1[addr1] = true
	list1[addr2] = true
	mediatorListBytes, err := json.Marshal(list1)
	assert.Nil(t, err, "json marshal error: ")
	version := &modules.StateVersion{Height: &modules.ChainIndex{Index: 123}, TxIndex: 1}
	ws := modules.NewWriteSet("MediatorList", mediatorListBytes)

	err = statedb.SaveContractState(contractId, ws, version)
	assert.Nil(t, err, "save mediatorlist error: ")
	//list2, err := statedb.GetApprovedMediatorList()
	//assert.Nil(t, err, "get mediator candidate list error: ")
	//assert.True(t, len(list2) == 2, "len is erroe")
	//for k, b := range list2 {
	//	fmt.Println(k, b)
	//}
}

func TestGetContract(t *testing.T) {
	//var keys []string
	//var results []interface{}
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
	//log.Debug("store error: ", StoreToRlpBytes(db, append(CONTRACT_PREFIX, contract.Id[:]...), contract))
	//keys = append(keys, "Id", "id", "Name", "Code", "code", "codes", "inputs")
	//results = append(results, common.HexToHash("123456"), nil, "test", []byte(`logger.PrintLn("hello world")`), nil, nil, nil)
	//log.Debug("test data: ", keys)
	//
	//for i, k := range keys {
	//	data, err := GetContractKeyValue(db, contract.Id, k)
	//	if !reflect.DeepEqual(data, results[i]) {
	//		t.Error("test error:", err, "the expect key is:", k, " value is :", results[i], ",but the return value is: ", data)
	//	}
	//}
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
	//[{\"Key\":\"depositAmountForJury\",\"Value\":\"9000000\"}]
	//err := statedb.SaveContractState(syscontract.SysConfigContractAddress.Bytes21(), modules.DesiredSysParams, modifiesByte, version)
	err := statedb.SaveSysConfig(modules.DesiredSysParamsWithoutVote, modifiesByte, version)
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
	err := statedb.SaveSysConfig(modules.DesiredSysParamsWithVote, infoByte, version)
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

func TestStateDb_UpdateSysParams(t *testing.T) {
	// 初始化 环境
	version := &modules.StateVersion{Height: &modules.ChainIndex{Index: 123}, TxIndex: 1}
	db, _ := ptndb.NewMemDatabase()
	statedb := NewStateDb(db)

	// 初始化3个参数
	err := statedb.SaveSysConfig("key1", []byte("lala1"), version)
	if err != nil {
		t.Error(err.Error())
	}
	err = statedb.SaveSysConfig("key2", []byte("lala2"), version)
	if err != nil {
		t.Error(err.Error())
	}
	err = statedb.SaveSysConfig(modules.DepositAmountForMediator, []byte("1000"), version)
	if err != nil {
		t.Error(err.Error())
	}

	// 1, 不通过投票修改参数
	modifies := make(map[string]string)
	modifies["key1"] = "val1"
	modifies["key2"] = "val2"
	modifiesByte, _ := json.Marshal(modifies)
	err = statedb.SaveSysConfig(modules.DesiredSysParamsWithoutVote, modifiesByte, version)
	if err != nil {
		t.Error(err.Error())
	}

	// 2, 通过投票修改参数
	sysTokenIDInfo := &modules.SysTokenIDInfo{}
	sysSupportResult := &modules.SysSupportResult{}
	sysTokenIDInfo.CreateTime = time.Now().UTC()
	sysTokenIDInfo.AssetID = modules.DesiredSysParamsWithVote
	sysTokenIDInfo.CreateAddr = "P1--------xxxxxxxxxxxxxxxxx"
	sysTokenIDInfo.IsVoteEnd = true
	sysTokenIDInfo.TotalSupply = 20
	sysTokenIDInfo.LeastNum = 10
	sysSupportResult.TopicIndex = 1
	sysSupportResult.TopicTitle = modules.DepositAmountForMediator
	sysVoteResult1 := &modules.SysVoteResult{}
	sysVoteResult1.SelectOption = "2000"
	sysVoteResult1.Num = 13
	sysVoteResult2 := &modules.SysVoteResult{}
	sysVoteResult2.SelectOption = "4000"
	sysVoteResult2.Num = 7
	sysSupportResult.VoteResults = []*modules.SysVoteResult{sysVoteResult1, sysVoteResult2}
	sysTokenIDInfo.SupportResults = []*modules.SysSupportResult{sysSupportResult}
	infoByte, _ := json.Marshal(sysTokenIDInfo)
	err = statedb.SaveSysConfig(modules.DesiredSysParamsWithVote, infoByte, version)
	if err != nil {
		t.Error(err.Error())
	}

	t.Logf("\n============换届之前，还没有更改系统参数")
	val1, _, err := statedb.GetSysConfig("key1")
	if err != nil {
		t.Error(err.Error())
	}
	t.Logf("key1=%s\n", val1)
	val2, _, err := statedb.GetSysConfig("key2")
	if err != nil {
		t.Error(err.Error())
	}
	t.Logf("key2=%s\n", val1)
	depositAmountForMediator, _, err := statedb.GetSysConfig(modules.DepositAmountForMediator)
	if err != nil {
		t.Error(err.Error())
	}
	t.Logf("DepositAmountForMediator=%s\n", depositAmountForMediator)

	// 环境更新参数
	err = statedb.UpdateSysParams(version)
	if err != nil {
		t.Error(err.Error())
	}

	t.Logf("\n============换届之后，已经更改系统参数")
	val1, _, err = statedb.GetSysConfig("key1")
	if err != nil {
		t.Error(err.Error())
	}
	t.Logf("key1=%s\n", val1)
	val2, _, err = statedb.GetSysConfig("key2")
	if err != nil {
		t.Error(err.Error())
	}
	t.Logf("key2=%s\n", val2)
	depositAmountForMediator, _, err = statedb.GetSysConfig(modules.DepositAmountForMediator)
	if err != nil {
		t.Error(err.Error())
	}
	t.Logf("DepositAmountForMediator=%s\n", depositAmountForMediator)

	// 检查是否重置为nil
	sysParam, _, err := statedb.GetSysConfig(modules.DesiredSysParamsWithoutVote)
	if err != nil {
		t.Error(err.Error())
	}
	if sysParam == nil {
		t.Log("update sysParam success")
	} else if len(sysParam) > 0 {
		t.Logf("%#v\n", sysParam)
		t.Error("update sysParam fail")
	} else {
		t.Log("update sysParams success")
	}
	sysParams, _, err := statedb.GetSysConfig(modules.DesiredSysParamsWithVote)
	if err != nil {
		t.Error(err.Error())
	}
	if sysParams == nil {
		t.Log("update sysParams success")
	} else if len(sysParams) > 0 {
		t.Logf("%#v\n", sysParams)
		t.Error("update sysParams fail")
	} else {
		t.Log("update sysParams success")
	}
}
