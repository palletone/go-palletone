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
	"github.com/palletone/go-palletone/dag/dagconfig"
	"github.com/palletone/go-palletone/dag/modules"
)

func developerPayToDepositContract(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	log.Info("developerPayToDepositContract")
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
		if invokeTokens.Amount != cp.DepositAmountForDeveloper {
			return shim.Error("Too many or too little.")
		}
		//  加入候选列表
		err = addCandaditeList(stub, invokeAddr, modules.DeveloperList)
		if err != nil {
			log.Error("addCandaditeList err: ", "error", err)
			return shim.Error(err.Error())
		}
		balance.EnterTime = getTiem(stub)
		//  没有
		balance.Balance = invokeTokens.Amount
		balance.Role = Developer
		err = SaveNodeBalance(stub, invokeAddr.String(), balance)
		if err != nil {
			log.Error("save node balance err: ", "error", err)
			return shim.Error(err.Error())
		}
		return shim.Success(nil)
	} else {
		//  追缴逻辑
		if balance.Role != Developer {
			return shim.Error("not developer")
		}
		all := balance.Balance + invokeTokens.Amount
		if all != cp.DepositAmountForDeveloper {
			return shim.Error("Too many or too little.")
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

//  申请
func devApplyQuit(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	log.Info("devApplyQuit")
	//  处理逻辑
	err := applyQuitList(Developer, stub)
	if err != nil {
		log.Error("devApplyQuit err: ", "error", err)
		return shim.Error(err.Error())
	}
	return shim.Success(nil)
}

//  处理
func handleDev(stub shim.ChaincodeStubInterface, quitAddr common.Address) error {
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
	//  获取该节点保证金数量
	b, err := GetNodeBalance(stub, quitAddr.String())
	if err != nil {
		return err
	}
	//  调用从合约把token转到请求地址
	gasToken := dagconfig.DagConfig.GetGasToken().ToAsset()
	err = stub.PayOutToken(quitAddr.String(), modules.NewAmountAsset(b.Balance, gasToken), 0)
	if err != nil {
		log.Error("stub.PayOutToken err:", "error", err)
		return err
	}
	//  移除候选列表
	err = moveCandidate(modules.DeveloperList, quitAddr.String(), stub)
	if err != nil {
		log.Error("moveCandidate err:", "error", err)
		return err
	}
	//  删除节点
	err = stub.DelState(string(constants.DEPOSIT_BALANCE_PREFIX) + quitAddr.String())
	if err != nil {
		log.Error("stub.DelState err:", "error", err)
		return err
	}
	return nil
}
