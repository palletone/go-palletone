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
	"github.com/palletone/go-palletone/common/uint128"
	"github.com/palletone/go-palletone/dag/modules"
)

func Test_UnitProduceRepository_UpdateSysParams(t *testing.T) {
	// 初始化 环境
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

	version := &modules.StateVersion{Height: &modules.ChainIndex{Index: 123}, TxIndex: 1}
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
	sysTokenIDInfo.CreateTime = time.Now().Unix()
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

func Test_updateAndSaveRecentSlotsFilled(t *testing.T) {
	dgp := modules.NewDynGlobalProp()
	t.Log(dgp.RecentSlotsFilled.BinaryStr())

	missedUnits := 10
	totalSlot := missedUnits + 1
	dgp.RecentSlotsFilled = dgp.RecentSlotsFilled.Lsh(uint(totalSlot)).Add64(1)
	t.Log(dgp.RecentSlotsFilled.BinaryStr())

	// 初始化 环境
	db, err := ptndb.NewMemDatabase()
	if err != nil {
		t.Error(err.Error())
	}
	upRep := NewUnitProduceRepository4Db(db, tokenengine.Instance)

	upRep.propRep.StoreDynGlobalProp(dgp)
	dgp1, err := upRep.propRep.RetrieveDynGlobalProp()
	if err != nil {
		t.Error(err.Error())
	}
	t.Log(dgp1.RecentSlotsFilled.BinaryStr())
}

func Test_LowAndHightDisplayOfRecentSlotsFilled(t *testing.T) {
	dgp := modules.NewDynGlobalProp()
	dgp.RecentSlotsFilled = uint128.New(1, 0)
	t.Log(dgp.RecentSlotsFilled.BinaryStr())

	dgp.RecentSlotsFilled = dgp.RecentSlotsFilled.Lsh(10).Add64(1)
	t.Log(dgp.RecentSlotsFilled.BinaryStr())

	dgp.RecentSlotsFilled = uint128.New(0, 1)
	t.Log(dgp.RecentSlotsFilled.BinaryStr())

	dgp.RecentSlotsFilled = dgp.RecentSlotsFilled.Lsh(10).Add64(1)
	t.Log(dgp.RecentSlotsFilled.BinaryStr())
}
