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
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/contracts/shim"
	pb "github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
	"strconv"
	"strings"
)

//同意申请没收请求
func (d *DepositChaincode) agreeForApplyForfeiture(stub shim.ChaincodeStubInterface, foundationAddr, forfeitureAddr string, applyTime int64, balance *DepositBalance) pb.Response {
	log.Info("Start entering agreeForApplyForfeiture func.")
	//获取列表
	listForForfeiture, err := GetListForForfeiture(stub)
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
	forfeiture := &Forfeiture{}
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
	newList, _ := moveInApplyForForfeitureList(listForForfeiture, forfeitureAddr, applyTime)
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
		return shim.Error("Please enter validate role.")
	}
}

//处理没收Mediator保证金
func (d *DepositChaincode) handleMediatorForfeitureDeposit(foundationAddr string, forfeiture *Forfeiture, balance *DepositBalance, stub shim.ChaincodeStubInterface) pb.Response {
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

func (d *DepositChaincode) handleJuryForfeitureDeposit(foundationAddr string, forfeiture *Forfeiture, balance *DepositBalance, stub shim.ChaincodeStubInterface) pb.Response {
	depositAmountsForJuryStr, err := stub.GetSystemConfig("DepositAmountForJury")
	if err != nil {
		log.Error("Stub.GetSystemConfig with DepositAmountForJury err:", "error", err)
		return shim.Error(err.Error())
	}
	//转换
	depositAmountsForJury, err := strconv.ParseUint(depositAmountsForJuryStr, 10, 64)
	if err != nil {
		log.Error("Strconv.ParseUint err:", "error", err)
		return shim.Error(err.Error())
	}
	log.Info("Stub.GetSystemConfig with DepositAmountForJury:", "value", depositAmountsForJury)
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

func (d *DepositChaincode) handleDeveloperForfeitureDeposit(foundationAddr string, forfeiture *Forfeiture, balance *DepositBalance, stub shim.ChaincodeStubInterface) pb.Response {
	depositAmountsForDeveloperStr, err := stub.GetSystemConfig("DepositAmountForDeveloper")
	if err != nil {
		log.Error("Stub.GetSystemConfig with DepositAmountForDeveloper err:", "error", err)
		return shim.Error(err.Error())
	}
	//转换
	depositAmountsForDeveloper, err := strconv.ParseUint(depositAmountsForDeveloperStr, 10, 64)
	if err != nil {
		log.Error("Strconv.ParseUint err:", "error", err)
		return shim.Error(err.Error())
	}
	log.Info("Stub.GetSystemConfig with DepositAmountForDeveloper:", "value", depositAmountsForDeveloper)
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

//不同意这样没收请求，则直接从没收列表中移除该节点
func (d *DepositChaincode) disagreeForApplyForfeiture(stub shim.ChaincodeStubInterface, forfeiture string, applyTime int64) pb.Response {
	//获取没收列表
	listForForfeiture, err := GetListForForfeiture(stub)
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
	newList, isFound := moveInApplyForForfeitureList(listForForfeiture, forfeiture, applyTime)
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

//基金会处理没收请求
func (d *DepositChaincode) handleForfeitureDepositApplication(stub shim.ChaincodeStubInterface, foundationAddr, forfeitureAddr string, applyTime int64, balance *DepositBalance, check string) pb.Response {
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
	foundationAddress, err := stub.GetSystemConfig("FoundationAddress")
	if err != nil {
		//fmt.Println(err.Error())
		log.Error("Stub.GetSystemConfig with FoundationAddress err:", "error", err)
		return shim.Error(err.Error())
	}
	//foundationAddress = "P129MFVxaLP4N9FZxYQJ3QPJks4gCeWsF9p"
	log.Info("Stub.GetSystemConfig with FoundationAddress:", "value", foundationAddress)
	//判断没收请求地址是否是基金会地址
	if strings.Compare(invokeAddr.String(), foundationAddress) != 0 {
		log.Error("Please use foundation address.")
		return shim.Error("Please use foundation address.")
	}
	//获取传入参数信息
	addr := args[0]
	applyTimeStr := args[1]
	isOk := args[2]
	//获取一下该用户下的账簿情况
	balance, err := GetDepositBalance(stub, addr)
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
	return d.handleForfeitureDepositApplication(stub, invokeAddr.String(), addr, applyTime, balance, isOk)
}

//基金会处理
func handleForDeveloperApplyCashback(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	//地址，申请时间，是否同意
	if len(args) != 3 {
		log.Error("Args need three parameters.")
		return shim.Error("Args need three parameters.")
	}
	//基金会地址
	invokeAddr, err := stub.GetInvokeAddress()
	if err != nil {
		log.Error("Stub.GetInvokeAddress err:", "error", err)
		return shim.Error(err.Error())
	}
	//判断没收请求地址是否是基金会地址
	foundationAddress, err := stub.GetSystemConfig("FoundationAddress")
	if err != nil {
		//fmt.Println(err.Error())
		log.Error("Stub.GetSystemConfig with FoundationAddress err:", "error", err)
		return shim.Error(err.Error())
	}
	//foundationAddress = "P129MFVxaLP4N9FZxYQJ3QPJks4gCeWsF9p"
	log.Info("Stub.GetSystemConfig with FoundationAddress:", "value", foundationAddress)
	if strings.Compare(invokeAddr.String(), foundationAddress) != 0 {
		log.Error("Please use foundation address.")
		return shim.Error("Please use foundation address.")
	}
	//获取一下该用户下的账簿情况
	addr := args[0]
	balance, err := GetDepositBalance(stub, addr)
	if err != nil {
		log.Error("Stub.GetDepositBalance err:", "error", err)
		return shim.Error(err.Error())
	}
	//判断没收节点账户是否为空
	if balance == nil {
		log.Error("Balance is nil.")
		return shim.Error("Balance is nil.")
	}
	//获取申请时间戳
	strTime := args[1]
	applyTime, err := strconv.ParseInt(strTime, 10, 64)
	if err != nil {
		log.Error("Strconv.ParseInt err", "error", err)
		return shim.Error(err.Error())
	}
	isOk := args[2]
	if strings.Compare(isOk, "ok") == 0 {
		//对余额处理
		err = handleDeveloper(stub, addr, applyTime, balance)
		if err != nil {
			log.Error("handleDeveloper err", "error", err)
			return shim.Error(err.Error())
		}
	} else if strings.Compare(isOk, "no") == 0 {
		//移除提取申请列表
		err = moveAndPutStateFromCashbackList(stub, addr, applyTime)
		if err != nil {
			log.Error("moveAndPutStateFromCashbackList err", "error", err)
			return shim.Error(err.Error())
		}
	} else {
		log.Error("please enter ok or no.")
		return shim.Error("please enter ok or no.")
	}
	return shim.Success([]byte("ok"))
}

//对Developer退保证金的处理
func handleDeveloperDepositCashback(stub shim.ChaincodeStubInterface, cashbackAddr string, cashbackValue *Cashback, balance *DepositBalance) error {
	depositAmountsForDeveloperStr, err := stub.GetSystemConfig("DepositAmountForDeveloper")
	if err != nil {
		log.Error("Stub.GetSystemConfig with DepositAmountForDeveloper err:", "error", err)
		return err
	}
	//转换
	depositAmountsForDeveloper, err := strconv.ParseUint(depositAmountsForDeveloperStr, 10, 64)
	if err != nil {
		log.Error("Strconv.ParseUint err:", "error", err)
		return err
	}
	log.Info("Stub.GetSystemConfig with DepositAmountForDeveloper:", "value", depositAmountsForDeveloper)
	if balance.TotalAmount >= depositAmountsForDeveloper {
		//已在列表中
		err := handleDeveloperFromList(stub, cashbackAddr, cashbackValue, balance)
		if err != nil {
			log.Error("HandleDeveloperFromList err:", "error", err)
			return err
		}
	} else {
		////TODO 不在列表中,没有奖励，直接退
		err := handleCommonJuryOrDev(stub, cashbackAddr, cashbackValue, balance)
		if err != nil {
			log.Error("handleCommonJuryOrDev err:", "error", err)
			return err
		}
	}
	return nil
}

//基金会处理
func handleForJuryApplyCashback(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	//地址，申请时间，是否同意
	if len(args) != 3 {
		log.Error("Args need three parameters.")
		return shim.Error("Args need three parameters.")
	}
	//基金会地址
	invokeAddr, err := stub.GetInvokeAddress()
	if err != nil {
		log.Error("Stub.GetInvokeAddress err:", "error", err)
		return shim.Error(err.Error())
	}
	//判断没收请求地址是否是基金会地址
	foundationAddress, err := stub.GetSystemConfig("FoundationAddress")
	if err != nil {
		//fmt.Println(err.Error())
		log.Error("Stub.GetSystemConfig with FoundationAddress err:", "error", err)
		return shim.Error(err.Error())
	}
	//foundationAddress = "P129MFVxaLP4N9FZxYQJ3QPJks4gCeWsF9p"
	log.Info("Stub.GetSystemConfig with FoundationAddress:", "value", foundationAddress)
	if strings.Compare(invokeAddr.String(), foundationAddress) != 0 {
		log.Error("Please use foundation address.")
		return shim.Error("Please use foundation address.")
	}
	//获取一下该用户下的账簿情况
	addr := args[0]
	balance, err := GetDepositBalance(stub, addr)
	if err != nil {
		log.Error("Stub.GetDepositBalance err:", "error", err)
		return shim.Error(err.Error())
	}
	//判断没收节点账户是否为空
	if balance == nil {
		log.Error("Balance is nil.")
		return shim.Error("Balance is nil.")
	}
	//获取申请时间戳
	strTime := args[1]
	applyTime, err := strconv.ParseInt(strTime, 10, 64)
	if err != nil {
		log.Error("Strconv.ParseInt err", "error", err)
		return shim.Error(err.Error())
	}
	isOk := args[2]
	if strings.Compare(isOk, "ok") == 0 {
		//对余额处理
		err = handleJury(stub, addr, applyTime, balance)
		if err != nil {
			log.Error("handleJury err", "error", err)
			return shim.Error(err.Error())
		}
	} else if strings.Compare(isOk, "no") == 0 {
		//移除提取申请列表
		err = moveAndPutStateFromCashbackList(stub, addr, applyTime)
		if err != nil {
			log.Error("moveAndPutStateFromCashbackList err", "error", err)
			return shim.Error(err.Error())
		}
	} else {
		log.Error("Please enter ok or no.")
		return shim.Error("Please enter ok or no.")
	}
	return shim.Success([]byte("ok"))
}

//对Jury退保证金的处理
func handleJuryDepositCashback(stub shim.ChaincodeStubInterface, cashbackAddr string, cashbackValue *Cashback, balance *DepositBalance) error {
	depositAmountsForJuryStr, err := stub.GetSystemConfig("DepositAmountForJury")
	if err != nil {
		log.Error("Stub.GetSystemConfig with DepositAmountForJury err:", "error", err)
		return err
	}
	//转换
	depositAmountsForJury, err := strconv.ParseUint(depositAmountsForJuryStr, 10, 64)
	if err != nil {
		log.Error("Strconv.ParseUint err:", "error", err)
		return err
	}
	log.Info("Stub.GetSystemConfig with DepositAmountForJury:", "value", depositAmountsForJury)
	if balance.TotalAmount >= depositAmountsForJury {
		//已在列表中
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
	log.Info("Start entering handleForApplyBecomeMediator func")
	invokeAddr, err := stub.GetInvokeAddress()
	if err != nil {
		log.Error("Stub.GetInvokeAddress err:", "error", err)
		return shim.Error(err.Error())
	}
	foundationAddress, err := stub.GetSystemConfig("FoundationAddress")
	if err != nil {
		//fmt.Println(err.Error())
		log.Error("Stub.GetSystemConfig with FoundationAddress err:", "error", err)
		return shim.Error(err.Error())
	}
	//foundationAddress = "P129MFVxaLP4N9FZxYQJ3QPJks4gCeWsF9p"
	log.Info("Stub.GetSystemConfig with FoundationAddress:", "value", foundationAddress)
	if strings.Compare(invokeAddr.String(), foundationAddress) != 0 {
		log.Error("Please use foundation address.")
		return shim.Error("Please use foundation address.")
	}
	if len(args) != 2 {
		log.Error("Arg need two parameters.")
		return shim.Error("arg need two parameters.")
	}
	//获取申请列表
	becomeList, err := GetBecomeMediatorApplyList(stub)
	if err != nil {
		log.Error("Stub.GetBecomeMediatorApplyList err:", "error", err)
		return shim.Error(err.Error())
	}
	if becomeList == nil {
		log.Error("Stub.GetBecomeMediatorApplyList: list is nil")
		return shim.Error("Become list is nil.")
	}
	isOk := args[0]
	addr := args[1]
	isExist := isInMediatorInfolist(addr, becomeList)
	if !isExist {
		log.Error("Node is not exist in the become list.")
		return shim.Error("Node is not exist in the become list.")
	}
	//var mediatorList []*MediatorInfo
	mediator := &MediatorRegisterInfo{}
	//不同意，移除申请列表
	if strings.Compare(isOk, "no") == 0 {
		log.Info("foundation is not agree with application.")
		becomeList, _ = moveMediatorFromList(addr, becomeList)
	} else if strings.Compare(isOk, "ok") == 0 {
		log.Info("foundation is agree with application.")
		//同意，移除列表，并且加入同意申请列表
		becomeList, mediator = moveMediatorFromList(addr, becomeList)
		//获取同意列表
		agreeList, err := GetAgreeForBecomeMediatorList(stub)
		if err != nil {
			log.Error("Stub.GetAgreeForBecomeMediatorList err:", "error", err)
			return shim.Error(err.Error())
		}
		if agreeList == nil {
			log.Info("Stub.GetAgreeForBecomeMediatorList: list is nil")
			agreeList = []*MediatorRegisterInfo{mediator}
		} else {
			isExist := isInMediatorInfolist(mediator.Address, agreeList)
			if isExist {
				log.Error("Node is exist in the agree list.")
				return shim.Error("Node is exist in the agree list.")
			}
			agreeList = append(agreeList, mediator)
		}
		err = marshalAndPutStateForMediatorList(stub, "ListForAgreeBecomeMediator", agreeList)
		if err != nil {
			log.Error("MarshalAndPutStateForMediatorList err:", "error", err)
			return shim.Error(err.Error())
		}
	} else {
		log.Error("Please enter ok or no.")
		return shim.Error("Please enter ok or no.")
	}
	err = marshalAndPutStateForMediatorList(stub, "ListForApplyBecomeMediator", becomeList)
	if err != nil {
		log.Error("MarshalAndPutStateForMediatorList err:", "error", err)
		return shim.Error(err.Error())
	}
	log.Info("End entering handleForApplyBecomeMediator func")
	return shim.Success([]byte("ok"))
}

//处理退出 参数：同意或不同意，节点的地址
func handleForApplyQuitMediator(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	log.Info("Start enter handleForApplyQuitMediator func")
	//基金会
	invokeAddr, err := stub.GetInvokeAddress()
	if err != nil {
		log.Error("Stub.GetInvokeAddress err:", "error", err)
		return shim.Error(err.Error())
	}
	foundationAddress, err := stub.GetSystemConfig("FoundationAddress")
	if err != nil {
		//fmt.Println(err.Error())
		log.Error("Stub.GetSystemConfig with FoundationAddress err:", "error", err)
		return shim.Error(err.Error())
	}
	//foundationAddress = "P129MFVxaLP4N9FZxYQJ3QPJks4gCeWsF9p"
	log.Info("Stub.GetSystemConfig with FoundationAddress:", "value", foundationAddress)
	if strings.Compare(invokeAddr.String(), foundationAddress) != 0 {
		log.Error("Please use foundation address.")
		return shim.Success([]byte("Please use foundation address."))
	}
	//参数
	if len(args) != 2 {
		log.Error("Arg need two parameter.")
		return shim.Error("Arg need two parameter.")
	}
	//获取申请列表
	quitList, err := GetQuitMediatorApplyList(stub)
	if err != nil {
		log.Error("Stub.GetQuitMediatorApplyList err:", "error", err)
		return shim.Error(err.Error())
	}
	if quitList == nil {
		log.Error("Stub.GetQuitMediatorApplyList err:list is nil.")
		return shim.Error("Stub.GetQuitMediatorApplyList err:list is nil.")
	}
	//
	isOk := args[0]
	addr := args[1]
	isExist := isInMediatorInfolist(addr, quitList)
	if !isExist {
		log.Error("Node is not exist in the quit list.")
		return shim.Error("Node is not exist in the quit list.")
	}
	//var mediatorList []*MediatorInfo
	//不同意，移除申请列表
	if strings.Compare(isOk, "no") == 0 {
		log.Info("foundation is not agree with application.")
		quitList, _ = moveMediatorFromList(addr, quitList)
	} else if strings.Compare(isOk, "ok") == 0 {
		log.Info("foundation is agree with application.")
		//同意，移除列表，并且全款退出
		quitList, _ = moveMediatorFromList(addr, quitList)
		//获取该账户
		balance, err := GetDepositBalance(stub, addr)
		if err != nil {
			log.Error("Stub.GetDepositBalance err:", "error", err)
			return shim.Error(err.Error())
		}
		if balance == nil {
			log.Error("Stub.GetDepositBalance err: balance is nil.")
			return shim.Error("Stub.GetDepositBalance err: balance is nil.")
		}
		err = deleteNode(stub, balance, addr)
		if err != nil {
			log.Error("DeleteNode err:", "error", err)
			return shim.Error(err.Error())
		}
		//从同意列表中删除
		agreeList, err := GetAgreeForBecomeMediatorList(stub)
		if err != nil {
			log.Error("Stub.GetAgreeForBecomeMediatorList err:", "error", err)
			return shim.Error(err.Error())
		}
		if agreeList == nil {
			log.Error("Stub.GetAgreeForBecomeMediatorList err: list is nil")
			shim.Error("Stub.GetAgreeForBecomeMediatorList err: list is nil")
		}
		isExist = isInMediatorInfolist(addr, agreeList)
		if !isExist {
			log.Error("Node is not exist in the agree list.")
			shim.Error("Node is not exist in the agree list.")
		}
		//移除
		agreeList, _ = moveMediatorFromList(addr, agreeList)
		err = marshalAndPutStateForMediatorList(stub, "ListForAgreeBecomeMediator", agreeList)
		if err != nil {
			log.Error("MarshalAndPutStateForMediatorList err:", "error", err)
			return shim.Error(err.Error())
		}
	} else {
		log.Error("Please enter ok or no.")
		return shim.Error("Please enter ok or no.")
	}
	err = marshalAndPutStateForMediatorList(stub, "ListForApplyQuitMediator", quitList)
	if err != nil {
		log.Error("MarshalAndPutStateForMediatorList err:", "error", err)
		return shim.Error(err.Error())
	}
	log.Info("End entering handleForApplyQuitMediator func.")
	return shim.Success([]byte("ok"))
}

//基金会处理
func handleForMediatorApplyCashback(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	log.Info("Start entering handleForMediatorApplyCashback func.")
	//地址，申请时间，是否同意
	if len(args) != 3 {
		log.Error("Arg need three parameters.")
		return shim.Error("Arg need three parameters.")
	}
	//基金会地址
	invokeAddr, err := stub.GetInvokeAddress()
	if err != nil {
		log.Error("Stub.GetInvokeAddress err:", "error", err)
		return shim.Error(err.Error())
	}
	//判断没收请求地址是否是基金会地址
	foundationAddress, err := stub.GetSystemConfig("FoundationAddress")
	if err != nil {
		//fmt.Println(err.Error())
		log.Error("Stub.GetSystemConfig with FoundationAddress err:", "error", err)
		return shim.Error(err.Error())
	}
	//foundationAddress = "P129MFVxaLP4N9FZxYQJ3QPJks4gCeWsF9p"
	log.Info("Stub.GetSystemConfig with FoundationAddress:", "value", foundationAddress)
	if strings.Compare(invokeAddr.String(), foundationAddress) != 0 {
		log.Error("Please use foundation address.")
		return shim.Error("Please use foundation address.")
	}
	//获取一下该用户下的账簿情况
	addr := args[0]
	balance, err := GetDepositBalance(stub, addr)
	if err != nil {
		log.Error("Stub.GetDepositBalance err: ", "error", err)
		return shim.Error(err.Error())
	}
	//判断没收节点账户是否为空
	if balance == nil {
		log.Error("Stub.GetDepositBalance err: balance is nil.")
		return shim.Error("Stub.GetDepositBalance err: balance is nil.")
	}
	//获取申请时间戳
	strTime := args[1]
	applyTime, err := strconv.ParseInt(strTime, 10, 64)
	if err != nil {
		log.Error("Strconv.ParseInt err:", "error", err)
		return shim.Error(err.Error())
	}
	isOk := args[2]
	if strings.Compare(isOk, "ok") == 0 {
		//对余额处理
		err = handleMediator(stub, addr, applyTime, balance)
		if err != nil {
			log.Error("HandleMediator err:", "error", err)
			return shim.Error(err.Error())
		}
	} else if strings.Compare(isOk, "no") == 0 {
		//移除提取申请列表
		err = moveAndPutStateFromCashbackList(stub, addr, applyTime)
		if err != nil {
			log.Error("MoveAndPutStateFromCashbackList err:", "error", err)
			return shim.Error(err.Error())
		}
	} else {
		log.Error("Please enter ok or no.")
		return shim.Error("Please enter ok or no.")
	}
	log.Info("End entering handleForMediatorApplyCashback func.")
	return shim.Success([]byte("ok"))
}
