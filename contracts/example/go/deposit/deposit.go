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
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/contracts/shim"
	pb "github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
	"github.com/palletone/go-palletone/dag/modules"
	"strconv"
	"strings"
	"time"
)

var (
	isLoad                     bool
	depositAmountsForJury      uint64
	depositAmountsForMediator  uint64
	depositAmountsForDeveloper uint64
	depositPeriod              int
	foundationAddress          string
)

type DepositChaincode struct {
}

func (d *DepositChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("*** DepositChaincode system contract init ***")
	return shim.Success([]byte("ok"))
}

func (d *DepositChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	//初始化保证金合约的配置
	if !isLoad {
		initDepositCfg(stub)
	}
	funcName, args := stub.GetFunctionAndParameters()
	switch funcName {
	case "ApplyBecomeMediator":
		//申请成为Mediator
		return d.applyBecomeMediator(stub, args)
	case "HandleForApplyBecomeMediator":
		//基金会对加入申请Mediator进行处理
		return d.handleForApplyBecomeMediator(stub, args)
	case "MediatorApplyQuitMediator":
		//申请退出Mediator
		return d.mediatorApplyQuitMediator(stub, args)
	case "HandleForApplyQuitMediator":
		//基金会对退出申请Mediator进行处理
		return d.handleForApplyQuitMediator(stub, args)
	case "MediatorPayToDepositContract":
		//mediator 交付保证金
		return d.mediatorPayToDepositContract(stub, args)
	case "JuryPayToDepositContract":
		//jury 交付保证金
		return d.juryPayToDepositContract(stub, args)
	case "DeveloperPayToDepositContract":
		//developer 交付保证金
		return d.developerPayToDepositContract(stub, args)
	case "MediatorApplyCashback":
		//mediator 申请提取保证金
		return d.mediatorApplyCashback(stub, args)
	case "HandleForMediatorApplyCashback":
		//基金会处理提取保证金
		return d.handleForMediatorApplyCashback(stub, args)
	case "JuryApplyCashback":
		//jury 申请提取保证金
		return d.juryApplyCashback(stub, args)
	case "HandleForJuryApplyCashback":
		//基金会处理提取保证金
		return d.handleForJuryApplyCashback(stub, args)
	case "DeveloperApplyCashback":
		//developer 申请提取保证金
		return d.developerApplyCashback(stub, args)
	case "HandleForDeveloperApplyCashback":
		//基金会处理提取保证金
		return d.handleForDeveloperApplyCashback(stub, args)
	case "ApplyForForfeitureDeposit":
		//申请保证金没收
		//void forfeiture_deposit(const witness_object& wit, token_type amount)
		return d.applyForForfeitureDeposit(stub, args)
	case "HandleForForfeitureApplication":
		//基金会对申请做相应的处理
		return d.handleForForfeitureApplication(stub, args)
	//获取提取保证金申请列表
	case "GetListForCashbackApplication":
		list, err := stub.GetState("ListForCashback")
		if err != nil {
			return shim.Error(err.Error())
		}
		if list == nil {
			return shim.Success([]byte("[]"))
		}
		return shim.Success(list)
		//获取没收保证金申请列表
	case "GetListForForfeitureApplication":
		list, err := stub.GetState("ListForForfeiture")
		if err != nil {
			return shim.Error(err.Error())
		}
		if list == nil {
			return shim.Success([]byte("[]"))
		}
		return shim.Success(list)
		//获取Mediator候选列表
	case "GetListForMediatorCandidate":
		list, err := stub.GetState("MediatorList")
		if err != nil {
			return shim.Error(err.Error())
		}
		if list == nil {
			return shim.Success([]byte("[]"))
		}
		return shim.Success(list)
		//获取Jury候选列表
	case "GetListForJuryCandidater":
		list, err := stub.GetState("JuryList")
		if err != nil {
			return shim.Error(err.Error())
		}
		if list == nil {
			return shim.Success([]byte("[]"))
		}
		return shim.Success(list)
		//获取Contract Developer候选列表
	case "GetListForDeveloperCandidate":
		list, err := stub.GetState("DeveloperList")
		if err != nil {
			return shim.Error(err.Error())
		}
		if list == nil {
			return shim.Success([]byte("[]"))
		}
		return shim.Success(list)
		//获取某个节点的账户
	case "GetBalanceWithAddr":
		balance, err := stub.GetState(args[0])
		if err != nil {
			return shim.Error(err.Error())
		}
		if balance == nil {
			return shim.Success([]byte("balance is nil"))
		}
		if string(balance) == "" {
			return shim.Success([]byte("balance is nil"))
		}
		return shim.Success(balance)
		//获取Mediator申请加入列表
	case "GetBecomeMediatorApplyList":
		list, err := stub.GetState("ListForApplyBecomeMediator")
		if err != nil {
			return shim.Error(err.Error())
		}
		if list == nil {
			return shim.Success([]byte("[]"))
		}
		return shim.Success(list)
		//获取已同意的mediator列表
	case "GetAgreeForBecomeMediatorList":
		list, err := stub.GetState("ListForAgreeBecomeMediator")
		if err != nil {
			return shim.Error(err.Error())
		}
		if list == nil {
			return shim.Success([]byte("[]"))
		}
		return shim.Success(list)
		//获取Mediator申请退出列表
	case "GetQuitMediatorApplyList":
		list, err := stub.GetState("ListForApplyQuitMediator")
		if err != nil {
			return shim.Error(err.Error())
		}
		if list == nil {
			return shim.Success([]byte("[]"))
		}
		return shim.Success(list)
	}
	return shim.Success([]byte("Invoke error"))
}

func initDepositCfg(stub shim.ChaincodeStubInterface) {
	depositPeriod, err := stub.GetSystemConfig("DepositPeriod")
	if err != nil {
		log.Error("Stub.GetSystemConfig with DepositPeriod err:", "error", err)
		return
	}
	day, err := strconv.Atoi(depositPeriod)
	if err != nil {
		log.Error("Strconv.Atoi err:", "error", err)
		return
	}
	log.Info("Stub.GetSystemConfig with DepositPeriod:", "value", day)
	//foundationAddress, err = stub.GetSystemConfig("FoundationAddress")
	//if err != nil {
	//	fmt.Println(err.Error())
	//	return
	//}
	foundationAddress = "P129MFVxaLP4N9FZxYQJ3QPJks4gCeWsF9p"
	log.Info("Stub.GetSystemConfig with FoundationAddress:", "value", foundationAddress)

	depositAmountsForMediatorStr, err := stub.GetSystemConfig("DepositAmountForMediator")
	if err != nil {
		log.Error("Stub.GetSystemConfig with DepositAmountForMediator err:", "error", err)
		return
	}
	//转换
	depositAmountsForMediator, err = strconv.ParseUint(depositAmountsForMediatorStr, 10, 64)
	if err != nil {
		log.Error("Strconv.ParseUint err:", "error", err)
		return
	}
	log.Info("Stub.GetSystemConfig with DepositAmountForMediator:", "value", depositAmountsForMediator)
	//
	//fmt.Println("需要的mediator保证金数量=", depositAmountsForMediator)
	//fmt.Println()
	depositAmountsForJuryStr, err := stub.GetSystemConfig("DepositAmountForJury")
	if err != nil {
		log.Error("Stub.GetSystemConfig with DepositAmountForJury err:", "error", err)
		return
	}
	//转换
	depositAmountsForJury, err = strconv.ParseUint(depositAmountsForJuryStr, 10, 64)
	if err != nil {
		log.Error("Strconv.ParseUint err:", "error", err)
		return
	}
	log.Info("Stub.GetSystemConfig with DepositAmountForJury:", "value", depositAmountsForJury)

	depositAmountsForDeveloperStr, err := stub.GetSystemConfig("DepositAmountForDeveloper")
	if err != nil {
		log.Error("Stub.GetSystemConfig with DepositAmountForDeveloper err:", "error", err)
		return
	}
	//转换
	depositAmountsForDeveloper, err = strconv.ParseUint(depositAmountsForDeveloperStr, 10, 64)
	if err != nil {
		log.Error("Strconv.ParseUint err:", "error", err)
		return
	}
	log.Info("Stub.GetSystemConfig with DepositAmountForDeveloper:", "value", depositAmountsForDeveloper)
	isLoad = true
	log.Info("Init deposit config success.")
}

func (d *DepositChaincode) mediatorPayToDepositContract(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	return mediatorPayToDepositContract(stub, args)
}

func (d *DepositChaincode) juryPayToDepositContract(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	return juryPayToDepositContract(stub, args)
}

func (d *DepositChaincode) developerPayToDepositContract(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	return developerPayToDepositContract(stub, args)
}
func (d *DepositChaincode) mediatorApplyCashback(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	return mediatorApplyCashback(stub, args)
}
func (d *DepositChaincode) juryApplyCashback(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	return juryApplyCashback(stub, args)
}
func (d *DepositChaincode) developerApplyCashback(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	return developerApplyCashback(stub, args)
}

func (d *DepositChaincode) handleForMediatorApplyCashback(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	return handleForMediatorApplyCashback(stub, args)
}

func (d *DepositChaincode) handleForJuryApplyCashback(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	return handleForJuryApplyCashback(stub, args)
}

func (d *DepositChaincode) handleForDeveloperApplyCashback(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	return handleForDeveloperApplyCashback(stub, args)
}

//申请加入Mediator
func (d *DepositChaincode) applyBecomeMediator(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	return applyBecomeMediator(stub, args)
}

//基金会对申请加入Mediator进行处理
func (d *DepositChaincode) handleForApplyBecomeMediator(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	return handleForApplyBecomeMediator(stub, args)
}

//申请退出Mediator
func (d *DepositChaincode) mediatorApplyQuitMediator(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	return mediatorApplyQuitMediator(stub, args)
}

//基金会对申请退出Mediator进行处理
func (d *DepositChaincode) handleForApplyQuitMediator(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	return handleForApplyQuitMediator(stub, args)
}

//对结果序列化并更新数据
func (d *DepositChaincode) marshalForBalance(stub shim.ChaincodeStubInterface, nodeAddr string, balance *modules.DepositBalance) pb.Response {
	balanceByte, err := json.Marshal(balance)
	if err != nil {
		log.Error("Json.Marshal err:", "error", err)
		return shim.Error(err.Error())
	}
	err = stub.PutState(nodeAddr, balanceByte)
	if err != nil {
		log.Error("Stub.PutState err:", "error", err)
		return shim.Error(err.Error())
	}
	return shim.Success([]byte("ok"))
}

func (d *DepositChaincode) handleForForfeitureApplication(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	log.Info("Start entering handleForForfeitureApplication func.")
	//地址，申请时间，是否同意
	if len(args) != 3 {
		log.Error("Arg need three parameters.")
		return shim.Error("Arg need three parameters.")
	}
	//基金会地址
	invokeAddr, err := stub.GetInvokeAddress()
	if err != nil {
		log.Error("Stub.GetInvokeAddress err", "error", err)
		return shim.Error(err.Error())
	}
	//判断没收请求地址是否是基金会地址
	if strings.Compare(invokeAddr, foundationAddress) != 0 {
		log.Error("Please use foundation address.")
		return shim.Error("Please use foundation address.")
	}
	//获取传入参数信息
	addr := args[0]
	applyTimeStr := args[1]
	isOk := args[2]
	//获取一下该用户下的账簿情况
	balance, err := stub.GetDepositBalance(addr)
	if err != nil {
		log.Error("Stub.GetDepositBalance err:", "error", err)
		return shim.Error(err.Error())
	}
	//判断没收节点账户是否为空
	if balance == nil {
		log.Error("Stub.GetDepositBalance: balance is nil.")
		return shim.Error("Stub.GetDepositBalance: balance is nil.")
	}
	//获取申请时间戳
	applyTime, err := strconv.ParseInt(applyTimeStr, 10, 64)
	if err != nil {
		log.Error("Strconv.ParseInt err:", "error", err)
		return shim.Error(err.Error())
	}
	return d.handleForfeitureDepositApplication(stub, invokeAddr, addr, applyTime, balance, isOk)
}

//社区申请没收某节点的保证金数量
func (d DepositChaincode) applyForForfeitureDeposit(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	log.Info("Start entering applyForForfeitureDeposit func.")
	//没收地址 数量 角色 额外说明
	//forfeiture string, invokeTokens modules.InvokeTokens, role, extra string
	if len(args) != 3 {
		log.Error("Arg need three parameters.")
		return shim.Error("Arg need three parameters.")
	}
	//获取参数信息
	forfeitureAddr := args[0]
	amount := args[1]
	forfeitureRole := args[2]
	//申请地址
	invokeAddr, err := stub.GetInvokeAddress()
	if err != nil {
		log.Error("Stub.GetInvokeAddress err:", "error", err)
		return shim.Error(err.Error())
	}

	forfeiture := new(modules.Forfeiture)
	forfeiture.ApplyAddress = invokeAddr
	//判断被没收时，该节点是否在相应的候选列表当中
	isFound := isFoundInCandidateList(stub, forfeitureRole, forfeitureAddr)
	if !isFound {
		log.Error("Node is not exist in the candidate list.")
		return shim.Error("Node is not exist in the candidate list.")
	}
	//获取没收保证金数量，将 string 转 uint64
	ptnAccount, err := strconv.ParseUint(amount, 10, 64)
	if err != nil {
		log.Error("Strconv.ParseUint err:", "error", err)
		return shim.Error(err.Error())
	}
	//fmt.Println("ptnAccount  args[1] ", ptnAccount)
	//判断账户余额和没收请求数量
	//if balanceValue.TotalAmount < ptnAccount {
	//	return shim.Success([]byte("Forfeiture too many."))
	//}
	forfeiture.ForfeitureAddress = forfeitureAddr
	asset := modules.NewPTNAsset()
	invokeTokens := &modules.InvokeTokens{
		Amount: ptnAccount,
		Asset:  asset,
	}
	forfeiture.ApplyTokens = invokeTokens
	forfeiture.ForfeitureRole = forfeitureRole
	forfeiture.ApplyTime = time.Now().UTC().Unix()
	//先获取列表，再更新列表
	listForForfeiture, err := stub.GetListForForfeiture()
	if err != nil {
		log.Error("Stub.GetListForForfeiture err:", "error", err)
		return shim.Error(err.Error())
	}
	if listForForfeiture == nil {
		log.Info("Stub.GetListForForfeiture:list is nil.")
		listForForfeiture = []*modules.Forfeiture{forfeiture}
	} else {
		isExist := isInForfeiturelist(forfeiture.ForfeitureAddress, listForForfeiture)
		if isExist {
			log.Error("Node is exist in the forfeiture list.")
			return shim.Error("Node is exist in the forfeiture list.")
		}
		listForForfeiture = append(listForForfeiture, forfeiture)
	}
	listForForfeitureByte, err := json.Marshal(listForForfeiture)
	if err != nil {
		log.Error("Json.Marshal err:", "error", err)
		return shim.Error(err.Error())
	}
	err = stub.PutState("ListForForfeiture", listForForfeitureByte)
	if err != nil {
		log.Error("Stub.PutState err:", "error", err)
		return shim.Error(err.Error())
	}
	log.Info("End entering applyForForfeitureDeposit func.")
	return shim.Success([]byte("ok"))
}

func isFoundInCandidateList(stub shim.ChaincodeStubInterface, role string, addr string) bool {
	if strings.Compare(role, "Mediator") == 0 {
		candidateList, err := stub.GetCandidateListForMediator()
		if err != nil {
			return false
		}
		if candidateList == nil {
			return false
		}
		return isInMediatorInfolist(addr, candidateList)

	} else if strings.Compare(role, "Jury") == 0 {
		candidateList, err := stub.GetCandidateList("JuryList")
		if err != nil {
			return false
		}
		if candidateList == nil {
			return false
		}
		return isInCandidateList(addr, candidateList)
	} else if strings.Compare(role, "Developer") == 0 {
		candidateList, err := stub.GetCandidateList("DeveloperList")
		if err != nil {
			return false
		}
		if candidateList == nil {
			return false
		}
		return isInCandidateList(addr, candidateList)
	} else {
		return false
	}
}

func isInCandidateList(addr string, list []string) bool {
	for _, a := range list {
		if strings.Compare(addr, a) == 0 {
			return true
		}
	}
	return false
}

//查找节点是否在列表中
func isInForfeiturelist(addr string, list []*modules.Forfeiture) bool {
	for _, m := range list {
		if strings.Compare(addr, m.ForfeitureAddress) == 0 {
			return true
		}
	}
	return false
}

//基金会处理没收请求
func (d *DepositChaincode) handleForfeitureDepositApplication(stub shim.ChaincodeStubInterface, foundationAddr, forfeitureAddr string, applyTime int64, balance *modules.DepositBalance, check string) pb.Response {
	//check 如果为ok，则同意此申请，如果为no，则不同意此申请
	if strings.Compare(check, "ok") == 0 {
		return d.agreeForApplyForfeiture(stub, foundationAddr, forfeitureAddr, applyTime, balance)
	} else if strings.Compare(check, "no") == 0 {
		//移除申请列表，不做处理
		return d.disagreeForApplyForfeiture(stub, forfeitureAddr, applyTime)
	} else {
		log.Error("Please enter ok or no.")
		return shim.Error("Please enter ok or no.")
	}
}

//不同意提取请求，则直接从提保证金列表中移除该节点
//func (d *DepositChaincode) disagreeForApplyCashback(stub shim.ChaincodeStubInterface, cashbackAddr string, applyTime int64) pb.Response {
//	//获取没收列表
//	listForCashback, err := stub.GetListForCashback()
//	if err != nil {
//		log.Error("Stub.GetListForCashback err:", "error", err)
//		return shim.Error(err.Error())
//	}
//	if listForCashback == nil {
//		log.Error("Stub.GetListForCashback err: list is nil.")
//		return shim.Error("Stub.GetListForCashback err: list is nil.")
//	}
//	isExist := isInCashbacklist(cashbackAddr, listForCashback)
//	if !isExist {
//		log.Error("Node is not exist in the cashback list.")
//		return shim.Error("Node is not exist in the cashback list.")
//	}
//	//fmt.Println("moveInApplyForCashbackList==>", listForCashback)
//	newList := moveInApplyForCashbackList(stub, listForCashback, cashbackAddr, applyTime)
//	listForCashbackByte, err := json.Marshal(newList)
//	if err != nil {
//		log.Error("Json.Marshal err:", "error", err)
//		return shim.Error(err.Error())
//	}
//	//更新列表
//	err = stub.PutState("ListForCashback", listForCashbackByte)
//	if err != nil {
//		log.Error("Stub.PutState err:", "error", err)
//		return shim.Error(err.Error())
//	}
//	//fmt.Println("moveInApplyForCashbackList==>", listForCashback)
//	return shim.Success([]byte("ok"))
//}

//不同意这样没收请求，则直接从没收列表中移除该节点
func (d *DepositChaincode) disagreeForApplyForfeiture(stub shim.ChaincodeStubInterface, forfeiture string, applyTime int64) pb.Response {
	//获取没收列表
	listForForfeiture, err := stub.GetListForForfeiture()
	if err != nil {
		log.Error("Stub.GetListForForfeiture err:", "error", err)
		return shim.Error(err.Error())
	}
	if listForForfeiture == nil {
		log.Error("Stub.GetListForForfeiture:list is nil.")
		return shim.Error("Stub.GetListForForfeiture:list is nil.")
	}
	//判断是否在列表中
	isExist := isInForfeiturelist(forfeiture, listForForfeiture)
	if !isExist {
		log.Error("Node is not exist in the forfeiture list.")
		return shim.Error("Node is not exist in the forfeiture list.")
	}
	//从列表中移除
	newList, isFound := moveInApplyForForfeitureList(stub, listForForfeiture, forfeiture, applyTime)
	if !isFound {
		log.Error("Apply time is wrong.")
		return shim.Error("Apply time is wrong.")
	}
	//
	listForForfeitureByte, err := json.Marshal(newList)
	if err != nil {
		log.Error("Json.Marshal err:", "error", err)
		return shim.Error(err.Error())
	}
	//更新列表
	err = stub.PutState("ListForForfeiture", listForForfeitureByte)
	if err != nil {
		log.Error("Stub.PutState err:", "error", err)
		return shim.Error(err.Error())
	}
	return shim.Success([]byte("ok"))
}

//同意申请没收请求
func (d *DepositChaincode) agreeForApplyForfeiture(stub shim.ChaincodeStubInterface, foundationAddr, forfeitureAddr string, applyTime int64, balance *modules.DepositBalance) pb.Response {
	log.Info("Start entering agreeForApplyForfeiture func.")
	//获取列表
	listForForfeiture, err := stub.GetListForForfeiture()
	if err != nil {
		log.Error("Stub.GetListForForfeiture err:", "error", err)
		return shim.Error(err.Error())
	}
	if listForForfeiture == nil {
		log.Error("Stub.GetListForForfeiture err:list is nil.")
		return shim.Error("Stub.GetListForForfeiture err:list is nil.")
	}
	//判断是否在列表中
	isExist := isInForfeiturelist(forfeitureAddr, listForForfeiture)
	if !isExist {
		log.Error("Node is not exist in the forfeiture list.")
		return shim.Error("Node is not exist in the forfeiture list.")
	}
	//获取节点信息
	forfeiture := &modules.Forfeiture{}
	isFound := false
	for _, m := range listForForfeiture {
		if m.ForfeitureAddress == forfeitureAddr && m.ApplyTime == applyTime {
			forfeiture = m
			isFound = true
			break
		}
	}
	if !isFound {
		log.Error("Apply time is wrong.")
		return shim.Error("Apply time is wrong.")
	}
	//在列表中移除，并获取没收情况
	newList, _ := moveInApplyForForfeitureList(stub, listForForfeiture, forfeitureAddr, applyTime)
	listForForfeitureByte, err := json.Marshal(newList)
	if err != nil {
		log.Error("Json.Marshal err:", "error", err)
		return shim.Error(err.Error())
	}
	//更新列表
	err = stub.PutState("ListForForfeiture", listForForfeitureByte)
	if err != nil {
		log.Error("Stub.PutState err:", "error", err)
		return shim.Error(err.Error())
	}
	//判断余额
	if forfeiture.ApplyTokens.Amount > balance.TotalAmount {
		log.Error("The balance is not enough.")
		return shim.Error("The balance is not enough.")
	}
	//判断节点类型
	switch {
	case forfeiture.ForfeitureRole == "Mediator":
		return d.handleMediatorForfeitureDeposit(foundationAddr, forfeiture, balance, stub)
	case forfeiture.ForfeitureRole == "Jury":
		return d.handleJuryForfeitureDeposit(foundationAddr, forfeiture, balance, stub)
	case forfeiture.ForfeitureRole == "Developer":
		return d.handleDeveloperForfeitureDeposit(foundationAddr, forfeiture, balance, stub)
	default:
		return shim.Error("role error")
	}
}

//处理申请没收请求并移除列表
func (d *DepositChaincode) forfeitureAllDeposit(role string, stub shim.ChaincodeStubInterface, foundationAddr, forfeitureAddr string, invokeTokens *modules.InvokeTokens) error {
	//TODO 没收保证金是否需要计算利息
	//调用从合约把token转到请求地址
	err := stub.PayOutToken(foundationAddr, invokeTokens, 0)
	if err != nil {
		log.Error("Stub.PayOutToken err:", "error", err)
		return err
	}
	//移除出列表
	err = moveCandidate(role, forfeitureAddr, stub)
	if err != nil {
		log.Error("MoveCandidate err:", "error", err)
		return err
	}
	//删除节点
	err = stub.DelState(forfeitureAddr)
	if err != nil {
		log.Error("Stub.DelState err:", "error", err)
		return err
	}
	return nil
}

//处理没收Mediator保证金
func (d *DepositChaincode) handleMediatorForfeitureDeposit(foundationAddr string, forfeiture *modules.Forfeiture, balance *modules.DepositBalance, stub shim.ChaincodeStubInterface) pb.Response {
	var err error
	//计算余额
	result := balance.TotalAmount - forfeiture.ApplyTokens.Amount
	//判断是否没收全部，即在列表中移除该节点
	if result == 0 {
		//没收不考虑是否在规定周期内,其实它肯定是在列表中并已在周期内
		//没收全部，即删除,已经是计算好奖励了
		err = d.forfeitureAllDeposit("MediatorList", stub, foundationAddr, forfeiture.ForfeitureAddress, forfeiture.ApplyTokens)
		if err != nil {
			log.Error("ForfeitureAllDeposit err:", "error", err)
			return shim.Error(err.Error())
		}
		return shim.Success([]byte("ok"))
	} else {
		//TODO 对于mediator，要么全没收，要么退出一部分，且退出该部分金额后还在列表中
		return d.forfeitureSomeDeposit("Mediator", stub, foundationAddr, forfeiture, balance)
	}
}

func (d *DepositChaincode) forfertureAndMoveList(role string, stub shim.ChaincodeStubInterface, foundationAddr string, forfeiture *modules.Forfeiture, balance *modules.DepositBalance) pb.Response {
	//调用从合约把token转到请求地址
	err := stub.PayOutToken(foundationAddr, forfeiture.ApplyTokens, 0)
	if err != nil {
		log.Error("Stub.PayOutToken err:", "error", err)
		return shim.Error(err.Error())
	}
	//handleMember(role, forfeiture.ForfeitureAddress, stub)
	err = moveCandidate(role, forfeiture.ForfeitureAddress, stub)
	if err != nil {
		log.Error("MoveCandidate err:", "error", err)
		return shim.Error(err.Error())
	}
	//计算一部分的利息
	//获取币龄
	awards := award.GetAwardsWithCoins(balance.TotalAmount, balance.LastModifyTime.Unix())
	//fmt.Println("awards ", awards)
	balance.LastModifyTime = time.Now().UTC()
	//加上利息奖励
	balance.TotalAmount += awards
	//减去提取部分
	balance.TotalAmount -= forfeiture.ApplyTokens.Amount

	balance.ForfeitureValues = append(balance.ForfeitureValues, forfeiture)

	//序列化
	return d.marshalForBalance(stub, forfeiture.ForfeitureAddress, balance)
}

//不需要移除候选列表，但是要没收一部分保证金
func (d *DepositChaincode) forfeitureSomeDeposit(role string, stub shim.ChaincodeStubInterface, foundationAddr string, forfeiture *modules.Forfeiture, balance *modules.DepositBalance) pb.Response {
	//调用从合约把token转到请求地址
	err := stub.PayOutToken(foundationAddr, forfeiture.ApplyTokens, 0)
	if err != nil {
		log.Error("Stub.PayOutToken err:", "error", err)
		return shim.Error(err.Error())
	}
	//计算当前币龄奖励
	awards := award.GetAwardsWithCoins(balance.TotalAmount, balance.LastModifyTime.Unix())
	//fmt.Println("awards ", awards)
	balance.LastModifyTime = time.Now().UTC()
	//加上利息奖励
	balance.TotalAmount += awards
	//减去提取部分
	balance.TotalAmount -= forfeiture.ApplyTokens.Amount

	balance.ForfeitureValues = append(balance.ForfeitureValues, forfeiture)

	//序列化
	return d.marshalForBalance(stub, forfeiture.ForfeitureAddress, balance)
}

func (d *DepositChaincode) handleJuryForfeitureDeposit(foundationAddr string, forfeiture *modules.Forfeiture, balance *modules.DepositBalance, stub shim.ChaincodeStubInterface) pb.Response {
	var err error
	//计算余额
	result := balance.TotalAmount - forfeiture.ApplyTokens.Amount
	//判断是否没收全部，即在列表中移除该节点
	if result == 0 {
		//没收不考虑是否在规定周期内,其实它肯定是在列表中并已在周期内
		//没收全部，即删除
		err = d.forfeitureAllDeposit("JuryList", stub, foundationAddr, forfeiture.ForfeitureAddress, forfeiture.ApplyTokens)
		if err != nil {
			log.Error("ForfeitureAllDeposit err:", "error", err)
			return shim.Error(err.Error())
		}
		return shim.Success([]byte("ok"))
	} else if result < depositAmountsForJury {
		//TODO 对于jury，需要移除列表
		return d.forfertureAndMoveList("JuryList", stub, foundationAddr, forfeiture, balance)
	} else {
		//TODO 退出一部分，且退出该部分金额后还在列表中
		return d.forfeitureSomeDeposit("Jury", stub, foundationAddr, forfeiture, balance)
	}
}

func (d *DepositChaincode) handleDeveloperForfeitureDeposit(foundationAddr string, forfeiture *modules.Forfeiture, balance *modules.DepositBalance, stub shim.ChaincodeStubInterface) pb.Response {
	var err error
	//计算余额
	result := balance.TotalAmount - forfeiture.ApplyTokens.Amount
	//判断是否没收全部，即在列表中移除该节点
	if result == 0 {
		//没收不考虑是否在规定周期内,其实它肯定是在列表中并已在周期内
		//没收全部，即删除
		err = d.forfeitureAllDeposit("DeveloperList", stub, foundationAddr, forfeiture.ForfeitureAddress, forfeiture.ApplyTokens)
		if err != nil {
			log.Error("ForfeitureAllDeposit err:", "error", err)
			return shim.Error(err.Error())
		}
		return shim.Success([]byte("ok"))
	} else if result < depositAmountsForDeveloper {
		return d.forfertureAndMoveList("DeveloperList", stub, foundationAddr, forfeiture, balance)
	} else {
		//TODO 退出一部分，且退出该部分金额后还在列表中
		return d.forfeitureSomeDeposit("Developer", stub, foundationAddr, forfeiture, balance)
	}
}
