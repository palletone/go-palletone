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
 *  * @author PalletOne core developers <dev@pallet.one>
 *  * @date 2018
 *
 */

package common

import (
	"encoding/json"
	"github.com/palletone/go-palletone/tokenengine"
	"testing"
	"time"

	"github.com/palletone/go-palletone/common/ptndb"
	"github.com/palletone/go-palletone/dag/modules"
)

func Test_UnitProduceRepository_UpdateSysParams(t *testing.T) {
	// 初始化 环境
	version := &modules.StateVersion{Height: &modules.ChainIndex{Index: 123}, TxIndex: 1}
	db, err := ptndb.NewMemDatabase()
	if err != nil {
		t.Error(err.Error())
	}
	upRep := NewUnitProduceRepository4Db(db, tokenengine.Instance)

	// 初始化若干个链参数
	gp := modules.NewGlobalProp()
	gp.ChainParameters.ActiveMediatorCount = 3
	gp.ChainParameters.FoundationAddress = "P1Kp2hcLhGEP45Xgx7vmSrE37QXunJUd8gJ"
	// gp.ChainParameters.TxCoinYearRate = 0.01
	err = upRep.propRep.StoreGlobalProp(gp)
	if err != nil {
		t.Error(err.Error())
	}

	// 1, 不通过投票修改参数
	modifies := make(map[string]string)
	// modifies["TxCoinYearRate"] = "0.02"
	modifies["FoundationAddress"] = "P16bXzewsexHwhGYdt1c1qbzjBirCqDg8mN"
	modifiesByte, _ := json.Marshal(modifies)
	err = upRep.stateRep.SaveSysConfigContract(modules.DesiredSysParamsWithoutVote, modifiesByte, version)
	if err != nil {
		t.Error(err.Error())
	}

	// 2, 通过投票修改参数
	sysTokenIDInfo := &modules.SysTokenIDInfo{}
	sysSupportResult := &modules.SysSupportResult{}
	sysTokenIDInfo.CreateTime = time.Now()
	sysTokenIDInfo.AssetID = modules.DesiredSysParamsWithVote
	sysTokenIDInfo.CreateAddr = "P1--------xxxxxxxxxxxxxxxxx"
	sysTokenIDInfo.IsVoteEnd = true
	sysTokenIDInfo.TotalSupply = 20
	sysTokenIDInfo.LeastNum = 10
	sysSupportResult.TopicIndex = 1
	sysSupportResult.TopicTitle = modules.DesiredActiveMediatorCount
	sysVoteResult1 := &modules.SysVoteResult{}
	sysVoteResult1.SelectOption = "5"
	sysVoteResult1.Num = 13
	sysVoteResult2 := &modules.SysVoteResult{}
	sysVoteResult2.SelectOption = "7"
	sysVoteResult2.Num = 7
	sysSupportResult.VoteResults = []*modules.SysVoteResult{sysVoteResult1, sysVoteResult2}
	sysTokenIDInfo.SupportResults = []*modules.SysSupportResult{sysSupportResult}
	infoByte, _ := json.Marshal(sysTokenIDInfo)
	err = upRep.stateRep.SaveSysConfigContract(modules.DesiredSysParamsWithVote, infoByte, version)
	if err != nil {
		t.Error(err.Error())
	}

	t.Logf("\n============换届之前，还没有更改系统参数")
	cp1 := upRep.propRep.GetChainParameters()
	if cp1 == nil {
		t.Error("cp1 is nil")
	}
	t.Logf("%v=%v\n", modules.DesiredActiveMediatorCount, cp1.ActiveMediatorCount)
	// t.Logf("%v=%v\n", "TxCoinYearRate", cp1.TxCoinYearRate)
	t.Logf("%v=%v\n", "FoundationAddress", cp1.FoundationAddress)

	// 换届更新参数
	err = upRep.UpdateSysParams(version)
	if err != nil {
		t.Error(err.Error())
	}

	t.Logf("\n============换届之后，已经更改系统参数")
	cp2 := upRep.propRep.GetChainParameters()
	if cp2 == nil {
		t.Error("cp2 is nil")
	}
	t.Logf("%v=%v\n", modules.DesiredActiveMediatorCount, cp2.ActiveMediatorCount)
	// t.Logf("%v=%v\n", "TxCoinYearRate", cp2.TxCoinYearRate)
	t.Logf("%v=%v\n", "FoundationAddress", cp2.FoundationAddress)

	// 检查是否重置为nil
	sysParam, err := upRep.stateRep.GetSysParamWithoutVote()
	if sysParam != nil {
		t.Error(err.Error())
	} else {
		t.Log("update sysParams success")
	}

	sysParams, err := upRep.stateRep.GetSysParamsWithVotes()
	if sysParams != nil {
		t.Error(err.Error())
	} else {
		t.Log("update sysParams success")
	}
}
