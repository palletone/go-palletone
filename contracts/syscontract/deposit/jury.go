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
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/contracts/shim"
	"github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
	"github.com/palletone/go-palletone/dag/constants"
	"github.com/palletone/go-palletone/dag/modules"
)

func juryPayToDepositContract(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	log.Info("juryPayToDepositContract")
	//  判断是否交付保证金交易
	invokeTokens, err := isContainDepositContractAddr(stub)
	if err != nil {
		log.Error("isContainDepositContractAddr err: ", "error", err)
		return shim.Error(err.Error())
	}
	//  获取jury交付保证金的下线
	//depositAmountsForJuryStr, err := stub.GetSystemConfig(DepositAmountForJury)
	//if err != nil {
	//	log.Error("get deposit amount for jury err: ", "error", err)
	//	return shim.Error(err.Error())
	//}
	////  转换
	//depositAmountsForJury, err := strconv.ParseUint(depositAmountsForJuryStr, 10, 64)
	//if err != nil {
	//	log.Error("strconv.ParseUint err: ", "error", err)
	//	return shim.Error(err.Error())
	//}
	cp, err := stub.GetSystemConfig()
	if err != nil {
		//log.Error("strconv.ParseUint err:", "error", err)
		return shim.Error(err.Error())
	}
	depositAmountsForJury := cp.DepositAmountForJury
	//  交付地址
	invokeAddr, err := stub.GetInvokeAddress()
	if err != nil {
		log.Error("get invoke address err: ", "error", err)
		return shim.Error(err.Error())
	}
	//获取账户
	balance, err := GetNodeBalance(stub, invokeAddr.String())
	if err != nil {
		log.Error("get node balance err: ", "error", err)
		return shim.Error(err.Error())
	}
	//  第一次想加入
	if balance == nil {
		balance = &DepositBalance{}
		//  可以加入列表
		if invokeTokens.Amount >= depositAmountsForJury {
			//  加入候选列表
			err = addCandaditeList(stub, invokeAddr, modules.JuryList)
			if err != nil {
				log.Error("addCandaditeList err: ", "error", err)
				return shim.Error(err.Error())
			}
			balance.EnterTime = TimeStr()
		}
		//  没有
		balance.Balance += invokeTokens.Amount
	} else {
		//  TODO 再次交付保证金时，先计算当前余额的币龄奖励
		//  如果在候选列表当中，即可享受利息
		if balance.EnterTime != "" {
			awards := caculateAwards(stub, balance.Balance, balance.LastModifyTime)
			balance.Balance += awards
		}
		//  处理交付保证金数据
		balance.Balance += invokeTokens.Amount
	}
	//  判断再次交付后是否可以加入列表
	if balance.EnterTime == "" {
		//  判断此时交了保证金后是否超过了jury
		if balance.Balance >= depositAmountsForJury {
			//  加入候选列表
			err = addCandaditeList(stub, invokeAddr, modules.JuryList)
			if err != nil {
				log.Error("addCandaditeList err: ", "error", err)
				return shim.Error(err.Error())
			}
			balance.EnterTime = TimeStr()
		}
	}
	balance.LastModifyTime = TimeStr()
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

//Jury已在列表中,并发起退钱申请，需要判断是否需要删除该节点，移除列表等
func handleJuryFromList(stub shim.ChaincodeStubInterface, cashbackAddr common.Address, cashbackValue *Cashback, balance *DepositBalance) error {
	//depositAmountsForJuryStr, err := stub.GetSystemConfig(DepositAmountForJury)
	//if err != nil {
	//	log.Error("Stub.GetSystemConfig with DepositAmountForJury err:", "error", err)
	//	return err
	//}
	////  转换
	//depositAmountsForJury, err := strconv.ParseUint(depositAmountsForJuryStr, 10, 64)
	//if err != nil {
	//	log.Error("Strconv.ParseUint err:", "error", err)
	//	return err
	//}
	cp, err := stub.GetSystemConfig()
	if err != nil {
		//log.Error("strconv.ParseUint err:", "error", err)
		return err
	}
	depositAmountsForJury := cp.DepositAmountForJury
	//  这里计算这一次操作的币龄利息
	awards := caculateAwards(stub, balance.Balance, balance.LastModifyTime)
	//  剩下的余额
	result := balance.Balance - cashbackValue.CashbackTokens.Amount
	// 需要删除节点和移除列表
	if result == 0 {
		//
		cashbackValue.CashbackTokens.Amount += awards
		//  调用从合约把token转到请求地址
		err := stub.PayOutToken(cashbackAddr.String(), cashbackValue.CashbackTokens, 0)
		if err != nil {
			log.Error("stub.PayOutToken err:", "error", err)
			return err
		}
		//  移除出列表
		err = moveCandidate(modules.JuryList, cashbackAddr.String(), stub)
		if err != nil {
			log.Error("moveCandidate err:", "error", err)
			return err
		}
		//  删除节点
		err = stub.DelState(string(constants.DEPOSIT_BALANCE_PREFIX) + cashbackAddr.String())
		if err != nil {
			log.Error("stub.DelState err:", "error", err)
			return err
		}
		//  特殊处理
		return nil
	} else if result < depositAmountsForJury {
		//  移除列表并更新
		err = moveCandidate(modules.JuryList, cashbackAddr.String(), stub)
		if err != nil {
			log.Error("moveCandidate err:", "error", err)
			return err
		}
		balance.EnterTime = ""
		balance.LastModifyTime = TimeStr()
	} else {
		//  只更新账户
		balance.LastModifyTime = TimeStr()

	}
	//  调用从合约把token转到请求地址
	err = stub.PayOutToken(cashbackAddr.String(), cashbackValue.CashbackTokens, 0)
	if err != nil {
		log.Error("stub.PayOutToken err:", "error", err)
		return err
	}
	balance.Balance -= cashbackValue.CashbackTokens.Amount
	balance.Balance += awards
	err = SaveNodeBalance(stub, cashbackAddr.String(), balance)
	if err != nil {
		log.Error("SaveMedInfo err:", "error", err)
		return err
	}
	return nil
}
