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
	"encoding/json"
	"fmt"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/contracts/shim"
	pb "github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
	"github.com/palletone/go-palletone/dag/modules"
)

//同意申请没收请求
func agreeForApplyForfeiture(stub shim.ChaincodeStubInterface, foundationA string, forfeitureNode *Forfeiture) error {
	log.Info("Start entering agreeForApplyForfeiture func.")
	//判断节点类型
	switch {
	case forfeitureNode.ForfeitureRole == Mediator:
		return handleMediatorForfeitureDeposit(stub, foundationA, forfeitureNode)
	case forfeitureNode.ForfeitureRole == Jury:
		return handleJuryForfeitureDeposit(stub, foundationA, forfeitureNode)
	case forfeitureNode.ForfeitureRole == Developer:
		return handleDevForfeitureDeposit(stub, foundationA, forfeitureNode)
	default:
		return fmt.Errorf("please enter validate role.")
	}
}

//处理没收Mediator保证金
func handleMediatorForfeitureDeposit(stub shim.ChaincodeStubInterface, foundationA string, forfeiture *Forfeiture) error {
	//  获取mediator
	md, err := GetMediatorDeposit(stub, forfeiture.ForfeitureAddress)
	if err != nil {
		return err
	}
	if md == nil {
		return fmt.Errorf("node is nil")
	}
	//  计算余额
	result := md.Balance - forfeiture.ApplyTokens.Amount
	//  判断是否需要移除列表
	//  获取保证金下线，在状态数据库中
	//depositAmountsForMediatorStr, err := stub.GetSystemConfig(modules.DepositAmountForMediator)
	//if err != nil {
	//	log.Error("get deposit amount for mediator err: ", "error", err)
	//	return err
	//}
	////  转换保证金数量
	//depositAmountsForMediator, err := strconv.ParseUint(depositAmountsForMediatorStr, 10, 64)

	cp, err := stub.GetSystemConfig()
	if err != nil {
		//log.Error("strconv.ParseUint err:", "error", err)
		return err
	}
	depositAmountsForMediator := cp.DepositAmountForMediator
	//  需要移除列表forfeiture.ApplyTokens.Amount
	if result < depositAmountsForMediator {
		//  调用从合约把token转到请求地址
		forfeiture.ApplyTokens.Amount = md.Balance
		//  没收到基金会地址
		err := stub.PayOutToken(foundationA, forfeiture.ApplyTokens, 0)
		if err != nil {
			log.Error("Stub.PayOutToken err:", "error", err)
			return err
		}
		//  移除列表
		err = moveCandidate(modules.MediatorList, forfeiture.ForfeitureAddress, stub)
		if err != nil {
			log.Error("MoveCandidate err:", "error", err)
			return err
		}
		err = moveCandidate(modules.JuryList, forfeiture.ForfeitureAddress, stub)
		if err != nil {
			log.Error("MoveCandidate err:", "error", err)
			return err
		}
		//  更新
		md.Status = Quited
		md.Balance = 0
		md.EnterTime = ""
		md.LastModifyTime = ""
	} else {
		//  没收到基金会地址
		err := stub.PayOutToken(foundationA, forfeiture.ApplyTokens, 0)
		if err != nil {
			log.Error("Stub.PayOutToken err:", "error", err)
			return err
		}
		md.Balance -= forfeiture.ApplyTokens.Amount
	}
	//  保存
	err = SaveMediatorDeposit(stub, forfeiture.ForfeitureAddress, md)
	if err != nil {
		return err
	}
	return nil
}

func handleJuryForfeitureDeposit(stub shim.ChaincodeStubInterface, foundationA string, forfeiture *Forfeiture) error {
	node, err := GetNodeBalance(stub, forfeiture.ForfeitureAddress)
	if err != nil {
		return err
	}
	if node == nil {
		return fmt.Errorf("node is nil")
	}
	//depositAmountsForJuryStr, err := stub.GetSystemConfig(DepositAmountForJury)
	//if err != nil {
	//	log.Error("Stub.GetSystemConfig with DepositAmountForJury err:", "error", err)
	//	return err
	//}
	////转换
	//depositAmountsForJury, err := strconv.ParseUint(depositAmountsForJuryStr, 10, 64)
	//if err != nil {
	//	log.Error("Strconv.ParseUint err:", "error", err)
	//	return err
	//}

	cp, err := stub.GetSystemConfig()
	if err != nil {
		//log.Error("strconv.ParseUint err:", "error", err)
		return err
	}
	depositAmountsForJury := cp.DepositAmountForJury
	//计算余额
	result := node.Balance - forfeiture.ApplyTokens.Amount
	//判断是否没收全部，即在列表中移除该节点
	if result < depositAmountsForJury {
		//  移除列表
		err = moveCandidate(modules.JuryList, forfeiture.ForfeitureAddress, stub)
		if err != nil {
			return err
		}
		//  如果没收全部，则删除该节点
		if result == 0 {
			err := stub.PayOutToken(foundationA, forfeiture.ApplyTokens, 0)
			if err != nil {
				log.Error("Stub.PayOutToken err:", "error", err)
				return err
			}
			err = DelNodeBalance(stub, forfeiture.ForfeitureAddress)
			if err != nil {
				return err
			}
			return nil
		}
		//  更新
		node.EnterTime = ""
	} else {
	}
	node.Balance -= forfeiture.ApplyTokens.Amount
	node.LastModifyTime = TimeStr()
	err = stub.PayOutToken(foundationA, forfeiture.ApplyTokens, 0)
	if err != nil {
		log.Error("Stub.PayOutToken err:", "error", err)
		return err
	}
	err = SaveNodeBalance(stub, forfeiture.ForfeitureAddress, node)
	if err != nil {
		return err
	}
	return nil
}

func handleDevForfeitureDeposit(stub shim.ChaincodeStubInterface, foundationA string, forfeiture *Forfeiture) error {
	node, err := GetNodeBalance(stub, forfeiture.ForfeitureAddress)
	if err != nil {
		return err
	}
	if node == nil {
		return fmt.Errorf("node is nil")
	}
	//depositAmountsForDevStr, err := stub.GetSystemConfig(DepositAmountForDeveloper)
	//if err != nil {
	//	log.Error("Stub.GetSystemConfig with DepositAmountForJury err:", "error", err)
	//	return err
	//}
	////转换
	//depositAmountsForDev, err := strconv.ParseUint(depositAmountsForDevStr, 10, 64)
	//if err != nil {
	//	log.Error("Strconv.ParseUint err:", "error", err)
	//	return err
	//}

	cp, err := stub.GetSystemConfig()
	if err != nil {
		//log.Error("strconv.ParseUint err:", "error", err)
		return err
	}
	depositAmountsForDev := cp.DepositAmountForDeveloper
	//计算余额
	result := node.Balance - forfeiture.ApplyTokens.Amount
	//判断是否没收全部，即在列表中移除该节点
	if result < depositAmountsForDev {
		//  移除列表
		err = moveCandidate(modules.DeveloperList, forfeiture.ForfeitureAddress, stub)
		if err != nil {
			return err
		}
		//  如果没收全部，则删除该节点
		if result == 0 {
			err := stub.PayOutToken(foundationA, forfeiture.ApplyTokens, 0)
			if err != nil {
				log.Error("Stub.PayOutToken err:", "error", err)
				return err
			}
			err = DelNodeBalance(stub, forfeiture.ForfeitureAddress)
			if err != nil {
				return err
			}
			return nil
		}
		//  更新
		node.EnterTime = ""
	} else {
	}
	node.Balance -= forfeiture.ApplyTokens.Amount
	node.LastModifyTime = TimeStr()
	err = stub.PayOutToken(foundationA, forfeiture.ApplyTokens, 0)
	if err != nil {
		log.Error("Stub.PayOutToken err:", "error", err)
		return err
	}
	err = SaveNodeBalance(stub, forfeiture.ForfeitureAddress, node)
	if err != nil {
		return err
	}
	return nil
}

func handleForForfeitureApplication(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	log.Info("handleForForfeitureApplication")
	//  地址，是否同意
	if len(args) != 2 {
		log.Error("args need two parameters.")
		return shim.Error("args need two parameters.")
	}
	//  判断是否基金会发起的
	if !isFoundationInvoke(stub) {
		log.Error("please use foundation address")
		return shim.Error("please use foundation address")
	}
	//  获取基金会地址
	invokeAddr, err := stub.GetInvokeAddress()
	if err != nil {
		log.Error("get invoke address err: ", "error", err)
		return shim.Error(err.Error())
	}
	//  处理没收地址
	addr := args[0]
	//  判断没收地址是否正确
	f, err := common.StringToAddress(addr)
	if err != nil {
		return shim.Error(err.Error())
	}
	//  需要判断是否在列表
	listForForfeiture, err := GetListForForfeiture(stub)
	if err != nil {
		return shim.Error(err.Error())
	}
	//
	if listForForfeiture == nil {
		return shim.Error("was not in the list")
	} else {
		//
		if _, ok := listForForfeiture[f.String()]; !ok {
			return shim.Error("node was not in the forfeiture list")
		}
	}
	//获取节点信息
	forfeitureNode := listForForfeiture[addr]
	//  处理操作ok or no
	isOk := args[1]
	//check 如果为ok，则同意此申请，如果为no，则不同意此申请
	if isOk == Ok {
		err = agreeForApplyForfeiture(stub, invokeAddr.String(), forfeitureNode)
		if err != nil {
			return shim.Error(err.Error())
		}
	} else if isOk == No {
		//移除申请列表，不做处理
		log.Info("not agree to for apply forfeiture")
	} else {
		log.Error("Please enter ok or no.")
		return shim.Error("Please enter ok or no.")
	}
	//  不管同意与否都需要从列表中移除
	delete(listForForfeiture, addr)
	listForForfeitureByte, err := json.Marshal(listForForfeiture)
	if err != nil {
		log.Error("Json.Marshal err:", "error", err)
		return shim.Error(err.Error())
	}
	//更新列表
	err = stub.PutState(ListForForfeiture, listForForfeitureByte)
	if err != nil {
		log.Error("Stub.PutState err:", "error", err)
		return shim.Error(err.Error())
	}
	return shim.Success([]byte(nil))
}

//基金会处理
func handleForDeveloperApplyCashback(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	log.Info("handleForDeveloperApplyCashback")
	//  地址，申请时间，是否同意
	if len(args) != 2 {
		log.Error("args need two parameters")
		return shim.Error("args need two parameters")
	}
	//  判断是否基金会发起的
	if !isFoundationInvoke(stub) {
		log.Error("please use foundation address")
		return shim.Error("please use foundation address")
	}
	//  获取一下该用户下的账簿情况
	addr, err := common.StringToAddress(args[0])
	if err != nil {
		log.Error("common.StringToAddress err: ", "error", err)
		return shim.Error(err.Error())
	}
	balance, err := GetNodeBalance(stub, addr.String())
	if err != nil {
		log.Error("Stub.GetDepositBalance err: ", "error", err)
		return shim.Error(err.Error())
	}
	//  判断没收节点账户是否为空
	if balance == nil {
		log.Error("balance is nil")
		return shim.Error("balance is nil")
	}
	//  获取请求列表
	listForCashback, err := GetListForCashback(stub)
	if err != nil {
		log.Error("Stub.GetListForCashback err:", "error", err)
		return shim.Error(err.Error())
	}
	if listForCashback == nil {
		log.Error("listForCashback is nil.")
		return shim.Error("listForCashback is nil.")
	}
	if _, ok := listForCashback[addr.String()]; !ok {
		log.Error("node is not exist in the list.")
		return shim.Error("node is not exist in the list.")
	}
	cashbackNode := listForCashback[addr.String()]
	delete(listForCashback, addr.String())
	isOk := args[1]
	if isOk == Ok {
		//  对余额处理
		err := handleDeveloperDepositCashback(stub, addr, cashbackNode, balance)
		if err != nil {
			log.Error("HandleJuryDepositCashback err:", "error", err)
			return shim.Error(err.Error())
		}
	} else if isOk == No {
		log.Info("does not agree")
	} else {
		log.Error("please enter ok or no.")
		return shim.Error("please enter ok or no.")
	}
	err = SaveListForCashback(stub, listForCashback)
	if err != nil {
		log.Error("Stub.PutState err:", "error", err)
		return shim.Error(err.Error())
	}
	return shim.Success([]byte("ok"))
}

//对Developer退保证金的处理
func handleDeveloperDepositCashback(stub shim.ChaincodeStubInterface, cashbackAddr common.Address, cashbackValue *Cashback, balance *DepositBalance) error {
	//  如果在列表中，还要判断退出这部分钱后，是否需要移除候选列表
	if balance.EnterTime != "" {
		//  已在列表中
		err := handleDeveloperFromList(stub, cashbackAddr, cashbackValue, balance)
		if err != nil {
			log.Error("HandleJuryFromList err:", "error", err)
			return err
		}
	} else {
		////TODO 不在列表中,没有奖励，直接退
		err := handleCommonJuryOrDev(stub, cashbackAddr, cashbackValue, balance)
		if err != nil {
			log.Error("HandleCommonJuryOrDev err:", "error", err)
			return err
		}
	}
	return nil
}

//基金会处理
func handleForJuryApplyCashback(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	log.Info("handleForJuryApplyCashback")
	//  判断是否基金会发起的
	if !isFoundationInvoke(stub) {
		log.Error("please use foundation address")
		return shim.Error("please use foundation address")
	}
	//  地址，申请时间，是否同意
	if len(args) != 2 {
		log.Error("args need two parameters")
		return shim.Error("args need two parameters")
	}
	isOk := args[1]
	//  获取一下该用户下的账簿情况
	addr, err := common.StringToAddress(args[0])
	if err != nil {
		return shim.Error(err.Error())
	}
	balance, err := GetNodeBalance(stub, addr.String())
	if err != nil {
		log.Error("stub.GetDepositBalance err: ", "error", err)
		return shim.Error(err.Error())
	}
	//  判断没收节点账户是否为空
	if balance == nil {
		log.Error("balance is nil.")
		return shim.Error("balance is nil.")
	}
	//  获取请求列表
	listForCashback, err := GetListForCashback(stub)
	if err != nil {
		log.Error("Stub.GetListForCashback err:", "error", err)
		return shim.Error(err.Error())
	}
	if listForCashback == nil {
		log.Error("listForCashback is nil.")
		return shim.Error("listForCashback is nil.")
	}
	if _, ok := listForCashback[addr.String()]; !ok {
		log.Error("node is not exist in the list.")
		return shim.Error("node is not exist in the list.")
	}
	cashbackNode := listForCashback[addr.String()]
	delete(listForCashback, addr.String())

	if isOk == Ok {
		//  对余额处理
		err := handleJuryDepositCashback(stub, addr, cashbackNode, balance)
		if err != nil {
			log.Error("HandleJuryDepositCashback err:", "error", err)
			return shim.Error(err.Error())
		}
	} else if isOk == No {
		log.Info("does not agree")
	} else {
		log.Error("Please enter ok or no.")
		return shim.Error("Please enter ok or no.")
	}
	//更新列表
	err = SaveListForCashback(stub, listForCashback)
	if err != nil {
		log.Error("saveListForCashback err:", "error", err)
		return shim.Error(err.Error())
	}
	return shim.Success([]byte("ok"))
}

//对Jury退保证金的处理
func handleJuryDepositCashback(stub shim.ChaincodeStubInterface, cashbackAddr common.Address, cashbackValue *Cashback, balance *DepositBalance) error {
	//  如果在列表中，还要判断退出这部分钱后，是否需要移除候选列表
	if balance.EnterTime != "" {
		//  已在列表中
		err := handleJuryFromList(stub, cashbackAddr, cashbackValue, balance)
		if err != nil {
			log.Error("HandleJuryFromList err:", "error", err)
			return err
		}
	} else {
		////TODO 不在列表中,没有奖励，直接退
		err := handleCommonJuryOrDev(stub, cashbackAddr, cashbackValue, balance)
		if err != nil {
			log.Error("HandleCommonJuryOrDev err:", "error", err)
			return err
		}
	}
	return nil
}

//处理加入 参数：同意或不同意，节点的地址
func handleForApplyBecomeMediator(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	log.Info("handleForApplyBecomeMediator")
	if len(args) != 2 {
		log.Error("args need two parameters")
		return shim.Error("args need two parameters")
	}
	//  判断是否基金会发起的
	if !isFoundationInvoke(stub) {
		log.Error("please use foundation address")
		return shim.Error("please use foundation address")
	}
	//  判断处理地址是否申请过
	isOk := args[1]
	addr, err := common.StringToAddress(args[0])
	if err != nil {
		log.Error("string to address err: ", "error", err)
		return shim.Error(err.Error())
	}
	md, err := GetMediatorDeposit(stub, addr.String())
	if err != nil {
		log.Error("get mediator deposit error " + err.Error())
		return shim.Error(err.Error())
	}
	if md == nil {
		return shim.Error(addr.String() + " is nil")
	}
	if md.Status != Apply {
		return shim.Error(addr.String() + "is not applying")
	}

	//  不同意，直接删除
	if isOk == No {
		err = DelMediatorDeposit(stub, addr.String())
		if err != nil {
			return shim.Error(err.Error())
		}
	} else if isOk == Ok {
		//  获取同意列表
		agreeList, err := getList(stub, ListForAgreeBecomeMediator)
		if err != nil {
			log.Error("get agree list err: ", "error", err)
			return shim.Error(err.Error())
		}
		if agreeList == nil {
			agreeList = make(map[string]bool)
		}
		agreeList[addr.String()] = true
		//  保存同意列表
		err = saveList(stub, ListForAgreeBecomeMediator, agreeList)
		if err != nil {
			log.Error("save agree list err: ", "error", err)
			return shim.Error(err.Error())
		}
		// 修改同意时间
		md.AgreeTime = TimeStr()
		md.Status = Agree
		err = SaveMediatorDeposit(stub, addr.Str(), md)
		if err != nil {
			log.Error("save mediator info err: ", "error", err)
			return shim.Error(err.Error())
		}
	} else {
		log.Error("please enter ok")
		return shim.Error("please enter ok")
	}
	//  不管同意还是不同意都需要移除申请列表
	becomeList, err := getList(stub, ListForApplyBecomeMediator)
	if err != nil {
		log.Error("get become list err: ", "error", err)
		return shim.Error(err.Error())
	}
	delete(becomeList, addr.String())
	//  保存成为列表
	err = saveList(stub, ListForApplyBecomeMediator, becomeList)
	if err != nil {
		log.Error("save become list err: ", "error", err)
		return shim.Error(err.Error())
	}
	return shim.Success([]byte(nil))
}

//处理退出 参数：同意或不同意，节点的地址
func handleForApplyQuitMediator(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	log.Info("Start enter handleForApplyQuitMediator func")
	//  判断是否基金会发起的
	if !isFoundationInvoke(stub) {
		log.Error("please use foundation address")
		return shim.Error("please use foundation address")
	}
	//参数
	if len(args) != 2 {
		log.Error("Arg need two parameter.")
		return shim.Error("Arg need two parameter.")
	}
	//
	isOk := args[1]
	addr1 := args[0]
	addr, err := common.StringToAddress(addr1)
	if err != nil {
		log.Error("common.StringToAddress err:", "error", err)
		return shim.Error(err.Error())
	}
	//获取该账户
	md, err := GetMediatorDeposit(stub, addr.String())
	if err != nil {
		log.Error("Stub.GetDepositBalance err:", "error", err)
		return shim.Error(err.Error())
	}
	if md == nil {
		log.Error("Stub.GetDepositBalance err: balance is nil.")
		return shim.Error("Stub.GetDepositBalance err: balance is nil.")
	}
	if md.Status != Quitting {
		return shim.Error("not is quitting")
	}

	if isOk == No {

	} else if isOk == Ok {
		log.Info("foundation is agree with application.")
		err = deleteMediatorDeposit(stub, md, addr)
		if err != nil {
			log.Error("DeleteNode err:", "error", err)
			return shim.Error(err.Error())
		}
	} else {
		log.Error("please enter ok or no")
		return shim.Error("please enter ok or no")
	}
	//  删除退出列表
	ql, err := getList(stub, ListForApplyQuitMediator)
	if err != nil {
		return shim.Error("quit mediator list is nil")
	}
	delete(ql, addr.String())
	err = saveList(stub, ListForApplyQuitMediator, ql)
	if err != nil {
		return shim.Error("save quit mediator err " + err.Error())
	}
	return shim.Success([]byte(nil))
}

//基金会处理
func handleForMediatorApplyCashback(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	log.Info("handleForMediatorApplyCashback")
	//  判断是否基金会发起的
	if !isFoundationInvoke(stub) {
		log.Error("please use foundation address")
		return shim.Error("please use foundation address")
	}
	//  地址，是否同意
	if len(args) != 2 {
		log.Error("args need two parameters")
		return shim.Error("args need two parameters")
	}
	//  获取一下该用户下的账簿情况
	isOk := args[1]
	addr, err := common.StringToAddress(args[0])
	if err != nil {
		log.Error("string to address err: ", "error", err)
		return shim.Error(err.Error())
	}
	md, err := GetMediatorDeposit(stub, addr.String())
	if err != nil {
		log.Error("get cashback node balance err: ", "error", err)
		return shim.Error(err.Error())
	}
	//  判断没收节点账户是否为空
	if md == nil {
		log.Error("get cashback node balance: balance is nil")
		return shim.Error("get cashback node balance: balance is nil")
	}
	//  判断是否在退出部分保证金列表
	//获取没收列表
	listForCashback, err := GetListForCashback(stub)
	if err != nil {
		log.Error("stub.GetListForCashback err:", "error", err)
		return shim.Error(err.Error())
	}
	if listForCashback == nil {
		log.Error("listForCashback is nil")
		return shim.Error("listForCashback is nil")
	}
	if _, ok := listForCashback[addr.String()]; !ok {
		log.Error("node is not exist in the cashback list.")
		return shim.Error("node is not exist in the cashback list")
	}
	cashbackNode := listForCashback[addr.String()]
	//  不过如何，都会移除列表
	delete(listForCashback, addr.String())
	//  判断处理结果
	if isOk == Ok {
		//TODO 这是只退一部分钱，剩下余额还是在规定范围之内
		err = cashbackSomeMediatorDeposit(stub, addr, cashbackNode, md)
		if err != nil {
			log.Error("cashbackSomeDeposit err: ", "error", err)
			return shim.Error(err.Error())
		}
	} else if isOk == No {
		log.Info("does not agree")
	} else {
		log.Error("please enter ok or no")
		return shim.Error("please enter ok or no")
	}
	//更新列表
	err = SaveListForCashback(stub, listForCashback)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success([]byte(nil))
}
