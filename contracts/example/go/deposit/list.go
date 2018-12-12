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
	"github.com/palletone/go-palletone/dag/modules"
)

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
	//先获取状态数据库中的 Developer 列表
	developerListBytes, _ := stub.GetState("DeveloperList")
	developerList := []*common.Address{}
	_ = json.Unmarshal(developerListBytes, &developerList)
	//fmt.Printf("developerList = %#v\n", developerList)
	developerList = append(developerList, &invokeAddr)
	developerListBytes, _ = json.Marshal(developerList)
	stub.PutState("DeveloperList", developerListBytes)
}

//加入 Mediator 列表
func addMediatorList(invokeAddr common.Address, stub shim.ChaincodeStubInterface) {
	//先获取状态数据库中的 Mediator 列表
	mediatorListBytes, _ := stub.GetState("MediatorList")
	mediatorList := []*common.Address{}
	_ = json.Unmarshal(mediatorListBytes, &mediatorList)
	//fmt.Printf("MediatorList = %#v\n", mediatorList)
	mediatorList = append(mediatorList, &invokeAddr)
	mediatorListBytes, _ = json.Marshal(mediatorList)
	//更新列表
	stub.PutState("MediatorList", mediatorListBytes)
}

//加入 Jury 列表
func addJuryList(invokeAddr common.Address, stub shim.ChaincodeStubInterface) {
	//先获取状态数据库中的 Jury 列表
	juryListBytes, _ := stub.GetState("JuryList")
	juryList := []*common.Address{}
	_ = json.Unmarshal(juryListBytes, &juryList)
	//fmt.Printf("JuryList = %#v\n", juryList)
	juryList = append(juryList, &invokeAddr)
	juryListBytes, _ = json.Marshal(juryList)
	stub.PutState("JuryList", juryListBytes)
}

//无论是退款还是罚款，在相应列表中移除节点
func handleMember(who string, invokeFromAddr common.Address, stub shim.ChaincodeStubInterface) {
	switch {
	case who == "Mediator":
		listBytes, _ := stub.GetState("MediatorList")
		mediatorList := []common.Address{}
		_ = json.Unmarshal(listBytes, &mediatorList)
		//从列表中移除该节点
		move("MediatorList", mediatorList, invokeFromAddr, stub)
	case who == "Jury":
		listBytes, _ := stub.GetState("JuryList")
		juryList := []common.Address{}
		_ = json.Unmarshal(listBytes, &juryList)
		//从列表中移除该节点
		move("JuryList", juryList, invokeFromAddr, stub)
	case who == "Developer":
		listBytes, _ := stub.GetState("DeveloperList")
		developerList := []common.Address{}
		_ = json.Unmarshal(listBytes, &developerList)
		//从列表中移除该节点
		move("DeveloperList", developerList, invokeFromAddr, stub)
	}
}

//从候选列表中移除
func move(who string, list []common.Address, invokeAddr common.Address, stub shim.ChaincodeStubInterface) {
	for i := 0; i < len(list); i++ {
		if list[i] == invokeAddr {
			list = append(list[:i], list[i+1:]...)
			break
		}
	}
	listBytes, _ := json.Marshal(list)
	//更新列表
	stub.PutState(who, listBytes)
}

//从申请没收保证金列表中移除
func moveInApplyForForfeitureList(stub shim.ChaincodeStubInterface, listForForfeiture *modules.ListForForfeiture, forfeitureAddr common.Address, applyTime int64) (*modules.Forfeiture, error) {
	//
	forfeiture := new(modules.Forfeiture)
	for i := 0; i < len(listForForfeiture.Forfeitures); i++ {
		if listForForfeiture.Forfeitures[i].ApplyTime == applyTime && listForForfeiture.Forfeitures[i].ForfeitureAddress == forfeitureAddr {
			forfeiture = listForForfeiture.Forfeitures[i]
			listForForfeiture.Forfeitures = append(listForForfeiture.Forfeitures[:i], listForForfeiture.Forfeitures[i+1:]...)
			break
		}
	}
	listForForfeitureByte, err := json.Marshal(listForForfeiture)
	if err != nil {
		return nil, err
	}
	//更新列表
	stub.PutState("ListForForfeiture", listForForfeitureByte)
	return forfeiture, nil
}

//从申请没收保证金列表中移除
func moveInApplyForCashbackList(stub shim.ChaincodeStubInterface, listForCashback *modules.ListForCashback, cashbackAddr common.Address, applyTime int64) (*modules.Cashback, error) {
	//
	cashback := new(modules.Cashback)
	for i := 0; i < len(listForCashback.Cashbacks); i++ {
		if listForCashback.Cashbacks[i].CashbackTime == applyTime && listForCashback.Cashbacks[i].CashbackAddress == cashbackAddr {
			cashback = listForCashback.Cashbacks[i]
			listForCashback.Cashbacks = append(listForCashback.Cashbacks[:i], listForCashback.Cashbacks[i+1:]...)
			break
		}
	}
	listForCashbackByte, err := json.Marshal(listForCashback)
	if err != nil {
		return nil, err
	}
	//更新列表
	stub.PutState("ListForCashback", listForCashbackByte)
	return cashback, nil
}
