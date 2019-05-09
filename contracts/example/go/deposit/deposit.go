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
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/contracts/shim"
	pb "github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
	"github.com/palletone/go-palletone/dag/modules"
)

type DepositChaincode struct {
}

func (d *DepositChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	log.Info("*** DepositChaincode system contract init ***")
	return shim.Success([]byte("init ok"))
}

func (d *DepositChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	funcName, args := stub.GetFunctionAndParameters()
	switch funcName {
	case modules.ApplyMediator:
		//log.Info("Enter DepositChaincode Contract ApplyBecomeMediator Invoke")
		//申请成为Mediator
		return d.applyBecomeMediator(stub, args)
	case "HandleForApplyBecomeMediator":
		//log.Info("Enter DepositChaincode Contract HandleForApplyBecomeMediator Invoke")
		//基金会对加入申请Mediator进行处理
		return d.handleForApplyBecomeMediator(stub, args)
	case "MediatorApplyQuitMediator":
		//申请退出Mediator
		return d.mediatorApplyQuitMediator(stub, args)
	case "HandleForApplyQuitMediator":
		//基金会对退出申请Mediator进行处理
		return d.handleForApplyQuitMediator(stub, args)
	case "MediatorPayToDepositContract":
		//mediator 交付保证金
		return d.mediatorPayToDepositContract(stub, args)
	case "JuryPayToDepositContract":
		//jury 交付保证金
		return d.juryPayToDepositContract(stub, args)
	case "DeveloperPayToDepositContract":
		//developer 交付保证金
		return d.developerPayToDepositContract(stub, args)
	case "MediatorApplyCashback":
		//mediator 申请提取保证金
		return d.mediatorApplyCashback(stub, args)
	case "HandleForMediatorApplyCashback":
		//基金会处理提取保证金
		return d.handleForMediatorApplyCashback(stub, args)
	case "JuryApplyCashback":
		//jury 申请提取保证金
		return d.juryApplyCashback(stub, args)
	case "HandleForJuryApplyCashback":
		//基金会处理提取保证金
		return d.handleForJuryApplyCashback(stub, args)
	case "DeveloperApplyCashback":
		//developer 申请提取保证金
		return d.developerApplyCashback(stub, args)
	case "HandleForDeveloperApplyCashback":
		//基金会处理提取保证金
		return d.handleForDeveloperApplyCashback(stub, args)
	case "ApplyForForfeitureDeposit":
		//申请保证金没收
		//void forfeiture_deposit(const witness_object& wit, token_type amount)
		return d.applyForForfeitureDeposit(stub, args)
	case "HandleForForfeitureApplication":
		//基金会对申请做相应的处理
		return d.handleForForfeitureApplication(stub, args)
	//获取提取保证金申请列表
	case "GetListForCashbackApplication":
		list, err := stub.GetState(ListForCashback)
		if err != nil {
			return shim.Error(err.Error())
		}
		if list == nil {
			return shim.Success([]byte("[]"))
		}
		return shim.Success(list)
		//获取没收保证金申请列表
	case "GetListForForfeitureApplication":
		list, err := stub.GetState(ListForForfeiture)
		if err != nil {
			return shim.Error(err.Error())
		}
		if list == nil {
			return shim.Success([]byte("[]"))
		}
		return shim.Success(list)
		//获取Mediator候选列表
	case "GetListForMediatorCandidate":
		list, err := stub.GetState(modules.MediatorList)
		if err != nil {
			return shim.Error(err.Error())
		}
		if list == nil {
			return shim.Success([]byte("[]"))
		}
		return shim.Success(list)
		//获取Jury候选列表
	case "GetListForJuryCandidate":
		list, err := stub.GetState(modules.JuryList)
		if err != nil {
			return shim.Error(err.Error())
		}
		if list == nil {
			return shim.Success([]byte("[]"))
		}
		return shim.Success(list)
		//获取Contract Developer候选列表
	case "GetListForDeveloperCandidate":
		list, err := stub.GetState(DeveloperList)
		if err != nil {
			return shim.Error(err.Error())
		}
		if list == nil {
			return shim.Success([]byte("[]"))
		}
		return shim.Success(list)
		//获取某个节点的账户
	case "GetBalanceWithAddr":
		balance, err := stub.GetState(args[0])
		if err != nil {
			return shim.Error(err.Error())
		}
		if balance == nil {
			return shim.Success([]byte("balance is nil"))
		}
		if string(balance) == "" {
			return shim.Success([]byte("balance is nil"))
		}
		depositB := &DepositBalance{}
		err = json.Unmarshal(balance, depositB)
		if err != nil {
			return shim.Error(err.Error())
		}
		if depositB.EnterTime != "" {
			depositB.EnterTime = timeFormat(depositB.EnterTime)
		}
		balance, err = json.Marshal(depositB)
		if err != nil {
			return shim.Error(err.Error())
		}
		return shim.Success(balance)
		//获取Mediator申请加入列表
	case "GetBecomeMediatorApplyList":
		log.Info("Enter DepositChaincode Contract GetBecomeMediatorApplyList Invoke")
		list, err := stub.GetState(ListForApplyBecomeMediator)
		if err != nil {
			return shim.Error(err.Error())
		}
		if list == nil {
			return shim.Success([]byte("[]"))
		}
		return shim.Success(list)
		//获取已同意的mediator列表
	case "GetAgreeForBecomeMediatorList":
		log.Info("Enter DepositChaincode Contract GetAgreeForBecomeMediatorList Invoke")
		list, err := stub.GetState(ListForAgreeBecomeMediator)
		if err != nil {
			return shim.Error(err.Error())
		}
		if list == nil {
			return shim.Success([]byte("[]"))
		}
		return shim.Success(list)
		//获取Mediator申请退出列表
	case "GetQuitMediatorApplyList":
		list, err := stub.GetState(ListForApplyQuitMediator)
		if err != nil {
			return shim.Error(err.Error())
		}
		if list == nil {
			return shim.Success([]byte("[]"))
		}
		return shim.Success(list)
		//查看是否申请Mediator通过
	case "IsSelected":
		mediatorRegisterInfo, err := GetAgreeForBecomeMediatorList(stub)
		if err != nil {
			return shim.Error(err.Error())
		}
		if mediatorRegisterInfo == nil {
			return shim.Success([]byte("list is nil"))
		}
		for _, m := range mediatorRegisterInfo {
			if args[0] == m.Address {
				return shim.Success([]byte("had pass"))
			}
		}
		return shim.Success([]byte("no pass"))
	}

	return shim.Error("Please enter validate function name.")
}

func (d *DepositChaincode) mediatorPayToDepositContract(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	return mediatorPayToDepositContract(stub, args)
}

func (d *DepositChaincode) juryPayToDepositContract(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	return juryPayToDepositContract(stub, args)
}

func (d *DepositChaincode) developerPayToDepositContract(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	return developerPayToDepositContract(stub, args)
}
func (d *DepositChaincode) mediatorApplyCashback(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	return mediatorApplyCashback(stub, args)
}
func (d *DepositChaincode) juryApplyCashback(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	return juryApplyCashback(stub, args)
}
func (d *DepositChaincode) developerApplyCashback(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	return developerApplyCashback(stub, args)
}

func (d *DepositChaincode) handleForMediatorApplyCashback(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	return handleForMediatorApplyCashback(stub, args)
}

func (d *DepositChaincode) handleForJuryApplyCashback(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	return handleForJuryApplyCashback(stub, args)
}

func (d *DepositChaincode) handleForDeveloperApplyCashback(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	return handleForDeveloperApplyCashback(stub, args)
}

//申请加入Mediator
func (d *DepositChaincode) applyBecomeMediator(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	return applyBecomeMediator(stub, args)
}

//基金会对申请加入Mediator进行处理
func (d *DepositChaincode) handleForApplyBecomeMediator(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	return handleForApplyBecomeMediator(stub, args)
}

//申请退出Mediator
func (d *DepositChaincode) mediatorApplyQuitMediator(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	return mediatorApplyQuitMediator(stub, args)
}

//基金会对申请退出Mediator进行处理
func (d *DepositChaincode) handleForApplyQuitMediator(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	return handleForApplyQuitMediator(stub, args)
}
