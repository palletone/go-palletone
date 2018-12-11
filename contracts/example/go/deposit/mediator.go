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
/*
 * Copyright IBM Corp. All Rights Reserved.
 * @author PalletOne core developers <dev@pallet.one>
 * @date 2018
 */

//Package deposit implements some functions for deposit contract.
package deposit

import (
	"encoding/json"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/award"
	"github.com/palletone/go-palletone/contracts/shim"
	"github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
	"github.com/palletone/go-palletone/dag/modules"
	"strings"
	"time"
)

//申请加入  参数：暂时  姓名和节点地址
func ApplyBecomeMediator(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 2 {
		return shim.Error("arg need two parameter.")
	}
	name := args[0]
	addr := args[1]
	mediatorInfo := modules.MediatorInfo{
		Name:    name,
		Address: addr,
	}
	//获取列表
	list, err := stub.GetBecomeMediatorApplyList()
	if err != nil {
		return shim.Error(err.Error())
	}
	if list == nil {
		list = new(modules.BecomeMediatorApplyList)
		list.MediatorList = append(list.MediatorList, &mediatorInfo)
	} else {
		list.MediatorList = append(list.MediatorList, &mediatorInfo)
	}
	listByte, err := json.Marshal(list)
	if err != nil {
		return shim.Error(err.Error())
	}
	stub.PutState("ListForApplyBecomeMediator", listByte)
	return shim.Success([]byte("ok"))
}

//处理加入 参数：同意或不同意，节点的地址
func HandleForApplyBecomeMediator(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 2 {
		return shim.Error("arg need two parameter.")
	}
	//获取申请列表
	list, err := stub.GetBecomeMediatorApplyList()
	if err != nil {
		return shim.Error(err.Error())
	}
	if list == nil {
		return shim.Error("申请列表为空。")
	}
	var mediatorList []*modules.MediatorInfo
	var mediator *modules.MediatorInfo
	//不同意，移除申请列表
	if args[0] == "不同意" {
		mediatorList, _ = MoveFromList(args[1], list.MediatorList)
	} else if args[0] == "同意" {
		//同意，移除列表，并且加入同意申请列表
		mediatorList, mediator = MoveFromList(args[1], list.MediatorList)
		//获取同意列表
		agreeList, err := stub.GetAgreeForBecomeMediatorList()
		if err != nil {
			return shim.Error(err.Error())
		}
		if agreeList == nil {
			agreeList = new(modules.AgreeForBecomeMediatorList)
			agreeList.MediatorList = append(agreeList.MediatorList, mediator)
		} else {
			agreeList.MediatorList = append(agreeList.MediatorList, mediator)
		}
		agreeListByte, err := json.Marshal(agreeList)
		if err != nil {
			return shim.Error(err.Error())
		}
		stub.PutState("ListForAgreeBecomeMediator", agreeListByte)
	}
	list.MediatorList = mediatorList
	listByte, err := json.Marshal(list)
	if err != nil {
		return shim.Error(err.Error())
	}
	stub.PutState("ListForApplyBecomeMediator", listByte)
	return shim.Success([]byte("ok"))
}

func MoveFromList(address string, list []*modules.MediatorInfo) (list1 []*modules.MediatorInfo, mediator *modules.MediatorInfo) {
	for i := 0; i < len(list); i++ {
		if strings.Compare(list[i].Address, address) == 0 {
			mediator = list[i]
			list1 = append(list[:i], list[i+1:]...)
			break
		}
	}
	return
}

//申请加入  参数：暂时 节点地址
func ApplyForQuitMediator(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 1 {
		return shim.Error("arg need one parameter.")
	}
	addr := args[0]
	mediatorInfo := modules.MediatorInfo{
		Address: addr,
	}
	//获取列表
	list, err := stub.GetQuitMediatorApplyList()
	if err != nil {
		return shim.Error(err.Error())
	}
	if list == nil {
		list = new(modules.QuitMediatorApplyList)
		list.MediatorList = append(list.MediatorList, &mediatorInfo)
	} else {
		list.MediatorList = append(list.MediatorList, &mediatorInfo)
	}
	listByte, err := json.Marshal(list)
	if err != nil {
		return shim.Error(err.Error())
	}
	stub.PutState("ListForApplyQuitMediator", listByte)
	return shim.Success([]byte("ok"))
}

//处理退出 参数：同意或不同意，节点的地址
func HandleForApplyForQuitMediator(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 2 {
		return shim.Error("arg need two parameter.")
	}
	//获取申请列表
	list, err := stub.GetQuitMediatorApplyList()
	if err != nil {
		return shim.Error(err.Error())
	}
	if list == nil {
		return shim.Error("申请列表为空。")
	}
	var mediatorList []*modules.MediatorInfo
	//不同意，移除申请列表
	if args[0] == "不同意" {
		mediatorList, _ = MoveFromList(args[1], list.MediatorList)
	} else if args[0] == "同意" {
		//同意，移除列表，并且全款退出
		mediatorList, _ = MoveFromList(args[1], list.MediatorList)
		//处理退款
		//获取该账户
		balance, err := stub.GetDepositBalance(args[0])
		if err != nil {
			return shim.Error(err.Error())
		}
		if balance == nil {
			return shim.Error("balance is nil.")
		}
		endTime := time.Now().UTC()
		coinDays := award.GetCoinDay(balance.TotalAmount, balance.LastModifyTime, endTime)
		//计算币龄收益
		awards := award.CalculateAwardsForDepositContractNodes(coinDays)
		//本金+利息
		balance.TotalAmount += awards
		invokeTokens := new(modules.InvokeTokens)
		invokeTokens.Amount = balance.TotalAmount
		asset := modules.NewPTNAsset()
		invokeTokens.Asset = asset
		//调用从合约把token转到请求地址
		err = stub.PayOutToken(args[0], invokeTokens, 0)
		if err != nil {
			return shim.Error(err.Error())
		}
		add, err := common.StringToAddress(args[0])
		if err != nil {
			return shim.Error(err.Error())
		}
		//移除出候选列表
		handleMember("Mediator", add, stub)
		//删除节点
		err = stub.DelState(args[0])
		if err != nil {
			return shim.Error(err.Error())
		}
	}
	list.MediatorList = mediatorList
	listByte, err := json.Marshal(list)
	if err != nil {
		return shim.Error(err.Error())
	}
	stub.PutState("ListForApplyQuitMediator", listByte)
	return shim.Success([]byte("ok"))
}
