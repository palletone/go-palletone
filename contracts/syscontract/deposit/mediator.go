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
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/contracts/shim"
	pb "github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
	"github.com/palletone/go-palletone/dag/dagconfig"
	"github.com/palletone/go-palletone/dag/modules"
)

//  申请加入
func applyBecomeMediator(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	log.Info("Start entering apply for become mediator func")
	//  检查参数
	if len(args) != 1 {
		errStr := "Arg need only one parameter."
		log.Error(errStr)
		return shim.Error(errStr)
	}

	var mco modules.MediatorCreateArgs
	err := json.Unmarshal([]byte(args[0]), &mco)
	if err != nil {
		errStr := fmt.Sprintf("invalid args: %v", err.Error())
		log.Errorf(errStr)
		return shim.Error(errStr)
	}

	// 参数验证
	if mco.MediatorInfoBase == nil || mco.MediatorApplyInfo == nil {
		errStr := fmt.Sprintf("invalid args, is null")
		log.Errorf(errStr)
		return shim.Error(errStr)
	}

	_, err = mco.Validate()
	if err != nil {
		errStr := fmt.Sprintf("invalid args: %v", err.Error())
		log.Errorf(errStr)
		return shim.Error(errStr)
	}

	//  获取请求地址
	applyingAddr, err := stub.GetInvokeAddress()
	if err != nil {
		log.Error("Stub.GetInvokeAddress err:", "error", err)
		return shim.Error(err.Error())
	}

	applyingAddrStr := applyingAddr.Str()
	if mco.AddStr != applyingAddrStr /*|| mco.RewardAdd != applyingAddrStr*/ {
		errStr := fmt.Sprintf("the calling account(%v) is not applying account(%v)",
			applyingAddrStr, mco.AddStr)
		log.Error(errStr)
		return shim.Error(errStr)
	}

	//  判断该地址是否是第一次申请
	mDeposit, err := GetMediatorDeposit(stub, mco.AddStr)
	if err != nil {
		return shim.Error(err.Error())
	}
	if mDeposit != nil {
		return shim.Error(mco.AddStr + " has applied for become mediator")
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

	becomeList[mco.AddStr] = true
	//  保存列表
	err = saveList(stub, ListForApplyBecomeMediator, becomeList)
	if err != nil {
		log.Error("saveList err:", "error", err)
		return shim.Error(err.Error())
	}

	// 保存账户信息
	md := NewMediatorDeposit()
	md.ApplyEnterTime = getTiem(stub)
	md.Status = Apply
	md.Role = Mediator
	err = SaveMediatorDeposit(stub, mco.AddStr, md)
	if err != nil {
		log.Error("SaveMedInfo err:", "error", err)
		return shim.Error(err.Error())
	}

	log.Info("End entering apply for become mediator func")
	return shim.Success([]byte("ok"))
}

//mediator 交付保证金：
func mediatorPayToDepositContract(stub shim.ChaincodeStubInterface /*, args []string*/) pb.Response {
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
	//  TODO 添加进入质押记录
	//err = pledgeDepositRep(stub, invokeAddr, invokeTokens.Amount)
	//if err != nil {
	//	return shim.Error(err.Error())
	//}
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
		md.ApplyQuitTime = ""
	}
	//  判断是否已经获得同意状态
	if md.Status != Agree {
		return shim.Error(invokeAddr.String() + "does not in the agree list")
	}
	gp, err := stub.GetSystemConfig()
	cp := gp.ChainParameters
	if err != nil {
		return shim.Error(err.Error())
	}
	//
	if md.Balance == 0 {
		if invokeTokens.Amount != cp.DepositAmountForMediator {
			str := fmt.Errorf("Mediator needs to pay only %d  deposit.", cp.DepositAmountForMediator)
			log.Error(str.Error())
			return shim.Error(str.Error())
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
		md.EnterTime = getTiem(stub)
		md.Balance = invokeTokens.Amount
		//  保存账户信息
		err = SaveMediatorDeposit(stub, invokeAddr.String(), md)
		if err != nil {
			log.Error("save node balance err: ", "error", err)
			return shim.Error(err.Error())
		}
		return shim.Success(nil)
	} else {
		//  如果账户不是零，考虑到系统参数升级，导致保证金数量问题，所以需要添加追缴逻辑
		//1判断追缴+账户余额是否为当前系统配置
		if md.Role != Mediator {
			return shim.Error("not mediator")
		}
		all := invokeTokens.Amount + md.Balance
		if all != cp.DepositAmountForMediator {
			str := fmt.Errorf("Mediator needs to pay only %d  deposit.", cp.DepositAmountForMediator)
			log.Error(str.Error())
			return shim.Error(str.Error())
		}
		//这里需要判断是否以及被基金会提前移除候选列表，即在规定时间内该节点没有追缴保证金
		b, err := isInCandidate(stub, invokeAddr.String(), modules.MediatorList)
		if err != nil {
			log.Debugf("isInCandidate error: %s", err.Error())
			return shim.Error(err.Error())
		}
		if !b {
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
		}
		md.Balance = all
		//  保存账户信息
		err = SaveMediatorDeposit(stub, invokeAddr.String(), md)
		if err != nil {
			log.Error("save node balance err: ", "error", err)
			return shim.Error(err.Error())
		}
		return shim.Success(nil)
	}
}

//  申请退出 参数：暂时 节点地址
func mediatorApplyQuit(stub shim.ChaincodeStubInterface /*, args []string*/) pb.Response {
	err := applyQuitList(Mediator, stub)
	if err != nil {
		log.Error("mediatorApplyQuitMediator err: ", "error", err)
		return shim.Error(err.Error())
	}
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
	mediator.ApplyQuitTime = getTiem(stub)
	mediator.Status = Quitting
	//  保存账户信息
	err = SaveMediatorDeposit(stub, invokeAddr.Str(), mediator)
	if err != nil {
		log.Error("save mediator info err: ", "error", err)
		return shim.Error(err.Error())
	}
	log.Info("end entering mediatorApplyQuitMediator func.")
	return shim.Success(nil)
}

//  申请加入
func updateMediatorInfo(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	log.Info("Start entering updateMediatorInfo func")
	//  检查参数
	if len(args) != 1 {
		errStr := "Arg need only one parameter."
		log.Error(errStr)
		return shim.Error(errStr)
	}

	var mua modules.MediatorUpdateArgs
	err := json.Unmarshal([]byte(args[0]), &mua)
	if err != nil {
		errStr := fmt.Sprintf("invalid args: %v", err.Error())
		log.Errorf(errStr)
		return shim.Error(errStr)
	}

	addr, err := mua.Validate()
	if err != nil {
		errStr := fmt.Sprintf("invalid args: %v", err.Error())
		log.Errorf(errStr)
		return shim.Error(errStr)
	}

	//  获取请求地址
	invokeAddr, err := stub.GetInvokeAddress()
	if err != nil {
		log.Error("Stub.GetInvokeAddress err:", "error", err)
		return shim.Error(err.Error())
	}

	if addr != invokeAddr {
		errStr := fmt.Sprintf("the calling account(%v) is not not produce account(%v)", invokeAddr.String(),
			mua.AddStr)
		log.Error(errStr)
	}

	// 判断该地址是否是mediator
	// 这样不能判断genesis文件中那几个mediator
	//mdeposit, err := GetMediatorDeposit(stub, mua.AddStr)
	//if err != nil {
	//	return shim.Error(err.Error())
	//}
	//if mdeposit == nil {
	//	return shim.Error(mua.AddStr + " is not a mediator")
	//}

	list, err := getList(stub, modules.MediatorList)
	if err != nil {
		return shim.Error(err.Error())
	}
	if list == nil {
		return shim.Success([]byte("false"))
	}

	if _, ok := list[mua.AddStr]; !ok {
		return shim.Error(mua.AddStr + " is not a mediator")
	}

	log.Info("End entering updateMediatorInfo func")
	return shim.Success([]byte("ok"))
}

func handleMediator(stub shim.ChaincodeStubInterface, quitAddr common.Address) error {
	md, err := GetMediatorDeposit(stub, quitAddr.Str())
	if err != nil {
		log.Error("get node balance err: ", "error", err)
		return err
	}
	//  移除退出列表
	listForQuit, err := GetListForQuit(stub)
	if err != nil {
		return err
	}
	delete(listForQuit, quitAddr.String())
	err = SaveListForQuit(stub, listForQuit)
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
	err = stub.PayOutToken(quitAddr.String(), modules.NewAmountAsset(md.Balance, gasToken), 0)
	if err != nil {
		log.Error("stub.PayOutToken err:", "error", err)
		return err
	}
	//  移除候选列表
	err = moveCandidate(modules.MediatorList, quitAddr.String(), stub)
	if err != nil {
		log.Error("MoveCandidate err:", "error", err)
		return err
	}
	err = moveCandidate(modules.JuryList, quitAddr.String(), stub)
	if err != nil {
		log.Error("MoveCandidate err:", "error", err)
		return err
	}
	//  更新
	md.Status = Quited
	md.Balance = 0
	md.EnterTime = ""
	//  保存
	err = SaveMediatorDeposit(stub, quitAddr.Str(), md)
	if err != nil {
		return err
	}
	return nil
}
