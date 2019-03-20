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
	"github.com/palletone/go-palletone/common/award"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/contracts/shim"
	"github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
	"strconv"
	"strings"
	"time"
)

func developerPayToDepositContract(stub shim.ChaincodeStubInterface, args []string) peer.Response {
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
	//交付地址
	invokeAddr, err := stub.GetInvokeAddress()
	if err != nil {
		log.Error("Stub.GetInvokeAddress err:", "error", err)
		return shim.Error(err.Error())
	}
	//交付数量
	//交付数量
	//invokeTokens, err := stub.GetInvokeTokens()
	invokeTokens, err := isContainDepositContractAddr(stub)
	if err != nil {
		log.Error("Stub.GetInvokeTokens err:", "error", err)
		return shim.Error(err.Error())
	}
	//获取账户
	balance, err := GetDepositBalance(stub, invokeAddr)
	if err != nil {
		log.Error("Stub.GetDepositBalance err:", "error", err)
		return shim.Error(err.Error())
	}
	isDeveloper := false
	if balance == nil {
		balance = &DepositBalance{}
		if invokeTokens.Amount >= depositAmountsForDeveloper {
			//加入列表
			//addList("Developer", invokeAddr, stub)
			err = addCandaditeList(invokeAddr, stub, "DeveloperList")
			if err != nil {
				log.Error("AddCandaditeList err:", "error", err)
				return shim.Error(err.Error())
			}
			isDeveloper = true
			balance.EnterTime = time.Now().UTC().Unix() / 1800
		}
		updateForPayValue(balance, invokeTokens)
	} else {
		//账户已存在，进行信息的更新操作
		if balance.TotalAmount >= depositAmountsForDeveloper {
			//原来就是Developer
			isDeveloper = true
			//TODO 再次交付保证金时，先计算当前余额的币龄奖励
			endTime := balance.LastModifyTime * 1800
			awards := award.GetAwardsWithCoins(balance.TotalAmount, endTime)
			balance.TotalAmount += awards

		}
		//处理交付保证金数据
		updateForPayValue(balance, invokeTokens)
	}
	if !isDeveloper {
		//判断交了保证金后是否超过了Developer
		if balance.TotalAmount >= depositAmountsForDeveloper {
			//addList("Developer", invokeAddr, stub)
			err = addCandaditeList(invokeAddr, stub, "DeveloperList")
			if err != nil {
				log.Error("AddCandaditeList err:", "error", err)
				return shim.Error(err.Error())
			}
			balance.EnterTime = time.Now().UTC().Unix() / 1800
		}
	}
	err = marshalAndPutStateForBalance(stub, invokeAddr, balance)
	if err != nil {
		log.Error("MarshalAndPutStateForBalance err:", "error", err)
		return shim.Error(err.Error())
	}
	return shim.Success([]byte("ok"))
}

func developerApplyCashback(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	err := applyCashbackList("Developer", stub, args)
	if err != nil {
		log.Error("ApplyCashbackList err:", "error", err)
		return shim.Error(err.Error())
	}
	return shim.Success([]byte("ok"))
}

//基金会处理
func handleForDeveloperApplyCashback(stub shim.ChaincodeStubInterface, args []string) peer.Response {
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
	if strings.Compare(invokeAddr, foundationAddress) != 0 {
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

func handleDeveloper(stub shim.ChaincodeStubInterface, cashbackAddr string, applyTime int64, balance *DepositBalance) error {
	//获取请求列表
	listForCashback, err := GetListForCashback(stub)
	if err != nil {
		log.Error("Stub.GetListForCashback err:", "error", err)
		return err
	}
	if listForCashback == nil {
		log.Error("listForCashback is nil.")
		return fmt.Errorf("%s", "listForCashback is nil.")
	}
	isExist := isInCashbacklist(cashbackAddr, listForCashback)
	if !isExist {
		log.Error("node is not exist in the list.")
		return fmt.Errorf("%s", "node is not exist in the list.")
	}
	//获取节点信息
	cashbackNode := &Cashback{}
	isFound := false
	for _, m := range listForCashback {
		if m.CashbackAddress == cashbackAddr && m.CashbackTime == applyTime {
			cashbackNode = m
			isFound = true
			break
		}
	}
	if !isFound {
		log.Error("Apply time is wrong.")
		return fmt.Errorf("%s", "Apply time is wrong.")
	}
	newList, _ := moveInApplyForCashbackList(stub, listForCashback, cashbackAddr, applyTime)
	listForCashbackByte, err := json.Marshal(newList)
	if err != nil {
		log.Error("Json.Marshal err:", "error", err)
		return err
	}
	//更新列表
	err = stub.PutState("ListForCashback", listForCashbackByte)
	if err != nil {
		log.Error("Stub.PutState err:", "error", err)
		return err
	}
	//还得判断一下是否超过余额
	if cashbackNode.CashbackTokens.Amount > balance.TotalAmount {
		log.Error("Balance is not enough.")
		return fmt.Errorf("%s", "Balance is not enough.")
	}
	err = handleDeveloperDepositCashback(stub, cashbackAddr, cashbackNode, balance)
	if err != nil {
		log.Error("HandleDeveloperDepositCashback err:", "error", err)
		return err
	}
	return nil
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

//Developer已在列表中
func handleDeveloperFromList(stub shim.ChaincodeStubInterface, cashbackAddr string, cashbackValue *Cashback, balance *DepositBalance) error {
	depositPeriod, err := stub.GetSystemConfig("DepositPeriod")
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
	//退出列表
	//计算余额
	result := balance.TotalAmount - cashbackValue.CashbackTokens.Amount
	//判断是否退出列表
	if result == 0 {
		//加入列表时的时间
		startTime := time.Unix(balance.EnterTime*1800, 0).UTC().YearDay()
		//当前退出时间
		endTime := time.Now().UTC().YearDay()
		//判断是否已到期
		if endTime-startTime >= day {
			//退出全部，即删除cashback，利息计算好了
			err = cashbackAllDeposit("Developer", stub, cashbackAddr, cashbackValue.CashbackTokens, balance)
			if err != nil {
				return err
			}
		} else {
			log.Error("Not exceeding the valid time,can not cashback some.")
			return fmt.Errorf("%s", "Not exceeding the valid time,can not cashback some.")
		}
	} else {
		//TODO 退出一部分，且退出该部分金额后还在列表中，还没有计算利息
		//d.addListForCashback("Developer", stub, cashbackAddr, invokeTokens)
		err = cashbackSomeDeposit("Developer", stub, cashbackAddr, cashbackValue, balance)
		if err != nil {
			log.Error("CashbackSomeDeposit err:", "error", err)
			return err
		}
	}
	return nil
}
