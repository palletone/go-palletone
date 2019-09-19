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
	"fmt"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/contracts/shim"
	pb "github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
	"github.com/palletone/go-palletone/dag/dagconfig"
	"github.com/palletone/go-palletone/dag/modules"
	"strings"
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
	isOk := strings.ToLower(args[1])
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
	if md.Status != modules.Apply {
		return shim.Error(addr.String() + "is not applying")
	}

	//  不同意，直接删除
	if isOk == modules.No {
		err = DelMediatorDeposit(stub, addr.String())
		if err != nil {
			return shim.Error(err.Error())
		}
	} else if isOk == modules.Ok {
		//  获取同意列表
		agreeList, err := getList(stub, modules.ListForAgreeBecomeMediator)
		if err != nil {
			log.Error("get agree list err: ", "error", err)
			return shim.Error(err.Error())
		}
		if agreeList == nil {
			agreeList = make(map[string]bool)
		}
		agreeList[addr.String()] = true
		//  保存同意列表
		err = saveList(stub, modules.ListForAgreeBecomeMediator, agreeList)
		if err != nil {
			log.Error("save agree list err: ", "error", err)
			return shim.Error(err.Error())
		}
		// 修改同意时间
		md.AgreeTime = getTime(stub)
		md.Status = modules.Agree
		err = SaveMediatorDeposit(stub, addr.Str(), md)
		if err != nil {
			log.Error("save mediator info err: ", "error", err)
			return shim.Error(err.Error())
		}
	} else {
		log.Error("please enter Ok or No")
		return shim.Error("please enter Ok or No")
	}
	//  不管同意还是不同意都需要移除申请列表
	becomeList, err := getList(stub, modules.ListForApplyBecomeMediator)
	if err != nil {
		log.Error("get become list err: ", "error", err)
		return shim.Error(err.Error())
	}
	delete(becomeList, addr.String())
	//  保存成为列表
	err = saveList(stub, modules.ListForApplyBecomeMediator, becomeList)
	if err != nil {
		log.Error("save become list err: ", "error", err)
		return shim.Error(err.Error())
	}
	return shim.Success([]byte(nil))
}

//处理退出 参数：同意或不同意，节点的地址
func handleForApplyQuitJury(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	return handleForApplyQuitNode(stub, args, modules.Jury)
}

//处理退出 参数：同意或不同意，节点的地址
func handleForApplyQuitDev(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	return handleForApplyQuitNode(stub, args, modules.Developer)
}

func handleForApplyQuitNode(stub shim.ChaincodeStubInterface, args []string, role string) pb.Response {
	log.Info("Start enter handleForApplyQuitMediator func")
	//参数
	if len(args) != 2 {
		log.Error("Arg need two parameter.")
		return shim.Error("Arg need two parameter.")
	}
	//  判断是否基金会发起的
	if !isFoundationInvoke(stub) {
		log.Error("please use foundation address")
		return shim.Error("please use foundation address")
	}
	addr, err := common.StringToAddress(args[0])
	if err != nil {
		log.Error("common.StringToAddress err:", "error", err)
		return shim.Error(err.Error())
	}
	isOk := strings.ToLower(args[1])
	if isOk == modules.Ok {
		if role == modules.Developer {
			err = handleDev(stub, addr)
		}
		if role == modules.Jury {
			err = handleJury(stub, addr)
		}
		if err != nil {
			return shim.Error(err.Error())
		}
	} else if isOk == modules.No {
		//  移除退出列表
		listForQuit, err := GetListForQuit(stub)
		if err != nil {
			return shim.Error(err.Error())
		}
		delete(listForQuit, addr.String())
		err = SaveListForQuit(stub, listForQuit)
		if err != nil {
			return shim.Error(err.Error())
		}
	} else {
		log.Error("please enter Ok or No")
		return shim.Error("please enter Ok or No")
	}
	return shim.Success(nil)
}

//处理退出 参数：同意或不同意，节点的地址
func handleForApplyQuitMediator(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	log.Info("Start enter handleForApplyQuitMediator func")
	//参数
	if len(args) != 2 {
		log.Error("Arg need two parameter.")
		return shim.Error("Arg need two parameter.")
	}
	//  判断是否基金会发起的
	if !isFoundationInvoke(stub) {
		log.Error("please use foundation address")
		return shim.Error("please use foundation address")
	}
	addr, err := common.StringToAddress(args[0])
	if err != nil {
		log.Error("common.StringToAddress err:", "error", err)
		return shim.Error(err.Error())
	}
	isOk := strings.ToLower(args[1])
	if isOk == modules.Ok {
		err = handleMediator(stub, addr)
		if err != nil {
			return shim.Error(err.Error())
		}
	} else if isOk == modules.No {
		//  移除退出列表
		listForQuit, err := GetListForQuit(stub)
		if err != nil {
			return shim.Error(err.Error())
		}
		delete(listForQuit, addr.String())
		err = SaveListForQuit(stub, listForQuit)
		if err != nil {
			return shim.Error(err.Error())
		}
	} else {
		log.Error("please enter Ok or No")
		return shim.Error("please enter Ok or No")
	}
	return shim.Success(nil)
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
	//  处理没收地址
	addr := args[0]
	//  判断没收地址是否正确
	f, err := common.StringToAddress(addr)
	if err != nil {
		return shim.Error(err.Error())
	}
	//  获取基金会地址
	invokeAddr, err := stub.GetInvokeAddress()
	if err != nil {
		log.Error("get invoke address err: ", "error", err)
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
	forfeitureNode := listForForfeiture[f.String()]
	//  处理操作ok or no
	isOk := strings.ToLower(args[1])
	//check 如果为ok，则同意此申请，如果为no，则不同意此申请
	if isOk == modules.Ok {
		err = agreeForApplyForfeiture(stub, invokeAddr.String(), f.String(), forfeitureNode.ForfeitureRole)
		if err != nil {
			return shim.Error(err.Error())
		}
	} else if isOk == modules.No {
		//移除申请列表，不做处理
		log.Info("not agree to for apply forfeiture")
	} else {
		log.Error("Please enter Ok or No.")
		return shim.Error("Please enter Ok or No.")
	}
	//  不管同意与否都需要从列表中移除
	delete(listForForfeiture, addr)
	err = SaveListForForfeiture(stub, listForForfeiture)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(nil)
}

//同意申请没收请求
func agreeForApplyForfeiture(stub shim.ChaincodeStubInterface, foundationA string, forfeitureAddr string, forfeitureRole string) error {
	log.Info("Start entering agreeForApplyForfeiture func.")
	//判断节点类型
	switch {
	case forfeitureRole == modules.Mediator:
		return handleMediatorForfeitureDeposit(stub, foundationA, forfeitureAddr)
	case forfeitureRole == modules.Jury:
		return handleJuryForfeitureDeposit(stub, foundationA, forfeitureAddr)
	case forfeitureRole == modules.Developer:
		return handleDevForfeitureDeposit(stub, foundationA, forfeitureAddr)
	default:
		return fmt.Errorf("%s", "please enter validate role.")
	}
}
func handleJuryForfeitureDeposit(stub shim.ChaincodeStubInterface, foundationA string, forfeitureAddr string) error {
	node, err := GetJuryBalance(stub, forfeitureAddr)
	if err != nil {
		return err
	}
	if node == nil {
		return fmt.Errorf("node is nil")
	}

	//  移除列表
	err = moveCandidate(modules.JuryList, forfeitureAddr, stub)
	if err != nil {
		return err
	}
	//  退还保证金
	//cp, err := stub.GetSystemConfig()
	//if err != nil {
	//	return err
	//}
	//  调用从合约把token转到请求地址
	gasToken := dagconfig.DagConfig.GetGasToken().ToAsset()
	err = stub.PayOutToken(foundationA, modules.NewAmountAsset(node.Balance, gasToken), 0)
	if err != nil {
		log.Error("stub.PayOutToken err:", "error", err)
		return err
	}
	err = DelJuryBalance(stub, forfeitureAddr)
	if err != nil {
		return err
	}
	return nil
}

func handleDevForfeitureDeposit(stub shim.ChaincodeStubInterface, foundationA string, forfeitureAddr string) error {
	node, err := GetNodeBalance(stub, forfeitureAddr)
	if err != nil {
		return err
	}
	if node == nil {
		return fmt.Errorf("node is nil")
	}
	//  移除列表
	err = moveCandidate(modules.DeveloperList, forfeitureAddr, stub)
	if err != nil {
		return err
	}
	//  退还保证金
	//cp, err := stub.GetSystemConfig()
	//if err != nil {
	//	return err
	//}
	//  调用从合约把token转到请求地址
	gasToken := dagconfig.DagConfig.GetGasToken().ToAsset()
	err = stub.PayOutToken(foundationA, modules.NewAmountAsset(node.Balance, gasToken), 0)
	if err != nil {
		log.Error("stub.PayOutToken err:", "error", err)
		return err
	}
	err = DelNodeBalance(stub, forfeitureAddr)
	if err != nil {
		return err
	}
	return nil
}

//func handleNodeForfeitureDeposit(stub shim.ChaincodeStubInterface, foundationA string, forfeitureAddr string) error {
//
//}

//处理没收Mediator保证金
func handleMediatorForfeitureDeposit(stub shim.ChaincodeStubInterface, foundationA string, forfeitureAddr string) error {
	//  获取mediator
	md, err := GetMediatorDeposit(stub, forfeitureAddr)
	if err != nil {
		return err
	}
	if md == nil {
		return fmt.Errorf("node is nil")
	}
	//cp, err := stub.GetSystemConfig()
	//if err != nil {
	//	//log.Error("strconv.ParseUint err:", "error", err)
	//	return err
	//}
	//  调用从合约把token转到请求地址
	gasToken := dagconfig.DagConfig.GetGasToken().ToAsset()
	err = stub.PayOutToken(foundationA, modules.NewAmountAsset(md.Balance, gasToken), 0)
	if err != nil {
		log.Error("stub.PayOutToken err:", "error", err)
		return err
	}
	//  移除列表
	err = moveCandidate(modules.MediatorList, forfeitureAddr, stub)
	if err != nil {
		log.Error("MoveCandidate err:", "error", err)
		return err
	}
	err = moveCandidate(modules.JuryList, forfeitureAddr, stub)
	if err != nil {
		log.Error("MoveCandidate err:", "error", err)
		return err
	}
	//  更新
	md.Status = modules.Quited
	md.Balance = 0
	md.EnterTime = ""
	//  保存
	err = SaveMediatorDeposit(stub, forfeitureAddr, md)
	if err != nil {
		return err
	}
	return nil
}

func hanldeNodeRemoveFromAgreeList(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	address, err := common.StringToAddress(args[0])
	if err != nil {
		return shim.Error(err.Error())
	}
	if !isFoundationInvoke(stub) {
		return shim.Error("please use foundation address")
	}
	agreeList, err := getList(stub, modules.ListForAgreeBecomeMediator)
	if err != nil {
		return shim.Error(err.Error())
	}
	delete(agreeList, address.String())
	err = saveList(stub, modules.ListForAgreeBecomeMediator, agreeList)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(nil)
}

//func handleRemoveMediatorNode(stub shim.ChaincodeStubInterface, args []string) pb.Response {
//	address, err := common.StringToAddress(args[0])
//	if err != nil {
//		return shim.Error(err.Error())
//	}
//	if !isFoundationInvoke(stub) {
//		return shim.Error("please use foundation address")
//	}
//	err = DelMediatorDeposit(stub, address.String())
//	if err != nil {
//		return shim.Error(err.Error())
//	}
//	return shim.Success(nil)
//}
//func handleRemoveNormalNode(stub shim.ChaincodeStubInterface, args []string) pb.Response {
//	address, err := common.StringToAddress(args[0])
//	if err != nil {
//		return shim.Error(err.Error())
//	}
//	if !isFoundationInvoke(stub) {
//		return shim.Error("please use foundation address")
//	}
//	err = DelNodeBalance(stub, address.String())
//	if err != nil {
//		return shim.Error(err.Error())
//	}
//	return shim.Success(nil)
//}

func handleNodeInList(stub shim.ChaincodeStubInterface, args []string, role string) pb.Response {
	if len(args) > 0 {
		if !isFoundationInvoke(stub) {
			log.Debugf("please use foundation address")
			return shim.Error("please use foundation address")
		}
		//
		list := ""
		switch role {
		case modules.Mediator:
			list = modules.MediatorList
		case modules.Jury:
			list = modules.JuryList
		case modules.Developer:
			list = modules.DeveloperList
		}
		for _, a := range args {
			// 判断地址是否合法
			_, err := common.StringToAddress(a)
			if err != nil {
				log.Debugf("string to address error: %s", err.Error())
				return shim.Error(err.Error())
			}
			//  从候选列表中移除
			err = moveCandidate(list, a, stub)
			if err != nil {
				log.Debugf("move list error: %s", err.Error())
				return shim.Error(err.Error())
			}
		}
		return shim.Success(nil)
	}
	return shim.Error("invoke err")
}
