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
	"github.com/palletone/go-palletone/dag/constants"
	"github.com/palletone/go-palletone/dag/modules"
)

type DepositChaincode struct {
}

func (d *DepositChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	log.Info("*** DepositChaincode system contract init ***")
	return shim.Success(nil)
}

func (d *DepositChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	funcName, args := stub.GetFunctionAndParameters()
	switch funcName {
	//
	// 申请成为Mediator
	case modules.ApplyMediator:
		log.Info("Enter DepositChaincode Contract " + modules.ApplyMediator + " Invoke")
		if len(args) != 1 {
			errStr := "Arg need only one parameter."
			log.Error(errStr)
			return shim.Error(errStr)
		}
		return d.ApplyBecomeMediator(stub, args[0])
	// mediator 交付保证金
	case modules.MediatorPayDeposit:
		log.Info("Enter DepositChaincode Contract " + modules.MediatorPayDeposit + " Invoke")
		return d.MediatorPayToDepositContract(stub)
	// 申请退出Mediator
	case modules.MediatorApplyQuit:
		log.Info("Enter DepositChaincode Contract " + modules.MediatorApplyQuit + " Invoke")
		return d.MediatorApplyQuit(stub)
	// 更新 Mediator 信息
	case modules.UpdateMediatorInfo:
		log.Info("Enter DepositChaincode Contract " + modules.UpdateMediatorInfo + " Invoke")
		//  检查参数
		if len(args) != 1 {
			errStr := "Arg need only one parameter."
			log.Error(errStr)
			return shim.Error(errStr)
		}
		return d.UpdateMediatorInfo(stub, args[0])
	//
	//  jury 交付保证金
	case modules.JuryPayToDepositContract:
		log.Info("Enter DepositChaincode Contract " + modules.JuryPayToDepositContract + " Invoke")
		if len(args) != 1 {
			return shim.Error("need 1 parameter")
		}
		return d.JuryPayToDepositContract(stub, args[0])
		//  jury 申请退出
	case modules.JuryApplyQuit:
		log.Info("Enter DepositChaincode Contract " + modules.JuryApplyQuit + " Invoke")
		return d.JuryApplyQuit(stub)
	//
	//  developer 交付保证金
	case modules.DeveloperPayToDepositContract:
		log.Info("Enter DepositChaincode Contract " + modules.DeveloperPayToDepositContract + " Invoke")
		return d.DeveloperPayToDepositContract(stub)
		//  developer 申请退出
	case modules.DeveloperApplyQuit:
		log.Info("Enter DepositChaincode Contract " + modules.DeveloperApplyQuit + " Invoke")
		return d.DevApplyQuit(stub)
	//
	//  基金会对加入申请Mediator进行处理
	case modules.HandleForApplyBecomeMediator:
		log.Info("Enter DepositChaincode Contract " + modules.HandleForApplyBecomeMediator + " Invoke")
		if len(args) != 2 {
			log.Error("args need two parameters")
			return shim.Error("args need two parameters")
		}
		return d.HandleForApplyBecomeMediator(stub, args[0],args[1])
	//  基金会移除某个节点
	case modules.HanldeNodeRemoveFromAgreeList:
		log.Info("Enter DepositChaincode Contract " + modules.HanldeNodeRemoveFromAgreeList + " Invoke")
		if len(args) != 1 {
			return shim.Error("need 1 parameter")
		}
		return d.HandleNodeRemoveFromAgreeList(stub, args[0])
		//  基金会对退出申请Mediator进行处理
	case modules.HandleForApplyQuitMediator:
		log.Info("Enter DepositChaincode Contract " + modules.HandleForApplyQuitMediator + " Invoke")
		//参数
		if len(args) != 2 {
			log.Error("Arg need two parameter.")
			return shim.Error("Arg need two parameter.")
		}
		return d.HandleForApplyQuitMediator(stub, args[0],args[1])
		//  基金会对退出申请Jury进行处理
	case modules.HandleForApplyQuitJury:
		log.Info("Enter DepositChaincode Contract " + modules.HandleForApplyQuitJury + " Invoke")
		//参数
		if len(args) != 2 {
			log.Error("Arg need two parameter.")
			return shim.Error("Arg need two parameter.")
		}
		return d.HandleForApplyQuitJury(stub, args[0],args[1])
		//  基金会对退出申请Developer进行处理
	case modules.HandleForApplyQuitDev:
		log.Info("Enter DepositChaincode Contract " + modules.HandleForApplyQuitDev + " Invoke")
		//参数
		if len(args) != 2 {
			log.Error("Arg need two parameter.")
			return shim.Error("Arg need two parameter.")
		}
		return d.HandleForApplyQuitDev(stub, args[0],args[1])
		//  基金会对申请没收做相应的处理
	case modules.HandleForForfeitureApplication:
		log.Info("Enter DepositChaincode Contract " + modules.HandleForForfeitureApplication + " Invoke")
		//  地址，是否同意
		if len(args) != 2 {
			log.Error("args need two parameters.")
			return shim.Error("args need two parameters.")
		}
		return d.HandleForForfeitureApplication(stub, args[0],args[1])
	//
	//  申请保证金没收
	case modules.ApplyForForfeitureDeposit:
		log.Info("Enter DepositChaincode Contract " + modules.ApplyForForfeitureDeposit + " Invoke")
		if len(args) != 3 {
			log.Error("args need three parameters")
			return shim.Error("args need three parameters")
		}
		return d.ApplyForForfeitureDeposit(stub, args[0],args[1],args[2])
	//
	//  获取Mediator申请加入列表
	case modules.GetBecomeMediatorApplyList:
		log.Info("Enter DepositChaincode Contract " + modules.GetBecomeMediatorApplyList + " Query")
		return d.GetBecomeMediatorApplyList(stub)
		//  查看是否在become列表中
	case modules.IsInBecomeList:
		log.Info("Enter DepositChaincode Contract " + modules.IsInBecomeList + " Query")
		if len(args) != 1 {
			return shim.Error("arg need one")
		}
		return d.IsInBecomeList(stub,args[0])
		//  获取已同意的mediator列表
	case modules.GetAgreeForBecomeMediatorList:
		log.Info("Enter DepositChaincode Contract " + modules.GetAgreeForBecomeMediatorList + " Query")
		return d.GetAgreeForBecomeMediatorList(stub)
		//  查看是否在agree列表中
	case modules.IsApproved:
		log.Info("Enter DepositChaincode Contract " + modules.IsApproved + " Query")
		if len(args) != 1 {
			return shim.Error("arg need one")
		}
		return d.IsInAgreeList(stub,args[0])
		//获取申请退出列表
	case modules.GetQuitApplyList:
		log.Info("Enter DepositChaincode Contract " + modules.GetQuitApplyList + " Query")
		return d.GetQuitApplyList(stub)
		//  查看是否在退出列表中
	case modules.IsInQuitList:
		log.Info("Enter DepositChaincode Contract " + modules.IsInQuitList + " Query")
		if len(args) != 1 {
			return shim.Error("arg need one")
		}
		return d.IsInQuitList(stub,args[0])
		//  获取没收保证金申请列表
	case modules.GetListForForfeitureApplication:
		log.Info("Enter DepositChaincode Contract " + modules.GetListForForfeitureApplication + " Query")
		return d.GetListForForfeitureApplication(stub)
		//
	case modules.IsInForfeitureList:
		log.Info("Enter DepositChaincode Contract " + modules.IsInForfeitureList + " Query")
		if len(args) != 1 {
			return shim.Error("arg need one")
		}
		return d.IsInForfeitureList(stub,args[0])

		//  获取Mediator候选列表
	case modules.GetListForMediatorCandidate:
		log.Info("Enter DepositChaincode Contract " + modules.GetListForMediatorCandidate + " Query")
		return d.GetListForMediatorCandidate(stub)
		//  查看节点是否在候选列表中
	case modules.IsInMediatorCandidateList:
		log.Info("Enter DepositChaincode Contract " + modules.IsInMediatorCandidateList + " Query")
		if len(args) != 1 {
			return shim.Error("arg need one")
		}
		return d.IsInMediatorCandidateList(stub,args[0])
		//  获取Jury候选列表
	case modules.GetListForJuryCandidate:
		log.Info("Enter DepositChaincode Contract " + modules.GetListForJuryCandidate + " Query")
		return d.GetListForJuryCandidate(stub)
		//  查看jury是否在候选列表中
	case modules.IsInJuryCandidateList:
		log.Info("Enter DepositChaincode Contract " + modules.IsInJuryCandidateList + " Query")
		if len(args) != 1 {
			return shim.Error("arg need one")
		}
		return d.IsInJuryCandidateList(stub,args[0])
		//  获取Contract Developer候选列表
	case modules.GetListForDeveloper:
		log.Info("Enter DepositChaincode Contract " + modules.GetListForDeveloper + " Query")
		return d.GetListForDeveloper(stub)
		//  查看developer是否在候选列表中
	case modules.IsInDeveloperList:
		log.Info("Enter DepositChaincode Contract " + modules.IsInDeveloperList + " Query")
		return d.IsInDeveloperList(stub,args)
		//  获取jury/dev节点的账户
	case modules.GetDeposit:
		log.Info("Enter DepositChaincode Contract " + modules.GetDeposit + " Query")
		if len(args) != 1 {
			return shim.Error("arg need one")
		}
		return d.GetNodeBalance(stub,args[0])
	case modules.GetJuryDeposit:
		log.Info("Enter DepositChaincode Contract " + modules.GetJuryDeposit + " Query")
		if len(args) != 1 {
			return shim.Error("arg need one")
		}
		return d.GetJuryDeposit(stub,args[0])
		// 获取mediator Deposit
	case modules.GetMediatorDeposit:
		log.Info("Enter DepositChaincode Contract " + modules.GetMediatorDeposit + " Query")
		if len(args) != 1 {
			return shim.Error("arg need one")
		}
		return d.GetMediatorDeposit(stub,args[0])

	//  普通用户质押投票
	case modules.PledgeDeposit:
		log.Info("Enter DepositChaincode Contract " + modules.PledgeDeposit + " Invoke")
		return d.ProcessPledgeDeposit(stub)
	case modules.PledgeWithdraw: //提币质押申请（如果提币申请金额为MaxUint64表示全部提现）
		log.Info("Enter DepositChaincode Contract " + modules.PledgeWithdraw + " Invoke")
		if len(args) != 1 {
			return shim.Error("need 1 arg, withdraw Dao amount")
		}
		return d.ProcessPledgeWithdraw(stub, args[0])

	case modules.QueryPledgeStatusByAddr: //查询某用户的质押状态
		log.Info("Enter DepositChaincode Contract " + modules.QueryPledgeStatusByAddr + " Query")
		if len(args) != 1 {
			return shim.Error("need 1 arg, Address")
		}
		return d.QueryPledgeStatusByAddr(stub, args[0])
	case modules.QueryAllPledgeHistory: //查询质押分红历史
		log.Info("Enter DepositChaincode Contract " + modules.QueryAllPledgeHistory + " Query")
		return d.QueryAllPledgeHistory(stub)

	case modules.HandlePledgeReward: //质押分红处理
		log.Info("Enter DepositChaincode Contract " + modules.HandlePledgeReward + " Invoke")
		return d.HandlePledgeReward(stub)
	case modules.QueryPledgeList:
		log.Info("Enter DepositChaincode Contract " + modules.QueryPledgeList + " Query")
		return d.QueryPledgeList(stub)
	case modules.QueryPledgeListByDate:
		log.Info("Enter DepositChaincode Contract " + modules.QueryPledgeListByDate + " Query")
		if len(args) != 1 {
			return shim.Error("need 1 arg, Address")
		}
		return d.QueryPledgeListByDate(stub, args[0])
	case modules.QueryPledgeWithdraw:
		log.Info("Enter DepositChaincode Contract " + modules.QueryPledgeWithdraw + " Query")
		return d.QueryPledgeWithdraw(stub)
		//TODO Devin一个用户，怎么查看自己的流水账？
		//case AllPledgeVotes:
		//	b, err := getVotes(stub)
		//	if err != nil {
		//		return shim.Error(err.Error())
		//	}
		//	st := strconv.FormatInt(b, 10)
		//	return shim.Success([]byte(st))
	case modules.HandleMediatorInCandidateList:
		log.Info("Enter DepositChaincode Contract " + modules.HandleMediatorInCandidateList + " Invoke")
		return d.HandleMediatorInCandidateList(stub, args)
	case modules.HandleJuryInCandidateList:
		log.Info("Enter DepositChaincode Contract " + modules.HandleJuryInCandidateList + " Invoke")
		return d.HandleJuryInCandidateList(stub, args)
	case modules.HandleDevInList:
		log.Info("Enter DepositChaincode Contract " + modules.HandleDevInList + " Invoke")
		return d.HandleDevInList(stub, args)
	case modules.GetAllMediator:
		log.Info("Enter DepositChaincode Contract " + modules.GetAllMediator + " Query")
		return d.GetAllMediator(stub)
	case modules.GetAllNode:
		log.Info("Enter DepositChaincode Contract " + modules.GetAllNode + " Query")
		return d.GetAllNode(stub)
	case modules.GetAllJury:
		log.Info("Enter DepositChaincode Contract " + modules.GetAllJury + " Query")
		return d.GetAllJury(stub)
	}
	return shim.Error("please enter validate function name")
}

func (d *DepositChaincode) GetMediatorDeposit(stub shim.ChaincodeStubInterface,address string) pb.Response {
	mediator, err := getMediatorDeposit(stub, address)
	if err != nil {
		return shim.Error(err.Error())
	}
	if mediator == nil {
		return shim.Success([]byte("mediator is nil"))
	}
	mdJson := convertMediatorDeposit2Json(mediator)
	byte, err := json.Marshal(mdJson)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(byte)
}

func (d *DepositChaincode) GetJuryDeposit(stub shim.ChaincodeStubInterface,address string) pb.Response {
	balance, err := getJuryBalance(stub, address)
	if err != nil {
		return shim.Error(err.Error())
	}
	if balance == nil {
		return shim.Success([]byte("balance is nil"))
	}
	dbJson := convertJuryDeposit2Json(balance)
	byte, err := json.Marshal(dbJson)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(byte)
}

func (d *DepositChaincode) GetNodeBalance(stub shim.ChaincodeStubInterface,address string) pb.Response {
	balance, err := getNodeBalance(stub, address)
	if err != nil {
		return shim.Error(err.Error())
	}
	if balance == nil {
		return shim.Success([]byte("balance is nil"))
	}
	dbJson := convertDepositBalance2Json(balance)
	byte, err := json.Marshal(dbJson)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(byte)
}


func (d *DepositChaincode) IsInDeveloperList(stub shim.ChaincodeStubInterface,args []string) pb.Response {
	list, err := getList(stub, modules.DeveloperList)
	if err != nil {
		return shim.Error(err.Error())
	}
	if list == nil {
		return shim.Success([]byte("false"))
	}
	if len(args) != 1 {
		return shim.Error("arg need one")
	}
	if _, ok := list[args[0]]; ok {
		return shim.Success([]byte("true"))
	}
	return shim.Success([]byte("false"))
}

func (d *DepositChaincode) GetListForDeveloper(stub shim.ChaincodeStubInterface) pb.Response {
	list, err := stub.GetState(modules.DeveloperList)
	if err != nil {
		return shim.Error(err.Error())
	}
	if list == nil {
		return shim.Success([]byte("{}"))
	}
	return shim.Success(list)
}


func (d *DepositChaincode) IsInJuryCandidateList(stub shim.ChaincodeStubInterface,address string) pb.Response {
	list, err := getList(stub, modules.JuryList)
	if err != nil {
		return shim.Error(err.Error())
	}
	if list == nil {
		return shim.Success([]byte("false"))
	}

	if _, ok := list[address]; ok {
		return shim.Success([]byte("true"))
	}
	return shim.Success([]byte("false"))
}


func (d *DepositChaincode) GetListForJuryCandidate(stub shim.ChaincodeStubInterface) pb.Response {
	list, err := stub.GetState(modules.JuryList)
	if err != nil {
		return shim.Error(err.Error())
	}
	if list == nil {
		return shim.Success([]byte("{}"))
	}
	return shim.Success(list)
}

func (d *DepositChaincode) IsInMediatorCandidateList(stub shim.ChaincodeStubInterface,address string) pb.Response {
	list, err := getList(stub, modules.MediatorList)
	if err != nil {
		return shim.Error(err.Error())
	}
	if list == nil {
		return shim.Success([]byte("false"))
	}

	if _, ok := list[address]; ok {
		return shim.Success([]byte("true"))
	}
	return shim.Success([]byte("false"))
}

func (d *DepositChaincode) GetListForMediatorCandidate(stub shim.ChaincodeStubInterface) pb.Response {
	list, err := stub.GetState(modules.MediatorList)
	if err != nil {
		return shim.Error(err.Error())
	}
	if list == nil {
		return shim.Success([]byte("{}"))
	}
	return shim.Success(list)
}

	func (d *DepositChaincode) IsInForfeitureList(stub shim.ChaincodeStubInterface,address string) pb.Response {
	list, err := getListForForfeiture(stub)
	if err != nil {
		return shim.Error(err.Error())
	}
	if list == nil {
		return shim.Success([]byte("false"))
	}
	if _, ok := list[address]; ok {
		return shim.Success([]byte("true"))
	}
	return shim.Success([]byte("false"))
}


	func (d *DepositChaincode) GetListForForfeitureApplication(stub shim.ChaincodeStubInterface) pb.Response {
	list, err := stub.GetState(modules.ListForForfeiture)
	if err != nil {
		return shim.Error(err.Error())
	}
	if list == nil {
		return shim.Success([]byte("{}"))
	}
	return shim.Success(list)
}
	func (d *DepositChaincode) IsInQuitList(stub shim.ChaincodeStubInterface,address string) pb.Response {
	list, err := getListForQuit(stub)
	if err != nil {
		return shim.Error(err.Error())
	}
	if list == nil {
		return shim.Success([]byte("false"))
	}
	if _, ok := list[address]; ok {
		return shim.Success([]byte("true"))
	}
	return shim.Success([]byte("false"))
}

	func (d *DepositChaincode) GetQuitApplyList(stub shim.ChaincodeStubInterface) pb.Response {
	list, err := stub.GetState(modules.ListForQuit)
	if err != nil {
		return shim.Error(err.Error())
	}
	if list == nil {
		return shim.Success([]byte("{}"))
	}
	return shim.Success(list)
}

func (d *DepositChaincode) IsInAgreeList(stub shim.ChaincodeStubInterface, address string) pb.Response {
	list, err := getList(stub, modules.ListForAgreeBecomeMediator)
	if err != nil {
		return shim.Error(err.Error())
	}
	if list == nil {
		return shim.Success([]byte("false"))
	}

	if _, ok := list[address]; ok {
		return shim.Success([]byte("true"))
	}
	return shim.Success([]byte("false"))
}

func (d *DepositChaincode) GetAgreeForBecomeMediatorList(stub shim.ChaincodeStubInterface) pb.Response {
	list, err := stub.GetState(modules.ListForAgreeBecomeMediator)
	if err != nil {
		return shim.Error(err.Error())
	}
	if list == nil {
		return shim.Success([]byte("{}"))
	}
	return shim.Success(list)
}

func (d *DepositChaincode) IsInBecomeList(stub shim.ChaincodeStubInterface, address string) pb.Response {
	list, err := getList(stub, modules.ListForApplyBecomeMediator)
	if err != nil {
		return shim.Error(err.Error())
	}
	if list == nil {
		return shim.Success([]byte("false"))
	}

	if _, ok := list[address]; ok {
		return shim.Success([]byte("true"))
	}
	return shim.Success([]byte("false"))
}
func (d *DepositChaincode) GetBecomeMediatorApplyList(stub shim.ChaincodeStubInterface) pb.Response {
	list, err := stub.GetState(modules.ListForApplyBecomeMediator)
	if err != nil {
		return shim.Error(err.Error())
	}
	if list == nil {
		return shim.Success([]byte("{}"))
	}
	return shim.Success(list)
}

//  超级节点申请加入
func (d *DepositChaincode) ApplyBecomeMediator(stub shim.ChaincodeStubInterface,   mediatorCreateArgs string) pb.Response {
	return applyBecomeMediator(stub, mediatorCreateArgs)
}

//  超级节点交付保证金
func (d *DepositChaincode) MediatorPayToDepositContract(stub shim.ChaincodeStubInterface) pb.Response {
	return mediatorPayToDepositContract(stub)
}

//  超级节点申请退出候选列表
func (d *DepositChaincode) MediatorApplyQuit(stub shim.ChaincodeStubInterface) pb.Response {
	return mediatorApplyQuit(stub)
}

//  超级节点更新信息
func (d *DepositChaincode) UpdateMediatorInfo(stub shim.ChaincodeStubInterface, mediatorUpdateArgs string) pb.Response {
	return updateMediatorInfo(stub, mediatorUpdateArgs)
}

//  陪审员交付保证金
func (d *DepositChaincode) JuryPayToDepositContract(stub shim.ChaincodeStubInterface, pubkey string) pb.Response {
	return juryPayToDepositContract(stub, pubkey)
}

//  陪审员申请退出候选列表
func (d *DepositChaincode) JuryApplyQuit(stub shim.ChaincodeStubInterface) pb.Response {
	return juryApplyQuit(stub)
}

//  开发者交付保证金
func (d *DepositChaincode) DeveloperPayToDepositContract(stub shim.ChaincodeStubInterface) pb.Response {
	return developerPayToDepositContract(stub)
}

//  开发者申请退出列表
func (d *DepositChaincode) DevApplyQuit(stub shim.ChaincodeStubInterface) pb.Response {
	return devApplyQuit(stub)
}

//  基金会对申请加入Mediator进行处理
func (d *DepositChaincode) HandleForApplyBecomeMediator(stub shim.ChaincodeStubInterface, address string, okOrNo string) pb.Response {
	return handleForApplyBecomeMediator(stub, address,okOrNo)
}

//  基金会对申请退出Mediator进行处理
func (d *DepositChaincode) HandleForApplyQuitMediator(stub shim.ChaincodeStubInterface, address string, okOrNo string) pb.Response {
	return handleForApplyQuitMediator(stub,  address,okOrNo)
}

//  处理陪审员申请退出候选列表
func (d *DepositChaincode) HandleForApplyQuitJury(stub shim.ChaincodeStubInterface, address string ,okOrNo string) pb.Response {
	return handleForApplyQuitJury(stub, address,okOrNo)
}

//  处理开发者申请退出列表
func (d *DepositChaincode) HandleForApplyQuitDev(stub shim.ChaincodeStubInterface, address string ,okOrNo string) pb.Response {
	return handleForApplyQuitDev(stub, address,okOrNo)
}

//  处理没收节点
func (d *DepositChaincode) HandleForForfeitureApplication(stub shim.ChaincodeStubInterface, address string,okOrNo string) pb.Response {
	return handleForForfeitureApplication(stub, address,okOrNo)
}

//  移除超级节点同意列表
func (d DepositChaincode) HandleNodeRemoveFromAgreeList(stub shim.ChaincodeStubInterface, address string) pb.Response {
	return hanldeNodeRemoveFromAgreeList(stub, address)
}

//  申请没收节点保证金
func (d DepositChaincode) ApplyForForfeitureDeposit(stub shim.ChaincodeStubInterface, forfeitureAddress string,role string,reason string) pb.Response {
	return applyForForfeitureDeposit(stub, forfeitureAddress,role,reason)
}

//  质押

func (d DepositChaincode) ProcessPledgeDeposit(stub shim.ChaincodeStubInterface) pb.Response {
	return processPledgeDeposit(stub)
}

func (d DepositChaincode) ProcessPledgeWithdraw(stub shim.ChaincodeStubInterface, amount string) pb.Response {
	return processPledgeWithdraw(stub, amount)
}

func (d DepositChaincode) HandlePledgeReward(stub shim.ChaincodeStubInterface) pb.Response {
	return handlePledgeReward(stub)
}

func (d DepositChaincode) QueryPledgeStatusByAddr(stub shim.ChaincodeStubInterface, address string) pb.Response {
	return queryPledgeStatusByAddr(stub, address)
}

func (d DepositChaincode) QueryAllPledgeHistory(stub shim.ChaincodeStubInterface) pb.Response {
	return queryAllPledgeHistory(stub)
}

func (d DepositChaincode) QueryPledgeList(stub shim.ChaincodeStubInterface) pb.Response {
	return queryPledgeList(stub)
}


func (d DepositChaincode) QueryPledgeListByDate(stub shim.ChaincodeStubInterface, date string) pb.Response {
	return queryPledgeListByDate(stub,date)
}



func (d DepositChaincode) QueryPledgeWithdraw(stub shim.ChaincodeStubInterface) pb.Response {
	list, err := getAllPledgeWithdrawRecords(stub)
	if err != nil {
		return shim.Error(err.Error())
	}
	result, _ := json.Marshal(list)
	return shim.Success(result)
}
//  质押

//  移除超级节点候选列表
func (d DepositChaincode) HandleMediatorInCandidateList(stub shim.ChaincodeStubInterface, addresses []string) pb.Response {
	return handleNodeInList(stub, addresses, modules.Mediator)
}

//  移除陪审员候选列表
func (d DepositChaincode) HandleJuryInCandidateList(stub shim.ChaincodeStubInterface, addresses []string) pb.Response {
	return handleNodeInList(stub, addresses, modules.Jury)
}

//  移除开发者列表
func (d DepositChaincode) HandleDevInList(stub shim.ChaincodeStubInterface, addresses []string) pb.Response {
	return handleNodeInList(stub, addresses, modules.Developer)
}


func (d DepositChaincode) GetAllMediator(stub shim.ChaincodeStubInterface) pb.Response {
	values, err := stub.GetStateByPrefix(string(constants.MEDIATOR_INFO_PREFIX) +
		string(constants.DEPOSIT_BALANCE_PREFIX))
	if err != nil {
		log.Debugf("stub.GetStateByPrefix error: %s", err.Error())
		return shim.Error(err.Error())
	}
	if len(values) > 0 {
		mediators := make(map[string]*modules.MediatorDeposit)
		for _, v := range values {
			m := modules.MediatorDeposit{}
			err := json.Unmarshal(v.Value, &m)
			if err != nil {
				log.Debugf("json.Unmarshal error: %s", err.Error())
				return shim.Error(err.Error())
			}
			mediators[v.Key] = &m
		}
		bytes, err := json.Marshal(mediators)
		if err != nil {
			log.Debugf("json.Marshal error: %s", err.Error())
			return shim.Error(err.Error())
		}
		return shim.Success(bytes)
	}
	return shim.Success([]byte("{}"))
}
func (d DepositChaincode) GetAllNode(stub shim.ChaincodeStubInterface) pb.Response {
	values, err := stub.GetStateByPrefix(string(constants.DEPOSIT_BALANCE_PREFIX))
	if err != nil {
		log.Debugf("stub.GetStateByPrefix error: %s", err.Error())
		return shim.Error(err.Error())
	}
	if len(values) > 0 {
		node := make(map[string]*modules.DepositBalance)
		for _, v := range values {
			n := modules.DepositBalance{}
			err := json.Unmarshal(v.Value, &n)
			if err != nil {
				log.Debugf("json.Unmarshal error: %s", err.Error())
				return shim.Error(err.Error())
			}
			node[v.Key] = &n
		}
		bytes, err := json.Marshal(node)
		if err != nil {
			log.Debugf("json.Marshal error: %s", err.Error())
			return shim.Error(err.Error())
		}
		return shim.Success(bytes)
	}
	return shim.Success([]byte("{}"))
}

func (d DepositChaincode) GetAllJury(stub shim.ChaincodeStubInterface) pb.Response {
	listb, err := stub.GetState(modules.JuryList)
	if err != nil {
		return shim.Error(err.Error())
	}
	if listb == nil {
		return shim.Success([]byte("{}"))
	}
	allJurorAddrs := make(map[string]bool)
	err = json.Unmarshal(listb, &allJurorAddrs)
	if err != nil {
		return shim.Error(err.Error())
	}
	jurynodes := make(map[string]*modules.JurorDeposit)
	for a := range allJurorAddrs {
		j, err := stub.GetState(string(constants.DEPOSIT_JURY_BALANCE_PREFIX) + a)
		if err != nil {
			return shim.Error(err.Error())
		}
		juror := modules.JurorDeposit{}
		err = json.Unmarshal(j, &juror)
		if err != nil {
			shim.Error(err.Error())
		}
		jurynodes[a] = &juror
	}
	juryb, err := json.Marshal(jurynodes)
	if err != nil {
		shim.Error(err.Error())
	}
	return shim.Success(juryb)
}
//
//func (d DepositChaincode) handleRemoveMediatorNode(stub shim.ChaincodeStubInterface, args []string) pb.Response {
//	return handleRemoveMediatorNode(stub, args)
//}
//
////
//func (d DepositChaincode) handleRemoveNormalNode(stub shim.ChaincodeStubInterface, args []string) pb.Response {
//	return handleRemoveNormalNode(stub, args)
//}

//  更新陪审员信息
//func (d DepositChaincode) updateJuryInfo(stub shim.ChaincodeStubInterface, args []string) pb.Response {
//	return updateJuryInfo(stub, args)
//}
