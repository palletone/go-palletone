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
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/award"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/contracts/shim"
	pb "github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
	"github.com/palletone/go-palletone/dag/modules"
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
	//  判断该地址是否是第一次申请
	mdeposit, err := GetMediatorDeposit(stub, invokeAddr.String())
	if err != nil {
		return shim.Error(err.Error())
	}
	if mdeposit != nil {
		return shim.Error(invokeAddr.String() + " has applied for become mediator")
	}
	//  获取申请列表
	becomeList, err := getList(stub, ListForApplyBecomeMediator)
	if err != nil {
		log.Error("Stub.GetBecomeMediatorApplyList err:", "error", err)
		return shim.Error(err.Error())
	}
	//  判断
	if becomeList == nil {
		log.Info("Stub.GetBecomeMediatorApplyList: list is nil")
		becomeList = make(map[string]bool)
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

//  申请退出 参数：暂时 节点地址
func mediatorApplyQuitMediator(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	log.Info("start entering mediatorApplyQuitMediator func")
	//  获取请求地址
	invokeAddr, err := stub.GetInvokeAddress()
	if err != nil {
		log.Error("get invoke address err: ", "error", err)
		return shim.Error(err.Error())
	}
	//  判断该地址是否为空并是否在候选列表中
	mediator, err := GetMediatorDeposit(stub, invokeAddr.Str())
	if err != nil {
		log.Error("get node balance err: ", "error", err)
		return shim.Error(err.Error())
	}
	if mediator == nil {
		return shim.Error(invokeAddr.String() + " is nil")
	}
	if mediator.Status == Quitting {
		return shim.Error("was in the list")
	}
	//  判断是否超过质押日期
	if !isOverDeadline(stub, mediator.EnterTime) {
		return shim.Error("does not over deadline")
	}
	//  获取候选列表
	candidateList, err := getList(stub, modules.MediatorList)
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
	//  加入在退出列表
	quitList, err := getList(stub, ListForApplyQuitMediator)
	if err != nil {
		log.Error("get quit list err: ", "error", err)
		return shim.Error(err.Error())
	}
	if quitList == nil {
		quitList = make(map[string]bool)
	}
	quitList[invokeAddr.String()] = true
	//  保存退出列表
	err = saveList(stub, ListForApplyQuitMediator, quitList)
	if err != nil {
		log.Error("save quit list err: ", "error", err)
		return shim.Error(err.Error())
	}
	mediator.ApplyQuitTime = TimeStr()
	mediator.Status = Quitting
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
	awards := caculateAwards(stub, md.Balance, md.LastModifyTime)
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

	//  移除列表
	err = moveCandidate(modules.MediatorList, nodeAddr.String(), stub)
	if err != nil {
		log.Error("MoveCandidate err:", "error", err)
		return err
	}
	err = moveCandidate(modules.JuryList, nodeAddr.String(), stub)
	if err != nil {
		log.Error("MoveCandidate err:", "error", err)
		return err
	}
	//  更新
	md.Status = Quited
	md.Balance = 0
	md.EnterTime = ""
	md.LastModifyTime = ""
	md.AgreeTime = ""
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
	//  判断是否是交付保证金到保证金合约地址
	invokeTokens, err := isContainDepositContractAddr(stub)
	if err != nil {
		log.Error("get deposit invoke tokens err: ", "error", err)
		return shim.Error(err.Error())
	}
	//  获取交付地址
	invokeAddr, err := stub.GetInvokeAddress()
	if err != nil {
		log.Error("get invoke address err: ", "error", err)
		return shim.Error(err.Error())
	}
	//  判断是否已经申请了
	md, err := GetMediatorDeposit(stub, invokeAddr.String())
	if err != nil {
		log.Error("get node balance err: ", "error", err)
		return shim.Error(err.Error())
	}
	if md == nil {
		return shim.Error(invokeAddr.String() + " does not apply for mediator")
	}
	//  TODO 退出后，再交付的状态
	if md.Status == Quited {
		md.Status = Agree
		md.AgreeTime = TimeStr()
		md.ApplyQuitTime = ""
	}
	//  判断是否已经获得同意状态
	if md.Status != Agree {
		return shim.Error(invokeAddr.String() + "does not in the agree list")
	}
	//  获取保证金下线，在状态数据库中
	//depositAmountsForMediatorStr, err := stub.GetSystemConfig(modules.DepositAmountForMediator)
	//if err != nil {
	//	log.Error("get deposit amount for mediator err: ", "error", err)
	//	return shim.Error(err.Error())
	//}
	////  转换保证金数量
	//depositAmountsForMediator, err := strconv.ParseUint(depositAmountsForMediatorStr, 10, 64)
	//if err != nil {
	//	log.Error("strconv.ParseUint err:", "error", err)
	//	return shim.Error(err.Error())
	//}
	cp, err := stub.GetSystemConfig()
	if err != nil {
		//log.Error("strconv.ParseUint err:", "error", err)
		return shim.Error(err.Error())
	}
	depositAmountsForMediator := cp.DepositAmountForMediator
	//  第一次交付
	if md.Balance == 0 {
		//  判断保证金是否足够(Mediator第一次交付必须足够)
		if invokeTokens.Amount < depositAmountsForMediator {
			//TODO 第一次交付不够的话，这里必须终止
			log.Error("Payment amount is not enough.")
			return shim.Error("Payment amount is not enough.")
		}
		//  加入候选列表
		err = addCandaditeList(stub, invokeAddr, modules.MediatorList)
		if err != nil {
			log.Error("addCandidateListAndPutStateForMediator err: ", "error", err)
			return shim.Error(err.Error())
		}
		//  自动加入jury候选列表
		err = addCandaditeList(stub, invokeAddr, modules.JuryList)
		if err != nil {
			log.Error("addCandidateListAndPutStateForMediator err: ", "error", err)
			return shim.Error(err.Error())
		}
		//  处理数据
		md.EnterTime = TimeStr()
	} else {
		//  TODO 再次交付保证金时，先计算当前余额的币龄奖励
		//  获取上次加入最后更改的时间
		//endTime := md.LastModifyTime * DTimeDuration
		//endTime, _ := time.Parse(Layout, md.LastModifyTime)
		endTime := StrToTime(md.LastModifyTime)
		//  获取保证金的年利率
		//depositRateStr, err := stub.GetSystemConfig(modules.DepositRate)
		//if err != nil {
		//	log.Error("get depositRate config err: ", "error", err)
		//	return shim.Error(err.Error())
		//}
		//depositRateFloat64, err := strconv.ParseFloat(depositRateStr, 64)
		//if err != nil {
		//	log.Errorf("string to float64 error: %s", err.Error())
		//	return shim.Error(err.Error())
		//}
		cp, err := stub.GetSystemConfig()
		if err != nil {
			//log.Error("strconv.ParseUint err:", "error", err)
			return shim.Error(err.Error())
		}
		depositRateFloat64 := cp.DepositRate
		//  计算币龄收益
		awards := award.GetAwardsWithCoins(md.Balance, endTime.Unix(), depositRateFloat64)
		md.Balance += awards
		//  处理数据
	}
	md.Balance += invokeTokens.Amount
	md.LastModifyTime = TimeStr()
	//  保存账户信息
	err = SaveMediatorDeposit(stub, invokeAddr.String(), md)
	if err != nil {
		log.Error("save node balance err: ", "error", err)
		return shim.Error(err.Error())
	}

	return shim.Success([]byte(nil))
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

//  提取一部分保证金
func cashbackSomeMediatorDeposit(stub shim.ChaincodeStubInterface, cashbackAddr common.Address,
	cashbackValue *Cashback, md *MediatorDeposit) error {
	//  调用从合约把token转到请求地址
	err := stub.PayOutToken(cashbackAddr.String(), cashbackValue.CashbackTokens, 0)
	if err != nil {
		log.Error("stub.PayOutToken err: ", "error", err)
		return err
	}
	awards := caculateAwards(stub, md.Balance, md.LastModifyTime)
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
