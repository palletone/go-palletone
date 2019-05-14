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

//Package deposit implements some functions for deposit contract.
package deposit

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/award"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/contracts/shim"
	pb "github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
	"github.com/palletone/go-palletone/dag/modules"
)

//申请加入  参数： jsonString
func applyBecomeMediator(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	log.Info("Start entering apply for become mediator func")
	if len(args) != 1 {
		log.Error("Arg need only one parameter.")
		return shim.Error("Arg need only one parameter.")
	}
	//var nodeInfo map[string]
	invokeAddr, err := stub.GetInvokeAddress()
	if err != nil {
		log.Error("Stub.GetInvokeAddress err:", "error", err)
		return shim.Error(err.Error())
	}
	var mco modules.MediatorCreateOperation

	err = json.Unmarshal([]byte(args[0]), &mco)
	if err != nil {
		log.Debugf("Save Apply Mediator(%v) Invoke Req", mco.AddStr)
		return shim.Error(err.Error())

	}
	mi := modules.NewMediatorInfo()
	*mi.MediatorInfoBase = *mco.MediatorInfoBase
	*mi.MediatorApplyInfo = *mco.MediatorApplyInfo

	mi.MediatorApplyInfo.ApplyEnterTime = time.Now().Unix() / DTimeDuration
	mi.MediatorApplyInfo.Status = modules.Apply
	mi.MediatorApplyInfo.Address = invokeAddr.String()
	//mediatorInfo := core.MediatorApplyInfo{
	//	Address: invokeAddr.String(),
	//	Content: args[0],
	//	Time:    time.Now().Unix() / DTimeDuration,
	//	Status:  Apply,
	//}
	//  获取同意列表
	agreeList, err := GetAgreeForBecomeMediatorLists(stub)
	if err != nil {
		log.Error("Stub.GetAgreeForBecomeMediatorList err:", "error", err)
		return shim.Error(err.Error())
	}
	if agreeList != nil {
		//  判断是否已经申请过了
		if _, ok := agreeList[invokeAddr.String()]; ok {
			log.Error("Node is exist in the agree list.")
			return shim.Error("Node is exist in the agree list.")
		}
		//isExist := isInMediatorInfolist(invokeAddr.String(), agreeList)
		//if isExist {
		//	log.Error("Node is exist in the agree list.")
		//	return shim.Error("Node is exist in the agree list.")
		//}
	}
	//  获取申请列表
	becomeList, err := GetBecomeMediatorApplyLists(stub)
	if err != nil {
		log.Error("Stub.GetBecomeMediatorApplyList err:", "error", err)
		return shim.Error(err.Error())
	}
	if becomeList == nil {
		log.Info("Stub.GetBecomeMediatorApplyList: list is nil")
		becomeList = make(map[string]bool)
		becomeList[invokeAddr.String()] = true
		//becomeList = []*core.MediatorApplyInfo{&mediatorInfo}
	} else {
		if _, ok := becomeList[invokeAddr.String()]; ok {
			log.Debug("Node is exist in the become list.")
			return shim.Error("Node is exist in the become list.")
		}
		//isExist := isInMediatorInfolist(invokeAddr.String(), becomeList)
		//if isExist {
		//	log.Debug("Node is exist in the become list.")
		//	return shim.Error("Node is exist in the become list.")
		//}
		becomeList[invokeAddr.String()] = true
	}
	err = marshalAndPutStateForMediatorList(stub, ListForApplyBecomeMediator, becomeList)
	if err != nil {
		log.Error("MarshalAndPutStateForMediatorList err:", "error", err)
		return shim.Error(err.Error())
	}
	//  保存账户信息
	err = SaveMedInfo(stub, mi)
	if err != nil {
		log.Error("SaveMedInfo err:", "error", err)
		return shim.Error(err.Error())
	}
	log.Info("End entering apply for become mediator func")
	return shim.Success([]byte("ok"))
}

//查找节点是否在列表中
//func isInMediatorInfolist(addr string, list []*core.MediatorApplyInfo) bool {
//	for _, m := range list {
//		if strings.Compare(addr, m.Address) == 0 {
//			return true
//		}
//	}
//	return false
//}

//序列化list for mediator
func marshalAndPutStateForMediatorList(stub shim.ChaincodeStubInterface, key string, list map[string]bool) error {
	listByte, err := json.Marshal(list)
	if err != nil {
		log.Error("Json.Marshal err:", "error", err)
		return err
	}
	err = stub.PutState(key, listByte)
	if err != nil {
		log.Error("Stub.PutState err:", "error", err)
		return err
	}
	return nil
}

//从列表中删除并返回该节点
//func moveMediatorFromList(address string, list []*core.MediatorApplyInfo) (newList []*core.MediatorApplyInfo,
//	mediator *core.MediatorApplyInfo) {
//	for i := 0; i < len(list); i++ {
//		if strings.Compare(list[i].Address, address) == 0 {
//			mediator = list[i]
//			newList = append(list[:i], list[i+1:]...)
//			return
//		}
//	}
//	return
//}

//申请退出  参数：暂时 节点地址
func mediatorApplyQuitMediator(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	log.Info("Start entering mediatorApplyQuitMediator func.")
	invokeAddr, err := stub.GetInvokeAddress()
	if err != nil {
		log.Error("Stub.GetInvokeAddress err:", "error", err)
		return shim.Error(err.Error())
	}
	//获取同意列表
	agreeList, err := GetAgreeForBecomeMediatorLists(stub)
	if err != nil {
		log.Error("Stub.GetAgreeForBecomeMediatorList err:", "error", err)
		return shim.Error(err.Error())
	}
	if agreeList == nil {
		log.Error("Stub.GetAgreeForBecomeMediatorList:list is nil")
		return shim.Error("Agree list is nil.")
	}
	if _, ok := agreeList[invokeAddr.String()]; !ok {
		log.Error("Node is not exist in the agree list.")
		return shim.Error("Node is not exist in the agree list.")
	}
	//isExist := isInMediatorInfolist(invokeAddr.String(), agreeList)
	//if !isExist {
	//	log.Error("Node is not exist in the agree list.")
	//	return shim.Error("Node is not exist in the agree list.")
	//}
	//获取候选列表
	candidateList, err := GetCandidateListForMediator(stub)
	if err != nil {
		log.Error("Stub.GetCandidateListForMediator err:", "error", err)
		return shim.Error(err.Error())
	}
	if candidateList == nil {
		log.Error("Stub.GetCandidateListForMediator err:", "error", "list is nil.")
		return shim.Error("Stub.GetCandidateListForMediator err: list is nil.")

	}
	if _, ok := candidateList[invokeAddr.String()]; !ok {
		log.Error("Node is not exist in the candidate list.")
		return shim.Error("Node is not exist in the candidate list.")
	}
	//isExist = isInMediatorInfolist(invokeAddr.String(), candidateList)
	//if !isExist {
	//	log.Error("Node is not exist in the candidate list.")
	//	return shim.Error("Node is not exist in the candidate list.")
	//}
	//获取节点信息
	mediator, err := GetMedNodeInfo(stub, invokeAddr.String())
	if err != nil {
		log.Error("GetMedNodeInfo err:", "error", err)
		return shim.Error(err.Error())
	}
	//mediator := &core.MediatorApplyInfo{}
	//for _, m := range agreeList {
	//	if strings.Compare(m.Address, invokeAddr.String()) == 0 {
	//		mediator = m
	//		break
	//	}
	//}
	mediator.ApllyQuitTime = time.Now().Unix() / DTimeDuration
	mediator.Status = modules.Quit
	//获取列表
	quitList, err := GetQuitMediatorApplyLists(stub)
	if err != nil {
		log.Error("Stub.GetQuitMediatorApplyList err:", "error", err)
		return shim.Error(err.Error())
	}
	if quitList == nil {
		log.Info("Stub.GetQuitMediatorApplyList err:list is nil.")
		quitList = make(map[string]bool)
		quitList[invokeAddr.String()] = true
		//quitList = []*core.MediatorApplyInfo{mediator}
	} else {
		if _, ok := quitList[invokeAddr.String()]; ok {
			log.Error("Node is exist in the quit list.")
			return shim.Error("Node is exist in the quit list.")
		}
		//isExist := isInMediatorInfolist(mediator.Address, quitList)
		//if isExist {
		//	log.Error("Node is exist in the quit list.")
		//	return shim.Error("Node is exist in the quit list.")
		//}
		//quitList = append(quitList, mediator)
		quitList[invokeAddr.String()] = true
	}
	err = marshalAndPutStateForMediatorList(stub, ListForApplyQuitMediator, quitList)
	if err != nil {
		log.Error("MarshalAndPutStateForMediatorList err:", "error", err)
		return shim.Error(err.Error())
	}
	//  保存账户信息
	err = SaveMedInfo(stub, mediator)
	if err != nil {
		log.Error("SaveMedInfo err:", "error", err)
		return shim.Error(err.Error())
	}
	log.Info("End entering mediatorApplyQuitMediator func.")
	return shim.Success([]byte("ok"))
}

func deleteNode(stub shim.ChaincodeStubInterface, balance *modules.MediatorInfo, nodeAddr common.Address) error {
	//计算币龄收益
	endTime := balance.LastModifyTime * DTimeDuration
	depositRate, err := stub.GetSystemConfig(modules.DepositRate)
	if err != nil {
		log.Error("stub.GetSystemConfig err:", "error", err)
		return err
	}
	awards := award.GetAwardsWithCoins(balance.Balance, endTime, depositRate)
	//本金+利息
	balance.Balance += awards
	invokeTokens := new(modules.AmountAsset)
	invokeTokens.Amount = balance.Balance
	fees, err := stub.GetInvokeFees()
	if err != nil {
		log.Error("stub.GetInvokeFees err:", "error", err)
		return err
	}
	invokeTokens.Asset = fees.Asset
	//调用从合约把token转到请求地址
	err = stub.PayOutToken(nodeAddr.String(), invokeTokens, 0)
	if err != nil {
		log.Error("Stub.PayOutToken err:", "error", err)
		return err
	}
	//删除节点
	err = stub.DelState(nodeAddr.String())
	if err != nil {
		log.Error("Stub.DelState err:", "error", err)
		return err
	}
	//获取候选列表
	candidateList, err := GetCandidateListForMediator(stub)
	if err != nil {
		log.Error("Stub.GetCandidateListForMediator err:", "error", err)
		return err
	}
	if candidateList == nil {
		log.Error("Stub.GetCandidateListForMediator:list is nil.")
		return fmt.Errorf("%s", "Stub.GetCandidateListForMediator:list is nil.")
	}
	//移除
	delete(candidateList, nodeAddr.String())
	//candidateList, _ = moveMediatorFromList(nodeAddr, candidateList)
	err = marshalAndPutStateForMediatorList(stub, modules.MediatorList, candidateList)
	if err != nil {
		log.Error("MarshalAndPutStateForMediatorList err:", "error", err)
		return err
	}
	return nil
}

//mediator 交付保证金：
func mediatorPayToDepositContract(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	depositAmountsForMediatorStr, err := stub.GetSystemConfig(DepositAmountForMediator)
	if err != nil {
		log.Error("Stub.GetSystemConfig with DepositAmountForMediator err:", "error", err)
		return shim.Error(err.Error())
	}
	//转换
	depositAmountsForMediator, err := strconv.ParseUint(depositAmountsForMediatorStr, 10, 64)
	if err != nil {
		log.Error("Strconv.ParseUint err:", "error", err)
		return shim.Error(err.Error())
	}
	log.Info("Stub.GetSystemConfig with DepositAmountForMediator:", "value", depositAmountsForMediator)
	log.Info("Starting entering mediatorPayToDepositContract func.")
	//交付地址
	invokeAddr, err := stub.GetInvokeAddress()
	if err != nil {
		log.Error("Stub.GetInvokeAddress err:", "error", err)
		return shim.Error(err.Error())
	}
	//交付数量
	//invokeTokens, err := stub.GetInvokeTokens()
	invokeTokens, err := isContainDepositContractAddr(stub)
	if err != nil {
		log.Error("Stub.GetInvokeTokens err:", "error", err)
		return shim.Error(err.Error())
	}
	//获取同意列表
	agreeList, err := GetAgreeForBecomeMediatorLists(stub)
	if err != nil {
		log.Error("Stub.GetAgreeForBecomeMediatorList err:", "error", err)
		return shim.Error(err.Error())
	}
	if agreeList == nil {
		log.Error("Stub.GetAgreeForBecomeMediatorList err: list is nil.")
		return shim.Error("Stub.GetAgreeForBecomeMediatorList err: list is nil.")
	}
	if _, ok := agreeList[invokeAddr.String()]; !ok {
		log.Error("Node is not exist in the agree list,you should apply for it.")
		return shim.Error("Node is not exist in the agree list,you should apply for it.")
	}
	//isExist := isInMediatorInfolist(invokeAddr.String(), agreeList)
	//if !isExist {
	//	log.Error("Node is not exist in the agree list,you should apply for it.")
	//	return shim.Error("Node is not exist in the agree list,you should apply for it.")
	//}
	//获取节点信息
	//mediator := &core.MediatorApplyInfo{}
	//isFound := false
	//for _, m := range agreeList {
	//	if strings.Compare(m.Address, invokeAddr.String()) == 0 {
	//		mediator = m
	//		isFound = true
	//		break
	//	}
	//}
	//if !isFound {
	//	log.Error("Apply time is wrong.")
	//	return shim.Error("Apply time is wrong.")
	//}
	//获取账户
	balance, err := GetMedNodeInfo(stub, invokeAddr.String())
	if err != nil {
		log.Error("Stub.GetDepositBalance err:", "error", err)
		return shim.Error(err.Error())
	}
	//账户不存在，第一次参与
	if balance.Balance == 0 {
		log.Info("Stub.GetDepositBalance: list is nil.")
		//判断保证金是否足够(Mediator第一次交付必须足够)
		if invokeTokens.Amount < depositAmountsForMediator {
			//TODO 第一次交付不够的话，这里必须终止
			log.Error("Payment amount is not enough.")
			return shim.Error("Payment amount is not enough.")
		}
		//加入候选列表
		err = addCandidateListAndPutStateForMediator(stub, invokeAddr)
		if err != nil {
			log.Error("AddCandidateListAndPutStateForMediator err:", "error", err)
			return shim.Error(err.Error())
		}
		//balance = &core.MediatorApplyInfo{}
		//处理数据
		balance.EnterTime = strconv.FormatInt(time.Now().Unix()/DTimeDuration, 10)
		updateForPayValue(balance, invokeTokens)
	} else {
		//TODO 再次交付保证金时，先计算当前余额的币龄奖励
		endTime := balance.LastModifyTime * DTimeDuration
		depositRate, err := stub.GetSystemConfig(modules.DepositRate)
		if err != nil {
			log.Error("stub.GetSystemConfig err:", "error", err)
			return shim.Error(err.Error())
		}
		awards := award.GetAwardsWithCoins(balance.Balance, endTime, depositRate)
		balance.Balance += awards
		//处理数据
		updateForPayValue(balance, invokeTokens)
	}
	err = marshalAndPutStateForBalance(stub, invokeAddr, balance)
	if err != nil {
		log.Error("MarshalAndPutStateForBalance err:", "error", err)
		return shim.Error(err.Error())
	}
	log.Info("End entering mediatorPayToDepositContract func.")
	return shim.Success([]byte("ok"))
}

//加入候选列表并保存
func addCandidateListAndPutStateForMediator(stub shim.ChaincodeStubInterface,
	addr common.Address) error {
	candidateList, err := GetCandidateListForMediator(stub)
	if err != nil {
		log.Error("Stub.GetCandidateListForMediator err:", "error", err)
		return err
	}
	if candidateList == nil {
		log.Info("Stub.GetCandidateListForMediator:list is nil.")
		candidateList = make(map[string]bool)
		candidateList[addr.String()] = true
		//candidateList = []*core.MediatorApplyInfo{mediator}
	} else {
		if _, ok := candidateList[addr.String()]; ok {
			log.Error("Node is exist in the candidate list.")
			return fmt.Errorf("%s", "Node is exist in the candidate list.")
		}
		//isExist := isInMediatorInfolist(mediator.Address, candidateList)
		//if isExist {
		//	log.Error("Node is exist in the candidate list.")
		//	return fmt.Errorf("%s", "Node is exist in the candidate list.")
		//}
		//candidateList = append(candidateList, mediator)
		candidateList[addr.String()] = true
	}
	err = marshalAndPutStateForMediatorList(stub, modules.MediatorList, candidateList)
	if err != nil {
		log.Error("MarshalAndPutStateForMediatorList err:", "error", err)
		return err
	}
	return nil
}

//申请提取保证金
func mediatorApplyCashback(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	log.Info("Start entering mediatorApplyCashback func.")
	err := applyCashbackList(Mediator, stub, args)
	if err != nil {
		log.Error("ApplyCashbackList err:", "error", err)
		return shim.Error(err.Error())
	}
	log.Info("End entering mediatorApplyCashback func.")
	return shim.Success([]byte("ok"))
}

func handleMediator(stub shim.ChaincodeStubInterface, cashbackAddr common.Address, applyTime int64, balance *modules.MediatorInfo) error {
	depositPeriod, err := stub.GetSystemConfig(DepositPeriod)
	if err != nil {
		log.Error("Stub.GetSystemConfig with DepositPeriod err:", "error", err)
		return err
	}
	day, err := strconv.Atoi(depositPeriod)
	if err != nil {
		log.Error("Strconv.Atoi err:", "error", err)
		return err
	}
	log.Info("Stub.GetSystemConfig with DepositPeriod:", "value", day)
	depositAmountsForMediatorStr, err := stub.GetSystemConfig(DepositAmountForMediator)
	if err != nil {
		log.Error("Stub.GetSystemConfig with DepositAmountForMediator err:", "error", err)
		return err
	}
	//转换
	depositAmountsForMediator, err := strconv.ParseUint(depositAmountsForMediatorStr, 10, 64)
	if err != nil {
		log.Error("Strconv.ParseUint err:", "error", err)
		return err
	}
	log.Info("Stub.GetSystemConfig with DepositAmountForMediator:", "value", depositAmountsForMediator)
	//获取提取列表
	listForCashback, err := GetListForCashback(stub)
	if err != nil {
		log.Error("Stub.GetListForCashback err:", "error", err)
		return err
	}
	if listForCashback == nil {
		log.Error("Stub.GetListForCashback:list is nil.")
		return fmt.Errorf("%s", "listForCashback is nil")
	}
	if _, ok := listForCashback[cashbackAddr]; !ok {
		log.Error("Node is not exist in the list.")
		return fmt.Errorf("%s", "Node is not exist in the list.")
	}
	//isExist := isInCashbacklist(cashbackAddr.String(), listForCashback)
	//if !isExist {
	//	log.Error("Node is not exist in the list.")
	//	return fmt.Errorf("%s", "Node is not exist in the list.")
	//}
	//获取节点信息
	//cashbackNode := &Cashback{}
	//isFound := false
	//for _, m := range listForCashback {
	//	if m.CashbackAddress == cashbackAddr && m.CashbackTime == applyTime {
	//		cashbackNode = m
	//		isFound = true
	//		break
	//	}
	//}
	//if !isFound {
	//	log.Error("Apply time is wrong.")
	//	return fmt.Errorf("%s", "Apply time is wrong.")
	//}
	//newList, _ := moveInApplyForCashbackList(stub, listForCashback, cashbackAddr, applyTime)
	delete(listForCashback, cashbackAddr)
	listForCashbackByte, err := json.Marshal(listForCashback)
	if err != nil {
		log.Error("Json.Marshal err:", "error", err)
		return err
	}
	//更新列表
	err = stub.PutState(ListForCashback, listForCashbackByte)
	if err != nil {
		log.Error("Stub.PutState err:", "error", err)
		return err
	}
	//计算余额
	result := balance.Balance - listForCashback[cashbackAddr].CashbackTokens.Amount
	//判断是否全部退
	if result == 0 {
		//加入候选列表的时的时间
		ent, err := strconv.ParseInt(balance.EnterTime, 10, 64)
		startTime := time.Unix(ent*DTimeDuration, 0).UTC().YearDay()
		//当前时间
		endTime := time.Now().UTC().YearDay()
		//判断是否已超过规定周期
		if endTime-startTime >= day {
			//退出全部，即删除cashback
			err = deleteNode(stub, balance, cashbackAddr)
			if err != nil {
				log.Error("DeleteNode err:", "error", err)
				return err
			}
		} else {
			//没有超过周期，不能退出
			log.Error("Not exceeding the valid time,can not quit.")
			return fmt.Errorf("%s", "Not exceeding the valid time,can not quit.")
		}
	} else if result < depositAmountsForMediator {
		//说明退款后，余额少于规定数量
		log.Error("Can not cashback some.")
		return fmt.Errorf("%s", "Can not cashback some.")
	} else {
		//TODO 这是只退一部分钱，剩下余额还是在规定范围之内
		err = cashbackSomeDeposit(Mediator, stub, cashbackAddr, listForCashback[cashbackAddr], balance)
		if err != nil {
			log.Error("CashbackSomeDeposit err:", "error", err)
			return err
		}
	}
	return nil
}
