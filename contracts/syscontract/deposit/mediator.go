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
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/award"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/contracts/shim"
	pb "github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
	"github.com/palletone/go-palletone/dag/modules"
	"strconv"
)

//  申请加入
func applyBecomeMediator(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	log.Info("Start entering apply for become mediator func")
	//  检查参数
	if len(args) != 1 {
		log.Error("Arg need only one parameter.")
		return shim.Error("Arg need only one parameter.")
	}
	//  获取请求地址
	invokeAddr, err := stub.GetInvokeAddress()
	if err != nil {
		log.Error("Stub.GetInvokeAddress err:", "error", err)
		return shim.Error(err.Error())
	}

	//  获取同意列表
	agreeList, err := GetList(stub, ListForAgreeBecomeMediator)
	if err != nil {
		log.Error("Stub.GetAgreeForBecomeMediatorList err:", "error", err)
		return shim.Error(err.Error())
	}
	if agreeList != nil {
		//  判断是否已经申请过，并且通过了申请
		if agreeList[invokeAddr.String()] {
			log.Error("Node is exist in the agree list.")
			return shim.Error("Node is exist in the agree list.")
		}
	}
	//  获取申请列表
	becomeList, err := GetList(stub, ListForApplyBecomeMediator)
	if err != nil {
		log.Error("Stub.GetBecomeMediatorApplyList err:", "error", err)
		return shim.Error(err.Error())
	}
	//  判断
	if becomeList == nil {
		log.Info("Stub.GetBecomeMediatorApplyList: list is nil")
		becomeList = make(map[string]bool)
	} else {
		//  判断是否已经申请过了
		if becomeList[invokeAddr.String()] {
			log.Debug("Node is exist in the become list.")
			return shim.Error("Node is exist in the become list.")
		}
	}
	becomeList[invokeAddr.String()] = true
	//  保存列表
	err = saveList(stub, ListForApplyBecomeMediator, becomeList)
	if err != nil {
		log.Error("saveList err:", "error", err)
		return shim.Error(err.Error())
	}

	// 保存账户信息
	md := NewMediatorDeposit()
	md.ApplyEnterTime = TimeStr()
	md.Status = Apply
	err = SaveMediatorDeposit(stub, invokeAddr.Str(), md)
	if err != nil {
		log.Error("SaveMedInfo err:", "error", err)
		return shim.Error(err.Error())
	}

	log.Info("End entering apply for become mediator func")
	return shim.Success([]byte("ok"))
}

//序列化list for mediator
func saveList(stub shim.ChaincodeStubInterface, key string, list map[string]bool) error {
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

//  申请退出 参数：暂时 节点地址
func mediatorApplyQuitMediator(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	log.Info("start entering mediatorApplyQuitMediator func")
	//  获取请求地址
	invokeAddr, err := stub.GetInvokeAddress()
	if err != nil {
		log.Error("get invoke address err: ", "error", err)
		return shim.Error(err.Error())
	}
	//  获取候选列表
	candidateList, err := GetList(stub, modules.MediatorList)
	if err != nil {
		log.Error("get candidate list err: ", "error", err)
		return shim.Error(err.Error())
	}
	if candidateList == nil {
		log.Error("get candidate list: list is nil")
		return shim.Error("get candidate list: list is nil")
	}
	if !candidateList[invokeAddr.String()] {
		log.Error("node is not exist in the candidate list")
		return shim.Error("node is not exist in the candidate list")
	}
	//  获取节点信息
	mediator, err := GetMediatorDeposit(stub, invokeAddr.Str())
	if err != nil {
		log.Error("get node balance err: ", "error", err)
		return shim.Error(err.Error())
	}
	mediator.ApplyQuitTime = TimeStr()
	mediator.Status = Quitting
	//  获取退出列表
	quitList, err := GetList(stub, ListForApplyQuitMediator)
	if err != nil {
		log.Error("get quit list err: ", "error", err)
		return shim.Error(err.Error())
	}
	if quitList == nil {
		quitList = make(map[string]bool)
		quitList[invokeAddr.String()] = true
	} else {
		if quitList[invokeAddr.String()] {
			log.Error("node is exist in the quit list")
			return shim.Error("node is exist in the quit list")
		}
		quitList[invokeAddr.String()] = true
	}
	//  保存退出列表
	err = saveList(stub, ListForApplyQuitMediator, quitList)
	if err != nil {
		log.Error("save quit list err: ", "error", err)
		return shim.Error(err.Error())
	}
	//  保存账户信息
	err = SaveMediatorDeposit(stub, invokeAddr.Str(), mediator)
	if err != nil {
		log.Error("save mediator info err: ", "error", err)
		return shim.Error(err.Error())
	}
	log.Info("end entering mediatorApplyQuitMediator func.")
	return shim.Success([]byte(nil))
}

func deleteMediatorDeposit(stub shim.ChaincodeStubInterface, md *MediatorDeposit, nodeAddr common.Address) error {
	//  计算币龄收益
	//endTime := md.LastModifyTime * DTimeDuration
	//endTime, _ := time.Parse(Layout, md.LastModifyTime)
	endTime := StrToTime(md.LastModifyTime)
	//
	depositRate, err := stub.GetSystemConfig(modules.DepositRate)
	if err != nil {
		log.Error("stub.GetSystemConfig err:", "error", err)
		return err
	}
	//
	awards := award.GetAwardsWithCoins(md.Balance, endTime.Unix(), depositRate)
	//  本金+利息
	md.Balance += awards
	invokeTokens := new(modules.AmountAsset)
	invokeTokens.Amount = md.Balance
	//
	fees, err := stub.GetInvokeFees()
	if err != nil {
		log.Error("stub.GetInvokeFees err:", "error", err)
		return err
	}
	invokeTokens.Asset = fees.Asset
	//  调用从合约把token转到请求地址
	err = stub.PayOutToken(nodeAddr.String(), invokeTokens, 0)
	if err != nil {
		log.Error("Stub.PayOutToken err:", "error", err)
		return err
	}

	//  获取候选列表
	candidateList, err := GetList(stub, modules.MediatorList)
	if err != nil {
		log.Error("Stub.GetCandidateListForMediator err:", "error", err)
		return err
	}
	//
	if candidateList == nil {
		log.Error("Stub.GetCandidateListForMediator:list is nil.")
		return fmt.Errorf("%s", "Stub.GetCandidateListForMediator:list is nil.")
	}
	//  移除
	delete(candidateList, nodeAddr.String())
	//
	err = saveList(stub, modules.MediatorList, candidateList)
	if err != nil {
		log.Error("MarshalAndPutStateForMediatorList err:", "error", err)
		return err
	}

	//  更新
	md.Status = Quited
	md.Balance = 0
	//  保存
	err = SaveMediatorDeposit(stub, nodeAddr.Str(), md)
	if err != nil {
		return err
	}

	return nil
}

//mediator 交付保证金：
func mediatorPayToDepositContract(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	log.Info("starting entering mediatorPayToDepositContract func.")
	//  获取保证金下线，在状态数据库中
	depositAmountsForMediatorStr, err := stub.GetSystemConfig(modules.DepositAmountForMediator)
	if err != nil {
		log.Error("get deposit amount for mediator err: ", "error", err)
		return shim.Error(err.Error())
	}
	//  转换保证金数量
	depositAmountsForMediator, err := strconv.ParseUint(depositAmountsForMediatorStr, 10, 64)
	if err != nil {
		log.Error("strconv.ParseUint err:", "error", err)
		return shim.Error(err.Error())
	}
	//  获取交付地址
	invokeAddr, err := stub.GetInvokeAddress()
	if err != nil {
		log.Error("get invoke address err: ", "error", err)
		return shim.Error(err.Error())
	}
	//  判断是否是交付保证金到保证金合约地址
	invokeTokens, err := isContainDepositContractAddr(stub)
	if err != nil {
		log.Error("get deposit invoke tokens err: ", "error", err)
		return shim.Error(err.Error())
	}
	//  获取同意列表
	agreeList, err := GetList(stub, ListForAgreeBecomeMediator)
	if err != nil {
		log.Error("get agree list err: ", "error", err)
		return shim.Error(err.Error())
	}
	//  判断是否已经通过了申请
	if agreeList == nil {
		log.Error("agree list is nil")
		return shim.Error("agree list is nil")
	}
	//  判断是否已经通过了申请
	if !agreeList[invokeAddr.String()] {
		log.Error("node is not exist in the agree list,you should apply for it.")
		return shim.Error("node is not exist in the agree list,you should apply for it.")
	}
	//  获取保证金账户信息
	md, err := GetMediatorDeposit(stub, invokeAddr.String())
	if err != nil {
		log.Error("get node balance err: ", "error", err)
		return shim.Error(err.Error())
	}
	//  第一次交付
	if md.Balance == 0 {
		//  判断保证金是否足够(Mediator第一次交付必须足够)
		if invokeTokens.Amount < depositAmountsForMediator {
			//TODO 第一次交付不够的话，这里必须终止
			log.Error("Payment amount is not enough.")
			return shim.Error("Payment amount is not enough.")
		}
		//  加入候选列表
		err = addCandidateListAndPutStateForMediator(stub, invokeAddr)
		if err != nil {
			log.Error("addCandidateListAndPutStateForMediator err: ", "error", err)
			return shim.Error(err.Error())
		}

		//  处理数据
		md.EnterTime = TimeStr()
		md.Balance += invokeTokens.Amount
		md.LastModifyTime = TimeStr()
	} else {
		//  TODO 再次交付保证金时，先计算当前余额的币龄奖励
		//  获取上次加入最后更改的时间
		//endTime := md.LastModifyTime * DTimeDuration
		//endTime, _ := time.Parse(Layout, md.LastModifyTime)
		endTime := StrToTime(md.LastModifyTime)
		//  获取保证金的年利率
		depositRate, err := stub.GetSystemConfig(modules.DepositRate)
		if err != nil {
			log.Error("get depositRate config err: ", "error", err)
			return shim.Error(err.Error())
		}
		//  计算币龄收益
		awards := award.GetAwardsWithCoins(md.Balance, endTime.Unix(), depositRate)
		md.Balance += awards
		//  处理数据
		md.Balance += invokeTokens.Amount
		md.LastModifyTime = TimeStr()
	}

	//  退出后，再交付的状态
	if md.Status == Quited {
		md.Status = Agree
		md.AgreeTime = TimeStr()
		md.ApplyQuitTime = ""
	}
	//  保存账户信息
	err = SaveMediatorDeposit(stub, invokeAddr.String(), md)
	if err != nil {
		log.Error("save node balance err: ", "error", err)
		return shim.Error(err.Error())
	}

	return shim.Success([]byte(nil))
}

//  加入候选列表并保存
func addCandidateListAndPutStateForMediator(stub shim.ChaincodeStubInterface,
	addr common.Address) error {
	//  获取节点候选列表
	candidateList, err := GetList(stub, modules.MediatorList)
	if err != nil {
		log.Error("get mediator candidate list err: ", "error", err)
		return err
	}
	//  判断是否为空
	if candidateList == nil {
		candidateList = make(map[string]bool)
	} else {
		if candidateList[addr.String()] {
			log.Error("node was in the candidate list.")
			return fmt.Errorf("%s", "node was in the candidate list.")
		}
	}
	candidateList[addr.String()] = true
	//  保存候选列表
	err = saveList(stub, modules.MediatorList, candidateList)
	if err != nil {
		log.Error("save candidate list err: ", "error", err)
		return err
	}
	return nil
}

//申请提取保证金
func mediatorApplyCashback(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	log.Info("start entering mediatorApplyCashback func")
	err := applyCashbackList(Mediator, stub, args)
	if err != nil {
		log.Error("apply cashback list err: ", "error", err)
		return shim.Error(err.Error())
	}
	log.Info("end entering mediatorApplyCashback func")
	return shim.Success([]byte(nil))
}

func handleMediator(stub shim.ChaincodeStubInterface, cashbackAddr common.Address, md *MediatorDeposit) error {
	//
	//depositPeriod, err := stub.GetSystemConfig(DepositPeriod)
	//if err != nil {
	//	log.Error("get deposit period err: ", "error", err)
	//	return err
	//}
	////
	//day, err := strconv.Atoi(depositPeriod)
	//if err != nil {
	//	log.Error("strconv.Atoi err: ", "error", err)
	//	return err
	//}
	//
	depositAmountsForMediatorStr, err := stub.GetSystemConfig(modules.DepositAmountForMediator)
	if err != nil {
		log.Error("get deposit amount for mediator err: ", "error", err)
		return err
	}
	//  转换
	depositAmountsForMediator, err := strconv.ParseUint(depositAmountsForMediatorStr, 10, 64)
	if err != nil {
		log.Error("strconv.ParseUint err:", "error", err)
		return err
	}
	//  获取提取列表
	listForCashback, err := GetListForCashback(stub)
	if err != nil {
		log.Error("get list for cashback err: ", "error", err)
		return err
	}
	if listForCashback == nil {
		log.Error("get list for cashback: list is nil")
		return fmt.Errorf("%s", "get list for cashback: list is nil")
	}
	if _, ok := listForCashback[cashbackAddr.String()]; !ok {
		log.Error("node is not exist in the list")
		return fmt.Errorf("%s", "node is not exist in the list")
	}
	//  退出信息
	cashbackNode := listForCashback[cashbackAddr.String()]
	delete(listForCashback, cashbackAddr.String())
	//更新列表
	err = SaveListForCashback(stub, listForCashback)
	if err != nil {
		return err
	}
	//  计算余额
	result := md.Balance - cashbackNode.CashbackTokens.Amount
	//  判断是否全部退
	//if result == 0 {
	//	//  加入候选列表的时的时间
	//	ent, err := strconv.ParseInt(md.EnterTime, 10, 64)
	//	startTime := time.Unix(ent*DTimeDuration, 0).UTC().YearDay()
	//	//  当前时间
	//	endTime := time.Now().UTC().YearDay()
	//	//  判断是否已超过规定周期
	//	if endTime-startTime >= day {
	//		//  退出全部，即删除cashback
	//		err = deleteNode(stub, md, cashbackAddr)
	//		if err != nil {
	//			log.Error("deleteNode err: ", "error", err)
	//			return err
	//		}
	//	} else {
	//		//  没有超过周期，不能退出
	//		log.Error("not exceeding the valid time,can not quit.")
	//		return fmt.Errorf("%s", "not exceeding the valid time,can not quit.")
	//	}
	//} else if result < depositAmountsForMediator {
	//  说明退款后，余额少于规定数量
	// 判断退款后，余额是否少于规定数量
	if result < depositAmountsForMediator {
		log.Error("can not cashback some")
		return fmt.Errorf("%s", "can not cashback some")
	} else {
		//TODO 这是只退一部分钱，剩下余额还是在规定范围之内
		err = cashbackSomeMediatorDeposit(stub, cashbackAddr, cashbackNode, md)
		if err != nil {
			log.Error("cashbackSomeDeposit err: ", "error", err)
			return err
		}
	}

	return nil
}

//  提取一部分保证金
func cashbackSomeMediatorDeposit(stub shim.ChaincodeStubInterface, cashbackAddr common.Address,
	cashbackValue *Cashback, md *MediatorDeposit) error {
	//  调用从合约把token转到请求地址
	err := stub.PayOutToken(cashbackAddr.String(), cashbackValue.CashbackTokens, 0)
	if err != nil {
		log.Error("stub.PayOutToken err: ", "error", err)
		return err
	}
	//endTime := md.LastModifyTime * DTimeDuration
	//endTime, _ := time.Parse(Layout, md.LastModifyTime)
	endTime := StrToTime(md.LastModifyTime)
	//
	depositRate, err := stub.GetSystemConfig(modules.DepositRate)
	if err != nil {
		log.Error("stub.GetSystemConfig err:", "error", err)
		return err
	}
	//
	awards := award.GetAwardsWithCoins(md.Balance, endTime.Unix(), depositRate)
	md.LastModifyTime = TimeStr()
	//  加上利息奖励
	md.Balance += awards
	//  减去提取部分
	md.Balance -= cashbackValue.CashbackTokens.Amount

	err = SaveMediatorDeposit(stub, cashbackAddr.String(), md)
	if err != nil {
		log.Error("SaveMedInfo err:", "error", err)
		return err
	}
	return nil
}
