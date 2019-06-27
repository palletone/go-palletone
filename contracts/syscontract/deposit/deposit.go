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

//  Package deposit implements some functions for deposit contract.
package deposit

import (
	"encoding/json"

	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/contracts/shim"
	pb "github.com/palletone/go-palletone/core/vmContractPub/protos/peer"

	"github.com/palletone/go-palletone/dag/modules"
	"strconv"
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
	//
	//  申请成为Mediator
	case modules.ApplyMediator:
		log.Info("Enter DepositChaincode Contract " + modules.ApplyMediator + " Invoke")
		return d.applyBecomeMediator(stub, args)
		//  mediator 交付保证金
	case modules.MediatorPayDeposit:
		log.Info("Enter DepositChaincode Contract " + modules.MediatorPayDeposit + " Invoke")
		return d.mediatorPayToDepositContract(stub, args)
		//  申请退出Mediator
	case modules.MediatorApplyQuitList:
		log.Info("Enter DepositChaincode Contract " + modules.MediatorApplyQuitList + " Invoke")
		return d.mediatorApplyQuitMediator(stub, args)
	//
	//  jury 交付保证金
	case JuryPayToDepositContract:
		log.Info("Enter DepositChaincode Contract " + JuryPayToDepositContract + " Invoke")
		return d.juryPayToDepositContract(stub, args)
		//  jury 申请提取保证金
	case JuryApplyQuit:
		log.Info("Enter DepositChaincode Contract " + JuryApplyCashback + " Invoke")
		return d.juryApplyQuit(stub, args)
	//
	//  developer 交付保证金
	case DeveloperPayToDepositContract:
		log.Info("Enter DepositChaincode Contract " + DeveloperPayToDepositContract + " Invoke")
		return d.developerPayToDepositContract(stub, args)
		//  developer 申请提取保证金
	case DeveloperApplyQuit:
		log.Info("Enter DepositChaincode Contract " + DeveloperApplyCashback + " Invoke")
		return d.devApplyQuit(stub, args)
	//
	//  基金会对加入申请Mediator进行处理
	case HandleForApplyBecomeMediator:
		log.Info("Enter DepositChaincode Contract " + HandleForApplyBecomeMediator + " Invoke")
		return d.handleForApplyBecomeMediator(stub, args)
		//  基金会对退出申请Mediator进行处理
	case HandleForApplyQuitMediator:
		log.Info("Enter DepositChaincode Contract " + HandleForApplyQuitMediator + " Invoke")
		return d.handleForApplyQuitMediator(stub, args)
	case HandleForApplyQuitJury:
		log.Info("Enter DepositChaincode Contract " + HandleForApplyQuitJury + " Invoke")
		return d.handleForApplyQuitJury(stub, args)
	case HandleForApplyQuitDev:
		log.Info("Enter DepositChaincode Contract " + HandleForApplyQuitDev + " Invoke")
		return d.handleForApplyQuitDev(stub, args)
		//  基金会对申请做相应的处理
	case HandleForForfeitureApplication:
		log.Info("Enter DepositChaincode Contract " + HandleForForfeitureApplication + " Invoke")
		return d.handleForForfeitureApplication(stub, args)
	//
	//  申请保证金没收
	case ApplyForForfeitureDeposit:
		log.Info("Enter DepositChaincode Contract " + ApplyForForfeitureDeposit + " Invoke")

		return d.applyForForfeitureDeposit(stub, args)
	//
	//  获取Mediator申请加入列表
	case GetBecomeMediatorApplyList:
		log.Info("Enter DepositChaincode Contract " + GetBecomeMediatorApplyList + " Invoke")
		list, err := stub.GetState(ListForApplyBecomeMediator)
		if err != nil {
			return shim.Error(err.Error())
		}
		if list == nil {
			return shim.Success([]byte("{}"))
		}
		return shim.Success(list)
		//  获取已同意的mediator列表
	case GetAgreeForBecomeMediatorList:
		log.Info("Enter DepositChaincode Contract " + GetAgreeForBecomeMediatorList + " Invoke")
		list, err := stub.GetState(ListForAgreeBecomeMediator)
		if err != nil {
			return shim.Error(err.Error())
		}
		if list == nil {
			return shim.Success([]byte("{}"))
		}
		return shim.Success(list)
		//  获取Mediator候选列表
	case GetListForMediatorCandidate:
		log.Info("Enter DepositChaincode Contract " + GetListForMediatorCandidate + " Invoke")
		list, err := stub.GetState(modules.MediatorList)
		if err != nil {
			return shim.Error(err.Error())
		}
		if list == nil {
			return shim.Success([]byte("{}"))
		}
		return shim.Success(list)
		//获取Mediator申请退出列表
	case GetQuitApplyList:
		log.Info("Enter DepositChaincode Contract " + GetQuitApplyList + " Invoke")
		list, err := stub.GetState(ListForQuit)
		if err != nil {
			return shim.Error(err.Error())
		}
		if list == nil {
			return shim.Success([]byte("{}"))
		}
		return shim.Success(list)
		//  获取Jury候选列表
	case GetListForJuryCandidate:
		log.Info("Enter DepositChaincode Contract " + GetListForJuryCandidate + " Invoke")
		list, err := stub.GetState(modules.JuryList)
		if err != nil {
			return shim.Error(err.Error())
		}
		if list == nil {
			return shim.Success([]byte("{}"))
		}
		return shim.Success(list)
		//  获取Contract Developer候选列表
	case GetListForDeveloperCandidate:
		log.Info("Enter DepositChaincode Contract " + GetListForDeveloperCandidate + " Invoke")
		list, err := stub.GetState(modules.DeveloperList)
		if err != nil {
			return shim.Error(err.Error())
		}
		if list == nil {
			return shim.Success([]byte("{}"))
		}
		return shim.Success(list)
		//  获取没收保证金申请列表
	case GetListForForfeitureApplication:
		log.Info("Enter DepositChaincode Contract " + GetListForForfeitureApplication + " Invoke")
		list, err := stub.GetState(ListForForfeiture)
		if err != nil {
			return shim.Error(err.Error())
		}
		if list == nil {
			return shim.Success([]byte("{}"))
		}
		return shim.Success(list)
		//  获取节点的账户
	case GetDeposit:
		log.Info("Enter DepositChaincode Contract " + GetDeposit + " Invoke")
		balance, err := GetNodeBalance(stub, args[0])
		if err != nil {
			return shim.Error(err.Error())
		}
		if balance == nil {
			return shim.Success([]byte("balance is nil"))
		}
		byte, err := json.Marshal(balance)
		if err != nil {
			return shim.Error(err.Error())
		}
		return shim.Success(byte)
		// 获取mediator Deposit
	case modules.GetMediatorDeposit:
		log.Info("Enter DepositChaincode Contract " + modules.GetMediatorDeposit + " Invoke")
		mediator, err := GetMediatorDeposit(stub, args[0])
		if err != nil {
			return shim.Error(err.Error())
		}
		if mediator == nil {
			return shim.Success([]byte("mediator is nil"))
		}
		byte, err := json.Marshal(mediator)
		if err != nil {
			return shim.Error(err.Error())
		}
		return shim.Success(byte)
	// 查看是否申请Mediator通过
	case modules.IsApproved:
		log.Info("Enter DepositChaincode Contract " + modules.IsApproved + " Invoke")
		mediatorRegisterInfo, err := getList(stub, ListForAgreeBecomeMediator)
		if err != nil {
			return shim.Error(err.Error())
		}
		if mediatorRegisterInfo == nil {
			return shim.Success([]byte("not in the supernode become list,please wait while approval..."))
		}
		if _, ok := mediatorRegisterInfo[args[0]]; ok {
			return shim.Success([]byte("you have already applied, please go to the delivery deposit to join the super node candidate list as soon as possible"))
		}
		return shim.Success([]byte("not in the supernode become list,please wait while approval..."))
		//  查看是否在候选列表中
	case IsInMediatorCandidateList:
		log.Info("Enter DepositChaincode Contract " + IsInMediatorCandidateList + " Invoke")
		mediatorRegisterInfo, err := getList(stub, modules.MediatorList)
		if err != nil {
			return shim.Error(err.Error())
		}
		if mediatorRegisterInfo == nil {
			return shim.Success([]byte("super node candidate list is empty"))
		}
		if _, ok := mediatorRegisterInfo[args[0]]; ok {
			return shim.Success([]byte("joined the super node candidate list"))
		}
		return shim.Success([]byte("not in the supernode candidate list"))
		//  查看是否在退出列表中
	case IsInMediatorQuitList:
		log.Info("Enter DepositChaincode Contract " + IsInMediatorQuitList + " Invoke")
		list, err := getList(stub, ListForApplyQuitMediator)
		if err != nil {
			return shim.Error(err.Error())
		}
		if list == nil {
			return shim.Success([]byte("not in the supernode quit list"))
		}
		if _, ok := list[args[0]]; ok {
			return shim.Success([]byte("in the supernode quit list"))
		}
		return shim.Success([]byte("not in the supernode quit list"))
		//  查看是否在退还列表中，并返回节点信息
	case IsInCashbackList:
		log.Info("Enter DepositChaincode Contract " + IsInCashbackList + " Invoke")
		list, err := GetListForQuit(stub)
		if err != nil {
			return shim.Error(err.Error())
		}
		if list == nil {
			return shim.Success([]byte("not in the cashback list"))
		}
		if _, ok := list[args[0]]; ok {
			byte, err := json.Marshal(list[args[0]])
			if err != nil {
				return shim.Error(err.Error())
			}
			return shim.Success(byte)
		}
		return shim.Success([]byte("not in the cashback list"))
		//  查看jury是否在候选列表中
	case IsInJuryCandidateList:
		log.Info("Enter DepositChaincode Contract " + IsInJuryCandidateList + " Invoke")
		jByte, err := getList(stub, modules.JuryList)
		if err != nil {
			return shim.Error(err.Error())
		}
		if jByte == nil {
			return shim.Success([]byte("jury node candidate list is empty"))
		}
		if _, ok := jByte[args[0]]; ok {
			return shim.Success([]byte("joined the jury node candidate list"))
		}
		return shim.Success([]byte("not in the jury candidate list"))
		//  查看developer是否在候选列表中
	case IsInDeveloperCandidateList:
		log.Info("Enter DepositChaincode Contract " + IsInDeveloperCandidateList + " Invoke")
		dByte, err := getList(stub, modules.DeveloperList)
		if err != nil {
			return shim.Error(err.Error())
		}
		if dByte == nil {
			return shim.Success([]byte("jury node candidate list is empty"))
		}
		if _, ok := dByte[args[0]]; ok {
			return shim.Success([]byte("joined the jury node candidate list"))
		}
		return shim.Success([]byte("not in the jury candidate list"))

		//  普通用户质押投票
	case PledgeDeposit:
		log.Info("Enter DepositChaincode Contract " + PledgeDeposit + " Invoke")
		return processPledgeDeposit(stub, args)
	case PledgeWithdraw: //提币质押申请（如果提币申请金额为MaxUint64表示全部提现）
		log.Info("Enter DepositChaincode Contract " + PledgeWithdraw + " Invoke")
		return processPledgeWithdraw(stub, args)

	case QueryPledgeStatusByAddr:
		log.Info("Enter DepositChaincode Contract " + QueryPledgeStatusByAddr + " Query")
		return queryPledgeStatusByAddr(stub, args)
	//case HandleExtractVote:
	//	log.Info("Enter DepositChaincode Contract " + HandleExtractVote + " Invoke")
	//	return d.handleExtractVote(stub, args)
	//质押分红处理
	case HandleEachDayAward:
		log.Info("Enter DepositChaincode Contract " + HandleEachDayAward + " Invoke")
		return d.handleEachDayAward(stub, args)
		//TODO Devin 查询质押分红历史
		//TODO Devin查询当前质押分红情况
		//TODO Devin一个用户，怎么查看自己的流水账？
	case AllPledgeVotes:
		b, err := getVotes(stub)
		if err != nil {
			return shim.Error(err.Error())
		}
		st := strconv.FormatInt(b, 10)
		return shim.Success([]byte(st))
	case GetPledgeList:
		log.Info("Enter DepositChaincode Contract " + GetPledgeList + " Invoke")
		return d.getPledgeList(stub, args)
		//case ExtractPtnList:
		//	b, err := stub.GetState(ExtractPtnList)
		//	if err != nil {
		//		return shim.Error(err.Error())
		//	}
		//	return shim.Success(b)
		//case NormalNodeList:
		//	b, err := stub.GetState(NormalNodeList)
		//	if err != nil {
		//		return shim.Error(err.Error())
		//	}
		//	return shim.Success(b)
		//case "normalNode":
		//	b, err := stub.GetState(string(constants.DEPOSIT_MEDIATOR_VOTE_PREFIX) + args[0])
		//	if err != nil {
		//		return shim.Error(err.Error())
		//	}
		//	return shim.Success(b)
	}
	return shim.Error("please enter validate function name")
}
func (d *DepositChaincode) getPledgeList(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	list, err := getLastPledgeList(stub)
	if err != nil {
		return shim.Error(err.Error())
	}
	result, _ := json.Marshal(list)
	return shim.Success(result)
}
func (d *DepositChaincode) applyBecomeMediator(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	return applyBecomeMediator(stub, args)
}

func (d *DepositChaincode) mediatorPayToDepositContract(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	return mediatorPayToDepositContract(stub, args)
}

func (d *DepositChaincode) mediatorApplyQuitMediator(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	return mediatorApplyQuit(stub, args)
}

//

func (d *DepositChaincode) juryPayToDepositContract(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	return juryPayToDepositContract(stub, args)
}
func (d *DepositChaincode) juryApplyQuit(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	return juryApplyQuit(stub, args)
}

//

func (d *DepositChaincode) developerPayToDepositContract(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	return developerPayToDepositContract(stub, args)
}
func (d *DepositChaincode) devApplyQuit(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	return devApplyQuit(stub, args)
}

//

//基金会对申请加入Mediator进行处理
func (d *DepositChaincode) handleForApplyBecomeMediator(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	return handleForApplyBecomeMediator(stub, args)
}

//基金会对申请退出Mediator进行处理
func (d *DepositChaincode) handleForApplyQuitMediator(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	return handleForApplyQuitMediator(stub, args)
}

func (d *DepositChaincode) handleForApplyQuitJury(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	return handleForApplyQuitJury(stub, args)
}

func (d *DepositChaincode) handleForApplyQuitDev(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	return handleForApplyQuitDev(stub, args)
}

func (d *DepositChaincode) handleForForfeitureApplication(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	return handleForForfeitureApplication(stub, args)
}

//

func (d DepositChaincode) applyForForfeitureDeposit(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	return applyForForfeitureDeposit(stub, args)
}

//
func (d DepositChaincode) normalNodePledgeVote(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	return processPledgeDeposit(stub, args)
}

//func (d DepositChaincode) normalNodeChangeVote(stub shim.ChaincodeStubInterface, args []string) pb.Response {
//	return normalNodeChangeVote(stub, args)
//}

func (d DepositChaincode) normalNodeExtractVote(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	return processPledgeWithdraw(stub, args)
}

//func (d DepositChaincode) handleExtractVote(stub shim.ChaincodeStubInterface, args []string) pb.Response {
//	return handleExtractVote(stub, args)
//}

func (d DepositChaincode) handleEachDayAward(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	return handleEachDayAward(stub, args)
}
