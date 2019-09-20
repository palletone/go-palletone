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
	return shim.Success([]byte("init ok"))
}

func (d *DepositChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	funcName, args := stub.GetFunctionAndParameters()
	switch funcName {
	//
	// 申请成为Mediator
	case modules.ApplyMediator:
		log.Info("Enter DepositChaincode Contract " + modules.ApplyMediator + " Invoke")
		return d.applyBecomeMediator(stub, args)
	// mediator 交付保证金
	case modules.MediatorPayDeposit:
		log.Info("Enter DepositChaincode Contract " + modules.MediatorPayDeposit + " Invoke")
		return d.mediatorPayToDepositContract(stub)
	// 申请退出Mediator
	case modules.MediatorApplyQuit:
		log.Info("Enter DepositChaincode Contract " + modules.MediatorApplyQuit + " Invoke")
		return d.mediatorApplyQuit(stub)
	// 更新 Mediator 信息
	case modules.UpdateMediatorInfo:
		log.Info("Enter DepositChaincode Contract " + modules.UpdateMediatorInfo + " Invoke")
		return d.updateMediatorInfo(stub, args)
	//
	//  jury 交付保证金
	case modules.JuryPayToDepositContract:
		log.Info("Enter DepositChaincode Contract " + modules.JuryPayToDepositContract + " Invoke")
		return d.juryPayToDepositContract(stub, args)
		//  jury 申请退出
	case modules.JuryApplyQuit:
		log.Info("Enter DepositChaincode Contract " + modules.JuryApplyQuit + " Invoke")
		return d.juryApplyQuit(stub)
	//
	//  developer 交付保证金
	case modules.DeveloperPayToDepositContract:
		log.Info("Enter DepositChaincode Contract " + modules.DeveloperPayToDepositContract + " Invoke")
		return d.developerPayToDepositContract(stub)
		//  developer 申请退出
	case modules.DeveloperApplyQuit:
		log.Info("Enter DepositChaincode Contract " + modules.DeveloperApplyQuit + " Invoke")
		return d.devApplyQuit(stub)
	//
	//  基金会对加入申请Mediator进行处理
	case modules.HandleForApplyBecomeMediator:
		log.Info("Enter DepositChaincode Contract " + modules.HandleForApplyBecomeMediator + " Invoke")
		return d.handleForApplyBecomeMediator(stub, args)
	//  基金会移除某个节点
	case modules.HanldeNodeRemoveFromAgreeList:
		log.Info("Enter DepositChaincode Contract " + modules.HanldeNodeRemoveFromAgreeList + " Invoke")
		return d.handleNodeRemoveFromAgreeList(stub, args)
		//  基金会对退出申请Mediator进行处理
	case modules.HandleForApplyQuitMediator:
		log.Info("Enter DepositChaincode Contract " + modules.HandleForApplyQuitMediator + " Invoke")
		return d.handleForApplyQuitMediator(stub, args)
		//  基金会对退出申请Jury进行处理
	case modules.HandleForApplyQuitJury:
		log.Info("Enter DepositChaincode Contract " + modules.HandleForApplyQuitJury + " Invoke")
		return d.handleForApplyQuitJury(stub, args)
		//  基金会对退出申请Developer进行处理
	case modules.HandleForApplyQuitDev:
		log.Info("Enter DepositChaincode Contract " + modules.HandleForApplyQuitDev + " Invoke")
		return d.handleForApplyQuitDev(stub, args)
		//  基金会对申请没收做相应的处理
	case modules.HandleForForfeitureApplication:
		log.Info("Enter DepositChaincode Contract " + modules.HandleForForfeitureApplication + " Invoke")
		return d.handleForForfeitureApplication(stub, args)
	//
	//  申请保证金没收
	case modules.ApplyForForfeitureDeposit:
		log.Info("Enter DepositChaincode Contract " + modules.ApplyForForfeitureDeposit + " Invoke")
		return d.applyForForfeitureDeposit(stub, args)
	//
	//  获取Mediator申请加入列表
	case modules.GetBecomeMediatorApplyList:
		log.Info("Enter DepositChaincode Contract " + modules.GetBecomeMediatorApplyList + " Query")
		list, err := stub.GetState(modules.ListForApplyBecomeMediator)
		if err != nil {
			return shim.Error(err.Error())
		}
		if list == nil {
			return shim.Success([]byte("{}"))
		}
		return shim.Success(list)
		//  查看是否在become列表中
	case modules.IsInBecomeList:
		log.Info("Enter DepositChaincode Contract " + modules.IsInBecomeList + " Query")
		list, err := getList(stub, modules.ListForApplyBecomeMediator)
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
		//  获取已同意的mediator列表
	case modules.GetAgreeForBecomeMediatorList:
		log.Info("Enter DepositChaincode Contract " + modules.GetAgreeForBecomeMediatorList + " Query")
		list, err := stub.GetState(modules.ListForAgreeBecomeMediator)
		if err != nil {
			return shim.Error(err.Error())
		}
		if list == nil {
			return shim.Success([]byte("{}"))
		}
		return shim.Success(list)
		//  查看是否在agree列表中
	case modules.IsApproved:
		log.Info("Enter DepositChaincode Contract " + modules.IsApproved + " Query")
		list, err := getList(stub, modules.ListForAgreeBecomeMediator)
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
		//获取申请退出列表
	case modules.GetQuitApplyList:
		log.Info("Enter DepositChaincode Contract " + modules.GetQuitApplyList + " Query")
		list, err := stub.GetState(modules.ListForQuit)
		if err != nil {
			return shim.Error(err.Error())
		}
		if list == nil {
			return shim.Success([]byte("{}"))
		}
		return shim.Success(list)
		//  查看是否在退出列表中
	case modules.IsInQuitList:
		log.Info("Enter DepositChaincode Contract " + modules.IsInQuitList + " Query")
		list, err := GetListForQuit(stub)
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
		//  获取没收保证金申请列表
	case modules.GetListForForfeitureApplication:
		log.Info("Enter DepositChaincode Contract " + modules.GetListForForfeitureApplication + " Query")
		list, err := stub.GetState(modules.ListForForfeiture)
		if err != nil {
			return shim.Error(err.Error())
		}
		if list == nil {
			return shim.Success([]byte("{}"))
		}
		return shim.Success(list)
		//
	case modules.IsInForfeitureList:
		log.Info("Enter DepositChaincode Contract " + modules.IsInForfeitureList + " Query")
		list, err := GetListForForfeiture(stub)
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

		//  获取Mediator候选列表
	case modules.GetListForMediatorCandidate:
		log.Info("Enter DepositChaincode Contract " + modules.GetListForMediatorCandidate + " Query")
		list, err := stub.GetState(modules.MediatorList)
		if err != nil {
			return shim.Error(err.Error())
		}
		if list == nil {
			return shim.Success([]byte("{}"))
		}
		return shim.Success(list)
		//  查看节点是否在候选列表中
	case modules.IsInMediatorCandidateList:
		log.Info("Enter DepositChaincode Contract " + modules.IsInMediatorCandidateList + " Query")
		list, err := getList(stub, modules.MediatorList)
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
		//  获取Jury候选列表
	case modules.GetListForJuryCandidate:
		log.Info("Enter DepositChaincode Contract " + modules.GetListForJuryCandidate + " Query")
		list, err := stub.GetState(modules.JuryList)
		if err != nil {
			return shim.Error(err.Error())
		}
		if list == nil {
			return shim.Success([]byte("{}"))
		}
		return shim.Success(list)
		//  查看jury是否在候选列表中
	case modules.IsInJuryCandidateList:
		log.Info("Enter DepositChaincode Contract " + modules.IsInJuryCandidateList + " Query")
		list, err := getList(stub, modules.JuryList)
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
		//  获取Contract Developer候选列表
	case modules.GetListForDeveloper:
		log.Info("Enter DepositChaincode Contract " + modules.GetListForDeveloper + " Query")
		list, err := stub.GetState(modules.DeveloperList)
		if err != nil {
			return shim.Error(err.Error())
		}
		if list == nil {
			return shim.Success([]byte("{}"))
		}
		return shim.Success(list)
		//  查看developer是否在候选列表中
	case modules.IsInDeveloperList:
		log.Info("Enter DepositChaincode Contract " + modules.IsInDeveloperList + " Query")
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
		//  获取jury/dev节点的账户
	case modules.GetDeposit:
		log.Info("Enter DepositChaincode Contract " + modules.GetDeposit + " Query")
		balance, err := GetNodeBalance(stub, args[0])
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
	case modules.GetJuryDeposit:
		log.Info("Enter DepositChaincode Contract " + modules.GetJuryDeposit + " Query")
		balance, err := GetJuryBalance(stub, args[0])
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
		// 获取mediator Deposit
	case modules.GetMediatorDeposit:
		log.Info("Enter DepositChaincode Contract " + modules.GetMediatorDeposit + " Query")
		mediator, err := GetMediatorDeposit(stub, args[0])
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

	//  普通用户质押投票
	case modules.PledgeDeposit:
		log.Info("Enter DepositChaincode Contract " + modules.PledgeDeposit + " Invoke")
		return d.processPledgeDeposit(stub)
	case modules.PledgeWithdraw: //提币质押申请（如果提币申请金额为MaxUint64表示全部提现）
		log.Info("Enter DepositChaincode Contract " + modules.PledgeWithdraw + " Invoke")
		return d.processPledgeWithdraw(stub, args)

	case modules.QueryPledgeStatusByAddr: //查询某用户的质押状态
		log.Info("Enter DepositChaincode Contract " + modules.QueryPledgeStatusByAddr + " Query")
		return queryPledgeStatusByAddr(stub, args)
	case modules.QueryAllPledgeHistory: //查询质押分红历史
		log.Info("Enter DepositChaincode Contract " + modules.QueryAllPledgeHistory + " Query")
		return queryAllPledgeHistory(stub)

	case modules.HandlePledgeReward: //质押分红处理
		log.Info("Enter DepositChaincode Contract " + modules.HandlePledgeReward + " Invoke")
		return d.handlePledgeReward(stub, args)
	case modules.QueryPledgeList:
		log.Info("Enter DepositChaincode Contract " + modules.QueryPledgeList + " Query")
		return queryPledgeList(stub)
	case modules.QueryPledgeListByDate:
		log.Info("Enter DepositChaincode Contract " + modules.QueryPledgeListByDate + " Query")
		return queryPledgeListByDate(stub, args)
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
		return d.handleMediatorInCandidateList(stub, args)
	case modules.HandleJuryInCandidateList:
		log.Info("Enter DepositChaincode Contract " + modules.HandleJuryInCandidateList + " Invoke")
		return d.handleJuryInCandidateList(stub, args)
	case modules.HandleDevInList:
		log.Info("Enter DepositChaincode Contract " + modules.HandleDevInList + " Invoke")
		return d.handleDevInList(stub, args)
	case modules.GetAllMediator:
		log.Info("Enter DepositChaincode Contract " + modules.GetAllMediator + " Query")
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
	case modules.GetAllNode:
		log.Info("Enter DepositChaincode Contract " + modules.GetAllNode + " Query")
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
	case modules.GetAllJury:
		log.Info("Enter DepositChaincode Contract " + modules.GetAllJury + " Query")
		juryvalues, err := stub.GetStateByPrefix(string(constants.DEPOSIT_JURY_BALANCE_PREFIX))
		if err != nil {
			return shim.Error(err.Error())
		}
		if len(juryvalues) > 0 {
			jurynode := make(map[string]*modules.Juror)
			for _, v := range juryvalues {
				n := modules.Juror{}
				err := json.Unmarshal(v.Value, &n)
				if err != nil {
					return shim.Error(err.Error())
				}
				jurynode[v.Key] = &n
			}
			bytes, err := json.Marshal(jurynode)
			if err != nil {
				return shim.Error(err.Error())
			}
			return shim.Success(bytes)
		}
		return shim.Success([]byte("{}"))
	case modules.UpdateJuryInfo:
		log.Info("Enter DepositChaincode Contract " + modules.UpdateJuryInfo + " Invoke")
		return d.updateJuryInfo(stub, args)
	}
	return shim.Error("please enter validate function name")
}

//  超级节点申请加入
func (d *DepositChaincode) applyBecomeMediator(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	return applyBecomeMediator(stub, args)
}

//  超级节点交付保证金
func (d *DepositChaincode) mediatorPayToDepositContract(stub shim.ChaincodeStubInterface) pb.Response {
	return mediatorPayToDepositContract(stub /*, args*/)
}

//  超级节点申请退出候选列表
func (d *DepositChaincode) mediatorApplyQuit(stub shim.ChaincodeStubInterface) pb.Response {
	return mediatorApplyQuit(stub)
}

//  超级节点更新信息
func (d *DepositChaincode) updateMediatorInfo(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	return updateMediatorInfo(stub, args)
}

//  陪审员交付保证金
func (d *DepositChaincode) juryPayToDepositContract(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	return juryPayToDepositContract(stub, args)
}

//  陪审员申请退出候选列表
func (d *DepositChaincode) juryApplyQuit(stub shim.ChaincodeStubInterface) pb.Response {
	return juryApplyQuit(stub)
}

//  开发者交付保证金
func (d *DepositChaincode) developerPayToDepositContract(stub shim.ChaincodeStubInterface) pb.Response {
	return developerPayToDepositContract(stub)
}

//  开发者申请退出列表
func (d *DepositChaincode) devApplyQuit(stub shim.ChaincodeStubInterface) pb.Response {
	return devApplyQuit(stub)
}

//  基金会对申请加入Mediator进行处理
func (d *DepositChaincode) handleForApplyBecomeMediator(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	return handleForApplyBecomeMediator(stub, args)
}

//  基金会对申请退出Mediator进行处理
func (d *DepositChaincode) handleForApplyQuitMediator(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	return handleForApplyQuitMediator(stub, args)
}

//  处理陪审员申请退出候选列表
func (d *DepositChaincode) handleForApplyQuitJury(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	return handleForApplyQuitJury(stub, args)
}

//  处理开发者申请退出列表
func (d *DepositChaincode) handleForApplyQuitDev(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	return handleForApplyQuitDev(stub, args)
}

//  处理没收节点
func (d *DepositChaincode) handleForForfeitureApplication(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	return handleForForfeitureApplication(stub, args)
}

//  移除超级节点同意列表
func (d DepositChaincode) handleNodeRemoveFromAgreeList(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	return hanldeNodeRemoveFromAgreeList(stub, args)
}

//  申请没收节点保证金
func (d DepositChaincode) applyForForfeitureDeposit(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	return applyForForfeitureDeposit(stub, args)
}

//  质押

func (d DepositChaincode) processPledgeDeposit(stub shim.ChaincodeStubInterface) pb.Response {
	return processPledgeDeposit(stub)
}

func (d DepositChaincode) processPledgeWithdraw(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	return processPledgeWithdraw(stub, args)
}

func (d DepositChaincode) handlePledgeReward(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	return handlePledgeReward(stub, args)
}

//  质押

//  移除超级节点候选列表
func (d DepositChaincode) handleMediatorInCandidateList(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	return handleNodeInList(stub, args, modules.Mediator)
}

//  移除陪审员候选列表
func (d DepositChaincode) handleJuryInCandidateList(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	return handleNodeInList(stub, args, modules.Jury)
}

//  移除开发者列表
func (d DepositChaincode) handleDevInList(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	return handleNodeInList(stub, args, modules.Developer)
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
func (d DepositChaincode) updateJuryInfo(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	return updateJuryInfo(stub, args)
}
