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

//  处理mediator申请退出保证金
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
		md.AgreeTime = getTiem(stub)
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
func handleForApplyQuitJury(stub shim.ChaincodeStubInterface, args []string) pb.Response {
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
	addr, err := common.StringToAddress(args[0])
	if err != nil {
		log.Error("common.StringToAddress err:", "error", err)
		return shim.Error(err.Error())
	}
	err = handleJury(stub, addr)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(nil)
}

//处理退出 参数：同意或不同意，节点的地址
func handleForApplyQuitDev(stub shim.ChaincodeStubInterface, args []string) pb.Response {
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
	addr, err := common.StringToAddress(args[0])
	if err != nil {
		log.Error("common.StringToAddress err:", "error", err)
		return shim.Error(err.Error())
	}
	err = handleDev(stub, addr)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(nil)
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
	addr, err := common.StringToAddress(args[0])
	if err != nil {
		log.Error("common.StringToAddress err:", "error", err)
		return shim.Error(err.Error())
	}
	err = handleMediator(stub, addr)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(nil)
}

//  处理普通节点提取质押PTN
//func handleExtractVote(stub shim.ChaincodeStubInterface, args []string) pb.Response {
//	if len(args) != 0 {
//		return shim.Error("need 0 args")
//	}
//	//  判断是否是基金会
//	if !isFoundationInvoke(stub) {
//		return shim.Error("please use foundation address")
//	}
//	//  保存质押提取
//	extPtnLis, err := getExtPtn(stub)
//	if err != nil {
//		return shim.Error(err.Error())
//	}
//	if extPtnLis == nil {
//		return shim.Error("list is nil")
//	}
//	cp, err := stub.GetSystemConfig()
//	if err != nil {
//		return shim.Error(err.Error())
//	}
//	day := cp.DepositPeriod
//	for k, v := range extPtnLis {
//		tim := StrToTime(v.Time)
//		dur := int(time.Since(tim).Hours())
//		if dur/24 < day {
//			continue
//		}
//		err := stub.PayOutToken(k, v.Amount, 0)
//		if err != nil {
//			return shim.Error(err.Error())
//		}
//	}
//	return shim.Success(nil)
//}

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
	node.LastModifyTime = getTiem(stub)
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
	node.LastModifyTime = getTiem(stub)
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
