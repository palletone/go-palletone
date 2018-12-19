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
	"fmt"
	"github.com/palletone/go-palletone/common/award"
	"github.com/palletone/go-palletone/contracts/shim"
	pb "github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
	"github.com/palletone/go-palletone/dag/modules"
	"strconv"
	"strings"
	"time"
)

//申请加入  参数：暂时  姓名
func applyBecomeMediator(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 4 {
		return shim.Error("arg need four parameters.")
	}
	invokeAddr, err := stub.GetInvokeAddress()
	if err != nil {
		return shim.Error(err.Error())
	}

	name := args[0]
	info := args[1]
	url := args[2]
	email := args[3]
	mediatorInfo := modules.MediatorInfo{
		Name:    name,
		Address: invokeAddr,
		Info:    info,
		Url:     url,
		Email:   email,
		Time:    time.Now().UTC(),
	}
	//获取列表
	becomeList, err := stub.GetBecomeMediatorApplyList()
	if err != nil {
		return shim.Error(err.Error())
	}
	if becomeList == nil {
		becomeList = []*modules.MediatorInfo{}
		becomeList = append(becomeList, &mediatorInfo)
	} else {
		isExist := isInMediatorInfolist(mediatorInfo.Address, becomeList)
		if isExist {
			return shim.Error("You is exist in the list.")
		}
		becomeList = append(becomeList, &mediatorInfo)
	}
	err = marshalAndPutStateForMediatorList(stub, "ListForApplyBecomeMediator", becomeList)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success([]byte("ok"))
}

//查找节点是否在列表中
func isInMediatorInfolist(addr string, list []*modules.MediatorInfo) bool {
	for _, m := range list {
		if strings.Compare(addr, m.Address) == 0 {
			return true
		}
	}
	return false
}

//处理加入 参数：同意或不同意，节点的地址
func handleForApplyBecomeMediator(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	invokeAddr, err := stub.GetInvokeAddress()
	if err != nil {
		return shim.Error(err.Error())
	}
	if strings.Compare(invokeAddr, foundationAddress) != 0 {
		return shim.Success([]byte("请求地址不正确，请使用基金会的地址"))
	}
	if len(args) != 2 {
		return shim.Error("arg need two parameter.")
	}
	//获取申请列表
	becomeList, err := stub.GetBecomeMediatorApplyList()
	if err != nil {
		return shim.Error(err.Error())
	}
	if becomeList == nil {
		return shim.Error("申请列表为空。")
	}
	isExist := isInMediatorInfolist(invokeAddr, becomeList)
	if !isExist {
		return shim.Error("不在申请列表中")
	}
	//var mediatorList []*modules.MediatorInfo
	mediator := &modules.MediatorInfo{}
	//不同意，移除申请列表
	if args[0] == "no" {
		becomeList, _ = moveMediatorFromList(args[1], becomeList)
	} else if args[0] == "ok" {
		//同意，移除列表，并且加入同意申请列表
		becomeList, mediator = moveMediatorFromList(args[1], becomeList)
		//获取同意列表
		agreeList, err := stub.GetAgreeForBecomeMediatorList()
		if err != nil {
			return shim.Error(err.Error())
		}
		if agreeList == nil {
			agreeList = []*modules.MediatorInfo{mediator}
		} else {
			isExist := isInMediatorInfolist(mediator.Address, agreeList)
			if isExist {
				return shim.Error("node in list")
			}
			agreeList = append(agreeList, mediator)
		}
		err = marshalAndPutStateForMediatorList(stub, "ListForAgreeBecomeMediator", agreeList)
		if err != nil {
			return shim.Error(err.Error())
		}
	}
	err = marshalAndPutStateForMediatorList(stub, "ListForApplyBecomeMediator", becomeList)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success([]byte("ok"))
}

//序列化list for mediator
func marshalAndPutStateForMediatorList(stub shim.ChaincodeStubInterface, key string, list []*modules.MediatorInfo) error {
	listByte, err := json.Marshal(list)
	if err != nil {
		return err
	}
	err = stub.PutState(key, listByte)
	if err != nil {
		return err
	}
	return nil
}

//func moveFromList(address string, list []*modules.MediatorInfo) (list1 []*modules.MediatorInfo) {
//	for i := 0; i < len(list); i++ {
//		if strings.Compare(list[i].Address, address) == 0 {
//			list1 = append(list[:i], list[i+1:]...)
//			break
//		}
//	}
//	return
//}

//从列表中删除并返回该节点
func moveMediatorFromList(address string, list []*modules.MediatorInfo) (list1 []*modules.MediatorInfo, mediator *modules.MediatorInfo) {
	for i := 0; i < len(list); i++ {
		if strings.Compare(list[i].Address, address) == 0 {
			mediator = list[i]
			list1 = append(list[:i], list[i+1:]...)
			return
		}
	}
	return
}

//申请退出  参数：暂时 节点地址
func mediatorApplyQuitMediator(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	invokeAddr, err := stub.GetInvokeAddress()
	if err != nil {
		return shim.Error(err.Error())
	}
	//获取同意列表
	agreeList, err := stub.GetAgreeForBecomeMediatorList()
	if err != nil {
		return shim.Error(err.Error())
	}
	if agreeList == nil {
		return shim.Error("Your node does not in agree list for mediator.")
	}
	isExist := isInMediatorInfolist(invokeAddr, agreeList)
	if !isExist {
		return shim.Error("node is not exist in the agree list.")
	}
	//获取节点信息
	mediator := &modules.MediatorInfo{}
	for _, m := range agreeList {
		if strings.Compare(m.Address, invokeAddr) == 0 {
			mediator = m
			break
		}
	}
	mediator.Time = time.Now().UTC()
	//获取列表
	quitList, err := stub.GetQuitMediatorApplyList()
	if err != nil {
		return shim.Error(err.Error())
	}
	if quitList == nil {
		quitList = []*modules.MediatorInfo{mediator}
	} else {
		isExist := isInMediatorInfolist(mediator.Address, quitList)
		if isExist {
			return shim.Error("node is exist in the quit list.")
		}
		quitList = append(quitList, mediator)
	}
	err = marshalAndPutStateForMediatorList(stub, "ListForApplyQuitMediator", quitList)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success([]byte("申请成功"))
}

//处理退出 参数：同意或不同意，节点的地址
func handleForApplyQuitMediator(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	//基金会
	invokeAddr, err := stub.GetInvokeAddress()
	if err != nil {
		return shim.Error(err.Error())
	}
	if strings.Compare(invokeAddr, foundationAddress) != 0 {
		return shim.Success([]byte("请求地址不正确，请使用基金会的地址"))
	}
	//参数
	if len(args) != 2 {
		return shim.Error("arg need two parameter.")
	}
	//获取申请列表
	quitList, err := stub.GetQuitMediatorApplyList()
	if err != nil {
		return shim.Error(err.Error())
	}
	if quitList == nil {
		return shim.Error("申请列表为空。")
	}
	//
	isExist := isInMediatorInfolist(args[1], quitList)
	if !isExist {
		return shim.Error("node is not exist in the quit list.")
	}
	//var mediatorList []*modules.MediatorInfo
	//不同意，移除申请列表
	if args[0] == "不同意" {
		quitList, _ = moveMediatorFromList(args[1], quitList)
	} else if args[0] == "同意" {
		//同意，移除列表，并且全款退出
		quitList, _ = moveMediatorFromList(args[1], quitList)
		//获取该账户
		balance, err := stub.GetDepositBalance(args[1])
		if err != nil {
			return shim.Error(err.Error())
		}
		if balance == nil {
			return shim.Error("balance is nil.")
		}
		err = deleteNode(stub, balance, args[1])
		if err != nil {
			return shim.Error(err.Error())
		}
	}
	err = marshalAndPutStateForMediatorList(stub, "ListForApplyQuitMediator", quitList)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success([]byte("ok"))
}

func deleteNode(stub shim.ChaincodeStubInterface, balance *modules.DepositBalance, nodeAddr string) error {
	//计算币龄收益
	awards := award.GetAwardsWithCoins(balance.TotalAmount, balance.LastModifyTime.Unix())
	//本金+利息
	balance.TotalAmount += awards
	//TODO 是否传入
	invokeTokens := new(modules.InvokeTokens)
	invokeTokens.Amount = balance.TotalAmount
	asset := modules.NewPTNAsset()
	invokeTokens.Asset = asset
	//调用从合约把token转到请求地址
	err := stub.PayOutToken(nodeAddr, invokeTokens, 0)
	if err != nil {
		return err
	}
	//删除节点
	err = stub.DelState(nodeAddr)
	if err != nil {
		return err
	}
	//获取候选列表
	candidateList, err := stub.GetCandidateListForMediator()
	if err != nil {
		return err
	}
	if candidateList == nil {
		return fmt.Errorf("%s", "candidate list for mediator is nil.")
	}
	//移除
	candidateList, _ = moveMediatorFromList(nodeAddr, candidateList)
	err = marshalAndPutStateForMediatorList(stub, "MediatorList", candidateList)
	if err != nil {
		return err
	}
	return nil
}

//mediator 交付保证金：
func mediatorPayToDepositContract(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	//交付地址
	invokeAddr, err := stub.GetInvokeAddress()
	if err != nil {
		return shim.Error("GetInvokeFromAddr error:")
	}
	//交付数量
	invokeTokens, err := stub.GetInvokeTokens()
	if err != nil {
		return shim.Success([]byte("GetPayToContractPtnTokens error:"))
	}
	//获取同意列表
	agreeList, err := stub.GetAgreeForBecomeMediatorList()
	if err != nil {
		return shim.Error(err.Error())
	}
	if agreeList == nil {
		return shim.Error("agree list is nil.")
	}
	isExist := isInMediatorInfolist(invokeAddr, agreeList)
	if !isExist {
		return shim.Error("node is not exist in the agree list,you should apply for it.")
	}
	//获取节点信息
	mediator := &modules.MediatorInfo{}
	for _, m := range agreeList {
		if strings.Compare(m.Address, invokeAddr) == 0 {
			mediator = m
			break
		}
	}
	//获取账户
	balance, err := stub.GetDepositBalance(invokeAddr)
	if err != nil {
		return shim.Error(err.Error())
	}
	//账户不存在，第一次参与
	if balance == nil {
		//判断保证金是否足够(Mediator第一次交付必须足够)
		if invokeTokens.Amount < depositAmountsForMediator {
			//TODO 第一次交付不够的话，这里必须终止
			return shim.Error("Payment amount is enough.")
		}
		//加入候选列表
		err = addCandidateListAndPutStateForMediator(stub, mediator)
		if err != nil {
			return shim.Error(err.Error())
		}
		balance = &modules.DepositBalance{}
		//处理数据
		balance.EnterTime = time.Now().UTC()
		updateForPayValue(balance, invokeTokens)
	} else {
		//TODO 再次交付保证金时，先计算当前余额的币龄奖励
		awards := award.GetAwardsWithCoins(balance.TotalAmount, balance.LastModifyTime.Unix())
		balance.TotalAmount += awards
		//处理数据
		updateForPayValue(balance, invokeTokens)
	}
	err = marshalAndPutStateForBalance(stub, invokeAddr, balance)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success([]byte("mediator pay ok."))
}

//加入候选列表并保存
func addCandidateListAndPutStateForMediator(stub shim.ChaincodeStubInterface, mediator *modules.MediatorInfo) error {
	candidateList, err := stub.GetCandidateListForMediator()
	if err != nil {
		return err
	}
	if candidateList == nil {
		candidateList = []*modules.MediatorInfo{mediator}
	} else {
		isExist := isInMediatorInfolist(mediator.Address, candidateList)
		if isExist {
			return fmt.Errorf("%s", "node is exist in the candidate list.")
		}
		candidateList = append(candidateList, mediator)
	}
	err = marshalAndPutStateForMediatorList(stub, "MediatorList", candidateList)
	if err != nil {
		return err
	}
	return nil
}

//申请提取保证金
func mediatorApplyCashback(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	err := applyCashbackList("Mediator", stub, args)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success([]byte("apply for cashback success."))
}

//基金会处理
func handleForMediatorApplyCashback(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	//地址，申请时间，是否同意
	if len(args) != 3 {
		return shim.Error("Input parameter error,need three parameters.")
	}
	//基金会地址
	invokeAddr, err := stub.GetInvokeAddress()
	if err != nil {
		return shim.Error(err.Error())
	}
	//判断没收请求地址是否是基金会地址
	if strings.Compare(invokeAddr, foundationAddress) != 0 {
		return shim.Error("请求地址不正确，请使用基金会的地址")
	}
	//获取一下该用户下的账簿情况
	addr := args[0]
	balance, err := stub.GetDepositBalance(addr)
	if err != nil {
		return shim.Error(err.Error())
	}
	//判断没收节点账户是否为空
	if balance == nil {
		return shim.Error("you have not depositWitnessPay for deposit.")
	}
	//获取申请时间戳
	strTime := args[1]
	applyTime, err := strconv.ParseInt(strTime, 10, 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	check := args[2]
	if check == "ok" {
		//对余额处理
		err = handleMediator(stub, addr, applyTime, balance)
		if err != nil {
			return shim.Error(err.Error())
		}
	} else {
		//移除提取申请列表
		err = moveAndPutStateFromCashbackList(stub, addr, applyTime)
		if err != nil {
			return shim.Error(err.Error())
		}
	}
	return shim.Success([]byte("handle for cashback success."))
}

func handleMediator(stub shim.ChaincodeStubInterface, cashbackAddr string, applyTime int64, balance *modules.DepositBalance) error {
	//获取提取列表
	listForCashback, err := stub.GetListForCashback()
	if err != nil {
		return err
	}
	if listForCashback == nil {
		return fmt.Errorf("%s", "listForCashback is nil")
	}
	isExist := isInCashbacklist(cashbackAddr, listForCashback)
	if !isExist {
		return fmt.Errorf("%s", "node is not exist in the list.")
	}
	cashbackNode, err := moveInApplyForCashbackList(stub, listForCashback, cashbackAddr, applyTime)
	if err != nil {
		return err
	}
	//计算余额
	result := balance.TotalAmount - cashbackNode.CashbackTokens.Amount
	//判断是否全部退
	if result == 0 {
		//加入候选列表的时的时间
		startTime := balance.EnterTime.YearDay()
		//当前时间
		endTime := time.Now().UTC().YearDay()
		//判断是否已超过规定周期
		if endTime-startTime >= depositPeriod {
			//退出全部，即删除cashback
			err = deleteNode(stub, balance, cashbackAddr)
			if err != nil {
				return err
			}
		} else {
			//没有超过周期，不能退出
			return fmt.Errorf("%s", "还在规定周期之内，不得退出列表")
		}
	} else if result < depositAmountsForMediator {
		//说明退款后，余额少于规定数量
		return fmt.Errorf("%s", "说明退款后，余额少于规定数量，对于Mediator来说，如果退部分保证后，余额少于规定数量，则不允许提款或者没收")
	} else {
		//TODO 这是只退一部分钱，剩下余额还是在规定范围之内
		err = cashbackSomeDeposit("Mediator", stub, cashbackAddr, cashbackNode, balance)
		if err != nil {
			return err
		}
	}
	return nil
}
