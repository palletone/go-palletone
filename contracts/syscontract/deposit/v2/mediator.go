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

//Package v2 implements some functions for v2 contract.
package v2

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/palletone/go-palletone/dag/errors"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/contracts/shim"
	pb "github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
	"github.com/palletone/go-palletone/dag/dagconfig"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/contracts/syscontract"
)

//  基金会手动将生产节点移除候选列表
func handleForRemoveMediator(stub shim.ChaincodeStubInterface, address string) error {
	// 条件检查
	if !isFoundationInvoke(stub) {
		return errors.New("please use foundation address")
	}
	_, err := common.StringToAddress(address)
	if err != nil {
		return err
	}

	// 从列表移除
	err = moveCandidate(modules.MediatorList, address, stub)
	if err != nil {
		log.Error("MoveCandidate err:", "error", err)
		return err
	}
	err = moveCandidate(modules.JuryList, address, stub)
	if err != nil {
		log.Error("MoveCandidate err:", "error", err)
		return err
	}

	// 减少活跃mediator数量
	gp, err := stub.GetSystemConfig()
	if err != nil {
		log.Errorf("fail to get system config err")
		return err
	}

	desiredActiveMediatorCount := uint64(gp.ChainParameters.ActiveMediatorCount - 1)
	if desiredActiveMediatorCount == 0 {
		log.Errorf("ActiveMediatorCount will become 0")
		return err
	}

	err = modifyActiveMediatorCount(stub, desiredActiveMediatorCount)
	if err != nil {
		log.Debugf(err.Error())
		return err
	}

	return nil
}

func modifyActiveMediatorCount(stub shim.ChaincodeStubInterface, desiredActiveMediatorCount uint64) error {
	desiredActiveMediatorCountStr := strconv.FormatUint(desiredActiveMediatorCount, 10)

	// 获取已经要修改的系统参数
	resultBytes, err := stub.GetContractState(syscontract.SysConfigContractAddress, modules.DesiredSysParamsWithoutVote)
	if err != nil {
		log.Debugf(err.Error())
		return err
	}

	var modifies map[string]string
	if resultBytes != nil && string(resultBytes) != "" {
		err := json.Unmarshal(resultBytes, &modifies)
		if err != nil {
			log.Debugf(err.Error())
			return err
		}
	}

	// 增加修改活跃mediator数量
	if modifies == nil {
		modifies = make(map[string]string)
	}
	modifies[modules.DesiredActiveMediatorCount] = desiredActiveMediatorCountStr

	modifyByte, err := json.Marshal(modifies)
	if err != nil {
		log.Debugf(err.Error())
		return err
	}
	err = stub.PutContractState(syscontract.SysConfigContractAddress,
		modules.DesiredSysParamsWithoutVote, modifyByte)
	if err != nil {
		log.Debugf(err.Error())
		return err
	}

	return nil
}

// 处理基金会手动添加生产节点
func handleForAddMediator(stub shim.ChaincodeStubInterface, mediatorCreateArgs string) error {
	if !isFoundationInvoke(stub) {
		return errors.New("please use foundation address")
	}

	log.Info("Start entering handleForAddMediator func")
	// 解析参数
	var mco modules.MediatorCreateArgs
	err := json.Unmarshal([]byte(mediatorCreateArgs), &mco)
	if err != nil {
		errStr := fmt.Sprintf("invalid args: %v", err.Error())
		log.Errorf(errStr)
		return errors.New(errStr)
	}

	// 参数验证
	if mco.MediatorInfoBase == nil || mco.MediatorApplyInfo == nil {
		errStr := fmt.Sprintf("invalid args, is null")
		log.Errorf(errStr)
		return errors.New(errStr)
	}

	// 补全参数
	if mco.RewardAddr == "" {
		mco.RewardAddr = mco.RewardAdd
	}

	// 获取作为juror的相关参数
	_, jde, err := mco.Validate()
	if err != nil {
		errStr := fmt.Sprintf("invalid args: %v", err.Error())
		log.Errorf(errStr)
		return errors.New(errStr)
	}

	// 保存juror保证金和相关设置
	jd := &modules.JurorDeposit{}
	jd.Balance = 0
	jd.EnterTime = getTime(stub)
	jd.Role = modules.Jury
	jd.Address = mco.AddStr
	jd.JurorDepositExtra = jde
	err = saveJuryBalance(stub, mco.AddStr, jd)
	if err != nil {
		log.Error("save node balance err: ", "error", err)
		return err
	}

	//  加入候选列表
	err = addToCandaditeList(stub, mco.AddStr, modules.MediatorList)
	if err != nil {
		log.Error("addCandidateListAndPutStateForMediator err: ", "error", err)
		return err
	}

	//  自动加入jury候选列表
	err = addToCandaditeList(stub, mco.AddStr, modules.JuryList)
	if err != nil {
		log.Error("addCandidateListAndPutStateForMediator err: ", "error", err)
		return err
	}

	// 增加活跃mediator数量
	gp, err := stub.GetSystemConfig()
	if err != nil {
		log.Errorf("fail to get system config err")
		return err
	}
	desiredActiveMediatorCount := uint64(gp.ChainParameters.ActiveMediatorCount + 1)

	err = modifyActiveMediatorCount(stub, desiredActiveMediatorCount)
	if err != nil {
		log.Debugf(err.Error())
		return err
	}

	log.Info("End entering handleForAddMediator func")
	return nil
}

//  加入相应候选列表，mediator jury dev
func addToCandaditeList(stub shim.ChaincodeStubInterface, invokeAddr string, candidate string) error {
	//  获取列表
	list, err := getList(stub, candidate)
	if err != nil {
		return err
	}
	if list == nil {
		list = make(map[string]bool)
	}

	list[invokeAddr] = true
	err = saveList(stub,candidate,list)
	if err != nil {
		return err
	}
	return nil
}

//  申请加入
func applyBecomeMediator(stub shim.ChaincodeStubInterface, mediatorCreateArgs string) error {
	log.Info("Start entering apply for become mediator func")
	var mco modules.MediatorCreateArgs
	err := json.Unmarshal([]byte(mediatorCreateArgs), &mco)
	if err != nil {
		errStr := fmt.Sprintf("invalid args: %v", err.Error())
		log.Errorf(errStr)
		return errors.New(errStr)
	}

	// 参数验证
	if mco.MediatorInfoBase == nil || mco.MediatorApplyInfo == nil {
		errStr := fmt.Sprintf("invalid args, is null")
		log.Errorf(errStr)
		return errors.New(errStr)
	}

	if mco.RewardAddr == "" {
		mco.RewardAddr = mco.RewardAdd
	}

	_, jde, err := mco.Validate()
	if err != nil {
		errStr := fmt.Sprintf("invalid args: %v", err.Error())
		log.Errorf(errStr)
		return errors.New(errStr)
	}

	//  获取请求地址
	applyingAddr, err := stub.GetInvokeAddress()
	if err != nil {
		log.Error("Stub.GetInvokeAddress err:", "error", err)
		return err
	}

	applyingAddrStr := applyingAddr.Str()
	if mco.AddStr != applyingAddrStr {
		errStr := fmt.Sprintf("the calling account(%v) is not applying account(%v)",
			applyingAddrStr, mco.AddStr)
		log.Error(errStr)
		return errors.New(errStr)
	}

	//  判断该地址是否是第一次申请
	mDeposit, err := getMediatorDeposit(stub, mco.AddStr)
	if err != nil {
		return err
	}
	if mDeposit != nil {
		return errors.New(mco.AddStr + " has applied for become mediator")
	}

	//  获取申请列表
	becomeList, err := getList(stub, modules.ListForApplyBecomeMediator)
	if err != nil {
		log.Error("Stub.GetBecomeMediatorApplyList err:", "error", err)
		return err
	}
	//  判断
	if becomeList == nil {
		log.Info("Stub.GetBecomeMediatorApplyList: list is nil")
		becomeList = make(map[string]bool)
	}

	becomeList[mco.AddStr] = true
	//  保存列表
	err = saveList(stub, modules.ListForApplyBecomeMediator, becomeList)
	if err != nil {
		log.Error("saveList err:", "error", err)
		return err
	}

	// 保存账户信息
	md := modules.NewMediatorDeposit()
	md.ApplyEnterTime = getTime(stub)
	md.Status = modules.Apply
	md.Role = modules.Mediator
	err = saveMediatorDeposit(stub, mco.AddStr, md)
	if err != nil {
		log.Error("SaveMedInfo err:", "error", err)
		return err
	}

	// 保存juror保证金
	jd := &modules.JurorDeposit{}
	jd.Balance = 0
	jd.EnterTime = md.ApplyEnterTime
	jd.Role = modules.Jury
	jd.Address = applyingAddrStr
	jd.JurorDepositExtra = jde
	err = saveJuryBalance(stub, applyingAddrStr, jd)
	if err != nil {
		log.Error("save node balance err: ", "error", err)
		return err
	}

	log.Info("End entering apply for become mediator func")
	return nil
}

//mediator 交付保证金：
func mediatorPayToDepositContract(stub shim.ChaincodeStubInterface) error {
	log.Info("starting entering MediatorPayToDepositContract func.")
	//  判断是否是交付保证金到保证金合约地址
	invokeTokens, err := isContainDepositContractAddr(stub)
	if err != nil {
		log.Error("get deposit invoke tokens err: ", "error", err)
		return err
	}
	//  获取交付地址
	invokeAddr, err := stub.GetInvokeAddress()
	if err != nil {
		log.Error("get invoke address err: ", "error", err)
		return err
	}

	//  TODO 添加进入质押记录
	//err = pledgeDepositRep(stub, invokeAddr, invokeTokens.Amount)
	//if err != nil {
	//	return shim.Error(err.Error())
	//}

	// 缴纳保证金的几种情况：
	// 1. 在正式网申请后，缴纳50w ptn；
	// 2. 前期的mediator，后面追缴保证金；
	// 3. 退出mediator列表后，再次缴纳保证；

	// 判断是否已经申请过，即是否创建保证金对象
	invokeAddrStr := invokeAddr.String()
	md, err := getMediatorDeposit(stub, invokeAddrStr)
	if err != nil {
		log.Error("get node balance err: ", "error", err)
		return err
	}

	if md == nil {
		return errors.New(invokeAddrStr + " does not apply for mediator")
	}

	// 退出mediator列表后，再次缴纳保证
	if md.Status == modules.Quited {
		md.Status = modules.Agree
		md.ApplyQuitTime = ""
	}
	//  判断是否已经获得同意状态
	if !strings.EqualFold(md.Status, modules.Agree) {
		// if strings.ToLower(md.Status) != strings.ToLower(modules.Agree) {
		return errors.New(invokeAddrStr + " does not in the agree list")
	}

	gp, err := stub.GetSystemConfig()
	cp := gp.ChainParameters
	if err != nil {
		return err
	}

	all := invokeTokens.Amount + md.Balance
	if all != cp.DepositAmountForMediator {
		str := fmt.Errorf("Mediator needs to pay only %d  deposit.", cp.DepositAmountForMediator-md.Balance)
		log.Warn(str.Error())
		return str
	}

	//  加入候选列表
	err = addCandaditeList(stub, invokeAddr, modules.MediatorList)
	if err != nil {
		log.Error("addCandidateListAndPutStateForMediator err: ", "error", err)
		return err
	}

	//  处理数据
	md.Status = modules.Agree
	md.Role = modules.Mediator
	md.EnterTime = getTime(stub)
	md.Balance = all
	//  保存账户信息
	err = saveMediatorDeposit(stub, invokeAddrStr, md)
	if err != nil {
		log.Error("save node balance err: ", "error", err)
		return err
	}

	//  自动加入jury候选列表
	err = addCandaditeList(stub, invokeAddr, modules.JuryList)
	if err != nil {
		log.Error("addCandidateListAndPutStateForMediator err: ", "error", err)
		return err
	}

	// 兼mediator的juror的保证金余额永远为0，防止货币增发
	//jd, err := getJuryBalance(stub, invokeAddrStr)
	//if err != nil {
	//	log.Error("get node balance err: ", "error", err)
	//	return shim.Error(err.Error())
	//}
	//
	//if jd == nil {
	//	return shim.Error(invokeAddrStr + " does not apply for mediator")
	//}
	//
	//jd.Balance = all
	//err = saveJuryBalance(stub, invokeAddrStr, jd)
	//if err != nil {
	//	log.Error("save node balance err: ", "error", err)
	//	return shim.Error(err.Error())
	//}

	return nil
}

//  申请退出 参数：暂时 节点地址
func mediatorApplyQuit(stub shim.ChaincodeStubInterface) pb.Response {
	err := applyQuitList(modules.Mediator, stub)
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
	mediator, err := getMediatorDeposit(stub, invokeAddr.Str())
	if err != nil {
		log.Error("get node balance err: ", "error", err)
		return shim.Error(err.Error())
	}
	mediator.ApplyQuitTime = getTime(stub)
	mediator.Status = modules.Quitting
	//  保存账户信息
	err = saveMediatorDeposit(stub, invokeAddr.Str(), mediator)
	if err != nil {
		log.Error("save mediator info err: ", "error", err)
		return shim.Error(err.Error())
	}
	log.Info("end entering mediatorApplyQuitMediator func.")
	return shim.Success(nil)
}

//  申请加入
func updateMediatorInfo(stub shim.ChaincodeStubInterface, mediatorUpdateArgs string) pb.Response {
	log.Info("Start entering UpdateMediatorInfo func")

	var mua modules.MediatorUpdateArgs
	err := json.Unmarshal([]byte(mediatorUpdateArgs), &mua)
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
	mdeposit, err := getMediatorDeposit(stub, mua.AddStr)
	if err != nil {
		return shim.Error(err.Error())
	}
	if mdeposit == nil {
		return shim.Error(mua.AddStr + " is not a mediator")
	}

	log.Info("End entering UpdateMediatorInfo func")
	return shim.Success([]byte("ok"))
}

func handleMediator(stub shim.ChaincodeStubInterface, quitAddr common.Address) error {
	md, err := getMediatorDeposit(stub, quitAddr.Str())
	if err != nil {
		log.Error("get node balance err: ", "error", err)
		return err
	}
	//  移除退出列表
	listForQuit, err := getListForQuit(stub)
	if err != nil {
		return err
	}
	delete(listForQuit, quitAddr.String())
	err = saveListForQuit(stub, listForQuit)
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
	md.Status = modules.Quited
	md.Balance = 0
	md.EnterTime = ""
	//  保存
	err = saveMediatorDeposit(stub, quitAddr.Str(), md)
	if err != nil {
		return err
	}
	return nil
}

func convertMediatorDeposit2Json(md *modules.MediatorDeposit) *modules.MediatorDepositJson {
	mdJson := &modules.MediatorDepositJson{}

	dbJson := convertDepositBalance2Json(&md.DepositBalance)
	mdJson.DepositBalanceJson = *dbJson
	mdJson.MediatorDepositExtra = md.MediatorDepositExtra

	return mdJson
}
