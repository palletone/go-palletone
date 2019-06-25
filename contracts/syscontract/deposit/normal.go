/*
	This file is part of go-palletone.
	go-palletone is free software: you can redistribute it and/or modify
	it under the terms of the GNU General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.
	go-palletone is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU General Public License for more details.
	You should have received a copy of the GNU General Public License
	along with go-palletone.  If not, see <http://www.gnu.org/licenses/>.
*/

package deposit

import (
	"github.com/palletone/go-palletone/contracts/shim"
	pb "github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
	"github.com/palletone/go-palletone/dag/modules"
	"strconv"
	"time"
)

//  质押PTN投票mediator
func normalNodePledgeVote(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	//  获取是否是保证金合约
	invokeToken, err := isContainDepositContractAddr(stub)
	if err != nil {
		return shim.Error(err.Error())
	}
	if len(args) != 1 {
		return shim.Error("need 1 arg")
	}
	//  获取请求地址
	inAddr, err := stub.GetInvokeAddress()
	if err != nil {
		return shim.Error(err.Error())
	}
	//  获取是否存在
	nor, err := getNor(stub, inAddr.String())
	if err != nil {
		return shim.Error(err.Error())
	}
	if nor == nil {
		nor = &NorNodBal{}
		nor.AmountAsset = &modules.AmountAsset{}
	}
	//allVote, err := getPledgeVotes(stub)
	//if err != nil {
	//	return shim.Error(err.Error())
	//}
	//allVote += int64(invokeToken.Amount)
	//err = savePledgeVotes(stub, allVote)
	//if err != nil {
	//	return shim.Error(err.Error())
	//}
	nor.AmountAsset.Amount += invokeToken.Amount
	nor.AmountAsset.Asset = invokeToken.Asset
	mediatorAddr := args[0]
	nor.MediatorAddr = mediatorAddr
	//  保存
	err = saveNor(stub, inAddr.String(), nor)
	if err != nil {
		return shim.Error(err.Error())
	}
	//  获取质押PTN列表
	//norMap, err := getNorMap(stub)
	//if err != nil {
	//	return shim.Error(err.Error())
	//}
	//if norMap == nil {
	//	norMap = make(map[string]*modules.AmountAsset)
	//}
	//var norB *modules.AmountAsset
	////  判断是否已经存在
	//if _, ok := norMap[inAddr.String()]; ok {
	//	norB = norMap[inAddr.String()]
	//}
	//norB.Amount += invokeToken.Amount
	//norB.Asset = invokeToken.Asset
	//err = saveNorMap(stub, norMap)
	//if err != nil {
	//	return shim.Error(err.Error())
	//}
	return shim.Success(nil)
}

func normalNodeChangeVote(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("need 1 arg")
	}
	//  获取请求地址
	inAddr, err := stub.GetInvokeAddress()
	if err != nil {
		return shim.Error(err.Error())
	}
	//  获取是否存在
	nor, err := getNor(stub, inAddr.String())
	if err != nil {
		return shim.Error(err.Error())
	}
	if nor == nil {
		return shim.Error("node was nil")
	}
	mediatorAddr := args[0]
	nor.MediatorAddr = mediatorAddr
	//  保存
	err = saveNor(stub, inAddr.String(), nor)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(nil)
}

func normalNodeExtractVote(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("need 1 arg")
	}
	//  获取请求地址
	inAddr, err := stub.GetInvokeAddress()
	if err != nil {
		return shim.Error(err.Error())
	}
	//  获取是否存在
	nor, err := getNor(stub, inAddr.String())
	if err != nil {
		return shim.Error(err.Error())
	}
	if nor == nil {
		return shim.Error("node was nil")
	}
	amount := args[0]
	ptnAccount, err := strconv.ParseUint(amount, 10, 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	if ptnAccount > nor.AmountAsset.Amount {
		return shim.Error("PTN was not enough")
	}
	//  保存质押提取
	extPtnLis, err := getExtPtn(stub)
	if err != nil {
		return shim.Error(err.Error())
	}
	if extPtnLis == nil {
		extPtnLis = make(map[string]*extractPtn)
	}
	var extPtn *extractPtn
	if _, ok := extPtnLis[inAddr.String()]; ok {
		extPtn = extPtnLis[inAddr.String()]
	}
	fees, err := stub.GetInvokeFees()
	if err != nil {
		return shim.Error(err.Error())
	}
	extPtn.Amount.Amount += ptnAccount
	extPtn.Amount.Asset = fees.Asset
	extPtn.Time = TimeStr()
	//  保存
	err = saveExtPtn(stub, extPtnLis)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(nil)
}

func handleExtractVote(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 0 {
		return shim.Error("need 0 args")
	}
	//  判断是否是基金会
	if !isFoundationInvoke(stub) {
		return shim.Error("please use foundation address")
	}
	//  保存质押提取
	extPtnLis, err := getExtPtn(stub)
	if err != nil {
		return shim.Error(err.Error())
	}
	if extPtnLis == nil {
		return shim.Error("list is nil")
	}
	cp, err := stub.GetSystemConfig()
	if err != nil {
		return shim.Error(err.Error())
	}
	day := cp.DepositPeriod
	for k, v := range extPtnLis {
		tim := StrToTime(v.Time)
		dur := int(time.Since(tim).Hours())
		if dur/24 < day {
			continue
		}
		err := stub.PayOutToken(k, v.Amount, 0)
		if err != nil {
			return shim.Error(err.Error())
		}
	}
	return shim.Success(nil)
}

func handleEachDayAward(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 0 {
		return shim.Error("need 0 args")
	}
	//  判断是否是基金会
	if !isFoundationInvoke(stub) {
		return shim.Error("please use foundation address")
	}
	norMap, err := getNorMap(stub)
	if err != nil {
		return shim.Error(err.Error())
	}
	if norMap == nil {
		return shim.Error("list is nil")
	}
	cp, err := stub.GetSystemConfig()
	if err != nil {
		return shim.Error(err.Error())
	}
	depositExtraReward := cp.DepositExtraReward
	pledgeVotes, err := getPledgeVotes(stub)
	if err != nil {
		return shim.Error(err.Error())
	}
	rate := depositExtraReward / pledgeVotes
	for k, v := range norMap {
		dayAward := uint64(rate) * v.Amount
		nor, _ := getNor(stub, k)
		nor.AmountAsset.Amount += dayAward
		_ = saveNor(stub, k, nor)
	}
	return shim.Success(nil)
}
