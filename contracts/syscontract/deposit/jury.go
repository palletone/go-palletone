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

	cp, err := stub.GetSystemConfig()
	if err != nil {
		//log.Error("strconv.ParseUint err:", "error", err)
		return shim.Error(err.Error())
	}
	//  交付地址
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
		if invokeTokens.Amount != cp.DepositAmountForJury {
			return shim.Error("Too many or too little.")
		}
		//  加入候选列表
		err = addCandaditeList(stub, invokeAddr, modules.JuryList)
		if err != nil {
			log.Error("addCandaditeList err: ", "error", err)
			return shim.Error(err.Error())
		}
		balance.EnterTime = getTiem(stub)
		//  没有
		balance.Balance = invokeTokens.Amount
		balance.Role = Jury
		err = SaveNodeBalance(stub, invokeAddr.String(), balance)
		if err != nil {
			log.Error("save node balance err: ", "error", err)
			return shim.Error(err.Error())
		}
		return shim.Success(nil)
	} else {
		//  追缴逻辑
		if balance.Role != Jury {
			return shim.Error("not jury")
		}
		all := balance.Balance + invokeTokens.Amount
		if all != cp.DepositAmountForJury {
			return shim.Error("Too many or too little.")
		}
		//这里需要判断是否以及被基金会提前移除候选列表，即在规定时间内该节点没有追缴保证金
		b, err := isInCandidate(stub, invokeAddr.String(), modules.JuryList)
		if err != nil {
			log.Debugf("isInCandidate error: %s", err.Error())
			return shim.Error(err.Error())
		}
		if !b {
			//  加入jury候选列表
			err = addCandaditeList(stub, invokeAddr, modules.JuryList)
			if err != nil {
				log.Error("addCandidateListAndPutStateForMediator err: ", "error", err)
				return shim.Error(err.Error())
			}
		}
		balance.Balance = all
		err = SaveNodeBalance(stub, invokeAddr.String(), balance)
		if err != nil {
			log.Error("save node balance err: ", "error", err)
			return shim.Error(err.Error())
		}
		return shim.Success(nil)
	}
}

func juryApplyQuit(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	err := applyQuitList(Jury, stub)
	if err != nil {
		log.Error("applyQuitList err: ", "error", err)
		return shim.Error(err.Error())
	}
	return shim.Success([]byte(nil))
}

//  处理
func handleJury(stub shim.ChaincodeStubInterface, quitAddr common.Address) error {
	return handleNode(stub, quitAddr, Jury)
}
