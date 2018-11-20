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
	"github.com/palletone/go-palletone/contracts/shim"
)

//判断原来是 Jury 还是 Mediator 还是什么都不是
func whoIs(amount uint64) string {
	switch {
	case amount >= depositAmountsForMediator:
		return "Mediator"
	case amount >= depositAmountsForJury:
		return "Jury"
	default:
		return ""
	}
}

//判断要成为 Jury 还是 Mediator 还是 Developer
func addList(role string, invokeaddr common.Address, stub shim.ChaincodeStubInterface) {
	switch {
	case role == "Mediator":
		//加入 Mediator 列表
		addMediatorList(invokeaddr, stub)
		//加入 Jury 列表
	case role == "Jury":
		addJuryList(invokeaddr, stub)
	case role == "Developer":
		//加入 Developer 列表
		addDeveloperList(invokeaddr, stub)
	}
}

//加入 Developer 列表
func addDeveloperList(invokeAddr common.Address, stub shim.ChaincodeStubInterface) {
	//先获取状态数据库中的 Mediator 列表
	developerListBytes, _ := stub.GetState("Developer")
	developerList := []common.Address{}
	_ = json.Unmarshal(developerListBytes, &developerList)
	//fmt.Printf("developerList = %#v\n", developerList)
	developerList = append(developerList, invokeAddr)
	developerListBytes, _ = json.Marshal(developerList)
	stub.PutState("Developer", developerListBytes)
}

//加入 Mediator 列表
func addMediatorList(invokeAddr common.Address, stub shim.ChaincodeStubInterface) {
	//先获取状态数据库中的 Mediator 列表
	mediatorListBytes, _ := stub.GetState("MediatorList")
	mediatorList := []common.Address{}
	_ = json.Unmarshal(mediatorListBytes, &mediatorList)
	//fmt.Printf("MediatorList = %#v\n", mediatorList)
	mediatorList = append(mediatorList, invokeAddr)
	mediatorListBytes, _ = json.Marshal(mediatorList)
	stub.PutState("MediatorList", mediatorListBytes)
}

//加入 Jury 列表
func addJuryList(invokeAddr common.Address, stub shim.ChaincodeStubInterface) {
	//先获取状态数据库中的 Jury 列表
	juryListBytes, _ := stub.GetState("JuryList")
	juryList := []common.Address{}
	_ = json.Unmarshal(juryListBytes, &juryList)
	//fmt.Printf("JuryList = %#v\n", juryList)
	juryList = append(juryList, invokeAddr)
	juryListBytes, _ = json.Marshal(juryList)
	stub.PutState("JuryList", juryListBytes)
}

//无论是退款还是罚款，作相应处理
func handleMember(who string, invokeFromAddr common.Address, amount uint64, stub shim.ChaincodeStubInterface) {
	//判断退款后是 Jury 还是 Mediator 还是  都不是（即移除列表）
	switch {
	case amount >= depositAmountsForJury && amount < depositAmountsForMediator:
		//如果一开始是mediator,无论是退款还是罚款，那么都先从mediatorlist 移除，再加入jurylist
		if who == "Mediator" {
			listBytes, _ := stub.GetState(who)
			mediatorList := []common.Address{}
			_ = json.Unmarshal(listBytes, &mediatorList)
			//从Mediator列表移除
			move(who, mediatorList, invokeFromAddr, stub)
			//加入JuryList
			addJuryList(invokeFromAddr, stub)
		}
		//who == "Jury"
		//如果一开始是Jury,无论是退款还是罚款,还在列表中，所以不用操作
	//case amount >= depositAmountsForMediator && who == "Mediator":
	//如果一开始是mediator,无论是退款还是罚款,还在列表中，所以不用操作
	case amount < depositAmountsForJury:
		//如果一开始是Jury,无论是退款还是罚款,移除jury列表
		if who == "Jury" {
			listBytes, _ := stub.GetState(who)
			juryList := []common.Address{}
			_ = json.Unmarshal(listBytes, &juryList)
			move(who, juryList, invokeFromAddr, stub)
		}
		if who == "Mediator" {
			//如果一开始是mediator,无论是退款还是罚款,移除mediator列表
			listBytes, _ := stub.GetState(who)
			mediatorList := []common.Address{}
			_ = json.Unmarshal(listBytes, &mediatorList)
			move(who, mediatorList, invokeFromAddr, stub)
		}
		//case amount >= depositAmountsForMediator && who == "":
		//一开始不在列表中，但是就是成为 Mediator
		//addMediatorList(invokeFromAddr, stub)
		//case amount >= depositAmountsForJury && who == "":
		//一开始不在列表中，但是就是成为 Jury
		//addJuryList(invokeFromAddr, stub)
		//case amount < depositAmountsForJury && who == "":
		//不够，不操作
	}
}

//从列表中移除
func move(who string, list []common.Address, invokeAddr common.Address, stub shim.ChaincodeStubInterface) {
	for i := 0; i < len(list); i++ {
		if list[i] == invokeAddr {
			list = append(list[:i], list[i+1:]...)
			break
		}
	}
	listBytes, _ := json.Marshal(list)
	stub.PutState(who, listBytes)
}
