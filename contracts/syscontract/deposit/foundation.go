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
	"strconv"
	"strings"
	"time"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/contracts/shim"
	pb "github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
	"github.com/palletone/go-palletone/dag/modules"
)

//同意申请没收请求
func agreeForApplyForfeiture(stub shim.ChaincodeStubInterface, foundationAddr,
	forfeitureAddr string, balance *DepositBalance) error {
	log.Info("Start entering agreeForApplyForfeiture func.")
	//获取列表
	listForForfeiture, err := GetListForForfeiture(stub)
	if err != nil {
		log.Error("Stub.GetListForForfeiture err:", "error", err)
		return err
	}
	if listForForfeiture == nil {
		log.Error("Stub.GetListForForfeiture err:list is nil.")
		return fmt.Errorf("stub.GetListForForfeiture err:list is nil.")
	}
	//判断是否在列表中
	if _, ok := listForForfeiture[forfeitureAddr]; !ok {
		log.Error("Node is not exist in the forfeiture list.")
		return fmt.Errorf("node is not exist in the forfeiture list.")
	}
	//获取节点信息
	forfeitureNode := listForForfeiture[forfeitureAddr]
	//在列表中移除，并获取没收情况
	delete(listForForfeiture, forfeitureAddr)
	//更新列表
	err = SaveListForForfeiture(stub, listForForfeiture)
	if err != nil {
		log.Error("Stub.PutState err:", "error", err)
		return err
	}
	//判断余额
	if forfeitureNode.ApplyTokens.Amount > balance.Balance {
		log.Error("The balance is not enough.")
		return fmt.Errorf("The balance is not enough.")
	}
	//判断节点类型
	switch {
	case forfeitureNode.ForfeitureRole == Mediator:
		return handleMediatorForfeitureDeposit(foundationAddr, forfeitureAddr, forfeitureNode, balance, stub)
	case forfeitureNode.ForfeitureRole == Jury:
		return handleJuryForfeitureDeposit(foundationAddr, forfeitureAddr, forfeitureNode, balance, stub)
	case forfeitureNode.ForfeitureRole == Developer:
		return handleDeveloperForfeitureDeposit(foundationAddr, forfeitureAddr, forfeitureNode, balance, stub)
	default:
		return fmt.Errorf("please enter validate role.")
	}
}

//处理没收Mediator保证金
func handleMediatorForfeitureDeposit(foundationAddr string, forfeitureAddress string, forfeiture *Forfeiture, balance *DepositBalance, stub shim.ChaincodeStubInterface) error {
	var err error
	//计算余额
	result := balance.Balance - forfeiture.ApplyTokens.Amount
	//判断是否没收全部，即在列表中移除该节点
	if result == 0 {
		//没收不考虑是否在规定周期内,其实它肯定是在列表中并已在周期内
		//没收全部，即删除,已经是计算好奖励了
		err = forfeitureAllDeposit(modules.MediatorList, stub, foundationAddr, forfeitureAddress, forfeiture.ApplyTokens)
		if err != nil {
			log.Error("ForfeitureAllDeposit err:", "error", err)
			return err
		}
		return nil
	} else {
		//TODO 对于mediator，要么全没收，要么退出一部分，且退出该部分金额后还在列表中
		return forfeitureSomeDeposit(stub, foundationAddr, forfeitureAddress, forfeiture, balance)
	}
}

func handleJuryForfeitureDeposit(foundationAddr string, forfeitureAddr string, forfeiture *Forfeiture, balance *DepositBalance, stub shim.ChaincodeStubInterface) error {
	depositAmountsForJuryStr, err := stub.GetSystemConfig(DepositAmountForJury)
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
	//计算余额
	result := balance.Balance - forfeiture.ApplyTokens.Amount
	//判断是否没收全部，即在列表中移除该节点
	if result == 0 {
		//没收不考虑是否在规定周期内,其实它肯定是在列表中并已在周期内
		//没收全部，即删除
		err = forfeitureAllDeposit(modules.JuryList, stub, foundationAddr, forfeitureAddr, forfeiture.ApplyTokens)
		if err != nil {
			log.Error("ForfeitureAllDeposit err:", "error", err)
			return err
		}
		return nil
	} else if result < depositAmountsForJury {
		//TODO 对于jury，需要移除列表
		err = forfertureAndMoveList(modules.JuryList, stub, foundationAddr, forfeitureAddr, forfeiture, balance)
		if err != nil {
			return err
		}
	} else {
		//TODO 退出一部分，且退出该部分金额后还在列表中
		err = forfeitureSomeDeposit(stub, foundationAddr, forfeitureAddr, forfeiture, balance)
		if err != nil {
			return err
		}
	}
	return nil
}

func handleDeveloperForfeitureDeposit(foundationAddr string, forfeitureAddr string, forfeiture *Forfeiture, balance *DepositBalance, stub shim.ChaincodeStubInterface) error {
	depositAmountsForDeveloperStr, err := stub.GetSystemConfig(DepositAmountForDeveloper)
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
	//计算余额
	result := balance.Balance - forfeiture.ApplyTokens.Amount
	//判断是否没收全部，即在列表中移除该节点
	if result == 0 {
		//没收不考虑是否在规定周期内,其实它肯定是在列表中并已在周期内
		//没收全部，即删除
		err = forfeitureAllDeposit(DeveloperList, stub, foundationAddr, forfeitureAddr, forfeiture.ApplyTokens)
		if err != nil {
			log.Error("ForfeitureAllDeposit err:", "error", err)
			return err
		}
		return nil
	} else if result < depositAmountsForDeveloper {
		return forfertureAndMoveList(DeveloperList, stub, foundationAddr, forfeitureAddr, forfeiture, balance)
	} else {
		//TODO 退出一部分，且退出该部分金额后还在列表中
		return forfeitureSomeDeposit(stub, foundationAddr, foundationAddr, forfeiture, balance)
	}
}

//不同意这样没收请求，则直接从没收列表中移除该节点
func disagreeForApplyForfeiture(stub shim.ChaincodeStubInterface, forfeiture string) error {
	//获取没收列表
	listForForfeiture, err := GetListForForfeiture(stub)
	if err != nil {
		log.Error("Stub.GetListForForfeiture err:", "error", err)
		return err
	}
	if listForForfeiture == nil {
		log.Error("Stub.GetListForForfeiture:list is nil.")
		return fmt.Errorf("Stub.GetListForForfeiture:list is nil.")
	}
	//判断是否在列表中
	//f, _ := common.StringToAddress(forfeiture)
	if _, ok := listForForfeiture[forfeiture]; !ok {
		log.Error("Node is not exist in the forfeiture list.")
		return fmt.Errorf("Node is not exist in the forfeiture list.")
	}
	//从列表中移除
	delete(listForForfeiture, forfeiture)
	listForForfeitureByte, err := json.Marshal(listForForfeiture)
	if err != nil {
		log.Error("Json.Marshal err:", "error", err)
		return err
	}
	//更新列表
	err = stub.PutState(ListForForfeiture, listForForfeitureByte)
	if err != nil {
		log.Error("Stub.PutState err:", "error", err)
		return err
	}
	return nil
}

func handleForForfeitureApplication(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	log.Info("handleForForfeitureApplication")
	//  地址，是否同意
	if len(args) != 2 {
		log.Error("args need two parameters.")
		return shim.Error("args need two parameters.")
	}
	//  基金会地址
	invokeAddr, err := stub.GetInvokeAddress()
	if err != nil {
		log.Error("Stub.GetInvokeAddress err", "error", err)
		return shim.Error(err.Error())
	}
	//
	foundationAddress, err := stub.GetSystemConfig(modules.FoundationAddress)
	if err != nil {
		log.Error("Stub.GetSystemConfig with FoundationAddress err:", "error", err)
		return shim.Error(err.Error())
	}
	//  判断没收请求地址是否是基金会地址
	if strings.Compare(invokeAddr.String(), foundationAddress) != 0 {
		log.Error("please use foundation address")
		return shim.Error("please use foundation address")
	}
	//获取传入参数信息
	addr := args[0]
	isOk := args[1]
	//  获取一下该用户下的账簿情况
	balance, err := GetNodeBalance(stub, addr)
	if err != nil {
		log.Error("Stub.GetDepositBalance err:", "error", err)
		return shim.Error(err.Error())
	}
	//  判断没收节点账户是否为空
	if balance == nil {
		log.Error("Stub.GetDepositBalance: balance is nil.")
		return shim.Error("Stub.GetDepositBalance: balance is nil.")
	}
	//check 如果为ok，则同意此申请，如果为no，则不同意此申请
	if strings.Compare(isOk, Ok) == 0 {
		err = agreeForApplyForfeiture(stub, invokeAddr.String(), addr, balance)
		if err != nil {
			return shim.Error(err.Error())
		}
	} else if strings.Compare(isOk, No) == 0 {
		//移除申请列表，不做处理
		err = disagreeForApplyForfeiture(stub, addr)
		if err != nil {
			return shim.Error(err.Error())
		}
	} else {
		log.Error("Please enter ok or no.")
		return shim.Error("Please enter ok or no.")
	}
	return shim.Success([]byte(nil))
}

//基金会处理
func handleForDeveloperApplyCashback(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	log.Info("handleForDeveloperApplyCashback")
	//  地址，申请时间，是否同意
	if len(args) != 2 {
		log.Error("args need two parameters")
		return shim.Error("args need two parameters")
	}
	//  基金会地址
	invokeAddr, err := stub.GetInvokeAddress()
	if err != nil {
		log.Error("Stub.GetInvokeAddress err: ", "error", err)
		return shim.Error(err.Error())
	}
	//  判断没收请求地址是否是基金会地址
	foundationAddress, err := stub.GetSystemConfig(modules.FoundationAddress)
	if err != nil {
		log.Error("stub.GetSystemConfig with FoundationAddress err: ", "error", err)
		return shim.Error(err.Error())
	}
	if strings.Compare(invokeAddr.String(), foundationAddress) != 0 {
		log.Error("please use foundation address")
		return shim.Error("please use foundation address")
	}
	//  获取一下该用户下的账簿情况
	addr, err := common.StringToAddress(args[0])
	if err != nil {
		log.Error("common.StringToAddress err: ", "error", err)
		return shim.Error(err.Error())
	}
	balance, err := GetNodeBalance(stub, addr.String())
	if err != nil {
		log.Error("Stub.GetDepositBalance err: ", "error", err)
		return shim.Error(err.Error())
	}
	//  判断没收节点账户是否为空
	if balance == nil {
		log.Error("balance is nil")
		return shim.Error("balance is nil")
	}
	isOk := args[1]
	if strings.Compare(isOk, Ok) == 0 {
		//  对余额处理
		err = handleDeveloper(stub, addr, balance)
		if err != nil {
			log.Error("handleDeveloper err ", "error", err)
			return shim.Error(err.Error())
		}
	} else if strings.Compare(isOk, No) == 0 {
		//移除提取申请列表
		err = moveAndPutStateFromCashbackList(stub, addr)
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
func handleDeveloperDepositCashback(stub shim.ChaincodeStubInterface, cashbackAddr common.Address, cashbackValue *Cashback, balance *DepositBalance) error {
	//
	depositAmountsForDeveloperStr, err := stub.GetSystemConfig(DepositAmountForDeveloper)
	if err != nil {
		log.Error("Stub.GetSystemConfig with DepositAmountForDeveloper err:", "error", err)
		return err
	}
	//  转换
	depositAmountsForDeveloper, err := strconv.ParseUint(depositAmountsForDeveloperStr, 10, 64)
	if err != nil {
		log.Error("Strconv.ParseUint err:", "error", err)
		return err
	}
	if balance.Balance >= depositAmountsForDeveloper {
		//  已在列表中
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
	log.Info("handleForJuryApplyCashback")
	//  地址，申请时间，是否同意
	if len(args) != 2 {
		log.Error("args need two parameters")
		return shim.Error("args need two parameters")
	}
	//  基金会地址
	invokeAddr, err := stub.GetInvokeAddress()
	if err != nil {
		log.Error("Stub.GetInvokeAddress err: ", "error", err)
		return shim.Error(err.Error())
	}
	//  判断没收请求地址是否是基金会地址
	foundationAddress, err := stub.GetSystemConfig(modules.FoundationAddress)
	if err != nil {
		log.Error("Stub.GetSystemConfig with FoundationAddress err: ", "error", err)
		return shim.Error(err.Error())
	}
	if strings.Compare(invokeAddr.String(), foundationAddress) != 0 {
		log.Error("please use foundation address")
		return shim.Error("please use foundation address")
	}
	//  获取一下该用户下的账簿情况
	addr, _ := common.StringToAddress(args[0])
	balance, err := GetNodeBalance(stub, addr.String())
	if err != nil {
		log.Error("stub.GetDepositBalance err: ", "error", err)
		return shim.Error(err.Error())
	}
	//  判断没收节点账户是否为空
	if balance == nil {
		log.Error("balance is nil.")
		return shim.Error("balance is nil.")
	}
	isOk := args[1]
	if strings.Compare(isOk, Ok) == 0 {
		//  对余额处理
		err = handleJury(stub, addr, balance)
		if err != nil {
			log.Error("handleJury err", "error", err)
			return shim.Error(err.Error())
		}
	} else if strings.Compare(isOk, No) == 0 {
		//  移除提取申请列表
		err = moveAndPutStateFromCashbackList(stub, addr)
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
func handleJuryDepositCashback(stub shim.ChaincodeStubInterface, cashbackAddr common.Address, cashbackValue *Cashback, balance *DepositBalance) error {
	//
	depositAmountsForJuryStr, err := stub.GetSystemConfig(DepositAmountForJury)
	if err != nil {
		log.Error("Stub.GetSystemConfig with DepositAmountForJury err:", "error", err)
		return err
	}
	//  转换
	depositAmountsForJury, err := strconv.ParseUint(depositAmountsForJuryStr, 10, 64)
	if err != nil {
		log.Error("Strconv.ParseUint err:", "error", err)
		return err
	}
	//
	if balance.Balance >= depositAmountsForJury {
		//  已在列表中
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
	log.Info("handleForApplyBecomeMediator")
	//  获取
	invokeAddr, err := stub.GetInvokeAddress()
	if err != nil {
		log.Error("get invoke address err: ", "error", err)
		return shim.Error(err.Error())
	}
	//  获取
	foundationAddress, err := stub.GetSystemConfig(modules.FoundationAddress)
	if err != nil {
		log.Error("get foundation address err: ", "error", err)
		return shim.Error(err.Error())
	}
	// 判断当前请求的是否为基金会
	if strings.Compare(invokeAddr.String(), foundationAddress) != 0 {
		log.Error("please use foundation address")
		return shim.Error("please use foundation address")
	}
	if len(args) != 2 {
		log.Error("args need two parameters")
		return shim.Error("args need two parameters")
	}
	//  获取申请列表
	becomeList, err := GetList(stub, ListForApplyBecomeMediator)
	if err != nil {
		log.Error("get become list err: ", "error", err)
		return shim.Error(err.Error())
	}
	if becomeList == nil {
		log.Error("get become list: list is nil")
		return shim.Error("get become list: list is nil")
	}
	isOk := args[0]
	addr, err := common.StringToAddress(args[1])
	if err != nil {
		log.Error("string to address err: ", "error", err)
		return shim.Error(err.Error())
	}
	if _, ok := becomeList[addr.String()]; !ok {
		log.Error("node is not exist in the become list")
		return shim.Error("node is not exist in the become list")
	}
	//  不同意，移除申请列表
	if isOk == No {
		delete(becomeList, addr.String())
		//  保存成为列表
		err = saveList(stub, ListForApplyBecomeMediator, becomeList)
		if err != nil {
			log.Error("save become list err: ", "error", err)
			return shim.Error(err.Error())
		}

		err = DelMediatorDeposit(stub, addr.String())
		if err != nil {
			return shim.Error(err.Error())
		}
	} else if isOk == Ok {
		//  同意，移除列表，并且加入同意申请列表
		delete(becomeList, addr.String())
		//  保存成为列表
		err = saveList(stub, ListForApplyBecomeMediator, becomeList)
		if err != nil {
			log.Error("save become list err: ", "error", err)
			return shim.Error(err.Error())
		}
		//  获取同意列表
		agreeList, err := GetList(stub, ListForAgreeBecomeMediator)
		if err != nil {
			log.Error("get agree list err: ", "error", err)
			return shim.Error(err.Error())
		}
		if agreeList == nil {
			agreeList = make(map[string]bool)
		} else {
			if _, ok := agreeList[addr.String()]; ok {
				log.Error("node was in the agree list")
				return shim.Error("node was in the agree list")
			}
		}
		agreeList[addr.String()] = true
		//  保存同意列表
		err = saveList(stub, ListForAgreeBecomeMediator, agreeList)
		if err != nil {
			log.Error("save agree list err: ", "error", err)
			return shim.Error(err.Error())
		}
		// 修改同意时间
		mediator, err := GetMediatorDeposit(stub, addr.Str())
		if err != nil {
			log.Error("get mediator info err: ", "error", err)
			return shim.Error(err.Error())
		}
		mediator.AgreeTime = time.Now().Unix() / DTimeDuration
		mediator.Status = Agree
		err = SaveMediatorDeposit(stub, addr.Str(), mediator)
		if err != nil {
			log.Error("save mediator info err: ", "error", err)
			return shim.Error(err.Error())
		}
	} else {
		log.Error("please enter ok")
		return shim.Error("please enter ok")
	}
	return shim.Success([]byte(nil))
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
	foundationAddress, err := stub.GetSystemConfig(modules.FoundationAddress)
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
	quitList, err := GetList(stub, ListForApplyQuitMediator)
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
	addr1 := args[1]
	addr, err := common.StringToAddress(addr1)
	if err != nil {
		log.Error("common.StringToAddress err:", "error", err)
		return shim.Error(err.Error())
	}
	if _, ok := quitList[addr.String()]; !ok {
		log.Error("Node is not exist in the quit list.")
		return shim.Error("Node is not exist in the quit list.")
	}
	if strings.Compare(isOk, No) == 0 {
		delete(quitList, addr.String())
		err = saveList(stub, ListForApplyQuitMediator, quitList)
		if err != nil {
			log.Error("MarshalAndPutStateForMediatorList err:", "error", err)
			return shim.Error(err.Error())
		}
	} else if strings.Compare(isOk, Ok) == 0 {
		log.Info("foundation is agree with application.")
		//同意，移除列表，并且全款退出
		delete(quitList, addr.String())
		err = saveList(stub, ListForApplyQuitMediator, quitList)
		if err != nil {
			log.Error("MarshalAndPutStateForMediatorList err:", "error", err)
			return shim.Error(err.Error())
		}
		//获取该账户
		md, err := GetMediatorDeposit(stub, addr.String())
		if err != nil {
			log.Error("Stub.GetDepositBalance err:", "error", err)
			return shim.Error(err.Error())
		}
		if md == nil {
			log.Error("Stub.GetDepositBalance err: balance is nil.")
			return shim.Error("Stub.GetDepositBalance err: balance is nil.")
		}
		err = deleteMediatorDeposit(stub, md, addr)
		if err != nil {
			log.Error("DeleteNode err:", "error", err)
			return shim.Error(err.Error())
		}
		//从同意列表中删除
		//agreeList, err := GetList(stub, ListForAgreeBecomeMediator)
		//if err != nil {
		//	log.Error("Stub.GetAgreeForBecomeMediatorList err:", "error", err)
		//	return shim.Error(err.Error())
		//}
		//if agreeList == nil {
		//	log.Error("Stub.GetAgreeForBecomeMediatorList err: list is nil")
		//	shim.Error("Stub.GetAgreeForBecomeMediatorList err: list is nil")
		//}
		//if _, ok := agreeList[addr.String()]; !ok {
		//	log.Error("Node is not exist in the agree list.")
		//	shim.Error("Node is not exist in the agree list.")
		//}
		////  移除
		//delete(agreeList, addr.String())
		//err = saveList(stub, ListForAgreeBecomeMediator, agreeList)
		//if err != nil {
		//	log.Error("MarshalAndPutStateForMediatorList err:", "error", err)
		//	return shim.Error(err.Error())
		//}
	} else {
		log.Error("please enter ok or no")
		return shim.Error("please enter ok or no")
	}
	return shim.Success([]byte(nil))
}

//基金会处理
func handleForMediatorApplyCashback(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	log.Info("handleForMediatorApplyCashback")
	//  地址，是否同意
	if len(args) != 2 {
		log.Error("args need two parameters")
		return shim.Error("args need two parameters")
	}
	//  基金会地址
	invokeAddr, err := stub.GetInvokeAddress()
	if err != nil {
		log.Error("get invoke address err: ", "error", err)
		return shim.Error(err.Error())
	}
	//  判断没收请求地址是否是基金会地址
	foundationAddress, err := stub.GetSystemConfig(modules.FoundationAddress)
	if err != nil {
		log.Error("get foundation address err: ", "error", err)
		return shim.Error(err.Error())
	}
	if invokeAddr.String() != foundationAddress {
		log.Error("please use foundation address")
		return shim.Error("please use foundation address")
	}
	//  获取一下该用户下的账簿情况
	addr, err := common.StringToAddress(args[0])
	if err != nil {
		log.Error("string to address err: ", "error", err)
		return shim.Error(err.Error())
	}
	md, err := GetMediatorDeposit(stub, addr.String())
	if err != nil {
		log.Error("get cashback node balance err: ", "error", err)
		return shim.Error(err.Error())
	}
	//  判断没收节点账户是否为空
	if md == nil {
		log.Error("get cashback node balance: balance is nil")
		return shim.Error("get cashback node balance: balance is nil")
	}
	isOk := args[1]
	//  判断处理结果
	if isOk == Ok {
		//  对余额处理
		err = handleMediator(stub, addr, md)
		if err != nil {
			log.Error("handle mediator err: ", "error", err)
			return shim.Error(err.Error())
		}
	} else if isOk == No {
		//移除提取申请列表
		err = moveAndPutStateFromCashbackList(stub, addr)
		if err != nil {
			log.Error("moveAndPutStateFromCashbackList err:", "error", err)
			return shim.Error(err.Error())
		}
	} else {
		log.Error("please enter ok or no")
		return shim.Error("please enter ok or no")
	}
	return shim.Success([]byte(nil))
}
