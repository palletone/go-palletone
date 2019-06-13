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
	"fmt"
	"strconv"
	"time"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/award"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/contracts/shim"
	"github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
	"github.com/palletone/go-palletone/dag/modules"
)

func juryPayToDepositContract(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	log.Info("juryPayToDepositContract")
	//  获取jury交付保证金的下线
	depositAmountsForJuryStr, err := stub.GetSystemConfig(DepositAmountForJury)
	if err != nil {
		log.Error("get deposit amount for jury err: ", "error", err)
		return shim.Error(err.Error())
	}
	//  转换
	depositAmountsForJury, err := strconv.ParseUint(depositAmountsForJuryStr, 10, 64)
	if err != nil {
		log.Error("strconv.ParseUint err: ", "error", err)
		return shim.Error(err.Error())
	}
	//  交付地址
	invokeAddr, err := stub.GetInvokeAddress()
	if err != nil {
		log.Error("get invoke address err: ", "error", err)
		return shim.Error(err.Error())
	}
	//  判断是否交付保证金交易
	invokeTokens, err := isContainDepositContractAddr(stub)
	if err != nil {
		log.Error("isContainDepositContractAddr err: ", "error", err)
		return shim.Error(err.Error())
	}
	//获取账户
	balance, err := GetNodeBalance(stub, invokeAddr.String())
	if err != nil {
		log.Error("get node balance err: ", "error", err)
		return shim.Error(err.Error())
	}
	isJury := false
	if balance == nil {
		balance = &DepositBalance{}
		if invokeTokens.Amount >= depositAmountsForJury {
			//  加入候选列表
			err = addCandaditeList(invokeAddr, stub, modules.JuryList)
			if err != nil {
				log.Error("addCandaditeList err: ", "error", err)
				return shim.Error(err.Error())
			}
			isJury = true
			balance.EnterTime = TimeStr()
		}
		balance.Balance += invokeTokens.Amount
		balance.LastModifyTime = TimeStr()
	} else {
		//  账户已存在，进行信息的更新操作
		if balance.Balance >= depositAmountsForJury {
			//  原来就是jury
			isJury = true
			//  TODO 再次交付保证金时，先计算当前余额的币龄奖励
			//endTime := balance.LastModifyTime * DTimeDuration
			//endTime, _ := time.Parse(Layout, balance.LastModifyTime)
			endTime := StrToTime(balance.LastModifyTime)
			//  获取保证金年利率
			depositRate, err := stub.GetSystemConfig(modules.DepositRate)
			if err != nil {
				log.Error("get deposit rate err: ", "error", err)
				return shim.Error(err.Error())
			}
			//  计算币龄收益
			awards := award.GetAwardsWithCoins(balance.Balance, endTime.Unix(), depositRate)
			balance.Balance += awards
		}
		//  处理交付保证金数据
		balance.Balance += invokeTokens.Amount
		balance.LastModifyTime = TimeStr()
	}
	if !isJury {
		//  判断交了保证金后是否超过了jury
		if balance.Balance >= depositAmountsForJury {
			//  加入候选列表
			err = addCandaditeList(invokeAddr, stub, modules.JuryList)
			if err != nil {
				log.Error("addCandaditeList err: ", "error", err)
				return shim.Error(err.Error())
			}
			balance.EnterTime = TimeStr()
		}
	}
	err = SaveNodeBalance(stub, invokeAddr.String(), balance)
	if err != nil {
		log.Error("save node balance err: ", "error", err)
		return shim.Error(err.Error())
	}
	return shim.Success([]byte(nil))
}

func juryApplyCashback(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	err := applyCashbackList(Jury, stub, args)
	if err != nil {
		log.Error("applyCashbackList err: ", "error", err)
		return shim.Error(err.Error())
	}
	return shim.Success([]byte(nil))
}

func handleJury(stub shim.ChaincodeStubInterface, cashbackAddr common.Address, balance *DepositBalance) error {
	//  获取请求列表
	listForCashback, err := GetListForCashback(stub)
	if err != nil {
		log.Error("Stub.GetListForCashback err:", "error", err)
		return err
	}
	if listForCashback == nil {
		log.Error("listForCashback is nil.")
		return fmt.Errorf("%s", "listForCashback is nil.")
	}
	if _, ok := listForCashback[cashbackAddr.String()]; !ok {
		log.Error("node is not exist in the list.")
		return fmt.Errorf("%s", "node is not exist in the list.")
	}
	cashbackNode := listForCashback[cashbackAddr.String()]
	delete(listForCashback, cashbackAddr.String())
	//更新列表
	err = SaveListForCashback(stub, listForCashback)
	if err != nil {
		log.Error("saveListForCashback err:", "error", err)
		return err
	}
	//  还得判断一下是否超过余额
	if cashbackNode.CashbackTokens.Amount > balance.Balance {
		log.Error("Balance is not enough.")
		return fmt.Errorf("%s", "Balance is not enough.")
	}
	//
	err = handleJuryDepositCashback(stub, cashbackAddr, cashbackNode, balance)
	if err != nil {
		log.Error("HandleJuryDepositCashback err:", "error", err)
		return err
	}
	return nil
}

//Jury已在列表中
func handleJuryFromList(stub shim.ChaincodeStubInterface, cashbackAddr common.Address, cashbackValue *Cashback, balance *DepositBalance) error {
	//
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
	//  退出列表
	//  计算余额
	result := balance.Balance - cashbackValue.CashbackTokens.Amount
	//  判断是否退出列表
	if result == 0 {
		//  加入列表时的时间
		//enterTime, _ := time.Parse(Layout, balance.EnterTime)
		enterTime := StrToTime(balance.EnterTime)
		startTime := time.Unix(enterTime.Unix(), 0).UTC().YearDay()
		//  当前退出时间
		endTime := time.Now().UTC().YearDay()
		//  判断是否已到期
		if endTime-startTime >= day {
			//  退出全部，即删除cashback，利息计算好了
			err = cashbackAllDeposit(modules.JuryList, stub, cashbackAddr, cashbackValue.CashbackTokens, balance)
			if err != nil {
				return err
			}
		} else {
			log.Error("not exceeding the valid time,can not cashback some")
			return fmt.Errorf("%s", "not exceeding the valid time,can not cashback some")
		}
	} else {
		//TODO 退出一部分，且退出该部分金额后还在列表中，还没有计算利息
		err = cashbackSomeDeposit(modules.JuryList, stub, cashbackAddr, cashbackValue, balance)
		if err != nil {
			log.Error("cashbackSomeDeposit err: ", "error", err)
			return err
		}
	}
	return nil
}
