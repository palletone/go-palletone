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
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/palletone/go-palletone/common/ptndb"
	"github.com/palletone/go-palletone/dag/modules"
)

func MockStateMemDb() *StateDb {
	db, _ := ptndb.NewMemDatabase()
	statedb := NewStateDb(db)
	return statedb
}

func TestStateDb_GetPartitionChains(t *testing.T) {
	db := MockStateMemDb()
	partitions, err := db.GetPartitionChains()
	assert.Nil(t, err)
	assert.Equal(t, 0, len(partitions))
}

func TestSaveAndGetConfig(t *testing.T) {
	db := MockStateMemDb()
	version := &modules.StateVersion{Height: &modules.ChainIndex{Index: 123}, TxIndex: 1}
	err := db.SaveSysConfigContract("key1", nil, version)
	assert.Nil(t, err)
	data, version, err := db.getSysConfigContract("key1")
	assert.Nil(t, err)
	t.Log(data)
	// assert.Nil(t, data)
	assert.NotNil(t, version)

}

//func TestStateDb_UpdateSysParams(t *testing.T) {
//	// 初始化 环境
//	version := &modules.StateVersion{Height: &modules.ChainIndex{Index: 123}, TxIndex: 1}
//	db, _ := ptndb.NewMemDatabase()
//	statedb := NewStateDb(db)
//
//	// 初始化3个参数
//	err := statedb.SaveSysConfig("key1", []byte("lala1"), version)
//	if err != nil {
//		t.Error(err.Error())
//	}
//	err = statedb.SaveSysConfig("key2", []byte("lala2"), version)
//	if err != nil {
//		t.Error(err.Error())
//	}
//	err = statedb.SaveSysConfig(modules.DepositAmountForMediator, []byte("1000"), version)
//	if err != nil {
//		t.Error(err.Error())
//	}
//
//	// 1, 不通过投票修改参数
//	modifies := make(map[string]string)
//	modifies["key1"] = "val1"
//	modifies["key2"] = "val2"
//	modifiesByte, _ := json.Marshal(modifies)
//	err = statedb.SaveSysConfig(modules.DesiredSysParamsWithoutVote, modifiesByte, version)
//	if err != nil {
//		t.Error(err.Error())
//	}
//
//	// 2, 通过投票修改参数
//	sysTokenIDInfo := &modules.SysTokenIDInfo{}
//	sysSupportResult := &modules.SysSupportResult{}
//	sysTokenIDInfo.CreateTime = time.Now().UTC()
//	sysTokenIDInfo.AssetID = modules.DesiredSysParamsWithVote
//	sysTokenIDInfo.CreateAddr = "P1--------xxxxxxxxxxxxxxxxx"
//	sysTokenIDInfo.IsVoteEnd = true
//	sysTokenIDInfo.TotalSupply = 20
//	sysTokenIDInfo.LeastNum = 10
//	sysSupportResult.TopicIndex = 1
//	sysSupportResult.TopicTitle = modules.DepositAmountForMediator
//	sysVoteResult1 := &modules.SysVoteResult{}
//	sysVoteResult1.SelectOption = "2000"
//	sysVoteResult1.Num = 13
//	sysVoteResult2 := &modules.SysVoteResult{}
//	sysVoteResult2.SelectOption = "4000"
//	sysVoteResult2.Num = 7
//	sysSupportResult.VoteResults = []*modules.SysVoteResult{sysVoteResult1, sysVoteResult2}
//	sysTokenIDInfo.SupportResults = []*modules.SysSupportResult{sysSupportResult}
//	infoByte, _ := json.Marshal(sysTokenIDInfo)
//	err = statedb.SaveSysConfig(modules.DesiredSysParamsWithVote, infoByte, version)
//	if err != nil {
//		t.Error(err.Error())
//	}
//
//	t.Logf("\n============换届之前，还没有更改系统参数")
//	val1, _, err := statedb.GetSysConfig("key1")
//	if err != nil {
//		t.Error(err.Error())
//	}
//	t.Logf("key1=%s\n", val1)
//	val2, _, err := statedb.GetSysConfig("key2")
//	if err != nil {
//		t.Error(err.Error())
//	}
//	t.Logf("key2=%s\n", val1)
//	depositAmountForMediator, _, err := statedb.GetSysConfig(modules.DepositAmountForMediator)
//	if err != nil {
//		t.Error(err.Error())
//	}
//	t.Logf("DepositAmountForMediator=%s\n", depositAmountForMediator)
//
//	// 环境更新参数
//	err = statedb.UpdateSysParams(version)
//	if err != nil {
//		t.Error(err.Error())
//	}
//
//	t.Logf("\n============换届之后，已经更改系统参数")
//	val1, _, err = statedb.GetSysConfig("key1")
//	if err != nil {
//		t.Error(err.Error())
//	}
//	t.Logf("key1=%s\n", val1)
//	val2, _, err = statedb.GetSysConfig("key2")
//	if err != nil {
//		t.Error(err.Error())
//	}
//	t.Logf("key2=%s\n", val2)
//	depositAmountForMediator, _, err = statedb.GetSysConfig(modules.DepositAmountForMediator)
//	if err != nil {
//		t.Error(err.Error())
//	}
//	t.Logf("DepositAmountForMediator=%s\n", depositAmountForMediator)
//
//	// 检查是否重置为nil
//	sysParam, _, err := statedb.GetSysConfig(modules.DesiredSysParamsWithoutVote)
//	if err != nil {
//		t.Error(err.Error())
//	}
//	if sysParam == nil {
//		t.Log("update sysParam success")
//	} else if len(sysParam) > 0 {
//		t.Logf("%#v\n", sysParam)
//		t.Error("update sysParam fail")
//	} else {
//		t.Log("update sysParams success")
//	}
//	sysParams, _, err := statedb.GetSysConfig(modules.DesiredSysParamsWithVote)
//	if err != nil {
//		t.Error(err.Error())
//	}
//	if sysParams == nil {
//		t.Log("update sysParams success")
//	} else if len(sysParams) > 0 {
//		t.Logf("%#v\n", sysParams)
//		t.Error("update sysParams fail")
//	} else {
//		t.Log("update sysParams success")
//	}
//}
